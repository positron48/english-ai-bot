<template>
  <Teleport to="body">
    <div v-if="open" class="report-modal-backdrop" @click.self="emit('close')">
      <div class="report-modal">
        <h3 class="report-modal-title">{{ t('training.reportIssue') }}</h3>
        <p class="report-modal-hint">{{ t('training.reportCategoryHint') }}</p>
        <div class="report-category-chips">
          <button
            v-for="code in categories"
            :key="code"
            type="button"
            class="report-category-chip"
            :class="{ active: modelCategory === code }"
            @click="modelCategory = code"
          >
            {{ t(`training.reportCategories.${code}`) }}
          </button>
        </div>
        <textarea
          v-model.trim="modelDetails"
          class="report-modal-textarea"
          :placeholder="detailsPlaceholder"
          rows="4"
          maxlength="1000"
        />
        <div class="report-modal-actions">
          <button type="button" class="report-modal-cancel" @click="emit('close')">
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="report-modal-submit"
            :disabled="submitting || !canSubmit"
            @click="emit('submit')"
          >
            {{ t('training.reportSend') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { canSubmitReport } from '../constants/contentReportCategories'

const props = defineProps<{
  open: boolean
  submitting: boolean
  categories: readonly string[]
  category: string
  details: string
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'update:category': [value: string]
  'update:details': [value: string]
}>()

const { t } = useI18n()

const modelCategory = computed({
  get: () => props.category,
  set: (v: string) => emit('update:category', v)
})

const modelDetails = computed({
  get: () => props.details,
  set: (v: string) => emit('update:details', v)
})

const canSubmit = computed(() => canSubmitReport(props.category, props.details))

const detailsPlaceholder = computed(() => {
  if (props.category === 'other') {
    return t('training.reportCommentPlaceholder')
  }
  return t('training.reportDetailsOptional')
})
</script>

<style scoped>
.report-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
  padding: 16px;
}
.report-modal {
  background: var(--card-bg, #fff);
  border-radius: 12px;
  padding: 20px;
  width: 100%;
  max-width: 480px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}
.report-modal-title {
  margin: 0 0 8px;
  font-size: 1.1rem;
}
.report-modal-hint {
  margin: 0 0 12px;
  font-size: 0.85rem;
  color: var(--text-secondary, #666);
}
.report-category-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.report-category-chip {
  border: 1px solid var(--border-primary, #ddd);
  background: var(--bg-secondary, #f5f5f5);
  color: var(--text-primary, #222);
  border-radius: 999px;
  padding: 6px 12px;
  font-size: 0.8rem;
  cursor: pointer;
}
.report-category-chip.active {
  border-color: var(--accent, #4a7cff);
  background: rgba(74, 124, 255, 0.12);
}
.report-modal-textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 8px;
  padding: 10px;
  resize: vertical;
  font: inherit;
  margin-bottom: 14px;
}
.report-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.report-modal-cancel,
.report-modal-submit {
  border: none;
  border-radius: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font: inherit;
}
.report-modal-cancel {
  background: var(--bg-secondary, #eee);
}
.report-modal-submit {
  background: var(--accent, #4a7cff);
  color: #fff;
}
.report-modal-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
