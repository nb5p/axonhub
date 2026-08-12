import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const detailSource = await readFile(new URL('./components/request-detail-content.tsx', import.meta.url), 'utf8');
const conversationSource = await readFile(new URL('./components/request-conversation-viewer.tsx', import.meta.url), 'utf8');

test('request headers start collapsed and report their root expansion state', () => {
  const requestHeadersViewer = detailSource.slice(
    detailSource.indexOf('data={request.requestHeaders}'),
    detailSource.indexOf('data={request.requestBody}')
  );

  assert.match(requestHeadersViewer, /defaultExpanded=\{false\}/);
  assert.match(requestHeadersViewer, /viewerId=\{`request-headers-\$\{request\.id\}`\}/);
  assert.match(requestHeadersViewer, /onExpandedChange=\{handleJsonViewerExpandedChange\}/);
});

test('page action collapses expanded JSON before returning to the page top', () => {
  assert.match(detailSource, /expandedJsonViewerCount > 0 \? collapseExpandedJsonViewers : scrollToPageTop/);
  assert.match(detailSource, /requests\.detail\.collapseExpandedContent/);
  assert.match(detailSource, /requests\.conversation\.backToTop/);
  assert.match(detailSource, /data-testid='request-detail-page-action'/);
});

test('JSON viewer no longer renders its own centered collapse control', () => {
  assert.doesNotMatch(detailSource, /showCollapseAll/);
  assert.doesNotMatch(detailSource, /-translate-x-1\/2 -translate-y-1\/2/);
  assert.doesNotMatch(conversationSource, /showBackTop/);
});
