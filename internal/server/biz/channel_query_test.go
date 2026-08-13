package biz

import (
	"context"
	"strconv"
	"testing"

	"entgo.io/contrib/entgql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
)

func TestChannelService_QueryChannels_EnabledFirst(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))

	disabledHighest := createOrderedTestChannel(t, client, ctx, "Disabled 100", channel.StatusDisabled, 100)
	enabledHigh := createOrderedTestChannel(t, client, ctx, "Enabled 50", channel.StatusEnabled, 50)
	disabledLow := createOrderedTestChannel(t, client, ctx, "Disabled 20", channel.StatusDisabled, 20)
	enabledLow := createOrderedTestChannel(t, client, ctx, "Enabled 10", channel.StatusEnabled, 10)

	weightDescending := &ent.ChannelOrder{
		Direction: entgql.OrderDirectionDesc,
		Field:     ent.ChannelOrderFieldOrderingWeight,
	}

	conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
		First:        lo.ToPtr(2),
		OrderBy:      weightDescending,
		EnabledFirst: true,
	})
	require.NoError(t, err)
	require.Equal(t, []int{enabledHigh.ID, enabledLow.ID}, channelConnectionIDs(conn))
	require.True(t, conn.PageInfo.HasNextPage)
	require.False(t, conn.PageInfo.HasPreviousPage)
	require.Equal(t, 4, conn.TotalCount)

	conn, err = svc.QueryChannels(ctx, QueryChannelsInput{
		After:        conn.PageInfo.EndCursor,
		First:        lo.ToPtr(2),
		OrderBy:      weightDescending,
		EnabledFirst: true,
	})
	require.NoError(t, err)
	require.Equal(t, []int{disabledHighest.ID, disabledLow.ID}, channelConnectionIDs(conn))
	require.False(t, conn.PageInfo.HasNextPage)
	require.True(t, conn.PageInfo.HasPreviousPage)

	conn, err = svc.QueryChannels(ctx, QueryChannelsInput{
		OrderBy: weightDescending,
	})
	require.NoError(t, err)
	require.Equal(t, []int{disabledHighest.ID, enabledHigh.ID, disabledLow.ID, enabledLow.ID}, channelConnectionIDs(conn))

	conn, err = svc.QueryChannels(ctx, QueryChannelsInput{
		Model:        lo.ToPtr("gpt-4"),
		OrderBy:      weightDescending,
		EnabledFirst: true,
	})
	require.NoError(t, err)
	require.Equal(t, []int{enabledHigh.ID, enabledLow.ID, disabledHighest.ID, disabledLow.ID}, channelConnectionIDs(conn))
}

func channelConnectionIDs(conn *ent.ChannelConnection) []int {
	ids := make([]int, len(conn.Edges))
	for i, edge := range conn.Edges {
		ids[i] = edge.Node.ID
	}
	return ids
}

