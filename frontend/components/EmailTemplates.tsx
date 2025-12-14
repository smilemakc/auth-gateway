
import React, { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getEmailTemplates } from '../services/mockData';
import { EmailTemplate } from '../types';
import { Mail, Edit2, Calendar, FileText, ArrowLeft } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const EmailTemplates: React.FC = () => {
  const [templates, setTemplates] = useState<EmailTemplate[]>([]);
  const navigate = useNavigate();
  const { t } = useLanguage();

  useEffect(() => {
    setTemplates(getEmailTemplates());
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => navigate('/settings')}
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
           <h1 className="text-2xl font-bold text-gray-900">{t('email.templates')}</h1>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {templates.map((template) => (
          <div key={template.id} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden flex flex-col hover:border-blue-300 transition-colors group">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="p-3 bg-blue-50 text-blue-600 rounded-lg group-hover:bg-blue-600 group-hover:text-white transition-colors">
                   <Mail size={24} />
                </div>
                <Link 
                   to={`/settings/email-templates/${template.id}`}
                   className="text-gray-400 hover:text-blue-600 transition-colors"
                >
                  <Edit2 size={20} />
                </Link>
              </div>
              
              <h3 className="text-lg font-bold text-gray-900 mb-1">{template.name}</h3>
              <p className="text-sm text-gray-500 mb-4 line-clamp-1">{t('email.subject')}: {template.subject}</p>
              
              <div className="flex flex-wrap gap-2 mt-4">
                {template.variables.slice(0, 3).map(v => (
                  <span key={v} className="bg-gray-100 text-gray-600 px-2 py-1 rounded text-xs font-mono border border-gray-200">
                    {v}
                  </span>
                ))}
                {template.variables.length > 3 && (
                  <span className="text-xs text-gray-400 self-center">+{template.variables.length - 3} more</span>
                )}
              </div>
            </div>
            
            <div className="bg-gray-50 px-6 py-4 border-t border-gray-100 flex items-center justify-between text-xs text-gray-500">
              <div className="flex items-center gap-2">
                <FileText size={14} />
                <span className="font-mono">HTML</span>
              </div>
              <div className="flex items-center gap-2">
                <Calendar size={14} />
                <span>Updated {new Date(template.updated_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default EmailTemplates;
