import React, { createContext, useCallback, useEffect, useMemo, useState } from 'react';
import {
  ThemeConfig,
  ThemeColors,
  lightTheme,
  darkTheme,
} from './theme-config';

// ============ TYPES ============

export type ThemeMode = 'light' | 'dark' | 'system';

export interface ThemeContextValue {
  /** Current resolved theme (after system preference resolution) */
  theme: ThemeConfig;
  /** Current mode setting (light/dark/system) */
  mode: ThemeMode;
  /** Resolved mode (always light or dark, never system) */
  resolvedMode: 'light' | 'dark';
  /** Whether dark mode is active */
  isDark: boolean;
  /** Set theme mode */
  setMode: (mode: ThemeMode) => void;
  /** Toggle between light and dark */
  toggleMode: () => void;
  /** Apply a custom theme */
  setTheme: (theme: ThemeConfig) => void;
  /** Apply custom colors (partial update) */
  setColors: (colors: Partial<ThemeColors>) => void;
  /** Reset to default theme */
  resetTheme: () => void;
  /** Light theme config */
  lightTheme: ThemeConfig;
  /** Dark theme config */
  darkTheme: ThemeConfig;
}

// ============ CONTEXT ============

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

// ============ CSS VARIABLE MAPPING ============

const colorVarMapping: Record<keyof ThemeColors, string> = {
  background: '--background',
  foreground: '--foreground',
  card: '--card',
  cardForeground: '--card-foreground',
  popover: '--popover',
  popoverForeground: '--popover-foreground',
  primary: '--primary',
  primaryForeground: '--primary-foreground',
  primary50: '--primary-50',
  primary100: '--primary-100',
  primary500: '--primary-500',
  primary600: '--primary-600',
  primary700: '--primary-700',
  secondary: '--secondary',
  secondaryForeground: '--secondary-foreground',
  muted: '--muted',
  mutedForeground: '--muted-foreground',
  accent: '--accent',
  accentForeground: '--accent-foreground',
  destructive: '--destructive',
  destructiveForeground: '--destructive-foreground',
  success: '--success',
  successForeground: '--success-foreground',
  warning: '--warning',
  warningForeground: '--warning-foreground',
  border: '--border',
  input: '--input',
  ring: '--ring',
  sidebar: '--sidebar',
  sidebarForeground: '--sidebar-foreground',
  sidebarMuted: '--sidebar-muted',
  sidebarAccent: '--sidebar-accent',
  sidebarBorder: '--sidebar-border',
};

// ============ HELPERS ============

function applyThemeToDOM(theme: ThemeConfig) {
  const root = document.documentElement;

  // Apply colors as CSS variables
  Object.entries(theme.colors).forEach(([key, value]) => {
    const varName = colorVarMapping[key as keyof ThemeColors];
    if (varName) {
      root.style.setProperty(varName, value);
    }
  });

  // Apply typography
  root.style.setProperty('--font-sans', theme.typography.fontSans);
  root.style.setProperty('--font-mono', theme.typography.fontMono);

  // Apply radius
  root.style.setProperty('--radius', theme.radius);

  // Set data attribute for CSS selectors
  root.setAttribute('data-theme', theme.mode);

  // Update document title if branding name provided
  if (theme.branding.name) {
    document.title = `${theme.branding.name} Admin`;
  }

  // Update favicon if provided
  if (theme.branding.favicon) {
    let favicon = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    if (!favicon) {
      favicon = document.createElement('link');
      favicon.rel = 'icon';
      document.head.appendChild(favicon);
    }
    favicon.href = theme.branding.favicon;
  }
}

function getSystemPreference(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

// ============ PROVIDER ============

export interface ThemeProviderProps {
  children: React.ReactNode;
  /** Default theme mode */
  defaultMode?: ThemeMode;
  /** Custom light theme */
  lightTheme?: ThemeConfig;
  /** Custom dark theme */
  darkTheme?: ThemeConfig;
  /** Storage key for persisting mode */
  storageKey?: string;
  /** Disable persistence */
  disableStorage?: boolean;
}

export function ThemeProvider({
  children,
  defaultMode = 'system',
  lightTheme: customLightTheme,
  darkTheme: customDarkTheme,
  storageKey = 'auth-gateway-theme-mode',
  disableStorage = false,
}: ThemeProviderProps) {
  // Theme configs
  const [lightConfig, setLightConfig] = useState<ThemeConfig>(
    customLightTheme || lightTheme
  );
  const [darkConfig, setDarkConfig] = useState<ThemeConfig>(
    customDarkTheme || darkTheme
  );

  // Mode state
  const [mode, setModeState] = useState<ThemeMode>(() => {
    if (disableStorage || typeof window === 'undefined') {
      return defaultMode;
    }
    const stored = localStorage.getItem(storageKey);
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored;
    }
    return defaultMode;
  });

  // System preference tracking
  const [systemPreference, setSystemPreference] = useState<'light' | 'dark'>(
    getSystemPreference
  );

  // Listen for system preference changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    const handleChange = (e: MediaQueryListEvent) => {
      setSystemPreference(e.matches ? 'dark' : 'light');
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  // Resolved mode (never 'system')
  const resolvedMode = mode === 'system' ? systemPreference : mode;

  // Current theme based on resolved mode
  const theme = resolvedMode === 'dark' ? darkConfig : lightConfig;

  // Apply theme to DOM when it changes
  useEffect(() => {
    applyThemeToDOM(theme);
  }, [theme]);

  // Persist mode to localStorage
  useEffect(() => {
    if (!disableStorage) {
      localStorage.setItem(storageKey, mode);
    }
  }, [mode, storageKey, disableStorage]);

  // Actions
  const setMode = useCallback((newMode: ThemeMode) => {
    setModeState(newMode);
  }, []);

  const toggleMode = useCallback(() => {
    setModeState((current) => {
      if (current === 'system') {
        return systemPreference === 'dark' ? 'light' : 'dark';
      }
      return current === 'dark' ? 'light' : 'dark';
    });
  }, [systemPreference]);

  const setTheme = useCallback((newTheme: ThemeConfig) => {
    if (newTheme.mode === 'dark') {
      setDarkConfig(newTheme);
    } else {
      setLightConfig(newTheme);
    }
  }, []);

  const setColors = useCallback(
    (colors: Partial<ThemeColors>) => {
      if (resolvedMode === 'dark') {
        setDarkConfig((prev) => ({
          ...prev,
          colors: { ...prev.colors, ...colors },
        }));
      } else {
        setLightConfig((prev) => ({
          ...prev,
          colors: { ...prev.colors, ...colors },
        }));
      }
    },
    [resolvedMode]
  );

  const resetTheme = useCallback(() => {
    setLightConfig(customLightTheme || lightTheme);
    setDarkConfig(customDarkTheme || darkTheme);
    setModeState(defaultMode);
  }, [customLightTheme, customDarkTheme, defaultMode]);

  // Context value
  const value = useMemo<ThemeContextValue>(
    () => ({
      theme,
      mode,
      resolvedMode,
      isDark: resolvedMode === 'dark',
      setMode,
      toggleMode,
      setTheme,
      setColors,
      resetTheme,
      lightTheme: lightConfig,
      darkTheme: darkConfig,
    }),
    [
      theme,
      mode,
      resolvedMode,
      setMode,
      toggleMode,
      setTheme,
      setColors,
      resetTheme,
      lightConfig,
      darkConfig,
    ]
  );

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
}

// ============ HOOK ============

export function useTheme(): ThemeContextValue {
  const context = React.useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}

// ============ EXPORTS ============

export { ThemeContext };
