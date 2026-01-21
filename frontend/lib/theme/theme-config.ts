/**
 * Auth Gateway Theme Configuration
 *
 * This module defines the theme structure for white-label customization.
 * Colors use HSL format (without hsl() wrapper) for CSS variable compatibility.
 *
 * Example HSL values: "199 89% 48%" (not "hsl(199, 89%, 48%)")
 */

export interface ThemeColors {
  /** Page background */
  background: string;
  /** Primary text color */
  foreground: string;

  /** Card/panel background */
  card: string;
  cardForeground: string;

  /** Popover/dropdown background */
  popover: string;
  popoverForeground: string;

  /** Primary brand color */
  primary: string;
  primaryForeground: string;
  /** Primary color scale for gradients/states */
  primary50: string;
  primary100: string;
  primary500: string;
  primary600: string;
  primary700: string;

  /** Secondary/neutral color */
  secondary: string;
  secondaryForeground: string;

  /** Muted/subtle elements */
  muted: string;
  mutedForeground: string;

  /** Accent color for highlights */
  accent: string;
  accentForeground: string;

  /** Error/danger color */
  destructive: string;
  destructiveForeground: string;

  /** Success color */
  success: string;
  successForeground: string;

  /** Warning color */
  warning: string;
  warningForeground: string;

  /** Border color */
  border: string;
  /** Input border color */
  input: string;
  /** Focus ring color */
  ring: string;

  /** Sidebar colors */
  sidebar: string;
  sidebarForeground: string;
  sidebarMuted: string;
  sidebarAccent: string;
  sidebarBorder: string;
}

export interface ThemeBranding {
  /** Company/product name */
  name: string;
  /** Logo URL (supports SVG, PNG, etc.) */
  logo?: string;
  /** Favicon URL */
  favicon?: string;
}

export interface ThemeTypography {
  /** Primary font family */
  fontSans: string;
  /** Monospace font family */
  fontMono: string;
}

export interface ThemeConfig {
  /** Unique theme identifier */
  id: string;
  /** Display name */
  name: string;
  /** Theme mode */
  mode: 'light' | 'dark';
  /** Branding options */
  branding: ThemeBranding;
  /** Color palette */
  colors: ThemeColors;
  /** Typography settings */
  typography: ThemeTypography;
  /** Border radius (CSS value) */
  radius: string;
}

// ============ DEFAULT THEMES ============

export const lightTheme: ThemeConfig = {
  id: 'light',
  name: 'Light',
  mode: 'light',
  branding: {
    name: 'Auth Gateway',
  },
  colors: {
    background: '0 0% 96%',
    foreground: '222 47% 11%',

    card: '0 0% 100%',
    cardForeground: '222 47% 11%',

    popover: '0 0% 100%',
    popoverForeground: '222 47% 11%',

    primary: '199 89% 48%',
    primaryForeground: '0 0% 100%',
    primary50: '204 100% 97%',
    primary100: '204 94% 94%',
    primary500: '199 89% 48%',
    primary600: '200 98% 39%',
    primary700: '201 96% 32%',

    secondary: '210 40% 96%',
    secondaryForeground: '222 47% 11%',

    muted: '210 40% 96%',
    mutedForeground: '215 16% 47%',

    accent: '210 40% 96%',
    accentForeground: '222 47% 11%',

    destructive: '0 84% 60%',
    destructiveForeground: '0 0% 100%',

    success: '142 76% 36%',
    successForeground: '0 0% 100%',

    warning: '38 92% 50%',
    warningForeground: '0 0% 100%',

    border: '214 32% 91%',
    input: '214 32% 91%',
    ring: '199 89% 48%',

    sidebar: '222 47% 11%',
    sidebarForeground: '210 40% 98%',
    sidebarMuted: '215 25% 27%',
    sidebarAccent: '217 91% 60%',
    sidebarBorder: '217 33% 17%',
  },
  typography: {
    fontSans: "'Inter', system-ui, sans-serif",
    fontMono: 'ui-monospace, monospace',
  },
  radius: '0.5rem',
};

