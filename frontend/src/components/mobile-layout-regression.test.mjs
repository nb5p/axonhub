import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const readSource = (relativePath) => readFile(new URL(relativePath, import.meta.url), 'utf8');

const profileDialogSource = await readSource('../features/apikeys/components/apikeys-profiles-dialog.tsx');
const profilePreviewSource = await readSource('../features/apikeys/components/apikey-profile-preview-panel.tsx');
const requestDetailPageSource = await readSource('../features/requests/components/request-detail-page.tsx');
const requestDetailGlobalPageSource = await readSource('../features/requests/components/request-detail-global-page.tsx');
const requestDetailHeaderSource = await readSource('../features/requests/components/request-detail-header.tsx').catch(() => '');
const requestTableSource = await readSource('../features/requests/components/requests-table.tsx');
const requestToolbarSource = await readSource('../features/requests/components/data-table-toolbar.tsx');
const serverPaginationSource = await readSource('./server-side-pagination.tsx');

test('API key profile management uses a full-height mobile dialog with editor and preview panels', () => {
  assert.match(profileDialogSource, /h-\[100dvh\]/);
  assert.match(profileDialogSource, /mobilePanel/);
  assert.match(profileDialogSource, /role='tablist'/);
  assert.match(profilePreviewSource, /grid-cols-2[^']*sm:grid-cols-4/);
});

test('request detail pages share a vertically flowing mobile header', () => {
  assert.match(requestDetailPageSource, /<RequestDetailHeader/);
  assert.match(requestDetailGlobalPageSource, /<RequestDetailHeader/);
  assert.match(requestDetailHeaderSource, /flex-col sm:flex-row/);
  assert.match(requestDetailHeaderSource, /h-auto/);
});

test('request logs use mobile cards and retain the desktop table above the mobile breakpoint', () => {
  assert.match(requestTableSource, /data-testid='requests-mobile-list'/);
  assert.match(requestTableSource, /data-testid='requests-desktop-table'/);
  assert.match(requestTableSource, /hidden[^']*md:block/);
  assert.match(requestToolbarSource, /basis-full/);
});

test('server pagination exposes mobile-sized navigation targets', () => {
  assert.match(serverPaginationSource, /min-h-12 min-w-12/);
  assert.match(serverPaginationSource, /pagination\.totalRows/);
});
