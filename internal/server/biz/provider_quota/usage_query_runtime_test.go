package provider_quota

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testUsageQueryScript = "({\n" +
	"  request: { url: \"{{baseUrl}}/user/balance\", method: \"GET\", headers: { Authorization: \"Bearer {{apiKey}}\" } },\n" +
	"  extractor: function(response) {\n" +
	"    const balance = response?.balance ?? 0;\n" +
	"    return { isValid: response.is_active, remaining: balance, unit: \"USD\" };\n" +
	"  }\n" +
	"})"

func TestGojaUsageQueryRuntime_ParsesRequestAndExtractsResult(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()

	request, err := runtime.ParseRequest(t.Context(), testUsageQueryScript, map[string]string{
		"baseUrl": "https://example.com",
		"apiKey":  "key-quoted",
	})
	require.NoError(t, err)
	require.Equal(t, "https://example.com/user/balance", request.URL)
	require.Equal(t, "Bearer key-quoted", request.Headers["Authorization"])

	result, err := runtime.Extract(t.Context(), testUsageQueryScript, map[string]any{
		"is_active": true,
		"balance":   12.5,
	})
	require.NoError(t, err)
	require.NotNil(t, result.IsValid)
	require.True(t, *result.IsValid)
	require.NotNil(t, result.Remaining)
	require.Equal(t, 12.5, *result.Remaining)
	require.Equal(t, "USD", result.Unit)
}

func TestGojaUsageQueryRuntime_ExtractsStructuredTextAndProgressWindows(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	script := `({
  request: { url: "{{baseUrl}}/quota", method: "GET" },
  extractor: function(response) {
    return {
      text: { isValid: true, planName: response.plan, remaining: 39.5, unit: "USD" },
      progress: {
        windows: [
          { id: "daily", label: "Daily", used: 10, total: 10, unit: "USD", windowStart: "2026-08-18T00:00:00Z", resetAt: "2026-08-19T00:00:00Z" },
          { id: "weekly", label: "Weekly", usedPercent: 20 },
          { id: "monthly", remaining: 70, total: 100, unit: "USD" }
        ]
      }
    };
  }
})`

	result, err := runtime.Extract(t.Context(), script, map[string]any{"plan": "Codex Lite"})
	require.NoError(t, err)
	require.NotNil(t, result.Text)
	require.True(t, *result.Text.IsValid)
	require.Equal(t, "Codex Lite", result.Text.PlanName)
	require.NotNil(t, result.Progress)
	require.Len(t, result.Progress.Windows, 3)
	require.Equal(t, "daily", result.Progress.Windows[0].ID)
	require.Equal(t, 10.0, *result.Progress.Windows[0].Used)
	require.Equal(t, "", result.Progress.Windows[1].ResetAt)
	require.Equal(t, 20.0, *result.Progress.Windows[1].UsedPercent)
	require.Equal(t, 70.0, *result.Progress.Windows[2].Remaining)
}

func TestGojaUsageQueryRuntime_RejectsUnknownVariables(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	script := "({ request: { url: \"{{baseUrl}}/{{unknown}}\", method: \"GET\" }, extractor: function(response) { return response; } })"
	_, err := runtime.ParseRequest(t.Context(), script, map[string]string{"baseUrl": "https://example.com"})
	require.ErrorContains(t, err, "unknown variable")
}

func TestGojaUsageQueryRuntime_RequiresExtractor(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	script := `({ request: { url: "{{baseUrl}}/balance", method: "GET" } })`
	_, err := runtime.ParseRequest(t.Context(), script, map[string]string{"baseUrl": "https://example.com"})
	require.ErrorContains(t, err, "extractor as a function")
}

func TestGojaUsageQueryRuntime_InterruptsInfiniteLoop(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	started := time.Now()
	_, err := runtime.ParseRequest(context.Background(), "(() => { while (true) {} })()", nil)
	require.ErrorContains(t, err, "timed out")
	require.Less(t, time.Since(started), 2*time.Second)
}
