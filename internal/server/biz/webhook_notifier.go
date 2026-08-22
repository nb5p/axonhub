package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/samber/lo"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/webhookdelivery"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm/httpclient"
)

const (
	EventChannelAutoDisabled   = "channel.auto_disabled"
	EventChannelError          = "channel.error"
	EventQuotaWindowExhausted  = "quota.window_exhausted"
	EventQuotaBalanceExhausted = "quota.balance_exhausted"

	webhookDeliveryHistoryLimit = 500
)

const (
	defaultWebhookMethod    = http.MethodPost
	defaultWebhookTimeoutMs = 3000
	defaultBarkTitle        = "AxonHub"
	defaultBarkBody         = "{{.Event}}"
)

type ChannelAutoDisabledEvent struct {
	ChannelID       int
	ChannelName     string
	ChannelProvider string
	ChannelBaseURL  string
	ChannelStatus   string
	StatusCode      int
	Threshold       int
	ActualCount     int
	Reason          string
	OccurredAt      time.Time
}

type ChannelErrorEvent struct {
	ChannelID       int
	ChannelName     string
	ChannelProvider string
	ChannelBaseURL  string
	ChannelStatus   string
	StatusCode      int
	ErrorMessage    string
	APIKeySuffix    string
	OccurredAt      time.Time
}

type QuotaWindowExhaustedEvent struct {
	ChannelID        int
	ChannelName      string
	ChannelProvider  string
	ChannelBaseURL   string
	ChannelStatus    string
	ProviderType     string
	LimitType        string
	Window           string
	RemainingPercent *float64
	ResetAt          *time.Time
	OccurredAt       time.Time
}

type QuotaBalanceExhaustedEvent struct {
	ChannelID       int
	ChannelName     string
	ChannelProvider string
	ChannelBaseURL  string
	ChannelStatus   string
	ProviderType    string
	Remaining       float64
	Unit            string
	OccurredAt      time.Time
}

// WebhookDeliveryHistoryItem is deliberately independent of the Ent entity so
// GraphQL only exposes the sanitized audit fields that are safe for operators.
type WebhookDeliveryHistoryItem struct {
	ID             int                   `json:"id"`
	Event          string                `json:"event"`
	TargetName     string                `json:"target_name"`
	TargetType     string                `json:"target_type"`
	URL            string                `json:"url"`
	Method         string                `json:"method"`
	RequestHeaders []objects.HeaderEntry `json:"request_headers"`
	RequestBody    string                `json:"request_body"`
	Status         string                `json:"status"`
	ResponseStatus int                   `json:"response_status"`
	ErrorMessage   string                `json:"error_message"`
	OccurredAt     time.Time             `json:"occurred_at"`
}

type WebhookRenderContext struct {
	Event      string `json:"event"`
	Severity   string `json:"severity"`
	OccurredAt string `json:"occurred_at"`

	Channel struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Provider string `json:"provider"`
		BaseURL  string `json:"base_url"`
		Status   string `json:"status"`
	} `json:"channel"`

	Trigger struct {
		Type        string `json:"type"`
		StatusCode  int    `json:"status_code"`
		Threshold   int    `json:"threshold"`
		ActualCount int    `json:"actual_count"`
		Reason      string `json:"reason"`
	} `json:"trigger"`

	Error struct {
		StatusCode   int    `json:"status_code"`
		Message      string `json:"message"`
		APIKeySuffix string `json:"api_key_suffix"`
	} `json:"error"`

	Quota struct {
		ProviderType     string   `json:"provider_type"`
		LimitType        string   `json:"limit_type"`
		Window           string   `json:"window"`
		RemainingPercent *float64 `json:"remaining_percent"`
		ResetAt          string   `json:"reset_at"`
		Remaining        *float64 `json:"remaining"`
		Unit             string   `json:"unit"`
	} `json:"quota"`
}

