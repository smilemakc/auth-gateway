import React from 'react';
import { Plus, X } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface ApplicationEditCallbackUrlsSectionProps {
  callbackUrls: string[];
  onAdd: () => void;
  onRemove: (index: number) => void;
  onUpdate: (index: number, value: string) => void;
}

const ApplicationEditCallbackUrlsSection: React.FC<ApplicationEditCallbackUrlsSectionProps> = ({
  callbackUrls,
  onAdd,
  onRemove,
  onUpdate,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold text-foreground">{t('apps.callback_urls')}</h2>
        <button
          type="button"
          onClick={onAdd}
          className="flex items-center gap-1 text-sm text-primary hover:text-primary-600"
        >
          <Plus size={16} />
          {t('apps.add_url')}
        </button>
      </div>
      <p className="text-sm text-muted-foreground mb-4">
        {t('apps.callback_hint')}
      </p>
      <div className="space-y-3">
        {callbackUrls.map((url, index) => (
          <div key={index} className="flex items-center gap-2">
            <input
              type="url"
              value={url}
              onChange={e => onUpdate(index, e.target.value)}
              className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              placeholder="https://myapp.example.com/callback"
            />
            {callbackUrls.length > 1 && (
              <button
                type="button"
                onClick={() => onRemove(index)}
                className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
              >
                <X size={18} />
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ApplicationEditCallbackUrlsSection;
