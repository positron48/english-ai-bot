import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import ru from './locales/ru.json'
import es from './locales/es.json'

export type SupportedLocale = 'en' | 'ru' | 'es'

// UI locales are always available regardless of the selected course.
export const AVAILABLE_LOCALES: { code: SupportedLocale; label: string }[] = [
  { code: 'ru', label: 'Русский' },
  { code: 'en', label: 'English' },
  { code: 'es', label: 'Español' },
]

const messages = {
  en,
  ru,
  es,
}

// Detect browser locale
export function detectBrowserLocale(): SupportedLocale {
  if (typeof window === 'undefined') {
    return 'en'
  }

  const browserLang = navigator.language || (navigator as any).userLanguage || 'en'
  const lang = browserLang.toLowerCase().split('-')[0]

  if (lang === 'ru') {
    return 'ru'
  }
  if (lang === 'es') {
    return 'es'
  }

  return 'en'
}

// Get locale from localStorage or detect from browser
export function getInitialLocale(): SupportedLocale {
  if (typeof window === 'undefined') {
    return 'en'
  }

  const stored = localStorage.getItem('locale') as SupportedLocale | null
  if (stored === 'en' || stored === 'ru' || stored === 'es') {
    return stored
  }

  return detectBrowserLocale()
}

// Save locale to localStorage
export function saveLocale(locale: SupportedLocale) {
  if (typeof window !== 'undefined') {
    localStorage.setItem('locale', locale)
  }
}

const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: 'en',
  messages,
  pluralRules: {
    ru: (choice: number) => {
      // Russian pluralization rules
      // Return index: 0 = 1, 21, 31, ... (ends with 1, but not 11)
      //               1 = 2, 3, 4, 22, 23, ... (ends with 2-4, but not 12-14)
      //               2 = 0, 5-20, 25-30, ... (everything else)
      
      if (choice === 0) {
        return 2 // 0 uses form 2 (карточек)
      }
      
      const mod10 = choice % 10
      const mod100 = choice % 100
      
      // Special cases: 11-14 use form 2
      if (mod100 >= 11 && mod100 <= 14) {
        return 2
      }
      
      // Cases based on last digit
      if (mod10 === 1) {
        return 0 // 1, 21, 31, ... (карточка)
      } else if (mod10 >= 2 && mod10 <= 4) {
        return 1 // 2-4, 22-24, 32-34, ... (карточки)
      } else {
        return 2 // 0, 5-20, 25-30, ... (карточек)
      }
    },
  },
})

export default i18n