type WebhookNotifier struct {
	SystemService *SystemService
	httpClient    *httpclient.HttpClient
}

func NewWebhookNotifier(systemService *SystemService, httpClient *httpclient.HttpClient) *WebhookNotifier {
	return &WebhookNotifier{
		SystemService: systemService,
		httpClient:    httpClient,
	}
}

func (n *WebhookNotifier) NotifyChannelAutoDisabled(ctx context.Context, event ChannelAutoDisabledEvent) {
	n.dispatch(ctx, EventChannelAutoDisabled, func() (WebhookRenderContext, error) {
		renderCtx := newWebhookRenderContext(EventChannelAutoDisabled, "warning", event.OccurredAt)
		populateWebhookChannel(&renderCtx, event.ChannelID, event.ChannelName, event.ChannelProvider, event.ChannelBaseURL, event.ChannelStatus)
		renderCtx.Trigger.Type = "error_status_rule"
		renderCtx.Trigger.StatusCode = event.StatusCode
		renderCtx.Trigger.Threshold = event.Threshold
		renderCtx.Trigger.ActualCount = event.ActualCount
		renderCtx.Trigger.Reason = sanitizeWebhookText(event.Reason)
		return renderCtx, nil
	})
}

func (n *WebhookNotifier) NotifyChannelAutoDisabledAsync(ctx context.Context, event ChannelAutoDisabledEvent) {
	n.dispatchAsync(ctx, EventChannelAutoDisabled, func() (WebhookRenderContext, error) {
		renderCtx := newWebhookRenderContext(EventChannelAutoDisabled, "warning", event.OccurredAt)
		populateWebhookChannel(&renderCtx, event.ChannelID, event.ChannelName, event.ChannelProvider, event.ChannelBaseURL, event.ChannelStatus)
		renderCtx.Trigger.Type = "error_status_rule"
		renderCtx.Trigger.StatusCode = event.StatusCode
		renderCtx.Trigger.Threshold = event.Threshold
		renderCtx.Trigger.ActualCount = event.ActualCount
		renderCtx.Trigger.Reason = sanitizeWebhookText(event.Reason)
		return renderCtx, nil
	})
}

func (n *WebhookNotifier) NotifyChannelErrorAsync(ctx context.Context, event ChannelErrorEvent) {
	n.dispatchAsync(ctx, EventChannelError, func() (WebhookRenderContext, error) {
		renderCtx := newWebhookRenderContext(EventChannelError, "error", event.OccurredAt)
		populateWebhookChannel(&renderCtx, event.ChannelID, event.ChannelName, event.ChannelProvider, event.ChannelBaseURL, event.ChannelStatus)
		renderCtx.Error.StatusCode = event.StatusCode
		renderCtx.Error.Message = sanitizeWebhookText(event.ErrorMessage)
		renderCtx.Error.APIKeySuffix = event.APIKeySuffix
		return renderCtx, nil
	})
}

func (n *WebhookNotifier) NotifyQuotaWindowExhaustedAsync(ctx context.Context, event QuotaWindowExhaustedEvent) {
	n.dispatchAsync(ctx, EventQuotaWindowExhausted, func() (WebhookRenderContext, error) {
		renderCtx := newWebhookRenderContext(EventQuotaWindowExhausted, "warning", event.OccurredAt)
		populateWebhookChannel(&renderCtx, event.ChannelID, event.ChannelName, event.ChannelProvider, event.ChannelBaseURL, event.ChannelStatus)
		renderCtx.Quota.ProviderType = event.ProviderType
		renderCtx.Quota.LimitType = event.LimitType
		renderCtx.Quota.Window = event.Window
		renderCtx.Quota.RemainingPercent = event.RemainingPercent
		if event.ResetAt != nil {
			renderCtx.Quota.ResetAt = event.ResetAt.Format(time.RFC3339)
		}
		return renderCtx, nil
	})
}

