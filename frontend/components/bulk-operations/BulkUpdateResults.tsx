import React from 'react';
import type { BulkOperationResult } from '@auth-gateway/client-sdk';
import { useLanguage } from '../../services/i18n';

interface BulkUpdateResultsProps {
  result: BulkOperationResult;
  onUpdateMore: () => void;
  onDone: () => void;
}

const BulkUpdateResults: React.FC<BulkUpdateResultsProps> = ({
  result,
  onUpdateMore,
  onDone,
}) => {
  const { t } = useLanguage();

  return (
    <div className="bg-card rounded-xl shadow-sm border border-border p-6">
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-muted rounded-lg p-4">
          <div className="text-sm text-muted-foreground">{t('bulk_update.total')}</div>
          <div className="text-2xl font-bold text-foreground">{result.total}</div>
        </div>
        <div className="bg-success/10 rounded-lg p-4">
          <div className="text-sm text-success">{t('bulk_update.success')}</div>
          <div className="text-2xl font-bold text-success">{result.success}</div>
        </div>
        <div className="bg-destructive/10 rounded-lg p-4">
          <div className="text-sm text-destructive">{t('bulk_update.failed')}</div>
          <div className="text-2xl font-bold text-destructive">{result.failed}</div>
        </div>
      </div>

      {result.errors && result.errors.length > 0 && (
        <div className="mb-4">
          <h3 className="font-semibold text-foreground mb-2">{t('bulk_update.errors')}</h3>
          <div className="space-y-2">
            {result.errors.map((error, index) => (
              <div key={index} className="bg-destructive/10 border border-border rounded-lg p-3">
                <div className="text-sm text-destructive">{error.message}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="flex justify-end gap-3 pt-4 border-t border-border">
        <button
          onClick={onUpdateMore}
          className="px-4 py-2 border border-input rounded-lg text-foreground hover:bg-accent transition-colors"
        >
          {t('bulk_update.update_more')}
        </button>
        <button
          onClick={onDone}
          className="px-4 py-2 bg-primary hover:bg-primary-600 text-primary-foreground rounded-lg transition-colors"
        >
          {t('common.done')}
        </button>
      </div>
    </div>
  );
};

export default BulkUpdateResults;
