import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import test from 'node:test';
import { QueryClient } from '@tanstack/react-query';
import { invalidateChannelDependentQueries } from './channel-query-invalidation.ts';

const dataDir = import.meta.dirname;

test('channel changes make a cached API key profile preview refetch', async () => {
  const queryClient = new QueryClient();
  const queryKey = ['apiKeyProfilePreview', 'api-key-1'];
  let visibleModels = ['model-from-enabled-channel'];
  let fetchCount = 0;

  const fetchPreview = () =>
    queryClient.fetchQuery({
      queryKey,
      staleTime: Infinity,
      queryFn: () => {
        fetchCount += 1;
        return [...visibleModels];
      },
    });

  assert.deepEqual(await fetchPreview(), ['model-from-enabled-channel']);
  visibleModels = [];
  assert.deepEqual(await fetchPreview(), ['model-from-enabled-channel']);

  invalidateChannelDependentQueries(queryClient);

  assert.deepEqual(await fetchPreview(), []);
  assert.equal(fetchCount, 2);
  queryClient.clear();
});

test('channel mutations use the shared dependent-query invalidation', () => {
  for (const filename of ['channels.ts', 'templates.ts']) {
    const source = readFileSync(join(dataDir, filename), 'utf8');

    assert.match(source, /invalidateChannelDependentQueries\(queryClient\)/);
    assert.doesNotMatch(source, /queryClient\.invalidateQueries\(\{ queryKey: \['channels'\] \}\)/);
  }
});
