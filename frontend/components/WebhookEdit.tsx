
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getWebhook, createWebhook, updateWebhook } from '../services/mockData';
import { WebhookEndpoint } from '../types';
import { ArrowLeft, Save, Activity } from 'lucide-react';
import { useLanguage } from '../services/i18n';

const AVAILABLE_EVENTS = [
  'user.created',
  'user.updated',
  'user.deleted',
  'user.blocked',
  'auth.login.success',
  'auth.login.failed',
  'org.created',
  'org.updated'
];

const WebhookEdit: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = id && id !== 'new';
  const isNewMode = id === 'new';

  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState<Partial<WebhookEndpoint>>({
    url: '',
    description: '',
    events: [],
    is_active: true
  });

  useEffect(() => {
    if (isEditMode) {
      const existing = getWebhook(id);
      if (existing) {
        setFormData(existing);
      } else {
        navigate('/developers/webhooks');
      }
    }
  }, [id, isEditMode, navigate]);

  const toggleEvent = (event: string) => {
    setFormData(prev => {
      const current = prev.events || [];
      if (current.includes(event)) {
        return { ...prev, events: current.filter(e => e !== event) };
      } else {
        return { ...prev, events: [...current, event] };
      }
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setTimeout(() => {
      if (isNewMode) {
        createWebhook(formData);
      } else if (id) {
        updateWebhook(id, formData);
      }
      setLoading(false);
      navigate('/developers/webhooks');
    }, 800);
  };

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button 
          onClick={() => navigate('/developers/webhooks')}
          className="p-2 hover:bg-white rounded-lg transition-colors text-gray-500"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">{isNewMode ? t('hooks.add') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 space-y-6">
          
          <div className="grid grid-cols-1 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t('hooks.url')}</label>
              <input 
                type="url" 
                required
                placeholder="https://api.example.com/webhooks"
                value={formData.url}
                onChange={(e) => setFormData(prev => ({ ...prev, url: e.target.value }))}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input 
                type="text" 
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="e.g. Sync users to marketing CRM"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-3">{t('hooks.events')}</label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {AVAILABLE_EVENTS.map(event => {
                  const isSelected = formData.events?.includes(event);
                  return (
                    <div 
                      key={event}
                      onClick={() => toggleEvent(event)}
                      className={`
                        cursor-pointer px-4 py-3 rounded-lg border flex items-center justify-between transition-all
                        ${isSelected ? 'bg-blue-50 border-blue-200 text-blue-700' : 'bg-white border-gray-200 hover:border-gray-300 text-gray-700'}
                      `}
                    >
                      <span className="text-sm font-mono">{event}</span>
                      <div className={`w-4 h-4 rounded-full border flex items-center justify-center ${isSelected ? 'bg-blue-600 border-blue-600' : 'border-gray-300'}`}>
                         {isSelected && <div className="w-1.5 h-1.5 rounded-full bg-white"></div>}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="pt-4 border-t border-gray-100">
              <label className="flex items-center cursor-pointer">
                <div className="relative">
                  <input
                    type="checkbox"
                    className="sr-only"
                    checked={formData.is_active}
                    onChange={(e) => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                  />
                  <div className={`block w-10 h-6 rounded-full transition-colors ${formData.is_active ? 'bg-blue-600' : 'bg-gray-300'}`}></div>
                  <div className={`dot absolute left-1 top-1 bg-white w-4 h-4 rounded-full transition-transform ${formData.is_active ? 'transform translate-x-4' : ''}`}></div>
                </div>
                <div className="ml-3 text-sm font-medium text-gray-700">
                  {t('oauth.enable')}
                </div>
              </label>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/developers/webhooks')}
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
                {isNewMode ? t('common.create') : t('common.save')}
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default WebhookEdit;
