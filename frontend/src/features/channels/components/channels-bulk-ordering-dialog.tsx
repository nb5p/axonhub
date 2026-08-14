import { useState, useEffect, memo, useCallback, Fragment, useMemo } from 'react';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import { arrayMove, SortableContext, sortableKeyboardCoordinates, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical, ArrowUpToLine, ArrowDownToLine, ArrowDownWideNarrow, Check } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { useAllChannelSummarys, useBulkDisableChannels, useBulkEnableChannels, useBulkUpdateChannelOrdering } from '../data/channels';
import { ChannelSummary, type ChannelStatus, type ChannelSummaryConnection } from '../data/schema';

const WEIGHT_PRECISION = 0;
const MIN_WEIGHT = 0;
const MAX_WEIGHT = 100;

const formatWeight = (value: number) => Math.round(value);

const clampWeight = (value: number) => formatWeight(Math.min(MAX_WEIGHT, Math.max(MIN_WEIGHT, value)));

const calculateRelativeWeight = (prev?: number, next?: number) => {
  if (prev == null && next == null) {
    return clampWeight(1);
  }
  if (prev == null) {
    return clampWeight((next ?? 0) + 1);
  }
  if (next == null) {
    return clampWeight(prev - 1);
  }
  if (prev === next) {
    return clampWeight(prev);
  }
  return clampWeight(Math.floor((prev + next) / 2));
};

type OrderedChannel = {
  channel: ChannelSummary;
  orderingWeight: number;
};

const STATUS_DIVIDER_ID = 'channel-status-divider';

const isEnabledStatus = (status: ChannelStatus) => status === 'enabled';

const sortOrderedChannels = (items: OrderedChannel[], enabledFirst: boolean) =>
  items
    .map((item, index) => ({ item, index }))
    .sort((a, b) => {
      if (enabledFirst) {
        const statusDifference = Number(isEnabledStatus(b.item.channel.status)) - Number(isEnabledStatus(a.item.channel.status));
        if (statusDifference !== 0) return statusDifference;
      }

      return b.item.orderingWeight - a.item.orderingWeight || a.index - b.index;
    })
    .map(({ item }) => item);

const createOrderedChannels = (channelsData?: ChannelSummaryConnection): OrderedChannel[] => {
  if (!channelsData?.edges) return [];

  return sortOrderedChannels(
    channelsData.edges.map((edge, index) => ({
      channel: edge.node,
      orderingWeight: clampWeight(edge.node.orderingWeight ?? channelsData.edges.length - index),
    })),
    false
  );
};

function ChannelStatusDivider({ enabledCount, disabledCount }: { enabledCount: number; disabledCount: number }) {
  const { t } = useTranslation();
  const { isOver, setNodeRef } = useDroppable({ id: STATUS_DIVIDER_ID });

  return (
    <div
      ref={setNodeRef}
      className={`flex items-center gap-3 rounded-md py-2 transition-colors ${isOver ? 'bg-primary/10 px-2' : ''}`}
      data-testid='channel-status-divider'
    >
      <Badge className='border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'>
        {t('channels.dialogs.bulkOrdering.enabledSection', { count: enabledCount })}
      </Badge>
      <Separator className='flex-1' />
      <Badge variant='secondary'>{t('channels.dialogs.bulkOrdering.disabledSection', { count: disabledCount })}</Badge>
    </div>
  );
}

interface ChannelOrderingItemProps {
  channel: ChannelSummary;
  orderingWeight: number;
  index: number;
  onMoveToTop: (index: number) => void;
  onMoveToBottom: (index: number) => void;
  onWeightChange: (id: string, weight: number) => void;
  onStatusChange: (id: string, enabled: boolean) => void;
  canMoveToTop: boolean;
  canMoveToBottom: boolean;
}

