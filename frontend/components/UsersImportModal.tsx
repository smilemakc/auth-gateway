import React, { useState } from 'react';
import { Upload, X, ChevronLeft, CheckCircle, AlertCircle, ChevronDown, ChevronUp } from 'lucide-react';
import { useLanguage } from '../services/i18n';
import { useImportUsers } from '../hooks/useApplications';
import { toast } from '../services/toast';
import type { ImportUserEntry, ImportUsersRequest, ImportUsersResponse, ImportDetail } from '../types';

interface UsersImportModalProps {
  applicationId: string;
  onClose: () => void;
  onSuccess?: () => void;
}

type Step = 'input' | 'preview' | 'result';

const UsersImportModal: React.FC<UsersImportModalProps> = ({ applicationId, onClose, onSuccess }) => {
  const { t } = useLanguage();
  const [step, setStep] = useState<Step>('input');
  const [jsonInput, setJsonInput] = useState('');
  const [parsedUsers, setParsedUsers] = useState<ImportUserEntry[]>([]);
  const [onConflict, setOnConflict] = useState<'skip' | 'update' | 'error'>('skip');
  const [result, setResult] = useState<ImportUsersResponse | null>(null);
  const [parseError, setParseError] = useState('');
  const [showAllDetails, setShowAllDetails] = useState(false);

  const importUsers = useImportUsers();

  const handleParse = () => {
    setParseError('');
    try {
      const parsed = JSON.parse(jsonInput);
      if (!Array.isArray(parsed)) {
        setParseError(t('apps.import.parse_error'));
        return;
      }
      const valid = parsed.every(entry => entry && typeof entry === 'object' && entry.email);
      if (!valid) {
        setParseError(t('apps.import.parse_error'));
        return;
      }
      setParsedUsers(parsed);
      setStep('preview');
    } catch (error) {
      setParseError(t('apps.import.parse_error'));
    }
  };

  const handleImport = async () => {
    try {
      const data: ImportUsersRequest = {
        users: parsedUsers,
        on_conflict: onConflict,
      };
      const response = await importUsers.mutateAsync({ appId: applicationId, data });
      setResult(response);
      setStep('result');
      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Import failed');
    }
  };

  const handleBack = () => {
    if (step === 'preview') {
      setStep('input');
    }
  };

  const handleClose = () => {
    onClose();
  };

  const previewUsers = parsedUsers.slice(0, 10);
  const displayDetails = showAllDetails ? result?.details || [] : (result?.details || []).slice(0, 10);
  const remainingDetails = (result?.details?.length || 0) - 10;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-card rounded-xl shadow-sm border border-border w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border flex items-center justify-between">
          <h2 className="text-xl font-semibold text-foreground">{t('apps.import.title')}</h2>
          <button
            onClick={handleClose}
            className="p-2 text-muted-foreground hover:text-foreground hover:bg-accent rounded-lg transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Step 1: Input */}
          {step === 'input' && (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  JSON {t('apps.import.placeholder').split('\n')[0]}
                </label>
                <textarea
                  value={jsonInput}
                  onChange={(e) => setJsonInput(e.target.value)}
                  placeholder={t('apps.import.placeholder')}
                  className="w-full h-96 px-3 py-2 bg-background border border-input rounded-lg text-foreground placeholder-muted-foreground font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>
              <p className="text-sm text-muted-foreground">{t('apps.import.format_hint')}</p>
              {parseError && (
                <div className="flex items-center gap-2 text-destructive text-sm">
                  <AlertCircle size={16} />
                  <span>{parseError}</span>
                </div>
              )}
              <div className="flex justify-end">
                <button
                  onClick={handleParse}
                  disabled={!jsonInput.trim()}
                  className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Upload size={18} />
                  {t('apps.import.parse')}
                </button>
              </div>
            </div>
          )}

          {/* Step 2: Preview */}
          {step === 'preview' && (
            <div className="space-y-4">
              <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
                <p className="text-sm text-foreground font-medium">
                  {t('apps.import.preview_count').replace('{count}', parsedUsers.length.toString())}
                </p>
              </div>

              <div className="bg-card rounded-lg border border-border overflow-hidden">
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-muted">
                      <tr>
                        <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">Email</th>
                        <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">Username</th>
                        <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">UUID</th>
                        <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">Password</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {previewUsers.map((user, idx) => (
                        <tr key={idx} className="hover:bg-accent transition-colors">
                          <td className="px-4 py-3 text-sm text-foreground">{user.email}</td>
                          <td className="px-4 py-3 text-sm text-muted-foreground">{user.username || '-'}</td>
                          <td className="px-4 py-3 text-sm text-muted-foreground font-mono">
                            {user.id ? user.id.substring(0, 8) + '...' : '-'}
                          </td>
                          <td className="px-4 py-3 text-sm">
                            {user.password_hash_import ? (
                              <span className="text-success">✓</span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                {parsedUsers.length > 10 && (
                  <div className="px-4 py-3 bg-muted text-sm text-muted-foreground">
                    {t('apps.import.and_more').replace('{count}', (parsedUsers.length - 10).toString())}
                  </div>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  {t('apps.import.on_conflict')}
                </label>
                <select
                  value={onConflict}
                  onChange={(e) => setOnConflict(e.target.value as 'skip' | 'update' | 'error')}
                  className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  <option value="skip">{t('apps.import.conflict_skip')}</option>
                  <option value="update">{t('apps.import.conflict_update')}</option>
                  <option value="error">{t('apps.import.conflict_error')}</option>
                </select>
              </div>

              <div className="flex items-center justify-between gap-4">
                <button
                  onClick={handleBack}
                  className="flex items-center gap-2 px-4 py-2 text-foreground hover:bg-accent rounded-lg transition-colors"
                >
                  <ChevronLeft size={18} />
                  {t('common.back')}
                </button>
                <button
                  onClick={handleImport}
                  disabled={importUsers.isPending}
                  className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors disabled:opacity-50"
                >
                  <Upload size={18} />
                  {importUsers.isPending ? t('apps.import.importing') : t('apps.import.start')}
                </button>
              </div>
            </div>
          )}

          {/* Step 3: Result */}
          {step === 'result' && result && (
            <div className="space-y-6">
              {/* Stats Cards */}
              <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
                <div className="bg-success/10 border border-success/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <CheckCircle className="text-success" size={20} />
                    <span className="text-sm font-medium text-foreground">{t('apps.import.imported')}</span>
                  </div>
                  <p className="text-3xl font-bold text-success">{result.imported}</p>
                </div>

                <div className="bg-warning/10 border border-warning/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <AlertCircle className="text-warning" size={20} />
                    <span className="text-sm font-medium text-foreground">{t('apps.import.skipped')}</span>
                  </div>
                  <p className="text-3xl font-bold text-warning">{result.skipped}</p>
                </div>

                <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <CheckCircle className="text-primary" size={20} />
                    <span className="text-sm font-medium text-foreground">{t('apps.import.updated')}</span>
                  </div>
                  <p className="text-3xl font-bold text-primary">{result.updated}</p>
                </div>

                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <AlertCircle className="text-destructive" size={20} />
                    <span className="text-sm font-medium text-foreground">{t('apps.import.errors')}</span>
                  </div>
                  <p className="text-3xl font-bold text-destructive">{result.errors}</p>
                </div>
              </div>

              {/* Details Table */}
              {result.details && result.details.length > 0 && (
                <div className="bg-card rounded-lg border border-border overflow-hidden">
                  <div className="px-4 py-3 bg-muted border-b border-border flex items-center justify-between">
                    <h3 className="text-sm font-semibold text-foreground">{t('apps.import.details')}</h3>
                    {result.details.length > 10 && (
                      <button
                        onClick={() => setShowAllDetails(!showAllDetails)}
                        className="flex items-center gap-1 text-sm text-primary hover:text-primary/80 transition-colors"
                      >
                        {showAllDetails ? (
                          <>
                            <ChevronUp size={16} />
                            {t('common.show_less') || 'Show less'}
                          </>
                        ) : (
                          <>
                            <ChevronDown size={16} />
                            {t('common.show_all') || 'Show all'}
                          </>
                        )}
                      </button>
                    )}
                  </div>
                  <div className="overflow-x-auto max-h-96 overflow-y-auto">
                    <table className="w-full">
                      <thead className="bg-muted sticky top-0">
                        <tr>
                          <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">Email</th>
                          <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">{t('common.status')}</th>
                          <th className="px-4 py-3 text-left text-xs font-semibold text-muted-foreground uppercase">Reason</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {displayDetails.map((detail: ImportDetail, idx) => (
                          <tr key={idx} className="hover:bg-accent transition-colors">
                            <td className="px-4 py-3 text-sm text-foreground">{detail.email}</td>
                            <td className="px-4 py-3">
                              <span
                                className={`inline-flex px-2 py-0.5 text-xs font-medium rounded ${
                                  detail.status === 'imported' || detail.status === 'updated'
                                    ? 'bg-success/10 text-success'
                                    : detail.status === 'skipped'
                                    ? 'bg-warning/10 text-warning'
                                    : 'bg-destructive/10 text-destructive'
                                }`}
                              >
                                {detail.status}
                              </span>
                            </td>
                            <td className="px-4 py-3 text-sm text-muted-foreground">
                              {detail.reason || '-'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              <div className="flex justify-end">
                <button
                  onClick={handleClose}
                  className="px-4 py-2 bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-medium transition-colors"
                >
                  {t('common.close') || 'Close'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default UsersImportModal;
