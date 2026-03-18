import React from 'react';
import { Boxes } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface ApplicationEditBasicInfoSectionProps {
  name: string;
  displayName: string;
  description: string;
  homepageUrl: string;
  isEditMode: boolean;
  errors: Record<string, string>;
  onNameChange: (value: string) => void;
  onDisplayNameChange: (value: string) => void;
  onDescriptionChange: (value: string) => void;
  onHomepageUrlChange: (value: string) => void;
}

const ApplicationEditBasicInfoSection: React.FC<ApplicationEditBasicInfoSectionProps> = ({
  name,
  displayName,
  description,
  homepageUrl,
  isEditMode,
  errors,
  onNameChange,
  onDisplayNameChange,
  onDescriptionChange,
  onHomepageUrlChange,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Boxes className="text-primary" size={20} />
        </div>
        <h2 className="text-lg font-semibold text-foreground">{t('apps.basic_info')}</h2>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('apps.name')} <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            value={name}
            onChange={e => onNameChange(e.target.value.toLowerCase())}
            disabled={isEditMode}
            className={`w-full px-3 py-2 bg-background border rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary ${
              errors.name ? 'border-destructive' : 'border-input'
            } ${isEditMode ? 'opacity-50 cursor-not-allowed' : ''}`}
            placeholder="my-application"
          />
          {errors.name && <p className="text-destructive text-sm mt-1">{errors.name}</p>}
          <p className="text-xs text-muted-foreground mt-1">
            {t('apps.name_hint')}
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('apps.display_name')} <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            value={displayName}
            onChange={e => onDisplayNameChange(e.target.value)}
            className={`w-full px-3 py-2 bg-background border rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary ${
              errors.display_name ? 'border-destructive' : 'border-input'
            }`}
            placeholder="My Application"
          />
          {errors.display_name && <p className="text-destructive text-sm mt-1">{errors.display_name}</p>}
        </div>

        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('apps.description')}
          </label>
          <textarea
            value={description}
            onChange={e => onDescriptionChange(e.target.value)}
            rows={3}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder={t('apps.description_placeholder')}
          />
        </div>

        <div className="md:col-span-2">
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('apps.homepage')}
          </label>
          <input
            type="url"
            value={homepageUrl}
            onChange={e => onHomepageUrlChange(e.target.value)}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="https://myapp.example.com"
          />
        </div>
      </div>
    </div>
  );
};

export default ApplicationEditBasicInfoSection;
