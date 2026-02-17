import React from 'react';
import { Globe } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatDate } from '../../lib/date';

interface OAuthAccount {
  id: string;
  provider: string;
  created_at: string;
}

interface UserOAuthSectionProps {
  oauthAccounts: OAuthAccount[];
}

const UserOAuthSection: React.FC<UserOAuthSectionProps> = ({ oauthAccounts }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border">
        <h3 className="font-semibold text-foreground flex items-center gap-2">
          <Globe size={18} className="text-primary" />
          {t('user.linked_accounts')}
        </h3>
      </div>
      <div className="p-6">
        {oauthAccounts.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {oauthAccounts.map((acc: any) => (
              <div key={acc.id} className="flex items-center justify-between p-3 border border-border rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded bg-muted flex items-center justify-center font-bold text-muted-foreground uppercase">
                    {acc.provider?.[0] || '?'}
                  </div>
                  <div>
                    <p className="text-sm font-medium capitalize">{acc.provider}</p>
                    <p className="text-xs text-muted-foreground">{t('user.connected')} {formatDate(acc.created_at)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground italic">{t('user.no_linked_accounts')}</p>
        )}
      </div>
    </div>
  );
};

export default UserOAuthSection;
