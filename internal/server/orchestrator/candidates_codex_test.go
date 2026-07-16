package orchestrator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
)

type staticCandidateSelector struct {
	candidates []*ChannelModelsCandidate
}

func (s staticCandidateSelector) Select(context.Context, *llm.Request) ([]*ChannelModelsCandidate, error) {
	return s.candidates, nil
}

func TestCodexAlphaSearchSelector_OnlyKeepsCodexChannels(t *testing.T) {
	codexCandidate := &ChannelModelsCandidate{Channel: &biz.Channel{Channel: &ent.Channel{Type: channel.TypeCodex}}}
	openAICandidate := &ChannelModelsCandidate{Channel: &biz.Channel{Channel: &ent.Channel{Type: channel.TypeOpenai}}}
	selector := WithCodexAlphaSearchSelector(staticCandidateSelector{
		candidates: []*ChannelModelsCandidate{openAICandidate, codexCandidate},
	})

	got, err := selector.Select(context.Background(), &llm.Request{RequestType: llm.RequestTypeAlphaSearch})
	require.NoError(t, err)
	require.Equal(t, []*ChannelModelsCandidate{codexCandidate}, got)
}
