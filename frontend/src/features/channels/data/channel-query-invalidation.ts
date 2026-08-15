import type { QueryClient } from '@tanstack/react-query';

export function invalidateChannelDependentQueries(queryClient: Pick<QueryClient, 'invalidateQueries'>) {
  void queryClient.invalidateQueries({ queryKey: ['channels'] });
  void queryClient.invalidateQueries({ queryKey: ['apiKeyProfilePreview'] });
}
