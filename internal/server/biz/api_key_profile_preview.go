package biz

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/model"
	"github.com/looplj/axonhub/internal/objects"
)

// APIKeyProfilePreview describes the API formats and models that an unsaved API
// key profile would expose.
type APIKeyProfilePreview struct {
	APIFormats        []string                     `json:"apiFormats"`
	Models            []*APIKeyProfilePreviewModel `json:"models"`
	PreferPassThrough bool                         `json:"preferPassThrough"`
}

type APIKeyProfilePreviewModel struct {
	ID       string                         `json:"id"`
	Channels []*APIKeyProfilePreviewChannel `json:"channels"`
}

type APIKeyProfilePreviewChannel struct {
	ID                    int      `json:"id"`
	Name                  string   `json:"name"`
	OrderingWeight        int      `json:"orderingWeight"`
	PassThroughAPIFormats []string `json:"passThroughApiFormats"`
}

// PreviewAPIKeyProfile calculates visibility with the same ListEnabledModels
// path used by the public model-list endpoints, while substituting the unsaved
// profile supplied by the editor.
func (svc *ModelService) PreviewAPIKeyProfile(
	ctx context.Context,
	apiKey *ent.APIKey,
	profile objects.APIKeyProfile,
) (*APIKeyProfilePreview, error) {
	previewProfile := profile
	if strings.TrimSpace(previewProfile.Name) == "" {
		previewProfile.Name = "__preview__"
	}

	previewAPIKey := *apiKey
	previewAPIKey.Profiles = &objects.APIKeyProfiles{
		ActiveProfile: previewProfile.Name,
		Profiles:      []objects.APIKeyProfile{previewProfile},
	}

	previewCtx := contexts.WithAPIKey(ctx, &previewAPIKey)
	visibleModels, err := svc.ListEnabledModels(previewCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to list visible models: %w", err)
	}

	channels, _ := filterChannelsForAPIKey(svc.channelService.GetEnabledChannels(), &previewAPIKey)
	configuredModels, err := svc.previewConfiguredModels(previewCtx, visibleModels)
	if err != nil {
		return nil, err
	}

	configuredByID := make(map[string]*ent.Model, len(configuredModels))
	for _, configuredModel := range configuredModels {
		configuredByID[configuredModel.ModelID] = configuredModel
	}

	preferPassThrough, err := svc.systemService.PreferPassThrough(previewCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pass-through preference: %w", err)
	}
	globalPassThrough, err := svc.systemService.PassThrough(previewCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global pass-through setting: %w", err)
	}

	previewModels := make([]*APIKeyProfilePreviewModel, 0, len(visibleModels))
	relevantChannels := make(map[int]*Channel)
	settings := svc.modelSettingsOrDefault(previewCtx)
	for _, visibleModel := range visibleModels {
		modelChannels := previewChannelsForModel(settings, visibleModel.ID, configuredByID[visibleModel.ID], channels)
		previewChannels := make([]*APIKeyProfilePreviewChannel, 0, len(modelChannels))
		for _, channel := range modelChannels {
			relevantChannels[channel.ID] = channel
			passThroughEnabled := globalPassThrough
			if channel.Settings != nil && channel.Settings.PassThroughBody != nil {
				passThroughEnabled = *channel.Settings.PassThroughBody
			}
			previewChannels = append(previewChannels, &APIKeyProfilePreviewChannel{
				ID:                    channel.ID,
				Name:                  channel.Name,
				OrderingWeight:        channel.OrderingWeight,
				PassThroughAPIFormats: previewPassThroughAPIFormats(channel, passThroughEnabled),
			})
		}

		sort.Slice(previewChannels, func(i, j int) bool {
			if previewChannels[i].OrderingWeight != previewChannels[j].OrderingWeight {
				return previewChannels[i].OrderingWeight > previewChannels[j].OrderingWeight
			}
			if previewChannels[i].Name == previewChannels[j].Name {
				return previewChannels[i].ID < previewChannels[j].ID
			}
			return previewChannels[i].Name < previewChannels[j].Name
		})
		previewModels = append(previewModels, &APIKeyProfilePreviewModel{ID: visibleModel.ID, Channels: previewChannels})
	}

	sort.Slice(previewModels, func(i, j int) bool { return previewModels[i].ID < previewModels[j].ID })

	apiFormatSet := make(map[string]struct{})
	for _, channel := range relevantChannels {
		for _, endpoint := range channel.ResolveEndpoints() {
			if endpoint.APIFormat != "" {
				apiFormatSet[endpoint.APIFormat] = struct{}{}
			}
		}
	}

	apiFormats := make([]string, 0, len(apiFormatSet))
	for apiFormat := range apiFormatSet {
		apiFormats = append(apiFormats, apiFormat)
	}
	sort.Slice(apiFormats, func(i, j int) bool {
		iPriority := previewAPIFormatPriority(apiFormats[i])
		jPriority := previewAPIFormatPriority(apiFormats[j])
		if iPriority != jPriority {
			return iPriority < jPriority
		}

		return apiFormats[i] < apiFormats[j]
	})

	return &APIKeyProfilePreview{
		APIFormats:        apiFormats,
		Models:            previewModels,
		PreferPassThrough: preferPassThrough,
	}, nil
}

