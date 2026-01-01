import { ref } from 'vue'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'app-theme'

const currentTheme = ref<Theme>('light')

const applyTheme = (theme: Theme) => {
  const root = document.documentElement
  if (theme === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }
}

const initTheme = () => {
  if (typeof window === 'undefined') return
  
  const savedTheme = localStorage.getItem(THEME_STORAGE_KEY) as Theme | null
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  
  if (savedTheme) {
    currentTheme.value = savedTheme
  } else if (prefersDark) {
    currentTheme.value = 'dark'
  } else {
    currentTheme.value = 'light'
  }
  
  applyTheme(currentTheme.value)
}

// Initialize theme immediately when module loads
if (typeof window !== 'undefined') {
  initTheme()
  
  // Watch for system theme changes
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  const handleChange = (e: MediaQueryListEvent) => {
    // Only auto-switch if user hasn't manually set a preference
    if (!localStorage.getItem(THEME_STORAGE_KEY)) {
      currentTheme.value = e.matches ? 'dark' : 'light'
      applyTheme(currentTheme.value)
    }
  }
  
  mediaQuery.addEventListener('change', handleChange)
}

export function useTheme() {
  const toggleTheme = () => {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem(THEME_STORAGE_KEY, currentTheme.value)
    applyTheme(currentTheme.value)
  }

  const setTheme = (theme: Theme) => {
    currentTheme.value = theme
    localStorage.setItem(THEME_STORAGE_KEY, theme)
    applyTheme(theme)
  }

  return {
    theme: currentTheme,
    toggleTheme,
    setTheme,
    isDark: () => currentTheme.value === 'dark',
    isLight: () => currentTheme.value === 'light'
  }
}
