import React, { useSyncExternalStore, useEffect, useRef } from 'react';
import { AlertTriangle, HelpCircle } from 'lucide-react';
import { confirmStore, resolveConfirm } from '../services/confirm';
import { useLanguage } from '../services/i18n';

export const ConfirmDialog: React.FC = () => {
  const state = useSyncExternalStore(confirmStore.subscribe, confirmStore.getSnapshot);
  const { t } = useLanguage();
  const cancelRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!state) return;

    // Focus cancel button on open
    cancelRef.current?.focus();

    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        resolveConfirm(false);
      }
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [state]);

  if (!state) return null;

  const { options } = state;
  const isDanger = options.variant === 'danger';

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in duration-150"
      onClick={() => resolveConfirm(false)}
    >
      <div
        className="bg-card rounded-xl shadow-xl border border-border w-full max-w-md animate-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-6">
          <div className="flex items-start gap-4">
            <div className={`p-2 rounded-full ${isDanger ? 'bg-destructive/10' : 'bg-primary/10'}`}>
              {isDanger
                ? <AlertTriangle size={24} className="text-destructive" />
                : <HelpCircle size={24} className="text-primary" />
              }
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-lg font-semibold text-foreground">
                {options.title || t('confirm.title')}
              </h3>
              <p className="mt-2 text-sm text-muted-foreground">
                {options.description}
              </p>
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-border">
          <button
            ref={cancelRef}
            onClick={() => resolveConfirm(false)}
            className="px-4 py-2 text-sm font-medium text-foreground border border-input rounded-lg hover:bg-accent transition-colors"
          >
            {options.cancelText || t('confirm.cancel')}
          </button>
          <button
            onClick={() => resolveConfirm(true)}
            className={`px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
              isDanger
                ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            }`}
          >
            {options.confirmText || (isDanger ? t('confirm.delete') : t('confirm.confirm'))}
          </button>
        </div>
      </div>
    </div>
  );
};
