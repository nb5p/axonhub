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
