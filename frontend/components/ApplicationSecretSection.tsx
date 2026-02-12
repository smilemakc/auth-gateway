import React, { useState } from 'react';
import { Copy, Check, RefreshCw, AlertTriangle, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useRotateApplicationSecret } from '../hooks/useApplications';
import { confirm } from '../services/confirm';
import { toast } from '../services/toast';
import { formatDateTime } from '../lib/date';
import type { Application } from '../types';

interface ApplicationSecretSectionProps {
  application: Application;
  initialSecret?: string;
}

const ApplicationSecretSection: React.FC<ApplicationSecretSectionProps> = ({ application, initialSecret }) => {
  const { t } = useLanguage();
  const rotateSecret = useRotateApplicationSecret();
  const [copiedSecret, setCopiedSecret] = useState(false);
  const [revealedSecret, setRevealedSecret] = useState<string | undefined>(initialSecret);

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSecret(true);
    setTimeout(() => setCopiedSecret(false), 2000);
  };

  const handleRotate = async () => {
    const ok = await confirm({
      title: t('apps.secret.rotate_confirm_title'),
      description: t('apps.secret.rotate_confirm_message'),
      confirmText: t('apps.secret.rotate_confirm_button'),
      variant: 'danger',
    });

    if (!ok) return;

    try {
      const result = await rotateSecret.mutateAsync(application.id);
      setRevealedSecret(result.secret);
      toast.success(t('apps.secret.rotated'));
    } catch (error) {
      toast.error(String(error instanceof Error ? error.message : error));
    }
  };

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-foreground">{t('apps.secret.title')}</h2>
        <button
          type="button"
          onClick={handleRotate}
          disabled={rotateSecret.isPending}
          className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-destructive hover:bg-destructive/10 rounded-lg transition-colors disabled:opacity-50"
        >
          {rotateSecret.isPending ? (
            <Loader2 size={14} className="animate-spin" />
          ) : (
            <RefreshCw size={14} />
          )}
          {t('apps.secret.rotate')}
        </button>
      </div>

      {revealedSecret ? (
        <div className="space-y-3">
          <div className="flex items-start gap-3 p-3 bg-warning/10 border border-warning/30 rounded-lg">
            <AlertTriangle size={18} className="text-warning shrink-0 mt-0.5" />
            <p className="text-sm text-warning">{t('apps.secret.warning_once')}</p>
          </div>
          <div className="flex items-center gap-2">
            <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-foreground font-mono border border-border break-all">
              {revealedSecret}
            </code>
            <button
              onClick={() => copyToClipboard(revealedSecret)}
              className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded shrink-0"
            >
              {copiedSecret ? <Check size={16} /> : <Copy size={16} />}
            </button>
          </div>
        </div>
      ) : (
        <dl className="space-y-3">
          {application.secret_prefix && (
            <div>
              <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                {t('apps.secret.prefix')}
              </dt>
              <dd>
                <code className="bg-muted rounded px-3 py-2 text-sm text-foreground font-mono border border-border inline-block">
                  {application.secret_prefix} {'••••••••'}
                </code>
              </dd>
            </div>
          )}
          {application.secret_last_rotated_at && (
            <div>
              <dt className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-1">
                {t('apps.secret.last_rotated')}
              </dt>
              <dd className="text-sm text-foreground">
                {formatDateTime(application.secret_last_rotated_at)}
              </dd>
            </div>
          )}
        </dl>
      )}
    </div>
  );
};

export default ApplicationSecretSection;
