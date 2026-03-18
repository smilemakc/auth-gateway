import React from 'react';
import { Link } from 'react-router-dom';
import { Edit2, Trash2, ToggleLeft, ToggleRight, Key, Copy, Check, Shield, Globe } from 'lucide-react';
import type { OAuthClient } from '@auth-gateway/client-sdk';
import { formatDate } from '../../lib/date';
import { useLanguage } from '../../services/i18n';

interface OAuthClientCardProps {
  client: OAuthClient;
  copiedId: string | null;
  isToggling: boolean;
  isDeleting: boolean;
  isRotating: boolean;
  onToggle: (client: OAuthClient) => void;
  onDelete: (id: string) => void;
  onRotateSecret: (clientId: string) => void;
  onCopy: (text: string, id: string) => void;
}

const getClientTypeIcon = (clientType: string) => {
  if (clientType === 'confidential') {
    return <Shield className="text-success" size={20} />;
  }
  return <Globe className="text-primary" size={20} />;
};

export const OAuthClientCard: React.FC<OAuthClientCardProps> = ({
  client,
  copiedId,
  isToggling,
  isDeleting,
  isRotating,
  onToggle,
  onDelete,
  onRotateSecret,
  onCopy,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden flex flex-col">
      <div className="p-6 flex-1">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-xl bg-muted flex items-center justify-center shadow-sm">
              {getClientTypeIcon(client.client_type)}
            </div>
            <div>
              <h3 className="font-semibold text-foreground text-lg">{client.name}</h3>
              <div className="flex items-center gap-2 mt-1">
                <span className={`w-2 h-2 rounded-full ${client.is_active ? 'bg-success' : 'bg-muted-foreground'}`}></span>
                <span className="text-xs text-muted-foreground font-medium uppercase tracking-wide">
                  {client.is_active ? t('common.active') : t('common.inactive')}
                </span>
                <span className="text-xs text-muted-foreground">|</span>
                <span className="text-xs text-muted-foreground capitalize">{client.client_type}</span>
              </div>
            </div>
          </div>
          <button
            onClick={() => onToggle(client)}
            className={`transition-colors ${client.is_active ? 'text-success hover:text-success' : 'text-muted-foreground hover:text-muted-foreground'}`}
            disabled={isToggling}
          >
            {client.is_active ? <ToggleRight size={36} /> : <ToggleLeft size={36} />}
          </button>
        </div>

        {client.description && (
          <p className="text-sm text-muted-foreground mb-4 line-clamp-2">{client.description}</p>
        )}

        <div className="space-y-3">
          <div>
            <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('oauth_clients.client_id')}</label>
            <div className="flex items-center gap-2">
              <code className="flex-1 bg-muted rounded px-3 py-2 text-sm text-muted-foreground font-mono truncate border border-border">
                {client.client_id}
              </code>
              <button
                onClick={() => onCopy(client.client_id, `client-${client.id}`)}
                className="p-1.5 text-muted-foreground hover:text-foreground hover:bg-accent rounded"
              >
                {copiedId === `client-${client.id}` ? <Check size={14} /> : <Copy size={14} />}
              </button>
            </div>
          </div>
          <div>
            <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('oauth_clients.redirect_uris')}</label>
            <div className="text-xs text-muted-foreground truncate" title={client.redirect_uris.join(', ')}>
              {client.redirect_uris.length > 0 ? client.redirect_uris[0] : <span className="italic text-muted-foreground">{t('oauth_clients.none_configured')}</span>}
              {client.redirect_uris.length > 1 && (
                <span className="text-muted-foreground"> (+{client.redirect_uris.length - 1} {t('oauth_clients.more_uris')})</span>
              )}
            </div>
          </div>
          <div>
            <label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block mb-1">{t('oauth_clients.grant_types')}</label>
            <div className="flex flex-wrap gap-1">
              {client.allowed_grant_types.slice(0, 2).map((grant) => (
                <span key={grant} className="px-2 py-0.5 bg-primary/10 text-primary text-xs rounded">
                  {grant.replace('urn:ietf:params:oauth:grant-type:', '')}
                </span>
              ))}
              {client.allowed_grant_types.length > 2 && (
                <span className="px-2 py-0.5 bg-muted text-muted-foreground text-xs rounded">
                  +{client.allowed_grant_types.length - 2}
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="bg-muted px-6 py-4 border-t border-border flex items-center justify-between">
        <span className="text-xs text-muted-foreground">
          {formatDate(client.created_at)}
        </span>
        <div className="flex items-center gap-1">
          <button
            onClick={() => onRotateSecret(client.id)}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
            title={t('oauth_clients.rotate_secret')}
            disabled={isRotating}
          >
            <Key size={18} />
          </button>
          <Link
            to={`/oauth-clients/${client.id}`}
            className="p-2 text-muted-foreground hover:text-primary hover:bg-primary/10 rounded-lg transition-colors"
          >
            <Edit2 size={18} />
          </Link>
          <button
            onClick={() => onDelete(client.id)}
            className="p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
            disabled={isDeleting}
          >
            <Trash2 size={18} />
          </button>
        </div>
      </div>
    </div>
  );
};
