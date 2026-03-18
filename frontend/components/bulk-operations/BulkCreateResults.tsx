import React from 'react';
import { CheckCircle, XCircle } from 'lucide-react';
import type { BulkOperationResult } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface BulkCreateResultsProps {
  result: BulkOperationResult;
  onCreateMore: () => void;
  onDone: () => void;
}

const BulkCreateResults: React.FC<BulkCreateResultsProps> = ({
  result,
  onCreateMore,
  onDone,
}) => {
  const { t } = useLanguage();

  return (
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
                    <span className="font-medium">{t('bulk.row')} {error.index + 1}:</span> {error.message}
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
            onClick={onCreateMore}
            className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
          >
            {t('bulk.create_more')}
          </button>
          <button
            onClick={onDone}
            className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
          >
            {t('common.done')}
          </button>
        </div>
      </div>
    </div>
  );
};

export default BulkCreateResults;
