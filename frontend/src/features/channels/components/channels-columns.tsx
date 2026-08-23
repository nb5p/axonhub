import { useCallback, useState, memo, useRef, useEffect, useLayoutEffect } from 'react';
import { format } from 'date-fns';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { ColumnDef, Row, Table } from '@tanstack/react-table';
import {
  IconPlayerPlay,
  IconChevronDown,
  IconChevronRight,
  IconAlertTriangle,
  IconEdit,
  IconArchive,
  IconTrash,
  IconCheck,
  IconWeight,
  IconTransform,
  IconNetwork,
  IconAdjustments,
  IconRoute,
  IconCopy,
  IconCoin,
  IconLoader2,
  IconKey,
  IconKeyOff,
  IconGauge,
  IconHistory,
  IconPlugConnected,
  IconClockPlay,
  IconCircleOff,
  IconSearch,
  IconRefresh,
} from '@tabler/icons-react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { cn, extractNumberID } from '@/lib/utils';
import { formatUsageQueryBalance } from '@/lib/usage-query-balance';
import { usePermissions } from '@/hooks/usePermissions';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Switch } from '@/components/ui/switch';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { DataTableColumnHeader } from '@/components/data-table-column-header';
import { useChannels } from '../context/channels-context';
import { useTestChannel, useUpdateChannel } from '../data/channels';
import { getDefaultAvailableChannelTestAPIFormats } from '../data/channel-test-api-formats';
import { useRefreshAllUsageQueries, useRefreshChannelUsageQuery } from '../data/usage-query';
import { CHANNEL_CONFIGS, getProvider } from '../data/config_channels';
import { Channel } from '../data/schema';
import type { ProviderUsageQueryQuotaData } from '@/features/system/data/quotas';
import { ChannelEndpointsCell } from './channel-endpoints-cell';
import { ChannelHealthCell } from './channel-health-cell';
import { ChannelLimiterCell } from './channel-limiter-cell';
import { ChannelsStatusDialog } from './channels-status-dialog';

const WEIGHT_PRECISION = 4;
const MIN_WEIGHT = 0;
const MAX_WEIGHT = 100;

const formatWeight = (value: number) => Number(value.toFixed(WEIGHT_PRECISION));
const clampWeight = (value: number) => formatWeight(Math.min(MAX_WEIGHT, Math.max(MIN_WEIGHT, value)));

// Status Switch Cell Component to handle status toggle with confirmation dialog
const StatusSwitchCell = memo(({ row }: { row: Row<Channel> }) => {
  const channel = row.original;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { channelPermissions } = usePermissions();

  const isEnabled = channel.status === 'enabled';
  const isArchived = channel.status === 'archived';

  const handleSwitchClick = useCallback(() => {
    if (!isArchived) {
      setDialogOpen(true);
    }
  }, [isArchived]);

  if (!channelPermissions.canWrite) {
    return <Badge variant='outline'>{channel.status}</Badge>;
  }

  return (
    <div className='flex justify-center'>
      <Switch checked={isEnabled} onCheckedChange={handleSwitchClick} disabled={isArchived} data-testid='channel-status-switch' />
      {dialogOpen && <ChannelsStatusDialog open={dialogOpen} onOpenChange={setDialogOpen} currentRow={channel} />}
    </div>
  );
});

StatusSwitchCell.displayName = 'StatusSwitchCell';

