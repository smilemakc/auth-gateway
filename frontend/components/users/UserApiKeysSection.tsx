import React from 'react';
import { Key } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatDate } from '../../lib/date';

interface ApiKey {
  id: string;
  name: string;
  key_prefix: string;
  is_active: boolean;
  created_at: string;
}

interface UserApiKeysSectionProps {
  apiKeys: ApiKey[];
}

const UserApiKeysSection: React.FC<UserApiKeysSectionProps> = ({ apiKeys }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border flex justify-between items-center">
        <h3 className="font-semibold text-foreground flex items-center gap-2">
          <Key size={18} className="text-primary" />
          {t('dash.api_keys')}
        </h3>
        <span className="text-xs font-medium bg-muted text-muted-foreground px-2 py-1 rounded-full">{apiKeys.length}</span>
      </div>
      <div className="overflow-x-auto">
        {apiKeys.length > 0 ? (
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">{t('common.name')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">{t('keys.prefix')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">{t('common.status')}</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase">{t('common.created')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {apiKeys.map(key => (
                <tr key={key.id}>
                  <td className="px-6 py-4 text-sm font-medium text-foreground">{key.name}</td>
                  <td className="px-6 py-4 text-sm font-mono text-muted-foreground">{key.key_prefix}...</td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium capitalize
                      ${key.is_active ? 'bg-success/20 text-success' : 'bg-destructive/20 text-destructive'}`}>
                      {key.is_active ? t('common.active') : t('common.inactive')}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-muted-foreground">{formatDate(key.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <div className="p-6 text-sm text-muted-foreground italic">{t('user.no_keys')}</div>
        )}
      </div>
    </div>
  );
};

export default UserApiKeysSection;
