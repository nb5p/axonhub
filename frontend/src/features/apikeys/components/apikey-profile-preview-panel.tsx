import { useMemo, useState } from 'react';
import { IconLoader2, IconSearch } from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { ApiKeyProfilePreview } from '../data/schema';

interface ApiKeyProfilePreviewPanelProps {
  profileName?: string;
  preview?: ApiKeyProfilePreview;
  loading?: boolean;
  error?: boolean;
}

export function ApiKeyProfilePreviewPanel({ profileName, preview, loading = false, error = false }: ApiKeyProfilePreviewPanelProps) {
  const { t } = useTranslation();
  const [modelSearch, setModelSearch] = useState('');

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
                <h4 className='text-sm font-medium'>{t('apikeys.profiles.preview.apis')}</h4>
                <Badge variant='secondary'>{preview?.apiFormats.length ?? 0}</Badge>
              </div>
              {(preview?.apiFormats.length ?? 0) > 0 ? (
                <div className='flex flex-wrap gap-1.5'>
                  {preview?.apiFormats.map((apiFormat) => (
                    <Badge key={apiFormat} variant='outline' className='font-normal'>
                      {apiFormatLabel(apiFormat)}
                    </Badge>
                  ))}
                </div>
              ) : (
                <p className='text-muted-foreground text-xs'>{t('apikeys.profiles.preview.noApis')}</p>
              )}
            </section>

            <section className='space-y-2'>
              <div className='flex items-center justify-between gap-2'>
                <h4 className='text-sm font-medium'>{t('apikeys.profiles.preview.models')}</h4>
                <Badge variant='secondary'>{preview?.models.length ?? 0}</Badge>
              </div>
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
                            {model.channels.map((channel) => (
                              <p key={channel.id} className='text-xs'>
                                {channel.name}
                              </p>
                            ))}
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