// Action Cell Component to handle hooks properly
const ActionCell = memo(({ row }: { row: Row<Channel> }) => {
  const { t } = useTranslation();
  const channel = row.original;
  const { setOpen, setCurrentRow } = useChannels();
  const { channelPermissions } = usePermissions();
  const testChannel = useTestChannel();
  const isArchived = channel.status === 'archived';
  const hasError = !!channel.errorMessage;
  const hasDisabledAPIKeys = channelPermissions.canWrite && (channel.disabledAPIKeys?.length ?? 0) > 0;
  const defaultTestFormat = getDefaultAvailableChannelTestAPIFormats(
    new Set([...(channel.defaultEndpoints ?? []), ...(channel.endpoints ?? [])].map((endpoint) => endpoint.apiFormat))
  )[0];

  const handleDefaultTest = async () => {
    if (!defaultTestFormat) {
      toast.error(t('channels.dialogs.test.formatUnavailable'));
      return;
    }

    try {
      await testChannel.mutateAsync({
        channelID: channel.id,
        modelID: channel.defaultTestModel || undefined,
        apiFormat: defaultTestFormat,
      });
    } catch (_error) {}
  };

  const handleOpenTestDialog = useCallback(() => {
    setCurrentRow(channel);
    setOpen('test');
  }, [channel, setCurrentRow, setOpen]);

  const handleEdit = useCallback(() => {
    setCurrentRow(channel);
    setOpen('edit');
  }, [channel, setCurrentRow, setOpen]);

  return (
    <div className='flex items-center justify-center gap-1'>
      <Button size='sm' variant='outline' className='h-8 w-8 p-0' onClick={handleEdit}>
        <IconEdit className='h-3 w-3' />
      </Button>
      <Button size='sm' variant='outline' className='h-8 px-3' onClick={handleDefaultTest} disabled={testChannel.isPending || !defaultTestFormat}>
        <IconPlayerPlay className='mr-1 h-3 w-3' />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button size='sm' variant='outline' className='h-8 w-8 p-0' data-testid='row-actions'>
            <DotsHorizontalIcon className='h-3 w-3' />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end' className='w-[160px]'>
          <DropdownMenuItem onClick={handleOpenTestDialog}>
            <IconPlayerPlay size={16} className='mr-2' />
            {t('channels.actions.test')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('testHistory');
            }}
          >
            <IconHistory size={16} className='mr-2' />
            {t('channels.actions.testHistory')}
          </DropdownMenuItem>
          <DropdownMenuSeparator />

          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('duplicate');
            }}
          >
            <IconCopy size={16} className='mr-2' />
            {t('common.actions.duplicate')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('modelMapping');
            }}
          >
            <IconRoute size={16} className='mr-2' />
            {t('channels.dialogs.settings.modelMapping.title')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('price');
            }}
          >
            <IconCoin size={16} className='mr-2' />
            {t('channels.actions.modelPrice')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('overrides');
            }}
          >
            <IconAdjustments size={16} className='mr-2' />
            {t('channels.dialogs.settings.overrides.action')}
          </DropdownMenuItem>

          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('proxy');
            }}
          >
            <IconNetwork size={16} className='mr-2' />
            {t('channels.dialogs.proxy.action')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('transformOptions');
            }}
          >
            <IconTransform size={16} className='mr-2' />
            {t('channels.dialogs.transformOptions.action')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('rateLimit');
            }}
          >
            <IconGauge size={16} className='mr-2' />
            {t('channels.dialogs.rateLimit.action')}
          </DropdownMenuItem>
          {channel.type !== 'xai_subscription' && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('endpoints');
              }}
            >
              <IconPlugConnected size={16} className='mr-2' />
              {t('channels.endpoints.title')}
            </DropdownMenuItem>
          )}
          {channelPermissions.canWrite && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('modelFormatDisables');
              }}
            >
              <IconCircleOff size={16} className='mr-2' />
              {t('channels.dialogs.modelFormatDisables.action')}
            </DropdownMenuItem>
          )}
          {channelPermissions.canWrite && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('usageQuery');
              }}
            >
              <IconSearch size={16} className='mr-2' />
              {t('channels.dialogs.usageQuery.action')}
            </DropdownMenuItem>
          )}
          {channelPermissions.canWrite && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('keyManagement');
              }}
            >
              <IconKey size={16} className='mr-2' />
              {t('channels.actions.keyManagement')}
            </DropdownMenuItem>
          )}
          {channelPermissions.canWrite && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('availability');
              }}
            >
              <IconClockPlay size={16} className='mr-2' />
              {t('channels.dialogs.availability.action')}
            </DropdownMenuItem>
          )}
          {hasDisabledAPIKeys && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('disabledAPIKeys');
              }}
              className='text-orange-500!'
            >
              <IconKeyOff size={16} className='mr-2' />
              {t('channels.actions.disabledAPIKeys', { count: channel.disabledAPIKeys?.length ?? 0 })}
            </DropdownMenuItem>
          )}
          {hasError && (
            <DropdownMenuItem
              onClick={() => {
                setCurrentRow(channel);
                setOpen('errorResolved');
              }}
              className='text-green-600!'
            >
              <IconCheck size={16} className='mr-2' />
              {t('channels.actions.markErrorResolved')}
            </DropdownMenuItem>
          )}
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('archive');
            }}
            className={isArchived ? 'text-green-600!' : 'text-orange-500!'}
          >
            {isArchived ? <IconCheck size={16} className='mr-2' /> : <IconArchive size={16} className='mr-2' />}
            {t(isArchived ? 'common.buttons.restore' : 'common.buttons.archive')}
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => {
              setCurrentRow(channel);
              setOpen('delete');
            }}
            className='text-red-500!'
          >
            <IconTrash size={16} className='mr-2' />
            {t('common.buttons.delete')}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
});

