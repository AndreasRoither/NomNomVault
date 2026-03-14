import type { Accessor, JSX } from 'solid-js'
import {
  createContext,
  createEffect,
  createSignal,
  onMount,
  useContext,
} from 'solid-js'

type ThemeMode = 'light' | 'dark'

type ThemeContextValue = {
  theme: Accessor<ThemeMode>
  setTheme: (value: ThemeMode) => void
  toggleTheme: () => void
}

const STORAGE_KEY = 'nnv-theme'

const ThemeContext = createContext<ThemeContextValue>()

function getPreferredTheme(): ThemeMode {
  if (typeof window === 'undefined') {
    return 'light'
  }

  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') {
    return stored
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

function applyTheme(theme: ThemeMode) {
  if (typeof document === 'undefined') {
    return
  }

  document.documentElement.dataset.theme = theme
}

export function ThemeScript() {
  const script = `
    (function () {
      var key = '${STORAGE_KEY}';
      var theme = localStorage.getItem(key);
      if (theme !== 'light' && theme !== 'dark') {
        theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
      }
      document.documentElement.dataset.theme = theme;
    })();
  `

  return <script innerHTML={script} />
}

export function ThemeProvider(props: { children: JSX.Element }) {
  const [theme, setTheme] = createSignal<ThemeMode>('light')

  onMount(() => {
    setTheme(getPreferredTheme())
  })

  createEffect(() => {
    const value = theme()
    applyTheme(value)

    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, value)
    }
  })

  const context: ThemeContextValue = {
    theme,
    setTheme,
    toggleTheme: () => setTheme(theme() === 'dark' ? 'light' : 'dark'),
  }

  return (
    <ThemeContext.Provider value={context}>
      {props.children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const context = useContext(ThemeContext)

  if (!context) {
    throw new Error('useTheme must be used inside ThemeProvider')
  }

  return context
}
