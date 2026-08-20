package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm"
)

func TestBuildTestRequestUsesConfiguredPrompts(t *testing.T) {
	req := buildChannelTestRequest("test-model", true, "system prompt", "user prompt")

	require.Equal(t, "test-model", req.Model)
	require.Len(t, req.Messages, 2)
	require.Equal(t, "system", req.Messages[0].Role)
	require.Equal(t, "system prompt", *req.Messages[0].Content.Content)
	require.Equal(t, "user", req.Messages[1].Role)
	require.Equal(t, "user prompt", *req.Messages[1].Content.Content)
	require.Equal(t, int64(256), *req.MaxCompletionTokens)
	require.True(t, *req.Stream)
}

func TestBuildChannelTestHTTPRequestUsesSelectedAPIFormat(t *testing.T) {
	tests := []struct {
		name      string
		apiFormat llm.APIFormat
	}{
		{name: "OpenAI Chat Completions", apiFormat: llm.APIFormatOpenAIChatCompletion},
		{name: "OpenAI Responses", apiFormat: llm.APIFormatOpenAIResponse},
		{name: "Anthropic Messages", apiFormat: llm.APIFormatAnthropicMessage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inbound, httpRequest, err := buildChannelTestHTTPRequest(tt.apiFormat, "test-model", true, "system prompt", "user prompt")
			require.NoError(t, err)
			require.NotNil(t, inbound)

			request, err := inbound.TransformRequest(context.Background(), httpRequest)
			require.NoError(t, err)
			require.Equal(t, tt.apiFormat, request.APIFormat)
			require.Equal(t, "test-model", request.Model)
			require.Equal(t, llm.RequestTypeChat, request.RequestType)
			require.True(t, *request.Stream)
		})
	}
}
