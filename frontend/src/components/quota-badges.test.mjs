import assert from 'node:assert/strict';
import test from 'node:test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const componentsDir = import.meta.dirname;
const srcRoot = join(componentsDir, '..');

function read(relativePath) {
  return readFileSync(join(srcRoot, relativePath), 'utf8');
}

// Isolate the Codex render branch (between the Codex and Cline branch
// markers) so assertions about the usage bars cannot bleed into Claude Code
// or Cline, which legitimately keep duration-aware severity.
function isolateCodexBlock(source) {
  const start = source.indexOf("{channel.type === 'codex' &&");
  const end = source.indexOf("{channel.type === 'cline' &&", start);

  assert.ok(start !== -1, 'Codex render branch should exist in quota-badges source');
  assert.ok(end !== -1 && end > start, 'Cline render branch should follow the Codex branch');

  return source.slice(start, end);
}

test('Codex usage bar color tracks used percentage, not reset-window elapsed time', () => {
  const quotaBadges = read('components/quota-badges.tsx');
  const codexBlock = isolateCodexBlock(quotaBadges);

  // Both usage bars must render the user-visible used percentage so their
  // severity reflects actual usage rather than elapsed reset-window time.
  assert.match(
    codexBlock,
    /usagePercent=\{primaryWindow\.used_percent \|\| 0\}[\s\S]*?timeAffectsSeverity=\{false\}/,
    'Codex primary usage bar should use the primary-window percentage without time-based severity'
  );
  assert.match(
    codexBlock,
    /usagePercent=\{secondaryWindow\.used_percent\}[\s\S]*?timeAffectsSeverity=\{false\}/,
    'Codex secondary usage bar should use the secondary-window percentage without time-based severity'
  );

  // Reset-window elapsed time stays visible through UsageTimeBar's marker or
  // duration bar, but does not alter Codex usage severity.
  assert.equal(
    (codexBlock.match(/timeAffectsSeverity=\{false\}/g) || []).length,
    2,
    'Codex should opt both quota windows out of time-based severity'
  );
});
