import { DropdownMenuTrigger } from '@radix-ui/react-dropdown-menu';
import { MixerHorizontalIcon } from '@radix-ui/react-icons';
import { Table } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu';
import { DataTableColumnSizingReset } from '@/components/data-table-column-sizing';

interface DataTableViewOptionsProps<TData> {
  table: Table<TData>;
}

export function DataTableViewOptions<TData>({ table }: DataTableViewOptionsProps<TData>) {
  const { t } = useTranslation();
  const columnLabels: Record<string, string> = {
    select: t('common.columns.selection'),
    id: t('common.columns.id'),
    name: t('common.columns.name'),
    key: t('apikeys.columns.key'),
    creator: t('apikeys.columns.creator'),
    type: t('apikeys.columns.type'),
    status: t('common.columns.status'),
    activeProfile: t('apikeys.columns.activeProfile'),
    createdAt: t('common.columns.createdAt'),
    updatedAt: t('common.columns.updatedAt'),
  };

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button variant='outline' size='sm' className='ml-2 h-8 shrink-0'>
          <MixerHorizontalIcon className='mr-2 h-4 w-4' />
          {t('common.view')}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[180px]'>
        <DropdownMenuLabel>{t('common.toggleColumns')}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {table
          .getAllColumns()
          .filter((column) => {
            const accessorKey = column.columnDef.accessorKey;
            return (
              (typeof column.accessorFn !== 'undefined' || typeof accessorKey !== 'undefined' || column.id === 'select') &&
              column.getCanHide()
            );
          })
          .map((column) => {
            return (
              <DropdownMenuCheckboxItem
                key={column.id}
                className='capitalize'
                checked={column.getIsVisible()}
                onCheckedChange={(value) => column.toggleVisibility(!!value)}
              >
                {columnLabels[column.id] ?? column.id}
              </DropdownMenuCheckboxItem>
            );
          })}
        <DataTableColumnSizingReset table={table} />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
