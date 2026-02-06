import React from 'react';
import {Clock, Key, MapPin, Monitor, Trash2, User} from 'lucide-react';
import {useRevokeSession, useSessions} from '../hooks/useSessions';
import { useLanguage } from '../services/i18n';

const Sessions: React.FC = () => {
    const {data, isLoading, error} = useSessions(1, 100);
    const revokeSession = useRevokeSession();
    const { t } = useLanguage();

    const sessions = data?.sessions || [];

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <div
                    className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-8 text-center">
                <p className="text-destructive">{t('sessions.error_loading')}: {(error as Error).message}</p>
            </div>
        );
    }

    const handleRevoke = (sessionId: string) => {
        if (confirm(t('sessions.revoke_confirm'))) {
            revokeSession.mutate(sessionId);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold text-foreground">{t('sessions.title')}</h1>
                <span className="text-sm text-muted-foreground">{sessions.length} {t('sessions.active_sessions')}</span>
            </div>

            {sessions.length === 0 && !isLoading && (
                <div className="text-center py-12 bg-card rounded-xl border border-border">
                    <Key className="mx-auto h-12 w-12 text-muted-foreground" />
                    <h3 className="mt-2 text-sm font-semibold text-foreground">{t('sessions.no_sessions')}</h3>
                    <p className="mt-1 text-sm text-muted-foreground">{t('sessions.no_sessions_desc')}</p>
                </div>
            )}

            {sessions.length > 0 && (
                <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full text-left text-sm whitespace-nowrap">
                            <thead className="uppercase tracking-wider border-b border-border bg-muted">
                            <tr>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_name')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_user')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_device')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_os')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_ip')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_user_agent')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_last_active')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('sessions.col_created')}</th>
                                <th scope="col" className="px-6 py-4 font-semibold text-foreground">{t('common.actions')}</th>
                            </tr>
                            </thead>
                            <tbody className="divide-y divide-border">
                            {sessions.map((session: any) => (
                                <tr key={session.id} className="hover:bg-accent">
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <User size={16}/>
                                            {session.session_name || t('common.unknown')}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <User size={16}/>
                                            {session.user_id || t('common.unknown')}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Monitor size={16}/>
                                            <span className="truncate max-w-xs" title={session.device_type}>
                                                    {(session.device_type || t('common.unknown'))}
                                                </span>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Monitor size={16}/>
                                            <span className="truncate max-w-xs" title={session.os}>
                                                    {(session.os || t('common.unknown'))}
                                                </span>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground font-mono text-xs">
                                            <MapPin size={14}/>
                                            {session.ip_address || '-'}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Monitor size={16}/>
                                            <span className="truncate max-w-xs" title={session.user_agent}>
                                                    {(session.user_agent || t('common.unknown')).slice(0, 40)}...
                                                </span>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-muted-foreground">
                                        <div className="flex items-center gap-2">
                                            <Clock size={14}/>
                                            {new Date(session.last_active_at || session.updated_at).toLocaleString()}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-muted-foreground">
                                        {new Date(session.created_at).toLocaleString()}
                                    </td>
                                    <td className="px-6 py-4">
                                        <button
                                            onClick={() => handleRevoke(session.id)}
                                            className="text-destructive hover:text-destructive p-1 rounded"
                                            title={t('sessions.revoke')}
                                        >
                                            <Trash2 size={16}/>
                                        </button>
                                    </td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </div>
                    <div className="p-4 border-t border-border flex justify-between items-center text-sm text-muted-foreground">
                        <span>{t('sessions.showing')} {sessions.length} {t('sessions.sessions_count')}</span>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Sessions;