const ChannelOrderingItemComponent = memo(function ChannelOrderingItemComponent({
  channel,
  orderingWeight,
  index,
  onMoveToTop,
  onMoveToBottom,
  onWeightChange,
  onStatusChange,
  canMoveToTop,
  canMoveToBottom,
}: ChannelOrderingItemProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: channel.id });
  const { t } = useTranslation();
  const [localWeight, setLocalWeight] = useState(orderingWeight.toString());

  useEffect(() => {
    setLocalWeight(orderingWeight.toString());
  }, [orderingWeight]);

  const handleWeightBlur = () => {
    if (localWeight.trim() === '') {
      setLocalWeight(orderingWeight.toString());
      return;
    }

    const val = Number(localWeight);
    if (!Number.isNaN(val) && val !== orderingWeight) {
      onWeightChange(channel.id, val);
    } else {
      setLocalWeight(orderingWeight.toString());
    }
  };

  const getTypeDisplayName = (type: string) => {
    const typeKey = `channels.types.${type}` as const;
    return t(typeKey, type);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'enabled':
        return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950 dark:text-emerald-400 dark:border-emerald-800';
      case 'disabled':
        return 'bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-900 dark:text-gray-400 dark:border-gray-700';
      case 'archived':
        return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950 dark:text-amber-400 dark:border-amber-800';
      default:
        return 'bg-gray-50 text-gray-600 border-gray-200 dark:bg-gray-900 dark:text-gray-400 dark:border-gray-700';
    }
  };

  const getTypeColor = (type: string) => {
    const colors = {
      openai: 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950 dark:text-blue-400',
      anthropic: 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-950 dark:text-purple-400',
      deepseek: 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-950 dark:text-indigo-400',
      doubao: 'bg-orange-50 text-orange-700 border-orange-200 dark:bg-orange-950 dark:text-orange-400',
      kimi: 'bg-pink-50 text-pink-700 border-pink-200 dark:bg-pink-950 dark:text-pink-400',
    };
    return colors[type as keyof typeof colors] || 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-900 dark:text-gray-400';
  };

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`group bg-card flex items-center gap-2 rounded-md border p-1 hover:shadow-sm ${
        isDragging ? 'ring-primary/20 relative z-50 shadow-xl ring-2' : 'hover:border-primary/20'
      }`}
    >
      {/* Drag Handle */}
      <div
        className='text-muted-foreground hover:text-foreground flex min-w-[40px] cursor-grab items-center gap-1 px-1 active:cursor-grabbing'
        {...attributes}
        {...listeners}
      >
        <GripVertical className='h-3.5 w-3.5' />
        <span className='w-[20px] text-center font-mono text-[10px]'>{index + 1}</span>
      </div>

      {/* Channel Info - Single Line Optimized */}
      <div className='flex min-w-0 flex-1 items-center gap-2'>
        <div className='flex min-w-0 items-center gap-1.5'>
          <span className='truncate text-sm font-medium'>{channel.name}</span>
          <div className='flex flex-shrink-0 gap-1'>
            <Badge variant='outline' className={`h-3.5 px-1 text-[10px] font-normal ${getTypeColor(channel.type)}`}>
              {getTypeDisplayName(channel.type)}
            </Badge>
            <Badge variant='outline' className={`h-3.5 px-1 text-[10px] font-normal ${getStatusColor(channel.status)}`}>
              {t(`channels.status.${channel.status}`)}
            </Badge>
          </div>
        </div>

        <div className='hidden flex-1 items-center gap-2 sm:flex'>
          <div className='bg-border h-3 w-[1px]' />
          <span className='text-muted-foreground truncate font-mono text-[10px] opacity-70'>{channel.baseURL}</span>
        </div>
      </div>

      {/* Controls */}
      <div className='flex items-center gap-1 pr-1'>
        <div className='flex items-center gap-1.5' onPointerDown={(event) => event.stopPropagation()}>
          <Switch
            checked={channel.status === 'enabled'}
            onCheckedChange={(checked) => onStatusChange(channel.id, checked)}
            aria-label={t(
              channel.status === 'enabled' ? 'channels.dialogs.bulkOrdering.disableChannel' : 'channels.dialogs.bulkOrdering.enableChannel'
            )}
            className='scale-90'
          />
          <span className='text-muted-foreground hidden w-8 text-[10px] lg:inline'>{t(`channels.status.${channel.status}`)}</span>
        </div>
        <div className='bg-muted/30 hidden items-center gap-1.5 rounded px-1.5 py-0.5 sm:flex'>
          <span className='text-muted-foreground text-[10px]'>{t('channels.dialogs.bulkOrdering.orderingWeight')}</span>
          <Input
            type='number'
            inputMode='decimal'
            step='any'
            min={MIN_WEIGHT}
            max={MAX_WEIGHT}
            className='h-6 w-16 px-1 text-center text-xs'
            value={localWeight}
            onChange={(e) => setLocalWeight(e.target.value)}
            onBlur={handleWeightBlur}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.currentTarget.blur();
              }
            }}
            onClick={(e) => e.stopPropagation()}
            onPointerDown={(e) => e.stopPropagation()}
          />
        </div>

        <div className='flex items-center gap-0.5'>
          <Button
            variant='ghost'
            size='icon'
            className='text-muted-foreground hover:text-foreground h-6 w-6'
            onClick={() => onMoveToTop(index)}
            disabled={!canMoveToTop}
            title={t('common.moveToTop', 'Move to top')}
          >
            <ArrowUpToLine className='h-3.5 w-3.5' />
          </Button>
          <Button
            variant='ghost'
            size='icon'
            className='text-muted-foreground hover:text-foreground h-6 w-6'
            onClick={() => onMoveToBottom(index)}
            disabled={!canMoveToBottom}
            title={t('common.moveToBottom', 'Move to bottom')}
          >
            <ArrowDownToLine className='h-3.5 w-3.5' />
          </Button>
        </div>
      </div>
    </div>
  );
});

interface ChannelsBulkOrderingDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ChannelsBulkOrderingDialog({ open, onOpenChange }: ChannelsBulkOrderingDialogProps) {
  const { t } = useTranslation();

  // Only fetch channels when dialog is open (lazy loading)
  const { data: channelsData, isLoading } = useAllChannelSummarys(undefined, { enabled: open });

  const bulkUpdateMutation = useBulkUpdateChannelOrdering();
  const bulkEnableMutation = useBulkEnableChannels();
  const bulkDisableMutation = useBulkDisableChannels();

  // Local state for ordering
  const [orderedChannels, setOrderedChannels] = useState<OrderedChannel[]>([]);
  const [enabledFirst, setEnabledFirst] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  // Initialize ordered channels when data loads
  useEffect(() => {
    if (channelsData) {
      setOrderedChannels(createOrderedChannels(channelsData));
      setEnabledFirst(false);
      setHasChanges(false);
    }
  }, [channelsData]);

  const enabledCount = useMemo(() => orderedChannels.filter((item) => isEnabledStatus(item.channel.status)).length, [orderedChannels]);

  const disabledCount = orderedChannels.length - enabledCount;

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { active, over } = event;

      if (!over || active.id === over.id) {
        return;
      }

      setOrderedChannels((items) => {
        const oldIndex = items.findIndex((item) => item.channel.id === active.id);
        if (oldIndex === -1) {
          return items;
        }

        const activeItem = items[oldIndex];
        const droppedOnDivider = over.id === STATUS_DIVIDER_ID;
        const newIndex = droppedOnDivider
          ? isEnabledStatus(activeItem.channel.status)
            ? Math.max(0, enabledCount - 1)
            : enabledCount
          : items.findIndex((item) => item.channel.id === over.id);

        if (newIndex === -1) {
          return items;
        }

        const destinationStatus =
          enabledFirst && droppedOnDivider
            ? isEnabledStatus(activeItem.channel.status)
              ? 'disabled'
              : 'enabled'
            : enabledFirst
              ? items[newIndex].channel.status
              : activeItem.channel.status;
        const newItems = arrayMove(items, oldIndex, newIndex);

        if (enabledFirst && newItems[newIndex].channel.status !== destinationStatus) {
          newItems[newIndex] = {
            ...newItems[newIndex],
            channel: {
              ...newItems[newIndex].channel,
              status: destinationStatus,
            },
          };
        }

        const previousInGroup = [...newItems]
          .slice(0, newIndex)
          .reverse()
          .find((item) => !enabledFirst || item.channel.status === destinationStatus);
        const nextInGroup = newItems.slice(newIndex + 1).find((item) => !enabledFirst || item.channel.status === destinationStatus);

        newItems[newIndex] = {
          ...newItems[newIndex],
          orderingWeight: calculateRelativeWeight(previousInGroup?.orderingWeight, nextInGroup?.orderingWeight),
        };

        setHasChanges(true);
        return newItems;
      });
    },
    [enabledCount, enabledFirst]
  );

  const handleWeightChange = useCallback(
    (id: string, weight: number) => {
      const normalizedWeight = clampWeight(weight);
      setOrderedChannels((items) => {
        const newItems = items.map((item) => (item.channel.id === id ? { ...item, orderingWeight: normalizedWeight } : item));
        const sortedItems = sortOrderedChannels(newItems, enabledFirst);
        setHasChanges(true);
        return sortedItems;
      });
    },
    [enabledFirst]
  );

  const handleStatusChange = useCallback(
    (id: string, enabled: boolean) => {
      const nextStatus: ChannelStatus = enabled ? 'enabled' : 'disabled';
      setOrderedChannels((items) => {
        const newItems = items.map((item) =>
          item.channel.id === id
            ? {
                ...item,
                channel: {
                  ...item.channel,
                  status: nextStatus,
                },
              }
            : item
        );

        setHasChanges(true);
        return enabledFirst ? sortOrderedChannels(newItems, true) : newItems;
      });
    },
    [enabledFirst]
  );

  const handleEnabledFirstChange = useCallback((pressed: boolean) => {
    setEnabledFirst(pressed);
    setOrderedChannels((items) => sortOrderedChannels(items, pressed));
  }, []);

  const handleMoveToTop = useCallback(
    (index: number) => {
      setOrderedChannels((items) => {
        if (!items.length || index === 0) {
          return items;
        }

        const status = items[index].channel.status;
        const targetIndex = enabledFirst && !isEnabledStatus(status) ? enabledCount : 0;
        if (index === targetIndex) return items;

        const newItems = arrayMove(items, index, targetIndex);
        const nextWeight = newItems.slice(targetIndex + 1).find((item) => !enabledFirst || item.channel.status === status)?.orderingWeight;

        newItems[targetIndex] = {
          ...newItems[targetIndex],
          orderingWeight: calculateRelativeWeight(undefined, nextWeight),
        };

        setHasChanges(true);
        return newItems;
      });
    },
    [enabledCount, enabledFirst]
  );

  const handleMoveToBottom = useCallback(
    (index: number) => {
      setOrderedChannels((items) => {
        if (!items.length) {
          return items;
        }

        const status = items[index].channel.status;
        const targetIndex = enabledFirst && isEnabledStatus(status) ? enabledCount - 1 : items.length - 1;
        if (index === targetIndex) return items;

        const newItems = arrayMove(items, index, targetIndex);
        const prevWeight = [...newItems]
          .slice(0, targetIndex)
          .reverse()
          .find((item) => !enabledFirst || item.channel.status === status)?.orderingWeight;

        newItems[targetIndex] = {
          ...newItems[targetIndex],
          orderingWeight: calculateRelativeWeight(prevWeight, undefined),
        };

        setHasChanges(true);
        return newItems;
      });
    },
    [enabledCount, enabledFirst]
  );

  const handleSave = async () => {
    try {
      const originalStatuses = new Map(channelsData?.edges.map((edge) => [edge.node.id, edge.node.status]));
      const enableIds = orderedChannels
        .filter((item) => item.channel.status === 'enabled' && originalStatuses.get(item.channel.id) !== 'enabled')
        .map((item) => item.channel.id);
      const disableIds = orderedChannels
        .filter((item) => item.channel.status === 'disabled' && originalStatuses.get(item.channel.id) !== 'disabled')
        .map((item) => item.channel.id);

      if (disableIds.length > 0) {
        await bulkDisableMutation.mutateAsync(disableIds);
      }
      if (enableIds.length > 0) {
        await bulkEnableMutation.mutateAsync(enableIds);
      }

      const updates = orderedChannels.map((item) => ({
        id: item.channel.id,
        orderingWeight: item.orderingWeight,
      }));

      await bulkUpdateMutation.mutateAsync({
        channels: updates,
      });

      setHasChanges(false);
      onOpenChange(false);
    } catch (_error) {
      // Error is handled by the mutation hook
    }
  };

  const handleCancel = () => {
    // Reset to original order
    if (channelsData) {
      setOrderedChannels(createOrderedChannels(channelsData));
      setEnabledFirst(false);
      setHasChanges(false);
    }
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='flex max-h-[90vh] flex-col sm:max-w-5xl'>
        <DialogHeader className='flex-shrink-0 text-left'>
          <DialogTitle className='flex items-center gap-2'>
            <GripVertical className='text-muted-foreground h-5 w-5' />
            {t('channels.dialogs.bulkOrdering.title')}
          </DialogTitle>
          <DialogDescription className='text-muted-foreground text-sm'>{t('channels.dialogs.bulkOrdering.description')}</DialogDescription>
        </DialogHeader>

        <Separator className='flex-shrink-0' />

        <div className='-mr-4 h-[40rem] w-full flex-1 overflow-y-auto py-1 pr-4'>
          {isLoading ? (
            <div className='flex items-center justify-center py-12'>
              <div className='flex flex-col items-center gap-3'>
                <div className='border-primary h-8 w-8 animate-spin rounded-full border-b-2'></div>
                <div className='text-muted-foreground text-sm'>{t('common.loading', 'Loading')}...</div>
              </div>
            </div>
          ) : orderedChannels.length === 0 ? (
            <div className='flex items-center justify-center py-12'>
              <div className='flex flex-col items-center gap-3 text-center'>
                <GripVertical className='text-muted-foreground/30 h-12 w-12' />
                <div className='text-muted-foreground text-sm'>{t('channels.dialogs.bulkOrdering.noChannels')}</div>
              </div>
            </div>
          ) : (
            <div className='flex h-full flex-col gap-4 p-0.5'>
              {/* Summary Header */}
              <div className='flex items-center justify-between px-1 py-2'>
                <div className='text-muted-foreground flex items-center gap-4 text-sm'>
                  <span>{t('channels.dialogs.bulkOrdering.dragHint')}</span>
                  <Badge variant='secondary' className='font-mono'>
                    {t('channels.dialogs.bulkOrdering.channelCount', {
                      count: orderedChannels.length,
                    })}
                  </Badge>
                  <Button
                    type='button'
                    variant={enabledFirst ? 'secondary' : 'outline'}
                    size='sm'
                    className='h-8 gap-1.5'
                    aria-pressed={enabledFirst}
                    onClick={() => handleEnabledFirstChange(!enabledFirst)}
                  >
                    {enabledFirst ? <Check className='h-3.5 w-3.5' /> : <ArrowDownWideNarrow className='h-3.5 w-3.5' />}
                    {t('channels.dialogs.bulkOrdering.enabledFirst')}
                  </Button>
                  {hasChanges && (
                    <Badge
                      variant='outline'
                      className='border-amber-200 bg-amber-50 text-amber-600 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-400'
                    >
                      {t('common.unsavedChanges', 'Unsaved changes')}
                    </Badge>
                  )}
                </div>
              </div>

              {/* Channels List */}
              <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
                <SortableContext items={orderedChannels.map((item) => item.channel.id)} strategy={verticalListSortingStrategy}>
                  <div className='flex-1 space-y-1'>
                    {orderedChannels.map((item, index) => {
                      const groupStart = enabledFirst && item.channel.status === 'disabled' ? enabledCount : 0;
                      const groupEnd = enabledFirst && item.channel.status === 'enabled' ? enabledCount - 1 : orderedChannels.length - 1;

                      return (
                        <Fragment key={item.channel.id}>
                          {enabledFirst && index === enabledCount && (
                            <ChannelStatusDivider enabledCount={enabledCount} disabledCount={disabledCount} />
                          )}
                          <ChannelOrderingItemComponent
                            channel={item.channel}
                            orderingWeight={item.orderingWeight}
                            index={index}
                            onMoveToTop={handleMoveToTop}
                            onMoveToBottom={handleMoveToBottom}
                            onWeightChange={handleWeightChange}
                            onStatusChange={handleStatusChange}
                            canMoveToTop={index > groupStart}
                            canMoveToBottom={index < groupEnd}
                          />
                        </Fragment>
                      );
                    })}
                    {enabledFirst && enabledCount === orderedChannels.length && (
                      <ChannelStatusDivider enabledCount={enabledCount} disabledCount={disabledCount} />
                    )}
                  </div>
                </SortableContext>
              </DndContext>
            </div>
          )}
        </div>

        <DialogFooter className='flex-shrink-0'>
          <div className='flex w-full items-center justify-between'>
            <div className='text-muted-foreground text-xs'>
              {hasChanges && (
                <span className='flex items-center gap-1'>
                  <div className='h-2 w-2 rounded-full bg-amber-500'></div>
                  {t('common.unsavedChanges', 'You have unsaved changes')}
                </span>
              )}
            </div>
            <div className='flex items-center gap-2'>
              <Button variant='outline' onClick={handleCancel}>
                {t('common.buttons.cancel')}
              </Button>
              <Button
                onClick={handleSave}
                disabled={!hasChanges || bulkUpdateMutation.isPending || bulkEnableMutation.isPending || bulkDisableMutation.isPending}
                className='min-w-[120px]'
              >
                {bulkUpdateMutation.isPending || bulkEnableMutation.isPending || bulkDisableMutation.isPending ? (
                  <div className='flex items-center gap-2'>
                    <div className='h-4 w-4 animate-spin rounded-full border-b-2 border-white'></div>
                    {t('common.buttons.saving')}
                  </div>
                ) : (
                  t('channels.dialogs.bulkOrdering.saveButton')
                )}
              </Button>
            </div>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
