const FILTER_STORAGE_PREFIX = 'axonhub:filters:v1';
const DATE_MARKER = '__axonhubFilterDate';

type PersistedDate = {
  [DATE_MARKER]: string;
};

export function buildFilterStorageKey(scope: string, name: string) {
  return `${FILTER_STORAGE_PREFIX}:${scope}:${name}`;
}

export function serializeFilterValue<T>(value: T): string | undefined {
  return JSON.stringify(value, function (this: Record<string, unknown>, key, nestedValue) {
    const originalValue = key === '' ? value : this[key];
    if (originalValue instanceof Date) {
      return { [DATE_MARKER]: originalValue.toISOString() } satisfies PersistedDate;
    }
    return nestedValue;
  });
}

export function deserializeFilterValue<T>(value: string): T {
  return JSON.parse(value, (_key, nestedValue: unknown) => {
    if (
      typeof nestedValue === 'object' &&
      nestedValue !== null &&
      DATE_MARKER in nestedValue &&
      typeof (nestedValue as PersistedDate)[DATE_MARKER] === 'string'
    ) {
      const date = new Date((nestedValue as PersistedDate)[DATE_MARKER]);
      if (!Number.isNaN(date.getTime())) {
        return date;
      }
    }
    return nestedValue;
  }) as T;
}
