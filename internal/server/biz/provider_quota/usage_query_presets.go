package provider_quota

import (
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
)

// UsageQueryPresetScript returns the system-owned source for a preset. Scripts
// are deliberately fixed on the server so an API client cannot replace a
// provider preset while claiming to use it; CUSTOM remains user-authored.
func UsageQueryPresetScript(preset objects.ChannelUsageQueryPreset) (string, bool) {
	switch preset {
	case objects.ChannelUsageQueryPresetNewAPI:
		return newAPIUsageQueryScript, true
	case objects.ChannelUsageQueryPresetCodex:
		return codexUsageQueryScript, true
	case objects.ChannelUsageQueryPresetClaude:
		return claudeOAuthUsageQueryScript, true
	case objects.ChannelUsageQueryPresetOpenCode:
		return openCodeGoUsageQueryScript, true
	default:
		return "", false
	}
}

// BuiltInUsageQuerySettings provides the prefilled script for the native
// coding-subscription channels. It is intentionally disabled until an
// administrator explicitly saves the channel's usage-query configuration.
// A saved channel configuration still takes priority, including CUSTOM.
func BuiltInUsageQuerySettings(ch *ent.Channel) *objects.ChannelUsageQuerySettings {
	if ch == nil || (ch.Settings != nil && ch.Settings.UsageQuery != nil) {
		return nil
	}

	var preset objects.ChannelUsageQueryPreset
	var baseURL string
	switch ch.Type { //nolint:exhaustive
	case channel.TypeCodex:
		if !ch.Credentials.IsOAuth() {
			return nil
		}
		preset = objects.ChannelUsageQueryPresetCodex
		baseURL = "https://chatgpt.com"
	case channel.TypeClaudecode:
		if !ch.Credentials.IsOAuth() {
			return nil
		}
		preset = objects.ChannelUsageQueryPresetClaude
		baseURL = "https://api.anthropic.com"
	case channel.TypeOpencodeGo, channel.TypeOpencodeGoAnthropic:
		preset = objects.ChannelUsageQueryPresetOpenCode
		baseURL = ch.BaseURL
	default:
		return nil
	}

	script, ok := UsageQueryPresetScript(preset)
	if !ok {
		return nil
	}
	return &objects.ChannelUsageQuerySettings{
		Enabled:             false,
		ShowInProviderQuota: boolPtr(true),
		Preset:              preset,
		BaseURLOverride:     baseURL,
		Script:              script,
	}
}

func effectiveUsageQueryScript(settings *objects.ChannelUsageQuerySettings) string {
	if settings == nil {
		return ""
	}
	if script, ok := UsageQueryPresetScript(settings.Preset); ok {
		return script
	}
	return settings.Script
}

// EffectiveUsageQueryScript is exported for config reads so clients see the
// same immutable preset script the server executes.
func EffectiveUsageQueryScript(settings *objects.ChannelUsageQuerySettings) string {
	return effectiveUsageQueryScript(settings)
}

func boolPtr(value bool) *bool {
	return &value
}

const newAPIUsageQueryScript = `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/api/user/self",
    method: "GET",
    headers: {
      "Accept": "application/json",
      "Authorization": "Bearer {{accessToken}}",
      "User-Agent": "cc-switch/1.0",
      "New-Api-User": "{{userId}}"
    }
  },
  extractor: function (response) {
    const body = response.body || {};
    const data = body.data || {};
    const quota = Number(data.quota);
    if (response.status < 200 || response.status >= 300 || !body.success || !Number.isFinite(quota)) {
      throw new Error(body.message || "无法读取 New API 余额");
    }
    return { balance: { remaining: quota / 500000, unit: "A$" }, text: String(data.group || "默认套餐") };
  }
})`

const codexUsageQueryScript = `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/backend-api/wham/usage",
    method: "GET",
    headers: { "Accept": "application/json", "Authorization": "Bearer {{accessToken}}", "User-Agent": "cc-switch/1.0" }
  },
  extractor: function (response, context) {
    const body = response.body || {};
    const rateLimit = body.rate_limit || {};
    if (response.status < 200 || response.status >= 300 || !body.rate_limit) throw new Error(body.message || "无法读取 Codex 用量");
    function clamp(value) { return Math.max(0, Math.min(100, value)); }
    function formatTime(value) {
      const milliseconds = typeof value === "number" ? (value < 1000000000000 ? value * 1000 : value) : Date.parse(value || context.now);
      if (!Number.isFinite(milliseconds)) return undefined;
      const date = new Date(milliseconds), pad = function (number) { return String(number).padStart(2, "0"); };
      const offset = -date.getTimezoneOffset(), sign = offset >= 0 ? "+" : "-", absolute = Math.abs(offset);
      return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) + sign + pad(Math.floor(absolute / 60)) + ":" + pad(absolute % 60);
    }
    function windowResult(id, window) {
      if (!window) return null;
      const result = { id: id }, used = Number(window.used_percent), duration = Number(window.limit_window_seconds);
      if (Number.isFinite(used)) result.remainingPercent = clamp(100 - used);
      if (Number.isFinite(duration) && duration > 0) result.durationSeconds = Math.round(duration);
      const resetValue = window.reset_at != null ? window.reset_at : Date.parse(context.now) / 1000 + Number(window.reset_after_seconds);
      const resetAt = formatTime(resetValue);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }
    return { tags: [String(body.plan_type || "Codex")], progress: { windows: [windowResult("primary", rateLimit.primary_window), windowResult("secondary", rateLimit.secondary_window)].filter(Boolean) } };
  }
})`

