package gql

import (
	"fmt"

	"github.com/looplj/axonhub/internal/server/orchestrator"
)

func toChannelTestAPIFormat(apiFormat *ChannelTestAPIFormat) (*orchestrator.ChannelTestAPIFormat, error) {
	if apiFormat == nil {
		return nil, nil
	}

	var value orchestrator.ChannelTestAPIFormat
	switch *apiFormat {
	case ChannelTestAPIFormatOpenaiChatCompletion:
		value = orchestrator.ChannelTestAPIFormatOpenAIChatCompletion
	case ChannelTestAPIFormatOpenaiResponse:
		value = orchestrator.ChannelTestAPIFormatOpenAIResponse
	case ChannelTestAPIFormatAnthropicMessages:
		value = orchestrator.ChannelTestAPIFormatAnthropicMessages
	case ChannelTestAPIFormatGeminiContents:
		value = orchestrator.ChannelTestAPIFormatGeminiContents
	default:
		return nil, fmt.Errorf("unsupported channel test API format %q", *apiFormat)
	}

	return &value, nil
}
