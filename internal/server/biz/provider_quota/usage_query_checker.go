package provider_quota

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/llm/httpclient"
)

const (
	usageQueryProviderType     = "usage_query"
	usageQueryRequestTimeout   = 15 * time.Second
	maxUsageQueryBodyBytes     = 256 << 10
	maxUsageQueryResponseBytes = 1 << 20
)

var usageQueryAllowedMethods = map[string]struct{}{
	http.MethodGet: {}, http.MethodPost: {}, http.MethodPut: {},
	http.MethodPatch: {}, http.MethodDelete: {}, http.MethodHead: {},
}

type UsageQueryChecker struct {
	httpClient *httpclient.HttpClient
	runtime    UsageQueryScriptRuntime
}

func NewUsageQueryChecker(httpClient *httpclient.HttpClient) *UsageQueryChecker {
	return &UsageQueryChecker{
		httpClient: httpClient,
		runtime:    NewGojaUsageQueryRuntime(),
	}
}

func NewUsageQueryCheckerWithRuntime(httpClient *httpclient.HttpClient, runtime UsageQueryScriptRuntime) *UsageQueryChecker {
	return &UsageQueryChecker{httpClient: httpClient, runtime: runtime}
}

func (c *UsageQueryChecker) SupportsChannel(ch *ent.Channel) bool {
	return ch != nil && ch.Settings != nil && ch.Settings.UsageQuery != nil && ch.Settings.UsageQuery.Enabled
}

func (c *UsageQueryChecker) CheckQuota(ctx context.Context, ch *ent.Channel) (QuotaData, error) {
	if !c.SupportsChannel(ch) {
		return QuotaData{}, errors.New("usage query is not enabled")
	}

	settings := ch.Settings.UsageQuery
	baseURL := strings.TrimSpace(settings.BaseURLOverride)
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(ch.BaseURL), "/")
	}
	if baseURL == "" {
		return QuotaData{}, errors.New("usage query requires a base URL or request URL override")
	}

	apiKey := strings.TrimSpace(ch.Credentials.UsageQueryAPIKey)
	if apiKey == "" {
		for _, candidate := range ch.Credentials.GetEnabledAPIKeys(ch.DisabledAPIKeys) {
			if candidate = strings.TrimSpace(candidate); candidate != "" {
				apiKey = candidate
				break
			}
		}
	}
	if apiKey == "" && ch.Credentials.OAuth != nil {
		apiKey = strings.TrimSpace(ch.Credentials.OAuth.AccessToken)
	}

	requestConfig, err := c.runtime.ParseRequest(ctx, settings.Script, map[string]string{
		"baseUrl":     strings.TrimRight(baseURL, "/"),
		"apiKey":      apiKey,
		"accessToken": apiKey,
		"userId":      strings.TrimSpace(settings.UserID),
	})
	if err != nil {
		return QuotaData{}, fmt.Errorf("failed to parse usage query request: %w", err)
	}

	response, err := c.executeRequest(ctx, ch, baseURL, requestConfig)
	if err != nil {
		return QuotaData{}, err
	}

	result, err := c.runtime.Extract(ctx, settings.Script, response)
	if err != nil {
		return QuotaData{}, fmt.Errorf("failed to extract usage query response: %w", err)
	}

	quotaData := normalizeUsageQueryResult(result)
	quotaData.RawData["showInProviderQuota"] = usageQueryShowsInProviderQuota(ch)
	return quotaData, nil
}

func usageQueryShowsInProviderQuota(ch *ent.Channel) bool {
	return ch != nil && ch.Settings != nil && ch.Settings.UsageQuery != nil &&
		(ch.Settings.UsageQuery.ShowInProviderQuota == nil || *ch.Settings.UsageQuery.ShowInProviderQuota)
}

func (c *UsageQueryChecker) executeRequest(
	ctx context.Context,
	ch *ent.Channel,
	baseURL string,
	config UsageQueryRequest,
) (any, error) {
	method := strings.ToUpper(strings.TrimSpace(config.Method))
	if method == "" {
		method = http.MethodGet
	}
	if _, ok := usageQueryAllowedMethods[method]; !ok {
		return nil, fmt.Errorf("usage query HTTP method %q is not allowed", method)
	}

	requestURL, err := validateUsageQueryURL(ctx, config.URL, baseURL)
	if err != nil {
		return nil, err
	}

	body, contentType, err := encodeUsageQueryBody(config.Body)
	if err != nil {
		return nil, err
	}
	requestCtx, cancel := context.WithTimeout(ctx, usageQueryRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, method, requestURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build usage query request: %w", err)
	}
	if err := setUsageQueryHeaders(req.Header, config.Headers); err != nil {
		return nil, err
	}
	if contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	hc := c.httpClient
	if ch.Settings != nil && ch.Settings.Proxy != nil {
		hc = c.httpClient.WithProxy(ch.Settings.Proxy)
	}
	native := *hc.GetNativeClient()
	native.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("usage query stopped after 3 redirects")
		}
		_, err := validateUsageQueryURL(req.Context(), req.URL.String(), baseURL)
		return err
	}

	resp, err := native.Do(req)
	if err != nil {
		return nil, fmt.Errorf("usage query request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxUsageQueryResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read usage query response: %w", err)
	}
	if len(data) > maxUsageQueryResponseBytes {
		return nil, fmt.Errorf("usage query response exceeds %d bytes", maxUsageQueryResponseBytes)
	}

	var response any
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("usage query returned HTTP %d with invalid JSON: %w", resp.StatusCode, err)
	}
	return response, nil
}

