import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader } from 'lucide-react';
import type { BulkUserUpdate, BulkOperationResult } from '@auth-gateway/client-sdk';
import { useBulkUpdateUsers } from '../../hooks/useBulkOperations';
import { useUsers } from '../../hooks/useUsers';
import { toast } from '../../services/toast';
import { useLanguage } from '../../services/i18n';
import { logger } from '@/lib/logger';
import BulkUpdateFieldSelector from './BulkUpdateFieldSelector';
import BulkUpdateResults from './BulkUpdateResults';

const BulkUpdateUsers: React.FC = () => {
  const { t } = useLanguage();
  const navigate = useNavigate();
  const bulkUpdate = useBulkUpdateUsers();
  const { data: usersData } = useUsers(1, 1000);

  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [updateFields, setUpdateFields] = useState<{
    email?: string;
    username?: string;
    full_name?: string;
    is_active?: boolean;
  }>({});
  const [result, setResult] = useState<BulkOperationResult | null>(null);
  const [searchTerm, setSearchTerm] = useState('');

  const users = usersData?.users || [];
  const filteredUsers = searchTerm
    ? users.filter(
        (u) =>
          u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
          u.username.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : users;

  const toggleUserSelection = (userId: string) => {
    if (selectedUserIds.includes(userId)) {
      setSelectedUserIds(selectedUserIds.filter((id) => id !== userId));
    } else {
      setSelectedUserIds([...selectedUserIds, userId]);
    }
  };

  const handleSelectAll = () => {
    if (selectedUserIds.length === filteredUsers.length) {
      setSelectedUserIds([]);
    } else {
      setSelectedUserIds(filteredUsers.map((u) => u.id));
    }
  };

  const handleSubmit = async () => {
    if (selectedUserIds.length === 0) {
      toast.warning(t('bulk_update.select_user_warning'));
      return;
    }

    const updates: BulkUserUpdate[] = selectedUserIds.map((id) => ({
      id,
      email: updateFields.email || undefined,
      username: updateFields.username || undefined,
      full_name: updateFields.full_name || undefined,
      is_active: updateFields.is_active !== undefined ? updateFields.is_active : undefined,
    }));

    try {
      const result = await bulkUpdate.mutateAsync({ users: updates });
      setResult(result);
    } catch (error) {
      logger.error('Bulk update failed:', error);
      toast.error(t('bulk_update.update_error'));
    }
  };

  const handleUpdateMore = () => {
    setResult(null);
    setSelectedUserIds([]);
    setUpdateFields({});
  };

  if (result) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
            <ArrowLeft size={20} />
            {t('common.back')}
          </button>
          <h1 className="text-2xl font-bold text-foreground">{t('bulk_update.results_title')}</h1>
        </div>

        <BulkUpdateResults
          result={result}
          onUpdateMore={handleUpdateMore}
          onDone={() => navigate('/bulk')}
        />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
          <ArrowLeft size={20} />
          {t('common.back')}
        </button>
        <h1 className="text-2xl font-bold text-foreground">{t('bulk.update_users')}</h1>
      </div>

      <div className="bg-card rounded-xl shadow-sm border border-border p-6 space-y-6">
        <BulkUpdateFieldSelector
          updateFields={updateFields}
          onUpdateFieldsChange={setUpdateFields}
          filteredUsers={filteredUsers}
          selectedUserIds={selectedUserIds}
          searchTerm={searchTerm}
          onSearchTermChange={setSearchTerm}
          onToggleUserSelection={toggleUserSelection}
          onSelectAll={handleSelectAll}
        />

        <div className="flex justify-end gap-3 pt-4 border-t border-border">
          <button
            onClick={() => navigate('/bulk')}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            {t('common.cancel')}
          </button>
          <button
            onClick={handleSubmit}
            disabled={bulkUpdate.isPending || selectedUserIds.length === 0}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {bulkUpdate.isPending && <Loader size={16} className="animate-spin" />}
            {t('bulk_update.update')} {selectedUserIds.length > 0 && `(${selectedUserIds.length})`}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BulkUpdateUsers;
