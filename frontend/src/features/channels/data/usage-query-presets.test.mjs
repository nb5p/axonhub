import assert from 'node:assert/strict';
import test from 'node:test';
import { USAGE_QUERY_PRESETS } from './usage-query-presets.ts';

test('OpenCode Go preset returns the calculated minimum as a USD balance', () => {
  const script = USAGE_QUERY_PRESETS.OPENCODE_GO.script;

  assert.match(script, /balance:\s*\{[\s\S]*remaining:\s*Number\(remaining\.toFixed\(2\)\)[\s\S]*unit:\s*"USD"/);
  assert.doesNotMatch(script, /text:\s*"最小可用 USD/);
});
