export function clampQuotaPercentage(percentage: number): number {
  return Math.min(Math.max(Number.isFinite(percentage) ? percentage : 0, 0), 100);
}

export function getQuotaDisplayPercentage(usedPercentage: number, reverseUsageDisplay: boolean): number {
  const used = clampQuotaPercentage(usedPercentage);
  return reverseUsageDisplay ? 100 - used : used;
}
