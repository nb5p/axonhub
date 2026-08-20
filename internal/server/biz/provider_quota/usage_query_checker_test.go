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
		_, _ = w.Write([]byte("{\"success\":true,\"data\":{\"group\":\"pro\",\"quota\":1000000,\"used_quota\":4000000}}"))
	}))
	defer server.Close()

	script := "({\n" +
		" request: { url: \"{{baseUrl}}/api/user/self\", method: \"GET\", headers: { Authorization: \"Bearer {{accessToken}}\", \"New-Api-User\": \"{{userId}}\" } },\n" +
		" extractor: function(response) { return {\n" +
		"  planName: response.data.group, remaining: response.data.quota / 500000,\n" +
		"  used: response.data.used_quota / 500000,\n" +
		"  total: (response.data.quota + response.data.used_quota) / 500000, unit: \"USD\"\n" +
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
				Preset:              objects.ChannelUsageQueryPresetNewAPI,
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
	require.Equal(t, "pro", result.RawData["planName"])
	require.Equal(t, 2.0, result.RawData["remaining"])
	require.Equal(t, 8.0, result.RawData["used"])
	require.Equal(t, 10.0, result.RawData["total"])
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
	require.Equal(t, "expired", result.RawData["invalidMessage"])
}
