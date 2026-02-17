
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, HelpCircle, Loader2, Send } from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { useTelegramBotDetail, useCreateTelegramBot, useUpdateTelegramBot, useDeleteTelegramBot } from '../../hooks/useTelegramBots';
import { confirm } from '../../services/confirm';
import { logger } from '@/lib/logger';
import { TelegramBotFormFields } from './TelegramBotFormFields';

const TelegramBotEdit: React.FC = () => {
  const { applicationId, botId } = useParams<{ applicationId: string; botId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const isEditMode = botId && botId !== 'new';
  const isNewMode = !botId || botId === 'new';

  const [showToken, setShowToken] = useState(false);
  const [formData, setFormData] = useState({
    bot_token: '',
    bot_username: '',
    display_name: '',
    is_auth_bot: false,
    is_active: true
  });

  const { data: existingBot, isLoading: loadingBot } = useTelegramBotDetail(
    applicationId || '',
    isEditMode ? botId! : ''
  );
  const createMutation = useCreateTelegramBot();
  const updateMutation = useUpdateTelegramBot();
  const deleteMutation = useDeleteTelegramBot();

  useEffect(() => {
    if (isEditMode && existingBot) {
      setFormData({
        bot_token: '',
        bot_username: existingBot.bot_username || '',
        display_name: existingBot.display_name || '',
        is_auth_bot: existingBot.is_auth_bot ?? false,
        is_active: existingBot.is_active ?? true
      });
    }
  }, [existingBot, isEditMode]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value, type } = e.target;
    if (type === 'checkbox') {
      const checked = (e.target as HTMLInputElement).checked;
      setFormData(prev => ({ ...prev, [name]: checked }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
  };

  const handleToggleField = (field: 'is_auth_bot' | 'is_active') => {
    setFormData(prev => ({ ...prev, [field]: !prev[field] }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!applicationId) {
      logger.error('Application ID is required');
      return;
    }

    try {
      if (isNewMode) {
        await createMutation.mutateAsync({
          appId: applicationId,
          data: {
            bot_token: formData.bot_token,
            bot_username: formData.bot_username,
            display_name: formData.display_name,
            is_auth_bot: formData.is_auth_bot,
            is_active: formData.is_active
          }
        });
      } else if (botId) {
        await updateMutation.mutateAsync({
          appId: applicationId,
          id: botId,
          data: {
            bot_token: formData.bot_token || undefined,
            bot_username: formData.bot_username,
            display_name: formData.display_name,
            is_auth_bot: formData.is_auth_bot,
            is_active: formData.is_active
          }
        });
      }
      navigate(`/applications/${applicationId}/telegram-bots`);
    } catch (err) {
      logger.error('Failed to save bot:', err);
    }
  };

  const handleDelete = async () => {
    if (isEditMode && botId && applicationId) {
      const ok = await confirm({
        title: t('confirm.delete_title'),
        description: t('common.confirm_delete'),
        variant: 'danger'
      });
      if (ok) {
        try {
          await deleteMutation.mutateAsync({ appId: applicationId, id: botId });
          navigate(`/applications/${applicationId}/telegram-bots`);
        } catch (err) {
          logger.error('Failed to delete bot:', err);
        }
      }
    }
  };

  const isLoading = createMutation.isPending || updateMutation.isPending;

  if (isEditMode && loadingBot) {
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
          onClick={() => navigate(`/applications/${applicationId}/telegram-bots`)}
          className="p-2 hover:bg-accent rounded-lg transition-colors text-muted-foreground"
        >
          <ArrowLeft size={24} />
        </button>
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-lg bg-muted flex items-center justify-center">
            <Send className="text-blue-500" size={20} />
          </div>
          <h1 className="text-2xl font-bold text-foreground">
            {isEditMode ? t('tg.edit_title') : t('tg.add_title')}
          </h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 border-b border-border bg-muted flex items-start gap-3">
          <HelpCircle className="text-primary mt-0.5" size={20} />
          <div className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground mb-1">{t('tg.getting_started')}</p>
            <p>{t('tg.getting_started_desc')}</p>
          </div>
        </div>

        <TelegramBotFormFields
          formData={formData}
          isEditMode={!!isEditMode}
          isNewMode={!!isNewMode}
          showToken={showToken}
          onToggleShowToken={() => setShowToken(!showToken)}
          onChange={handleChange}
          onToggleField={handleToggleField}
        />

        <div className="px-6 py-4 bg-muted border-t border-border flex items-center justify-between">
          <div>
            {isEditMode && (
              <button
                type="button"
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="text-destructive hover:text-destructive text-sm font-medium px-2 py-1 rounded hover:bg-destructive/10 transition-colors disabled:opacity-50"
              >
                {t('common.delete')}
              </button>
            )}
          </div>
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => navigate(`/applications/${applicationId}/telegram-bots`)}
              className="px-4 py-2 text-sm font-medium text-muted-foreground bg-card border border-input rounded-lg hover:bg-accent focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring"
            >
              {t('common.cancel')}
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className={`flex items-center px-4 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-lg hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-ring
                ${isLoading ? 'opacity-70 cursor-not-allowed' : ''}`}
            >
              {isLoading ? (
                <Loader2 size={16} className="mr-2 animate-spin" />
              ) : (
                <Save size={16} className="mr-2" />
              )}
              {isEditMode ? t('common.save') : t('common.create')}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
};

export default TelegramBotEdit;
