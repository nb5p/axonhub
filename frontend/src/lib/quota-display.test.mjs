import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(join(import.meta.dirname, 'quota-display.ts'), 'utf8');
const transpiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2023,
  },
}).outputText;
const moduleUrl = `data:text/javascript;base64,${Buffer.from(transpiled).toString('base64')}`;
const { clampQuotaPercentage, getQuotaDisplayPercentage } = await import(moduleUrl);

test('quota display can switch between used and remaining percentages', () => {
  assert.equal(getQuotaDisplayPercentage(33, false), 33);
  assert.equal(getQuotaDisplayPercentage(33, true), 67);
});

test('quota display percentages stay within the visible range', () => {
  assert.equal(clampQuotaPercentage(-10), 0);
  assert.equal(clampQuotaPercentage(120), 100);
  assert.equal(getQuotaDisplayPercentage(Number.NaN, true), 100);
});
