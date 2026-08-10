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
	APIFormats []string                     `json:"apiFormats"`
	Models     []*APIKeyProfilePreviewModel `json:"models"`
}

type APIKeyProfilePreviewModel struct {
	ID       string                         `json:"id"`
	Channels []*APIKeyProfilePreviewChannel `json:"channels"`
}

type APIKeyProfilePreviewChannel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

	previewModels := make([]*APIKeyProfilePreviewModel, 0, len(visibleModels))
	relevantChannels := make(map[int]*Channel)
	settings := svc.modelSettingsOrDefault(previewCtx)
	for _, visibleModel := range visibleModels {
		modelChannels := previewChannelsForModel(settings, visibleModel.ID, configuredByID[visibleModel.ID], channels)
		previewChannels := make([]*APIKeyProfilePreviewChannel, 0, len(modelChannels))
		for _, channel := range modelChannels {
			relevantChannels[channel.ID] = channel
			previewChannels = append(previewChannels, &APIKeyProfilePreviewChannel{
				ID:   channel.ID,
				Name: channel.Name,
			})
		}

		sort.Slice(previewChannels, func(i, j int) bool {
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
	sort.Strings(apiFormats)

	return &APIKeyProfilePreview{APIFormats: apiFormats, Models: previewModels}, nil
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