func (n *WebhookNotifier) NotifyQuotaBalanceExhaustedAsync(ctx context.Context, event QuotaBalanceExhaustedEvent) {
	n.dispatchAsync(ctx, EventQuotaBalanceExhausted, func() (WebhookRenderContext, error) {
		renderCtx := newWebhookRenderContext(EventQuotaBalanceExhausted, "warning", event.OccurredAt)
		populateWebhookChannel(&renderCtx, event.ChannelID, event.ChannelName, event.ChannelProvider, event.ChannelBaseURL, event.ChannelStatus)
		remaining := event.Remaining
		renderCtx.Quota.ProviderType = event.ProviderType
		renderCtx.Quota.Remaining = &remaining
		renderCtx.Quota.Unit = event.Unit
		return renderCtx, nil
	})
}

func newWebhookRenderContext(event, severity string, occurredAt time.Time) WebhookRenderContext {
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	return WebhookRenderContext{
		Event:      event,
		Severity:   severity,
		OccurredAt: occurredAt.Format(time.RFC3339),
	}
}

func populateWebhookChannel(renderCtx *WebhookRenderContext, id int, name, provider, baseURL, status string) {
	renderCtx.Channel.ID = id
	renderCtx.Channel.Name = name
	renderCtx.Channel.Provider = provider
	renderCtx.Channel.BaseURL = sanitizeWebhookURL(baseURL)
	renderCtx.Channel.Status = status
}

func (n *WebhookNotifier) dispatch(ctx context.Context, eventName string, build func() (WebhookRenderContext, error)) {
	ctx = authz.WithSystemBypass(context.WithoutCancel(ctx), "webhook-notifier")
	targets := n.selectTargets(*n.SystemService.WebhookNotifierConfigOrDefault(ctx), eventName)
	if len(targets) == 0 {
		return
	}

	renderCtx, err := build()
	if err != nil {
		log.Warn(ctx, "failed to build webhook notification", log.String("event", eventName), log.Cause(err))
		return
	}
	n.notifyTargets(ctx, eventName, renderCtx, targets)
}

func (n *WebhookNotifier) dispatchAsync(ctx context.Context, eventName string, build func() (WebhookRenderContext, error)) {
	ctx = authz.WithSystemBypass(context.WithoutCancel(ctx), "webhook-notifier")
	targets := n.selectTargets(*n.SystemService.WebhookNotifierConfigOrDefault(ctx), eventName)
	if len(targets) == 0 {
		return
	}

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error(ctx, "webhook notification goroutine panicked", log.String("event", eventName), log.Any("panic", recovered))
			}
		}()

		renderCtx, err := build()
		if err != nil {
			log.Warn(ctx, "failed to build webhook notification", log.String("event", eventName), log.Cause(err))
			return
		}
		n.notifyTargets(ctx, eventName, renderCtx, targets)
	}()
}

func (n *WebhookNotifier) notifyTargets(ctx context.Context, eventName string, renderCtx WebhookRenderContext, targets []WebhookTarget) {
	for _, target := range targets {
		request, err := prepareWebhookRequest(target, renderCtx)
		if err != nil {
			log.Warn(ctx, "failed to render webhook request template",
				log.String("event", eventName),
				log.String("target", target.Name),
				log.Cause(err),
			)
			n.recordDelivery(ctx, eventName, target, webhookRequest{
				URL:     target.URL,
				Method:  defaultWebhookMethod,
				Headers: rawWebhookHeaders(target.Headers),
				Body:    target.Body,
			}, 0, err)
			continue
		}

		statusCode, sendErr := n.send(ctx, target, request)
		n.recordDelivery(ctx, eventName, target, request, statusCode, sendErr)
		if sendErr != nil {
			log.Warn(ctx, "failed to send webhook notification",
				log.String("event", eventName),
				log.String("target", target.Name),
				log.Cause(sendErr),
			)
		}
	}
}

