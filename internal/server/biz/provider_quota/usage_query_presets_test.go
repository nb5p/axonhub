package provider_quota

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
)

func TestUsageQueryPresetScripts_ParseAndExtract(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	tests := []struct {
		name     string
		preset   objects.ChannelUsageQueryPreset
		path     string
		response map[string]any
	}{
		{
			name:   "new api",
			preset: objects.ChannelUsageQueryPresetNewAPI,
			path:   "/api/user/self",
			response: map[string]any{
				"success": true,
				"data":    map[string]any{"quota": 1_000_000, "group": "default"},
			},
		},
		{
			name:   "codex",
			preset: objects.ChannelUsageQueryPresetCodex,
			path:   "/backend-api/wham/usage",
			response: map[string]any{
				"plan_type": "plus",
				"rate_limit": map[string]any{
					"primary_window": map[string]any{"used_percent": 20, "limit_window_seconds": 18_000, "reset_at": 1_787_064_000},
				},
			},
		},
		{
			name:     "claude oauth",
			preset:   objects.ChannelUsageQueryPresetClaude,
			path:     "/api/oauth/usage",
			response: map[string]any{"rate_limit_tier": "pro", "five_hour": map[string]any{"utilization": 0.2, "resets_at": "2026-08-19T07:36:46+08:00"}},
		},
		{
			name:   "opencode go",
			preset: objects.ChannelUsageQueryPresetOpenCode,
			path:   "/usage",
			response: map[string]any{
				"usage": map[string]any{
					"rolling": map[string]any{"percent": 20, "resetsAt": "2026-08-19T07:36:46+08:00"},
					"weekly":  map[string]any{"percent": 40, "resetsAt": "2026-08-25T07:36:46+08:00"},
					"monthly": map[string]any{"percent": 10, "resetsAt": "2026-09-17T07:36:46+08:00"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, ok := UsageQueryPresetScript(tt.preset)
			require.True(t, ok)

			request, err := runtime.ParseRequest(t.Context(), script, map[string]string{
				"baseUrl":     "https://example.com",
				"apiKey":      "api-key",
				"accessToken": "access-token",
				"userId":      "42",
			})
			require.NoError(t, err)
			require.Equal(t, "https://example.com"+tt.path, request.URL)

			result, err := runtime.Extract(t.Context(), script, UsageQueryHTTPResponse{
				Status:  200,
				Headers: map[string]string{"content-type": "application/json"},
				Body:    tt.response,
			}, UsageQueryScriptContext{Now: "2026-08-18T07:36:46+08:00"})
			require.NoError(t, err)
			if tt.preset != objects.ChannelUsageQueryPresetNewAPI {
				require.NotNil(t, result.Progress)
			}
			if tt.preset == objects.ChannelUsageQueryPresetCodex || tt.preset == objects.ChannelUsageQueryPresetClaude {
				require.NotEmpty(t, result.Tags)
			} else if tt.preset == objects.ChannelUsageQueryPresetOpenCode {
				require.NotNil(t, result.Balance)
				require.Equal(t, 9.6, result.Balance.Remaining)
				require.Equal(t, "USD", result.Balance.Unit)
				require.Empty(t, result.Text)
			} else {
				require.NotEmpty(t, result.Text)
			}
		})
	}
}

func TestBuiltInUsageQuerySettings_UsesDisabledPresetsAndHonorsSavedDisable(t *testing.T) {
	codex := &ent.Channel{
		Type: channel.TypeCodex,
		Credentials: objects.ChannelCredentials{
			OAuth: &objects.OAuthCredentials{AccessToken: "token"},
		},
	}
	settings := BuiltInUsageQuerySettings(codex)
	require.NotNil(t, settings)
	require.False(t, settings.Enabled)
	require.Equal(t, objects.ChannelUsageQueryPresetCodex, settings.Preset)
	require.Equal(t, "https://chatgpt.com", settings.BaseURLOverride)
	require.False(t, NewUsageQueryChecker(nil).SupportsChannel(codex))

	codex.Settings = &objects.ChannelSettings{UsageQuery: &objects.ChannelUsageQuerySettings{Enabled: false}}
	require.Nil(t, BuiltInUsageQuerySettings(codex))
}

func TestBuiltInUsageQuerySettings_RequiresOAuthForCodexAndClaude(t *testing.T) {
	for _, channelType := range []channel.Type{channel.TypeCodex, channel.TypeClaudecode} {
		t.Run(channelType.String(), func(t *testing.T) {
			require.Nil(t, BuiltInUsageQuerySettings(&ent.Channel{
				Type:        channelType,
				Credentials: objects.ChannelCredentials{APIKey: "plain-api-key"},
			}))

			require.NotNil(t, BuiltInUsageQuerySettings(&ent.Channel{
				Type:        channelType,
				Credentials: objects.ChannelCredentials{OAuth: &objects.OAuthCredentials{AccessToken: "token"}},
			}))
		})
	}
}
