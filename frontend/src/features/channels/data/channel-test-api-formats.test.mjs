import assert from 'node:assert/strict';
import test from 'node:test';
import {
  getAvailableChannelTestAPIFormats,
  getDefaultAvailableChannelTestAPIFormats,
} from './channel-test-api-formats.ts';

test('single-channel testing only exposes endpoint formats the channel supports', () => {
  const endpointFormats = new Set(['openai/responses']);

  assert.deepEqual(getAvailableChannelTestAPIFormats(endpointFormats), ['OPENAI_RESPONSE']);
  assert.deepEqual(getDefaultAvailableChannelTestAPIFormats(endpointFormats), ['OPENAI_RESPONSE']);
});

test('single-channel testing keeps the common supported defaults in display order', () => {
  const endpointFormats = new Set(['gemini/contents', 'anthropic/messages', 'openai/chat_completions']);

  assert.deepEqual(getAvailableChannelTestAPIFormats(endpointFormats), ['OPENAI_CHAT_COMPLETION', 'ANTHROPIC_MESSAGES', 'GEMINI_CONTENTS']);
  assert.deepEqual(getDefaultAvailableChannelTestAPIFormats(endpointFormats), ['OPENAI_CHAT_COMPLETION', 'ANTHROPIC_MESSAGES']);
});