ActionCell.displayName = 'ActionCell';

const ExpandCell = ({ row }: { row: any }) => (
  <div className='flex justify-center'>
    <Button
      variant='ghost'
      size='sm'
      className='h-6 w-6 p-0'
      onClick={(e) => {
        e.stopPropagation();
        row.toggleExpanded();
      }}
    >
      {row.getIsExpanded() ? <IconChevronDown className='h-4 w-4' /> : <IconChevronRight className='h-4 w-4' />}
    </Button>
  </div>
);

// ExpandCell.displayName = 'ExpandCell'; // Removed since it's not memoized now, but can keep if desired

function getChannelWebsiteURL(baseURL: string): string | null {
  try {
    const url = new URL(baseURL);
    return url.origin;
  } catch {
    return null;
  }
}

function getProxyURLSummary(proxyURL: string): { label: string; detail?: string } {
  try {
    const url = new URL(proxyURL);
    const pathname = url.pathname === '/' ? '' : url.pathname;
    return {
      label: url.host || proxyURL,
      detail: `${url.protocol}//${url.host}${pathname}`,
    };
  } catch {
    return { label: proxyURL };
  }
}

// Memoized cell components to avoid recreating on every render
const NameCell = memo(({ row }: { row: Row<Channel> }) => {
  const { t } = useTranslation();
  const channel = row.original;
  const hasError = !!channel.errorMessage;
  const disabledKeysCount = channel.disabledAPIKeys?.length ?? 0;
  const hasDisabledKeys = disabledKeysCount > 0;
  const websiteURL = getChannelWebsiteURL(channel.baseURL);

  const nameElement = websiteURL ? (
    <a
      href={websiteURL}
      target='_blank'
      rel='noopener noreferrer'
      className={cn('truncate font-medium hover:underline', hasError ? 'text-destructive' : '')}
      onClick={(e) => e.stopPropagation()}
    >
      {row.getValue('name')}
    </a>
  ) : (
    <div className={cn('truncate font-medium', hasError && 'text-destructive')}>{row.getValue('name')}</div>
  );

  // Both indicators are shown independently: a channel disabled because every
  // credential is unavailable carries an error *and* disabled credentials, and
  // hiding the key icon behind the error would lose the reason it went down.
  const content = (
    <div className='flex justify-center'>
      <div className='flex max-w-56 items-center gap-2'>
        {hasError && <IconAlertTriangle className='text-destructive h-4 w-4 shrink-0' />}
        {hasDisabledKeys && <IconKeyOff className='h-4 w-4 shrink-0 text-amber-500' />}
        {nameElement}
      </div>
    </div>
  );

  if (!hasError && !hasDisabledKeys) {
    return content;
  }

  return (
    <Tooltip>
      <TooltipTrigger asChild>{content}</TooltipTrigger>
      <TooltipContent>
        <div className='space-y-1'>
          {hasError && (
            <p className='text-destructive text-sm'>
              {t(`channels.messages.${channel.errorMessage}`, {
                defaultValue: channel.errorMessage,
              })}
            </p>
          )}
          {hasDisabledKeys && (
            <p className='text-sm text-amber-500'>{t('channels.actions.disabledAPIKeys', { count: disabledKeysCount })}</p>
          )}
        </div>
      </TooltipContent>
    </Tooltip>
  );
});

NameCell.displayName = 'NameCell';