func TestChannelService_QueryChannels_WithModelFilter(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test channels with different models
	channels := []*ent.Channel{
		createTestChannel(t, client, ctx, "Channel 1", []string{"gpt-4", "gpt-3.5-turbo"}, nil),
		createTestChannel(t, client, ctx, "Channel 2", []string{"gpt-4"}, nil),
		createTestChannel(t, client, ctx, "Channel 3", []string{"claude-3-opus"}, nil),
		createTestChannel(t, client, ctx, "Channel 4", []string{"gpt-4", "claude-3-opus"}, nil),
		createTestChannel(t, client, ctx, "Channel 5", []string{"gpt-3.5-turbo"}, nil),
		createTestChannel(t, client, ctx, "Channel 6", []string{"gpt-4"}, nil),
	}

	tests := []struct {
		name              string
		input             QueryChannelsInput
		expectedIDs       []int
		expectedHasNext   bool
		expectedHasPrev   bool
		expectedEdgeCount int
	}{
		{
			name: "filter by gpt-4 model - should return all without pagination",
			input: QueryChannelsInput{
				Model: lo.ToPtr("gpt-4"),
				// First should be ignored when model is specified
				First: lo.ToPtr(2),
			},
			expectedIDs:       []int{channels[0].ID, channels[1].ID, channels[3].ID, channels[5].ID},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 4,
		},
		{
			name: "filter by gpt-4 model - all",
			input: QueryChannelsInput{
				Model: lo.ToPtr("gpt-4"),
			},
			expectedIDs:       []int{channels[0].ID, channels[1].ID, channels[3].ID, channels[5].ID},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 4,
		},
		{
			name: "filter by claude-3-opus - should return all without pagination",
			input: QueryChannelsInput{
				Model: lo.ToPtr("claude-3-opus"),
				First: lo.ToPtr(2), // Should be ignored
			},
			expectedIDs:       []int{channels[2].ID, channels[3].ID},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 2,
		},
		{
			name: "filter by gpt-3.5-turbo",
			input: QueryChannelsInput{
				Model: lo.ToPtr("gpt-3.5-turbo"),
			},
			expectedIDs:       []int{channels[0].ID, channels[4].ID},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 2,
		},
		{
			name: "filter by non-existent model",
			input: QueryChannelsInput{
				Model: lo.ToPtr("non-existent-model"),
			},
			expectedIDs:       []int{},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 0,
		},
		{
			name: "filter by gpt-4 with Last parameter - should ignore pagination",
			input: QueryChannelsInput{
				Model: lo.ToPtr("gpt-4"),
				Last:  lo.ToPtr(1), // Should be ignored
			},
			expectedIDs:       []int{channels[0].ID, channels[1].ID, channels[3].ID, channels[5].ID},
			expectedHasNext:   false,
			expectedHasPrev:   false,
			expectedEdgeCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := svc.QueryChannels(ctx, tt.input)

			require.NoError(t, err)
			require.NotNil(t, conn)
			require.Len(t, conn.Edges, tt.expectedEdgeCount)
			require.Equal(t, tt.expectedHasNext, conn.PageInfo.HasNextPage)
			require.Equal(t, tt.expectedHasPrev, conn.PageInfo.HasPreviousPage)

			// Verify returned channel IDs
			actualIDs := make([]int, len(conn.Edges))
			for i, edge := range conn.Edges {
				actualIDs[i] = edge.Node.ID
			}

			require.ElementsMatch(t, tt.expectedIDs, actualIDs)

			// Verify cursors are set when there are results
			if len(conn.Edges) > 0 {
				require.NotNil(t, conn.PageInfo.StartCursor)
				require.NotNil(t, conn.PageInfo.EndCursor)
			}
		})
	}
}

func TestChannelService_QueryChannels_WithMultipleModelFilters(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	channels := []*ent.Channel{
		createTestChannel(t, client, ctx, "Both", []string{"gpt-4", "claude-3-opus"}, nil),
		createTestChannel(t, client, ctx, "GPT only", []string{"gpt-4"}, nil),
		createTestChannel(t, client, ctx, "Claude only", []string{"claude-3-opus"}, nil),
		createTestChannel(t, client, ctx, "Other", []string{"gemini-pro"}, nil),
	}

	tests := []struct {
		name        string
		matchMode   ChannelModelsMatchMode
		expectedIDs []int
	}{
		{
			name:        "defaults to matching any selected model",
			expectedIDs: []int{channels[0].ID, channels[1].ID, channels[2].ID},
		},
		{
			name:        "matches every selected model",
			matchMode:   ChannelModelsMatchModeAll,
			expectedIDs: []int{channels[0].ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
				Models:          []string{"gpt-4", "claude-3-opus"},
				ModelsMatchMode: tt.matchMode,
				First:           lo.ToPtr(1),
			})

			require.NoError(t, err)
			require.Len(t, conn.Edges, len(tt.expectedIDs))
			require.False(t, conn.PageInfo.HasNextPage)

			actualIDs := make([]int, len(conn.Edges))
			for i, edge := range conn.Edges {
				actualIDs[i] = edge.Node.ID
			}
			require.ElementsMatch(t, tt.expectedIDs, actualIDs)
		})
	}
}

func TestChannelService_QueryChannels_WithEndpointFormats(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	channels := []*ent.Channel{
		createTestChannelWithCapabilities(t, client, ctx, "Chat", channel.TypeQiniu, []string{"gpt-4"}, nil),
		createTestChannelWithCapabilities(t, client, ctx, "Responses", channel.TypeOpenaiResponses, []string{"gpt-4"}, nil),
		createTestChannelWithCapabilities(t, client, ctx, "Anthropic", channel.TypeAnthropic, []string{"claude-3-opus"}, nil),
		createTestChannelWithCapabilities(t, client, ctx, "Custom Gemini", channel.TypeQiniu, []string{"gemini-pro"}, []objects.ChannelEndpoint{
			{APIFormat: "gemini/contents"},
		}),
	}

	tests := []struct {
		name            string
		endpointFormats []string
		models          []string
		expectedIDs     []int
	}{
		{
			name:            "matches any selected endpoint format",
			endpointFormats: []string{"openai/responses", "anthropic/messages"},
			expectedIDs:     []int{channels[1].ID, channels[2].ID},
		},
		{
			name:            "includes user-configured endpoints in effective capabilities",
			endpointFormats: []string{"gemini/contents"},
			expectedIDs:     []int{channels[3].ID},
		},
		{
			name:            "combines endpoint and model filters",
			endpointFormats: []string{"openai/chat_completions"},
			models:          []string{"gemini-pro"},
			expectedIDs:     []int{channels[3].ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
				EndpointFormats: tt.endpointFormats,
				Models:          tt.models,
				First:           lo.ToPtr(1),
			})

			require.NoError(t, err)
			require.Len(t, conn.Edges, len(tt.expectedIDs))
			require.False(t, conn.PageInfo.HasNextPage)

			actualIDs := make([]int, len(conn.Edges))
			for i, edge := range conn.Edges {
				actualIDs[i] = edge.Node.ID
			}
			require.ElementsMatch(t, tt.expectedIDs, actualIDs)
		})
	}
}

