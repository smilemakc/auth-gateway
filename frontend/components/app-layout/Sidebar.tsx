import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useLanguage } from '../../services/i18n';
import {
  LayoutDashboard,
  Users,
  Key,
  ShieldAlert,
  Settings,
  LogOut,
  X,
  Globe,
  Network,
  Bot,
  Search,
  Shield,
  FolderTree,
  Server,
  FileSpreadsheet,
  ChevronDown,
  ChevronRight,
  Lock,
  Mail,
  Boxes,
  MessageSquare
} from 'lucide-react';

interface NavItem {
  path: string;
  label: string;
  icon: React.ElementType;
}

interface NavGroup {
  id: string;
  label: string;
  icon: React.ElementType;
  items: NavItem[];
}

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  onLogout: () => void;
}

function useNavGroups(): { navGroups: NavGroup[]; devItems: NavItem[] } {
  const { t } = useLanguage();

  const navGroups: NavGroup[] = [
    {
      id: 'users',
      label: t('nav.users_identity'),
      icon: Users,
      items: [
        { path: '/users', label: t('nav.users'), icon: Users },
        { path: '/groups', label: t('nav.groups'), icon: FolderTree },
        { path: '/bulk', label: t('nav.bulk_operations'), icon: FileSpreadsheet },
      ],
    },
    {
      id: 'auth',
      label: t('nav.authentication'),
      icon: Key,
      items: [
        { path: '/applications', label: t('nav.applications'), icon: Boxes },
        { path: '/sessions', label: t('nav.sessions'), icon: Key },
        { path: '/api-keys', label: t('nav.api_keys'), icon: Key },
        { path: '/oauth', label: t('nav.oauth'), icon: Globe },
        { path: '/oauth-clients', label: t('nav.oauth_clients'), icon: Shield },
      ],
    },
    {
      id: 'security',
      label: t('nav.security'),
      icon: ShieldAlert,
      items: [
        { path: '/ldap', label: 'LDAP', icon: Server },
        { path: '/saml', label: 'SAML', icon: Shield },
        { path: '/ip-security', label: t('nav.ip_security'), icon: ShieldAlert },
        { path: '/audit-logs', label: t('nav.audit_logs'), icon: ShieldAlert },
      ],
    },
    {
      id: 'messaging',
      label: t('nav.messaging'),
      icon: Mail,
      items: [
        { path: '/email/templates', label: t('nav.email_templates'), icon: Mail },
        { path: '/email/providers', label: t('nav.email_providers'), icon: Server },
        { path: '/sms/providers', label: t('nav.sms_providers'), icon: MessageSquare },
      ],
    },
  ];

  const devItems: NavItem[] = [
    { path: '/developers/webhooks', label: t('nav.webhooks'), icon: Network },
    { path: '/developers/service-accounts', label: t('nav.service_accounts'), icon: Bot },
    { path: '/developers/token-inspector', label: t('nav.token_inspector'), icon: Search },
  ];

  return { navGroups, devItems };
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen, onClose, onLogout }) => {
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set(['users', 'auth', 'security', 'messaging']));
  const location = useLocation();
  const { t } = useLanguage();
  const { navGroups, devItems } = useNavGroups();

  const toggleGroup = (groupId: string) => {
    setExpandedGroups(prev => {
      const next = new Set(prev);
      if (next.has(groupId)) {
        next.delete(groupId);
      } else {
        next.add(groupId);
      }
      return next;
    });
  };

  const isGroupActive = (group: NavGroup) => {
    return group.items.some(item =>
      location.pathname === item.path ||
      (item.path !== '/' && location.pathname.startsWith(item.path + '/'))
    );
  };

  const isItemActive = (path: string) => {
    return location.pathname === path ||
      (path !== '/' &&
       location.pathname.startsWith(path + '/') &&
       !location.pathname.startsWith('/developers'));
  };

  return (
    <aside
      className={`
        fixed inset-y-0 left-0 z-30 w-64 bg-sidebar text-sidebar-foreground transform transition-transform duration-200 ease-in-out
        lg:translate-x-0 lg:static lg:inset-0
        ${isOpen ? 'translate-x-0' : '-translate-x-full'}
      `}
    >
      <div className="flex items-center justify-between h-16 px-6 bg-sidebar border-b border-sidebar-border">
        <span className="text-xl font-bold tracking-wider text-sidebar-accent">{t('auth.title')}</span>
        <button onClick={onClose} className="lg:hidden text-sidebar-muted hover:text-sidebar-foreground">
          <X size={24} />
        </button>
      </div>

      <div className="p-4 overflow-y-auto max-h-[calc(100vh-8rem)]">
        <nav className="space-y-1 mb-4">
          <Link
            to="/"
            onClick={onClose}
            className={`
              flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
              ${location.pathname === '/'
                ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
            `}
          >
            <LayoutDashboard className="mr-3 h-5 w-5" />
            {t('nav.dashboard')}
          </Link>
          <Link
            to="/settings/access-control"
            onClick={onClose}
            className={`
              flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
              ${location.pathname.startsWith('/settings/access-control') || location.pathname.startsWith('/roles') || location.pathname.startsWith('/permissions')
                ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
            `}
          >
            <Lock className="mr-3 h-5 w-5" />
            {t('nav.access_settings')}
          </Link>
        </nav>

        <nav className="space-y-2 mb-6">
          {navGroups.map((group) => {
            const GroupIcon = group.icon;
            const isExpanded = expandedGroups.has(group.id);
            const groupActive = isGroupActive(group);

            return (
              <div key={group.id} className="space-y-1">
                <button
                  onClick={() => toggleGroup(group.id)}
                  className={`
                    w-full flex items-center justify-between px-4 py-2.5 text-sm font-medium rounded-lg transition-colors
                    ${groupActive
                      ? 'text-primary bg-primary/10'
                      : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
                  `}
                >
                  <span className="flex items-center">
                    <GroupIcon className="mr-3 h-5 w-5" />
                    {group.label}
                  </span>
                  {isExpanded ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                </button>

                {isExpanded && (
                  <div className="ml-4 pl-4 border-l border-sidebar-border space-y-1">
                    {group.items.map((item) => {
                      const Icon = item.icon;
                      return (
                        <Link
                          key={item.path}
                          to={item.path}
                          onClick={onClose}
                          className={`
                            flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors
                            ${isItemActive(item.path)
                              ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                              : 'text-sidebar-foreground/70 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
                          `}
                        >
                          <Icon className="mr-3 h-4 w-4" />
                          {item.label}
                        </Link>
                      );
                    })}
                  </div>
                )}
              </div>
            );
          })}
        </nav>

        <div className="text-xs font-semibold text-sidebar-muted uppercase tracking-wider mb-3">
          {t('nav.developers')}
        </div>
        <nav className="space-y-1 mb-4">
          {devItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname.startsWith(item.path);
            return (
              <Link
                key={item.path}
                to={item.path}
                onClick={onClose}
                className={`
                  flex items-center px-4 py-2.5 text-sm font-medium rounded-lg transition-colors
                  ${isActive
                    ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                    : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
                `}
              >
                <Icon className="mr-3 h-5 w-5" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <nav>
          <Link
            to="/settings"
            onClick={onClose}
            className={`
              flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors
              ${location.pathname === '/settings'
                ? 'bg-primary text-primary-foreground shadow-lg shadow-primary/30'
                : 'text-sidebar-foreground/80 hover:bg-sidebar-muted hover:text-sidebar-foreground'}
            `}
          >
            <Settings className="mr-3 h-5 w-5" />
            {t('nav.settings')}
          </Link>
        </nav>
      </div>

      <div className="absolute bottom-0 w-full p-4 border-t border-sidebar-border bg-sidebar">
        <button
          onClick={onLogout}
          className="flex items-center w-full px-4 py-3 text-sm font-medium text-destructive hover:bg-sidebar-muted hover:text-destructive/80 rounded-lg transition-colors"
        >
          <LogOut className="mr-3 h-5 w-5" />
          {t('nav.logout')}
        </button>
      </div>
    </aside>
  );
};

export default Sidebar;
