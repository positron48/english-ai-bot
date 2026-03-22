import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'

export interface LearningPayload {
  pair: string
  native_lang: string
  target_lang: string
  app_code: string
  grammar_bundle_id: string
  target_lang_name_ru: string
  target_lang_name_en: string
}

const learning = ref<LearningPayload | null>(null)
let loadPromise: Promise<void> | null = null

const defaultLearning = (): LearningPayload => ({
  pair: 'ru-en',
  native_lang: 'ru',
  target_lang: 'en',
  app_code: 'english',
  grammar_bundle_id: 'en',
  target_lang_name_ru: 'английский',
  target_lang_name_en: 'English',
})

/** Loads learning metadata from GET /api/settings (cached). Safe to call multiple times. */
export async function ensureLearningLoaded(): Promise<void> {
  if (learning.value) return
  if (loadPromise) return loadPromise
  loadPromise = (async () => {
    try {
      const data = await apiClient.request<{ learning?: LearningPayload }>('/api/settings')
      learning.value = data.learning ?? defaultLearning()
    } catch {
      learning.value = defaultLearning()
    }
  })()
  return loadPromise
}

export function useLearningConfig() {
  const { locale } = useI18n()
  const targetLangDisplay = computed(() => {
    const l = learning.value
    if (!l) return locale.value === 'ru' ? defaultLearning().target_lang_name_ru : defaultLearning().target_lang_name_en
    return locale.value === 'ru' ? l.target_lang_name_ru : l.target_lang_name_en
  })
  return { learning, targetLangDisplay, ensureLearningLoaded }
}
