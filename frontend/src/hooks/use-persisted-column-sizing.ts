import { useEffect, useState } from 'react';
import type { ColumnOrderState, ColumnSizingState } from '@tanstack/react-table';

function parseColumnSizing(value: string | null): ColumnSizingState {
  if (!value) return {};

  try {
    const parsed: unknown = JSON.parse(value);
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {};

    return Object.fromEntries(
      Object.entries(parsed).filter(
        (entry): entry is [string, number] => typeof entry[1] === 'number' && Number.isFinite(entry[1]) && entry[1] > 0
      )
    );
  } catch {
    return {};
  }
}

export function usePersistedColumnSizing(storageKey: string) {
  const [columnSizing, setColumnSizing] = useState<ColumnSizingState>(() => {
    try {
      return parseColumnSizing(localStorage.getItem(storageKey));
    } catch {
      return {};
    }
  });

  useEffect(() => {
    try {
      if (Object.keys(columnSizing).length === 0) {
        localStorage.removeItem(storageKey);
      } else {
        localStorage.setItem(storageKey, JSON.stringify(columnSizing));
      }
    } catch {
      // Keep column resizing available for this session when storage is unavailable.
    }
  }, [columnSizing, storageKey]);

  return [columnSizing, setColumnSizing] as const;
}

function parseColumnOrder(value: string | null): ColumnOrderState {
  if (!value) return [];

  try {
    const parsed: unknown = JSON.parse(value);
    if (!Array.isArray(parsed)) return [];

    return [...new Set(parsed.filter((columnId): columnId is string => typeof columnId === 'string' && columnId.length > 0))];
  } catch {
    return [];
  }
}

export function usePersistedColumnOrder(storageKey: string) {
  const [columnOrder, setColumnOrder] = useState<ColumnOrderState>(() => {
    try {
      return parseColumnOrder(localStorage.getItem(storageKey));
    } catch {
      return [];
    }
  });

  useEffect(() => {
    try {
      if (columnOrder.length === 0) {
        localStorage.removeItem(storageKey);
      } else {
        localStorage.setItem(storageKey, JSON.stringify(columnOrder));
      }
    } catch {
      // Keep column ordering available for this session when storage is unavailable.
    }
  }, [columnOrder, storageKey]);

  return [columnOrder, setColumnOrder] as const;
}
