package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
)

func TestSelectAPIFormat(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: "openai/responses"},
		{APIFormat: "openai/embeddings"},
		{APIFormat: "openai/image_generation"},
		{APIFormat: "openai/moderations"},
	}

	require.Equal(t, "openai/responses", SelectAPIFormat(endpoints, &llm.Request{RequestType: llm.RequestTypeChat}))
	require.Equal(t, "openai/embeddings", SelectAPIFormat(endpoints, &llm.Request{RequestType: llm.RequestTypeEmbedding}))
	require.Equal(t, "openai/image_generation", SelectAPIFormat(endpoints, &llm.Request{RequestType: llm.RequestTypeImage}))
	require.Equal(t, "openai/moderations", SelectAPIFormat(endpoints, &llm.Request{RequestType: llm.RequestTypeModeration}))

	geminiEndpoints := []objects.ChannelEndpoint{
		{APIFormat: llm.APIFormatGeminiContents.String()},
		{APIFormat: llm.APIFormatGeminiEmbedding.String()},
	}

	require.Equal(t, llm.APIFormatGeminiContents.String(), SelectAPIFormat(geminiEndpoints, &llm.Request{RequestType: llm.RequestTypeChat}))
	require.Equal(t, llm.APIFormatGeminiEmbedding.String(), SelectAPIFormat(geminiEndpoints, &llm.Request{RequestType: llm.RequestTypeEmbedding}))
	require.Equal(t, llm.APIFormatGeminiContents.String(), SelectAPIFormat(geminiEndpoints, &llm.Request{RequestType: llm.RequestTypeImage}))
}

func TestSelectAPIFormat_RemoteCompactionForcesResponses(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: llm.APIFormatOpenAIChatCompletion.String()},
		{APIFormat: llm.APIFormatOpenAIResponse.String()},
	}

	req := &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
		ProviderExtensions: &llm.ProviderExtensions{
			OpenAIResponses: &llm.OpenAIResponsesProviderExtensions{
				Request: &llm.OpenAIResponsesRequestExtensions{
					RawInputItems: []llm.OpenAIResponsesRawFragment{{Type: "compaction_trigger"}},
				},
			},
		},
	}

	require.Equal(t, llm.APIFormatOpenAIResponse.String(), SelectAPIFormat(endpoints, req))
}

func TestSelectAPIFormat_PrefersMatchingFormat(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: "openai/responses"},
		{APIFormat: "openai/chat_completions"},
	}

	require.Equal(t, "openai/chat_completions", SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIChatCompletion,
	}))
}

func TestSelectAPIFormat_FallsBackWhenNoMatch(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: "openai/responses"},
	}

	require.Equal(t, "openai/responses", SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIChatCompletion,
	}))
}

func TestSelectAPIFormat_Video(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: "openai/video"},
		{APIFormat: "seedance/video"},
	}

	require.Equal(t, "openai/video", SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeVideo,
		APIFormat:   llm.APIFormatOpenAIVideo,
	}))

	require.Equal(t, "seedance/video", SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeVideo,
		APIFormat:   llm.APIFormatSeedanceVideo,
	}))
}

func TestSelectAPIFormat_Compact(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: "openai/responses"},
		{APIFormat: "openai/responses_compact"},
	}

	require.Equal(t, "openai/responses_compact", SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeCompact,
		APIFormat:   llm.APIFormatOpenAIResponseCompact,
	}))
}

func TestSelectAPIFormat_AlphaSearch(t *testing.T) {
	endpoints := []objects.ChannelEndpoint{
		{APIFormat: llm.APIFormatOpenAIResponse.String()},
		{APIFormat: llm.APIFormatOpenAICodexAlphaSearch.String()},
	}

	require.Equal(t, llm.APIFormatOpenAICodexAlphaSearch.String(), SelectAPIFormat(endpoints, &llm.Request{
		RequestType: llm.RequestTypeAlphaSearch,
		APIFormat:   llm.APIFormatOpenAICodexAlphaSearch,
	}))
}

func TestSupportedPassThroughConversions(t *testing.T) {
	options := SupportedPassThroughConversions()
	require.Len(t, options, 33)

	keys := make(map[string]struct{}, len(options))
	for _, option := range options {
		require.NotEqual(t, option.SourceFormat, option.TargetFormat)
		_, exists := keys[option.Key]
		require.False(t, exists, "duplicate conversion key %s", option.Key)
		keys[option.Key] = struct{}{}
	}

	chatToResponses := PassThroughConversionKey(llm.APIFormatOpenAIChatCompletion, llm.APIFormatOpenAIResponse)
	require.Contains(t, keys, chatToResponses)
	require.True(t, IsSupportedPassThroughConversion(chatToResponses))
	require.False(t, IsSupportedPassThroughConversion("openai/chat_completions->unsupported/format"))
}
