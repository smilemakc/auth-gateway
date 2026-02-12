
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Mail, Server } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import EmailTemplates from './EmailTemplates';
import EmailProviders from './EmailProviders';

type Tab = 'templates' | 'providers';

const EmailSettings: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [activeTab, setActiveTab] = useState<Tab>('templates');

  const tabs = [
    { id: 'templates' as Tab, label: t('email.tab_templates'), icon: Mail },
    { id: 'providers' as Tab, label: t('email.tab_providers'), icon: Server },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/settings')}
          className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">
            {t('email.settings_title')}
          </h1>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-border">
        <nav className="flex gap-6">
          {tabs.map(tab => {
            const Icon = tab.icon;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 px-1 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab.id
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                }`}
              >
                <Icon size={18} />
                {tab.label}
              </button>
            );
          })}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'templates' && <EmailTemplates embedded />}
      {activeTab === 'providers' && <EmailProviders embedded />}
    </div>
  );
};

export default EmailSettings;
