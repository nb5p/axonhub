package biz

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
)

func TestModelService_PreviewAPIKeyProfile(t *testing.T) {
	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&_fk=0")
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	openAIChannel := createPreviewChannel(t, client, ctx, channel.TypeOpenai, "OpenAI Primary", 100, []string{"shared", "openai-only"}, []string{"preview"})
	anthropicChannel := createPreviewChannel(t, client, ctx, channel.TypeAnthropic, "Anthropic Backup", 20, []string{"shared", "anthropic-only"}, []string{"preview"})
	_ = createPreviewChannel(t, client, ctx, channel.TypeOpenai, "Excluded", 50, []string{"shared"}, []string{"other"})

	channelSvc := NewChannelServiceForTest(client)
	enabledEntities, err := client.Channel.Query().Where(channel.StatusEQ(channel.StatusEnabled)).All(ctx)
	require.NoError(t, err)

	enabledChannels := make([]*Channel, 0, len(enabledEntities))
	for _, entity := range enabledEntities {
		built, buildErr := channelSvc.buildChannelWithTransformer(entity)
		require.NoError(t, buildErr)
		enabledChannels = append(enabledChannels, built)
	}
	channelSvc.SetEnabledChannelsForTest(enabledChannels)

	systemSvc := &SystemService{
		AbstractService: &AbstractService{db: client},
		Cache:           xcache.NewFromConfig[ent.System](xcache.Config{Mode: xcache.ModeMemory}),
	}
	require.NoError(t, systemSvc.SetModelSettings(ctx, SystemModelSettings{QueryAllChannelModels: true}))
	require.NoError(t, systemSvc.SetPassThrough(ctx, true))
	require.NoError(t, systemSvc.SetPreferPassThrough(ctx, true))

	modelSvc := &ModelService{
		AbstractService: &AbstractService{db: client},
		channelService:  channelSvc,
		systemService:   systemSvc,
	}
	apiKey := &ent.APIKey{
		ID:   1,
		Name: "preview-key",
		Edges: ent.APIKeyEdges{
			Project: &ent.Project{},
		},
	}

	preview, err := modelSvc.PreviewAPIKeyProfile(ctx, apiKey, objects.APIKeyProfile{
		Name:                 "unsaved",
		ChannelTags:          []string{"preview"},
		ChannelTagsMatchMode: objects.ChannelTagsMatchModeAny,
		ModelIDs:             []string{"shared"},
	})
	require.NoError(t, err)
	require.Len(t, preview.Models, 1)
	require.Len(t, preview.Models[0].Channels, 2)
	require.Equal(t, []string{"shared"}, []string{preview.Models[0].ID})
	require.True(t, preview.PreferPassThrough)
	require.Equal(t, []string{"OpenAI Primary", "Anthropic Backup"}, []string{
		preview.Models[0].Channels[0].Name,
		preview.Models[0].Channels[1].Name,
	})
	require.Equal(t, 100, preview.Models[0].Channels[0].OrderingWeight)
	require.Contains(t, preview.Models[0].Channels[0].PassThroughAPIFormats, "openai/chat_completions")
	require.Contains(t, preview.Models[0].Channels[1].PassThroughAPIFormats, "anthropic/messages")
	require.Equal(t, "openai/chat_completions", preview.APIFormats[0])
	require.Contains(t, preview.APIFormats, "openai/chat_completions")
	require.Contains(t, preview.APIFormats, "anthropic/messages")

	restrictedPreview, err := modelSvc.PreviewAPIKeyProfile(ctx, apiKey, objects.APIKeyProfile{
		Name:       "unsaved",
		ChannelIDs: []int{openAIChannel.ID},
		ModelIDs:   []string{"openai-only"},
	})
	require.NoError(t, err)
	require.Len(t, restrictedPreview.Models, 1)
	require.Equal(t, "openai-only", restrictedPreview.Models[0].ID)
	require.Equal(t, openAIChannel.ID, restrictedPreview.Models[0].Channels[0].ID)
	require.NotEqual(t, anthropicChannel.ID, restrictedPreview.Models[0].Channels[0].ID)
}

func createPreviewChannel(
	t *testing.T,
	client *ent.Client,
	ctx context.Context,
	channelType channel.Type,
	name string,
	orderingWeight int,
	models []string,
	tags []string,
) *ent.Channel {
	t.Helper()

	result, err := client.Channel.Create().
		SetType(channelType).
		SetName(name).
		SetBaseURL("https://example.com").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key"}).
		SetSupportedModels(models).
		SetDefaultTestModel(models[0]).
		SetOrderingWeight(orderingWeight).
		SetStatus(channel.StatusEnabled).
		SetTags(tags).
		Save(ctx)
	require.NoError(t, err)

	return result
}