func TestChannelService_QueryChannels_ModelFilterNoPagination(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create channels for testing
	for i := 1; i <= 10; i++ {
		var models []string
		if i%2 == 0 {
			models = []string{"gpt-4"}
		} else {
			models = []string{"claude-3-opus"}
		}

		_ = createTestChannel(t, client, ctx, "FwdChannel"+strconv.Itoa(i), models, nil)
	}

	t.Run("model filter returns all results ignoring pagination", func(t *testing.T) {
		// Even with First=2, should return all 5 gpt-4 channels
		result, err := svc.QueryChannels(ctx, QueryChannelsInput{
			Model: lo.ToPtr("gpt-4"),
			First: lo.ToPtr(2), // Should be ignored
		})
		require.NoError(t, err)
		require.Len(t, result.Edges, 5) // All 5 gpt-4 channels
		require.False(t, result.PageInfo.HasNextPage)
		require.False(t, result.PageInfo.HasPreviousPage)

		// Even with Last=1, should return all 5 gpt-4 channels
		result2, err := svc.QueryChannels(ctx, QueryChannelsInput{
			Model: lo.ToPtr("gpt-4"),
			Last:  lo.ToPtr(1), // Should be ignored
		})
		require.NoError(t, err)
		require.Len(t, result2.Edges, 5) // All 5 gpt-4 channels
		require.False(t, result2.PageInfo.HasNextPage)
		require.False(t, result2.PageInfo.HasPreviousPage)
	})
}

func TestChannelService_QueryChannels_WithModelMapping(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create channels with model mappings
	settings := &objects.ChannelSettings{
		ModelMappings: []objects.ModelMapping{
			{From: "gpt-4-latest", To: "gpt-4"},
			{From: "gpt-4-turbo", To: "gpt-4"},
		},
	}

	channels := []*ent.Channel{
		createTestChannel(t, client, ctx, "Channel 1", []string{"gpt-4"}, settings),
		createTestChannel(t, client, ctx, "Channel 2", []string{"gpt-3.5-turbo"}, nil),
		createTestChannel(t, client, ctx, "Channel 3", []string{"gpt-4"}, settings),
	}

	tests := []struct {
		name              string
		model             string
		expectedIDs       []int
		expectedEdgeCount int
	}{
		{
			name:              "query by actual model name",
			model:             "gpt-4",
			expectedIDs:       []int{channels[0].ID, channels[2].ID},
			expectedEdgeCount: 2,
		},
		{
			name:              "query by mapped model name",
			model:             "gpt-4-latest",
			expectedIDs:       []int{channels[0].ID, channels[2].ID},
			expectedEdgeCount: 2,
		},
		{
			name:              "query by another mapped model name",
			model:             "gpt-4-turbo",
			expectedIDs:       []int{channels[0].ID, channels[2].ID},
			expectedEdgeCount: 2,
		},
		{
			name:              "query by unmapped model",
			model:             "gpt-3.5-turbo",
			expectedIDs:       []int{channels[1].ID},
			expectedEdgeCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
				Model: &tt.model,
			})

			require.NoError(t, err)
			require.NotNil(t, conn)
			require.Len(t, conn.Edges, tt.expectedEdgeCount)

			// Verify returned channel IDs
			actualIDs := make([]int, len(conn.Edges))
			for i, edge := range conn.Edges {
				actualIDs[i] = edge.Node.ID
			}

			require.ElementsMatch(t, tt.expectedIDs, actualIDs)
		})
	}
}