func previewPassThroughAPIFormats(channel *Channel, enabled bool) []string {
	if !enabled {
		return []string{}
	}

	formats := make([]string, 0)
	seen := make(map[string]struct{})
	for _, endpoint := range channel.ResolveEndpoints() {
		if endpoint.APIFormat == "" || !previewPassThroughBodySupported(endpoint.APIFormat) {
			continue
		}
		if _, ok := seen[endpoint.APIFormat]; ok {
			continue
		}

		seen[endpoint.APIFormat] = struct{}{}
		formats = append(formats, endpoint.APIFormat)
	}

	return formats
}

func previewPassThroughBodySupported(apiFormat string) bool {
	switch apiFormat {
	case "openai/audio_transcriptions", "openai/audio_translations", "openai/image_edit", "openai/image_variation":
		return false
	default:
		return true
	}
}

func previewAPIFormatPriority(apiFormat string) int {
	switch apiFormat {
	case "openai/chat_completions":
		return 0
	case "openai/responses":
		return 1
	case "anthropic/messages":
		return 2
	case "gemini/contents":
		return 3
	default:
		return 100
	}
}

func (svc *ModelService) previewConfiguredModels(ctx context.Context, visibleModels []ModelFacade) ([]*ent.Model, error) {
	if len(visibleModels) == 0 {
		return nil, nil
	}

	modelIDs := make([]string, 0, len(visibleModels))
	for _, visibleModel := range visibleModels {
		modelIDs = append(modelIDs, visibleModel.ID)
	}

	models, err := svc.entFromContext(ctx).Model.Query().
		Where(model.StatusEQ(model.StatusEnabled), model.ModelIDIn(modelIDs...)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query configured models for preview: %w", err)
	}

	return models, nil
}

func previewChannelsForModel(
	settings *SystemModelSettings,
	modelID string,
	configuredModel *ent.Model,
	channels []*Channel,
) []*Channel {
	channelByID := make(map[int]*Channel, len(channels))
	for _, channel := range channels {
		channelByID[channel.ID] = channel
	}

	matched := make(map[int]*Channel)
	if configuredModel != nil {
		connections := MatchConnections(EffectiveModelAssociations(settings, configuredModel), channels)
		for _, connection := range connections {
			if channel := channelByID[connection.Channel.ID]; channel != nil {
				matched[channel.ID] = channel
			}
		}
	} else {
		for _, channel := range channels {
			if channel.IsModelSupported(modelID) {
				matched[channel.ID] = channel
			}
		}
	}

	result := make([]*Channel, 0, len(matched))
	for _, channel := range matched {
		result = append(result, channel)
	}

	return result
}
