import type { ChannelSettings } from '../data/schema';

export type DisabledModelAPIFormat = NonNullable<ChannelSettings['disabledModelApiFormats']>[number];

export function isModelAPIFormatDisabled(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string,
  client = 'unknown'
): boolean {
  return (disabledModelAPIFormats ?? []).some(
    (entry) =>
      entry.model === model &&
      entry.apiFormats.includes(endpointFormat) &&
      ((entry.clients?.length ?? 0) === 0 || entry.clients?.includes(client))
  );
}

export function disableModelAPIFormat(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string,
  clients?: string[]
): DisabledModelAPIFormat[] {
  const current = (disabledModelAPIFormats ?? []).map((entry) => ({
    model: entry.model,
    apiFormats: [...entry.apiFormats],
    ...(entry.clients?.length ? { clients: [...entry.clients] } : {}),
  }));
  const normalizedClients = Array.from(new Set((clients ?? []).map((client) => client.trim().toLowerCase()).filter(Boolean))).sort();
  const existing = current.find(
    (entry) => entry.model === model && JSON.stringify([...(entry.clients ?? [])].sort()) === JSON.stringify(normalizedClients)
  );

  if (!existing) {
    return [...current, { model, apiFormats: [endpointFormat], ...(normalizedClients.length ? { clients: normalizedClients } : {}) }];
  }

  if (!existing.apiFormats.includes(endpointFormat)) {
    existing.apiFormats.push(endpointFormat);
  }

  return current;
}

export function enableModelAPIFormat(
  disabledModelAPIFormats: ChannelSettings['disabledModelApiFormats'] | undefined | null,
  model: string,
  endpointFormat: string,
  clients?: string[]
): DisabledModelAPIFormat[] {
  const normalizedClients = Array.from(new Set((clients ?? []).map((client) => client.trim().toLowerCase()).filter(Boolean))).sort();
  return (disabledModelAPIFormats ?? []).flatMap((entry) => {
    const entryClients = [...(entry.clients ?? [])].sort();
    if (entry.model !== model || JSON.stringify(entryClients) !== JSON.stringify(normalizedClients)) {
      return [{ model: entry.model, apiFormats: [...entry.apiFormats], ...(entry.clients?.length ? { clients: [...entry.clients] } : {}) }];
    }

    const apiFormats = entry.apiFormats.filter((format) => format !== endpointFormat);
    return apiFormats.length > 0
      ? [{ model: entry.model, apiFormats, ...(entry.clients?.length ? { clients: [...entry.clients] } : {}) }]
      : [];
  });
}
