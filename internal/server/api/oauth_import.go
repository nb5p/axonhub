package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/looplj/axonhub/llm/oauth"
	"github.com/looplj/axonhub/llm/transformer/anthropic/claudecode"
	"github.com/looplj/axonhub/llm/transformer/antigravity"
	"github.com/looplj/axonhub/llm/transformer/openai/codex"
)

// OAuthImportHandlers accepts one exported Sub2API account object and converts
// its credentials to AxonHub's single-channel OAuth JSON. It intentionally
// does not create channels or accept arrays, so importing remains an explicit
// operator action in the channel editor.
type OAuthImportHandlers struct{}

func NewOAuthImportHandlers() *OAuthImportHandlers {
	return &OAuthImportHandlers{}
}

type importOAuthCredentialsRequest struct {
	Provider    string `json:"provider" binding:"required"`
	AccountJSON string `json:"account_json" binding:"required"`
}

type importOAuthCredentialsResponse struct {
	Credentials string `json:"credentials"`
}

// ImportSub2APICredentials normalizes a single Sub2API account export locally.
// POST /admin/oauth/import/sub2api.
func (h *OAuthImportHandlers) ImportSub2APICredentials(c *gin.Context) {
	var req importOAuthCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("invalid request format"))
		return
	}

	result, err := normalizeSub2APICredentials(req.Provider, req.AccountJSON, time.Now())
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func normalizeSub2APICredentials(provider, raw string, now time.Time) (*importOAuthCredentialsResponse, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "codex" && provider != "claudecode" && provider != "gemini" && provider != "grok" {
		return nil, errors.New("unsupported OAuth import provider")
	}

	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	var root map[string]any
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("invalid account JSON: %w", err)
	}
	if root == nil {
		return nil, errors.New("account JSON must be an object")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errors.New("account JSON must contain exactly one object")
	}

	candidate, ok := findOAuthCredentialObject(root)
	if !ok {
		return nil, errors.New("account JSON does not contain access_token and refresh_token")
	}
	if err := validateSub2APIAccountProvider(provider, root); err != nil {
		return nil, err
	}

	accessToken := firstString(candidate, "access_token", "accessToken")
	refreshToken := firstString(candidate, "refresh_token", "refreshToken", "rt")
	if accessToken == "" || refreshToken == "" {
		return nil, errors.New("account JSON requires access_token and refresh_token")
	}

	clientID := firstString(candidate, "client_id", "clientId")
	if clientID == "" {
		clientID = firstString(root, "client_id", "clientId")
	}
	switch provider {
	case "codex":
		clientID = codex.ClientID
	case "claudecode":
		if clientID == "" {
			clientID = claudecode.ClientID
		}
	case "gemini":
		if clientID == "" {
			clientID = antigravity.ClientID
		}
	case "grok":
		if clientID == "" {
			return nil, errors.New("Grok account JSON requires client_id")
		}
	}

	expiresAt, err := credentialExpiresAt(candidate, now)
	if err != nil {
		return nil, err
	}

	creds := &oauth.OAuthCredentials{
		ClientID:     clientID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IDToken:      firstString(candidate, "id_token", "idToken"),
		ExpiresAt:    expiresAt,
		TokenType:    firstString(candidate, "token_type", "tokenType"),
		Scopes:       credentialScopes(candidate),
	}

	if provider == "gemini" {
		creds.ProjectID = firstString(candidate, "project_id", "projectId")
		if creds.ProjectID == "" {
			creds.ProjectID = firstString(root, "project_id", "projectId")
		}
		if creds.ProjectID == "" {
			return nil, errors.New("Gemini account JSON requires project_id")
		}
	}

	credentials, err := creds.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("serialize credentials: %w", err)
	}

	return &importOAuthCredentialsResponse{Credentials: credentials}, nil
}

// validateSub2APIAccountProvider rejects an account export whose explicit
// platform does not match the selected AxonHub channel type. Plain credential
// JSON intentionally has no platform and remains supported.
func validateSub2APIAccountProvider(provider string, root map[string]any) error {
	platform := strings.ToLower(firstString(root, "platform"))
	if platform == "" {
		return nil
	}

	expected := map[string][]string{
		"codex":      {"openai"},
		"claudecode": {"anthropic"},
		"gemini":     {"gemini", "antigravity"},
		"grok":       {"grok"},
	}
	for _, candidate := range expected[provider] {
		if platform == candidate {
			return nil
		}
	}

	return fmt.Errorf("account platform %q cannot be imported as %s", platform, provider)
}

