import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(join(import.meta.dirname, 'usage-query-balance.ts'), 'utf8');
const transpiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2023,
  },
}).outputText;
const moduleUrl = `data:text/javascript;base64,${Buffer.from(transpiled).toString('base64')}`;
const { formatUsageQueryBalance } = await import(moduleUrl);

test('usage-query balance retains two decimals and puts ISO currency units after the amount', () => {
  assert.equal(formatUsageQueryBalance(9.99967699, 'USD'), '10.00 USD');
  assert.equal(formatUsageQueryBalance(12, 'USD'), '12.00 USD');
});

test('usage-query balance keeps A$ as a leading unit', () => {
  assert.equal(formatUsageQueryBalance(12, 'A$'), 'A$12.00');
});
