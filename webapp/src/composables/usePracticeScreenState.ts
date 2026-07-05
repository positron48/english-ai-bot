import { ref } from 'vue'

const ME_CACHE_KEY = 'me_profile_cache_v1'

export type VocabSummaryPayload = {
  total?: number
  new?: number
  learning?: number
  review?: number
  review_count?: number
  mastered?: number
  mastered_count?: number
}

const vocabStats = ref({
  total: 0,
  newCount: 0,
  learningCount: 0,
  reviewCount: 0,
  masteredCount: 0,
})
const vocabStatsLoaded = ref(false)
const sentenceAvailable = ref(false)
const conversationPro = ref(false)
const picturePro = ref(false)

export function syncPracticeProGatesFromMeCache(): void {
  if (typeof localStorage === 'undefined') return
  try {
    const raw = localStorage.getItem(ME_CACHE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as { data?: { features?: Record<string, boolean> } }
    const features = parsed?.data?.features
    if (!features) return
    conversationPro.value = !!features.conversation
    picturePro.value = !!features.picture_description
  } catch { /* ignore */ }
}

export function applyPracticeProGates(features?: Record<string, boolean> | null): void {
  if (!features) return
  conversationPro.value = !!features.conversation
  picturePro.value = !!features.picture_description
}

export function usePracticeScreenState() {
  function applyVocabSummary(raw?: VocabSummaryPayload | null) {
    if (!raw) return
    vocabStats.value = {
      total: raw.total ?? 0,
      newCount: raw.new ?? 0,
      learningCount: raw.learning ?? 0,
      reviewCount: raw.review ?? raw.review_count ?? 0,
      masteredCount: raw.mastered ?? raw.mastered_count ?? 0,
    }
    vocabStatsLoaded.value = true
  }

  function applySentenceToday(today: { available?: boolean; remaining?: number } | null | undefined) {
    sentenceAvailable.value = !!today?.available && (today?.remaining ?? 0) > 0
  }

  function resetPracticeScreenState() {
    vocabStats.value = {
      total: 0,
      newCount: 0,
      learningCount: 0,
      reviewCount: 0,
      masteredCount: 0,
    }
    vocabStatsLoaded.value = false
    sentenceAvailable.value = false
  }

  return {
    vocabStats,
    vocabStatsLoaded,
    sentenceAvailable,
    conversationPro,
    picturePro,
    applyVocabSummary,
    applySentenceToday,
    resetPracticeScreenState,
  }
}