func validateUsageQueryURL(ctx context.Context, rawURL, rawBaseURL string) (*url.URL, error) {
	requestURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("invalid usage query URL: %w", err)
	}
	baseURL, err := url.Parse(strings.TrimSpace(rawBaseURL))
	if err != nil {
		return nil, fmt.Errorf("invalid usage query base URL: %w", err)
	}
	for label, candidate := range map[string]*url.URL{"request": requestURL, "base": baseURL} {
		if candidate.Scheme != "http" && candidate.Scheme != "https" {
			return nil, fmt.Errorf("%s URL must use http or https", label)
		}
		if candidate.Hostname() == "" || candidate.User != nil {
			return nil, fmt.Errorf("%s URL must have a host and no user information", label)
		}
	}
	if requestURL.Scheme != baseURL.Scheme || !strings.EqualFold(requestURL.Host, baseURL.Host) {
		return nil, errors.New("usage query URL must use the configured base URL origin")
	}

	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, requestURL.Hostname())
	if err != nil {
		return nil, fmt.Errorf("failed to resolve usage query host: %w", err)
	}
	for _, address := range addresses {
		if address.IP.IsUnspecified() || address.IP.IsLinkLocalUnicast() || address.IP.IsLinkLocalMulticast() || address.IP.IsMulticast() {
			return nil, fmt.Errorf("usage query host resolves to a disallowed address: %s", address.IP)
		}
	}

	return requestURL, nil
}

func encodeUsageQueryBody(body any) ([]byte, string, error) {
	if body == nil {
		return nil, "", nil
	}
	if text, ok := body.(string); ok {
		if len(text) > maxUsageQueryBodyBytes {
			return nil, "", fmt.Errorf("usage query body exceeds %d bytes", maxUsageQueryBodyBytes)
		}
		return []byte(text), "text/plain", nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode usage query body: %w", err)
	}
	if len(data) > maxUsageQueryBodyBytes {
		return nil, "", fmt.Errorf("usage query body exceeds %d bytes", maxUsageQueryBodyBytes)
	}
	return data, "application/json", nil
}

func setUsageQueryHeaders(target http.Header, headers map[string]string) error {
	if len(headers) > 50 {
		return errors.New("usage query has too many headers")
	}
	for key, value := range headers {
		key = http.CanonicalHeaderKey(strings.TrimSpace(key))
		if key == "" || strings.ContainsAny(key, "\r\n") || strings.ContainsAny(value, "\r\n") {
			return errors.New("usage query contains an invalid header")
		}
		switch key {
		case "Host", "Content-Length", "Connection", "Proxy-Connection", "Transfer-Encoding", "Upgrade":
			return fmt.Errorf("usage query header %q is not allowed", key)
		}
		target.Set(key, value)
	}
	return nil
}

func normalizeUsageQueryResult(result UsageQueryResult) QuotaData {
	if result.Total == nil && result.Remaining != nil && result.Used != nil {
		total := *result.Remaining + *result.Used
		result.Total = &total
	}
	if result.Remaining == nil && result.Total != nil && result.Used != nil {
		remaining := *result.Total - *result.Used
		result.Remaining = &remaining
	}

	status := "unknown"
	isValid := result.IsValid == nil || *result.IsValid
	if isValid {
		status = "available"
		if result.Remaining != nil && *result.Remaining <= 0 {
			status = "exhausted"
		} else if result.Total != nil && result.Used != nil && *result.Total > 0 && *result.Used/(*result.Total) >= WarningThresholdRatio {
			status = "warning"
		}
	}

	rawData := map[string]any{"kind": usageQueryProviderType}
	if result.IsValid != nil {
		rawData["isValid"] = *result.IsValid
	}
	if result.InvalidMessage != "" {
		rawData["invalidMessage"] = result.InvalidMessage
	}
	if result.Remaining != nil {
		rawData["remaining"] = *result.Remaining
	}
	if result.Total != nil {
		rawData["total"] = *result.Total
	}
	if result.Used != nil {
		rawData["used"] = *result.Used
	}
	if result.Unit != "" {
		rawData["unit"] = result.Unit
	}
	if result.PlanName != "" {
		rawData["planName"] = result.PlanName
	}
	if result.Extra != "" {
		rawData["extra"] = result.Extra
	}

	return QuotaData{
		Status:       status,
		ProviderType: usageQueryProviderType,
		RawData:      rawData,
		Ready:        IsReadyStatus(status),
	}
}
