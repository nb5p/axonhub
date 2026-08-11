'use client';

import { cloneElement, useMemo } from 'react';
import { CheckIcon, PlusCircledIcon } from '@radix-ui/react-icons';
import { Loader2 } from 'lucide-react';
import { ActivityCalendar, type Activity } from 'react-activity-calendar';
import 'react-activity-calendar/tooltips.css';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { formatNumber } from '@/utils/format-number';
import { usePersistedFilter } from '@/hooks/use-persisted-filter';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandSeparator } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Separator } from '@/components/ui/separator';
import { Skeleton } from '@/components/ui/skeleton';
import { useGeneralSettings } from '../../system/data/system';
import { useAPIKeyActivityHeatmap, type APIKeyActivityHeatmapBucket } from '../data/dashboard';

const HEATMAP_DAYS = 90;
const MAX_LEVEL = 4;
const MAX_TOOLTIP_CONTRIBUTORS = 5;

interface APIKeyGroup {
  apiKeyId: string;
  apiKeyName: string;
  buckets: APIKeyActivityHeatmapBucket[];
  totalRequests: number;
}

interface AggregatedBucket {
  date: string;
  requestCount: number;
  totalTokens: number;
  cost: number;
  contributions: Array<{
    apiKeyName: string;
    requestCount: number;
  }>;
}

function getHeatmapRange() {
  const end = new Date();
  const endDay = new Date(end.getFullYear(), end.getMonth(), end.getDate() + 1);
  const startDay = new Date(endDay);
  startDay.setDate(startDay.getDate() - HEATMAP_DAYS);

  return {
    startDate: startDay,
    endDate: endDay,
  };
}

function getActivityLevel(count: number, maxCount: number) {
  if (count <= 0 || maxCount <= 0) return 0;
  return Math.max(1, Math.ceil((count / maxCount) * MAX_LEVEL));
}

function getCalendarTheme() {
  return {
    light: [
      'var(--muted)',
      'color-mix(in oklab, var(--chart-2) 35%, var(--muted))',
      'color-mix(in oklab, var(--chart-2) 55%, var(--muted))',
      'color-mix(in oklab, var(--chart-2) 75%, var(--muted))',
      'var(--chart-2)',
    ],
    dark: [
      'var(--muted)',
      'color-mix(in oklab, var(--chart-2) 35%, var(--muted))',
      'color-mix(in oklab, var(--chart-2) 55%, var(--muted))',
      'color-mix(in oklab, var(--chart-2) 75%, var(--muted))',
      'var(--chart-2)',
    ],
  };
}