const ProviderCell = memo(({ row }: { row: Row<Channel> }) => {
  const { t } = useTranslation();
  const type = row.original.type;
  const config = CHANNEL_CONFIGS[type];
  const provider = getProvider(type);
  const IconComponent = config.icon;
  return (
    <div className='flex justify-center'>
      <Badge variant='outline' className={cn('capitalize', config.color)}>
        <div className='flex items-center gap-2'>
          <IconComponent size={16} className='shrink-0' />
          <span>{t(`channels.providers.${provider}`)}</span>
        </div>
      </Badge>
    </div>
  );
});

ProviderCell.displayName = 'ProviderCell';

const TagsCell = memo(({ row }: { row: Row<Channel> }) => {
  const tags = (row.getValue('tags') as string[]) || [];
  const containerRef = useRef<HTMLDivElement>(null);
  const tagMeasureRefs = useRef<Array<HTMLSpanElement | null>>([]);
  const overflowMeasureRefs = useRef<Array<HTMLSpanElement | null>>([]);
  const [visibleCount, setVisibleCount] = useState(Math.min(tags.length, 2));

  const updateVisibleCount = useCallback(() => {
    const availableWidth = containerRef.current?.clientWidth ?? 0;
    if (availableWidth <= 0) return;

    const gap = 4;
    let occupiedWidth = 0;
    let nextVisibleCount = 0;

    for (let index = 0; index < tags.length; index += 1) {
      const tagWidth = tagMeasureRefs.current[index]?.offsetWidth ?? 0;
      const nextOccupiedWidth = occupiedWidth + (index > 0 ? gap : 0) + tagWidth;
      const remainingCount = tags.length - index - 1;
      const overflowWidth = remainingCount > 0 ? (overflowMeasureRefs.current[remainingCount]?.offsetWidth ?? 0) + gap : 0;

      if (nextOccupiedWidth + overflowWidth > availableWidth) break;

      occupiedWidth = nextOccupiedWidth;
      nextVisibleCount = index + 1;
    }

    setVisibleCount((current) => (current === nextVisibleCount ? current : nextVisibleCount));
  }, [tags]);

  useLayoutEffect(() => {
    updateVisibleCount();

    const container = containerRef.current;
    if (!container) return;

    const observer = new ResizeObserver(updateVisibleCount);
    observer.observe(container);
    return () => observer.disconnect();
  }, [updateVisibleCount]);

  if (tags.length === 0) {
    return (
      <div className='flex justify-center'>
        <span className='text-muted-foreground text-xs'>-</span>
      </div>
    );
  }

  const hiddenCount = tags.length - visibleCount;

  return (
    <div ref={containerRef} className='relative flex w-full min-w-0 justify-center gap-1 overflow-hidden'>
      {tags.slice(0, visibleCount).map((tag) => (
        <Badge key={tag} variant='outline' className='max-w-full truncate text-xs'>
          {tag}
        </Badge>
      ))}
      {hiddenCount > 0 && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge variant='outline' className='cursor-default text-xs'>
              +{hiddenCount}
            </Badge>
          </TooltipTrigger>
          <TooltipContent className='max-w-72'>
            <div className='flex flex-wrap gap-1'>
              {tags.slice(visibleCount).map((tag) => (
                <Badge key={tag} variant='outline' className='text-xs'>
                  {tag}
                </Badge>
              ))}
            </div>
          </TooltipContent>
        </Tooltip>
      )}
      <div aria-hidden='true' className='pointer-events-none invisible absolute top-0 left-0 flex w-max gap-1'>
        {tags.map((tag, index) => (
          <Badge
            key={`measure-${tag}`}
            ref={(element) => {
              tagMeasureRefs.current[index] = element;
            }}
            variant='outline'
            className='text-xs'
          >
            {tag}
          </Badge>
        ))}
        {tags.map((_, index) => {
          const count = index + 1;
          return (
            <Badge
              key={`measure-overflow-${count}`}
              ref={(element) => {
                overflowMeasureRefs.current[count] = element;
              }}
              variant='outline'
              className='text-xs'
            >
              +{count}
            </Badge>
          );
        })}
      </div>
    </div>
  );
});

TagsCell.displayName = 'TagsCell';

