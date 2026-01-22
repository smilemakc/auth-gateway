// Theme configuration and types
export {
  type ThemeConfig,
  type ThemeColors,
  type ThemeBranding,
  type ThemeTypography,
  lightTheme,
  darkTheme,
  themePresets,
  createTheme,
  createBrandTheme,
} from './theme-config';

// Theme provider and hook
export {
  ThemeProvider,
  useTheme,
  type ThemeMode,
  type ThemeContextValue,
  type ThemeProviderProps,
} from './theme-provider';