func findOAuthCredentialObject(root map[string]any) (map[string]any, bool) {
	candidates := []map[string]any{root}
	for _, parent := range []map[string]any{root, objectValue(root, "data")} {
		if parent == nil {
			continue
		}
		for _, key := range []string{"credentials", "credential", "tokens", "token"} {
			if value := objectValue(parent, key); value != nil {
				candidates = append(candidates, value)
			}
		}
	}

	for _, candidate := range candidates {
		if firstString(candidate, "access_token", "accessToken") != "" &&
			firstString(candidate, "refresh_token", "refreshToken", "rt") != "" {
			return candidate, true
		}
	}

	return nil, false
}

func objectValue(values map[string]any, key string) map[string]any {
	value, ok := values[key]
	if !ok {
		return nil
	}
	object, _ := value.(map[string]any)
	return object
}

func firstString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok {
			return strings.TrimSpace(text)
		}
	}

	return ""
}

func credentialScopes(values map[string]any) []string {
	for _, key := range []string{"scope", "scopes"} {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch scopes := value.(type) {
		case string:
			return strings.Fields(scopes)
		case []any:
			result := make([]string, 0, len(scopes))
			for _, scope := range scopes {
				if text, ok := scope.(string); ok && strings.TrimSpace(text) != "" {
					result = append(result, strings.TrimSpace(text))
				}
			}
			return result
		}
	}

	return nil
}

func credentialExpiresAt(values map[string]any, now time.Time) (time.Time, error) {
	for _, key := range []string{"expires_at", "expiresAt"} {
		if value, ok := values[key]; ok {
			expiresAt, err := parseCredentialTime(value)
			if err != nil {
				return time.Time{}, fmt.Errorf("invalid %s: %w", key, err)
			}
			return expiresAt, nil
		}
	}

	for _, key := range []string{"expires_in", "expiresIn"} {
		if value, ok := values[key]; ok {
			seconds, err := numericSeconds(value)
			if err != nil || seconds <= 0 {
				return time.Time{}, fmt.Errorf("invalid %s", key)
			}
			return now.Add(time.Duration(seconds) * time.Second), nil
		}
	}

	// Sub2API OAuth account exports normally contain expires_at. If an older
	// export omitted it, refresh on the next request rather than treating the
	// imported access token as indefinitely valid.
	return now.Add(time.Hour), nil
}

func parseCredentialTime(value any) (time.Time, error) {
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return time.Time{}, errors.New("empty timestamp")
		}
		if timestamp, err := time.Parse(time.RFC3339Nano, text); err == nil {
			return timestamp, nil
		}
		seconds, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return time.Time{}, errors.New("not RFC3339 or unix time")
		}
		return unixCredentialTime(seconds)
	case json.Number:
		seconds, err := typed.Float64()
		if err != nil {
			return time.Time{}, err
		}
		return unixCredentialTime(seconds)
	case float64:
		return unixCredentialTime(typed)
	default:
		return time.Time{}, errors.New("must be an RFC3339 string or unix timestamp")
	}
}

func numericSeconds(value any) (int64, error) {
	switch typed := value.(type) {
	case json.Number:
		return typed.Int64()
	case float64:
		if math.Trunc(typed) != typed || typed > math.MaxInt64 || typed < math.MinInt64 {
			return 0, errors.New("not an integer")
		}
		return int64(typed), nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
	default:
		return 0, errors.New("not a number")
	}
}

func unixCredentialTime(value float64) (time.Time, error) {
	if !isFinite(value) || value <= 0 {
		return time.Time{}, errors.New("must be a positive finite unix timestamp")
	}
	if value > 1000000000000 {
		value /= 1000
	}
	if value > math.MaxInt64 || math.Trunc(value) != value {
		return time.Time{}, errors.New("must be an integer unix timestamp")
	}
	return time.Unix(int64(value), 0).UTC(), nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
