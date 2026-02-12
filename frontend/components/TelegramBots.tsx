
import React from 'react';
import { Link, useParams } from 'react-router-dom';
import { Plus, Edit2, Trash2, Send, Loader2, Shield, ShieldOff } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useTelegramBots, useDeleteTelegramBot } from '../hooks/useTelegramBots';
import { formatDate } from '../lib/date';
import { confirm } from '../services/confirm';

interface TelegramBotsProps {
  applicationId: string;
}

const TelegramBots: React.FC<TelegramBotsProps> = ({ applicationId }) => {
  const { t } = useLanguage();
  const { data: botsResponse, isLoading, error } = useTelegramBots(applicationId);
  const deleteBotMutation = useDeleteTelegramBot();

  const bots = Array.isArray(botsResponse) ? botsResponse : [];

  const handleDelete = async (botId: string) => {
    const ok = await confirm({
      title: t('confirm.delete_title'),
      description: t('common.confirm_delete'),
      variant: 'danger'
    });
    if (ok) {
      try {
        await deleteBotMutation.mutateAsync({ appId: applicationId, id: botId });
      } catch (err) {
        console.error('Failed to delete bot:', err);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 text-destructive">
        {t('tg.load_error')}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-bold text-foreground">{t('tg.title')}</h2>
          <p className="text-muted-foreground mt-1">{t('tg.desc')}</p>
        </div>
        <Link
          to={`/applications/${applicationId}/telegram-bots/new`}
          className="flex items-center gap-2 bg-primary hover:bg-primary-600 text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={18} />
          {t('tg.add_bot')}
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
        {bots.map((bot) => (
          <div key={bot.id} className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
            <div className="p-6 flex-1">
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center shadow-sm">
                    <Send className="text-blue-500" size={24} />
                  </div>
                  <div>
                    <h3 className="font-semibold text-foreground text-lg">{bot.display_name}</h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span className={`w-2 h-2 rounded-full ${bot.is_active ? 'bg-green-500' : 'bg-gray-300'}`}></span>
                      <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                        {bot.is_active ? t('common.active') : t('common.inactive')}
                      </span>
                    </div>
                  </div>
                </div>
                <div>
                  {bot.is_auth_bot ? (
                    <Shield className="text-primary" size={20} title={t('tg.auth_bot')} />
                  ) : (
                    <ShieldOff className="text-muted-foreground" size={20} title={t('tg.not_auth_bot')} />
                  )}
                </div>
              </div>

              <div className="space-y-3 mt-6">
                <div>
                  <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('tg.bot_username')}</label>
                  <code className="block bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono truncate border border-border">
                    @{bot.bot_username}
                  </code>
                </div>
                {bot.is_auth_bot && (
                  <div className="flex items-center gap-2 text-xs text-primary bg-primary/10 px-3 py-2 rounded-lg border border-primary/20">
                    <Shield size={14} />
                    <span className="font-medium">{t('tg.auth_bot')}</span>
                  </div>
                )}
              </div>
            </div>

            <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {bot.created_at ? formatDate(bot.created_at) : '-'}
              </span>
              <div className="flex items-center gap-2">
                <Link
                  to={`/applications/${applicationId}/telegram-bots/${bot.id}`}
                  className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
                >
                  <Edit2 size={18} />
                </Link>
                <button
                  onClick={() => handleDelete(bot.id)}
                  disabled={deleteBotMutation.isPending}
                  className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors disabled:opacity-50"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          </div>
        ))}

        {bots.length === 0 && (
          <div className="col-span-full text-center py-12 bg-card rounded-xl border border-border">
            <Send size={48} className="mx-auto mb-4 text-muted-foreground opacity-50" />
            <p className="text-muted-foreground">{t('tg.no_bots')}</p>
            <Link
              to={`/applications/${applicationId}/telegram-bots/new`}
              className="mt-4 inline-block text-primary hover:underline text-sm font-medium"
            >
              {t('tg.add_first')}
            </Link>
          </div>
        )}
      </div>
    </div>
  );
};

export default TelegramBots;
