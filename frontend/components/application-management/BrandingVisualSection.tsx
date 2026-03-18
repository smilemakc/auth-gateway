import React from 'react';
import { Image } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface BrandingVisualSectionProps {
  logoUrl: string;
  faviconUrl: string;
  onLogoUrlChange: (value: string) => void;
  onFaviconUrlChange: (value: string) => void;
}

const BrandingVisualSection: React.FC<BrandingVisualSectionProps> = ({
  logoUrl,
  faviconUrl,
  onLogoUrlChange,
  onFaviconUrlChange,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Image className="text-primary" size={20} />
        </div>
        <h2 className="text-lg font-semibold text-foreground">{t('brand.visual')}</h2>
      </div>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('brand.logo')}
          </label>
          <input
            type="url"
            value={logoUrl}
            onChange={e => onLogoUrlChange(e.target.value)}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="https://example.com/logo.png"
          />
          {logoUrl && (
            <div className="mt-2 p-4 bg-muted rounded-lg flex items-center justify-center">
              <img src={logoUrl} alt="Logo preview" className="max-h-16 object-contain" />
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('brand.favicon')}
          </label>
          <input
            type="url"
            value={faviconUrl}
            onChange={e => onFaviconUrlChange(e.target.value)}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
            placeholder="https://example.com/favicon.ico"
          />
        </div>
      </div>
    </div>
  );
};

export default BrandingVisualSection;
