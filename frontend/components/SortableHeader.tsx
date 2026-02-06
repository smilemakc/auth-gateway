import React from 'react';
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react';
import type { SortDirection } from '../hooks/useSort';

interface SortableHeaderProps {
  label: string;
  sortKey: string;
  currentSortKey: string | null;
  currentDirection: SortDirection;
  onSort: (key: string) => void;
  className?: string;
}

const SortableHeader: React.FC<SortableHeaderProps> = ({
  label,
  sortKey,
  currentSortKey,
  currentDirection,
  onSort,
  className = '',
}) => {
  const isActive = currentSortKey === sortKey;

  return (
    <th
      scope="col"
      className={`px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider cursor-pointer select-none hover:text-foreground transition-colors ${className}`}
      onClick={() => onSort(sortKey)}
    >
      <div className="flex items-center gap-1">
        {label}
        <span className="inline-flex">
          {isActive && currentDirection === 'asc' && <ChevronUp size={14} />}
          {isActive && currentDirection === 'desc' && <ChevronDown size={14} />}
          {!isActive && <ChevronsUpDown size={14} className="opacity-40" />}
        </span>
      </div>
    </th>
  );
};

export default SortableHeader;
