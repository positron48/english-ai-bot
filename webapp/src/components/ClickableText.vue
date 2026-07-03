<template>
  <span class="ct">
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
import VocabWordCardsDetail, { type VocabCardsAPIResponse } from './VocabWordCardsDetail.vue'

// Renders a plain sentence with every word clickable: a click resolves the word through
// /api/reading/word-lookup (which lemmatizes surface forms and adds the word to training,
// same behaviour as clicking a word in a reading text) and opens the dictionary card modal.
const props = defineProps<{ text: string }>()

const { t } = useI18n()

const tokens = computed(() => {
  const parts = (props.text || '').split(/([\p{L}\p{M}'’-]+)/u)
  return parts
    .filter((p) => p !== '')
    .map((p) => ({ text: p, word: /[\p{L}]{2,}/u.test(p) }))
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
      `/api/reading/word-lookup?lemma=${encodeURIComponent(word)}`,
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
