<template>
  <section class="lg-word-lookup">
    <form class="lg-word-lookup-form" @submit.prevent="submit">
      <div class="lg-word-lookup-field">
        <span class="lg-word-lookup-icon" aria-hidden="true">
          <LgIcon name="book-open" :s="16" c="var(--subtext)" />
        </span>
        <input
          v-model="query"
          type="text"
          class="lg-word-lookup-input"
          :placeholder="t('dashboard.wordLookupPlaceholder')"
          :disabled="loading"
          autocomplete="off"
          enterkeyhint="search"
        />
        <button
          v-if="query"
          type="button"
          class="lg-word-lookup-clear"
          :disabled="loading"
          :aria-label="t('dashboard.wordLookupClear')"
          @click="clearQuery"
        >
          <LgIcon name="x" :s="14" c="var(--subtext)" />
        </button>
        <button
          type="submit"
          class="lg-word-lookup-btn"
          :disabled="!canSubmit"
          :aria-label="t('dashboard.wordLookupSubmit')"
        >
          <LgIcon name="chevron-right" :s="16" c="var(--subtext)" />
        </button>
      </div>
    </form>
    <p v-if="inlineError" class="lg-word-lookup-error">{{ inlineError }}</p>

    <Teleport to="body">
      <div v-if="modalVisible" class="word-modal-overlay" @click.self="closeModal">
        <div class="word-modal-panel">
          <LgLumiCardLoading
            v-if="loading"
            :message="generating ? t('reading.wordGenerating') : t('common.loading')"
          />
          <div v-else-if="error" class="word-modal-error">
            <p class="word-modal-error-text">{{ error }}</p>
            <button type="button" class="word-modal-close-btn" @click="closeModal">{{ t('common.close') }}</button>
          </div>
          <VocabWordCardsDetail
            v-else
            :lemma="lemma"
            :preloaded="preloaded"
            @close="closeModal"
          />
        </div>
      </div>
    </Teleport>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { isWordLookupCandidate, useWordLookup } from '../../composables/useWordLookup'
import LgIcon from './LgIcon.vue'
import LgLumiCardLoading from './LgLumiCardLoading.vue'
import VocabWordCardsDetail from '../VocabWordCardsDetail.vue'

const { t } = useI18n()
const query = ref('')
const inlineError = ref('')

const {
  modalVisible,
  loading,
  generating,
  error,
  lemma,
  preloaded,
  lookup,
  closeModal,
  bindModalKeydown,
} = useWordLookup()

watch(modalVisible, bindModalKeydown)

const canSubmit = computed(() => {
  const q = query.value.trim()
  return q.length > 0 && isWordLookupCandidate(q) && !loading.value
})

const clearQuery = () => {
  query.value = ''
  inlineError.value = ''
}

const submit = async () => {
  const q = query.value.trim()
  if (!q) return
  if (!isWordLookupCandidate(q)) {
    inlineError.value = t('reading.wordNotFound')
    return
  }
  inlineError.value = ''
  await lookup(q)
}
</script>

<style scoped>
.lg-word-lookup {
  padding: 0 4px;
  margin-top: -4px;
  margin-bottom: -2px;
}

.lg-word-lookup-form {
  display: block;
}

.lg-word-lookup-field {
  display: flex;
  align-items: center;
  gap: 6px;
  min-height: 34px;
  padding: 0 2px;
  border: none;
  border-bottom: 1px solid color-mix(in srgb, var(--border) 70%, transparent);
  border-radius: 0;
  background: transparent;
  transition: border-color 0.15s ease;
}

.lg-word-lookup-field:focus-within {
  border-bottom-color: color-mix(in srgb, var(--dorado, #c9a227) 55%, var(--border));
}

.lg-word-lookup-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  opacity: 0.55;
}

.lg-word-lookup-input {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--text);
  font-size: 15px;
  line-height: 1.25;
  padding: 6px 0;
  margin-bottom: 0;
}

.lg-word-lookup-input::placeholder {
  color: var(--subtext);
  opacity: 0.65;
}

.lg-word-lookup-input:focus {
  outline: none;
}

.lg-word-lookup-input:disabled {
  opacity: 0.6;
}

.lg-word-lookup-clear,
.lg-word-lookup-btn {
  flex-shrink: 0;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 6px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0;
  opacity: 0.55;
  transition: opacity 0.15s ease, background-color 0.15s ease;
}

.lg-word-lookup-clear:hover:not(:disabled),
.lg-word-lookup-btn:hover:not(:disabled) {
  opacity: 0.85;
  background: color-mix(in srgb, var(--border) 35%, transparent);
}

.lg-word-lookup-btn:not(:disabled) {
  opacity: 0.75;
}

.lg-word-lookup-clear:disabled,
.lg-word-lookup-btn:disabled {
  opacity: 0.25;
  cursor: not-allowed;
}

.lg-word-lookup-error {
  margin: 6px 0 0;
  padding: 0 2px;
  font-size: 12px;
  color: var(--color-danger, #c0392b);
}

.word-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: var(--bg-modal-overlay, rgba(0, 0, 0, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.word-modal-panel {
  background: var(--card-bg);
  border-radius: 8px;
  max-width: 800px;
  width: 100%;
  max-height: min(90vh, 900px);
  overflow-y: auto;
  padding: 24px 28px;
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
  box-shadow: 0 24px 64px var(--card-shadow);
}

.word-modal-error {
  text-align: center;
  padding: 48px 16px;
}

.word-modal-error-text {
  margin: 0 0 12px;
}

.word-modal-close-btn {
  border: 1px solid var(--border-primary);
  background: transparent;
  color: inherit;
  border-radius: 8px;
  padding: 8px 14px;
  cursor: pointer;
}
</style>
