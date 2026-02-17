import React from 'react';
import {
  Globe,
  Monitor,
  Smartphone,
  Tablet,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatRelative } from '../../lib/date';

interface Session {
  id: string;
  user_agent: string;
  ip_address: string;
  is_current: boolean;
  last_active_at: string;
}

interface UserSessionsSectionProps {
  sessions: Session[];
  onRevokeSession: (sessionId: string) => void;
}

function getDeviceIcon(userAgent: string) {
  const ua = userAgent.toLowerCase();
  if (ua.includes('mobile') || ua.includes('android') || ua.includes('iphone')) {
    return <Smartphone size={18} />;
  }
  if (ua.includes('tablet') || ua.includes('ipad')) {
    return <Tablet size={18} />;
  }
  return <Monitor size={18} />;
}

function parseUserAgent(userAgent: string) {
  const browsers = ['Chrome', 'Firefox', 'Safari', 'Edge', 'Opera'];
  const oses = ['Windows', 'Mac OS', 'Linux', 'Android', 'iOS'];

  let browser = 'Unknown Browser';
  let os = 'Unknown OS';

  for (const b of browsers) {
    if (userAgent.includes(b)) {
      browser = b;
      break;
    }
  }

  for (const o of oses) {
    if (userAgent.includes(o)) {
      os = o;
      break;
    }
  }

  return { browser, os };
}

const UserSessionsSection: React.FC<UserSessionsSectionProps> = ({ sessions, onRevokeSession }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 border-b border-border flex items-center justify-between">
        <h3 className="font-semibold text-foreground flex items-center gap-2">
          <Monitor size={18} className="text-primary" />
          {t('user.sessions')}
        </h3>
        <span className="text-xs font-medium bg-muted text-muted-foreground px-2 py-1 rounded-full">{sessions.length}</span>
      </div>
      <div className="divide-y divide-border">
        {sessions.length > 0 ? (
          sessions.map(session => {
            const { browser, os } = parseUserAgent(session.user_agent);
            return (
              <div key={session.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between hover:bg-accent gap-4">
                <div className="flex items-start gap-3">
                  <div className={`p-2 rounded-lg bg-muted text-muted-foreground`}>
                    {getDeviceIcon(session.user_agent)}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-foreground">{os} - {browser}</p>
                      {session.is_current && (
                        <span className="text-[10px] font-bold bg-success/20 text-success px-1.5 py-0.5 rounded uppercase">{t('user.current')}</span>
                      )}
                    </div>
                    <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground mt-1">
                      <span className="flex items-center gap-1">
                        <Globe size={12} /> {session.ip_address}
                      </span>
                      <span className="flex items-center gap-1">
                        • Active {formatRelative(session.last_active_at)}
                      </span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => onRevokeSession(session.id)}
                  className="text-sm text-destructive hover:text-destructive hover:bg-destructive/10 px-3 py-1.5 rounded-md font-medium transition-colors border border-transparent hover:border-destructive/20 self-start sm:self-center"
                >
                  {t('user.revoke')}
                </button>
              </div>
            );
          })
        ) : (
          <div className="p-6 text-sm text-muted-foreground italic">{t('user.no_sessions')}</div>
        )}
      </div>
    </div>
  );
};

export default UserSessionsSection;
