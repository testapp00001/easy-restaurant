import { createContext, useContext } from 'react';

export const THEME_DEFAULT = 'default';
export const THEME_OCEAN = 'theme-ocean';
export const THEME_SUNSET = 'theme-sunset';

export type Theme = 'dark' | 'light' | 'system';
export type CustomTheme = 'default' | 'theme-ocean' | 'theme-sunset';

export type ThemeProviderState = {
  theme: Theme;
  customTheme: CustomTheme;
  setTheme: (theme: Theme) => void;
  setCustomTheme: (theme: CustomTheme) => void;
};

const initialState: ThemeProviderState = {
  theme: 'system',
  customTheme: 'default',
  setTheme: () => null,
  setCustomTheme: () => null,
};

export const ThemeProviderContext =
  createContext<ThemeProviderState>(initialState);

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);
  if (context === undefined)
    throw new Error('useTheme must be used within a ThemeProvider');
  return context;
};
