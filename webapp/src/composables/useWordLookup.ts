import { onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { withCourseCode } from '../api/grammarClient'
import type { VocabCardsAPIResponse } from '../types/vocab'

const LOOKUP_GENERATING_DELAY_MS = 600

const LOOKUP_CANDIDATE_RE = /^[\p{L}\p{M}'’-]+( [\p{L}\p{M}'’-]+)?$/u

export function isWordLookupCandidate(text: string): boolean {
  return LOOKUP_CANDIDATE_RE.test(text.trim())
}

export function useWordLookup() {
  const { t } = useI18n()

  const modalVisible = ref(false)
  const loading = ref(false)
  const generating = ref(false)
  const error = ref('')
  const lemma = ref('')
  const preloaded = ref<VocabCardsAPIResponse | null>(null)

  let generatingTimer: ReturnType<typeof setTimeout> | null = null

  const clearGeneratingTimer = () => {
    if (generatingTimer) {
      clearTimeout(generatingTimer)
      generatingTimer = null
    }
  }

  const closeModal = () => {
    modalVisible.value = false
    error.value = ''
  }

  const modalKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') closeModal()
  }

  const bindModalKeydown = (open: boolean) => {
    if (open) window.addEventListener('keydown', modalKeydown)
    else window.removeEventListener('keydown', modalKeydown)
  }

  onUnmounted(() => {
    window.removeEventListener('keydown', modalKeydown)
    clearGeneratingTimer()
  })

  const mapLookupError = (err: unknown): string => {
    const status = typeof (err as { status?: number })?.status === 'number'
      ? (err as { status: number }).status
      : 0
    if ((err as { isNetworkError?: boolean })?.isNetworkError) {
      return t('reading.wordLookupNetwork')
    }
    if (status === 404) return t('reading.wordNotFound')
    if (status >= 500) return t('reading.wordLookupServerError')
    return t('reading.wordLookupFailed')
  }

  const lookup = async (word: string) => {
    const query = word.trim()
    if (!query || !isWordLookupCandidate(query) || loading.value) return

    modalVisible.value = true
    loading.value = true
    generating.value = false
    error.value = ''
    preloaded.value = null
    lemma.value = query

    clearGeneratingTimer()
    generatingTimer = setTimeout(() => {
      generating.value = true
    }, LOOKUP_GENERATING_DELAY_MS)

    try {
      const data: VocabCardsAPIResponse = await apiClient.request(
        withCourseCode(`/api/reading/word-lookup?lemma=${encodeURIComponent(query)}`),
      )
      lemma.value = data.lemma || query
      preloaded.value = data
    } catch (err) {
      console.error('Word lookup failed', err)
      preloaded.value = null
      error.value = mapLookupError(err)
    } finally {
      clearGeneratingTimer()
      loading.value = false
      generating.value = false
    }
  }

  return {
    modalVisible,
    loading,
    generating,
    error,
    lemma,
    preloaded,
    lookup,
    closeModal,
    bindModalKeydown,
    isWordLookupCandidate,
  }
}
