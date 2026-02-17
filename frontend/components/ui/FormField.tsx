import React from 'react';

interface FormFieldProps {
  label: string;
  required?: boolean;
  error?: string;
  children: React.ReactNode;
  className?: string;
}

const FormField: React.FC<FormFieldProps> = ({ label, required, error, children, className }) => {
  return (
    <div className={className}>
      <label className="block text-sm font-medium text-foreground mb-1">
        {label} {required && '*'}
      </label>
      {children}
      {error && (
        <p className="text-xs text-destructive mt-1">{error}</p>
      )}
    </div>
  );
};

export default FormField;
