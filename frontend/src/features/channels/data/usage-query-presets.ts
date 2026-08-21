export type ChannelUsageQueryPreset = 'NEW_API' | 'CODEX' | 'CLAUDE_OAUTH' | 'OPENCODE_GO' | 'CUSTOM';

export type UsageQueryPresetDefinition = {
  script: string;
  defaultBaseUrl?: string;
  requiresUserId?: boolean;
};

const timeHelpers = `
  function clamp(value) {
    return Math.max(0, Math.min(100, value));
  }

  function toMilliseconds(value, context) {
    if (typeof value === "number") {
      return value < 1000000000000 ? value * 1000 : value;
    }
    if (typeof value === "string") {
      const parsed = Date.parse(value);
      if (Number.isFinite(parsed)) return parsed;
    }
    return Date.parse(context.now);
  }

  function formatOffsetTime(value, context) {
    const milliseconds = toMilliseconds(value, context);
    if (!Number.isFinite(milliseconds)) return undefined;

    const date = new Date(milliseconds);
    const pad = function (number) { return String(number).padStart(2, "0"); };
    const offset = -date.getTimezoneOffset();
    const sign = offset >= 0 ? "+" : "-";
    const absoluteOffset = Math.abs(offset);
    return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) +
      "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) +
      sign + pad(Math.floor(absoluteOffset / 60)) + ":" + pad(absoluteOffset % 60);
  }
`;

