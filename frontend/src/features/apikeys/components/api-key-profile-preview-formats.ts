export const CONVERSATION_API_FORMATS = [
  'openai/chat_completions',
  'openai/responses',
  'anthropic/messages',
  'gemini/contents',
] as const;

export function selectConversationAPIFormats(apiFormats: readonly string[]): string[] {
  const availableFormats = new Set(apiFormats);
  return CONVERSATION_API_FORMATS.filter((apiFormat) => availableFormats.has(apiFormat));
}
