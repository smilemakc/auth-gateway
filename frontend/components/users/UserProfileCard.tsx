import React from 'react';
import {
  Mail,
  Phone,
  Calendar,
  Clock,
  CheckCircle,
  User as UserIcon,
} from 'lucide-react';
import { useLanguage } from '../../services/i18n';
import { formatDate, formatRelative } from '../../lib/date';

interface Role {
  id: string;
  name: string;
  display_name?: string;
}

interface UserProfileCardProps {
  user: {
    profile_picture_url: string;
    full_name: string;
    username: string;
    email: string;
    email_verified: boolean;
    phone?: string;
    is_active: boolean;
    roles?: Role[];
    created_at: string;
    last_login?: string;
  };
}

const UserProfileCard: React.FC<UserProfileCardProps> = ({ user }) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
      <div className="p-6 text-center border-b border-border">
        <img
          src={user.profile_picture_url}
          alt={user.full_name}
          className="w-24 h-24 rounded-full mx-auto mb-4 border-4 border-muted"
        />
        <h2 className="text-xl font-bold text-foreground">{user.username}</h2>
        <div className="flex justify-center gap-2 flex-wrap mt-2">
          {user.roles?.map(role => (
            <span
              key={role.id}
              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize
                ${role.name === 'admin' ? 'bg-purple-100 text-purple-800' :
                  role.name === 'moderator' ? 'bg-indigo-100 text-indigo-800' : 'bg-muted text-foreground'}`}>
              {role.display_name || role.name}
            </span>
          ))}
          <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium
            ${user.is_active ? 'bg-success/20 text-success' : 'bg-destructive/20 text-destructive'}`}>
            {user.is_active ? t('users.active') : t('users.blocked')}
          </span>
        </div>
      </div>

      <div className="p-6 space-y-4">
        <div className="flex items-center gap-3 text-muted-foreground">
          <UserIcon size={18} className="text-muted-foreground" />
          <span className="text-sm font-medium">{user.username}</span>
        </div>
        <div className="flex items-center gap-3 text-muted-foreground">
          <Mail size={18} className="text-muted-foreground" />
          <span className="text-sm">{user.email}</span>
          {user.email_verified && <CheckCircle size={14} className="text-green-500 ml-auto" />}
        </div>
        <div className="flex items-center gap-3 text-muted-foreground">
          <Phone size={18} className="text-muted-foreground" />
          <span className="text-sm">{user.phone || '-'}</span>
        </div>
        <div className="pt-4 border-t border-border space-y-3">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground flex items-center gap-2">
              <Calendar size={16} /> {t('users.col_created')}
            </span>
            <span className="text-foreground">{formatDate(user.created_at)}</span>
          </div>
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground flex items-center gap-2">
              <Clock size={16} /> {t('user.login')}
            </span>
            <span className="text-foreground">
              {user.last_login ? formatRelative(user.last_login) : '-'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserProfileCard;
