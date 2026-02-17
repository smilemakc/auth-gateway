import React from 'react';
import { X, CheckCircle, Copy } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

const AVAILABLE_SCOPES = [
  'users:read',
  'users:write',
  'users:sync',
  'users:import',
  'profile:read',
  'profile:write',
  'token:validate',
  'token:introspect',
  'auth:login',
  'auth:register',
  'auth:otp',
  'email:send',
  'oauth:read',
  'exchange:manage',
  'admin:all',
  'all',
];

interface CreateApiKeyModalProps {
  newKeyName: string;
  onNameChange: (value: string) => void;
  newKeyScopes: string[];
  onToggleScope: (scope: string) => void;
  newKeyExpiresIn: number | undefined;
  onExpiresChange: (value: number | undefined) => void;
  generatedKey: string | null;
  copiedId: string | null;
  onCopy: (text: string, id: string) => void;
  onClose: () => void;
  onCreate: () => void;
  isCreating: boolean;
}

export const CreateApiKeyModal: React.FC<CreateApiKeyModalProps> = ({
  newKeyName,
  onNameChange,
  newKeyScopes,
  onToggleScope,
  newKeyExpiresIn,
  onExpiresChange,
  generatedKey,
  copiedId,
  onCopy,
  onClose,
  onCreate,
  isCreating,
}) => {
  const { t } = useLanguage();

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div className="bg-card rounded-xl shadow-xl border border-border w-full max-w-md">
        <div className="flex items-center justify-between p-6 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">
            {generatedKey ? t('keys.key_generated') : t('keys.generate')}
          </h2>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <div className="p-6 space-y-4">
          {generatedKey ? (
            <>
              <div className="bg-warning/10 border border-warning/20 rounded-lg p-4">
                <p className="text-sm text-warning font-medium mb-2">
                  {t('keys.copy_warning')}
                </p>
                <div className="flex items-center gap-2 bg-card rounded border border-border p-2">
                  <code className="flex-1 text-sm font-mono text-foreground break-all">
                    {generatedKey}
                  </code>
                  <button
                    onClick={() => onCopy(generatedKey, 'new-key')}
                    className="p-2 text-muted-foreground hover:text-primary transition-colors"
                  >
                    {copiedId === 'new-key' ? <CheckCircle size={18} className="text-success" /> : <Copy size={18} />}
                  </button>
                </div>
              </div>
              <button
                onClick={onClose}
                className="w-full bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors"
              >
                {t('common.done')}
              </button>
            </>
          ) : (
            <>
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('keys.name')}
                </label>
                <input
                  type="text"
                  value={newKeyName}
                  onChange={(e) => onNameChange(e.target.value)}
                  placeholder={t('keys.name_placeholder')}
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('keys.scopes')}
                </label>
                <div className="flex flex-wrap gap-2">
                  {AVAILABLE_SCOPES.map(scope => (
                    <button
                      key={scope}
                      onClick={() => onToggleScope(scope)}
                      className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-colors ${
                        newKeyScopes.includes(scope)
                          ? 'bg-primary text-primary-foreground border-primary'
                          : 'bg-background text-muted-foreground border-input hover:border-primary'
                      }`}
                    >
                      {scope}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('keys.expires_in')}
                </label>
                <select
                  value={newKeyExpiresIn || ''}
                  onChange={(e) => onExpiresChange(e.target.value ? parseInt(e.target.value) : undefined)}
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">{t('keys.never')}</option>
                  <option value="30">30 {t('common.days')}</option>
                  <option value="90">90 {t('common.days')}</option>
                  <option value="180">180 {t('common.days')}</option>
                  <option value="365">365 {t('common.days')}</option>
                </select>
              </div>

              <div className="flex gap-3 pt-2">
                <button
                  onClick={onClose}
                  className="flex-1 px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={onCreate}
                  disabled={isCreating}
                  className="flex-1 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg font-medium transition-colors disabled:opacity-50"
                >
                  {isCreating ? t('common.creating') : t('keys.generate')}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};
