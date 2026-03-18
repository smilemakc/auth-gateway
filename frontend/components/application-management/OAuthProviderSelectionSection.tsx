import React from 'react';
import { useLanguage } from '../../services/i18n';

interface OAuthProviderSelectionSectionProps {
  selectedProvider: string;
  isEditMode: boolean;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const PROVIDERS = ['google', 'github', 'yandex', 'telegram', 'instagram'];

const OAuthProviderSelectionSection: React.FC<OAuthProviderSelectionSectionProps> = ({
  selectedProvider,
  isEditMode,
  onChange,
}) => {
  const { t } = useLanguage();

  return (
    <div>
      <label className="block text-sm font-medium text-muted-foreground mb-2">{t('app_oauth.provider')}</label>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
        {PROVIDERS.map(p => (
          <label key={p} className={`
            cursor-pointer border rounded-lg p-3 text-center transition-all hover:bg-accent
            ${selectedProvider === p ? 'ring-2 ring-ring border-transparent bg-primary/10' : 'border-border'}
          `}>
            <input
              type="radio"
              name="provider"
              value={p}
              checked={selectedProvider === p}
              onChange={onChange}
              className="sr-only"
              disabled={isEditMode}
            />
            <span className="capitalize font-medium block text-foreground">{p}</span>
          </label>
        ))}
      </div>
    </div>
  );
};

export default OAuthProviderSelectionSection;
