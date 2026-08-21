package biz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
)

const channelUsageQueryTestScript = `({
  request: {
    url: "{{baseUrl}}/balance",
    method: "GET",
    headers: { "Authorization": "Bearer {{apiKey}}" }
  },
  extractor: function(response) {
    return { remaining: response.balance, unit: "USD" };
  }
})`

func TestChannelUsageQueryConfig_SaveAndRegularUpdatePreserveSecret(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Usage Query Channel").
		SetBaseURL("https://api.example.com/v1").
		SetCredentials(objects.ChannelCredentials{
			APIKey:           "request-key",
			UsageQueryAPIKey: "query-key",
		}).
		SetSettings(&objects.ChannelSettings{ExtraModelPrefix: "old-prefix"}).
		SetSupportedModels([]string{"test-model"}).
		SetDefaultTestModel("test-model").
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)

	config, err := svc.SaveChannelUsageQueryConfig(ctx, ch.ID, ChannelUsageQueryConfigInput{
		Enabled:             true,
		ShowInProviderQuota: false,
		Preset:              objects.ChannelUsageQueryPresetCustom,
		Script:              channelUsageQueryTestScript,
	})
	require.NoError(t, err)
	require.True(t, config.Enabled)
	require.False(t, config.ShowInProviderQuota)
	require.True(t, config.APIKeyConfigured)
	require.Equal(t, channelUsageQueryTestScript, config.Script)

	config, err = svc.ChannelUsageQueryConfig(ctx, ch.ID)
	require.NoError(t, err)
	require.True(t, config.APIKeyConfigured)
	require.False(t, config.ShowInProviderQuota)
	require.NotContains(t, config.Script, "query-key")

	updated, err := svc.UpdateChannel(ctx, ch.ID, &ent.UpdateChannelInput{
		Credentials: &objects.ChannelCredentials{APIKey: "new-request-key"},
		Settings:    &objects.ChannelSettings{ExtraModelPrefix: "new-prefix"},
	})
	require.NoError(t, err)
	require.Equal(t, "new-request-key", updated.Credentials.APIKey)
	require.Equal(t, "query-key", updated.Credentials.UsageQueryAPIKey)
	require.NotNil(t, updated.Settings.UsageQuery)
	require.Equal(t, channelUsageQueryTestScript, updated.Settings.UsageQuery.Script)
	require.NotNil(t, updated.Settings.UsageQuery.ShowInProviderQuota)
	require.False(t, *updated.Settings.UsageQuery.ShowInProviderQuota)
	require.Equal(t, "new-prefix", updated.Settings.ExtraModelPrefix)
}

func TestChannelUsageQueryConfig_TestUsesTemporaryKeyWithoutPersisting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer temporary-query-key", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"balance":12.5}`))
	}))
	defer server.Close()

	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	ch, err := client.Channel.Create().
		SetType(channel.TypeOpenai).
		SetName("Usage Query Test Channel").
		SetBaseURL(server.URL).
		SetCredentials(objects.ChannelCredentials{
			APIKey:           "request-key",
			UsageQueryAPIKey: "saved-query-key",
		}).
		SetSupportedModels([]string{"test-model"}).
		SetDefaultTestModel("test-model").
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)

	result, err := svc.TestChannelUsageQuery(ctx, ch.ID, ChannelUsageQueryConfigInput{
		Enabled:         false,
		Preset:          objects.ChannelUsageQueryPresetCustom,
		BaseURLOverride: lo.ToPtr(server.URL),
		Script:          channelUsageQueryTestScript,
		APIKey:          lo.ToPtr("temporary-query-key"),
	})
	require.NoError(t, err)
	require.Equal(t, "available", result.Status)
	require.NotNil(t, result.Balance)
	require.Equal(t, 12.5, result.Balance.Remaining)

	stored, err := client.Channel.Get(ctx, ch.ID)
	require.NoError(t, err)
	require.Equal(t, "saved-query-key", stored.Credentials.UsageQueryAPIKey)
	require.NotNil(t, stored.Settings)
	require.Nil(t, stored.Settings.UsageQuery)
}

func TestChannelUsageQueryConfig_PresetScriptIsServerOwned(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()

	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	ch, err := client.Channel.Create().
		SetType(channel.TypeCodex).
		SetName("Codex usage query").
		SetBaseURL("https://chatgpt.com").
		SetCredentials(objects.ChannelCredentials{APIKey: "{\"access_token\":\"test-token\"}"}).
		SetSupportedModels([]string{"test-model"}).
		SetDefaultTestModel("test-model").
		SetStatus(channel.StatusEnabled).
		Save(ctx)
	require.NoError(t, err)

	config, err := svc.SaveChannelUsageQueryConfig(ctx, ch.ID, ChannelUsageQueryConfigInput{
		Enabled: true,
		Preset:  objects.ChannelUsageQueryPresetCodex,
		Script:  "not the preset script",
	})
	require.NoError(t, err)
	require.Equal(t, objects.ChannelUsageQueryPresetCodex, config.Preset)
	require.Contains(t, config.Script, "responseVersion: 2")
	require.NotContains(t, config.Script, "not the preset script")
}
