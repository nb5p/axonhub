import type { ApiKeyProfile } from '../data/schema';

const unrestrictedAPIKeyProfilePreview: ApiKeyProfile = {
  name: '__unrestricted_preview__',
  modelMappings: [],
  channelIDs: [],
  channelTags: [],
  channelTagsMatchMode: 'any',
  modelIDs: [],
  loadBalanceStrategy: 'default',
  traceStickyMode: 'default',
};

export function resolveAPIKeyProfilePreview(activeProfile: ApiKeyProfile | undefined): ApiKeyProfile {
  return activeProfile ?? unrestrictedAPIKeyProfilePreview;
}