const ProxyCell = memo(({ row }: { row: Row<Channel> }) => {
  const { t } = useTranslation();
  const proxy = row.original.settings?.proxy;

  if (!proxy || proxy.type === 'disabled') {
    return (
      <div className='flex justify-center'>
        <span className='text-muted-foreground text-xs'>-</span>
      </div>
    );
  }

  if (proxy.type === 'environment') {
    return (
      <div className='flex justify-center'>
        <span className='text-muted-foreground text-xs'>{t('channels.dialogs.proxy.types.environment')}</span>
      </div>
    );
  }

  const proxyURL = proxy.url?.trim();
  if (!proxyURL) {
    return (
      <div className='flex justify-center'>
        <span className='text-muted-foreground text-xs'>-</span>
      </div>
    );
  }

  const { label, detail } = getProxyURLSummary(proxyURL);
  const content = (
    <div className='flex justify-center'>
      <span className='max-w-40 truncate font-mono text-xs'>{label}</span>
    </div>
  );

  if (detail && detail !== label) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{content}</TooltipTrigger>
        <TooltipContent>{detail}</TooltipContent>
      </Tooltip>
    );
  }

  return content;
});

ProxyCell.displayName = 'ProxyCell';

const SupportedModelsCell = memo(({ row }: { row: Row<Channel> }) => {
  const { t } = useTranslation();
  const channel = row.original;
  const models = row.getValue('supportedModels') as string[];
  const { setOpen, setCurrentRow } = useChannels();

  const handleOpenModelsDialog = useCallback(() => {
    setCurrentRow(channel);
    setOpen('viewModels');
  }, [channel, setCurrentRow, setOpen]);

  return (
    <div className='flex items-center justify-center gap-2'>
      <div className='flex flex-wrap justify-center gap-1 overflow-hidden'>
        {models.slice(0, 5).map((model) => (
          <Badge key={model} variant='secondary' className='block max-w-48 truncate text-left text-xs'>
            {model}
          </Badge>
        ))}
        {models.length > 5 && (
          <Badge
            variant='secondary'
            className='hover:bg-primary hover:text-primary-foreground cursor-pointer text-xs transition-colors'
            onClick={handleOpenModelsDialog}
            title={t('channels.actions.viewModels')}
          >
            +{models.length - 5}
          </Badge>
        )}
      </div>
    </div>
  );
});

SupportedModelsCell.displayName = 'SupportedModelsCell';

const OrderingWeightCell = memo(({ row }: { row: Row<Channel> }) => {
  const channel = row.original;
  const initialWeight = row.getValue('orderingWeight') as number | null;
  const [isEditing, setIsEditing] = useState(false);
  const [weight, setWeight] = useState<string>(initialWeight?.toString() || '1');
  const updateChannel = useUpdateChannel();
  const { channelPermissions } = usePermissions();
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleDoubleClick = useCallback(() => {
    setIsEditing(true);
    setWeight(initialWeight?.toString() || '1');
  }, [initialWeight]);

  const handleSave = useCallback(async () => {
    const weightValue = clampWeight(Number(weight));
    if (weightValue === initialWeight) {
      setIsEditing(false);
      return;
    }

    try {
      await updateChannel.mutateAsync({
        id: channel.id,
        input: { orderingWeight: weightValue },
      });
      setIsEditing(false);
    } catch (_error) {
      // Error handled by mutation hook
    }
  }, [channel.id, weight, initialWeight, updateChannel]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        handleSave();
      } else if (e.key === 'Escape') {
        setIsEditing(false);
        setWeight(initialWeight?.toString() || '1');
      }
    },
    [handleSave, initialWeight]
  );

  if (!channelPermissions.canWrite) {
    return <span className={cn('font-mono text-sm', initialWeight == null && 'text-muted-foreground')}>{initialWeight ?? '-'}</span>;
  }

  if (isEditing) {
    return (
      <div className='flex justify-center px-2'>
        <Input
          ref={inputRef}
          type='number'
          inputMode='decimal'
          step='any'
          min={MIN_WEIGHT}
          max={MAX_WEIGHT}
          value={weight}
          onChange={(e) => setWeight(e.target.value)}
          onBlur={handleSave}
          onKeyDown={handleKeyDown}
          className='h-7 w-20 text-center font-mono text-sm'
          disabled={updateChannel.isPending}
        />
      </div>
    );
  }

  return (
    <div className='group flex cursor-pointer items-center justify-center gap-2' onDoubleClick={handleDoubleClick}>
      <span className={cn('font-mono text-sm', initialWeight == null && 'text-muted-foreground')}>{initialWeight ?? '-'}</span>
      {updateChannel.isPending && <IconLoader2 className='text-muted-foreground h-3 w-3 animate-spin' />}
    </div>
  );
});

