import React from 'react';
import { ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

type ProviderType = 'smtp' | 'sendgrid' | 'ses' | 'mailgun';

interface EmailProviderTestSectionProps {
  name: string;
  type: ProviderType;
  isActive: boolean;
  isNew: boolean;
  onNameChange: (value: string) => void;
  onTypeChange: (value: ProviderType) => void;
  onActiveToggle: () => void;
}

const EmailProviderTestSection: React.FC<EmailProviderTestSectionProps> = ({
  name,
  type,
  isActive,
  isNew,
  onNameChange,
  onTypeChange,
  onActiveToggle,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <h2 className="text-lg font-semibold text-foreground mb-4">
        {t('email.basic_settings')}
      </h2>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('email.provider_name')} *
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            required
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="My SMTP Server"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('email.provider_type')} *
          </label>
          <select
            value={type}
            onChange={(e) => onTypeChange(e.target.value as ProviderType)}
            disabled={!isNew}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50"
          >
            <option value="smtp">SMTP</option>
            <option value="sendgrid">SendGrid</option>
            <option value="ses">AWS SES</option>
            <option value="mailgun">Mailgun</option>
          </select>
        </div>

        <div className="flex items-center gap-3">
          <button type="button" onClick={onActiveToggle}
            className={`transition-colors ${isActive ? 'text-success' : 'text-muted-foreground'}`}>
            {isActive ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
          </button>
          <span className="text-sm font-medium text-foreground">
            {t('email.provider_active')}
          </span>
        </div>
      </div>
    </div>
  );
};

export default EmailProviderTestSection;
