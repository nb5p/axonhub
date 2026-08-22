import { DropdownMenuTrigger } from '@radix-ui/react-dropdown-menu';
import { MixerHorizontalIcon } from '@radix-ui/react-icons';
import { Table } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { DataTableColumnSettings } from '@/components/data-table-column-settings';

interface DataTableViewOptionsProps<TData> {
  table: Table<TData>;
}

export function DataTableViewOptions<TData>({ table }: DataTableViewOptionsProps<TData>) {
  const { t } = useTranslation();
  const configurableColumns = table
    .getAllLeafColumns()
    .filter((column) => (typeof column.accessorFn !== 'undefined' || column.id === 'actions') && column.getCanHide());

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' size='sm' className='ml-auto hidden h-8 lg:flex'>
          <MixerHorizontalIcon className='mr-2 h-4 w-4' />
          {t('common.view')}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[220px]'>
        <DropdownMenuLabel>{t('common.configureColumns')}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DataTableColumnSettings
          table={table}
          columns={configurableColumns}
          getColumnLabel={(column) => (column.id === 'actions' ? t('common.columns.actions') : column.id)}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
