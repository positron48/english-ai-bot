import { computed, ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { useLearningConfig, ensureLearningLoaded } from './useLearningConfig'

const verbFormsTotalCardsPool = ref<number | null>(null)
const verbFormsPoolResolved = ref(false)

export function useSpanishVerbFormsPractice(isOnline: Ref<boolean>) {
  const router = useRouter()
  const { learning } = useLearningConfig()

  async function refreshVerbFormsPoolCount(force = false) {
    if (!force && verbFormsPoolResolved.value) return
    await ensureLearningLoaded()
    const ly = learning.value
    if (!ly || ly.target_lang?.toLowerCase() !== 'es' || !ly.spanish_verb_forms_enabled) {
      verbFormsTotalCardsPool.value = null
      verbFormsPoolResolved.value = true
      return
    }
    try {
      const data = await apiClient.request<{ total_cards?: number }>('/api/verb-training/upcoming')
      verbFormsTotalCardsPool.value = typeof data?.total_cards === 'number' ? data.total_cards : null
    } catch {
      verbFormsTotalCardsPool.value = null
    } finally {
      verbFormsPoolResolved.value = true
    }
  }

  function applyVerbFormsPool(raw: { total_cards?: number } | null | undefined) {
    const ly = learning.value
    if (!ly || ly.target_lang?.toLowerCase() !== 'es' || !ly.spanish_verb_forms_enabled) {
      verbFormsTotalCardsPool.value = null
    } else {
      verbFormsTotalCardsPool.value = typeof raw?.total_cards === 'number' ? raw.total_cards : null
    }
    verbFormsPoolResolved.value = true
  }

  function resetVerbFormsPool() {
    verbFormsTotalCardsPool.value = null
    verbFormsPoolResolved.value = false
  }

  const showSpanishVerbFormsTraining = computed(
    () =>
      (learning.value?.target_lang || '').toLowerCase() === 'es' &&
      isOnline.value &&
      learning.value?.spanish_verb_forms_enabled === true &&
      verbFormsPoolResolved.value &&
      verbFormsTotalCardsPool.value !== null &&
      verbFormsTotalCardsPool.value > 0
  )

  function openVerbFormsTraining() {
    void router.push({ path: '/training/verbs', query: { start: '1' } })
  }

  return {
    verbFormsTotalCardsPool,
    showSpanishVerbFormsTraining,
    refreshVerbFormsPoolCount,
    applyVerbFormsPool,
    resetVerbFormsPool,
    openVerbFormsTraining,
  }
}
