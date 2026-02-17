import React from 'react';
import { Search } from 'lucide-react';

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

const SearchInput: React.FC<SearchInputProps> = ({
  value,
  onChange,
  placeholder = '',
  className = '',
}) => {
  return (
    <div className={`relative flex-1 ${className}`}>
      <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
      <input
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e: React.ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        className="w-full pl-10 pr-4 py-2.5 border border-input rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-ring bg-card"
      />
    </div>
  );
};

export default SearchInput;