export const USAGE_QUERY_PRESETS: Record<ChannelUsageQueryPreset, UsageQueryPresetDefinition> = {
  NEW_API: {
    requiresUserId: true,
    script: `({
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

    return {
      balance: { remaining: quota / 500000, unit: "A$" },
      text: String(data.group || "默认套餐")
    };
  }
})`,
  },
  CODEX: {
    defaultBaseUrl: 'https://chatgpt.com',
    script: `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/backend-api/wham/usage",
    method: "GET",
    headers: {
      "Accept": "application/json",
      "Authorization": "Bearer {{accessToken}}",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function (response, context) {
    const body = response.body || {};
    const rateLimit = body.rate_limit || {};
${timeHelpers}
    if (response.status < 200 || response.status >= 300 || !body.rate_limit) {
      throw new Error(body.message || "无法读取 Codex 用量");
    }

    function makeWindow(id, window) {
      if (!window) return null;
      const usedPercent = Number(window.used_percent);
      const result = { id: id };
      const duration = Number(window.limit_window_seconds);
      if (Number.isFinite(duration) && duration > 0) result.durationSeconds = Math.round(duration);
      if (Number.isFinite(usedPercent)) result.remainingPercent = clamp(100 - usedPercent);
      const resetValue = window.reset_at != null
        ? window.reset_at
        : Number(window.reset_after_seconds) + Date.parse(context.now) / 1000;
      const resetAt = formatOffsetTime(resetValue, context);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }

    const windows = [
      makeWindow("primary", rateLimit.primary_window),
      makeWindow("secondary", rateLimit.secondary_window)
    ].filter(Boolean);

    return {
      tags: [String(body.plan_type || "Codex")],
      progress: { windows: windows }
    };
  }
})`,
  },
  CLAUDE_OAUTH: {
    defaultBaseUrl: 'https://api.anthropic.com',
    script: `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/api/oauth/usage",
    method: "GET",
    headers: {
      "Accept": "application/json",
      "Authorization": "Bearer {{accessToken}}",
      "anthropic-beta": "oauth-2025-04-20",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function (response, context) {
    const body = response.body || {};
${timeHelpers}
    if (response.status < 200 || response.status >= 300) {
      throw new Error(body.message || "无法读取 Claude OAuth 用量");
    }

    function remainingPercent(window) {
      if (!window) return undefined;
      const utilization = Number(window.utilization);
      if (Number.isFinite(utilization)) return clamp(100 - (utilization <= 1 ? utilization * 100 : utilization));
      const usedPercent = Number(window.used_percent != null ? window.used_percent : window.percent_used);
      if (Number.isFinite(usedPercent)) return clamp(100 - usedPercent);
      const used = Number(window.used);
      const limit = Number(window.limit);
      if (Number.isFinite(used) && Number.isFinite(limit) && limit > 0) return clamp(100 - used * 100 / limit);
      return undefined;
    }

    function makeWindow(id, window, durationSeconds) {
      if (!window) return null;
      const result = { id: id, durationSeconds: durationSeconds };
      const percent = remainingPercent(window);
      if (percent !== undefined) result.remainingPercent = percent;
      const resetAt = formatOffsetTime(window.resets_at || window.reset_at || window.reset, context);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }

    const windows = [
      makeWindow("fiveHour", body.five_hour, 5 * 60 * 60),
      makeWindow("sevenDay", body.seven_day, 7 * 24 * 60 * 60),
      makeWindow("sevenDaySonnet", body.seven_day_sonnet, 7 * 24 * 60 * 60),
      makeWindow("sevenDayOpus", body.seven_day_opus, 7 * 24 * 60 * 60)
    ].filter(Boolean);
    return {
      tags: [String(body.rate_limit_tier || "Claude")],
      progress: { windows: windows }
    };
  }
})`,
  },
  OPENCODE_GO: {
    script: `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/usage",
    method: "GET",
    headers: {
      "Authorization": "Bearer {{apiKey}}",
      "Accept": "application/json",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function (response, context) {
    const body = response.body || {};
    const usage = body.usage || {};
${timeHelpers}
    if (response.status < 200 || response.status >= 300 || !body.usage) {
      throw new Error(body.message || "无法读取 OpenCode Go 用量信息");
    }

    function makeWindow(id, data, limit, durationSeconds) {
      if (!data) return null;
      const usedPercent = Number(data.percent);
      if (!Number.isFinite(usedPercent)) return null;
      const result = {
        id: id,
        durationSeconds: durationSeconds,
        remainingPercent: clamp(100 - usedPercent),
        limit: limit,
        usedPercent: usedPercent
      };
      const resetAt = formatOffsetTime(data.resetsAt, context);
      if (resetAt) result.resetAt = resetAt;
      return result;
    }

    const windows = [
      makeWindow("rolling", usage.rolling, 12, 5 * 60 * 60),
      makeWindow("weekly", usage.weekly, 30, 7 * 24 * 60 * 60),
      makeWindow("monthly", usage.monthly, 60, 30 * 24 * 60 * 60)
    ];

    if (windows.some(function (window) { return window === null; })) {
      throw new Error("无法读取 OpenCode Go 用量信息");
    }

    const remaining = Math.max(0, Math.min.apply(null, windows.map(function (window) {
      return window.limit * (100 - window.usedPercent) / 100;
    })));

    return {
      text: "最小可用 USD " + remaining.toFixed(2),
      progress: {
        windows: windows.map(function (window) {
          return {
            id: window.id,
            durationSeconds: window.durationSeconds,
            remainingPercent: window.remainingPercent,
            resetAt: window.resetAt
          };
        })
      }
    };
  }
})`,
  },
  CUSTOM: {
    script: `({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/user/balance",
    method: "GET",
    headers: {
      "Accept": "application/json",
      "Authorization": "Bearer {{apiKey}}",
      "User-Agent": "cc-switch/1.0"
    }
  },
  extractor: function (response) {
    const body = response.body || {};
    const balance = Number(body.balance);
    if (response.status < 200 || response.status >= 300 || !Number.isFinite(balance)) {
      throw new Error(body.message || "无法读取余额");
    }
    return { balance: { remaining: balance, unit: "A$" } };
  }
})`,
  },
};

export function getUsageQueryPreset(preset: ChannelUsageQueryPreset): UsageQueryPresetDefinition {
  return USAGE_QUERY_PRESETS[preset] ?? USAGE_QUERY_PRESETS.CUSTOM;
}
