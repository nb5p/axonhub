package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	providerquota "github.com/looplj/axonhub/internal/server/biz/provider_quota"
)

type ChannelUsageQueryConfig struct {
	Enabled             bool
	ShowInProviderQuota bool
	Preset              objects.ChannelUsageQueryPreset
	BaseURLOverride     *string
	UserID              *string
	Script              string
	APIKeyConfigured    bool
}

type ChannelUsageQueryConfigInput struct {
	Enabled             bool
	ShowInProviderQuota bool
	Preset              objects.ChannelUsageQueryPreset
	BaseURLOverride     *string
	UserID              *string
	Script              string
	APIKey              *string
	ClearAPIKey         bool
}

type ChannelUsageQueryTestResult struct {
	Status   string
	Balance  *providerquota.UsageQueryBalance
	Text     *string
	Progress *providerquota.UsageQueryProgress
}

func (svc *ChannelService) ChannelUsageQueryConfig(ctx context.Context, channelID int) (*ChannelUsageQueryConfig, error) {
	ch, err := svc.entFromContext(ctx).Channel.Get(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel usage query: %w", err)
	}
	return channelUsageQueryConfigFromEntity(ch), nil
}

func (svc *ChannelService) SaveChannelUsageQueryConfig(
	ctx context.Context,
	channelID int,
	input ChannelUsageQueryConfigInput,
) (*ChannelUsageQueryConfig, error) {
	input = canonicalizeUsageQueryPresetInput(input)
	if err := validateChannelUsageQueryInput(input); err != nil {
		return nil, err
	}

	var updated *ent.Channel
	err := svc.RunInTransaction(ctx, func(ctx context.Context) error {
		db := svc.entFromContext(ctx)
		ch, err := db.Channel.Get(ctx, channelID)
		if err != nil {
			return fmt.Errorf("failed to load channel usage query: %w", err)
		}

		settings := objects.ChannelSettings{}
		if ch.Settings != nil {
			settings = *ch.Settings
		}
		settings.UsageQuery = usageQuerySettingsFromInput(input)

		credentials := ch.Credentials
		if input.ClearAPIKey {
			credentials.UsageQueryAPIKey = ""
		} else if input.APIKey != nil && strings.TrimSpace(*input.APIKey) != "" {
			credentials.UsageQueryAPIKey = strings.TrimSpace(*input.APIKey)
		}

		updated, err = db.Channel.UpdateOne(ch).
			SetSettings(&settings).
			SetCredentials(credentials).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to save channel usage query: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if ent.TxFromContext(ctx) == nil {
		updated.Unwrap()
	}

	runAfterCommit(ctx, func(ctx context.Context) {
		svc.invalidateProviderQuota(ctx, channelID)
	})
	svc.reloadChannelsAfterCommit(ctx)

	return channelUsageQueryConfigFromEntity(updated), nil
}

func (svc *ChannelService) TestChannelUsageQuery(
	ctx context.Context,
	channelID int,
	input ChannelUsageQueryConfigInput,
) (*ChannelUsageQueryTestResult, error) {
	input = canonicalizeUsageQueryPresetInput(input)
	if err := validateChannelUsageQueryInput(input); err != nil {
		return nil, err
	}

	ch, err := svc.entFromContext(ctx).Channel.Get(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel usage query: %w", err)
	}
	clone := *ch
	settings := objects.ChannelSettings{}
	if ch.Settings != nil {
		settings = *ch.Settings
	}
	settings.UsageQuery = usageQuerySettingsFromInput(input)
	settings.UsageQuery.Enabled = true
	clone.Settings = &settings
	clone.Credentials = ch.Credentials
	if input.ClearAPIKey {
		clone.Credentials.UsageQueryAPIKey = ""
	} else if input.APIKey != nil && strings.TrimSpace(*input.APIKey) != "" {
		clone.Credentials.UsageQueryAPIKey = strings.TrimSpace(*input.APIKey)
	}

	quotaData, err := providerquota.NewUsageQueryChecker(svc.httpClient).CheckQuota(ctx, &clone)
	if err != nil {
		return nil, err
	}
	return usageQueryTestResultFromQuotaData(quotaData), nil
}

func validateChannelUsageQueryInput(input ChannelUsageQueryConfigInput) error {
	switch input.Preset {
	case objects.ChannelUsageQueryPresetNewAPI,
		objects.ChannelUsageQueryPresetCodex,
		objects.ChannelUsageQueryPresetClaude,
		objects.ChannelUsageQueryPresetOpenCode,
		objects.ChannelUsageQueryPresetCustom:
	default:
		return fmt.Errorf("unsupported usage query preset: %q", input.Preset)
	}
	if strings.TrimSpace(input.Script) == "" {
		return errors.New("usage query script is required")
	}
	if input.Preset == objects.ChannelUsageQueryPresetNewAPI && (input.UserID == nil || strings.TrimSpace(*input.UserID) == "") {
		return errors.New("New API usage query requires a user ID")
	}
	if input.ClearAPIKey && input.APIKey != nil && strings.TrimSpace(*input.APIKey) != "" {
		return errors.New("cannot set and clear the usage query API key at the same time")
	}

	baseURL := "https://usage-query.invalid"
	if input.BaseURLOverride != nil && strings.TrimSpace(*input.BaseURLOverride) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(*input.BaseURLOverride), "/")
	}
	_, err := providerquota.NewGojaUsageQueryRuntime().ParseRequest(context.Background(), input.Script, map[string]string{
		"baseUrl":     baseURL,
		"apiKey":      "validation-key",
		"accessToken": "validation-key",
		"userId":      "validation-user",
	})
	if err != nil {
		return fmt.Errorf("invalid usage query script: %w", err)
	}
	return nil
}

