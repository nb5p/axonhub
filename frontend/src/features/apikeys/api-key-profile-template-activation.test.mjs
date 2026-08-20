import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import { buildGUID } from '../../lib/utils.ts';

const featureDir = import.meta.dirname;

test('active profile quick switch sends the template GraphQL GUID', () => {
  const dataSource = readFileSync(join(featureDir, 'data/apikeys.ts'), 'utf8');
  const activationHookSource = dataSource.slice(dataSource.indexOf('export function useActivateApiKeyProfileTemplate'));

  assert.equal(buildGUID('APIKeyProfileTemplate', '12'), 'gid://axonhub/APIKeyProfileTemplate/12');
  assert.match(activationHookSource, /templateID:\s*buildGUID\('APIKeyProfileTemplate',\s*String\(templateID\)\)/);
  assert.doesNotMatch(activationHookSource, /templateID:\s*String\(templateID\)/);
});

test('active profile quick switch verifies and applies the server-confirmed profile', () => {
  const dataSource = readFileSync(join(featureDir, 'data/apikeys.ts'), 'utf8');
  const activationHookSource = dataSource.slice(dataSource.indexOf('export function useActivateApiKeyProfileTemplate'));

  assert.match(activationHookSource, /updatedApiKey\.profiles\?\.activeProfile\s*!==\s*activatedProfile\.name/);
  assert.match(activationHookSource, /setQueriesData<ApiKeyConnection>/);
  assert.match(activationHookSource, /setQueriesData<ApiKey>/);
  assert.match(activationHookSource, /handleError\(error/);
});

test('profile editor makes the save boundary explicit', () => {
  const dialogSource = readFileSync(join(featureDir, 'components/apikeys-profiles-dialog.tsx'), 'utf8');
  const english = JSON.parse(readFileSync(join(featureDir, '../../locales/en/apikeys.json'), 'utf8'));
  const chinese = JSON.parse(readFileSync(join(featureDir, '../../locales/zh-CN/apikeys.json'), 'utf8'));

  assert.match(dialogSource, /dirtyFields\.activeProfile/);
  assert.match(dialogSource, /apikeys\.profiles\.activeProfilePendingSave/);
  assert.match(english['apikeys.profiles.activeProfilePendingSave'], /Not active yet/);
  assert.match(chinese['apikeys.profiles.activeProfilePendingSave'], /尚未生效/);
});