func (n *WebhookNotifier) selectTargets(cfg WebhookNotifierConfig, eventName string) []WebhookTarget {
	subscription, ok := lo.Find(cfg.Subscriptions, func(item WebhookSubscription) bool {
		return item.Event == eventName
	})
	if !ok {
		return nil
	}

	names := subscription.TargetNames
	if len(names) == 0 || len(cfg.Targets) == 0 {
		return nil
	}

	targets := make([]WebhookTarget, 0, len(names))
	for _, name := range names {
		target, ok := lo.Find(cfg.Targets, func(item WebhookTarget) bool {
			return item.Name == name
		})
		if !ok || !target.Enabled || strings.TrimSpace(target.URL) == "" {
			continue
		}
		target.Type = normalizeWebhookTargetType(target.Type)
		targets = append(targets, target)
	}

	return targets
}

type webhookRequest struct {
	URL     string
	Method  string
	Headers http.Header
	Body    string
}

func prepareWebhookRequest(target WebhookTarget, renderCtx WebhookRenderContext) (webhookRequest, error) {
	headers, err := renderWebhookHeaders(target.Headers, renderCtx)
	if err != nil {
		return webhookRequest{}, err
	}

	request := webhookRequest{
		URL:     strings.TrimSpace(target.URL),
		Method:  defaultWebhookMethod,
		Headers: headers,
	}

	switch normalizeWebhookTargetType(target.Type) {
	case WebhookTargetTypeBark:
		if strings.TrimSpace(target.BarkDeviceKey) == "" {
			return webhookRequest{}, fmt.Errorf("Bark device key is required")
		}
		request.URL, err = barkPushURL(request.URL)
		if err != nil {
			return webhookRequest{}, err
		}

		title := target.BarkTitle
		if strings.TrimSpace(title) == "" {
			title = defaultBarkTitle
		}
		title, err = renderWebhookTemplate(title, renderCtx)
		if err != nil {
			return webhookRequest{}, fmt.Errorf("render bark title: %w", err)
		}

		bodyTemplate := target.Body
		if strings.TrimSpace(bodyTemplate) == "" {
			bodyTemplate = defaultBarkBody
		}
		body, err := renderWebhookTemplate(bodyTemplate, renderCtx)
		if err != nil {
			return webhookRequest{}, fmt.Errorf("render bark body: %w", err)
		}

		payload := map[string]string{
			"device_key": target.BarkDeviceKey,
			"title":      title,
			"body":       body,
		}
		if strings.TrimSpace(target.BarkLevel) != "" {
			payload["level"] = strings.TrimSpace(target.BarkLevel)
		}
		if strings.TrimSpace(target.BarkGroup) != "" {
			payload["group"] = strings.TrimSpace(target.BarkGroup)
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return webhookRequest{}, fmt.Errorf("encode bark payload: %w", err)
		}
		request.Body = string(data)
	default:
		request.Body, err = renderWebhookTemplate(target.Body, renderCtx)
		if err != nil {
			return webhookRequest{}, fmt.Errorf("render webhook body: %w", err)
		}
	}

	if request.Headers.Get("Content-Type") == "" {
		request.Headers.Set("Content-Type", "application/json")
	}

	return request, nil
}

func barkPushURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid Bark server URL")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(parsed.Path, "/push") {
		parsed.Path += "/push"
	}
	return parsed.String(), nil
}