OrderingWeightCell.displayName = 'OrderingWeightCell';

const CreatedAtCell = memo(({ row }: { row: Row<Channel> }) => {
  const raw = row.getValue('createdAt') as unknown;
  const date = raw instanceof Date ? raw : new Date(raw as string);

  if (Number.isNaN(date.getTime())) {
    return (
      <div className='flex justify-center'>
        <span className='text-muted-foreground text-xs'>-</span>
      </div>
    );
  }

  return (
    <div className='flex justify-center'>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className='text-muted-foreground cursor-help text-sm'>{format(date, 'yyyy-MM-dd')}</div>
        </TooltipTrigger>
        <TooltipContent>{format(date, 'yyyy-MM-dd HH:mm:ss')}</TooltipContent>
      </Tooltip>
    </div>
  );
});

CreatedAtCell.displayName = 'CreatedAtCell';

function normalizeUsageQueryText(value: string): string {
  return value.replace(/\r\n?/g, '\n').replace(/\n+/g, '\n').split('\n').slice(0, 2).join('\n');
}

const UsageQueryCell = memo(({ row, canWrite }: { row: Row<Channel>; canWrite: boolean }) => {
  const { t } = useTranslation();
  const refreshUsageQuery = useRefreshChannelUsageQuery();
  const channel = row.original;
  const configured = channel.settings?.usageQuery?.enabled === true;
  const quotaStatus = channel.providerQuotaStatus?.providerType === 'usage_query' ? channel.providerQuotaStatus : null;
  const quotaData = quotaStatus?.quotaData as ProviderUsageQueryQuotaData | null | undefined;
  const balance = quotaData?.balance ?? (quotaData?.remaining != null ? { remaining: quotaData.remaining, unit: quotaData.unit } : undefined);
  const hasListResult = (quotaData?.text?.trim().length ?? 0) > 0 || balance !== undefined;

  let fullText = '-';
  if (quotaData?.error) {
    fullText = quotaData.error;
  } else if (quotaData?.isValid === false) {
    fullText = quotaData.invalidMessage || t('channels.usageQuery.invalid');
  } else if (quotaData) {
    const details: string[] = [];
    if (quotaData.text) {
      details.push(quotaData.text);
    }
    if (balance) {
      details.push(`${t('channels.usageQuery.remaining')}${formatUsageQueryBalance(balance.remaining, balance.unit)}`);
    }
    if (details.length > 0) fullText = details.join('\n');
  }
  const text = normalizeUsageQueryText(fullText);

  const refresh = async (event: React.MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation();
    try {
      await refreshUsageQuery.mutateAsync({ channelID: channel.id });
      toast.success(t('channels.usageQuery.refresh.success'));
    } catch (error) {
      toast.error(t('channels.usageQuery.refresh.failed'), {
        description: error instanceof Error ? error.message : String(error),
      });
    }
  };

  return (
    <div className='flex min-w-0 items-center justify-center gap-1'>
      <Tooltip>
        <TooltipTrigger asChild>
          <span className='line-clamp-2 min-w-0 flex-1 whitespace-pre-line break-words text-left text-sm leading-5'>{text}</span>
        </TooltipTrigger>
        <TooltipContent className='max-w-96 whitespace-pre-line break-words'>{fullText}</TooltipContent>
      </Tooltip>
      {canWrite && configured && hasListResult && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7 shrink-0'
              disabled={refreshUsageQuery.isPending}
              onClick={refresh}
              aria-label={t('channels.usageQuery.refresh.channel')}
            >
              <IconRefresh className={cn('h-4 w-4', refreshUsageQuery.isPending && 'animate-spin')} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('channels.usageQuery.refresh.channel')}</TooltipContent>
        </Tooltip>
      )}
    </div>
  );
});

UsageQueryCell.displayName = 'UsageQueryCell';

