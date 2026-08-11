import { useMemo, useEffect } from 'react';
import { Cross2Icon } from '@radix-ui/react-icons';
import { IconChevronsDown, IconChevronsUp, IconSearch } from '@tabler/icons-react';
import { Table } from '@tanstack/react-table';
import { useQueryModels } from '@/gql/models';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { DataTableFacetedFilter } from '@/components/data-table-faceted-filter';
import { useHorizontalScroll } from '@/hooks/use-horizontal-scroll';
import { useAllChannelTags } from '../data/channels';
import { useChannels } from '../context/channels-context';
import { CHANNEL_CONFIGS } from '../data/config_channels';
import { ChannelModelsMatchMode } from '../data/channels';
import { channelEndpointFilterApiFormats } from '../data/schema';
import { DataTableViewOptions } from './data-table-view-options';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  isFiltered?: boolean;
  selectedCount?: number;
  selectedTypeTab?: string;
  showErrorOnly?: boolean;
  onExitErrorOnlyMode?: () => void;
  modelMatchMode: ChannelModelsMatchMode;
  onModelMatchModeChange: (mode: ChannelModelsMatchMode) => void;
}

export function DataTableToolbar<TData>({
  table,
  isFiltered: externalIsFiltered,
  selectedCount: externalSelectedCount,
  selectedTypeTab = 'all',
  showErrorOnly,
  onExitErrorOnlyMode,
  modelMatchMode,
  onModelMatchModeChange,
}: DataTableToolbarProps<TData>) {
  const { t } = useTranslation();
  const scrollRef = useHorizontalScroll<HTMLDivElement>();
  const { showTypeTabs, setShowTypeTabs } = useChannels();
  const tableState = table.getState();
  const isFiltered = externalIsFiltered ?? tableState.columnFilters.length > 0;

  // Get all channel tags from GraphQL
  const { data: allTags = [] } = useAllChannelTags();

  // Fetch models using the models query
  const { mutate: fetchModels, data: modelsData } = useQueryModels();

  // Fetch models on component mount
  useEffect(() => {
    fetchModels({
      statusIn: ['enabled', 'disabled'],
      includeAllChannelModels: true,
    });
  }, [fetchModels]);

  const tagOptions = useMemo(
    () =>
      allTags.map((tag) => ({
        value: tag,
        label: tag,
      })),
    [allTags]
  );

  const modelOptions = useMemo(() => {
    if (!modelsData) return [];
    return modelsData.map((model) => ({
      value: model.id,
      label: model.id,
    }));
  }, [modelsData]);

  const modelFilterValue = table.getColumn('model')?.getFilterValue();
  const selectedModelCount = Array.isArray(modelFilterValue) ? modelFilterValue.length : 0;

  const endpointOptions = useMemo(
    () =>
      channelEndpointFilterApiFormats.map((value) => {
        const key = `channels.dialogs.fields.apiFormat.formats.${value}`;
        const label = t(key);
        return { value, label: label === key ? value : label };
      }),
    [t]
  );

  // Generate channel types from CHANNEL_CONFIGS
  const channelTypes = useMemo(
    () =>
      Object.values(CHANNEL_CONFIGS).map((config) => ({
        value: config.channelType,
        label: t(`channels.types.${config.channelType}`),
      })),
    [t]
  );

  const channelStatuses = useMemo(
    () => [
      {
        value: 'enabled',
        label: t('channels.status.enabled'),
      },
      {
        value: 'disabled',
        label: t('channels.status.disabled'),
      },
      {
        value: 'archived',
        label: t('channels.status.archived'),
      },
    ],
    [t]
  );

  return (
    <div ref={scrollRef} className='flex items-center gap-4 overflow-x-auto pb-2 md:overflow-x-visible md:pb-0'>
      <div className='relative w-[150px] shrink-0 lg:flex-1 lg:w-auto'>
        <IconSearch className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
        <Input
          placeholder={t('channels.filters.filterByName')}
          value={(table.getColumn('name')?.getFilterValue() as string) ?? ''}
          onChange={(event) => table.getColumn('name')?.setFilterValue(event.target.value)}
          className='h-8 pl-8'
        />
      </div>
      <Button variant='outline' size='sm' className='h-8' onClick={() => setShowTypeTabs(!showTypeTabs)}>
        {showTypeTabs ? (
          <IconChevronsDown className='mr-1 h-4 w-4' />
        ) : (
          <IconChevronsUp className='mr-1 h-4 w-4' />
        )}
        {t('channels.filters.providerToggle')}
      </Button>
      {table.getColumn('status') && (
        <DataTableFacetedFilter column={table.getColumn('status')} title={t('channels.filters.status')} options={channelStatuses} />
      )}
      {table.getColumn('tags') && tagOptions?.length > 0 && (
        <DataTableFacetedFilter column={table.getColumn('tags')} title={t('channels.filters.tags')} options={tagOptions} singleSelect />
      )}
      {table.getColumn('model') && modelOptions?.length > 0 && (
        <DataTableFacetedFilter
          column={table.getColumn('model')}
          title={t('channels.filters.model')}
          options={modelOptions}
          selectedFirst
          selectionSummaryThreshold={1}
          selectionCountLabel={(count) => t('channels.filters.selectedModels', { count })}
          selectionControl={
            selectedModelCount > 1 ? (
              <span
                className='hover:bg-accent hover:text-accent-foreground -my-1 rounded px-1.5 py-1 font-medium'
                title={t(modelMatchMode === 'any' ? 'channels.filters.modelMatchAnyDescription' : 'channels.filters.modelMatchAllDescription')}
                onPointerDown={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                }}
                onClick={(event) => {
                  event.preventDefault();
                  event.stopPropagation();
                  onModelMatchModeChange(modelMatchMode === 'any' ? 'all' : 'any');
                }}
              >
                {t(modelMatchMode === 'any' ? 'channels.filters.modelMatchAny' : 'channels.filters.modelMatchAll')}
              </span>
            ) : undefined
          }
        />
      )}
      {table.getColumn('endpoints') && (
        <DataTableFacetedFilter
          column={table.getColumn('endpoints')}
          title={t('channels.filters.endpoint')}
          options={endpointOptions}
          selectedFirst
          selectionSummaryThreshold={1}
          selectionCountLabel={(count) => t('channels.filters.selectedEndpoints', { count })}
          contentClassName='w-[280px]'
        />
      )}
      {isFiltered && (
        <Button
          variant='ghost'
          onClick={() => table.resetColumnFilters()}
          className='h-8 px-2 lg:px-3'
        >
          {t('common.filters.reset')}
          <Cross2Icon className='ml-2 h-4 w-4' />
        </Button>
      )}
      {showErrorOnly && onExitErrorOnlyMode && (
        <Button
          variant='outline'
          onClick={onExitErrorOnlyMode}
          className='h-8 border-orange-600 text-orange-600 hover:bg-orange-600 hover:text-white'
        >
          {t('channels.errorBanner.exitErrorOnlyButton')}
        </Button>
      )}
      <DataTableViewOptions table={table} />
    </div>
  );
}
