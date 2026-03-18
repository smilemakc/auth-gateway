import React from 'react';

interface EmptyStateProps {
  icon: React.ReactNode;
  message: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

const EmptyState: React.FC<EmptyStateProps> = ({ icon, message, action }) => {
  return (
    <div className="text-center py-12 bg-card rounded-xl border border-border">
      <div className="mx-auto mb-4 text-muted-foreground opacity-50">
        {icon}
      </div>
      <p className="text-muted-foreground">{message}</p>
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 text-primary hover:underline text-sm font-medium"
        >
          {action.label}
        </button>
      )}
    </div>
  );
};

export default EmptyState;
