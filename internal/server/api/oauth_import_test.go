package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm/oauth"
	"github.com/looplj/axonhub/llm/transformer/anthropic/claudecode"
	"github.com/looplj/axonhub/llm/transformer/antigravity"
	"github.com/looplj/axonhub/llm/transformer/openai/codex"
	"github.com/looplj/axonhub/llm/transformer/xai"
)

func TestNormalizeSub2APICredentials(t *testing.T) {
	now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		provider         string
		raw              string
		wantClientID     string
		wantProjectID    string
		wantBaseURL      string
		wantExpiresAt    time.Time
		wantRefreshToken string
	}{
		{
			name:             "Codex account export uses AxonHub client identity",
			provider:         "codex",
			raw:              `{"platform":"openai","type":"oauth","credentials":{"access_token":"codex-access","refresh_token":"codex-refresh","id_token":"codex-id","client_id":"exported-client","expires_at":1787306400}}`,
			wantClientID:     codex.ClientID,
			wantExpiresAt:    time.Unix(1787306400, 0).UTC(),
			wantRefreshToken: "codex-refresh",
		},
		{
			name:             "Claude nested response accepts camel case tokens",
			provider:         "claudecode",
			raw:              `{"data":{"credentials":{"accessToken":"claude-access","refreshToken":"claude-refresh","expiresIn":"3600","scopes":["user:inference"]}}}`,
			wantClientID:     claudecode.ClientID,
			wantExpiresAt:    now.Add(time.Hour),
			wantRefreshToken: "claude-refresh",
		},
		{
			name:             "Gemini account export defaults Code Assist client identity",
			provider:         "gemini",
			raw:              `{"platform":"gemini","type":"oauth","credentials":{"access_token":"gemini-access","refresh_token":"gemini-refresh","project_id":"project-123","expires_at":"2026-08-21T18:00:00+08:00"}}`,
			wantClientID:     antigravity.ClientID,
			wantProjectID:    "project-123",
			wantExpiresAt:    time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC),
			wantRefreshToken: "gemini-refresh",
		},
		{
			name:             "Grok account export retains OAuth client and CLI endpoint",
			provider:         "grok",
			raw:              `{"platform":"grok","type":"oauth","credentials":{"access_token":"grok-access","refresh_token":"grok-refresh","client_id":"grok-cli-client","base_url":"https://cli-chat-proxy.grok.com/v1","expires_at":"1787306400000"}}`,
			wantClientID:     "grok-cli-client",
			wantBaseURL:      xai.OAuthBaseURL,
			wantExpiresAt:    time.Unix(1787306400, 0).UTC(),
			wantRefreshToken: "grok-refresh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeSub2APICredentials(tt.provider, tt.raw, now)
			require.NoError(t, err)
			require.Equal(t, tt.wantBaseURL, result.BaseURL)

			credentials, err := oauth.ParseCredentialsJSON(result.Credentials)
			require.NoError(t, err)
			require.Equal(t, tt.wantClientID, credentials.ClientID)
			require.Equal(t, tt.wantProjectID, credentials.ProjectID)
			require.True(t, tt.wantExpiresAt.Equal(credentials.ExpiresAt), "expires_at mismatch: want %s, got %s", tt.wantExpiresAt, credentials.ExpiresAt)
			require.Equal(t, tt.wantRefreshToken, credentials.RefreshToken)
		})
	}
}

func TestNormalizeSub2APICredentials_RejectsMismatchedPlatformAndUnsafeGrokURL(t *testing.T) {
	now := time.Date(2026, time.August, 21, 10, 0, 0, 0, time.UTC)

	_, err := normalizeSub2APICredentials("codex", `{"platform":"anthropic","credentials":{"access_token":"access","refresh_token":"refresh"}}`, now)
	require.ErrorContains(t, err, "cannot be imported as codex")

	_, err = normalizeSub2APICredentials("grok", `{"platform":"grok","credentials":{"access_token":"access","refresh_token":"refresh","client_id":"client","base_url":"http://grok.example.test/v1"}}`, now)
	require.ErrorContains(t, err, "HTTPS URL")
}