func TestChannelService_QueryChannels_WithExtraModelPrefix(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create channels with extra model prefix
	settings := &objects.ChannelSettings{
		ExtraModelPrefix: "deepseek",
	}

	channels := []*ent.Channel{
		createTestChannel(t, client, ctx, "DeepSeek Channel", []string{"deepseek-chat", "deepseek-reasoner"}, settings),
		createTestChannel(t, client, ctx, "OpenAI Channel", []string{"gpt-4"}, nil),
	}

	tests := []struct {
		name              string
		model             string
		expectedIDs       []int
		expectedEdgeCount int
	}{
		{
			name:              "query by model without prefix",
			model:             "deepseek-chat",
			expectedIDs:       []int{channels[0].ID},
			expectedEdgeCount: 1,
		},
		{
			name:              "query by model with prefix",
			model:             "deepseek/deepseek-chat",
			expectedIDs:       []int{channels[0].ID},
			expectedEdgeCount: 1,
		},
		{
			name:              "query by model with prefix - reasoner",
			model:             "deepseek/deepseek-reasoner",
			expectedIDs:       []int{channels[0].ID},
			expectedEdgeCount: 1,
		},
		{
			name:              "query by unsupported model with prefix",
			model:             "deepseek/gpt-4",
			expectedIDs:       []int{},
			expectedEdgeCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
				Model: &tt.model,
			})

			require.NoError(t, err)
			require.NotNil(t, conn)
			require.Len(t, conn.Edges, tt.expectedEdgeCount)

			// Verify returned channel IDs
			actualIDs := make([]int, len(conn.Edges))
			for i, edge := range conn.Edges {
				actualIDs[i] = edge.Node.ID
			}

			require.ElementsMatch(t, tt.expectedIDs, actualIDs)
		})
	}
}

func TestChannelService_QueryChannels_WithoutModelFilter(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := context.Background()
	ctx = ent.NewContext(ctx, client)
	ctx = authz.WithTestBypass(ctx)

	// Create test channels
	channels := []*ent.Channel{
		createTestChannel(t, client, ctx, "Channel 1", []string{"gpt-4"}, nil),
		createTestChannel(t, client, ctx, "Channel 2", []string{"claude-3-opus"}, nil),
		createTestChannel(t, client, ctx, "Channel 3", []string{"gpt-3.5-turbo"}, nil),
	}

	t.Run("query without model filter", func(t *testing.T) {
		conn, err := svc.QueryChannels(ctx, QueryChannelsInput{})

		require.NoError(t, err)
		require.NotNil(t, conn)
		require.Len(t, conn.Edges, 3)

		// Verify all channels are returned
		actualIDs := make([]int, len(conn.Edges))
		for i, edge := range conn.Edges {
			actualIDs[i] = edge.Node.ID
		}

		expectedIDs := []int{channels[0].ID, channels[1].ID, channels[2].ID}
		require.ElementsMatch(t, expectedIDs, actualIDs)
	})

	t.Run("query without model filter with pagination", func(t *testing.T) {
		conn, err := svc.QueryChannels(ctx, QueryChannelsInput{
			First: lo.ToPtr(2),
		})

		require.NoError(t, err)
		require.NotNil(t, conn)
		require.Len(t, conn.Edges, 2)
		require.True(t, conn.PageInfo.HasNextPage)
		require.False(t, conn.PageInfo.HasPreviousPage)
	})
}

// Helper function to create test channel.
func createTestChannel(
	t *testing.T,
	client *ent.Client,
	ctx context.Context,
	name string,
	models []string,
	settings *objects.ChannelSettings,
) *ent.Channel {
	t.Helper()

	builder := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName(name).
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key"}).
		SetSupportedModels(models).
		SetDefaultTestModel(models[0]).
		SetStatus(channel.StatusEnabled)

	if settings != nil {
		builder.SetSettings(settings)
	}

	ch, err := builder.Save(ctx)
	require.NoError(t, err)

	return ch
}

func createTestChannelWithCapabilities(
	t *testing.T,
	client *ent.Client,
	ctx context.Context,
	name string,
	channelType channel.Type,
	models []string,
	endpoints []objects.ChannelEndpoint,
) *ent.Channel {
	t.Helper()

	ch, err := client.Channel.Create().
		SetType(channelType).
		SetName(name).
		SetBaseURL("https://example.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key"}).
		SetSupportedModels(models).
		SetDefaultTestModel(models[0]).
		SetStatus(channel.StatusEnabled).
		SetEndpoints(endpoints).
		Save(ctx)
	require.NoError(t, err)

	return ch
}

func createOrderedTestChannel(
	t *testing.T,
	client *ent.Client,
	ctx context.Context,
	name string,
	status channel.Status,
	weight int,
) *ent.Channel {
	t.Helper()

	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName(name).
		SetBaseURL("https://api.openai.com/v1").
		SetCredentials(objects.ChannelCredentials{APIKey: "test-key"}).
		SetSupportedModels([]string{"gpt-4"}).
		SetDefaultTestModel("gpt-4").
		SetStatus(status).
		SetOrderingWeight(weight).
		Save(ctx)
	require.NoError(t, err)

	return ch
}
