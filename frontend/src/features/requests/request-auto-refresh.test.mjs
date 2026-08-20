import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const requestsSource = await readFile(new URL('./index.tsx', import.meta.url), 'utf8');
const tableSource = await readFile(new URL('./components/requests-table.tsx', import.meta.url), 'utf8');
const toolbarSource = await readFile(new URL('./components/data-table-toolbar.tsx', import.meta.url), 'utf8');

test('request auto refresh remains available when pagination is disabled', () => {
  assert.match(requestsSource, /const isFirstPage = !paginationEnabled \|\|/);
  assert.match(requestsSource, /autoRefresh && isFirstPage \? 10000 : null/);
  assert.doesNotMatch(requestsSource, /setAutoRefresh\(false\)/);
  assert.doesNotMatch(requestsSource, /autoRefreshDisabled=\{!paginationEnabled\}/);
  assert.match(tableSource, /useAnimatedList\(data, autoRefresh,/);

  const autoRefreshControls = toolbarSource.slice(
    toolbarSource.indexOf('{showRefresh && onAutoRefreshChange && ('),
    toolbarSource.indexOf('{showRefresh && onRefresh && (')
  );
  assert.doesNotMatch(autoRefreshControls, /disabled=/);
});
