package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
)

func TestPopulateAPIFormat_DisabledResponsesUsesSameChannelChatEndpoint(t *testing.T) {
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "glm-5.2"}
	candidates := []*ChannelModelsCandidate{{
		Channel: testChannelWithEndpoints(
			&objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
				Model:      "glm-5.2",
				APIFormats: []string{llm.APIFormatOpenAIResponse.String()},
			}}},
			[]objects.ChannelEndpoint{{APIFormat: llm.APIFormatOpenAIResponse.String()}},
		),
		Models: []biz.ChannelModelEntry{entry},
	}}

	populated := populateAPIFormat(candidates, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
	})

	require.Len(t, populated, 1)
	require.Equal(t, llm.APIFormatOpenAIChatCompletion.String(), populated[0].APIFormat)
	require.Equal(t, []biz.ChannelModelEntry{entry}, populated[0].Models)
}

func TestPopulateAPIFormat_UnrestrictedModelStillPrefersResponses(t *testing.T) {
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "glm-5.2"}
	candidates := []*ChannelModelsCandidate{{
		Channel: testChannelWithEndpoints(nil, []objects.ChannelEndpoint{{APIFormat: llm.APIFormatOpenAIResponse.String()}}),
		Models:  []biz.ChannelModelEntry{entry},
	}}

	populated := populateAPIFormat(candidates, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
	})

	require.Len(t, populated, 1)
	require.Equal(t, llm.APIFormatOpenAIResponse.String(), populated[0].APIFormat)
}

func TestPopulateAPIFormat_ExcludesChannelWhenEveryChatEndpointIsDisabled(t *testing.T) {
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "glm-5.2"}
	candidates := []*ChannelModelsCandidate{{
		Channel: testChannelWithEndpoints(
			&objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
				Model:      "glm-5.2",
				APIFormats: []string{llm.APIFormatOpenAIChatCompletion.String(), llm.APIFormatOpenAIResponse.String()},
			}}},
			[]objects.ChannelEndpoint{{APIFormat: llm.APIFormatOpenAIResponse.String()}},
		),
		Models: []biz.ChannelModelEntry{entry},
	}}

	populated := populateAPIFormat(candidates, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
	})

	require.Empty(t, populated)
}

func TestPopulateAPIFormat_RemoteCompactionExcludesChannelWithoutResponses(t *testing.T) {
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "glm-5.2"}
	candidates := []*ChannelModelsCandidate{{
		Channel: testChannelWithEndpoints(
			&objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
				Model:      "glm-5.2",
				APIFormats: []string{llm.APIFormatOpenAIResponse.String()},
			}}},
			[]objects.ChannelEndpoint{{APIFormat: llm.APIFormatOpenAIResponse.String()}},
		),
		Models: []biz.ChannelModelEntry{entry},
	}}

	populated := populateAPIFormat(candidates, &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
		ProviderExtensions: &llm.ProviderExtensions{OpenAIResponses: &llm.OpenAIResponsesProviderExtensions{
			Request: &llm.OpenAIResponsesRequestExtensions{RawInputItems: []llm.OpenAIResponsesRawFragment{{Type: "compaction_trigger"}}},
		}},
	})

	require.Empty(t, populated)
}

func TestIsModelAPIFormatDisabled_MatchesActualModel(t *testing.T) {
	settings := &objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
		Model:      "provider/glm-5.2",
		APIFormats: []string{llm.APIFormatOpenAIResponse.String()},
	}}}
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "provider/glm-5.2"}

	require.True(t, IsModelAPIFormatDisabled(settings, entry, llm.APIFormatOpenAIResponse.String(), llm.RequestClientOther))
	require.False(t, IsModelAPIFormatDisabled(settings, entry, llm.APIFormatOpenAIChatCompletion.String(), llm.RequestClientOther))
}

func TestPopulateAPIFormat_CodexScopedResponsesRestriction(t *testing.T) {
	entry := biz.ChannelModelEntry{RequestModel: "glm-5.2", ActualModel: "glm-5.2"}
	settings := &objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
		Model:      "glm-5.2",
		APIFormats: []string{llm.APIFormatOpenAIResponse.String()},
		Clients:    []string{string(llm.RequestClientCodex)},
	}}}
	makeCandidates := func() []*ChannelModelsCandidate {
		return []*ChannelModelsCandidate{{
			Channel: testChannelWithEndpoints(settings, []objects.ChannelEndpoint{{APIFormat: llm.APIFormatOpenAIResponse.String()}}),
			Models:  []biz.ChannelModelEntry{entry},
		}}
	}

	codex := populateAPIFormat(makeCandidates(), &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
		Client:      llm.RequestClientCodex,
	})
	require.Len(t, codex, 1)
	require.Equal(t, llm.APIFormatOpenAIChatCompletion.String(), codex[0].APIFormat)

	other := populateAPIFormat(makeCandidates(), &llm.Request{
		RequestType: llm.RequestTypeChat,
		APIFormat:   llm.APIFormatOpenAIResponse,
		Client:      llm.RequestClientOther,
	})
	require.Len(t, other, 1)
	require.Equal(t, llm.APIFormatOpenAIResponse.String(), other[0].APIFormat)
}

func testChannelWithEndpoints(settings *objects.ChannelSettings, endpoints []objects.ChannelEndpoint) *biz.Channel {
	return &biz.Channel{Channel: &ent.Channel{
		Type:      channel.TypeOpencodeGo,
		Settings:  settings,
		Endpoints: endpoints,
	}}
}
