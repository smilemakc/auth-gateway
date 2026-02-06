
import React from 'react';
import {
  Users,
  Activity,
  ShieldCheck,
  Key,
  ArrowUpRight,
  ArrowDownRight,
  Plus,
  Monitor,
  FileText,
} from 'lucide-react';
import { Link } from 'react-router-dom';
import { useLanguage } from '../services/i18n';
import { useDashboardStats } from '../hooks/useDashboard';
import { useAuditLogs } from '../hooks/useAuditLogs';
import { formatRelative } from '../lib/date';

const Dashboard: React.FC = () => {
  const { t } = useLanguage();
  const { data: stats, isLoading, error } = useDashboardStats();
  const { data: recentActivity } = useAuditLogs(1, 5);

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
        <p className="text-destructive">{t('dash.error_loading')}: {(error as Error).message}</p>
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
          subtext={t('dash.new_today')}
          color="blue"
        />
        <StatCard
          title={t('dash.active_now')}
          value={stats?.active_users || 0}
          icon={Activity}
          trend={null}
          subtext={t('dash.currently_active')}
          color="green"
        />
        <StatCard
          title={t('dash.total_api_keys')}
          value={stats?.total_api_keys || 0}
          icon={Key}
          trend={null}
          subtext={`${stats?.active_api_keys || 0} ${t('dash.active_keys')}`}
          color="purple"
        />
        <StatCard
          title={t('dash.login_attempts')}
          value={stats?.login_attempts_today || 0}
          icon={ShieldCheck}
          trend={null}
          subtext={`${stats?.failed_login_attempts_today || 0} ${t('dash.failed_today')}`}
          color="amber"
        />
      </div>

      {/* Quick Actions */}
      <div className="bg-card rounded-xl shadow-sm border border-border p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">{t('dash.quick_actions')}</h3>
        <div className="flex flex-wrap gap-3">
          <Link
            to="/users/new"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card hover:bg-accent text-sm font-medium text-foreground transition-colors"
          >
            <Plus size={16} />
            {t('common.create')} {t('nav.users').toLowerCase()}
          </Link>
          <Link
            to="/sessions"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card hover:bg-accent text-sm font-medium text-foreground transition-colors"
          >
            <Monitor size={16} />
            {t('nav.sessions')}
          </Link>
          <Link
            to="/audit-logs"
            className="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border bg-card hover:bg-accent text-sm font-medium text-foreground transition-colors"
          >
            <FileText size={16} />
            {t('nav.audit_logs')}
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Today's Activity */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h3 className="text-lg font-semibold text-foreground mb-4">{t('dash.todays_activity')}</h3>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="p-4 bg-primary/10 rounded-lg">
              <p className="text-sm text-primary font-medium">{t('dash.new_users')}</p>
              <p className="text-2xl font-bold text-foreground mt-1">{stats?.new_users_today || 0}</p>
            </div>
            <div className="p-4 bg-success/10 rounded-lg">
              <p className="text-sm text-success font-medium">{t('dash.login_attempts')}</p>
              <p className="text-2xl font-bold text-foreground mt-1">{stats?.login_attempts_today || 0}</p>
            </div>
            <div className="p-4 bg-destructive/10 rounded-lg">
              <p className="text-sm text-destructive font-medium">{t('dash.failed_logins')}</p>
              <p className="text-2xl font-bold text-foreground mt-1">{stats?.failed_login_attempts_today || 0}</p>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-foreground">{t('dash.recent_activity')}</h3>
            <Link to="/audit-logs" className="text-sm text-primary hover:underline">
              {t('dash.view_all')}
            </Link>
          </div>
          {recentActivity?.logs && recentActivity.logs.length > 0 ? (
            <div className="space-y-3">
              {recentActivity.logs.map((log: any) => (
                <div key={log.id} className="flex items-center justify-between py-2 border-b border-border last:border-0">
                  <div className="min-w-0 flex-1">
                    <p className="text-sm font-medium text-foreground truncate">
                      {log.action}
                      <span className="text-muted-foreground font-normal"> — {log.resource}</span>
                    </p>
                    <p className="text-xs text-muted-foreground truncate">
                      {log.user_email || log.user_id}
                    </p>
                  </div>
                  <div className="ml-4 flex items-center gap-2 shrink-0">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                      log.status === 'success'
                        ? 'bg-success/10 text-success'
                        : 'bg-destructive/10 text-destructive'
                    }`}>
                      {log.status}
                    </span>
                    <span className="text-xs text-muted-foreground whitespace-nowrap">
                      {formatRelative(log.created_at)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground text-center py-4">{t('audit.no_logs')}</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
