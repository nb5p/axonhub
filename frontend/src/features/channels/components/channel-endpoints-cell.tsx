import { memo } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { Channel, ChannelEndpoint } from '../data/schema';

const PRIMARY_ENDPOINTS = [
  {
    apiFormat: 'openai/chat_completions',
    abbreviation: 'C',
    className: 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
  },
  {
    apiFormat: 'openai/responses',
    abbreviation: 'R',
    className: 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300',
  },
  {
    apiFormat: 'anthropic/messages',
    abbreviation: 'M',
    className: 'border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-300',
  },
  {
    apiFormat: 'gemini/contents',
    abbreviation: 'G',
    className: 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300',
  },
] as const;

function resolveEndpointFormats(defaultEndpoints: ChannelEndpoint[], configuredEndpoints: ChannelEndpoint[]): string[] {
  const formats = defaultEndpoints.map((endpoint) => endpoint.apiFormat);
  const seen = new Set(formats);

  for (const endpoint of configuredEndpoints) {
    if (!seen.has(endpoint.apiFormat)) {
      formats.push(endpoint.apiFormat);
      seen.add(endpoint.apiFormat);
    }
  }

  return formats.filter(Boolean);
}

export const ChannelEndpointsCell = memo(({ channel }: { channel: Channel }) => {
  const { t } = useTranslation();
  const endpointFormats = resolveEndpointFormats(channel.defaultEndpoints ?? [], channel.endpoints ?? []);
  const supportedFormats = new Set(endpointFormats);
  const primaryEndpoints = PRIMARY_ENDPOINTS.filter((endpoint) => supportedFormats.has(endpoint.apiFormat));
  const primaryFormats = new Set<string>(PRIMARY_ENDPOINTS.map((endpoint) => endpoint.apiFormat));
  const remainingEndpoints = endpointFormats.filter((apiFormat) => !primaryFormats.has(apiFormat));

  const apiFormatLabel = (apiFormat: string) => {
    const key = `channels.dialogs.fields.apiFormat.formats.${apiFormat}`;
    const label = t(key);
    return label === key ? apiFormat : label;
  };

  if (endpointFormats.length === 0) {
    return <span className='text-muted-foreground text-xs'>-</span>;
  }

  return (
    <div className='flex items-center justify-center gap-1'>
      {primaryEndpoints.map((endpoint) => (
        <Tooltip key={endpoint.apiFormat}>
          <TooltipTrigger asChild>
            <span
              tabIndex={0}
              aria-label={apiFormatLabel(endpoint.apiFormat)}
              className={cn(
                'inline-flex h-6 min-w-8 cursor-help items-center justify-center rounded-md border px-2 font-mono text-xs font-semibold',
                endpoint.className
              )}
            >
              {endpoint.abbreviation}
            </span>
          </TooltipTrigger>
          <TooltipContent>{apiFormatLabel(endpoint.apiFormat)}</TooltipContent>
        </Tooltip>
      ))}

      {remainingEndpoints.length > 0 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <span
              tabIndex={0}
              aria-label={t('channels.endpoints.moreSupported', { count: remainingEndpoints.length })}
              className='bg-muted text-muted-foreground inline-flex h-6 min-w-8 cursor-help items-center justify-center rounded-md border px-2 text-xs font-semibold'
            >
              …
            </span>
          </TooltipTrigger>
          <TooltipContent side='left' className='max-w-80'>
            <p className='mb-1 font-medium'>{t('channels.endpoints.moreSupported', { count: remainingEndpoints.length })}</p>
            <div className='space-y-0.5'>
              {remainingEndpoints.map((apiFormat) => (
                <p key={apiFormat}>{apiFormatLabel(apiFormat)}</p>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      )}
    </div>
  );
});

ChannelEndpointsCell.displayName = 'ChannelEndpointsCell';
