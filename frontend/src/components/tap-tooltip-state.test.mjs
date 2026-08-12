import assert from 'node:assert/strict';
import test from 'node:test';
import { getTooltipOpenStateAfterPress, scheduleTooltipOpenStateAfterPress } from './ui/tap-tooltip-state.ts';

test('touch and pen presses toggle tooltip visibility', () => {
  assert.equal(getTooltipOpenStateAfterPress('touch', false), true);
  assert.equal(getTooltipOpenStateAfterPress('touch', true), false);
  assert.equal(getTooltipOpenStateAfterPress('pen', false), true);
});

test('mouse and keyboard activation keep the native tooltip behavior', () => {
  assert.equal(getTooltipOpenStateAfterPress('mouse', false), undefined);
  assert.equal(getTooltipOpenStateAfterPress('', false), undefined);
});

test('touch tooltip state is applied after the current click handler', async () => {
  const states = [];

  assert.equal(
    scheduleTooltipOpenStateAfterPress('touch', false, (open) => states.push(open)),
    true
  );
  assert.deepEqual(states, []);

  await Promise.resolve();
  assert.deepEqual(states, [true]);
  assert.equal(
    scheduleTooltipOpenStateAfterPress('mouse', false, (open) => states.push(open)),
    false
  );
});
