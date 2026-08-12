import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const tooltipSource = await readFile(new URL('./ui/tooltip.tsx', import.meta.url), 'utf8');
const hoverCardSource = await readFile(new URL('./ui/hover-card.tsx', import.meta.url), 'utf8');
const truncatedTextSource = await readFile(new URL('./truncated-text.tsx', import.meta.url), 'utf8');
const activityHeatmapSource = await readFile(
  new URL('../features/dashboard/components/api-key-activity-heatmap.tsx', import.meta.url),
  'utf8'
);

test('shared tooltip trigger handles touch-like pointer presses', () => {
  assert.match(tooltipSource, /scheduleTooltipOpenStateAfterPress/);
  assert.match(tooltipSource, /onPointerDown/);
  assert.match(tooltipSource, /onClick/);
});

test('shared hover card trigger handles touch-like pointer presses', () => {
  assert.match(hoverCardSource, /getTooltipOpenStateAfterPress/);
  assert.match(hoverCardSource, /onPointerDown/);
});

test('API key activity blocks use the touch-capable shared tooltip', () => {
  assert.doesNotMatch(activityHeatmapSource, /tooltips=\{/);
  assert.match(activityHeatmapSource, /<TooltipTrigger asChild>/);
  assert.match(activityHeatmapSource, /tabIndex: 0/);
});

test('truncated text uses the shared tooltip instead of a native title', () => {
  assert.match(truncatedTextSource, /<TooltipTrigger asChild>/);
  assert.doesNotMatch(truncatedTextSource, /title=/);
});
