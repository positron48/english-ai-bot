import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { grammarClient } from '../api/grammarClient'
import {
  type GrammarContinueChapter,
  toGrammarContinueChapter,
} from '../utils/grammarContinueChapter'

export function useGrammarContinueChapter() {
  const { locale } = useI18n()
  const continueChapter = ref<GrammarContinueChapter | null>(null)
  const loading = ref(false)

  // applyContinueChapter hydrates state from an already-fetched payload (same shape as
  // grammarClient.getContinueChapter returns), used by the aggregate overview endpoints so the
  // continue-chapter section doesn't need its own network request.
  const applyContinueChapter = (data: { chapter?: unknown } | null | undefined) => {
    continueChapter.value = data?.chapter
      ? toGrammarContinueChapter(data.chapter as any, locale.value)
      : null
  }

  const loadContinueChapter = async () => {
    loading.value = true
    try {
      const data = await grammarClient.getContinueChapter()
      applyContinueChapter(data)
    } catch {
      continueChapter.value = null
    } finally {
      loading.value = false
    }
  }

  return {
    continueChapter,
    loading,
    loadContinueChapter,
    applyContinueChapter,
  }
}
