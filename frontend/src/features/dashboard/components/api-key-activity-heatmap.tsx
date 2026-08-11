'use client';

import { cloneElement, useMemo, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { ActivityCalendar, type Activity } from 'react-activity-calendar';
import 'react-activity-calendar/tooltips.css';
import { useTranslation } from 'react-i18next';
import { formatNumber } from '@/utils/format-number';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useGeneralSettings } from '../../system/data/system';
import { useAPIKeyActivityHeatmap, type APIKeyActivityHeatmapBucket } from '../data/dashboard';

const HEATMAP_DAYS = 90;
const MAX_LEVEL = 4;

function dateOnly(date: Date) {
  return date.toISOString().slice(0, 10);
}

function getHeatmapRange() {
  const end = new Date();
  const endDay = new Date(end.getFullYear(), end.getMonth(), end.getDate() + 1);
  const startDay = new Date(endDay);
  startDay.setDate(startDay.getDate() - HEATMAP_DAYS);

  return {
    start: dateOnly(startDay),
    end: dateOnly(endDay),
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
  const [hiddenApiKeyIds, setHiddenApiKeyIds] = useState<Set<string>>(() => new Set());

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

  const apiKeyGroups = useMemo(() => {
    const grouped = new Map<string, { apiKeyId: string; apiKeyName: string; buckets: APIKeyActivityHeatmapBucket[] }>();
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

  const maxRequestCount = useMemo(() => {
    return Math.max(0, ...(data ?? []).map((item) => item.requestCount));
  }, [data]);

  const visibleGroups = apiKeyGroups.filter((group) => !hiddenApiKeyIds.has(group.apiKeyId));
  const hasHiddenKeys = hiddenApiKeyIds.size > 0;

  const toggleApiKey = (apiKeyId: string) => {
    setHiddenApiKeyIds((current) => {
      const next = new Set(current);
      if (next.has(apiKeyId)) {
        next.delete(apiKeyId);
      } else {
        next.add(apiKeyId);
      }
      return next;
    });
  };

  const showOnlyApiKey = (apiKeyId: string) => {
    setHiddenApiKeyIds(new Set(apiKeyGroups.filter((group) => group.apiKeyId !== apiKeyId).map((group) => group.apiKeyId)));
  };

  const activityByKey = useMemo(() => {
    const mapped = new Map<string, APIKeyActivityHeatmapBucket>();
    for (const item of data ?? []) {
      mapped.set(`${item.apiKeyId}:${item.date}`, item);
    }
    return mapped;
  }, [data]);

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
    <div className='relative space-y-5'>
      <div className='flex flex-wrap items-center gap-2'>
        {apiKeyGroups.map((group) => {
          const hidden = hiddenApiKeyIds.has(group.apiKeyId);
          return (
            <div key={group.apiKeyId} className='bg-background flex items-center rounded-full border shadow-xs'>
              <button
                type='button'
                className={`max-w-[220px] truncate rounded-l-full px-3 py-1 text-xs transition-colors ${
                  hidden ? 'text-muted-foreground hover:bg-accent line-through' : 'text-foreground hover:bg-accent'
                }`}
                onClick={() => toggleApiKey(group.apiKeyId)}
                title={group.apiKeyName}
              >
                {group.apiKeyName}
              </button>
              <button
                type='button'
                className='text-muted-foreground hover:bg-accent hover:text-foreground border-l px-2 py-1 text-[10px] transition-colors'
                onClick={() => showOnlyApiKey(group.apiKeyId)}
              >
                {t('dashboard.charts.apiKeyActivityOnly')}
              </button>
            </div>
          );
        })}
        {hasHiddenKeys && (
          <Button variant='ghost' size='sm' className='h-7 text-xs' onClick={() => setHiddenApiKeyIds(new Set())}>
            {t('dashboard.charts.apiKeyActivityShowAll')}
          </Button>
        )}
      </div>

      {visibleGroups.length === 0 ? (
        <div className='flex h-[220px] items-center justify-center rounded-lg border border-dashed'>
          <Button variant='outline' size='sm' onClick={() => setHiddenApiKeyIds(new Set())}>
            {t('dashboard.charts.apiKeyActivityShowAll')}
          </Button>
        </div>
      ) : (
        <div className='space-y-5 overflow-x-auto pb-2'>
          {visibleGroups.map((group) => {
            const activities: Activity[] = group.buckets.map((bucket) => ({
              date: bucket.date,
              count: bucket.requestCount,
              level: getActivityLevel(bucket.requestCount, maxRequestCount),
            }));

            return (
              <div key={group.apiKeyId} className='bg-muted/10 min-w-[720px] rounded-lg border p-4'>
                <div className='mb-3 flex flex-wrap items-center justify-between gap-2'>
                  <div className='min-w-0'>
                    <div className='truncate text-sm font-medium'>{group.apiKeyName}</div>
                    <div className='text-muted-foreground text-xs'>
                      {t('dashboard.charts.apiKeyActivityTotalRequests', { count: formatNumber(group.totalRequests) })}
                    </div>
                  </div>
                  <Button variant='ghost' size='sm' className='h-7 text-xs' onClick={() => toggleApiKey(group.apiKeyId)}>
                    {t('dashboard.charts.apiKeyActivityHide')}
                  </Button>
                </div>
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
                        const bucket = activityByKey.get(`${group.apiKeyId}:${activity.date}`);
                        if (!bucket) return activity.date;
                        return t('dashboard.charts.apiKeyActivityTooltip', {
                          name: group.apiKeyName,
                          date: activity.date,
                          requests: formatNumber(bucket.requestCount),
                          tokens: formatNumber(bucket.totalTokens),
                          cost: formatCurrency(bucket.cost),
                        });
                      },
                    },
                  }}
                  renderBlock={(block, activity) =>
                    cloneElement(block, {
                      'aria-label': t('dashboard.charts.apiKeyActivityAriaLabel', {
                        name: group.apiKeyName,
                        date: activity.date,
                        count: activity.count,
                      }),
                    })
                  }
                />
              </div>
            );
          })}
        </div>
      )}

      {isFetching && (
        <div className='bg-background/50 absolute inset-0 flex items-center justify-center'>
          <Loader2 className='text-muted-foreground h-6 w-6 animate-spin' />
        </div>
      )}
    </div>
  );
}
