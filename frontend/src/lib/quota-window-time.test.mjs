import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(join(import.meta.dirname, 'quota-window-time.ts'), 'utf8');
const quotaBadgesSource = readFileSync(join(import.meta.dirname, '..', 'components', 'quota-badges.tsx'), 'utf8');
const transpiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.ESNext,
    target: ts.ScriptTarget.ES2023,
  },
}).outputText;
const moduleUrl = `data:text/javascript;base64,${Buffer.from(transpiled).toString('base64')}`;
const { getQuotaWindowDurationPercent, getQuotaWindowResetAfterSeconds } = await import(moduleUrl);

const nowMs = Date.UTC(2026, 7, 20, 10, 0, 0);
const nowSeconds = nowMs / 1000;

test('absolute reset time overrides a stale relative snapshot', () => {
  const window = {
    reset_at: nowSeconds + 30 * 60,
    reset_after_seconds: 5 * 24 * 60 * 60,
  };

  assert.equal(getQuotaWindowResetAfterSeconds(window, nowMs), 30 * 60);
});

test('an expired absolute reset time is not replaced by a positive stale snapshot', () => {
  const window = {
    reset_at: nowSeconds - 60,
    reset_after_seconds: 5 * 24 * 60 * 60,
  };

  assert.equal(getQuotaWindowResetAfterSeconds(window, nowMs), -60);
});

test('relative reset time remains the fallback when the absolute time is absent', () => {
  assert.equal(getQuotaWindowResetAfterSeconds({ reset_after_seconds: 900 }, nowMs), 900);
  assert.equal(getQuotaWindowResetAfterSeconds({}, nowMs), undefined);
});

test('primary and secondary windows use the same absolute reset rule', () => {
  const windows = [
    { reset_at: nowSeconds + 900, reset_after_seconds: 999_999 },
    { reset_at: nowSeconds + 1800, reset_after_seconds: 888_888 },
  ];

  assert.deepEqual(
    windows.map((window) => getQuotaWindowResetAfterSeconds(window, nowMs)),
    [900, 1800]
  );
});

test('duration progress is calculated from the absolute reset time', () => {
  const window = {
    limit_window_seconds: 3600,
    reset_at: nowSeconds + 900,
    reset_after_seconds: 999_999,
  };

  assert.equal(getQuotaWindowDurationPercent(window, nowMs), 75);
});

test('Codex reset dates keep using the absolute timestamp in the browser local timezone', () => {
  assert.match(quotaBadgesSource, /new Date\(timestamp \* 1000\)/);
  assert.match(quotaBadgesSource, /toLocaleTimeString\(\[\], \{ hour: '2-digit', minute: '2-digit', hour12: false \}\)/);
  assert.match(quotaBadgesSource, /formatDate\(primaryWindow\.reset_at\)/);
  assert.match(quotaBadgesSource, /formatDate\(secondaryWindow\.reset_at\)/);
});