func (n *WebhookNotifier) send(ctx context.Context, target WebhookTarget, request webhookRequest) (int, error) {
	timeout := time.Duration(target.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = defaultWebhookTimeoutMs * time.Millisecond
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := n.httpClient
	if target.Proxy != nil {
		client = client.WithProxy(target.Proxy)
	}

	response, err := client.Do(reqCtx, &httpclient.Request{
		Method:      request.Method,
		URL:         request.URL,
		Headers:     request.Headers,
		ContentType: request.Headers.Get("Content-Type"),
		Body:        []byte(request.Body),
	})
	if err != nil {
		var httpErr *httpclient.Error
		if errors.As(err, &httpErr) {
			return httpErr.StatusCode, fmt.Errorf("do request: %w", err)
		}
		return 0, fmt.Errorf("do request: %w", err)
	}

	return response.StatusCode, nil
}

func (n *WebhookNotifier) recordDelivery(ctx context.Context, eventName string, target WebhookTarget, request webhookRequest, statusCode int, sendErr error) {
	status := webhookdelivery.StatusSuccess
	errorMessage := ""
	if sendErr != nil {
		status = webhookdelivery.StatusFailed
		errorMessage = sanitizeWebhookText(sendErr.Error())
	}

	client := n.SystemService.entFromContext(ctx)
	_, err := client.WebhookDelivery.Create().
		SetEvent(eventName).
		SetTargetName(target.Name).
		SetTargetType(normalizeWebhookTargetType(target.Type)).
		SetURL(sanitizeWebhookURL(request.URL)).
		SetMethod(request.Method).
		SetRequestHeaders(sanitizeWebhookHeaders(request.Headers)).
		SetRequestBody(sanitizeWebhookBody(request.Body)).
		SetStatus(status).
		SetResponseStatus(statusCode).
		SetErrorMessage(errorMessage).
		Save(ctx)
	if err != nil {
		log.Warn(ctx, "failed to record webhook delivery", log.String("event", eventName), log.String("target", target.Name), log.Cause(err))
		return
	}

	oldestRetained, err := client.WebhookDelivery.Query().
		Order(ent.Desc(webhookdelivery.FieldID)).
		Offset(webhookDeliveryHistoryLimit - 1).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			log.Warn(ctx, "failed to locate webhook history retention boundary", log.Cause(err))
		}
		return
	}
	if _, err := client.WebhookDelivery.Delete().Where(webhookdelivery.IDLT(oldestRetained.ID)).Exec(ctx); err != nil {
		log.Warn(ctx, "failed to prune webhook delivery history", log.Cause(err))
	}
}

