
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getServiceAccount, createServiceAccount, updateServiceAccount } from '../services/mockData';
import { ServiceAccount } from '../types';
import { ArrowLeft, Save, Bot, Copy, AlertTriangle, CheckCircle } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const ServiceAccountEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = id && id !== 'new';
  const isNewMode = id === 'new';

  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<Partial<ServiceAccount>>({
    name: '',
    description: '',
    is_active: true
  });
  
  // State for newly created credentials
  const [createdCredentials, setCreatedCredentials] = useState<{clientId: string, clientSecret: string} | null>(null);

  useEffect(() => {
    if (isEditMode) {
      const existing = getServiceAccount(id);
      if (existing) {
        setFormData(existing);
      } else {
        navigate('/developers/service-accounts');
      }
    }
  }, [id, isEditMode, navigate]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    
    setTimeout(() => {
      if (isNewMode) {
        const result = createServiceAccount(formData);
        setCreatedCredentials({ clientId: result.account.client_id, clientSecret: result.clientSecret });
        setLoading(false);
        // Don't navigate away yet, show credentials
      } else if (id) {
        updateServiceAccount(id, formData);
        setLoading(false);
        navigate('/developers/service-accounts');
      }
    }, 800);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (createdCredentials) {
    return (
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="bg-success/10 border border-success/20 rounded-xl p-6 text-center">
          <div className="w-16 h-16 bg-success/20 text-success rounded-full flex items-center justify-center mx-auto mb-4">
            <CheckCircle size={32} />
          </div>
          <h2 className="text-2xl font-bold text-foreground mb-2">{t('sa.generated')}</h2>
          <p className="text-muted-foreground">{t('sa.generated_desc')}</p>
        </div>

        <div className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
          <div>
            <label className="text-sm font-semibold text-muted-foreground uppercase tracking-wider block mb-2">{t('oauth.client_id')}</label>
            <div className="flex gap-2">
              <input readOnly value={createdCredentials.clientId} className="flex-1 bg-muted border border-border rounded-lg px-4 py-3 font-mono text-sm text-foreground" />
              <button onClick={() => copyToClipboard(createdCredentials.clientId)} className="p-3 bg-muted hover:bg-accent rounded-lg text-muted-foreground"><Copy size={20}/></button>
            </div>
          </div>

          <div>
            <label className="text-sm font-semibold text-muted-foreground uppercase tracking-wider block mb-2">{t('oauth.client_secret')}</label>
            <div className="flex gap-2">
              <input readOnly value={createdCredentials.clientSecret} className="flex-1 bg-muted border border-border rounded-lg px-4 py-3 font-mono text-sm text-foreground" />
              <button onClick={() => copyToClipboard(createdCredentials.clientSecret)} className="p-3 bg-muted hover:bg-accent rounded-lg text-muted-foreground"><Copy size={20}/></button>
            </div>
            <p className="text-xs text-destructive mt-2 flex items-center">
              <AlertTriangle size={12} className="mr-1" />
              Store this secret securely.
            </p>
          </div>

          <div className="pt-4 flex justify-center">
            <button
              onClick={() => navigate('/developers/service-accounts')}
              className="bg-primary hover:bg-primary-600 text-primary-foreground px-8 py-3 rounded-lg font-medium"
            >
              {t('common.back')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/developers/service-accounts')}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{isNewMode ? t('sa.create') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 space-y-6">
          <div className="flex items-center gap-4 bg-primary/10 p-4 rounded-lg border border-primary/20 mb-6">
            <Bot size={24} className="text-primary flex-shrink-0" />
            <p className="text-sm text-primary">
              Service accounts are used by backend systems to authenticate via Client Credentials flow.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Name</label>
            <input
              type="text"
              required
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              placeholder="e.g. Payment Microservice"
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1">Description</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              rows={3}
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none resize-none"
            />
          </div>

          <div className="pt-2 border-t border-border">
            <label className="flex items-center cursor-pointer">
              <input
                type="checkbox"
                className="sr-only"
                checked={formData.is_active}
                onChange={(e) => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
              />
              <div className={`block w-10 h-6 rounded-full transition-colors ${formData.is_active ? 'bg-success' : 'bg-muted'}`}></div>
              <div className={`dot absolute bg-card w-4 h-4 rounded-full transition-transform transform ${formData.is_active ? 'translate-x-5' : 'translate-x-1'} mt-1 ml-0.5`}></div>
              <span className="ml-3 text-sm font-medium text-foreground">{t('users.active')}</span>
            </label>
          </div>
        </div>

        <div className="px-6 py-4 bg-muted border-t border-border flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/developers/service-accounts')}
            className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none
              ${loading ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {loading ? t('common.saving') : (
              <>
                <Save size={16} className="mr-2" />
                {isNewMode ? t('keys.generate') : t('common.save')}
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ServiceAccountEdit;
