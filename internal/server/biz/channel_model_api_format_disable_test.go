package biz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/objects"
)

func TestNormalizeDisabledModelAPIFormats(t *testing.T) {
	settings := &objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{
		{
			Model:      " glm-5.2 ",
			APIFormats: []string{" openai/responses ", "openai/responses", "openai/chat_completions"},
		},
		{
			Model:      "glm-5.2",
			APIFormats: []string{"openai/chat_completions", "anthropic/messages"},
		},
	}}

	require.NoError(t, NormalizeDisabledModelAPIFormats(settings))
	require.Equal(t, []objects.DisabledModelAPIFormat{{
		Model:      "glm-5.2",
		APIFormats: []string{"openai/responses", "openai/chat_completions", "anthropic/messages"},
	}}, settings.DisabledModelAPIFormats)
}

func TestNormalizeDisabledModelAPIFormatsRejectsEmptyRestriction(t *testing.T) {
	settings := &objects.ChannelSettings{DisabledModelAPIFormats: []objects.DisabledModelAPIFormat{{
		Model:      "glm-5.2",
		APIFormats: []string{"  "},
	}}}

	require.EqualError(t, NormalizeDisabledModelAPIFormats(settings), "disabled model API format entry 1 requires at least one API format")
}
