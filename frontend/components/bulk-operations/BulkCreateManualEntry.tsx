import React from 'react';
import { FileText, ToggleLeft, ToggleRight } from 'lucide-react';
import type { BulkUserCreate } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface BulkCreateManualEntryProps {
  users: BulkUserCreate[];
  onAddUser: () => void;
  onRemoveUser: (index: number) => void;
  onUserChange: (index: number, field: keyof BulkUserCreate, value: string | boolean) => void;
}

const BulkCreateManualEntry: React.FC<BulkCreateManualEntryProps> = ({
  users,
  onAddUser,
  onRemoveUser,
  onUserChange,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-foreground">{t('bulk.users_count')} {users.length}</h2>
        <button
          onClick={onAddUser}
          className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors"
        >
          + {t('bulk.add_user')}
        </button>
      </div>

      <div className="space-y-4">
        {users.map((user, index) => (
          <div key={index} className="border border-border rounded-lg p-4">
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm font-medium text-foreground">{t('bulk.user_number')} {index + 1}</span>
              <button
                onClick={() => onRemoveUser(index)}
                className="text-destructive hover:text-destructive/80 text-sm"
              >
                {t('common.remove')}
              </button>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_email')} *</label>
                <input
                  type="email"
                  value={user.email}
                  onChange={(e) => onUserChange(index, 'email', e.target.value)}
                  className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="user@example.com"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_username')} *</label>
                <input
                  type="text"
                  value={user.username}
                  onChange={(e) => onUserChange(index, 'username', e.target.value)}
                  className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="username"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_password')} *</label>
                <input
                  type="password"
                  value={user.password}
                  onChange={(e) => onUserChange(index, 'password', e.target.value)}
                  className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="password"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_full_name')}</label>
                <input
                  type="text"
                  value={user.full_name}
                  onChange={(e) => onUserChange(index, 'full_name', e.target.value)}
                  className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder={t('users.col_full_name')}
                />
              </div>
              <div className="flex items-center gap-4">
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => onUserChange(index, 'is_active', !user.is_active)}
                    className={`transition-colors ${user.is_active ? 'text-success' : 'text-muted-foreground'}`}
                  >
                    {user.is_active ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </button>
                  <span className="text-xs text-foreground">{t('users.active')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    onClick={() => onUserChange(index, 'email_verified', !user.email_verified)}
                    className={`transition-colors ${user.email_verified ? 'text-success' : 'text-muted-foreground'}`}
                  >
                    {user.email_verified ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                  </button>
                  <span className="text-xs text-foreground">{t('bulk.email_verified')}</span>
                </div>
              </div>
            </div>
          </div>
        ))}

        {users.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <FileText size={48} className="mx-auto mb-2 text-muted-foreground" />
            <p>{t('bulk.no_users_added')}</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default BulkCreateManualEntry;
