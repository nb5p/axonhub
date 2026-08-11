import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(join(import.meta.dirname, 'filter-storage.ts'), 'utf8');
const transpiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2023,
  },
}).outputText;
const moduleUrl = `data:text/javascript;base64,${Buffer.from(transpiled).toString('base64')}`;
const { buildFilterStorageKey, deserializeFilterValue, serializeFilterValue } = await import(moduleUrl);

test('filter storage keys are versioned and scoped by page', () => {
  assert.equal(buildFilterStorageKey('channels', 'models'), 'axonhub:filters:v1:channels:models');
  assert.notEqual(buildFilterStorageKey('channels', 'models'), buildFilterStorageKey('requests', 'models'));
});

test('filter values preserve dates and nested arrays across serialization', () => {
  const value = {
    models: ['gpt-5', 'claude-sonnet'],
    range: {
      from: new Date('2026-08-10T00:00:00.000Z'),
      to: new Date('2026-08-11T23:59:59.000Z'),
    },
  };

  const restored = deserializeFilterValue(serializeFilterValue(value));
  assert.deepEqual(restored.models, value.models);
  assert.ok(restored.range.from instanceof Date);
  assert.ok(restored.range.to instanceof Date);
  assert.equal(restored.range.from.toISOString(), value.range.from.toISOString());
  assert.equal(restored.range.to.toISOString(), value.range.to.toISOString());
});

test('undefined filter values can be represented by removing the storage entry', () => {
  assert.equal(serializeFilterValue(undefined), undefined);
});
