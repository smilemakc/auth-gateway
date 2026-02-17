import React, { useState } from 'react';
import { Eye, EyeOff, Copy, Check } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface OAuthClientSecretSectionProps {
  clientSecret: string;
  onDone: () => void;
}

const OAuthClientSecretSection: React.FC<OAuthClientSecretSectionProps> = ({
  clientSecret,
  onDone,
}) => {
  const { t } = useLanguage();
  const [showSecret, setShowSecret] = useState(false);
  const [copiedSecret, setCopiedSecret] = useState(false);

  const copySecret = () => {
    navigator.clipboard.writeText(clientSecret);
    setCopiedSecret(true);
    setTimeout(() => setCopiedSecret(false), 2000);
  };

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="bg-success/10 border border-success rounded-xl p-6">
        <h2 className="text-xl font-bold text-success mb-2">{t('oauth_edit.success_title')}</h2>
        <p className="text-success mb-4">
          {t('oauth_edit.success_desc')}
        </p>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-success mb-1">{t('oauth_edit.client_secret')}</label>
            <div className="flex items-center gap-2">
              <code className="flex-1 bg-card rounded-lg px-4 py-3 text-sm font-mono border border-success break-all">
                {showSecret ? clientSecret : '\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022'}
              </code>
              <button
                onClick={() => setShowSecret(!showSecret)}
                className="p-2 text-success hover:bg-success/10 rounded-lg"
              >
                {showSecret ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
              <button
                onClick={copySecret}
                className="p-2 text-success hover:bg-success/10 rounded-lg"
              >
                {copiedSecret ? <Check size={20} /> : <Copy size={20} />}
              </button>
            </div>
          </div>
        </div>
        <div className="mt-6 flex justify-end">
          <button
            onClick={onDone}
            className="px-4 py-2 bg-success hover:bg-success text-primary-foreground rounded-lg font-medium"
          >
            {t('oauth_edit.done')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default OAuthClientSecretSection;
