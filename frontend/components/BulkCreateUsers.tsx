import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Upload, FileText, Loader, CheckCircle, XCircle, Download, ToggleLeft, ToggleRight } from 'lucide-react';
import type { BulkUserCreate, BulkOperationResult } from '@auth-gateway/client-sdk';
import { useBulkCreateUsers } from '../hooks/useBulkOperations';
import { toast } from '../services/toast';

const BulkCreateUsers: React.FC = () => {
  const navigate = useNavigate();
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
      toast.warning('CSV must have at least a header row and one data row');
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
      toast.error('Invalid JSON format');
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
      toast.warning('Please add at least one user');
      return;
    }

    const validUsers = users.filter((u) => u.email && u.username && u.password);
    if (validUsers.length === 0) {
      toast.warning('Please fill in email, username, and password for at least one user');
      return;
    }

    try {
      const result = await bulkCreate.mutateAsync({ users: validUsers });
      setResult(result);
    } catch (error) {
      console.error('Bulk create failed:', error);
      toast.error('Failed to create users');
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
          Back
        </button>
        <h1 className="text-2xl font-bold text-foreground">Bulk Create Users</h1>
      </div>

      {!result ? (
        <>
          {/* File Upload */}
          <div className="bg-card rounded-xl shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">Upload File</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <label className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg cursor-pointer flex items-center gap-2">
                  <Upload size={16} />
                  Choose File
                  <input type="file" accept=".csv,.json" onChange={handleFileUpload} className="hidden" />
                </label>
                <button
                  onClick={downloadTemplate}
                  className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent flex items-center gap-2"
                >
                  <Download size={16} />
                  Download Template
                </button>
              </div>

              {mode === 'csv' && (
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">CSV Content</label>
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
              <h2 className="text-lg font-semibold text-foreground">Users ({users.length})</h2>
              <button
                onClick={handleAddUser}
                className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg text-sm transition-colors"
              >
                + Add User
              </button>
            </div>

            <div className="space-y-4">
              {users.map((user, index) => (
                <div key={index} className="border border-border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-foreground">User #{index + 1}</span>
                    <button
                      onClick={() => handleRemoveUser(index)}
                      className="text-destructive hover:text-destructive/80 text-sm"
                    >
                      Remove
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">Email *</label>
                      <input
                        type="email"
                        value={user.email}
                        onChange={(e) => handleUserChange(index, 'email', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="user@example.com"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">Username *</label>
                      <input
                        type="text"
                        value={user.username}
                        onChange={(e) => handleUserChange(index, 'username', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="username"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">Password *</label>
                      <input
                        type="password"
                        value={user.password}
                        onChange={(e) => handleUserChange(index, 'password', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="password"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">Full Name</label>
                      <input
                        type="text"
                        value={user.full_name}
                        onChange={(e) => handleUserChange(index, 'full_name', e.target.value)}
                        className="w-full px-3 py-2 border border-input rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                        placeholder="Full Name"
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
                        <span className="text-xs text-foreground">Active</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handleUserChange(index, 'email_verified', !user.email_verified)}
                          className={`transition-colors ${user.email_verified ? 'text-success' : 'text-muted-foreground'}`}
                        >
                          {user.email_verified ? <ToggleRight size={24} /> : <ToggleLeft size={24} />}
                        </button>
                        <span className="text-xs text-foreground">Email Verified</span>
                      </div>
                    </div>
                  </div>
                </div>
              ))}

              {users.length === 0 && (
                <div className="text-center py-8 text-muted-foreground">
                  <FileText size={48} className="mx-auto mb-2 text-muted-foreground" />
                  <p>No users added. Click "Add User" or upload a file.</p>
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={bulkCreate.isPending || users.length === 0}
              className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {bulkCreate.isPending && <Loader size={16} className="animate-spin" />}
              Create {users.length > 0 && `(${users.length})`}
            </button>
          </div>
        </>
      ) : (
        <div className="bg-card rounded-xl shadow-sm border border-border p-6">
          <h2 className="text-lg font-semibold text-foreground mb-4">Operation Results</h2>
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-muted rounded-lg p-4">
                <div className="text-sm text-muted-foreground">Total</div>
                <div className="text-2xl font-bold text-foreground">{result.total}</div>
              </div>
              <div className="bg-success/10 rounded-lg p-4">
                <div className="text-sm text-success">Success</div>
                <div className="text-2xl font-bold text-success">{result.success}</div>
              </div>
              <div className="bg-destructive/10 rounded-lg p-4">
                <div className="text-sm text-destructive">Failed</div>
                <div className="text-2xl font-bold text-destructive">{result.failed}</div>
              </div>
            </div>

            {result.errors && result.errors.length > 0 && (
              <div>
                <h3 className="font-semibold text-foreground mb-2">Errors</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {result.errors.map((error, index) => (
                    <div key={index} className="bg-destructive/10 border border-border rounded-lg p-3">
                      <div className="text-sm text-destructive">
                        <span className="font-medium">Row {error.index + 1}:</span> {error.message}
                        {error.email && <span className="text-destructive"> ({error.email})</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {result.results && result.results.length > 0 && (
              <div>
                <h3 className="font-semibold text-foreground mb-2">Results</h3>
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
                          {item.email}: {item.message || (item.success ? 'Created' : 'Failed')}
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
                Create More
              </button>
              <button
                onClick={() => navigate('/bulk')}
                className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default BulkCreateUsers;

