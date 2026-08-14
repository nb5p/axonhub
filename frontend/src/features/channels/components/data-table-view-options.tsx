import { DropdownMenuTrigger } from '@radix-ui/react-dropdown-menu';
import { MixerHorizontalIcon } from '@radix-ui/react-icons';
import { Table } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuLabel, DropdownMenuSeparator } from '@/components/ui/dropdown-menu';
import { DataTableColumnSettings } from '@/components/data-table-column-settings';

interface DataTableViewOptionsProps<TData> {
  table: Table<TData>;
}

export function DataTableViewOptions<TData>({ table }: DataTableViewOptionsProps<TData>) {
  const { t } = useTranslation();
  const configurableColumns = table
    .getAllLeafColumns()
    .filter((column) => typeof column.accessorFn !== 'undefined' || column.id === 'select')
    .filter((column) => column.getCanHide() || column.id === 'status')
    .filter((column) => column.id !== 'tags' && column.id !== 'model');

  const getColumnLabel = (columnId: string) => {
    const labelKey =
      columnId === 'select'
        ? 'common.columns.selection'
        : columnId === 'id'
          ? 'common.columns.id'
          : columnId === 'status'
            ? 'common.columns.status'
            : columnId === 'endpoints'
              ? 'channels.columns.supportedEndpoints'
              : columnId === 'createdAt'
                ? 'common.columns.createdAt'
                : `channels.columns.${columnId}`;
    return t(labelKey);
  };

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' size='sm'>
          <MixerHorizontalIcon className='mr-2 h-4 w-4' />
          {t('common.view')}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[220px]'>
        <DropdownMenuLabel>{t('common.configureColumns')}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DataTableColumnSettings table={table} columns={configurableColumns} getColumnLabel={(column) => getColumnLabel(column.id)} />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
