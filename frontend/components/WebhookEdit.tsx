
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Loader2 } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useWebhookDetail, useCreateWebhook, useUpdateWebhook } from '../hooks/useWebhooks';

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

  const [formData, setFormData] = useState({
    url: '',
    description: '',
    events: [] as string[],
    is_active: true
  });

  const { data: existingWebhook, isLoading: loadingWebhook } = useWebhookDetail(isEditMode ? id! : '');
  const createWebhookMutation = useCreateWebhook();
  const updateWebhookMutation = useUpdateWebhook();

  useEffect(() => {
    if (isEditMode && existingWebhook) {
      setFormData({
        url: existingWebhook.url || '',
        description: existingWebhook.description || '',
        events: existingWebhook.events || [],
        is_active: existingWebhook.is_active ?? true
      });
    }
  }, [existingWebhook, isEditMode]);

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      if (isNewMode) {
        await createWebhookMutation.mutateAsync({
          url: formData.url,
          description: formData.description,
          events: formData.events,
          is_active: formData.is_active
        });
      } else if (id) {
        await updateWebhookMutation.mutateAsync({
          id,
          data: {
            url: formData.url,
            description: formData.description,
            events: formData.events,
            is_active: formData.is_active
          }
        });
      }
      navigate('/developers/webhooks');
    } catch (err) {
      console.error('Failed to save webhook:', err);
    }
  };

  const isLoading = createWebhookMutation.isPending || updateWebhookMutation.isPending;

  if (isEditMode && loadingWebhook) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <button
          onClick={() => navigate('/developers/webhooks')}
          className="p-2 hover:bg-card rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-foreground">{isNewMode ? t('hooks.add') : t('common.edit')}</h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 space-y-6">

          <div className="grid grid-cols-1 gap-6">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">{t('hooks.url')}</label>
              <input
                type="url"
                required
                placeholder="https://api.example.com/webhooks"
                value={formData.url}
                onChange={(e) => setFormData(prev => ({ ...prev, url: e.target.value }))}
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1">Description</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="e.g. Sync users to marketing CRM"
                className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring outline-none"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-3">{t('hooks.events')}</label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {AVAILABLE_EVENTS.map(event => {
                  const isSelected = formData.events?.includes(event);
                  return (
                    <div
                      key={event}
                      onClick={() => toggleEvent(event)}
                      className={`
                        cursor-pointer px-4 py-3 rounded-lg border flex items-center justify-between transition-all
                        ${isSelected ? 'bg-primary/10 border-primary text-primary' : 'bg-card border-border hover:border-input text-foreground'}
                      `}
                    >
                      <span className="text-sm font-mono">{event}</span>
                      <div className={`w-4 h-4 rounded-full border flex items-center justify-center ${isSelected ? 'bg-primary border-primary' : 'border-input'}`}>
                         {isSelected && <div className="w-1.5 h-1.5 rounded-full bg-primary-foreground"></div>}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="pt-4 border-t border-border">
              <label className="flex items-center cursor-pointer">
                <div className="relative">
                  <input
                    type="checkbox"
                    className="sr-only"
                    checked={formData.is_active}
                    onChange={(e) => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                  />
                  <div className={`block w-10 h-6 rounded-full transition-colors ${formData.is_active ? 'bg-primary' : 'bg-muted'}`}></div>
                  <div className={`dot absolute left-1 top-1 bg-primary-foreground w-4 h-4 rounded-full transition-transform ${formData.is_active ? 'transform translate-x-4' : ''}`}></div>
                </div>
                <div className="ml-3 text-sm font-medium text-foreground">
                  {t('oauth.enable')}
                </div>
              </label>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 bg-muted border-t border-border flex justify-end gap-3">
           <button
            type="button"
            onClick={() => navigate('/developers/webhooks')}
            className="px-4 py-2 text-sm font-medium text-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none"
          >
            {t('common.cancel')}
          </button>
          <button
            type="submit"
            disabled={isLoading}
            className={`flex items-center px-6 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none
              ${isLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
          >
            {isLoading ? (
              <Loader2 size={16} className="mr-2 animate-spin" />
            ) : (
              <Save size={16} className="mr-2" />
            )}
            {isNewMode ? t('common.create') : t('common.save')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default WebhookEdit;
