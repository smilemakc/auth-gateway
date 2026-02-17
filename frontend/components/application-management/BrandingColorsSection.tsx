import React from 'react';
import { Palette } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface BrandingColorsSectionProps {
  primaryColor: string;
  secondaryColor: string;
  backgroundColor: string;
  onPrimaryColorChange: (value: string) => void;
  onSecondaryColorChange: (value: string) => void;
  onBackgroundColorChange: (value: string) => void;
}

const BrandingColorsSection: React.FC<BrandingColorsSectionProps> = ({
  primaryColor,
  secondaryColor,
  backgroundColor,
  onPrimaryColorChange,
  onSecondaryColorChange,
  onBackgroundColorChange,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
          <Palette className="text-primary" size={20} />
        </div>
        <h2 className="text-lg font-semibold text-foreground">{t('brand.colors')}</h2>
      </div>

      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('brand.primary')}
          </label>
          <div className="flex items-center gap-3">
            <input
              type="color"
              value={primaryColor}
              onChange={e => onPrimaryColorChange(e.target.value)}
              className="w-12 h-10 rounded border border-input cursor-pointer"
            />
            <input
              type="text"
              value={primaryColor}
              onChange={e => onPrimaryColorChange(e.target.value)}
              className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('brand.secondary')}
          </label>
          <div className="flex items-center gap-3">
            <input
              type="color"
              value={secondaryColor}
              onChange={e => onSecondaryColorChange(e.target.value)}
              className="w-12 h-10 rounded border border-input cursor-pointer"
            />
            <input
              type="text"
              value={secondaryColor}
              onChange={e => onSecondaryColorChange(e.target.value)}
              className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            {t('brand.bg')}
          </label>
          <div className="flex items-center gap-3">
            <input
              type="color"
              value={backgroundColor}
              onChange={e => onBackgroundColorChange(e.target.value)}
              className="w-12 h-10 rounded border border-input cursor-pointer"
            />
            <input
              type="text"
              value={backgroundColor}
              onChange={e => onBackgroundColorChange(e.target.value)}
              className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
          </div>
        </div>

        <div className="mt-4 p-4 rounded-lg border border-border" style={{ backgroundColor }}>
          <div className="text-center">
            <div
              className="inline-block px-4 py-2 rounded-lg text-white font-medium"
              style={{ backgroundColor: primaryColor }}
            >
              {t('brand.preview_btn')}
            </div>
            <p className="mt-2 text-sm" style={{ color: secondaryColor }}>
              {t('brand.preview_text')}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BrandingColorsSection;