const claudeOAuthUsageQueryScript = `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/api/oauth/usage",
    method: "GET",
    headers: { "Accept": "application/json", "Authorization": "Bearer {{accessToken}}", "anthropic-beta": "oauth-2025-04-20", "User-Agent": "cc-switch/1.0" }
  },
  extractor: function (response, context) {
    const body = response.body || {};
    if (response.status < 200 || response.status >= 300) throw new Error(body.message || "无法读取 Claude OAuth 用量");
    function clamp(value) { return Math.max(0, Math.min(100, value)); }
    function formatTime(value) {
      const milliseconds = typeof value === "number" ? (value < 1000000000000 ? value * 1000 : value) : Date.parse(value || context.now);
      if (!Number.isFinite(milliseconds)) return undefined;
      const date = new Date(milliseconds), pad = function (number) { return String(number).padStart(2, "0"); };
      const offset = -date.getTimezoneOffset(), sign = offset >= 0 ? "+" : "-", absolute = Math.abs(offset);
      return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) + sign + pad(Math.floor(absolute / 60)) + ":" + pad(absolute % 60);
    }
    function windowResult(id, window, durationSeconds) {
      if (!window) return null;
      const result = { id: id, durationSeconds: durationSeconds }, utilization = Number(window.utilization), used = Number(window.used), limit = Number(window.limit);
      if (Number.isFinite(utilization)) result.remainingPercent = clamp(100 - (utilization <= 1 ? utilization * 100 : utilization));
      else if (Number.isFinite(used) && Number.isFinite(limit) && limit > 0) result.remainingPercent = clamp(100 - used * 100 / limit);
      const resetAt = formatTime(window.resets_at || window.reset_at || window.reset);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }
    return { tags: [String(body.rate_limit_tier || "Claude")], progress: { windows: [windowResult("fiveHour", body.five_hour, 18000), windowResult("sevenDay", body.seven_day, 604800), windowResult("sevenDaySonnet", body.seven_day_sonnet, 604800), windowResult("sevenDayOpus", body.seven_day_opus, 604800)].filter(Boolean) } };
  }
})`

const openCodeGoUsageQueryScript = `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/usage",
    method: "GET",
    headers: { "Authorization": "Bearer {{apiKey}}", "Accept": "application/json", "User-Agent": "cc-switch/1.0" }
  },
  extractor: function (response, context) {
    const body = response.body || {}, usage = body.usage || {};
    if (response.status < 200 || response.status >= 300 || !body.usage) throw new Error(body.message || "无法读取 OpenCode Go 用量信息");
    function clamp(value) { return Math.max(0, Math.min(100, value)); }
    function formatTime(value) {
      const milliseconds = typeof value === "number" ? (value < 1000000000000 ? value * 1000 : value) : Date.parse(value || context.now);
      if (!Number.isFinite(milliseconds)) return undefined;
      const date = new Date(milliseconds), pad = function (number) { return String(number).padStart(2, "0"); };
      const offset = -date.getTimezoneOffset(), sign = offset >= 0 ? "+" : "-", absolute = Math.abs(offset);
      return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) + sign + pad(Math.floor(absolute / 60)) + ":" + pad(absolute % 60);
    }
    function windowResult(id, window, limit, durationSeconds) {
      if (!window) return null;
      const used = Number(window.percent);
      if (!Number.isFinite(used)) return null;
      const result = { id: id, durationSeconds: durationSeconds, remainingPercent: clamp(100 - used), limit: limit, usedPercent: used };
      const resetAt = formatTime(window.resetsAt);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }
    const windows = [windowResult("rolling", usage.rolling, 12, 18000), windowResult("weekly", usage.weekly, 30, 604800), windowResult("monthly", usage.monthly, 60, 2592000)];
    if (windows.some(function (window) { return window === null; })) throw new Error("无法读取 OpenCode Go 用量信息");
    const remaining = Math.max(0, Math.min.apply(null, windows.map(function (window) { return window.limit * (100 - window.usedPercent) / 100; })));
    return {
      text: "最小可用 USD " + remaining.toFixed(2),
      progress: { windows: windows.map(function (window) { return { id: window.id, durationSeconds: window.durationSeconds, remainingPercent: window.remainingPercent, resetAt: window.resetAt }; }) }
    };
  }
})`
