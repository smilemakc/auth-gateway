import React, { useSyncExternalStore } from 'react';
import { CheckCircle2, XCircle, AlertTriangle, Info, X } from 'lucide-react';
import { toast as toastActions, toastStore, type Toast } from '../services/toast';

const icons: Record<Toast['type'], React.ReactNode> = {
  success: <CheckCircle2 size={20} className="text-emerald-500 shrink-0" />,
  error: <XCircle size={20} className="text-red-500 shrink-0" />,
  warning: <AlertTriangle size={20} className="text-amber-500 shrink-0" />,
  info: <Info size={20} className="text-blue-500 shrink-0" />,
};

const bgClasses: Record<Toast['type'], string> = {
  success: 'border-emerald-200 dark:border-emerald-800',
  error: 'border-red-200 dark:border-red-800',
  warning: 'border-amber-200 dark:border-amber-800',
  info: 'border-blue-200 dark:border-blue-800',
};

export const ToastContainer: React.FC = () => {
  const toasts = useSyncExternalStore(toastStore.subscribe, toastStore.getSnapshot);

  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`pointer-events-auto flex items-start gap-3 bg-card border ${bgClasses[t.type]} rounded-lg shadow-lg p-4 animate-in slide-in-from-right-5 fade-in duration-200`}
        >
          {icons[t.type]}
          <p className="text-sm text-foreground flex-1 break-words">{t.message}</p>
          <button
            onClick={() => toastActions.dismiss(t.id)}
            className="text-muted-foreground hover:text-foreground shrink-0"
          >
            <X size={16} />
          </button>
        </div>
      ))}
    </div>
  );
};
