<template>
  <div class="reading-text-view">
    <div v-if="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="!block">{{ t('reading.noTexts') }}</div>
    <template v-else>
      <div v-if="can('full_access')" class="reading-admin-bar">
        <button
          type="button"
          class="btn btn-sm btn-danger reading-delete-btn"
          :disabled="deleting"
          :title="t('reading.deleteText')"
          :aria-label="t('reading.deleteText')"
          @click="onDeleteText"
        >
          <Icon name="trash" />
        </button>
      </div>
      <ReadingPassageBlock
        :block="block"
        :text-id="textId"
        :category-id="categoryId"
        :is-read="readingIsRead"
        @marked-read="readingIsRead = true"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import ReadingPassageBlock from '../components/ReadingPassageBlock.vue'
import Icon from '../components/Icon.vue'
import { useAuth } from '../composables/useAuth'
import { showAlert, showConfirm } from '../composables/useDialog'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { can } = useAuth()

const textId = computed(() => route.params.textId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const block = ref<any>(null)
const readingIsRead = ref(false)
const categoryId = ref('')
const deleting = ref(false)

const readingTitle = computed(() => {
  const b = block.value
  if (!b) return ''
  return String(b.title || b.reading_passage?.title || '').trim() || textId.value
})

async function onDeleteText() {
  const ok = await showConfirm(t('reading.deleteConfirm', { title: readingTitle.value }))
  if (!ok) return
  deleting.value = true
  try {
    await apiClient.request(`/api/admin/reading/texts/${encodeURIComponent(textId.value)}`, {
      method: 'DELETE',
    })
    await showAlert(t('reading.deleteSuccess'))
    const cid = categoryId.value.trim()
    if (cid) {
      await router.push({ name: 'ReadingChapters', params: { categoryId: cid } })
    } else {
      await router.push({ name: 'ReadingCategories' })
    }
  } catch (e: any) {
    await showAlert(e?.message || t('reading.deleteFailed'))
  } finally {
    deleting.value = false
  }
}

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    const data: {
      block?: any
      category_id?: string
      reading_progress?: { is_read?: boolean }
    } = await apiClient.request(`/api/learning/reading/texts/${textId.value}`)
    block.value = data.block ?? null
    readingIsRead.value = !!data.reading_progress?.is_read
    categoryId.value = (data.category_id && String(data.category_id)) || ''
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading text'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-text-view { margin: 0; padding: 0; }

.reading-admin-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.reading-delete-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.25rem;
  padding: 0.35rem 0.5rem;
}
</style>

