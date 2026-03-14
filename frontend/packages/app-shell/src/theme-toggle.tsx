import { Button, MoonIcon, SunIcon } from '@nomnomvault/ui'

import { useTheme } from './theme'

export function ThemeToggle() {
  const theme = useTheme()

  return (
    <Button
      variant="secondary"
      size="icon"
      class="shrink-0"
      aria-label="Toggle theme"
      onClick={() => theme.toggleTheme()}
    >
      {theme.theme() === 'dark' ? <SunIcon size="md" /> : <MoonIcon size="md" />}
    </Button>
  )
}
