package provider_quota

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm/httpclient"
)

func TestUsageQueryChecker_CheckQuota(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/user/self", r.URL.Path)
		require.Equal(t, "Bearer admin-key", r.Header.Get("Authorization"))
		require.Equal(t, "42", r.Header.Get("New-Api-User"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Plan", "pro")
		_, _ = w.Write([]byte("{\"success\":true,\"data\":{\"group\":\"pro\",\"quota\":1000000,\"used_quota\":4000000}}"))
	}))
	defer server.Close()

	script := "({\n" +
		" responseVersion: 2,\n" +
		" request: { url: \"{{baseUrl}}/api/user/self\", method: \"GET\", headers: { Authorization: \"Bearer {{accessToken}}\", \"New-Api-User\": \"{{userId}}\" } },\n" +
		" extractor: function(response, context) { const body = response.body; return {\n" +
		"  balance: { remaining: response.status === 200 ? body.data.quota / 500000 : 0, unit: \"A$\" },\n" +
		"  text: response.headers[\"x-plan\"] + \" \" + context.now,\n" +
		"  progress: { windows: [{ id: \"daily\", remainingPercent: 20, resetAt: context.now }] }\n" +
		" }; }\n" +
		"})"
	ch := &ent.Channel{
		BaseURL: server.URL,
		Credentials: objects.ChannelCredentials{
			APIKeys:          []string{"request-key"},
			UsageQueryAPIKey: "admin-key",
		},
		Settings: &objects.ChannelSettings{
			UsageQuery: &objects.ChannelUsageQuerySettings{
				Enabled:             true,
				ShowInProviderQuota: lo.ToPtr(false),
				Preset:              objects.ChannelUsageQueryPresetCustom,
				UserID:              "42",
				Script:              script,
			},
		},
	}

	checker := NewUsageQueryChecker(httpclient.NewHttpClient())
	result, err := checker.CheckQuota(t.Context(), ch)
	require.NoError(t, err)
	require.Equal(t, "warning", result.Status)
	require.True(t, result.Ready)
	balance, ok := result.RawData["balance"].(*UsageQueryBalance)
	require.True(t, ok)
	require.Equal(t, 2.0, balance.Remaining)
	require.Equal(t, "A$", balance.Unit)
	require.Regexp(t, `^pro \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$`, result.RawData["text"])
	require.Equal(t, false, result.RawData["showInProviderQuota"])
}

func TestUsageQueryChecker_RejectsCrossOriginScriptURL(t *testing.T) {
	ch := &ent.Channel{
		BaseURL: "https://example.com",
		Settings: &objects.ChannelSettings{
			UsageQuery: &objects.ChannelUsageQuerySettings{
				Enabled: true,
				Preset:  objects.ChannelUsageQueryPresetCustom,
				Script:  "({ request: { url: \"https://other.example/user/balance\", method: \"GET\" }, extractor: function(response) { return response; } })",
			},
		},
	}

	checker := NewUsageQueryChecker(httpclient.NewHttpClient())
	_, err := checker.CheckQuota(t.Context(), ch)
	require.ErrorContains(t, err, "configured base URL origin")
}

func TestNormalizeUsageQueryResult_InvalidPlanIsUnknown(t *testing.T) {
	valid := false
	result := normalizeUsageQueryResult(UsageQueryResult{
		IsValid:        &valid,
		InvalidMessage: "expired",
	})
	require.Equal(t, "unknown", result.Status)
	require.False(t, result.Ready)
	require.Equal(t, "expired", result.RawData["error"])
}

func TestNormalizeUsageQueryResult_UsesStructuredProgressWindows(t *testing.T) {
	result := normalizeUsageQueryResult(UsageQueryResult{
		Text: "Codex Lite",
		Progress: &UsageQueryProgress{Windows: []UsageQueryProgressWindow{
			{ID: "daily", RemainingPercent: lo.ToPtr(0.0), ResetAt: "2026-08-19T00:00:00+00:00"},
			{ID: "weekly", RemainingPercent: lo.ToPtr(80.0)},
			{ID: "monthly", RemainingPercent: lo.ToPtr(70.0)},
		}},
	})

	require.Equal(t, "exhausted", result.Status)
	require.False(t, result.Ready)
	require.Equal(t, "Codex Lite", result.RawData["text"])
	progress, ok := result.RawData["progress"].(*UsageQueryProgress)
	require.True(t, ok)
	require.Len(t, progress.Windows, 3)
	require.Empty(t, progress.Windows[2].ResetAt)
}

func TestNormalizeUsageQueryResult_ProgressTakesPrecedenceOverZeroBalance(t *testing.T) {
	result := normalizeUsageQueryResult(UsageQueryResult{
		Balance: &UsageQueryBalance{Remaining: 0, Unit: "A$"},
		Progress: &UsageQueryProgress{Windows: []UsageQueryProgressWindow{
			{ID: "daily", RemainingPercent: lo.ToPtr(80.0)},
		}},
	})

	require.Equal(t, "available", result.Status)
	require.True(t, result.Ready)
}

func TestNormalizeUsageQueryResult_RejectsIncompleteProgressWindow(t *testing.T) {
	err := validateUsageQueryResult(UsageQueryResult{
		Progress: &UsageQueryProgress{Windows: []UsageQueryProgressWindow{{ID: "daily"}}},
	})
	require.NoError(t, err)
}
