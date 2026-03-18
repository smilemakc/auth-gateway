import React from 'react';

interface TextInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'className'> {
  variant?: 'default' | 'compact';
  datalistId?: string;
  datalistOptions?: string[];
}

const TextInput: React.FC<TextInputProps> = ({
  variant = 'default',
  datalistId,
  datalistOptions,
  ...props
}) => {
  const padding = variant === 'compact' ? 'px-3 py-2' : 'px-4 py-2';

  return (
    <>
      <input
        {...props}
        list={datalistId}
        className={`w-full ${padding} border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none text-sm`}
      />
      {datalistId && datalistOptions && (
        <datalist id={datalistId}>
          {datalistOptions.map((opt) => (
            <option key={opt} value={opt} />
          ))}
        </datalist>
      )}
    </>
  );
};

export default TextInput;
