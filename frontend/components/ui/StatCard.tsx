import React from 'react';

interface StatCardProps {
  icon: React.ReactNode;
  iconBgClass?: string;
  value: number | string;
  label: string;
}

const StatCard: React.FC<StatCardProps> = ({
  icon,
  iconBgClass = 'bg-primary/10',
  value,
  label,
}) => {
  return (
    <div className="bg-card border border-border rounded-xl p-4">
      <div className="flex items-center gap-3">
        <div className={`p-2 ${iconBgClass} rounded-lg`}>
          {icon}
        </div>
        <div>
          <p className="text-2xl font-bold text-foreground">{value}</p>
          <p className="text-sm text-muted-foreground">{label}</p>
        </div>
      </div>
    </div>
  );
};

export default StatCard;
