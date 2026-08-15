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
