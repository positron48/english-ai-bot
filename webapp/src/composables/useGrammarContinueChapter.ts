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

  const loadContinueChapter = async () => {
    loading.value = true
    try {
      const data = await grammarClient.getContinueChapter()
      const chapter = data?.chapter
        ? toGrammarContinueChapter(data.chapter, locale.value)
        : null
      continueChapter.value = chapter
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
  }
}
