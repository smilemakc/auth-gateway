import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader } from 'lucide-react';
import type { BulkUserCreate, BulkOperationResult } from '@auth-gateway/client-sdk';
import { useBulkCreateUsers } from '../../hooks/useBulkOperations';
import { toast } from '../../services/toast';
import { useLanguage } from '../../services/i18n';
import { logger } from '@/lib/logger';
import BulkCreateCSVParser from './BulkCreateCSVParser';
import BulkCreateManualEntry from './BulkCreateManualEntry';
import BulkCreateResults from './BulkCreateResults';

const BulkCreateUsers: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const bulkCreate = useBulkCreateUsers();

  const [users, setUsers] = useState<BulkUserCreate[]>([]);
  const [result, setResult] = useState<BulkOperationResult | null>(null);
  const [csvText, setCsvText] = useState('');
  const [mode, setMode] = useState<'manual' | 'csv' | 'json'>('manual');

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target?.result as string;
      if (file.name.endsWith('.csv')) {
        parseCSV(content);
        setMode('csv');
      } else if (file.name.endsWith('.json')) {
        parseJSON(content);
        setMode('json');
      }
    };
    reader.readAsText(file);
  };

  const parseCSV = (content: string) => {
    const lines = content.split('\n').filter((line) => line.trim());
    if (lines.length < 2) {
      toast.warning(t('bulk.csv_min_rows'));
      return;
    }

    const headers = lines[0].split(',').map((h) => h.trim().toLowerCase());
    const parsedUsers: BulkUserCreate[] = [];

    for (let i = 1; i < lines.length; i++) {
      const values = lines[i].split(',').map((v) => v.trim());
      const user: BulkUserCreate = {
        email: '',
        username: '',
        password: '',
        full_name: '',
        is_active: true,
        email_verified: false,
      };

      headers.forEach((header, index) => {
        const value = values[index] || '';
        switch (header) {
          case 'email':
            user.email = value;
            break;
          case 'username':
            user.username = value;
            break;
          case 'password':
            user.password = value;
            break;
          case 'full_name':
            user.full_name = value;
            break;
          case 'is_active':
            user.is_active = value.toLowerCase() === 'true' || value === '1';
            break;
          case 'email_verified':
            user.email_verified = value.toLowerCase() === 'true' || value === '1';
            break;
        }
      });

      if (user.email && user.username && user.password) {
        parsedUsers.push(user);
      }
    }

    setUsers(parsedUsers);
    setCsvText(content);
  };

  const parseJSON = (content: string) => {
    try {
      const data = JSON.parse(content);
      const parsedUsers: BulkUserCreate[] = Array.isArray(data) ? data : data.users || [];
      setUsers(parsedUsers);
    } catch (error) {
      toast.error(t('bulk.invalid_json'));
    }
  };

  const handleAddUser = () => {
    setUsers([
      ...users,
      {
        email: '',
        username: '',
        password: '',
        full_name: '',
        is_active: true,
        email_verified: false,
      },
    ]);
  };

  const handleRemoveUser = (index: number) => {
    setUsers(users.filter((_, i) => i !== index));
  };

  const handleUserChange = (index: number, field: keyof BulkUserCreate, value: string | boolean) => {
    const updated = [...users];
    updated[index] = { ...updated[index], [field]: value };
    setUsers(updated);
  };

  const handleSubmit = async () => {
    if (users.length === 0) {
      toast.warning(t('bulk.add_one_user'));
      return;
    }

    const validUsers = users.filter((u) => u.email && u.username && u.password);
    if (validUsers.length === 0) {
      toast.warning(t('bulk.fill_required'));
      return;
    }

    try {
      const result = await bulkCreate.mutateAsync({ users: validUsers });
      setResult(result);
    } catch (error) {
      logger.error('Bulk create failed:', error);
      toast.error(t('bulk.create_failed'));
    }
  };

  const handleCSVTextChange = (text: string) => {
    setCsvText(text);
    parseCSV(text);
  };

  const downloadTemplate = () => {
    const template = 'email,username,password,full_name,is_active,email_verified\njohn@example.com,john,password123,John Doe,true,false';
    const blob = new Blob([template], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'bulk-users-template.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleCreateMore = () => {
    setResult(null);
    setUsers([]);
    setCsvText('');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <button onClick={() => navigate('/bulk')} className="text-muted-foreground hover:text-foreground flex items-center gap-2">
          <ArrowLeft size={20} />
          {t('common.back')}
        </button>
        <h1 className="text-2xl font-bold text-foreground">{t('bulk.create_users')}</h1>
      </div>

      {!result ? (
        <>
          <BulkCreateCSVParser
            csvText={csvText}
            mode={mode}
            onFileUpload={handleFileUpload}
            onCSVTextChange={handleCSVTextChange}
            onDownloadTemplate={downloadTemplate}
          />

          <BulkCreateManualEntry
            users={users}
            onAddUser={handleAddUser}
            onRemoveUser={handleRemoveUser}
            onUserChange={handleUserChange}
          />

          <div className="flex justify-end gap-3">
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
            >
              {t('common.cancel')}
            </button>
            <button
              onClick={handleSubmit}
              disabled={bulkCreate.isPending || users.length === 0}
              className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {bulkCreate.isPending && <Loader size={16} className="animate-spin" />}
              {t('bulk.create')} {users.length > 0 && `(${users.length})`}
            </button>
          </div>
        </>
      ) : (
        <BulkCreateResults
          result={result}
          onCreateMore={handleCreateMore}
          onDone={() => navigate('/bulk')}
        />
      )}
    </div>
  );
};

export default BulkCreateUsers;
