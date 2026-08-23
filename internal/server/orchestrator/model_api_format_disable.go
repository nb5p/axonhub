package orchestrator

import (
	"strings"

	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
)

// IsModelAPIFormatDisabled reports whether a channel setting prohibits using
// apiFormat as the outbound endpoint for entry. A setting may name either the
// request alias or the actual provider model so model mappings, prefixes, and
// automatic model trimming share the same restriction.
func IsModelAPIFormatDisabled(settings *objects.ChannelSettings, entry biz.ChannelModelEntry, apiFormat string) bool {
	if settings == nil || len(settings.DisabledModelAPIFormats) == 0 {
		return false
	}

	apiFormat = strings.TrimSpace(apiFormat)
	if apiFormat == "" {
		return false
	}

	for _, disabled := range settings.DisabledModelAPIFormats {
		if !matchesDisabledModel(disabled.Model, entry) {
			continue
		}

		for _, format := range disabled.APIFormats {
			if strings.TrimSpace(format) == apiFormat {
				return true
			}
		}
	}

	return false
}

// FilterEndpointsForModel removes only the endpoint formats disabled for the
// given model. It intentionally returns an empty slice when every endpoint is
// disabled, so callers can exclude the channel instead of falling back to a
// forbidden endpoint.
func FilterEndpointsForModel(
	endpoints []objects.ChannelEndpoint,
	settings *objects.ChannelSettings,
	entry biz.ChannelModelEntry,
) []objects.ChannelEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	filtered := make([]objects.ChannelEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if IsModelAPIFormatDisabled(settings, entry, endpoint.APIFormat) {
			continue
		}

		filtered = append(filtered, endpoint)
	}

	return filtered
}

func matchesDisabledModel(model string, entry biz.ChannelModelEntry) bool {
	model = strings.TrimSpace(model)
	if model == "" {
		return false
	}

	return model == entry.RequestModel || model == entry.ActualModel
}
