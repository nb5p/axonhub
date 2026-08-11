import { useEffect, useRef, useState, type Dispatch, type SetStateAction } from 'react';
import { buildFilterStorageKey, deserializeFilterValue, serializeFilterValue } from '@/lib/filter-storage';

function resolveInitialValue<T>(initialValue: T | (() => T)): T {
  return typeof initialValue === 'function' ? (initialValue as () => T)() : initialValue;
}

/**
 * Persists a page-level filter in localStorage and mirrors updates from other
 * browser tabs. Malformed or unavailable storage falls back to the page's
 * normal default without blocking the filter UI.
 */
export function usePersistedFilter<T>(scope: string, name: string, initialValue: T | (() => T)): [T, Dispatch<SetStateAction<T>>] {
  const storageKey = buildFilterStorageKey(scope, name);
  const fallbackRef = useRef<T>(undefined as T);
  const initializedRef = useRef(false);

  if (!initializedRef.current) {
    fallbackRef.current = resolveInitialValue(initialValue);
    initializedRef.current = true;
  }

  const [value, setValue] = useState<T>(() => {
    try {
      const stored = localStorage.getItem(storageKey);
      return stored === null ? fallbackRef.current : deserializeFilterValue<T>(stored);
    } catch {
      return fallbackRef.current;
    }
  });

  useEffect(() => {
    try {
      const serialized = serializeFilterValue(value);
      if (serialized === undefined) {
        localStorage.removeItem(storageKey);
      } else {
        localStorage.setItem(storageKey, serialized);
      }
    } catch {
      // Storage may be unavailable; keep the current-tab state functional.
    }
  }, [storageKey, value]);

  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.storageArea !== localStorage || event.key !== storageKey) return;

      if (event.newValue === null) {
        setValue(fallbackRef.current);
        return;
      }

      try {
        setValue(deserializeFilterValue<T>(event.newValue));
      } catch {
        setValue(fallbackRef.current);
      }
    };

    window.addEventListener('storage', handleStorage);
    return () => window.removeEventListener('storage', handleStorage);
  }, [storageKey]);

  return [value, setValue];
}
