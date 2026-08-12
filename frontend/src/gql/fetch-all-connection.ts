import type { PageInfo } from './pagination';

export const MAX_CONNECTION_PAGE_SIZE = 1000;

interface ConnectionPage<TEdge extends { cursor: string }> {
  edges: TEdge[];
  pageInfo: PageInfo;
  totalCount: number;
}

export async function fetchAllConnectionPages<TEdge extends { cursor: string }, TConnection extends ConnectionPage<TEdge>>(
  fetchPage: (after?: string) => Promise<TConnection>
): Promise<TConnection> {
  const edges: TEdge[] = [];
  let after: string | undefined;
  let lastPage: TConnection | undefined;

  do {
    const page = await fetchPage(after);
    edges.push(...page.edges);
    lastPage = page;

    const nextCursor = page.pageInfo.hasNextPage ? (page.pageInfo.endCursor ?? undefined) : undefined;
    if (nextCursor && nextCursor === after) {
      throw new Error('Connection pagination returned the same cursor twice');
    }
    after = nextCursor;
  } while (after);

  if (!lastPage) {
    throw new Error('Connection pagination returned no page');
  }

  return {
    ...lastPage,
    edges,
    pageInfo: {
      hasNextPage: false,
      hasPreviousPage: false,
      startCursor: edges.length > 0 ? edges[0].cursor : null,
      endCursor: edges.length > 0 ? edges[edges.length - 1].cursor : null,
    },
    totalCount: lastPage.totalCount,
  };
}
