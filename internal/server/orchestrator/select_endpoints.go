package orchestrator

import (
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
)

// chatCapableAPIFormats lists the API formats that can handle chat requests.
var chatCapableAPIFormats = map[string]struct{}{
	"openai/chat_completions": {},
	"openai/responses":        {},
	"anthropic/messages":      {},
	"gemini/contents":         {},
	"ollama/chat":             {},
}

// IsChatCapableAPIFormat reports whether an API format can handle chat requests.
func IsChatCapableAPIFormat(apiFormat string) bool {
	_, ok := chatCapableAPIFormats[apiFormat]

	return ok
}

// compactCapableAPIFormats lists API formats for compact requests.
var compactCapableAPIFormats = map[string]struct{}{
	"openai/responses_compact": {},
}

// remoteCompactionCapableAPIFormats contains the only outbound format that can
// replay Codex native remote compaction v2 requests with their original
// protocol semantics. Generic chat conversion cannot represent the opaque
// compaction trigger and encrypted context.
var remoteCompactionCapableAPIFormats = map[string]struct{}{
	"openai/responses": {},
}

var alphaSearchCapableAPIFormats = map[string]struct{}{
	"openai/codex_alpha_search": {},
}

// completionCapableAPIFormats lists API formats for completion requests.
var completionCapableAPIFormats = map[string]struct{}{
	"openai/completions": {},
}

// embeddingCapableAPIFormats lists API formats for embedding requests.
var embeddingCapableAPIFormats = map[string]struct{}{
	"openai/embeddings": {},
	"jina/embeddings":   {},
	"gemini/embeddings": {},
}

// moderationCapableAPIFormats lists API formats for moderation requests.
var moderationCapableAPIFormats = map[string]struct{}{
	"openai/moderations": {},
}

// imageCapableAPIFormats lists API formats for image requests.
var imageCapableAPIFormats = map[string]struct{}{
	"openai/image_generation": {},
	"openai/image_edit":       {},
	"openai/image_variation":  {},
}

// rerankCapableAPIFormats lists API formats for rerank requests.
var rerankCapableAPIFormats = map[string]struct{}{
	"jina/rerank": {},
}

// videoCapableAPIFormats lists API formats for video requests.
var videoCapableAPIFormats = map[string]struct{}{
	"openai/video":   {},
	"seedance/video": {},
}

// speechCapableAPIFormats lists API formats for text-to-speech (TTS) requests.
var speechCapableAPIFormats = map[string]struct{}{
	"openai/audio_speech": {},
}

// transcriptionCapableAPIFormats lists API formats for speech-to-text (STT) transcription requests.
var transcriptionCapableAPIFormats = map[string]struct{}{
	"openai/audio_transcriptions": {},
}

// translationCapableAPIFormats lists API formats for speech-to-text (STT) translation requests.
var translationCapableAPIFormats = map[string]struct{}{
	"openai/audio_translations": {},
}

// PassThroughConversionOption describes one supported, directional API format
// conversion that can be exempted from pass-through-first routing.
type PassThroughConversionOption struct {
	Key          string
	RequestType  llm.RequestType
	SourceFormat llm.APIFormat
	TargetFormat llm.APIFormat
}

type conversionFormats struct {
	requestType llm.RequestType
	sources     []llm.APIFormat
	targets     []llm.APIFormat
}

var supportedConversionFormats = []conversionFormats{
	{
		requestType: llm.RequestTypeChat,
		sources: []llm.APIFormat{
			llm.APIFormatOpenAIChatCompletion,
			llm.APIFormatOpenAIResponse,
			llm.APIFormatAnthropicMessage,
			llm.APIFormatGeminiContents,
			llm.APIFormatAiSDKDataStream,
		},
		targets: []llm.APIFormat{
			llm.APIFormatOpenAIChatCompletion,
			llm.APIFormatOpenAIResponse,
			llm.APIFormatAnthropicMessage,
			llm.APIFormatGeminiContents,
			llm.APIFormatOllamaChat,
		},
	},
	{
		requestType: llm.RequestTypeEmbedding,
		sources: []llm.APIFormat{
			llm.APIFormatOpenAIEmbedding,
			llm.APIFormatJinaEmbedding,
		},
		targets: []llm.APIFormat{
			llm.APIFormatOpenAIEmbedding,
			llm.APIFormatJinaEmbedding,
			llm.APIFormatGeminiEmbedding,
		},
	},
	{
		requestType: llm.RequestTypeImage,
		sources: []llm.APIFormat{
			llm.APIFormatOpenAIImageGeneration,
			llm.APIFormatOpenAIImageEdit,
			llm.APIFormatOpenAIImageVariation,
		},
		targets: []llm.APIFormat{
			llm.APIFormatOpenAIImageGeneration,
			llm.APIFormatOpenAIImageEdit,
			llm.APIFormatOpenAIImageVariation,
		},
	},
	{
		requestType: llm.RequestTypeVideo,
		sources: []llm.APIFormat{
			llm.APIFormatOpenAIVideo,
			llm.APIFormatSeedanceVideo,
		},
		targets: []llm.APIFormat{
			llm.APIFormatOpenAIVideo,
			llm.APIFormatSeedanceVideo,
		},
	},
}

// PassThroughConversionKey returns the stable key used to persist a
// directional conversion exception.
func PassThroughConversionKey(source, target llm.APIFormat) string {
	return source.String() + "->" + target.String()
}

// SupportedPassThroughConversions returns every currently supported
// cross-format conversion. Same-format routes are omitted because they do not
// perform protocol conversion.
func SupportedPassThroughConversions() []PassThroughConversionOption {
	options := make([]PassThroughConversionOption, 0)
	for _, formats := range supportedConversionFormats {
		for _, source := range formats.sources {
			for _, target := range formats.targets {
				if source == target {
					continue
				}

				options = append(options, PassThroughConversionOption{
					Key:          PassThroughConversionKey(source, target),
					RequestType:  formats.requestType,
					SourceFormat: source,
					TargetFormat: target,
				})
			}
		}
	}

	return options
}

// IsSupportedPassThroughConversion reports whether a persisted conversion key
// is still supported by the current transformer/endpoint registry.
func IsSupportedPassThroughConversion(key string) bool {
	for _, option := range SupportedPassThroughConversions() {
		if option.Key == key {
			return true
		}
	}

	return false
}

// SelectAPIFormat selects the most appropriate APIFormat from a channel's resolved endpoints
// based on the request type and inbound API format. Prefers an endpoint whose API format
// matches the inbound request format so that pass-through can be enabled when identical
// formats are used. Falls back to the first capable endpoint, then the first endpoint.
func SelectAPIFormat(endpoints []objects.ChannelEndpoint, req *llm.Request) string {
	if len(endpoints) == 0 {
		return ""
	}

	preferredFormat := string(req.APIFormat)
	allowed := llm.CapableAPIFormats(req.RequestType)
	if isOpenAIResponsesRemoteCompaction(req) {
		allowed = remoteCompactionCapableAPIFormats
		preferredFormat = llm.APIFormatOpenAIResponse.String()
	} else if req.RequestType == llm.RequestTypeAlphaSearch {
		allowed = alphaSearchCapableAPIFormats
	}

	if allowed != nil {
		if preferredFormat != "" {
			for _, ep := range endpoints {
				if _, ok := allowed[ep.APIFormat]; ok && ep.APIFormat == preferredFormat {
					return ep.APIFormat
				}
			}
		}

		for _, ep := range endpoints {
			if _, ok := allowed[ep.APIFormat]; ok {
				return ep.APIFormat
			}
		}
	}

	return endpoints[0].APIFormat
}
