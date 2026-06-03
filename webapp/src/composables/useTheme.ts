import { ref } from 'vue'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'app-theme'

const currentTheme = ref<Theme>('dark')

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
  if (savedTheme) {
    currentTheme.value = savedTheme
  } else {
    currentTheme.value = 'dark'
  }
  
  applyTheme(currentTheme.value)
}

// Initialize theme immediately when module loads
if (typeof window !== 'undefined') {
  initTheme()
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
