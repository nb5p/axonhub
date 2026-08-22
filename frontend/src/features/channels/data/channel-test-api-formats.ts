export const channelTestAPIFormats = [
  {
    value: 'OPENAI_CHAT_COMPLETION',
    endpointFormat: 'openai/chat_completions',
    labelKey: 'channels.dialogs.test.apiFormats.openaiChatCompletion',
  },
  {
    value: 'OPENAI_RESPONSE',
    endpointFormat: 'openai/responses',
    labelKey: 'channels.dialogs.test.apiFormats.openaiResponse',
  },
  {
    value: 'ANTHROPIC_MESSAGES',
    endpointFormat: 'anthropic/messages',
    labelKey: 'channels.dialogs.test.apiFormats.anthropicMessages',
  },
  {
    value: 'GEMINI_CONTENTS',
    endpointFormat: 'gemini/contents',
    labelKey: 'channels.dialogs.test.apiFormats.geminiContents',
  },
] as const;

export type ChannelTestAPIFormat = (typeof channelTestAPIFormats)[number]['value'];

export const defaultChannelTestAPIFormats: ChannelTestAPIFormat[] = channelTestAPIFormats.slice(0, 3).map((format) => format.value);

export function getChannelTestAPIFormat(value: ChannelTestAPIFormat) {
  return channelTestAPIFormats.find((format) => format.value === value)!;
}

export function getAvailableChannelTestAPIFormats(availableEndpointFormats: ReadonlySet<string>): ChannelTestAPIFormat[] {
  return channelTestAPIFormats
    .filter((format) => availableEndpointFormats.has(format.endpointFormat))
    .map((format) => format.value);
}

export function getDefaultAvailableChannelTestAPIFormats(availableEndpointFormats: ReadonlySet<string>): ChannelTestAPIFormat[] {
  const availableFormats = getAvailableChannelTestAPIFormats(availableEndpointFormats);
  const selectedDefaults = defaultChannelTestAPIFormats.filter((format) => availableFormats.includes(format));
  return selectedDefaults.length > 0 ? selectedDefaults : availableFormats;
}

export function orderChannelTestAPIFormats(values: ChannelTestAPIFormat[]): ChannelTestAPIFormat[] {
  return channelTestAPIFormats.filter((format) => values.includes(format.value)).map((format) => format.value);
}
