
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
    isActive: true
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
        setCreatedCredentials({ clientId: result.account.clientId, clientSecret: result.clientSecret });
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
        <div className="bg-green-50 border border-green-200 rounded-xl p-6 text-center">
          <div className="w-16 h-16 bg-green-100 text-green-600 rounded-full flex items-center justify-center mx-auto mb-4">
            <CheckCircle size={32} />
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('sa.generated')}</h2>
          <p className="text-gray-600">{t('sa.generated_desc')}</p>
        </div>

        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-6">
          <div>
            <label className="text-sm font-semibold text-gray-500 uppercase tracking-wider block mb-2">{t('oauth.client_id')}</label>
            <div className="flex gap-2">
              <input readOnly value={createdCredentials.clientId} className="flex-1 bg-gray-50 border border-gray-200 rounded-lg px-4 py-3 font-mono text-sm text-gray-800" />
              <button onClick={() => copyToClipboard(createdCredentials.clientId)} className="p-3 bg-gray-100 hover:bg-gray-200 rounded-lg text-gray-600"><Copy size={20}/></button>
            </div>
          </div>

          <div>
            <label className="text-sm font-semibold text-gray-500 uppercase tracking-wider block mb-2">{t('oauth.client_secret')}</label>
            <div className="flex gap-2">
              <input readOnly value={createdCredentials.clientSecret} className="flex-1 bg-gray-50 border border-gray-200 rounded-lg px-4 py-3 font-mono text-sm text-gray-800" />
              <button onClick={() => copyToClipboard(createdCredentials.clientSecret)} className="p-3 bg-gray-100 hover:bg-gray-200 rounded-lg text-gray-600"><Copy size={20}/></button>
            </div>
            <p className="text-xs text-red-500 mt-2 flex items-center">
              <AlertTriangle size={12} className="mr-1" />
              Store this secret securely.
            </p>
          </div>

          <div className="pt-4 flex justify-center">
            <button 
              onClick={() => navigate('/developers/service-accounts')}
              className="bg-blue-600 hover:bg-blue-700 text-white px-8 py-3 rounded-lg font-medium"
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
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{isNewMode ? t('sa.create') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 space-y-6">
          <div className="flex items-center gap-4 bg-blue-50 p-4 rounded-lg border border-blue-100 mb-6">
            <Bot size={24} className="text-blue-600 flex-shrink-0" />
            <p className="text-sm text-blue-800">
              Service accounts are used by backend systems to authenticate via Client Credentials flow.
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input 
              type="text" 
              required
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              placeholder="e.g. Payment Microservice"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea 
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              rows={3}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none resize-none"
            />
          </div>

          <div className="pt-2 border-t border-gray-100">
            <label className="flex items-center cursor-pointer">
              <input 
                type="checkbox" 
                className="sr-only" 
                checked={formData.isActive}
                onChange={(e) => setFormData(prev => ({ ...prev, isActive: e.target.checked }))}
              />
              <div className={`block w-10 h-6 rounded-full transition-colors ${formData.isActive ? 'bg-green-600' : 'bg-gray-300'}`}></div>
              <div className={`dot absolute bg-white w-4 h-4 rounded-full transition-transform transform ${formData.isActive ? 'translate-x-5' : 'translate-x-1'} mt-1 ml-0.5`}></div>
              <span className="ml-3 text-sm font-medium text-gray-700">{t('users.active')}</span>
            </label>
          </div>
        </div>

        <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/developers/service-accounts')}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={loading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-lg hover:bg-blue-700 focus:outline-none
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
