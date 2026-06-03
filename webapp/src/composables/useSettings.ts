import { ref, watch } from 'vue'

const SETTINGS_STORAGE_KEY = 'app-settings'

interface Settings {
  soundsEnabled: boolean
  vibrationEnabled: boolean
  theme: 'light' | 'dark'
  soundTheme: string
  hideMorphInTraining: boolean
  autoplayPronunciation: boolean
}

const defaultSettings: Settings = {
  soundsEnabled: true,
  vibrationEnabled: true,
  theme: 'dark',
  soundTheme: 'tick',
  hideMorphInTraining: false,
  autoplayPronunciation: true
}

const currentSettings = ref<Settings>({ ...defaultSettings })

// Load settings from localStorage
const loadSettings = () => {
  if (typeof window === 'undefined') return
  
  try {
    const saved = localStorage.getItem(SETTINGS_STORAGE_KEY)
    if (saved) {
      const parsed = JSON.parse(saved)
      currentSettings.value = {
        soundsEnabled: parsed.soundsEnabled !== undefined ? parsed.soundsEnabled : defaultSettings.soundsEnabled,
        vibrationEnabled: parsed.vibrationEnabled !== undefined ? parsed.vibrationEnabled : defaultSettings.vibrationEnabled,
        theme: parsed.theme || defaultSettings.theme,
        soundTheme: parsed.soundTheme || defaultSettings.soundTheme,
        hideMorphInTraining: parsed.hideMorphInTraining !== undefined ? parsed.hideMorphInTraining : defaultSettings.hideMorphInTraining,
        autoplayPronunciation: parsed.autoplayPronunciation !== undefined ? parsed.autoplayPronunciation : defaultSettings.autoplayPronunciation
      }
    } else {
      // If no saved settings, try to get theme from useTheme's storage
      const themeFromStorage = localStorage.getItem('app-theme') as 'light' | 'dark' | null
      if (themeFromStorage) {
        currentSettings.value.theme = themeFromStorage
      }
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
    currentSettings.value = { ...defaultSettings }
  }
}

// Save settings to localStorage
const saveSettings = () => {
  if (typeof window === 'undefined') return
  
  try {
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(currentSettings.value))
  } catch (e) {
    console.error('Failed to save settings:', e)
  }
}

// Initialize settings on module load
if (typeof window !== 'undefined') {
  loadSettings()
  
  // Watch for changes and save automatically
  watch(currentSettings, () => {
    saveSettings()
  }, { deep: true })
}

export function useSettings() {
  const setSoundsEnabled = (enabled: boolean) => {
    currentSettings.value.soundsEnabled = enabled
  }

  const setVibrationEnabled = (enabled: boolean) => {
    currentSettings.value.vibrationEnabled = enabled
  }

  const setTheme = (theme: 'light' | 'dark') => {
    currentSettings.value.theme = theme
    // Also update theme storage for useTheme compatibility
    localStorage.setItem('app-theme', theme)
  }

  const setSoundTheme = (theme: string) => {
    currentSettings.value.soundTheme = theme
  }

  const setHideMorphInTraining = (hide: boolean) => {
    currentSettings.value.hideMorphInTraining = hide
  }
  const setAutoplayPronunciation = (enabled: boolean) => {
    currentSettings.value.autoplayPronunciation = enabled
  }

  return {
    settings: currentSettings,
    setSoundsEnabled,
    setVibrationEnabled,
    setTheme,
    setSoundTheme,
    setHideMorphInTraining,
    setAutoplayPronunciation,
    loadSettings,
    saveSettings
  }
}
