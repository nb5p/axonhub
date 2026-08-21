package provider_quota

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const tuziDayCardUsageQueryScript = `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/reseller/v1/quota",
    method: "GET",
    headers: {
      "Authorization": "Bearer {{apiKey}}",
      "Accept": "application/json",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function (response, context) {
    const body = response.body || {};
    const data = body.data || {};
    const subscription = data.subscription || {};
    if (response.status < 200 || response.status >= 300 || body.code !== 0) {
      throw new Error(body.message || "无法读取 Tuzi DayCard 用量");
    }

    function clamp(value) { return Math.max(0, Math.min(100, value)); }
    function formatTime(value) {
      const milliseconds = typeof value === "number"
        ? (value < 1000000000000 ? value * 1000 : value)
        : Date.parse(value || context.now);
      if (!Number.isFinite(milliseconds)) return undefined;
      const date = new Date(milliseconds), pad = function (n) { return String(n).padStart(2, "0"); };
      const offset = -date.getTimezoneOffset(), sign = offset >= 0 ? "+" : "-", absolute = Math.abs(offset);
      return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) + sign + pad(Math.floor(absolute / 60)) + ":" + pad(absolute % 60);
    }

    const windows = ["daily", "weekly", "monthly"].map(function (id) {
      const limit = Number(subscription[id + "_limit"]);
      const used = Number(subscription[id + "_used"]);
      if (!Number.isFinite(limit) || limit <= 0 || !Number.isFinite(used)) return null;
      const item = { id: id, remainingPercent: clamp(100 - used * 100 / limit) };
      const start = Date.parse(subscription[id + "_window_start"] || "");
      const reset = Date.parse(subscription[id + "_reset_at"] || "");
      if (Number.isFinite(start) && Number.isFinite(reset) && reset > start) item.durationSeconds = Math.round((reset - start) / 1000);
      const resetAt = formatTime(subscription[id + "_reset_at"]);
      if (resetAt) item.resetAt = resetAt;
      return item;
    }).filter(Boolean);

    const text = [subscription.group_name || "Tuzi DayCard"];
    if (data.concurrency != null) text.push("并发 " + data.concurrency);
    if (subscription.expires_at || data.expires_at) text.push("到期 " + (subscription.expires_at || data.expires_at));
    return { text: text.join(" · "), progress: { windows: windows } };
  }
})`

func TestTuziDayCardUsageQueryScript_ExtractsAllSubscriptionWindows(t *testing.T) {
	runtime := NewGojaUsageQueryRuntime()
	result, err := runtime.Extract(t.Context(), tuziDayCardUsageQueryScript, UsageQueryHTTPResponse{
		Status: 200,
		Body: map[string]any{
			"code": 0,
			"data": map[string]any{
				"concurrency": 8,
				"fuel_pack":   map[string]any{"available_usd": 0},
				"subscription": map[string]any{
					"group_name":           "Codex（月卡 lite）",
					"daily_used":           2.3340194,
					"daily_limit":          10,
					"weekly_used":          33.24014004,
					"weekly_limit":         50,
					"monthly_used":         33.24014004,
					"monthly_limit":        220,
					"daily_window_start":   "2026-08-21T04:30:00+08:00",
					"daily_reset_at":       "2026-08-22T04:30:00+08:00",
					"weekly_window_start":  "2026-08-18T07:36:46.820322+08:00",
					"weekly_reset_at":      "2026-08-25T07:36:46.820322+08:00",
					"monthly_window_start": "2026-08-18T07:36:46.820322+08:00",
					"monthly_reset_at":     "2026-09-17T07:36:46.820322+08:00",
				},
			},
		},
	}, UsageQueryScriptContext{Now: "2026-08-21T07:36:46+08:00"})
	require.NoError(t, err)
	require.Nil(t, result.Balance)
	require.Contains(t, result.Text, "Codex（月卡 lite）")
	require.Contains(t, result.Text, "并发 8")
	require.NotNil(t, result.Progress)
	require.Len(t, result.Progress.Windows, 3)
	require.Equal(t, []string{"daily", "weekly", "monthly"}, []string{
		result.Progress.Windows[0].ID,
		result.Progress.Windows[1].ID,
		result.Progress.Windows[2].ID,
	})
	require.InDelta(t, 76.659806, *result.Progress.Windows[0].RemainingPercent, 0.000001)
	require.InDelta(t, 33.519719, *result.Progress.Windows[1].RemainingPercent, 0.000001)
	require.InDelta(t, 84.890845, *result.Progress.Windows[2].RemainingPercent, 0.000001)
	require.Equal(t, "2026-08-22T04:30:00+08:00", result.Progress.Windows[0].ResetAt)
}
