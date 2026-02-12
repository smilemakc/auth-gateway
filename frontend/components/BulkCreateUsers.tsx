import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Upload, FileText, Loader, CheckCircle, XCircle, Download, ToggleLeft, ToggleRight } from 'lucide-react';
import type { BulkUserCreate, BulkOperationResult } from '@auth-gateway/client-sdk';
import { useBulkCreateUsers } from '../hooks/useBulkOperations';
import { toast } from '../services/toast';
import { useLanguage } from '../services/i18n';

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
      console.error('Bulk create failed:', error);
      toast.error(t('bulk.create_failed'));
    }
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
          {/* File Upload */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">{t('bulk.upload_file')}</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <label className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg cursor-pointer flex items-center gap-2">
                  <Upload size={16} />
                  {t('bulk.choose_file')}
                  <input type="file" accept=".csv,.json" onChange={handleFileUpload} className="hidden" />
                </label>
                <button
                  onClick={downloadTemplate}
                  className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent flex items-center gap-2"
                >
                  <Download size={16} />
                  {t('bulk.download_template')}
                </button>
              </div>

              {mode === 'csv' && (
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">{t('bulk.csv_content')}</label>
                  <textarea
                    value={csvText}
                    onChange={(e) => {
                      setCsvText(e.target.value);
                      parseCSV(e.target.value);
                    }}
                    rows={10}
                    className="w-full px-3 py-2 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring font-mono text-xs"
                  />
                </div>
              )}
            </div>
          </div>

          {/* Manual Entry */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-foreground">{t('bulk.users_count', { count: users.length })}</h2>
              <button
                onClick={handleAddUser}
                className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors"
              >
                + {t('bulk.add_user')}
              </button>
            </div>

            <div className="space-y-4">
              {users.map((user, index) => (
                <div key={index} className="border border-border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-foreground">{t('bulk.user_number', { number: index + 1 })}</span>
                    <button
                      onClick={() => handleRemoveUser(index)}
                      className="text-destructive hover:text-destructive/80 text-sm"
                    >
                      {t('common.remove')}
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_email')} *</label>
                      <input
                        type="email"
                        value={user.email}
                        onChange={(e) => handleUserChange(index, 'email', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="user@example.com"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_username')} *</label>
                      <input
                        type="text"
                        value={user.username}
                        onChange={(e) => handleUserChange(index, 'username', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="username"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_password')} *</label>
                      <input
                        type="password"
                        value={user.password}
                        onChange={(e) => handleUserChange(index, 'password', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="password"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">{t('users.col_full_name')}</label>
                      <input
                        type="text"
                        value={user.full_name}
                        onChange={(e) => handleUserChange(index, 'full_name', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder={t('users.col_full_name')}
                      />
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleUserChange(index, 'is_active', !user.is_active)}
                          className={`transition-colors ${user.is_active ? 'text-success' : 'text-muted-foreground'}`}
                        >
                          {user.is_active ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                        </button>
                        <span className="text-xs text-foreground">{t('users.active')}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleUserChange(index, 'email_verified', !user.email_verified)}
                          className={`transition-colors ${user.email_verified ? 'text-success' : 'text-muted-foreground'}`}
                        >
                          {user.email_verified ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                        </button>
                        <span className="text-xs text-foreground">{t('bulk.email_verified')}</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}

              {users.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">
                  <FileText size={48} className="mx-auto mb-2 text-muted-foreground" />
                  <p>{t('bulk.no_users_added')}</p>
                </div>
              )}
            </div>
          </div>

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
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">{t('bulk.operation_results')}</h2>
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-muted rounded-lg p-4">
                <div className="text-sm text-muted-foreground">{t('bulk.total')}</div>
                <div className="text-2xl font-bold text-foreground">{result.total}</div>
              </div>
              <div className="bg-success/10 rounded-lg p-4">
                <div className="text-sm text-success">{t('bulk.success')}</div>
                <div className="text-2xl font-bold text-success">{result.success}</div>
              </div>
              <div className="bg-destructive/10 rounded-lg p-4">
                <div className="text-sm text-destructive">{t('bulk.failed')}</div>
                <div className="text-2xl font-bold text-destructive">{result.failed}</div>
              </div>
            </div>

            {result.errors && result.errors.length > 0 && (
              <div>
                <h3 className="font-semibold text-foreground mb-2">{t('bulk.errors')}</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {result.errors.map((error, index) => (
                    <div key={index} className="bg-destructive/10 border border-border rounded-lg p-3">
                      <div className="text-sm text-destructive">
                        <span className="font-medium">{t('bulk.row', { row: error.index + 1 })}:</span> {error.message}
                        {error.email && <span className="text-destructive"> ({error.email})</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {result.results && result.results.length > 0 && (
              <div>
                <h3 className="font-semibold text-foreground mb-2">{t('bulk.results')}</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {result.results.map((item, index) => (
                    <div
                      key={index}
                      className={`border rounded-lg p-3 ${
                        item.success ? 'bg-success/10 border-border' : 'bg-destructive/10 border-border'
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {item.success ? (
                          <CheckCircle className="text-success" size={16} />
                        ) : (
                          <XCircle className="text-destructive" size={16} />
                        )}
                        <span className="text-sm">
                          {item.email}: {item.message || (item.success ? t('bulk.created') : t('bulk.failed'))}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-4 border-t border-border">
              <button
                onClick={() => {
                  setResult(null);
                  setUsers([]);
                  setCsvText('');
                }}
                className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
              >
                {t('bulk.create_more')}
              </button>
              <button
                onClick={() => navigate('/bulk')}
                className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
              >
                {t('common.done')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default BulkCreateUsers;
