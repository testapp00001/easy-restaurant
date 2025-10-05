import { useEffect, useState } from 'react';
import {
  THEME_OCEAN,
  THEME_SUNSET,
  ThemeProviderContext,
  type CustomTheme,
  type Theme,
} from './ThemeContext';

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  defaultCustomTheme = 'default',
  storageKey = 'vite-ui-theme',
  ...props
}: {
  children: React.ReactNode;
  defaultTheme?: Theme;
  defaultCustomTheme?: CustomTheme;
  storageKey?: string;
}) {
  const [theme, setTheme] = useState<Theme>(
    () => (localStorage.getItem(storageKey) as Theme) || defaultTheme
  );
  const [customTheme, setCustomTheme] = useState<CustomTheme>(
    () =>
      (localStorage.getItem(`${storageKey}-custom`) as CustomTheme) ||
      defaultCustomTheme
  );

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove('light', 'dark');

    if (theme === 'system') {
      const systemTheme = window.matchMedia('(prefers-color-scheme: dark)')
        .matches
        ? 'dark'
        : 'light';
      root.classList.add(systemTheme);
    } else {
      root.classList.add(theme);
    }
  }, [theme]);

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove(THEME_OCEAN, THEME_SUNSET); // Remove all custom themes
    if (customTheme !== 'default') {
      root.classList.add(customTheme);
    }
  }, [customTheme]);

  const value = {
    theme,
    customTheme,
    setTheme: (newTheme: Theme) => {
      localStorage.setItem(storageKey, newTheme);
      setTheme(newTheme);
    },
    setCustomTheme: (newCustomTheme: CustomTheme) => {
      localStorage.setItem(`${storageKey}-custom`, newCustomTheme);
      setCustomTheme(newCustomTheme);
    },
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}
