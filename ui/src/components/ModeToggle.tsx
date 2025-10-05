import { Moon, Sun } from 'lucide-react';

import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  THEME_DEFAULT,
  THEME_OCEAN,
  THEME_SUNSET,
  useTheme,
} from '@/contexts/ThemeContext';

export function ModeToggle() {
  const { setTheme, setCustomTheme } = useTheme();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="icon">
          <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
          <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
          <span className="sr-only">Toggle theme</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => setTheme('light')}>
          Light
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme('dark')}>
          Dark
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setTheme('system')}>
          System
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setCustomTheme(THEME_DEFAULT)}>
          Default Theme
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setCustomTheme(THEME_OCEAN)}>
          Ocean Theme
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => setCustomTheme(THEME_SUNSET)}>
          Sunset Theme
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
