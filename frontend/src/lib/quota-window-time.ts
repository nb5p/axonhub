export interface QuotaWindowReset {
  reset_at?: number | null;
  reset_after_seconds?: number | null;
}

export function getQuotaWindowResetAfterSeconds(window: QuotaWindowReset, nowMs: number = Date.now()): number | undefined {
  if (window.reset_at != null) {
    return window.reset_at - nowMs / 1000;
  }

  return window.reset_after_seconds ?? undefined;
}

export function getQuotaWindowDurationPercent(
  window: QuotaWindowReset & { limit_window_seconds?: number | null },
  nowMs: number = Date.now()
): number | undefined {
  const limit = window.limit_window_seconds;
  if (!limit) return undefined;

  const resetAfter = getQuotaWindowResetAfterSeconds(window, nowMs);
  if (resetAfter === undefined) return undefined;

  const elapsed = limit - resetAfter;
  return Math.max(0, Math.min(100, (elapsed / limit) * 100));
}
