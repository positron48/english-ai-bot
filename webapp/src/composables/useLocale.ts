import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SupportedLocale } from '../i18n'
import { detectBrowserLocale, getInitialLocale, saveLocale } from '../i18n'
import { apiClient } from '../api/client'

const currentLocale = ref<SupportedLocale>(getInitialLocale())

export function useLocale() {
  const { locale } = useI18n()

  const setLocale = async (newLocale: SupportedLocale) => {
    currentLocale.value = newLocale
    locale.value = newLocale
    saveLocale(newLocale)
    
    // Save language preference to backend if authenticated
    try {
      await apiClient.request('/api/settings/language', {
        method: 'POST',
        body: JSON.stringify({ language: newLocale }),
      })
    } catch (error) {
      // Silently fail - user might not be authenticated yet
      console.debug('Failed to save language preference:', error)
    }
  }

  const detectBrowserLocaleAndSet = () => {
    const detected = detectBrowserLocale()
    setLocale(detected)
    return detected
  }

  // Sync locale changes to reactive ref
  watch(() => locale.value, (newLocale) => {
    if (newLocale === 'en' || newLocale === 'ru' || newLocale === 'es') {
      currentLocale.value = newLocale
      saveLocale(newLocale)
    }
  })

  return {
    currentLocale,
    setLocale,
    detectBrowserLocale: detectBrowserLocaleAndSet,
  }
}