export const darkTheme: ThemeConfig = {
  id: 'dark',
  name: 'Dark',
  mode: 'dark',
  branding: {
    name: 'Auth Gateway',
  },
  colors: {
    background: '222 47% 11%',
    foreground: '210 40% 98%',

    card: '217 33% 17%',
    cardForeground: '210 40% 98%',

    popover: '217 33% 17%',
    popoverForeground: '210 40% 98%',

    primary: '199 89% 48%',
    primaryForeground: '0 0% 100%',
    primary50: '204 100% 12%',
    primary100: '204 94% 18%',
    primary500: '199 89% 48%',
    primary600: '200 98% 55%',
    primary700: '201 96% 62%',

    secondary: '217 33% 17%',
    secondaryForeground: '210 40% 98%',

    muted: '217 33% 17%',
    mutedForeground: '215 20% 65%',

    accent: '217 33% 17%',
    accentForeground: '210 40% 98%',

    destructive: '0 62% 55%',
    destructiveForeground: '0 0% 100%',

    success: '142 71% 45%',
    successForeground: '0 0% 100%',

    warning: '38 92% 55%',
    warningForeground: '0 0% 0%',

    border: '217 33% 25%',
    input: '217 33% 25%',
    ring: '199 89% 48%',

    sidebar: '222 47% 8%',
    sidebarForeground: '210 40% 98%',
    sidebarMuted: '215 25% 20%',
    sidebarAccent: '217 91% 60%',
    sidebarBorder: '217 33% 12%',
  },
  typography: {
    fontSans: "'Inter', system-ui, sans-serif",
    fontMono: 'ui-monospace, monospace',
  },
  radius: '0.5rem',
};

// ============ THEME PRESETS ============

export const themePresets: Record<string, { light: ThemeConfig; dark: ThemeConfig }> = {
  default: {
    light: lightTheme,
    dark: darkTheme,
  },
};

// ============ HELPERS ============

/**
 * Creates a custom theme by merging with defaults
 */
export function createTheme(
  overrides: Partial<ThemeConfig> & { colors?: Partial<ThemeColors> },
  base: ThemeConfig = lightTheme
): ThemeConfig {
  return {
    ...base,
    ...overrides,
    colors: {
      ...base.colors,
      ...overrides.colors,
    },
    branding: {
      ...base.branding,
      ...overrides.branding,
    },
    typography: {
      ...base.typography,
      ...overrides.typography,
    },
  };
}

/**
 * Creates a theme pair (light + dark) for a brand
 */
export function createBrandTheme(options: {
  id: string;
  name: string;
  branding: ThemeBranding;
  primaryColor: string;
  primaryScale?: {
    50?: string;
    100?: string;
    500?: string;
    600?: string;
    700?: string;
  };
}): { light: ThemeConfig; dark: ThemeConfig } {
  const { id, name, branding, primaryColor, primaryScale } = options;

  const lightColors: Partial<ThemeColors> = {
    primary: primaryColor,
    ring: primaryColor,
    ...(primaryScale?.['50'] && { primary50: primaryScale['50'] }),
    ...(primaryScale?.['100'] && { primary100: primaryScale['100'] }),
    ...(primaryScale?.['500'] && { primary500: primaryScale['500'] }),
    ...(primaryScale?.['600'] && { primary600: primaryScale['600'] }),
    ...(primaryScale?.['700'] && { primary700: primaryScale['700'] }),
  };

  return {
    light: createTheme(
      { id: `${id}-light`, name: `${name} Light`, mode: 'light', branding, colors: lightColors },
      lightTheme
    ),
    dark: createTheme(
      { id: `${id}-dark`, name: `${name} Dark`, mode: 'dark', branding, colors: lightColors },
      darkTheme
    ),
  };
}
