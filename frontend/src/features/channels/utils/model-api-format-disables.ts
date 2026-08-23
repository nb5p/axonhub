import type { ChannelSettings } from '../data/schema';

export type DisabledModelAPIFormat = NonNullable<ChannelSettings['disabledModelApiFormats']>[number];

export function isModelAPIFormatDisabled(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string
): boolean {
  return (disabledModelAPIFormats ?? []).some(
    (entry) => entry.model === model && entry.apiFormats.includes(endpointFormat)
  );
}

export function disableModelAPIFormat(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string
): DisabledModelAPIFormat[] {
  const current = (disabledModelAPIFormats ?? []).map((entry) => ({
    model: entry.model,
    apiFormats: [...entry.apiFormats],
  }));
  const existing = current.find((entry) => entry.model === model);

  if (!existing) {
    return [...current, { model, apiFormats: [endpointFormat] }];
  }

  if (!existing.apiFormats.includes(endpointFormat)) {
    existing.apiFormats.push(endpointFormat);
  }

  return current;
}

export function enableModelAPIFormat(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string
): DisabledModelAPIFormat[] {
  return (disabledModelAPIFormats ?? []).flatMap((entry) => {
    if (entry.model !== model) {
      return [{ model: entry.model, apiFormats: [...entry.apiFormats] }];
    }

    const apiFormats = entry.apiFormats.filter((format) => format !== endpointFormat);
    return apiFormats.length > 0 ? [{ model: entry.model, apiFormats }] : [];
  });
}
