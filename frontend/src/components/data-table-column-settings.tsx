import { useId } from 'react';
import { DndContext, KeyboardSensor, PointerSensor, closestCenter, useSensor, useSensors, type DragEndEvent } from '@dnd-kit/core';
import { SortableContext, arrayMove, sortableKeyboardCoordinates, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { ResetIcon } from '@radix-ui/react-icons';
import type { Column, Table } from '@tanstack/react-table';
import { GripVertical } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Checkbox } from '@/components/ui/checkbox';
import { DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu';
import { DataTableColumnSizingReset } from '@/components/data-table-column-sizing';
import { cn } from '@/lib/utils';

interface SortableColumnItemProps<TData> {
  column: Column<TData, unknown>;
  label: string;
}

function SortableColumnItem<TData>({ column, label }: SortableColumnItemProps<TData>) {
  const { t } = useTranslation();
  const checkboxId = useId();
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: column.id });

  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Transform.toString(transform), transition }}
      className={cn(
        'focus-within:bg-accent flex min-h-8 items-center gap-1 rounded-sm pr-2 pl-1 text-sm',
        isDragging && 'bg-accent relative z-10 shadow-sm'
      )}
    >
      <button
        type='button'
        className='text-muted-foreground hover:text-foreground flex size-7 touch-none cursor-grab items-center justify-center rounded-sm outline-none active:cursor-grabbing focus-visible:ring-2 focus-visible:ring-ring'
        aria-label={t('common.dragColumn', { column: label })}
        {...attributes}
        {...listeners}
      >
        <GripVertical className='size-4' />
      </button>
      <Checkbox
        id={checkboxId}
        checked={column.getIsVisible()}
        onCheckedChange={(value) => column.toggleVisibility(value === true)}
        aria-label={label}
      />
      <label htmlFor={checkboxId} className='min-w-0 flex-1 cursor-pointer truncate py-1.5'>
        {label}
      </label>
    </div>
  );
}

interface DataTableColumnSettingsProps<TData> {
  table: Table<TData>;
  columns: Column<TData, unknown>[];
  getColumnLabel: (column: Column<TData, unknown>) => string;
}

export function DataTableColumnSettings<TData>({ table, columns, getColumnLabel }: DataTableColumnSettingsProps<TData>) {
  const { t } = useTranslation();
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );
  const order = table.getState().columnOrder;
  const orderIndex = new Map(order.map((columnId, index) => [columnId, index]));
  const orderedColumns =
    order.length === 0
      ? columns
      : [...columns].sort((left, right) => {
          const leftIndex = orderIndex.get(left.id) ?? Number.MAX_SAFE_INTEGER;
          const rightIndex = orderIndex.get(right.id) ?? Number.MAX_SAFE_INTEGER;
          return leftIndex - rightIndex;
        });

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    if (!over || active.id === over.id) return;

    const configurableIds = orderedColumns.map((column) => column.id);
    const oldIndex = configurableIds.indexOf(String(active.id));
    const newIndex = configurableIds.indexOf(String(over.id));
    if (oldIndex < 0 || newIndex < 0) return;

    const reorderedIds = arrayMove(configurableIds, oldIndex, newIndex);
    const configurableIdSet = new Set(configurableIds);
    const allColumnIds = table.getAllLeafColumns().map((column) => column.id);
    let reorderedIndex = 0;
    const nextOrder = allColumnIds.map((columnId) => {
      if (!configurableIdSet.has(columnId)) return columnId;
      const nextColumnId = reorderedIds[reorderedIndex];
      reorderedIndex += 1;
      return nextColumnId;
    });

    table.setColumnOrder(nextOrder);
  };

  return (
    <>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={orderedColumns.map((column) => column.id)} strategy={verticalListSortingStrategy}>
          {orderedColumns.map((column) => (
            <SortableColumnItem key={column.id} column={column} label={getColumnLabel(column)} />
          ))}
        </SortableContext>
      </DndContext>
      <DataTableColumnSizingReset table={table} />
      <DropdownMenuItem disabled={table.getState().columnOrder.length === 0} onSelect={() => table.resetColumnOrder(true)}>
        <ResetIcon />
        {t('common.resetColumnOrder')}
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <div className='text-muted-foreground px-2 py-1 text-xs'>{t('common.dragColumnsHint')}</div>
    </>
  );
}
