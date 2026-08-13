import { useState } from 'react';
import type { CSSProperties, KeyboardEvent, PointerEvent as ReactPointerEvent } from 'react';
import { ResetIcon } from '@radix-ui/react-icons';
import type { Column, Header, Table } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { DropdownMenuItem, DropdownMenuSeparator } from '@/components/ui/dropdown-menu';

function getColumnBounds<TData>(column: Column<TData, unknown>) {
  return {
    min: column.columnDef.minSize ?? 20,
    max: column.columnDef.maxSize ?? Number.MAX_SAFE_INTEGER,
  };
}

function getMeasuredColumnSizes<TData>(table: Table<TData>, tableElement: HTMLTableElement) {
  return Object.fromEntries(
    table.getVisibleLeafColumns().map((column) => {
      const headerCell = tableElement.querySelector<HTMLElement>(`th[data-column-id="${CSS.escape(column.id)}"]`);
      return [column.id, headerCell?.getBoundingClientRect().width ?? column.getSize()];
    })
  );
}

function getResizePair<TData>(table: Table<TData>, columnId: string) {
  const columns = table.getVisibleLeafColumns().filter((column) => column.getCanResize());
  const columnIndex = columns.findIndex((column) => column.id === columnId);
  if (columnIndex < 0 || columns.length < 2) return null;

  const neighbor = columns[columnIndex + 1];
  if (!neighbor) return null;

  return {
    column: columns[columnIndex],
    neighbor,
  };
}

function resizeColumnPair<TData>(
  table: Table<TData>,
  measuredSizes: Record<string, number>,
  column: Column<TData, unknown>,
  neighbor: Column<TData, unknown>,
  requestedDelta: number
) {
  const columnSize = measuredSizes[column.id] ?? column.getSize();
  const neighborSize = measuredSizes[neighbor.id] ?? neighbor.getSize();
  const columnBounds = getColumnBounds(column);
  const neighborBounds = getColumnBounds(neighbor);
  const minimumDelta = Math.max(columnBounds.min - columnSize, neighborSize - neighborBounds.max);
  const maximumDelta = Math.min(columnBounds.max - columnSize, neighborSize - neighborBounds.min);
  const delta = Math.min(maximumDelta, Math.max(minimumDelta, requestedDelta));

  table.setColumnSizing((current) => ({
    ...current,
    ...measuredSizes,
    [column.id]: columnSize + delta,
    [neighbor.id]: neighborSize - delta,
  }));
}

interface DataTableColumnResizerProps<TData, TValue> {
  header: Header<TData, TValue>;
}

export function DataTableColumnResizer<TData, TValue>({ header }: DataTableColumnResizerProps<TData, TValue>) {
  const [isResizing, setIsResizing] = useState(false);
  const { column } = header;
  const table = header.getContext().table;
  const resizePair = getResizePair(table, column.id);

  if (!column.getCanResize() || !resizePair) return null;

  const beginResize = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (!event.isPrimary || event.button !== 0) return;

    const tableElement = event.currentTarget.closest('table');
    if (!tableElement) return;

    event.preventDefault();
    event.stopPropagation();
    event.currentTarget.setPointerCapture(event.pointerId);

    const startX = event.clientX;
    const measuredSizes = getMeasuredColumnSizes(table, tableElement);
    const previousCursor = document.body.style.cursor;
    const previousUserSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
    setIsResizing(true);

    const handlePointerMove = (moveEvent: PointerEvent) => {
      resizeColumnPair(table, measuredSizes, resizePair.column, resizePair.neighbor, moveEvent.clientX - startX);
    };
    const finishResize = () => {
      document.removeEventListener('pointermove', handlePointerMove);
      document.removeEventListener('pointerup', finishResize);
      document.removeEventListener('pointercancel', finishResize);
      document.body.style.cursor = previousCursor;
      document.body.style.userSelect = previousUserSelect;
      setIsResizing(false);
    };

    document.addEventListener('pointermove', handlePointerMove);
    document.addEventListener('pointerup', finishResize);
    document.addEventListener('pointercancel', finishResize);
  };

  const resizeWithKeyboard = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;

    const tableElement = event.currentTarget.closest('table');
    if (!tableElement) return;

    event.preventDefault();
    event.stopPropagation();
    const direction = event.key === 'ArrowRight' ? 1 : -1;
    const step = event.shiftKey ? 24 : 8;
    resizeColumnPair(table, getMeasuredColumnSizes(table, tableElement), resizePair.column, resizePair.neighbor, direction * step);
  };

  return (
    <div
      role='separator'
      aria-orientation='vertical'
      aria-label={column.id}
      aria-valuemin={column.columnDef.minSize}
      aria-valuemax={column.columnDef.maxSize}
      aria-valuenow={Math.round(column.getSize())}
      data-testid={`column-resizer-${column.id}`}
      tabIndex={0}
      className={cn(
        'absolute top-0 right-0 z-10 h-full w-3 translate-x-1/2 cursor-col-resize touch-none outline-none select-none',
        'after:bg-border hover:after:bg-primary focus-visible:after:bg-primary after:absolute after:top-1/4 after:right-1/2 after:h-1/2 after:w-px after:translate-x-1/2 after:transition-colors',
        isResizing && 'after:bg-primary after:w-0.5'
      )}
      onKeyDown={resizeWithKeyboard}
      onPointerDown={beginResize}
    />
  );
}

export function DataTableColGroup<TData>({ table }: { table: Table<TData> }) {
  const columns = table.getVisibleLeafColumns();
  const hasCustomSizing = columns.some((column) => table.getState().columnSizing[column.id] !== undefined);
  if (!hasCustomSizing) return null;

  const totalSize = columns.reduce((total, column) => total + column.getSize(), 0);

  return (
    <colgroup>
      {columns.map((column) => (
        <col key={column.id} style={{ width: `${(column.getSize() / totalSize) * 100}%` }} />
      ))}
    </colgroup>
  );
}

export function getDataTableSizingStyle<TData>(table: Table<TData>): CSSProperties {
  const columns = table.getVisibleLeafColumns();
  const minimumWidth = columns.reduce((total, column) => total + (column.columnDef.minSize ?? 20), 0);
  const hasCustomSizing = columns.some((column) => table.getState().columnSizing[column.id] !== undefined);

  if (!hasCustomSizing) return { minWidth: `${minimumWidth}px` };

  const totalSize = columns.reduce((total, column) => total + column.getSize(), 0);
  const constrainedMinimumWidth = columns.reduce((requiredWidth, column) => {
    const columnSize = column.getSize();
    if (columnSize <= 0 || totalSize <= 0) return requiredWidth;
    return Math.max(requiredWidth, ((column.columnDef.minSize ?? 20) * totalSize) / columnSize);
  }, minimumWidth);

  return { minWidth: `${Math.ceil(constrainedMinimumWidth)}px` };
}

export function getDataTableLayoutClass<TData>(table: Table<TData>) {
  const hasCustomSizing = table.getVisibleLeafColumns().some((column) => table.getState().columnSizing[column.id] !== undefined);
  return hasCustomSizing ? 'table-fixed' : 'table-auto';
}

export function DataTableColumnSizingReset<TData>({ table }: { table: Table<TData> }) {
  const { t } = useTranslation();

  return (
    <>
      <DropdownMenuSeparator />
      <DropdownMenuItem disabled={Object.keys(table.getState().columnSizing).length === 0} onSelect={() => table.resetColumnSizing(true)}>
        <ResetIcon />
        {t('common.resetColumnWidths')}
      </DropdownMenuItem>
    </>
  );
}
