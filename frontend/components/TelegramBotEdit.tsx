
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Save, Trash2, HelpCircle, Eye, EyeOff, Loader2, Send, ToggleLeft, ToggleRight } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useTelegramBotDetail, useCreateTelegramBot, useUpdateTelegramBot, useDeleteTelegramBot } from '../hooks/useTelegramBots';
import { confirm } from '../services/confirm';

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!applicationId) {
      console.error('Application ID is required');
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
      console.error('Failed to save bot:', err);
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
          console.error('Failed to delete bot:', err);
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
            {isEditMode ? 'Edit Telegram Bot' : 'Add Telegram Bot'}
          </h1>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
        <div className="p-6 border-b border-border bg-muted flex items-start gap-3">
          <HelpCircle className="text-primary mt-0.5" size={20} />
          <div className="text-sm text-muted-foreground">
            <p className="font-medium text-foreground mb-1">Getting Started</p>
            <p>Create a bot via @BotFather on Telegram to get the bot token and username. Auth bots are used for user authentication via Telegram.</p>
          </div>
        </div>

        <div className="p-6 space-y-8">
          {/* Bot Token */}
          <div>
            <label htmlFor="bot_token" className="block text-sm font-medium text-muted-foreground mb-1">
              Bot Token
            </label>
            <div className="relative">
              <input
                type={showToken ? "text" : "password"}
                id="bot_token"
                name="bot_token"
                value={formData.bot_token}
                onChange={handleChange}
                required={isNewMode}
                className="w-full pl-4 pr-12 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
                placeholder={isEditMode ? '(leave blank to keep current)' : 'e.g. 123456789:ABCdefGHIjklMNOpqrsTUVwxyz'}
              />
              <button
                type="button"
                onClick={() => setShowToken(!showToken)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showToken ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              Get this from @BotFather on Telegram
            </p>
          </div>

          {/* Bot Username */}
          <div>
            <label htmlFor="bot_username" className="block text-sm font-medium text-muted-foreground mb-1">
              Bot Username
            </label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-muted-foreground">@</span>
              <input
                type="text"
                id="bot_username"
                name="bot_username"
                value={formData.bot_username}
                onChange={handleChange}
                required
                className="w-full pl-8 pr-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all font-mono text-sm"
                placeholder="your_bot_username"
              />
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              The username without @ symbol
            </p>
          </div>

          {/* Display Name */}
          <div>
            <label htmlFor="display_name" className="block text-sm font-medium text-muted-foreground mb-1">
              Display Name
            </label>
            <input
              type="text"
              id="display_name"
              name="display_name"
              value={formData.display_name}
              onChange={handleChange}
              required
              className="w-full px-4 py-2 border border-input rounded-lg focus:ring-2 focus:ring-ring focus:border-transparent outline-none transition-all text-sm"
              placeholder="e.g. My App Auth Bot"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Friendly name for this bot in the admin panel
            </p>
          </div>

          {/* Bot Settings */}
          <div className="pt-6 border-t border-border space-y-4">
            <div className="flex items-start gap-3">
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, is_auth_bot: !prev.is_auth_bot }))}
                className={`transition-colors mt-0.5 ${formData.is_auth_bot ? 'text-success' : 'text-muted-foreground'}`}
              >
                {formData.is_auth_bot ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <div>
                <span className="font-medium text-foreground block">
                  Authentication Bot
                </span>
                <p className="text-sm text-muted-foreground mt-1">
                  Use this bot for user authentication via Telegram
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <button
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, is_active: !prev.is_active }))}
                className={`transition-colors mt-0.5 ${formData.is_active ? 'text-success' : 'text-muted-foreground'}`}
              >
                {formData.is_active ? <ToggleRight size={28} /> : <ToggleLeft size={28} />}
              </button>
              <div>
                <span className="font-medium text-foreground block">
                  Active
                </span>
                <p className="text-sm text-muted-foreground mt-1">
                  Inactive bots cannot be used for authentication
                </p>
              </div>
            </div>
          </div>
        </div>

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
