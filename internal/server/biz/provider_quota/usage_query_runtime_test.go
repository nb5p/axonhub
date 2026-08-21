package provider_quota

import (
	"context"
	"testing"
	"time"

	"github.com/dop251/goja"
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

	result, err := runtime.Extract(
		t.Context(),
		testUsageQueryScript,
		UsageQueryHTTPResponse{Body: map[string]any{"is_active": true, "balance": 12.5}},
		UsageQueryScriptContext{},
	)
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
	  responseVersion: 2,
	  request: { url: "{{baseUrl}}/quota", method: "GET" },
	  extractor: function(response, context) {
	    return {
	      balance: { remaining: response.body.balance, unit: "A$" },
	      text: response.body.plan + "\n" + context.now,
	      tags: ["Plus", "并发: 8"],
	      progress: {
	        windows: [
	          { id: "daily", durationSeconds: 86400, remainingPercent: 0, resetAt: "2026-08-19T00:00:00+08:00" },
	          { id: "weekly", remainingPercent: 80 },
	          { id: "monthly", remainingPercent: 70 }
	        ]
      }
    };
  }
})`

	result, err := runtime.Extract(
		t.Context(),
		script,
		UsageQueryHTTPResponse{Body: map[string]any{"plan": "Codex Lite", "balance": 39.5}},
		UsageQueryScriptContext{Now: "2026-08-18T07:36:46+08:00"},
	)
	require.NoError(t, err)
	require.NotNil(t, result.Balance)
	require.Equal(t, 39.5, result.Balance.Remaining)
	require.Equal(t, "Codex Lite\n2026-08-18T07:36:46+08:00", result.Text)
	require.Equal(t, []string{"Plus", "并发: 8"}, result.Tags)
	require.NotNil(t, result.Progress)
	require.Len(t, result.Progress.Windows, 3)
	require.Equal(t, "daily", result.Progress.Windows[0].ID)
	require.Equal(t, 86400, *result.Progress.Windows[0].DurationSeconds)
	require.Equal(t, "", result.Progress.Windows[1].ResetAt)
	require.Equal(t, 80.0, *result.Progress.Windows[1].RemainingPercent)
	require.Equal(t, 70.0, *result.Progress.Windows[2].RemainingPercent)
}

func TestGojaUsageQueryRuntime_ExtractsResponseEnvelopeAndContext(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	script := `({
	  responseVersion: 2,
	  request: { url: "{{baseUrl}}/quota", method: "GET" },
	  extractor: function(response, context) {
	    return {
	      balance: { remaining: response.status === 207 ? response.body.balance : 0, unit: "A$" },
	      text: response.headers["x-plan"] + " " + context.now
    };
  }
})`

	var usesEnvelope bool
	err := runtime.run(t.Context(), script, func(_ *goja.Runtime, config *goja.Object) error {
		usesEnvelope = usageQueryUsesResponseEnvelope(config)
		return validateUsageQueryResponseVersion(config)
	})
	require.NoError(t, err)
	require.True(t, usesEnvelope)

	result, err := runtime.Extract(
		t.Context(),
		script,
		UsageQueryHTTPResponse{
			Status:  207,
			Headers: map[string]string{"x-plan": "Pro"},
			Body:    map[string]any{"balance": 12.5},
		},
		UsageQueryScriptContext{Now: "2026-08-18T07:36:46+08:00"},
	)
	require.NoError(t, err)
	require.NotNil(t, result.Balance)
	require.Equal(t, 12.5, result.Balance.Remaining)
	require.Equal(t, "Pro 2026-08-18T07:36:46+08:00", result.Text)
}

func TestGojaUsageQueryRuntime_RejectsInvalidProgressWindow(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	script := `({ responseVersion: 2, request: { url: "{{baseUrl}}/quota" }, extractor: function() {
  return { progress: { windows: [{ id: "每日", remainingPercent: 101, resetAt: "2026-08-18T07:36:46Z" }] } };
} })`

	_, err := runtime.Extract(t.Context(), script, UsageQueryHTTPResponse{}, UsageQueryScriptContext{})
	require.ErrorContains(t, err, "ASCII English identifier")
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
