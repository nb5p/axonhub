import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(join(import.meta.dirname, 'list-pagination-settings.ts'), 'utf8');
const transpiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2023,
  },
}).outputText;
const moduleUrl = `data:text/javascript;base64,${Buffer.from(transpiled).toString('base64')}`;
const { buildListPaginationSettingsInput } = await import(moduleUrl);

test('pagination settings mutation input excludes frontend-only capability fields', () => {
  const input = buildListPaginationSettingsInput({
    channels: true,
    apiKeys: false,
    requests: true,
    supported: true,
  });

  assert.deepEqual(input, {
    channels: true,
    apiKeys: false,
    requests: true,
  });
});
