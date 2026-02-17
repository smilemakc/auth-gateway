import React from 'react';
import { Search } from 'lucide-react';
import type { AdminUserResponse } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface UpdateFields {
  email?: string;
  username?: string;
  full_name?: string;
  is_active?: boolean;
}

interface BulkUpdateFieldSelectorProps {
  updateFields: UpdateFields;
  onUpdateFieldsChange: (fields: UpdateFields) => void;
  filteredUsers: AdminUserResponse[];
  selectedUserIds: string[];
  searchTerm: string;
  onSearchTermChange: (term: string) => void;
  onToggleUserSelection: (userId: string) => void;
  onSelectAll: () => void;
}

const BulkUpdateFieldSelector: React.FC<BulkUpdateFieldSelectorProps> = ({
  updateFields,
  onUpdateFieldsChange,
  filteredUsers,
  selectedUserIds,
  searchTerm,
  onSearchTermChange,
  onToggleUserSelection,
  onSelectAll,
}) => {
  const { t } = useLanguage();

  return (
    <>
      <div>
        <h2 className="text-lg font-semibold text-foreground mb-4">{t('bulk_update.update_fields')}</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('auth.email')}</label>
            <input
              type="email"
              value={updateFields.email || ''}
              onChange={(e) => onUpdateFieldsChange({ ...updateFields, email: e.target.value || undefined })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder={t('bulk_update.leave_empty')}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('common.username')}</label>
            <input
              type="text"
              value={updateFields.username || ''}
              onChange={(e) => onUpdateFieldsChange({ ...updateFields, username: e.target.value || undefined })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder={t('bulk_update.leave_empty')}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('user.form.fullname')}</label>
            <input
              type="text"
              value={updateFields.full_name || ''}
              onChange={(e) => onUpdateFieldsChange({ ...updateFields, full_name: e.target.value || undefined })}
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder={t('bulk_update.leave_empty')}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">{t('common.status')}</label>
            <select
              value={updateFields.is_active === undefined ? '' : updateFields.is_active ? 'true' : 'false'}
              onChange={(e) =>
                onUpdateFieldsChange({
                  ...updateFields,
                  is_active: e.target.value === '' ? undefined : e.target.value === 'true',
                })
              }
              className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">{t('bulk_update.keep_current')}</option>
              <option value="true">{t('common.active')}</option>
              <option value="false">{t('common.inactive')}</option>
            </select>
          </div>
        </div>
      </div>

      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">
            {t('bulk_update.select_users')} ({selectedUserIds.length} {t('bulk_update.selected')})
          </h2>
          <div className="flex items-center gap-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" size={18} />
              <input
                type="text"
                placeholder={t('common.search')}
                value={searchTerm}
                onChange={(e) => onSearchTermChange(e.target.value)}
                className="pl-10 pr-4 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <button
              onClick={onSelectAll}
              className="px-3 py-2 border border-input rounded-lg text-sm text-foreground hover:bg-accent"
            >
              {selectedUserIds.length === filteredUsers.length ? t('bulk_update.deselect_all') : t('bulk_update.select_all')}
            </button>
          </div>
        </div>

        <div className="border border-border rounded-lg max-h-96 overflow-y-auto">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted sticky top-0">
              <tr>
                <th scope="col" className="px-4 py-3 text-left">
                  <input
                    type="checkbox"
                    checked={selectedUserIds.length === filteredUsers.length && filteredUsers.length > 0}
                    onChange={onSelectAll}
                    className="rounded border-input text-primary focus:ring-ring"
                  />
                </th>
                <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                  {t('users.col_user')}
                </th>
                <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                  {t('auth.email')}
                </th>
                <th scope="col" className="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase">
                  {t('common.status')}
                </th>
              </tr>
            </thead>
            <tbody className="bg-card divide-y divide-border">
              {filteredUsers.map((user) => (
                <tr key={user.id} className="hover:bg-accent">
                  <td className="px-4 py-3">
                    <input
                      type="checkbox"
                      checked={selectedUserIds.includes(user.id)}
                      onChange={() => onToggleUserSelection(user.id)}
                      className="rounded border-input text-primary focus:ring-ring"
                    />
                  </td>
                  <td className="px-4 py-3 text-sm font-medium text-foreground">{user.username}</td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{user.email}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${
                        user.is_active ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
                      }`}
                    >
                      {user.is_active ? t('common.active') : t('common.inactive')}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </>
  );
};

export default BulkUpdateFieldSelector;
