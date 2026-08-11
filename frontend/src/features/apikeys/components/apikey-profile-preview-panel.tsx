import { useEffect, useMemo, useState } from 'react';
import { IconLoader2, IconSearch } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { ApiKeyProfilePreview } from '../data/schema';
import { selectConversationAPIFormats } from './api-key-profile-preview-formats';

interface ApiKeyProfilePreviewPanelProps {
  profileName?: string;
  preview?: ApiKeyProfilePreview;
  loading?: boolean;
  error?: boolean;
}

export function ApiKeyProfilePreviewPanel({ profileName, preview, loading = false, error = false }: ApiKeyProfilePreviewPanelProps) {
  const { t } = useTranslation();
  const [modelSearch, setModelSearch] = useState('');
  const [selectedApiFormat, setSelectedApiFormat] = useState('');
  const conversationApiFormats = useMemo(() => selectConversationAPIFormats(preview?.apiFormats ?? []), [preview?.apiFormats]);

  useEffect(() => {
    if (conversationApiFormats.length === 0) {
      setSelectedApiFormat('');
    } else if (!conversationApiFormats.includes(selectedApiFormat)) {
      setSelectedApiFormat(conversationApiFormats[0]);
    }
  }, [conversationApiFormats, selectedApiFormat]);

  const visibleModels = useMemo(() => {
    const query = modelSearch.trim().toLowerCase();
    if (!query) return preview?.models ?? [];
    return (preview?.models ?? []).filter((model) => model.id.toLowerCase().includes(query));
  }, [modelSearch, preview?.models]);

  const apiFormatLabel = (apiFormat: string) => {
    const key = `channels.dialogs.fields.apiFormat.formats.${apiFormat}`;
    const label = t(key);
    return label === key ? apiFormat : label;
  };

  const channelsByRoutingOrder = (channels: ApiKeyProfilePreview['models'][number]['channels']) => {
    return [...channels].sort((left, right) => {
      const leftPassThrough = preview?.preferPassThrough === true && left.passThroughApiFormats.includes(selectedApiFormat);
      const rightPassThrough = preview?.preferPassThrough === true && right.passThroughApiFormats.includes(selectedApiFormat);
      if (leftPassThrough !== rightPassThrough) return leftPassThrough ? -1 : 1;
      if (left.orderingWeight !== right.orderingWeight) return right.orderingWeight - left.orderingWeight;
      if (left.name !== right.name) return left.name < right.name ? -1 : 1;
      return left.id - right.id;
    });
  };

  return (
    <aside className='bg-muted/20 flex min-h-0 flex-col rounded-lg border'>
      <div className='shrink-0 space-y-2 border-b p-4'>
        <div className='flex items-center justify-between gap-2'>
          <h3 className='font-medium'>{t('apikeys.profiles.preview.title')}</h3>
          {loading && <IconLoader2 className='text-muted-foreground h-4 w-4 animate-spin' />}
        </div>
        <p className='text-muted-foreground text-xs'>{t('apikeys.profiles.preview.description')}</p>
        {profileName && (
          <Badge variant='outline' className='max-w-full truncate'>
            {t('apikeys.profiles.preview.activeProfile', { name: profileName })}
          </Badge>
        )}
      </div>

      <div className='min-h-0 flex-1 space-y-5 overflow-y-auto p-4'>
        {error ? (
          <p className='text-destructive text-sm'>{t('apikeys.profiles.preview.error')}</p>
        ) : !preview && loading ? (
          <p className='text-muted-foreground text-sm'>{t('apikeys.profiles.preview.loading')}</p>
        ) : (
          <>
            <section className='space-y-2'>
              <div className='flex items-center justify-between gap-2'>
                <h4 className='text-sm font-medium'>{t('apikeys.profiles.preview.models')}</h4>
                <Badge variant='secondary'>{preview?.models.length ?? 0}</Badge>
              </div>
              {conversationApiFormats.length > 0 ? (
                <Tabs value={selectedApiFormat} onValueChange={setSelectedApiFormat} className='gap-0'>
                  <TabsList className='grid h-auto w-full grid-cols-4 rounded-none border-b bg-transparent p-0'>
                    {conversationApiFormats.map((apiFormat) => (
                      <TabsTrigger
                        key={apiFormat}
                        value={apiFormat}
                        className='text-muted-foreground data-[state=active]:text-foreground h-auto min-w-0 rounded-none border-0 border-b-2 border-transparent bg-transparent px-2 py-2 text-center text-xs leading-tight whitespace-normal shadow-none data-[state=active]:border-current data-[state=active]:bg-transparent data-[state=active]:shadow-none dark:data-[state=active]:text-foreground'
                      >
                        {apiFormatLabel(apiFormat)}
                      </TabsTrigger>
                    ))}
                  </TabsList>
                </Tabs>
              ) : (
                <p className='text-muted-foreground text-xs'>{t('apikeys.profiles.preview.noApis')}</p>
              )}
              {(preview?.models.length ?? 0) > 0 && (
                <div className='relative'>
                  <IconSearch className='text-muted-foreground absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2' />
                  <Input
                    value={modelSearch}
                    onChange={(event) => setModelSearch(event.target.value)}
                    placeholder={t('apikeys.profiles.preview.searchModels')}
                    className='h-8 pl-8'
                  />
                </div>
              )}
              {selectedApiFormat && (
                <p className='text-muted-foreground text-xs'>
                  {t('apikeys.profiles.preview.routingOrderHint', { api: apiFormatLabel(selectedApiFormat) })}
                </p>
              )}
              {visibleModels.length > 0 ? (
                <div className='flex flex-wrap gap-1.5'>
                  {visibleModels.map((model) => (
                    <Tooltip key={model.id}>
                      <TooltipTrigger asChild>
                        <button
                          type='button'
                          className='hover:bg-accent rounded-md border px-2 py-1 text-left font-mono text-xs transition-colors'
                        >
                          {model.id}
                        </button>
                      </TooltipTrigger>
                      <TooltipContent side='left' className='max-w-72'>
                        <p className='mb-1 text-xs font-medium'>{t('apikeys.profiles.preview.supportedChannels')}</p>
                        {model.channels.length > 0 ? (
                          <div className='space-y-0.5'>
                            {channelsByRoutingOrder(model.channels).map((channel, index) => {
                              const passThroughPreferred =
                                preview?.preferPassThrough === true && channel.passThroughApiFormats.includes(selectedApiFormat);
                              return (
                                <div key={channel.id} className='flex items-center gap-1.5 text-xs'>
                                  <span className='text-background/65 w-4 text-right font-mono'>{index + 1}.</span>
                                  <span className='min-w-0 flex-1 truncate'>{channel.name}</span>
                                  {passThroughPreferred && (
                                    <span className='rounded border border-emerald-400/50 px-1 text-[10px] text-emerald-300'>
                                      {t('apikeys.profiles.preview.passThroughPreferred')}
                                    </span>
                                  )}
                                  <span className='text-background/65 font-mono text-[10px]'>
                                    {t('apikeys.profiles.preview.channelWeight', { weight: channel.orderingWeight })}
                                  </span>
                                </div>
                              );
                            })}
                          </div>
                        ) : (
                          <p className='text-xs'>{t('apikeys.profiles.preview.noChannels')}</p>
                        )}
                      </TooltipContent>
                    </Tooltip>
                  ))}
                </div>
              ) : (
                <p className='text-muted-foreground text-xs'>
                  {(preview?.models.length ?? 0) > 0
                    ? t('apikeys.profiles.preview.noSearchResults')
                    : t('apikeys.profiles.preview.noModels')}
                </p>
              )}
            </section>
          </>
        )}
      </div>
    </aside>
  );
}
