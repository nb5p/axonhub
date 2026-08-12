export interface ListPaginationSettingsInput {
  channels: boolean;
  apiKeys: boolean;
  requests: boolean;
}

export function buildListPaginationSettingsInput(
  value: ListPaginationSettingsInput
): ListPaginationSettingsInput {
  return {
    channels: value.channels,
    apiKeys: value.apiKeys,
    requests: value.requests,
  };
}
