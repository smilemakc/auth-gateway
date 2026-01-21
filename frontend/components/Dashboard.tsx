
import React from 'react';
import {
  Users,
  Activity,
  ShieldCheck,
  Key,
  ArrowUpRight,
  ArrowDownRight
} from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useDashboardStats } from '../hooks/useDashboard';

const Dashboard: React.FC = () => {
  const { t } = useLanguage();
  const { data: stats, isLoading, error } = useDashboardStats();

  const StatCard = ({ title, value, icon: Icon, trend, subtext, color }: any) => (
    <div className="bg-card overflow-hidden rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-muted-foreground truncate">{title}</p>
          <p className="mt-2 text-3xl font-bold text-foreground">{value.toLocaleString()}</p>
        </div>
        <div className={`p-3 rounded-lg bg-${color}-50 text-${color}-600`}>
          <Icon size={24} />
        </div>
      </div>
      <div className="mt-4 flex items-center">
        {trend && (
          <span className={`flex items-center text-sm font-medium ${trend > 0 ? 'text-success' : 'text-destructive'}`}>
            {trend > 0 ? <ArrowUpRight size={16} className="mr-1" /> : <ArrowDownRight size={16} className="mr-1" />}
            {Math.abs(trend)}%
          </span>
        )}
        <span className="ml-2 text-sm text-muted-foreground">{subtext}</span>
      </div>
    </div>
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive">Error loading dashboard: {(error as Error).message}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-foreground">{t('dash.title')}</h1>
      
      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title={t('dash.total_users')}
          value={stats?.total_users || 0}
          icon={Users}
          trend={stats?.new_users_today}
          subtext="new today"
          color="blue"
        />
        <StatCard
          title={t('dash.active_now')}
          value={stats?.active_users || 0}
          icon={Activity}
          trend={null}
          subtext="currently active"
          color="green"
        />
        <StatCard
          title="Total API Keys"
          value={stats?.total_api_keys || 0}
          icon={Key}
          trend={null}
          subtext={`${stats?.active_api_keys || 0} active`}
          color="purple"
        />
        <StatCard
          title="Login Attempts"
          value={stats?.login_attempts_today || 0}
          icon={ShieldCheck}
          trend={null}
          subtext={`${stats?.failed_login_attempts_today || 0} failed today`}
          color="amber"
        />
      </div>

      {/* Additional Stats */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Today's Activity</h3>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="p-4 bg-primary/10 rounded-lg">
            <p className="text-sm text-primary font-medium">New Users</p>
            <p className="text-2xl font-bold text-foreground mt-1">{stats?.new_users_today || 0}</p>
          </div>
          <div className="p-4 bg-success/10 rounded-lg">
            <p className="text-sm text-success font-medium">Login Attempts</p>
            <p className="text-2xl font-bold text-foreground mt-1">{stats?.login_attempts_today || 0}</p>
          </div>
          <div className="p-4 bg-destructive/10 rounded-lg">
            <p className="text-sm text-destructive font-medium">Failed Logins</p>
            <p className="text-2xl font-bold text-foreground mt-1">{stats?.failed_login_attempts_today || 0}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
