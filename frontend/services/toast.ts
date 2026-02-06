type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  type: ToastType;
  message: string;
  duration: number;
}

type Listener = () => void;

let toasts: Toast[] = [];
let listeners: Listener[] = [];
let nextId = 0;

function emit() {
  for (const l of listeners) l();
}

function add(type: ToastType, message: string, duration = 4000) {
  const id = String(++nextId);
  toasts = [...toasts, { id, type, message, duration }];
  emit();
  if (duration > 0) {
    setTimeout(() => remove(id), duration);
  }
}

function remove(id: string) {
  toasts = toasts.filter((t) => t.id !== id);
  emit();
}

export const toast = {
  success: (message: string, duration?: number) => add('success', message, duration),
  error: (message: string, duration?: number) => add('error', message, duration ?? 6000),
  warning: (message: string, duration?: number) => add('warning', message, duration),
  info: (message: string, duration?: number) => add('info', message, duration),
  dismiss: remove,
};

// useSyncExternalStore contract
export const toastStore = {
  subscribe: (listener: Listener) => {
    listeners = [...listeners, listener];
    return () => {
      listeners = listeners.filter((l) => l !== listener);
    };
  },
  getSnapshot: () => toasts,
};