const UsageQueryHeader = ({ column, canWrite }: { column: any; canWrite: boolean }) => {
  const { t } = useTranslation();
  const refreshAllUsageQueries = useRefreshAllUsageQueries();

  const refresh = async (event: React.MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation();
    try {
      await refreshAllUsageQueries.mutateAsync();
      toast.success(t('channels.usageQuery.refresh.allSuccess'));
    } catch (error) {
      toast.error(t('channels.usageQuery.refresh.failed'), {
        description: error instanceof Error ? error.message : String(error),
      });
    }
  };

  return (
    <div className='flex items-center gap-1'>
      <DataTableColumnHeader column={column} title={t('channels.columns.usageQuery')} className='min-w-0 flex-1 justify-center' />
      {canWrite && (
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='ghost'
              size='icon'
              className='h-7 w-7 shrink-0'
              disabled={refreshAllUsageQueries.isPending}
              onClick={refresh}
              aria-label={t('channels.usageQuery.refresh.all')}
            >
              <IconRefresh className={cn('h-4 w-4', refreshAllUsageQueries.isPending && 'animate-spin')} />
            </Button>
          </TooltipTrigger>
          <TooltipContent>{t('channels.usageQuery.refresh.all')}</TooltipContent>
        </Tooltip>
      )}
    </div>
  );
};

interface ChannelSortingOptions {
  enabledFirst: boolean;
  onEnabledFirstChange: (enabled: boolean) => void;
}

