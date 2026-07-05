<template>
  <section class="lg-word-lookup">
    <form class="lg-word-lookup-form" @submit.prevent="submit">
      <div class="lg-word-lookup-field">
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
          <LgIcon name="x" :s="16" c="var(--subtext)" />
        </button>
      </div>
      <button
        type="submit"
        class="lg-word-lookup-btn"
        :disabled="!canSubmit"
        :aria-label="t('dashboard.wordLookupSubmit')"
      >
        <LgIcon name="book-open" :s="18" c="var(--text)" />
      </button>
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
  border: 1px solid var(--border);
  border-radius: 18px;
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  padding: 12px 14px;
  margin-bottom: 10px;
}

.lg-word-lookup-form {
  display: flex;
  gap: 8px;
  align-items: stretch;
}

.lg-word-lookup-field {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 4px;
  min-height: 42px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--input-bg, var(--card-bg));
  padding: 0 8px 0 12px;
}

.lg-word-lookup-field:focus-within {
  border-color: var(--dorado, #c9a227);
}

.lg-word-lookup-input {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--text);
  font-size: 16px;
  line-height: 1.25;
  padding: 10px 0;
}

.lg-word-lookup-input:focus {
  outline: none;
}

.lg-word-lookup-input:disabled {
  opacity: 0.6;
}

.lg-word-lookup-clear {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 8px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.lg-word-lookup-clear:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.lg-word-lookup-btn {
  width: 42px;
  flex-shrink: 0;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 0;
}

.lg-word-lookup-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.lg-word-lookup-error {
  margin: 8px 0 0;
  font-size: 13px;
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
