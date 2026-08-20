import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { graphqlRequest } from '@/gql/graphql';

export type ChannelUsageQueryPreset = 'NEW_API' | 'CUSTOM';

export interface ChannelUsageQueryConfig {
  enabled: boolean;
  preset: ChannelUsageQueryPreset;
  baseUrlOverride?: string | null;
  userId?: string | null;
  script: string;
  apiKeyConfigured: boolean;
}

export interface ChannelUsageQueryConfigInput {
  enabled: boolean;
  preset: ChannelUsageQueryPreset;
  baseUrlOverride?: string | null;
  userId?: string | null;
  script: string;
  apiKey?: string | null;
  clearApiKey: boolean;
}

export interface ChannelUsageQueryTestResult {
  status: 'available' | 'warning' | 'exhausted' | 'unknown';
  isValid?: boolean | null;
  invalidMessage?: string | null;
  remaining?: number | null;
  unit?: string | null;
  planName?: string | null;
  total?: number | null;
  used?: number | null;
  extra?: string | null;
}

const CHANNEL_USAGE_QUERY = `
  query ChannelUsageQuery($channelID: ID!) {
    channelUsageQuery(channelID: $channelID) {
      enabled
      preset
      baseUrlOverride
      userId
      script
      apiKeyConfigured
    }
  }
`;

const SAVE_CHANNEL_USAGE_QUERY = `
  mutation SaveChannelUsageQuery($channelID: ID!, $input: ChannelUsageQueryConfigInput!) {
    saveChannelUsageQuery(channelID: $channelID, input: $input) {
      enabled
      preset
      baseUrlOverride
      userId
      script
      apiKeyConfigured
    }
  }
`;

const TEST_CHANNEL_USAGE_QUERY = `
  mutation TestChannelUsageQuery($channelID: ID!, $input: ChannelUsageQueryConfigInput!) {
    testChannelUsageQuery(channelID: $channelID, input: $input) {
      status
      isValid
      invalidMessage
      remaining
      unit
      planName
      total
      used
      extra
    }
  }
`;

export function useChannelUsageQuery(channelID: string, enabled: boolean) {
  return useQuery({
    queryKey: ['channel-usage-query', channelID],
    enabled,
    queryFn: async () => {
      const data = await graphqlRequest<{ channelUsageQuery: ChannelUsageQueryConfig }>(CHANNEL_USAGE_QUERY, { channelID });
      return data.channelUsageQuery;
    },
  });
}

export function useSaveChannelUsageQuery() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ channelID, input }: { channelID: string; input: ChannelUsageQueryConfigInput }) => {
      const data = await graphqlRequest<{ saveChannelUsageQuery: ChannelUsageQueryConfig }>(SAVE_CHANNEL_USAGE_QUERY, {
        channelID,
        input,
      });
      return data.saveChannelUsageQuery;
    },
    onSuccess: (data, variables) => {
      queryClient.setQueryData(['channel-usage-query', variables.channelID], data);
      queryClient.invalidateQueries({ queryKey: ['provider-quotas'] });
    },
  });
}

export function useTestChannelUsageQuery() {
  return useMutation({
    mutationFn: async ({ channelID, input }: { channelID: string; input: ChannelUsageQueryConfigInput }) => {
      const data = await graphqlRequest<{ testChannelUsageQuery: ChannelUsageQueryTestResult }>(TEST_CHANNEL_USAGE_QUERY, {
        channelID,
        input,
      });
      return data.testChannelUsageQuery;
    },
  });
}