func (n *WebhookNotifier) DeliveryHistory(ctx context.Context, limit int) ([]*WebhookDeliveryHistoryItem, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > webhookDeliveryHistoryLimit {
		limit = webhookDeliveryHistoryLimit
	}

	deliveries, err := n.SystemService.entFromContext(ctx).WebhookDelivery.Query().
		Order(ent.Desc(webhookdelivery.FieldCreatedAt), ent.Desc(webhookdelivery.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query webhook delivery history: %w", err)
	}

	return lo.Map(deliveries, func(delivery *ent.WebhookDelivery, _ int) *WebhookDeliveryHistoryItem {
		return &WebhookDeliveryHistoryItem{
			ID:             delivery.ID,
			Event:          delivery.Event,
			TargetName:     delivery.TargetName,
			TargetType:     delivery.TargetType,
			URL:            delivery.URL,
			Method:         delivery.Method,
			RequestHeaders: delivery.RequestHeaders,
			RequestBody:    delivery.RequestBody,
			Status:         delivery.Status.String(),
			ResponseStatus: delivery.ResponseStatus,
			ErrorMessage:   delivery.ErrorMessage,
			OccurredAt:     delivery.CreatedAt,
		}
	}), nil
}

func rawWebhookHeaders(headers []objects.HeaderEntry) http.Header {
	result := make(http.Header, len(headers))
	for _, header := range headers {
		if key := strings.TrimSpace(header.Key); key != "" {
			result.Set(key, header.Value)
		}
	}
	return result
}

func renderWebhookHeaders(headers []objects.HeaderEntry, renderCtx WebhookRenderContext) (http.Header, error) {
	result := make(http.Header, len(headers))
	for _, header := range headers {
		key := strings.TrimSpace(header.Key)
		if key == "" {
			continue
		}

		value, err := renderWebhookTemplate(header.Value, renderCtx)
		if err != nil {
			return nil, fmt.Errorf("header %q: %w", key, err)
		}

		result.Set(key, value)
	}

	return result, nil
}

func renderWebhookTemplate(value string, renderCtx WebhookRenderContext) (string, error) {
	if !strings.Contains(value, "{{") || !strings.Contains(value, "}}") {
		return value, nil
	}

	tmpl, err := template.New("webhook").Funcs(template.FuncMap{}).Parse(value)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderCtx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func sanitizeWebhookHeaders(headers http.Header) []objects.HeaderEntry {
	result := make([]objects.HeaderEntry, 0, len(headers))
	for key, values := range headers {
		result = append(result, objects.HeaderEntry{
			Key:   key,
			Value: sanitizeWebhookHeaderValue(key, strings.Join(values, ", ")),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Key < result[j].Key })
	return result
}

func sanitizeWebhookHeaderValue(key, value string) string {
	lowerKey := strings.ToLower(strings.TrimSpace(key))
	if isSensitiveWebhookKey(lowerKey) {
		return maskWebhookSecret(value)
	}
	return sanitizeWebhookText(value)
}

func sanitizeWebhookBody(value string) string {
	var data any
	if err := json.Unmarshal([]byte(value), &data); err == nil {
		sanitizeWebhookJSON(data)
		if encoded, err := json.Marshal(data); err == nil {
			return string(encoded)
		}
	}
	return sanitizeWebhookText(value)
}

func sanitizeWebhookJSON(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, childValue := range typed {
			if isSensitiveWebhookKey(strings.ToLower(key)) {
				if secret, ok := childValue.(string); ok {
					typed[key] = maskWebhookSecret(secret)
				} else {
					typed[key] = "[REDACTED]"
				}
				continue
			}
			sanitizeWebhookJSON(childValue)
		}
	case []any:
		for _, child := range typed {
			sanitizeWebhookJSON(child)
		}
	}
}

func sanitizeWebhookURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return sanitizeWebhookText(rawURL)
	}
	if parsed.User != nil {
		parsed.User = url.User("[REDACTED]")
	}
	query := parsed.Query()
	for key, values := range query {
		if isSensitiveWebhookKey(strings.ToLower(key)) {
			for i := range values {
				values[i] = maskWebhookSecret(values[i])
			}
			query[key] = values
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func sanitizeWebhookText(value string) string {
	value = bearerTokenPattern.ReplaceAllStringFunc(value, func(match string) string {
		parts := strings.Fields(match)
		if len(parts) != 2 {
			return "Bearer [REDACTED]"
		}
		return parts[0] + " " + maskWebhookSecret(parts[1])
	})
	return assignmentSecretPattern.ReplaceAllStringFunc(value, func(match string) string {
		separator := strings.IndexAny(match, "=:")
		if separator < 0 {
			return "[REDACTED]"
		}
		return match[:separator+1] + maskWebhookSecret(match[separator+1:])
	})
}

var (
	bearerTokenPattern      = regexp.MustCompile(`(?i)bearer\s+[^\s,;\"}]+`)
	assignmentSecretPattern = regexp.MustCompile(`(?i)(?:api[_-]?key|token|secret|password|device[_-]?key)\s*[=:]\s*[^\s,;\"}]+`)
)

func isSensitiveWebhookKey(key string) bool {
	key = strings.ReplaceAll(key, "_", "-")
	return strings.Contains(key, "authorization") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "api-key") ||
		strings.Contains(key, "apikey") ||
		strings.Contains(key, "device-key") ||
		strings.Contains(key, "cookie")
}

func maskWebhookSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "[REDACTED]"
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return "Bearer " + maskWebhookSecret(strings.TrimSpace(value[len("Bearer "):]))
	}
	if len(value) <= 4 {
		return "[REDACTED]"
	}
	return "****" + value[len(value)-4:]
}
