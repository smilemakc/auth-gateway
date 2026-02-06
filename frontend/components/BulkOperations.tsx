import React from 'react';
import { Link } from 'react-router-dom';
import { Users, UserPlus, UserMinus, UserCheck, FileSpreadsheet } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const BulkOperations: React.FC = () => {
  const { t } = useLanguage();

  const operations = [
    {
      id: 'create',
      title: t('bulk.create_users'),
      description: t('bulk.create_desc'),
      icon: UserPlus,
      path: '/bulk/create',
      color: 'bg-blue-500',
    },
    {
      id: 'update',
      title: t('bulk.update_users'),
      description: t('bulk.update_desc'),
      icon: Users,
      path: '/bulk/update',
      color: 'bg-green-500',
    },
    {
      id: 'delete',
      title: t('bulk.delete_users'),
      description: t('bulk.delete_desc'),
      icon: UserMinus,
      path: '/bulk/delete',
      color: 'bg-red-500',
    },
    {
      id: 'assign-roles',
      title: t('bulk.assign_roles'),
      description: t('bulk.assign_desc'),
      icon: UserCheck,
      path: '/bulk/assign-roles',
      color: 'bg-purple-500',
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">{t('bulk.title')}</h1>
        <p className="text-muted-foreground mt-1">{t('bulk.desc')}</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {operations.map((op) => {
          const Icon = op.icon;
          return (
            <Link
              key={op.id}
              to={op.path}
              className="bg-card rounded-xl shadow-sm border border-border p-6 hover:shadow-md transition-shadow"
            >
              <div className="flex items-start gap-4">
                <div className={`${op.color} p-3 rounded-lg`}>
                  <Icon className="text-white" size={24} />
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-semibold text-foreground mb-1">{op.title}</h3>
                  <p className="text-sm text-muted-foreground">{op.description}</p>
                </div>
              </div>
            </Link>
          );
        })}
      </div>

      <div className="bg-primary/10 border border-border rounded-lg p-4">
        <h3 className="font-semibold text-primary mb-2 flex items-center gap-2">
          <FileSpreadsheet size={20} />
          {t('bulk.csv_format')}
        </h3>
        <p className="text-sm text-primary">
          {t('bulk.csv_hint')}
        </p>
      </div>
    </div>
  );
};

export default BulkOperations;

