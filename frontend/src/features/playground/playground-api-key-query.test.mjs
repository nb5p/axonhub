import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';

const playgroundSource = readFileSync(join(import.meta.dirname, 'index.tsx'), 'utf8');

test('API key mode stays within the GraphQL connection page limit', () => {
  const pageSizeMatch = playgroundSource.match(/useApiKeys\(\s*\{\s*first:\s*(\d+)/s);

  assert.ok(pageSizeMatch, 'expected API key mode to declare a numeric page size');
  assert.ok(Number(pageSizeMatch[1]) <= 1000, 'apiKeys first cannot exceed 1000');
});
