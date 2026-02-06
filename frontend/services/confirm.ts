export interface ConfirmOptions {
  title?: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'danger' | 'default';
}

export interface ConfirmState {
  options: ConfirmOptions;
  resolve: (value: boolean) => void;
}

type Listener = () => void;

let current: ConfirmState | null = null;
let listeners: Listener[] = [];

function emit() {
  for (const l of listeners) l();
}

export function confirm(options: ConfirmOptions): Promise<boolean> {
  return new Promise<boolean>((resolve) => {
    current = { options, resolve };
    emit();
  });
}

export function resolveConfirm(value: boolean) {
  if (current) {
    current.resolve(value);
    current = null;
    emit();
  }
}

// useSyncExternalStore contract
export const confirmStore = {
  subscribe: (listener: Listener) => {
    listeners = [...listeners, listener];
    return () => {
      listeners = listeners.filter((l) => l !== listener);
    };
  },
  getSnapshot: () => current,
};
