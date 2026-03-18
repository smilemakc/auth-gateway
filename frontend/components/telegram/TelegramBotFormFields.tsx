import React from 'react';
import { Eye, EyeOff, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface TelegramBotFormData {
  bot_token: string;
  bot_username: string;
  display_name: string;
  is_auth_bot: boolean;
  is_active: boolean;
}

interface TelegramBotFormFieldsProps {
  formData: TelegramBotFormData;
  isEditMode: boolean;
  isNewMode: boolean;
  showToken: boolean;
  onToggleShowToken: () => void;
  onChange: (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => void;
  onToggleField: (field: 'is_auth_bot' | 'is_active') => void;
}

export const TelegramBotFormFields: React.FC<TelegramBotFormFieldsProps> = ({
  formData,
  isEditMode,
  isNewMode,
  showToken,
  onToggleShowToken,
  onChange,
  onToggleField,
}) => {
  const { t } = useLanguage();

  return (
    <div className="p-6 space-y-8">
      {/* Bot Token */}
      <div>
        <label htmlFor="bot_token" className="block text-sm font-medium text-muted-foreground mb-1">
          {t('tg.bot_token')}
        </label>
        <div className="relative">
          <input
            type={showToken ? "text" : "password"}
            id="bot_token"
            name="bot_token"
            value={formData.bot_token}
            onChange={onChange}
            required={isNewMode}
            className="w-full pl-4 pr-12 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
            placeholder={isEditMode ? t('tg.token_keep_current') : 'e.g. 123456789:ABCdefGHIjklMNOpqrsTUVwxyz'}
          />
          <button
            type="button"
            onClick={onToggleShowToken}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          >
            {showToken ? <EyeOff size={18} /> : <Eye size={18} />}
          </button>
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {t('tg.token_hint')}
        </p>
      </div>

      {/* Bot Username */}
      <div>
        <label htmlFor="bot_username" className="block text-sm font-medium text-muted-foreground mb-1">
          {t('tg.bot_username')}
        </label>
        <div className="relative">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground">@</span>
          <input
            type="text"
            id="bot_username"
            name="bot_username"
            value={formData.bot_username}
            onChange={onChange}
            required
            className="w-full pl-8 pr-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
            placeholder="your_bot_username"
          />
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          {t('tg.username_hint')}
        </p>
      </div>

      {/* Display Name */}
      <div>
        <label htmlFor="display_name" className="block text-sm font-medium text-muted-foreground mb-1">
          {t('tg.display_name')}
        </label>
        <input
          type="text"
          id="display_name"
          name="display_name"
          value={formData.display_name}
          onChange={onChange}
          required
          className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all text-sm"
          placeholder="e.g. My App Auth Bot"
        />
        <p className="text-xs text-muted-foreground mt-1">
          {t('tg.display_name_hint')}
        </p>
      </div>

      {/* Bot Settings */}
      <div className="pt-6 border-t border-border space-y-4">
        <div className="flex items-start gap-3">
          <button
            type="button"
            onClick={() => onToggleField('is_auth_bot')}
            className={`transition-colors mt-0.5 ${formData.is_auth_bot ? 'text-success' : 'text-muted-foreground'}`}
          >
            {formData.is_auth_bot ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
          </button>
          <div>
            <span className="font-medium text-foreground block">
              {t('tg.auth_bot')}
            </span>
            <p className="text-sm text-muted-foreground mt-1">
              {t('tg.auth_bot_desc')}
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3">
          <button
            type="button"
            onClick={() => onToggleField('is_active')}
            className={`transition-colors mt-0.5 ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}
          >
            {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
          </button>
          <div>
            <span className="font-medium text-foreground block">
              {t('common.active')}
            </span>
            <p className="text-sm text-muted-foreground mt-1">
              {t('tg.inactive_hint')}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
