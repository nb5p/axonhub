package biz

import (
	"context"
	"slices"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/looplj/axonhub/internal/ent"
)

// QueryChannelsInput represents the input for querying channels with additional filters.
type QueryChannelsInput struct {
	After           *entgql.Cursor[int]
	First           *int
	Before          *entgql.Cursor[int]
	Last            *int
	OrderBy         *ent.ChannelOrder
	Where           *ent.ChannelWhereInput
	HasTag          *string
	Model           *string
	Models          []string
	ModelsMatchMode ChannelModelsMatchMode
	EndpointFormats []string
}

// ChannelModelsMatchMode controls whether any or all selected models must be
// supported by a channel.
type ChannelModelsMatchMode string

const (
	ChannelModelsMatchModeAny ChannelModelsMatchMode = "any"
	ChannelModelsMatchModeAll ChannelModelsMatchMode = "all"
)

func (m ChannelModelsMatchMode) OrDefault() ChannelModelsMatchMode {
	if m == ChannelModelsMatchModeAll {
		return ChannelModelsMatchModeAll
	}

	return ChannelModelsMatchModeAny
}

// QueryChannels queries channels with the specified input parameters, including
// model and effective endpoint filtering.
func (svc *ChannelService) QueryChannels(ctx context.Context, input QueryChannelsInput) (*ent.ChannelConnection, error) {
	// Build the base query
	var (
		query = svc.entFromContext(ctx).Channel.Query()
		err   error
	)

	// Apply standard filters
	if input.Where != nil {
		query, err = input.Where.Filter(query)
		if err != nil {
			return nil, err
		}
	}

	if input.HasTag != nil && *input.HasTag != "" {
		query = query.Where(func(s *sql.Selector) {
			s.Where(sqljson.ValueContains("tags", *input.HasTag))
		})
	}

	models := append([]string(nil), input.Models...)
	if input.Model != nil && *input.Model != "" && !slices.Contains(models, *input.Model) {
		models = append(models, *input.Model)
	}

	// If no advanced filter is specified, return the paginated query directly.
	if len(models) == 0 && len(input.EndpointFormats) == 0 {
		return query.Paginate(ctx, input.After, input.First, input.Before, input.Last,
			ent.WithChannelOrder(input.OrderBy),
		)
	}

	// Effective models and endpoints are resolved by the business layer, so fetch
	// the standard-filter result set and apply these filters in memory.
	return svc.queryChannelsWithAdvancedFilters(ctx, query, input, models)
}

// queryChannelsWithAdvancedFilters performs resolved capability filtering
// without pagination.
func (svc *ChannelService) queryChannelsWithAdvancedFilters(
	ctx context.Context,
	query *ent.ChannelQuery,
	input QueryChannelsInput,
	models []string,
) (*ent.ChannelConnection, error) {
	// Fetch all channels from the database
	if input.OrderBy != nil {
		query = query.Order(input.OrderBy.ToOrderOption())
	}

	channels, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	// Filter channels by resolved model and endpoint support.
	var filteredChannels []*ent.Channel

	for _, channel := range channels {
		channelObj := Channel{Channel: channel}
		if len(models) > 0 && !channelMatchesModels(channelObj, models, input.ModelsMatchMode) {
			continue
		}
		if len(input.EndpointFormats) > 0 && !channelMatchesAnyEndpointFormat(channelObj, input.EndpointFormats) {
			continue
		}

		filteredChannels = append(filteredChannels, channel)
	}

	// Build connection without pagination (ignore pagination for resolved capability filters).
	return svc.buildConnectionInMemory(filteredChannels, input.OrderBy), nil
}

func channelMatchesAnyEndpointFormat(channel Channel, endpointFormats []string) bool {
	wanted := make(map[string]struct{}, len(endpointFormats))
	for _, apiFormat := range endpointFormats {
		if apiFormat != "" {
			wanted[apiFormat] = struct{}{}
		}
	}

	for _, endpoint := range channel.ResolveEndpoints() {
		if _, ok := wanted[endpoint.APIFormat]; ok {
			return true
		}
	}

	return false
}

func channelMatchesModels(channel Channel, models []string, matchMode ChannelModelsMatchMode) bool {
	if matchMode.OrDefault() == ChannelModelsMatchModeAll {
		for _, model := range models {
			if !channel.IsModelSupported(model) {
				return false
			}
		}

		return true
	}

	for _, model := range models {
		if channel.IsModelSupported(model) {
			return true
		}
	}

	return false
}

// buildConnectionInMemory builds a relay-style connection from filtered channels.
func (svc *ChannelService) buildConnectionInMemory(
	channels []*ent.Channel,
	order *ent.ChannelOrder,
) *ent.ChannelConnection {
	conn := &ent.ChannelConnection{
		Edges:    []*ent.ChannelEdge{},
		PageInfo: ent.PageInfo{},
	}

	// Handle empty result
	if len(channels) == 0 {
		conn.TotalCount = 0
		return conn
	}

	// Return all channels without pagination
	conn.Edges = make([]*ent.ChannelEdge, len(channels))
	for i, ch := range channels {
		conn.Edges[i] = ch.ToEdge(order)
	}

	conn.PageInfo.HasNextPage = false
	conn.PageInfo.HasPreviousPage = false

	conn.TotalCount = len(channels)
	if len(conn.Edges) > 0 {
		conn.PageInfo.StartCursor = &conn.Edges[0].Cursor
		conn.PageInfo.EndCursor = &conn.Edges[len(conn.Edges)-1].Cursor
	}

	return conn
}