func usageQuerySettingsFromInput(input ChannelUsageQueryConfigInput) *objects.ChannelUsageQuerySettings {
	settings := &objects.ChannelUsageQuerySettings{
		Enabled:             input.Enabled,
		ShowInProviderQuota: lo.ToPtr(input.ShowInProviderQuota),
		Preset:              input.Preset,
		Script:              strings.TrimSpace(input.Script),
	}
	if input.BaseURLOverride != nil {
		settings.BaseURLOverride = strings.TrimSpace(*input.BaseURLOverride)
	}
	if input.UserID != nil {
		settings.UserID = strings.TrimSpace(*input.UserID)
	}
	return settings
}

func canonicalizeUsageQueryPresetInput(input ChannelUsageQueryConfigInput) ChannelUsageQueryConfigInput {
	if script, ok := providerquota.UsageQueryPresetScript(input.Preset); ok {
		input.Script = script
	}
	return input
}

func channelUsageQueryConfigFromEntity(ch *ent.Channel) *ChannelUsageQueryConfig {
	result := &ChannelUsageQueryConfig{
		Preset:              objects.ChannelUsageQueryPresetNewAPI,
		ShowInProviderQuota: true,
		APIKeyConfigured:    strings.TrimSpace(ch.Credentials.UsageQueryAPIKey) != "",
	}
	var settings *objects.ChannelUsageQuerySettings
	if ch.Settings != nil {
		settings = ch.Settings.UsageQuery
	}
	if settings == nil {
		settings = providerquota.BuiltInUsageQuerySettings(ch)
	}
	if settings == nil {
		return result
	}

	result.Enabled = settings.Enabled
	if settings.ShowInProviderQuota != nil {
		result.ShowInProviderQuota = *settings.ShowInProviderQuota
	}
	result.Preset = settings.Preset
	result.Script = providerquota.EffectiveUsageQueryScript(settings)
	if settings.BaseURLOverride != "" {
		result.BaseURLOverride = &settings.BaseURLOverride
	}
	if settings.UserID != "" {
		result.UserID = &settings.UserID
	}
	return result
}

func usageQueryTestResultFromQuotaData(quotaData providerquota.QuotaData) *ChannelUsageQueryTestResult {
	result := &ChannelUsageQueryTestResult{Status: quotaData.Status}
	data := quotaData.RawData
	if balance, ok := data["balance"].(*providerquota.UsageQueryBalance); ok {
		result.Balance = balance
	}
	if value, ok := data["text"].(string); ok {
		result.Text = &value
	}
	if progress, ok := usageQueryProgressFromRawData(data["progress"]); ok {
		result.Progress = progress
	}
	return result
}

func usageQueryProgressFromRawData(value any) (*providerquota.UsageQueryProgress, bool) {
	if value == nil {
		return nil, false
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, false
	}
	var progress providerquota.UsageQueryProgress
	if err := json.Unmarshal(encoded, &progress); err != nil || len(progress.Windows) == 0 {
		return nil, false
	}
	return &progress, true
}

func hasEnabledUsageQuery(ch *ent.Channel) bool {
	return ch != nil && ch.Settings != nil && ch.Settings.UsageQuery != nil && ch.Settings.UsageQuery.Enabled
}