export const createColumns = (
  t: ReturnType<typeof useTranslation>['t'],
  canWrite: boolean = true,
  sortingOptions?: ChannelSortingOptions
): ColumnDef<Channel>[] => {
  return [
    {
      id: 'expand',
      header: () => null,
      meta: {
        className: 'text-center',
      },
      size: 40,
      minSize: 36,
      maxSize: 56,
      cell: ExpandCell,
      enableSorting: false,
      enableHiding: false,
    },
    ...(canWrite
      ? [
          {
            id: 'select',
            header: ({ table }: { table: Table<Channel> }) => (
              <div className='flex justify-center'>
                <Checkbox
                  checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
                  onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
                  aria-label={t('common.columns.selectAll')}
                  className='translate-y-[2px]'
                />
              </div>
            ),
            cell: ({ row }: { row: Row<Channel> }) => (
              <div className='flex justify-center'>
                <Checkbox
                  checked={row.getIsSelected()}
                  onCheckedChange={(value) => row.toggleSelected(!!value)}
                  aria-label={t('common.columns.selectRow')}
                  className='translate-y-[2px]'
                />
              </div>
            ),
            meta: {
              className: 'text-center',
            },
            size: 44,
            minSize: 40,
            maxSize: 64,
            enableSorting: false,
            enableHiding: true,
          },
        ]
      : []),
    {
      accessorKey: 'id',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('common.columns.id')} className='justify-center' />,
      cell: ({ row }) => <div className='text-center font-mono text-xs'>#{extractNumberID(row.getValue('id'))}</div>,
      meta: {
        className: 'text-center',
      },
      size: 84,
      minSize: 72,
      maxSize: 128,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: 'name',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('common.columns.name')} className='justify-center' />,
      cell: NameCell,
      meta: {
        className: 'md:table-cell text-center',
      },
      size: 210,
      minSize: 144,
      maxSize: 480,
      enableHiding: false,
      enableSorting: true,
    },
    {
      id: 'provider',
      accessorKey: 'type',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('channels.columns.provider')} className='justify-center' />,
      cell: ProviderCell,
      meta: {
        className: 'text-center',
      },
      size: 148,
      minSize: 112,
      maxSize: 240,
      filterFn: (row, _id, value) => {
        return value.includes(row.original.type);
      },
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('common.columns.status')} className='justify-center'>
          {sortingOptions && (
            <DropdownMenuItem
              className='gap-2'
              onSelect={(event) => event.preventDefault()}
              onClick={() => sortingOptions.onEnabledFirstChange(!sortingOptions.enabledFirst)}
            >
              <span className='flex-1'>{t('channels.sort.enabledFirst')}</span>
              <Switch checked={sortingOptions.enabledFirst} className='pointer-events-none scale-75' aria-hidden='true' />
            </DropdownMenuItem>
          )}
        </DataTableColumnHeader>
      ),
      cell: StatusSwitchCell,
      meta: {
        className: 'text-center',
      },
      size: 96,
      minSize: 84,
      maxSize: 144,
      enableSorting: true,
      enableHiding: false,
    },

    {
      accessorKey: 'tags',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('channels.columns.tags')} className='justify-center' />,
      cell: TagsCell,
      meta: {
        className: 'text-center',
      },
      size: 192,
      minSize: 104,
      maxSize: 560,
      filterFn: (row, id, value) => {
        const tags = (row.getValue(id) as string[]) || [];
        // Single select: value is a string, not an array
        return tags.includes(value as string);
      },
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: 'model',
      accessorFn: () => '', // Virtual column for filtering only
      header: () => null,
      cell: () => null,
      filterFn: () => true, // Server-side filtering, always return true
      enableSorting: false,
      enableHiding: true,
      enableColumnFilter: false,
      enableGlobalFilter: false,
      enableResizing: false,
    },
    {
      accessorKey: 'supportedModels',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('channels.columns.supportedModels')} className='justify-center' />
      ),
      cell: SupportedModelsCell,
      meta: {
        className: 'text-center',
      },
      size: 210,
      minSize: 144,
      maxSize: 480,
      enableSorting: false,
    },
    {
      id: 'endpoints',
      accessorFn: (row) => [...(row.defaultEndpoints ?? []), ...(row.endpoints ?? [])],
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('channels.columns.supportedEndpoints')} className='justify-center' />
      ),
      cell: ({ row }) => <ChannelEndpointsCell channel={row.original} />,
      meta: {
        className: 'text-center',
      },
      size: 144,
      minSize: 120,
      maxSize: 192,
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: 'proxy',
      accessorFn: (row) => row.settings?.proxy?.url ?? row.settings?.proxy?.type ?? '',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('channels.columns.proxy')} className='justify-center' />,
      cell: ProxyCell,
      meta: {
        className: 'text-center',
      },
      size: 128,
      minSize: 104,
      maxSize: 280,
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: 'health',
      accessorKey: 'health',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('channels.columns.health')} className='justify-center' />,
      cell: ({ row }: { row: Row<Channel> }) => {
        const probePoints = (row.original as any).probePoints || [];
        const limiterStats = row.original.liveLimiterStats;
        return (
          <div className='flex flex-col items-center gap-1'>
            <ChannelHealthCell points={probePoints} />
            {limiterStats ? <ChannelLimiterCell stats={limiterStats} /> : null}
          </div>
        );
      },
      meta: {
        className: 'text-center',
      },
      size: 144,
      minSize: 112,
      maxSize: 240,
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: 'usageQuery',
      accessorFn: (row) => row.providerQuotaStatus?.quotaData ?? null,
      header: ({ column }) => <UsageQueryHeader column={column} canWrite={canWrite} />,
      cell: ({ row }: { row: Row<Channel> }) => <UsageQueryCell row={row} canWrite={canWrite} />,
      meta: {
        className: 'text-center',
      },
      size: 240,
      minSize: 160,
      maxSize: 360,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: 'orderingWeight',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('channels.columns.orderingWeight')} className='justify-center' />
      ),
      cell: OrderingWeightCell,
      meta: {
        className: 'text-center',
      },
      size: 92,
      minSize: 76,
      maxSize: 132,
      sortingFn: 'alphanumeric',
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: 'createdAt',
      header: ({ column }) => <DataTableColumnHeader column={column} title={t('common.columns.createdAt')} className='justify-center' />,
      cell: CreatedAtCell,
      meta: {
        className: 'text-center',
      },
      size: 168,
      minSize: 132,
      maxSize: 240,
      enableSorting: true,
      enableHiding: true,
    },
    ...(canWrite
      ? [
          {
            id: 'action',
            header: ({ column }: { column: any }) => (
              <DataTableColumnHeader column={column} title={t('common.columns.actions')} className='justify-center' />
            ),
            cell: ActionCell,
            meta: {
              className: 'text-center',
            },
            size: 144,
            minSize: 128,
            maxSize: 200,
            enableSorting: false,
            enableHiding: true,
          },
        ]
      : []),
  ];
};
