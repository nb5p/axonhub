package biz

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/internal/server/biz/provider_quota"
)

func TestQuotaWindowExhaustionTransitions_NotifiesOnlyWhenWindowBecomesExhausted(t *testing.T) {
	resetAt := time.Date(2026, 8, 23, 7, 0, 0, 0, time.FixedZone("+08", 8*60*60))
	quotaData := provider_quota.QuotaData{Limits: []provider_quota.QuotaLimitStatus{
		{
			Type:        provider_quota.QuotaLimitTypeSubscriptionCycle,
			Window:      provider_quota.QuotaWindowDaily,
			Status:      "exhausted",
			UsageRatio:  1,
			Ready:       false,
			NextResetAt: &resetAt,
		},
	}}

	previousAvailable := map[string]any{
		"_limits": []map[string]any{{
			"type":       string(provider_quota.QuotaLimitTypeSubscriptionCycle),
			"window":     provider_quota.QuotaWindowDaily,
			"status":     "available",
			"usageRatio": 0.4,
			"ready":      true,
		}},
	}
	transitions := quotaWindowExhaustionTransitions("codex", previousAvailable, quotaData)
	require.Len(t, transitions, 1)
	require.Equal(t, provider_quota.QuotaWindowDaily, transitions[0].window)
	require.Equal(t, string(provider_quota.QuotaLimitTypeSubscriptionCycle), transitions[0].limitType)
	require.Equal(t, 0.0, *transitions[0].remainingPercent)
	require.Equal(t, resetAt, *transitions[0].resetAt)

	previousExhausted := map[string]any{
		"_limits": []map[string]any{{
			"type":       string(provider_quota.QuotaLimitTypeSubscriptionCycle),
			"window":     provider_quota.QuotaWindowDaily,
			"status":     "exhausted",
			"usageRatio": 1.0,
			"ready":      false,
		}},
	}
	require.Empty(t, quotaWindowExhaustionTransitions("codex", previousExhausted, quotaData))
}

func TestQuotaWindowExhaustionTransitions_HandlesUsageQueryProgressWindows(t *testing.T) {
	remaining := 0.0
	quotaData := provider_quota.QuotaData{RawData: map[string]any{
		"progress": &provider_quota.UsageQueryProgress{Windows: []provider_quota.UsageQueryProgressWindow{{
			ID:               "daily",
			RemainingPercent: &remaining,
			ResetAt:          "2026-08-23T07:36:46+08:00",
		}}},
	}}

	previous := map[string]any{
		"progress": map[string]any{
			"windows": []any{map[string]any{
				"id":               "daily",
				"remainingPercent": 25.0,
				"resetAt":          "2026-08-23T07:36:46+08:00",
			}},
		},
	}
	transitions := quotaWindowExhaustionTransitions("usage_query", previous, quotaData)
	require.Len(t, transitions, 1)
	require.Equal(t, "daily", transitions[0].window)
	require.Equal(t, "progress", transitions[0].limitType)
	require.Equal(t, remaining, *transitions[0].remainingPercent)

	previous["progress"] = map[string]any{
		"windows": []any{map[string]any{
			"id":               "daily",
			"remainingPercent": 0.0,
		}},
	}
	require.Empty(t, quotaWindowExhaustionTransitions("usage_query", previous, quotaData))
}

func TestQuotaBalance_ReadsUsageQueryBalanceAndPersistedMap(t *testing.T) {
	remaining := 0.0
	value, unit, ok := quotaBalance(map[string]any{
		"balance": &provider_quota.UsageQueryBalance{Remaining: remaining, Unit: "USD"},
	})
	require.True(t, ok)
	require.Equal(t, remaining, value)
	require.Equal(t, "USD", unit)

	value, unit, ok = quotaBalance(map[string]any{
		"balance": map[string]any{"remaining": 12.5, "unit": "CNY"},
	})
	require.True(t, ok)
	require.Equal(t, 12.5, value)
	require.Equal(t, "CNY", unit)

	_, _, ok = quotaBalance(map[string]any{"balance": lo.ToPtr("not-a-number")})
	require.False(t, ok)
}
