import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import { getVisibleSidebarNavigationGroups, SIDEBAR_NAVIGATION_GROUPS } from './sidebar-navigation.ts';

test('sidebar visibility definitions cover every current navigation item', () => {
  const items = SIDEBAR_NAVIGATION_GROUPS.flatMap((group) => group.items);
  const itemIDs = items.map((item) => item.id);

  assert.equal(items.length, 18);
  assert.equal(new Set(itemIDs).size, items.length);
  assert.ok(itemIDs.includes('admin.projects'));
  assert.ok(itemIDs.includes('admin.users'));
  assert.ok(itemIDs.includes('admin.roles'));
  assert.ok(itemIDs.includes('project.threads'));
  assert.ok(itemIDs.includes('settings.system'));
});

test('sidebar visibility can hide selected items or every group', () => {
  const visibleGroups = getVisibleSidebarNavigationGroups(['admin.users', 'project.threads']);
  const visibleItemIDs = visibleGroups.flatMap((group) => group.items.map((item) => item.id));
  assert.ok(!visibleItemIDs.includes('admin.users'));
  assert.ok(!visibleItemIDs.includes('project.threads'));
  assert.ok(visibleItemIDs.includes('project.requests'));

  const allItemIDs = SIDEBAR_NAVIGATION_GROUPS.flatMap((group) => group.items.map((item) => item.id));
  assert.deepEqual(getVisibleSidebarNavigationGroups(allItemIDs), []);
});

test('authenticated sidebar consumes the persisted visibility settings', () => {
  const sidebarSource = readFileSync(join(import.meta.dirname, 'sidebar.ts'), 'utf8');
  assert.match(sidebarSource, /useSidebarNavigationSettings\(\)/);
  assert.match(sidebarSource, /getVisibleSidebarNavigationGroups\(sidebarNavigationSettings\?\.hiddenItems \?\? \[\]\)/);
});
