import { useState, useMemo } from 'react';

export type SortDirection = 'asc' | 'desc' | null;

export interface SortState {
  key: string | null;
  direction: SortDirection;
}

export function useSort<T>(defaultKey?: string, defaultDirection?: SortDirection) {
  const [sortState, setSortState] = useState<SortState>({
    key: defaultKey || null,
    direction: defaultDirection || null,
  });

  const requestSort = (key: string) => {
    setSortState((prev) => {
      if (prev.key !== key) return { key, direction: 'asc' };
      if (prev.direction === 'asc') return { key, direction: 'desc' };
      if (prev.direction === 'desc') return { key: null, direction: null };
      return { key, direction: 'asc' };
    });
  };

  const sortData = useMemo(() => {
    return (items: T[]): T[] => {
      if (!sortState.key || !sortState.direction) return items;

      const sorted = [...items].sort((a, b) => {
        const aVal = (a as any)[sortState.key!];
        const bVal = (b as any)[sortState.key!];

        if (aVal == null && bVal == null) return 0;
        if (aVal == null) return 1;
        if (bVal == null) return -1;

        if (typeof aVal === 'string' && typeof bVal === 'string') {
          return aVal.localeCompare(bVal, undefined, { sensitivity: 'base' });
        }

        if (aVal < bVal) return -1;
        if (aVal > bVal) return 1;
        return 0;
      });

      return sortState.direction === 'desc' ? sorted.reverse() : sorted;
    };
  }, [sortState.key, sortState.direction]);

  return { sortState, requestSort, sortData };
}
