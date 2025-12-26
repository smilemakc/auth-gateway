import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Upload, FileText, Loader, CheckCircle, XCircle, Download } from 'lucide-react';
import type { BulkUserCreate, BulkOperationResult } from '@auth-gateway/client-sdk';
import { useBulkCreateUsers } from '../hooks/useBulkOperations';

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
      alert('CSV must have at least a header row and one data row');
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
      alert('Invalid JSON format');
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
      alert('Please add at least one user');
      return;
    }

    const validUsers = users.filter((u) => u.email && u.username && u.password);
    if (validUsers.length === 0) {
      alert('Please fill in email, username, and password for at least one user');
      return;
    }

    try {
      const result = await bulkCreate.mutateAsync({ users: validUsers });
      setResult(result);
    } catch (error) {
      console.error('Bulk create failed:', error);
      alert('Failed to create users');
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
        <button onClick={() => navigate('/bulk')} className="text-gray-500 hover:text-gray-700 flex items-center gap-2">
          <ArrowLeft size={20} />
          Back
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Bulk Create Users</h1>
      </div>

      {!result ? (
        <>
          {/* File Upload */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Upload File</h2>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <label className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg cursor-pointer flex items-center gap-2">
                  <Upload size={16} />
                  Choose File
                  <input type="file" accept=".csv,.json" onChange={handleFileUpload} className="hidden" />
                </label>
                <button
                  onClick={downloadTemplate}
                  className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 flex items-center gap-2"
                >
                  <Download size={16} />
                  Download Template
                </button>
              </div>

              {mode === 'csv' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">CSV Content</label>
                  <textarea
                    value={csvText}
                    onChange={(e) => {
                      setCsvText(e.target.value);
                      parseCSV(e.target.value);
                    }}
                    rows={10}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-xs"
                  />
                </div>
              )}
            </div>
          </div>

          {/* Manual Entry */}
          <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-900">Users ({users.length})</h2>
              <button
                onClick={handleAddUser}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors"
              >
                + Add User
              </button>
            </div>

            <div className="space-y-4">
              {users.map((user, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm font-medium text-gray-700">User #{index + 1}</span>
                    <button
                      onClick={() => handleRemoveUser(index)}
                      className="text-red-600 hover:text-red-800 text-sm"
                    >
                      Remove
                    </button>
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Email *</label>
                      <input
                        type="email"
                        value={user.email}
                        onChange={(e) => handleUserChange(index, 'email', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="user@example.com"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Username *</label>
                      <input
                        type="text"
                        value={user.username}
                        onChange={(e) => handleUserChange(index, 'username', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="username"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Password *</label>
                      <input
                        type="password"
                        value={user.password}
                        onChange={(e) => handleUserChange(index, 'password', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="password"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-gray-700 mb-1">Full Name</label>
                      <input
                        type="text"
                        value={user.full_name}
                        onChange={(e) => handleUserChange(index, 'full_name', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        placeholder="Full Name"
                      />
                    </div>
                    <div className="flex items-center gap-4">
                      <label className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={user.is_active}
                          onChange={(e) => handleUserChange(index, 'is_active', e.target.checked)}
                          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-xs text-gray-700">Active</span>
                      </label>
                      <label className="flex items-center gap-2">
                        <input
                          type="checkbox"
                          checked={user.email_verified}
                          onChange={(e) => handleUserChange(index, 'email_verified', e.target.checked)}
                          className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-xs text-gray-700">Email Verified</span>
                      </label>
                    </div>
                  </div>
                </div>
              ))}

              {users.length === 0 && (
                <div className="text-center py-8 text-gray-500">
                  <FileText size={48} className="mx-auto mb-2 text-gray-400" />
                  <p>No users added. Click "Add User" or upload a file.</p>
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-3">
            <button
              onClick={() => navigate('/bulk')}
              className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={bulkCreate.isPending || users.length === 0}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {bulkCreate.isPending && <Loader size={16} className="animate-spin" />}
              Create {users.length > 0 && `(${users.length})`}
            </button>
          </div>
        </>
      ) : (
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Operation Results</h2>
          <div className="space-y-4">
            <div className="grid grid-cols-3 gap-4">
              <div className="bg-gray-50 rounded-lg p-4">
                <div className="text-sm text-gray-500">Total</div>
                <div className="text-2xl font-bold text-gray-900">{result.total}</div>
              </div>
              <div className="bg-green-50 rounded-lg p-4">
                <div className="text-sm text-green-600">Success</div>
                <div className="text-2xl font-bold text-green-700">{result.success}</div>
              </div>
              <div className="bg-red-50 rounded-lg p-4">
                <div className="text-sm text-red-600">Failed</div>
                <div className="text-2xl font-bold text-red-700">{result.failed}</div>
              </div>
            </div>

            {result.errors && result.errors.length > 0 && (
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Errors</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {result.errors.map((error, index) => (
                    <div key={index} className="bg-red-50 border border-red-200 rounded-lg p-3">
                      <div className="text-sm text-red-800">
                        <span className="font-medium">Row {error.index + 1}:</span> {error.message}
                        {error.email && <span className="text-red-600"> ({error.email})</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {result.results && result.results.length > 0 && (
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Results</h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {result.results.map((item, index) => (
                    <div
                      key={index}
                      className={`border rounded-lg p-3 ${
                        item.success ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200'
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        {item.success ? (
                          <CheckCircle className="text-green-600" size={16} />
                        ) : (
                          <XCircle className="text-red-600" size={16} />
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

            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
              <button
                onClick={() => {
                  setResult(null);
                  setUsers([]);
                  setCsvText('');
                }}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                Create More
              </button>
              <button
                onClick={() => navigate('/bulk')}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
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