export function ApiKeyActivityHeatmap() {
  const { t, i18n } = useTranslation();
  const range = useMemo(getHeatmapRange, []);
  const { data, isLoading, isFetching, error } = useAPIKeyActivityHeatmap({
    createdAtGTE: range.startDate.toISOString(),
    createdAtLT: range.endDate.toISOString(),
  });
  const { data: generalSettings } = useGeneralSettings();
  const [storedSelectedApiKeyIds, setStoredSelectedApiKeyIds] = usePersistedFilter<string[] | null>(
    'dashboard',
    'api-key-activity-keys',
    null
  );

  const currencyCode = generalSettings?.currencyCode || 'USD';
  const locale = i18n.language.startsWith('zh') ? 'zh-CN' : 'en-US';
  const formatCurrency = (val: number) =>
    t('currencies.format', {
      val,
      currency: currencyCode,
      locale,
      minimumFractionDigits: 4,
      maximumFractionDigits: 4,
    });

  const apiKeyGroups = useMemo<APIKeyGroup[]>(() => {
    const grouped = new Map<string, Omit<APIKeyGroup, 'totalRequests'>>();
    for (const item of data ?? []) {
      if (!grouped.has(item.apiKeyId)) {
        grouped.set(item.apiKeyId, {
          apiKeyId: item.apiKeyId,
          apiKeyName: item.apiKeyName,
          buckets: [],
        });
      }
      grouped.get(item.apiKeyId)?.buckets.push(item);
    }

    return Array.from(grouped.values())
      .map((group) => ({
        ...group,
        buckets: group.buckets.sort((a, b) => a.date.localeCompare(b.date)),
        totalRequests: group.buckets.reduce((sum, bucket) => sum + bucket.requestCount, 0),
      }))
      .sort((a, b) => b.totalRequests - a.totalRequests || a.apiKeyName.localeCompare(b.apiKeyName));
  }, [data]);

  const allApiKeyIds = useMemo(() => apiKeyGroups.map((group) => group.apiKeyId), [apiKeyGroups]);
  const selectedApiKeyIds = useMemo(() => {
    const availableIds = new Set(allApiKeyIds);
    return new Set(storedSelectedApiKeyIds === null ? allApiKeyIds : storedSelectedApiKeyIds.filter((id) => availableIds.has(id)));
  }, [allApiKeyIds, storedSelectedApiKeyIds]);

  const orderedApiKeyGroups = useMemo(
    () =>
      [...apiKeyGroups].sort(
        (a, b) => Number(selectedApiKeyIds.has(b.apiKeyId)) - Number(selectedApiKeyIds.has(a.apiKeyId)) || b.totalRequests - a.totalRequests
      ),
    [apiKeyGroups, selectedApiKeyIds]
  );

  const aggregatedBuckets = useMemo<AggregatedBucket[]>(() => {
    const bucketsByDate = new Map<string, AggregatedBucket>();

    for (const item of data ?? []) {
      if (!selectedApiKeyIds.has(item.apiKeyId)) continue;

      const bucket = bucketsByDate.get(item.date) ?? {
        date: item.date,
        requestCount: 0,
        totalTokens: 0,
        cost: 0,
        contributions: [],
      };
      bucket.requestCount += item.requestCount;
      bucket.totalTokens += item.totalTokens;
      bucket.cost += item.cost;
      if (item.requestCount > 0) {
        bucket.contributions.push({
          apiKeyName: item.apiKeyName,
          requestCount: item.requestCount,
        });
      }
      bucketsByDate.set(item.date, bucket);
    }

    return Array.from(bucketsByDate.values())
      .map((bucket) => ({
        ...bucket,
        contributions: bucket.contributions.sort((a, b) => b.requestCount - a.requestCount || a.apiKeyName.localeCompare(b.apiKeyName)),
      }))
      .sort((a, b) => a.date.localeCompare(b.date));
  }, [data, selectedApiKeyIds]);

  const activityByDate = useMemo(() => new Map(aggregatedBuckets.map((bucket) => [bucket.date, bucket])), [aggregatedBuckets]);
  const maxRequestCount = useMemo(() => Math.max(0, ...aggregatedBuckets.map((bucket) => bucket.requestCount)), [aggregatedBuckets]);
  const totalRequests = useMemo(() => aggregatedBuckets.reduce((sum, bucket) => sum + bucket.requestCount, 0), [aggregatedBuckets]);
  const activities: Activity[] = aggregatedBuckets.map((bucket) => ({
    date: bucket.date,
    count: bucket.requestCount,
    level: getActivityLevel(bucket.requestCount, maxRequestCount),
  }));

  const toggleApiKey = (apiKeyId: string) => {
    setStoredSelectedApiKeyIds((current) => {
      const next = new Set(current === null ? allApiKeyIds : current);
      if (next.has(apiKeyId)) {
        next.delete(apiKeyId);
      } else {
        next.add(apiKeyId);
      }
      return Array.from(next);
    });
  };

  if (isLoading) {
    return (
      <div className='flex h-[360px] items-center justify-center'>
        <Skeleton className='h-[300px] w-full rounded-md' />
      </div>
    );
  }

  if (error) {
    return (
      <div className='flex h-[300px] items-center justify-center'>
        <div className='text-sm text-red-500'>
          {t('dashboard.charts.errorLoadingAPIKeyActivity')} {error.message}
        </div>
      </div>
    );
  }

  if (apiKeyGroups.length === 0) {
    return (
      <div className='flex h-[300px] items-center justify-center'>
        <div className='text-muted-foreground text-sm'>{t('dashboard.charts.noAPIKeyActivityData')}</div>
      </div>
    );
  }

  return (
    <div className='relative space-y-4'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant='outline' size='sm' className='h-8 border-dashed'>
              <PlusCircledIcon className='size-4' />
              {t('dashboard.charts.apiKeyActivityFilter')}
              <Separator orientation='vertical' className='mx-1 h-4' />
              <Badge variant='secondary' className='rounded-sm px-1 font-normal'>
                {t('dashboard.charts.apiKeyActivitySelected', { count: selectedApiKeyIds.size })}
              </Badge>
            </Button>
          </PopoverTrigger>
          <PopoverContent className='w-[300px] p-0' align='start'>
            <Command>
              <CommandInput placeholder={t('dashboard.charts.apiKeyActivitySearch')} />
              <CommandList>
                <CommandEmpty>{t('common.noResultsFound')}</CommandEmpty>
                <CommandGroup>
                  {orderedApiKeyGroups.map((group) => {
                    const isSelected = selectedApiKeyIds.has(group.apiKeyId);
                    return (
                      <CommandItem
                        key={group.apiKeyId}
                        value={`${group.apiKeyName} ${group.apiKeyId}`}
                        onSelect={() => toggleApiKey(group.apiKeyId)}
                      >
                        <div
                          className={cn(
                            'border-primary flex size-4 items-center justify-center rounded-sm border',
                            isSelected ? 'bg-primary text-primary-foreground' : 'opacity-50 [&_svg]:invisible'
                          )}
                        >
                          <CheckIcon className='size-4' />
                        </div>
                        <span className='min-w-0 flex-1 truncate' title={group.apiKeyName}>
                          {group.apiKeyName}
                        </span>
                        <span className='text-muted-foreground ml-auto font-mono text-xs'>{formatNumber(group.totalRequests)}</span>
                      </CommandItem>
                    );
                  })}
                </CommandGroup>
                <CommandSeparator />
                <CommandGroup>
                  <CommandItem onSelect={() => setStoredSelectedApiKeyIds(null)} className='justify-center text-center'>
                    {t('dashboard.charts.apiKeyActivitySelectAll')}
                  </CommandItem>
                  <CommandItem onSelect={() => setStoredSelectedApiKeyIds([])} className='justify-center text-center'>
                    {t('dashboard.charts.apiKeyActivityClearSelection')}
                  </CommandItem>
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>

        {selectedApiKeyIds.size > 0 && (
          <div className='text-muted-foreground text-xs'>
            {t('dashboard.charts.apiKeyActivityTotalRequests', { count: formatNumber(totalRequests) })}
          </div>
        )}
      </div>

      {selectedApiKeyIds.size === 0 ? (
        <div className='flex h-[220px] flex-col items-center justify-center gap-3 rounded-lg border border-dashed'>
          <div className='text-muted-foreground text-sm'>{t('dashboard.charts.apiKeyActivityEmptySelection')}</div>
          <Button variant='outline' size='sm' onClick={() => setStoredSelectedApiKeyIds(null)}>
            {t('dashboard.charts.apiKeyActivitySelectAll')}
          </Button>
        </div>
      ) : (
        <div className='overflow-x-auto pb-2'>
          <div className='bg-muted/10 min-w-[720px] rounded-lg border p-4'>
            <ActivityCalendar
              data={activities}
              blockMargin={3}
              blockRadius={3}
              blockSize={11}
              fontSize={11}
              maxLevel={MAX_LEVEL}
              showColorLegend={false}
              showTotalCount={false}
              showWeekdayLabels={['mon', 'wed', 'fri']}
              theme={getCalendarTheme()}
              tooltips={{
                activity: {
                  text: (activity) => {
                    const bucket = activityByDate.get(activity.date);
                    if (!bucket) return activity.date;

                    const contributionLines = bucket.contributions.slice(0, MAX_TOOLTIP_CONTRIBUTORS).map((contribution) =>
                      t('dashboard.charts.apiKeyActivityContribution', {
                        name: contribution.apiKeyName,
                        requests: formatNumber(contribution.requestCount),
                      })
                    );
                    const remainingContributors = bucket.contributions.length - contributionLines.length;
                    if (remainingContributors > 0) {
                      contributionLines.push(t('dashboard.charts.apiKeyActivityMoreContributors', { count: remainingContributors }));
                    }
                    const breakdown = contributionLines.length
                      ? `\n${t('dashboard.charts.apiKeyActivityBreakdown')}\n${contributionLines.join('\n')}`
                      : '';

                    return t('dashboard.charts.apiKeyActivityTooltip', {
                      date: activity.date,
                      keys: selectedApiKeyIds.size,
                      requests: formatNumber(bucket.requestCount),
                      tokens: formatNumber(bucket.totalTokens),
                      cost: formatCurrency(bucket.cost),
                      breakdown,
                    });
                  },
                },
              }}
              renderBlock={(block, activity) =>
                cloneElement(block, {
                  'aria-label': t('dashboard.charts.apiKeyActivityAriaLabel', {
                    date: activity.date,
                    keys: selectedApiKeyIds.size,
                    count: activity.count,
                  }),
                })
              }
            />
          </div>
        </div>
      )}

      {isFetching && (
        <div className='bg-background/50 absolute inset-0 flex items-center justify-center'>
          <Loader2 className='text-muted-foreground size-6 animate-spin' />
        </div>
      )}
    </div>
  );
}
