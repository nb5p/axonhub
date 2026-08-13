package biz

import (
	"context"
	"fmt"
	"slices"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
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
	// EnabledFirst keeps enabled channels ahead of all other statuses while
	// preserving OrderBy as the secondary ordering rule.
	EnabledFirst bool
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

	if input.EnabledFirst && input.OrderBy == nil {
		input.OrderBy = defaultChannelOrder()
	}

	models := append([]string(nil), input.Models...)
	if input.Model != nil && *input.Model != "" && !slices.Contains(models, *input.Model) {
		models = append(models, *input.Model)
	}

	// If no advanced filter is specified, return the paginated query directly.
	if len(models) == 0 && len(input.EndpointFormats) == 0 {
		if input.EnabledFirst {
			return svc.queryChannelsEnabledFirst(ctx, query, input)
		}
		return query.Paginate(ctx, input.After, input.First, input.Before, input.Last,
			ent.WithChannelOrder(input.OrderBy),
		)
	}

	// Effective models and endpoints are resolved by the business layer, so fetch
	// the standard-filter result set and apply these filters in memory.
	return svc.queryChannelsWithAdvancedFilters(ctx, query, input, models)
}

func defaultChannelOrder() *ent.ChannelOrder {
	return &ent.ChannelOrder{
		Direction: entgql.OrderDirectionDesc,
		Field:     ent.ChannelOrderFieldOrderingWeight,
	}
}

// queryChannelsEnabledFirst applies a status partition before the user-selected
// ordering. Cursor pagination is performed against the resulting global order,
// rather than reordering only the current page in the frontend.
func (svc *ChannelService) queryChannelsEnabledFirst(
	ctx context.Context,
	query *ent.ChannelQuery,
	input QueryChannelsInput,
) (*ent.ChannelConnection, error) {
	if input.First != nil && input.Last != nil {
		return nil, fmt.Errorf("passing both first and last is not supported")
	}
	if input.First != nil && *input.First < 0 {
		return nil, fmt.Errorf("first cannot be less than zero")
	}
	if input.Last != nil && *input.Last < 0 {
		return nil, fmt.Errorf("last cannot be less than zero")
	}
	if err := input.OrderBy.Direction.Validate(); err != nil {
		return nil, err
	}

	directionOption := input.OrderBy.Direction.OrderTermOption()
	orderedIDs, err := query.Clone().Order(
		enabledChannelsFirstOrder(),
		input.OrderBy.ToOrderOption(),
		channel.ByID(directionOption),
	).IDs(ctx)
	if err != nil {
		return nil, err
	}

	start, end := 0, len(orderedIDs)
	if input.After != nil {
		index := indexOfChannelID(orderedIDs, input.After.ID)
		if index < 0 {
			return nil, fmt.Errorf("after cursor channel %d was not found", input.After.ID)
		}
		start = index + 1
	}
	if input.Before != nil {
		index := indexOfChannelID(orderedIDs, input.Before.ID)
		if index < 0 {
			return nil, fmt.Errorf("before cursor channel %d was not found", input.Before.ID)
		}
		end = index
	}
	if start > end {
		start = end
	}

	hasPreviousPage := start > 0
	hasNextPage := end < len(orderedIDs)
	pageIDs := orderedIDs[start:end]
	if input.First != nil && len(pageIDs) > *input.First {
		pageIDs = pageIDs[:*input.First]
		hasNextPage = true
	}
	if input.Last != nil && len(pageIDs) > *input.Last {
		pageIDs = pageIDs[len(pageIDs)-*input.Last:]
		hasPreviousPage = true
	}

	conn := &ent.ChannelConnection{
		Edges:      []*ent.ChannelEdge{},
		TotalCount: len(orderedIDs),
		PageInfo: ent.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: hasPreviousPage,
		},
	}
	if len(pageIDs) == 0 {
		return conn, nil
	}

	nodes, err := query.Clone().Where(channel.IDIn(pageIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	nodesByID := make(map[int]*ent.Channel, len(nodes))
	for _, node := range nodes {
		nodesByID[node.ID] = node
	}

	conn.Edges = make([]*ent.ChannelEdge, 0, len(pageIDs))
	for _, id := range pageIDs {
		node, ok := nodesByID[id]
		if !ok {
			continue
		}
		conn.Edges = append(conn.Edges, node.ToEdge(input.OrderBy))
	}
	if len(conn.Edges) > 0 {
		conn.PageInfo.StartCursor = &conn.Edges[0].Cursor
		conn.PageInfo.EndCursor = &conn.Edges[len(conn.Edges)-1].Cursor
	}

	return conn, nil
}

func enabledChannelsFirstOrder() channel.OrderOption {
	return func(selector *sql.Selector) {
		selector.OrderExprFunc(func(builder *sql.Builder) {
			builder.WriteString("CASE WHEN ").
				Ident(selector.C(channel.FieldStatus)).
				WriteOp(sql.OpEQ).
				WriteString("'enabled'").
				WriteString(" THEN 0 ELSE 1 END")
		})
	}
}

func indexOfChannelID(ids []int, id int) int {
	for i, candidate := range ids {
		if candidate == id {
			return i
		}
	}
	return -1
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
	if input.EnabledFirst {
		filteredChannels = partitionEnabledChannelsFirst(filteredChannels)
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

func partitionEnabledChannelsFirst(channels []*ent.Channel) []*ent.Channel {
	partitioned := make([]*ent.Channel, 0, len(channels))
	for _, candidate := range channels {
		if candidate.Status == channel.StatusEnabled {
			partitioned = append(partitioned, candidate)
		}
	}
	for _, candidate := range channels {
		if candidate.Status != channel.StatusEnabled {
			partitioned = append(partitioned, candidate)
		}
	}
	return partitioned
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
