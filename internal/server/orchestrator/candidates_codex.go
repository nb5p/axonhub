package orchestrator

import (
	"context"

	"github.com/samber/lo"

	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/llm"
)

type CodexAlphaSearchSelector struct {
	wrapped CandidateSelector
}

func WithCodexAlphaSearchSelector(wrapped CandidateSelector) *CodexAlphaSearchSelector {
	return &CodexAlphaSearchSelector{wrapped: wrapped}
}

func (s *CodexAlphaSearchSelector) Select(ctx context.Context, req *llm.Request) ([]*ChannelModelsCandidate, error) {
	candidates, err := s.wrapped.Select(ctx, req)
	if err != nil {
		return nil, err
	}

	return lo.Filter(candidates, func(candidate *ChannelModelsCandidate, _ int) bool {
		return candidate != nil && candidate.Channel != nil && candidate.Channel.Type == channel.TypeCodex
	}), nil
}
