import { format } from 'date-fns';
import { zhCN, enUS } from 'date-fns/locale';
import { ArrowDown, ArrowUp, FileText, KeyRound, Network, Route } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { extractNumberID } from '@/lib/utils';
import { formatDuration } from '@/utils/format-duration';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import type { Request } from '../data/schema';
import { getStatusColor } from './help';

interface RequestMobileCardProps {
  request: Request;
  visibleColumnIds: Set<string>;
  currencyCode: string;
  onViewDetail: (requestId: string) => void;
}

export function RequestMobileCard({ request, visibleColumnIds, currencyCode, onViewDetail }: RequestMobileCardProps) {
  const { t, i18n } = useTranslation();
  const locale = i18n.language.startsWith('zh') ? zhCN : enUS;
  const executions = request.executions?.edges?.flatMap((edge) => (edge.node ? [edge.node] : [])) ?? [];
  const finalExecution = executions[0];
  const channel = finalExecution?.channel ?? request.channel;
  const reasoningEffort = finalExecution?.reasoningEffort ?? request.reasoningEffort;
  const passThroughApplied = executions.some((execution) => execution.passThroughApplied);
  const usage = request.usageLogs?.edges?.[0]?.node;
  const promptTokens = usage?.promptTokens ?? 0;
  const completionTokens = usage?.completionTokens ?? 0;
  const cachedTokens = usage?.promptCachedTokens ?? 0;
  const cacheHitRate = promptTokens > 0 ? (cachedTokens / promptTokens) * 100 : null;

  const detailItems = [
    visibleColumnIds.has('channel')
      ? { id: 'channel', icon: Network, label: t('requests.columns.channel'), value: channel?.name || '-' }
      : null,
    visibleColumnIds.has('caller')
      ? {
          id: 'caller',
          icon: KeyRound,
          label: t('requests.columns.caller'),
          value: request.source === 'api' ? request.apiKey?.name || '-' : t(`requests.source.${request.source}`),
        }
      : null,
    visibleColumnIds.has('apiFormat')
      ? { id: 'apiFormat', icon: Route, label: t('requests.columns.apiFormat'), value: request.format || '-' }
      : null,
    visibleColumnIds.has('cacheHitRate')
      ? {
          id: 'cacheHitRate',
          icon: null,
          label: t('requests.columns.cacheHitRateLabel'),
          value: cacheHitRate == null ? '-' : t('requests.columns.cacheHitRate', { rate: cacheHitRate.toFixed(1) }),
        }
      : null,
    visibleColumnIds.has('cost')
      ? {
          id: 'cost',
          icon: null,
          label: t('requests.columns.cost'),
          value:
            usage?.totalCost == null
              ? '-'
              : t('currencies.format', {
                  val: usage.totalCost,
                  currency: currencyCode,
                  locale: i18n.language.startsWith('zh') ? 'zh-CN' : 'en-US',
                  minimumFractionDigits: 6,
                }),
        }
      : null,
    visibleColumnIds.has('duration')
      ? {
          id: 'duration',
          icon: null,
          label: t('requests.columns.duration'),
          value: request.metricsLatencyMs == null ? '-' : formatDuration(request.metricsLatencyMs),
        }
      : null,
  ].filter((item): item is NonNullable<typeof item> => item !== null);

  return (
    <article className='bg-background rounded-xl border border-[var(--table-border)] p-3 shadow-xs'>
      <div className='flex items-start justify-between gap-3'>
        <div className='min-w-0 flex-1'>
          <div className='flex flex-wrap items-center gap-2'>
            <button
              type='button'
              onClick={() => onViewDetail(request.id)}
              className='min-h-12 rounded-md py-2 font-mono text-sm font-semibold text-green-800 hover:underline dark:text-green-300'
            >
              #{extractNumberID(request.id)}
            </button>
            <Badge className={`${getStatusColor(request.status)} whitespace-nowrap`}>{t(`requests.status.${request.status}`)}</Badge>
          </div>
          <time className='text-muted-foreground block text-xs'>
            {format(new Date(request.createdAt), 'yyyy-MM-dd HH:mm:ss', { locale })}
          </time>
        </div>
        <Button
          type='button'
          variant='ghost'
          size='icon'
          className='h-12 w-12 shrink-0'
          onClick={() => onViewDetail(request.id)}
          aria-label={t('requests.actions.viewDetails')}
        >
          <FileText className='h-4 w-4' />
        </Button>
      </div>

      <div className='mt-3 flex min-w-0 items-center gap-2 border-t pt-3'>
        <span className='min-w-0 truncate font-mono text-sm font-medium'>{request.modelID || t('requests.columns.unknown')}</span>
        {reasoningEffort && (
          <Badge className='shrink-0 border-sky-200 bg-sky-100 text-sky-800 dark:border-sky-800 dark:bg-sky-900/20 dark:text-sky-300'>
            {reasoningEffort}
          </Badge>
        )}
        <Tooltip>
          <TooltipTrigger asChild>
            <span
              className={`ml-auto inline-flex h-12 w-12 shrink-0 items-center justify-center rounded-md ${
                passThroughApplied ? 'text-amber-700 dark:text-amber-300' : 'text-muted-foreground/45'
              }`}
              tabIndex={0}
              role='img'
              aria-label={t(passThroughApplied ? 'requests.tooltips.passThroughApplied' : 'requests.tooltips.passThroughNotApplied')}
            >
              <Route className='h-4 w-4' />
            </span>
          </TooltipTrigger>
          <TooltipContent>
            {t(passThroughApplied ? 'requests.tooltips.passThroughApplied' : 'requests.tooltips.passThroughNotApplied')}
          </TooltipContent>
        </Tooltip>
      </div>

      {usage && (
        <div className='text-muted-foreground mt-2 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs'>
          <span className='inline-flex items-center gap-1' aria-label={t('requests.tooltips.inputTokens')}>
            <ArrowUp className='h-3.5 w-3.5' />
            {promptTokens.toLocaleString()}
          </span>
          <span className='inline-flex items-center gap-1' aria-label={t('requests.tooltips.outputTokens')}>
            <ArrowDown className='h-3.5 w-3.5' />
            {completionTokens.toLocaleString()}
          </span>
        </div>
      )}

      {detailItems.length > 0 && (
        <dl className='bg-muted/25 mt-3 grid grid-cols-2 gap-x-3 gap-y-2 rounded-lg p-3'>
          {detailItems.map((item) => {
            const Icon = item.icon;
            return (
              <div key={item.id} className='min-w-0'>
                <dt className='text-muted-foreground flex items-center gap-1 text-[11px]'>
                  {Icon && <Icon className='h-3 w-3 shrink-0' />}
                  {item.label}
                </dt>
                <dd className='mt-0.5 truncate font-mono text-xs font-medium'>{item.value}</dd>
              </div>
            );
          })}
        </dl>
      )}
    </article>
  );
}
