import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { resolveAPIKeyProfilePreview } from './components/api-key-profile-preview-state.ts';

const featureDir = import.meta.dirname;

function read(relativePath) {
  return readFileSync(join(featureDir, relativePath), 'utf8');
}

test('profile preview does not retain data from another API key', () => {
  const dataSource = read('data/apikeys.ts');
  const previewHookSource = dataSource.slice(
    dataSource.indexOf('export function useApiKeyProfilePreview'),
    dataSource.indexOf('export function useApiKeyTokenUsageStats')
  );

  assert.doesNotMatch(previewHookSource, /keepPreviousData/);
  assert.doesNotMatch(previewHookSource, /placeholderData:/);
});

test('profile preview loads an unrestricted preview when the API key has no profiles', () => {
  const dialogSource = read('components/apikeys-profiles-dialog.tsx');
  const previewProfile = resolveAPIKeyProfilePreview(undefined);

  assert.match(dialogSource, /resolveAPIKeyProfilePreview\(activeProfile\)/);
  assert.match(dialogSource, /enabled:\s*open\s*&&\s*!!apiKeyId\s*&&\s*!initialDataLoading/);
  assert.deepEqual(previewProfile.modelMappings, []);
  assert.deepEqual(previewProfile.modelIDs, []);
  assert.deepEqual(previewProfile.channelIDs, []);
  assert.deepEqual(previewProfile.channelTags, []);
});
