<template>
  <span class="ct" :class="{ 'ct--subtle-underline': subtleUnderline }">
    <template v-for="(tk, i) in tokens" :key="i">
      <span
        v-if="tk.word"
        class="ct-word"
        :class="{ 'ct-word--selected': selectedIdx === i }"
        @click.stop="onWordClick(tk.text, i)"
      >{{ tk.text }}</span>
      <template v-else>{{ tk.text }}</template>
    </template>
  </span>
  <Teleport to="body">
    <div v-if="modalVisible" class="ct-modal-overlay" @click.self="closeModal">
      <div class="ct-modal-panel">
        <div v-if="lookupLoading" class="ct-modal-loading">
          {{ lookupGenerating ? t('reading.wordGenerating') : t('common.loading') }}
        </div>
        <div v-else-if="lookupError" class="ct-modal-error">
          <p class="ct-modal-error-text">{{ lookupError }}</p>
          <button type="button" class="ct-modal-close-btn" @click="closeModal">{{ t('common.close') }}</button>
        </div>
        <VocabWordCardsDetail
          v-else
          :lemma="modalLemma"
          :preloaded="modalPreloaded"
          @close="closeModal"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { withCourseCode } from '../api/grammarClient'
import VocabWordCardsDetail, { type VocabCardsAPIResponse } from './VocabWordCardsDetail.vue'

// Renders a plain sentence with every word clickable: a click resolves the word through
// /api/reading/word-lookup (which lemmatizes surface forms and adds the word to training,
// same behaviour as clicking a word in a reading text) and opens the dictionary card modal.
// `exclude` is the word being trained: opening its own card would spoil the answer, so the
// occurrence closest to it (exact match first, then Levenshtein distance to catch inflected
// forms) is rendered as plain text.
const props = defineProps<{ text: string; exclude?: string; subtleUnderline?: boolean }>()

const { t } = useI18n()

const levenshtein = (a: string, b: string): number => {
  if (a === b) return 0
  const m = a.length
  const n = b.length
  if (m === 0 || n === 0) return m + n
  let prev = Array.from({ length: n + 1 }, (_, j) => j)
  for (let i = 1; i <= m; i++) {
    const cur = [i]
    for (let j = 1; j <= n; j++) {
      cur[j] = Math.min(
        prev[j] + 1,
        cur[j - 1] + 1,
        prev[j - 1] + (a[i - 1] === b[j - 1] ? 0 : 1),
      )
    }
    prev = cur
  }
  return prev[n]
}

const tokens = computed(() => {
  const parts = (props.text || '').split(/([\p{L}\p{M}'’-]+)/u)
  const list = parts
    .filter((p) => p !== '')
    .map((p) => ({ text: p, word: /[\p{L}]{2,}/u.test(p) }))

  // Multi-word targets ("to go", "echar de menos"): match tokens against each target word.
  const targetWords = (props.exclude || '')
    .toLowerCase()
    .split(/[^\p{L}\p{M}'’-]+/u)
    .filter((w) => /[\p{L}]{2,}/u.test(w))
  if (targetWords.length > 0) {
    const exact = list.filter((tk) => tk.word && targetWords.includes(tk.text.toLowerCase()))
    if (exact.length > 0) {
      for (const tk of exact) tk.word = false
    } else {
      let best = -1
      let bestDist = Infinity
      list.forEach((tk, i) => {
        if (!tk.word) return
        const d = Math.min(...targetWords.map((w) => levenshtein(tk.text.toLowerCase(), w)))
        if (d < bestDist) {
          bestDist = d
          best = i
        }
      })
      if (best >= 0) list[best].word = false
    }
  }
  return list
})

const selectedIdx = ref(-1)
const modalVisible = ref(false)
const modalLemma = ref('')
const modalPreloaded = ref<VocabCardsAPIResponse | null>(null)
const lookupLoading = ref(false)
const lookupGenerating = ref(false)
const lookupError = ref('')
let generatingTimer: ReturnType<typeof setTimeout> | null = null

const closeModal = () => {
  modalVisible.value = false
  selectedIdx.value = -1
}

const modalKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') closeModal()
}
watch(modalVisible, (open) => {
  if (open) window.addEventListener('keydown', modalKeydown)
  else window.removeEventListener('keydown', modalKeydown)
})
onUnmounted(() => {
  window.removeEventListener('keydown', modalKeydown)
  if (generatingTimer) clearTimeout(generatingTimer)
})

const onWordClick = async (word: string, idx: number) => {
  selectedIdx.value = idx
  modalVisible.value = true
  lookupLoading.value = true
  lookupGenerating.value = false
  lookupError.value = ''
  modalPreloaded.value = null
  modalLemma.value = word

  if (generatingTimer) clearTimeout(generatingTimer)
  generatingTimer = setTimeout(() => { lookupGenerating.value = true }, 600)

  try {
    const data: VocabCardsAPIResponse = await apiClient.request(
      withCourseCode(`/api/reading/word-lookup?lemma=${encodeURIComponent(word)}`),
    )
    modalLemma.value = data.lemma || word
    modalPreloaded.value = data
  } catch (error: any) {
    console.error('Word lookup failed', error)
    modalPreloaded.value = null
    const status = typeof error?.status === 'number' ? error.status : 0
    if (error?.isNetworkError) {
      lookupError.value = t('reading.wordLookupNetwork')
    } else if (status === 404) {
      lookupError.value = t('reading.wordNotFound')
    } else if (status >= 500) {
      lookupError.value = t('reading.wordLookupServerError')
    } else {
      lookupError.value = t('reading.wordLookupFailed')
    }
  } finally {
    if (generatingTimer) {
      clearTimeout(generatingTimer)
      generatingTimer = null
    }
    lookupLoading.value = false
    lookupGenerating.value = false
  }
}
</script>

<style scoped>
.ct-word {
  cursor: pointer;
  border-bottom: 1px dashed currentColor;
}
.ct--subtle-underline .ct-word {
  border-bottom-color: color-mix(in srgb, currentColor 22%, transparent);
}
.ct-word:hover,
.ct-word--selected {
  opacity: 0.75;
}
.ct-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
  padding: 16px;
}
.ct-modal-panel {
  background: var(--card-bg, #fff);
  color: var(--text, inherit);
  border-radius: 14px;
  width: 100%;
  max-width: 560px;
  max-height: 88vh;
  overflow: auto;
  padding: 16px;
}
.ct-modal-loading,
.ct-modal-error {
  padding: 24px 8px;
  text-align: center;
}
.ct-modal-error-text {
  margin: 0 0 12px;
}
.ct-modal-close-btn {
  border: 1px solid var(--border, rgba(0, 0, 0, 0.15));
  background: transparent;
  color: inherit;
  border-radius: 10px;
  padding: 8px 14px;
  cursor: pointer;
}
</style>
