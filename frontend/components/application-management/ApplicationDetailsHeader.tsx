import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Edit2, Boxes } from 'lucide-react';
import { useLanguage } from '../../services/i18n';

interface ApplicationDetailsHeaderProps {
  applicationId: string;
  displayName: string;
  isActive: boolean;
  isSystem: boolean;
  logoUrl?: string;
}

const ApplicationDetailsHeader: React.FC<ApplicationDetailsHeaderProps> = ({
  applicationId,
  displayName,
  isActive,
  isSystem,
  logoUrl,
}) => {
  const { t } = useLanguage();

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-4">
        <Link
          to="/applications"
          className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </Link>
        <div className="flex items-center gap-4">
          <div className="w-14 h-14 rounded-xl bg-muted flex items-center justify-center shadow-sm">
            {logoUrl ? (
              <img src={logoUrl} alt={displayName} className="w-10 h-10 object-contain" />
            ) : (
              <Boxes className="text-primary" size={28} />
            )}
          </div>
          <div>
            <h1 className="text-2xl font-bold text-foreground">{displayName}</h1>
            <div className="flex items-center gap-2 mt-1">
              <span className={`w-2 h-2 rounded-full ${isActive ? 'bg-success' : 'bg-muted-foreground'}`}></span>
              <span className="text-sm text-muted-foreground">
                {isActive ? t('common.active') : t('common.inactive')}
              </span>
              {isSystem && (
                <>
                  <span className="text-muted-foreground">•</span>
                  <span className="text-sm text-warning">{t('apps.system')}</span>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
      <Link
        to={`/applications/${applicationId}/edit`}
        className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
      >
        <Edit2 size={18} />
        {t('common.edit')}
      </Link>
    </div>
  );
};

export default ApplicationDetailsHeader;
