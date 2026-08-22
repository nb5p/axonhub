export function formatUsageQueryBalance(value: number | null | undefined, unit?: string | null): string {
  if (value == null || !Number.isFinite(value)) return '-';

  const formatted = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
  const normalizedUnit = unit?.trim();

  if (normalizedUnit === 'A$') return `A$${formatted}`;
  return normalizedUnit ? `${formatted} ${normalizedUnit}` : formatted;
}
