import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'

export interface SpanishVerbScopeLadderStep {
  scope: string
  label_ru: string
  label_en: string
}

export interface LearningPayload {
  pair: string
  native_lang: string
  target_lang: string
  app_code: string
  grammar_bundle_id: string
  target_lang_name_ru: string
  target_lang_name_en: string
  /** From server: true only when target_lang is es and verb-forms feature is enabled */
  spanish_verb_forms_enabled: boolean
  /** From GET /api/settings learning.* when Spanish verb forms enabled */
  spanish_verb_scope_ladder?: SpanishVerbScopeLadderStep[]
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
  spanish_verb_forms_enabled: false,
  spanish_verb_scope_ladder: undefined,
})

function learningFromHealthPayload(raw: unknown): LearningPayload | null {
  if (!raw || typeof raw !== 'object') return null
  const l = raw as Record<string, unknown>
  if (typeof l.target_lang !== 'string' || typeof l.native_lang !== 'string') return null
  const d = defaultLearning()
  return {
    pair: typeof l.pair === 'string' ? l.pair : d.pair,
    native_lang: l.native_lang,
    target_lang: l.target_lang,
    app_code: typeof l.app_code === 'string' ? l.app_code : d.app_code,
    grammar_bundle_id: typeof l.grammar_bundle_id === 'string' ? l.grammar_bundle_id : d.grammar_bundle_id,
    target_lang_name_ru: typeof l.target_lang_name_ru === 'string' ? l.target_lang_name_ru : d.target_lang_name_ru,
    target_lang_name_en: typeof l.target_lang_name_en === 'string' ? l.target_lang_name_en : d.target_lang_name_en,
    spanish_verb_forms_enabled:
      typeof l.spanish_verb_forms_enabled === 'boolean' ? l.spanish_verb_forms_enabled : false,
    spanish_verb_scope_ladder: Array.isArray(l.spanish_verb_scope_ladder)
      ? (l.spanish_verb_scope_ladder as SpanishVerbScopeLadderStep[])
      : undefined,
  }
}

/** Loads learning metadata from GET /health (no auth) then GET /api/settings (cached). Safe to call multiple times. */
export async function ensureLearningLoaded(): Promise<void> {
  if (!loadPromise) {
    loadPromise = (async () => {
      try {
        const hr = await fetch('/health')
        if (hr.ok) {
          const hj = await hr.json()
          const pub = learningFromHealthPayload(hj?.learning)
          if (pub) {
            learning.value = pub
          }
        }
      } catch {
        /* ignore */
      }
      try {
        const data = await apiClient.request<{ learning?: Partial<LearningPayload> }>('/api/settings')
        const patch = data.learning
        const base = learning.value ?? defaultLearning()
        learning.value = {
          ...base,
          ...(patch ?? {}),
          spanish_verb_forms_enabled:
            typeof patch?.spanish_verb_forms_enabled === 'boolean'
              ? patch.spanish_verb_forms_enabled
              : base.spanish_verb_forms_enabled,
          spanish_verb_scope_ladder: Array.isArray(patch?.spanish_verb_scope_ladder)
            ? (patch.spanish_verb_scope_ladder as SpanishVerbScopeLadderStep[])
            : base.spanish_verb_scope_ladder,
        }
      } catch {
        if (!learning.value) {
          learning.value = defaultLearning()
        }
      }
    })()
  }
  await loadPromise
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
