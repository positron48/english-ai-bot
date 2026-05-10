<template>
  <div class="reading-text-view">
    <div v-if="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="!block">{{ t('reading.noTexts') }}</div>
    <ReadingPassageBlock
      v-else
      :block="block"
      :text-id="textId"
      :is-read="readingIsRead"
      @marked-read="readingIsRead = true"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import ReadingPassageBlock from '../components/ReadingPassageBlock.vue'

const route = useRoute()
const { t } = useI18n()

const textId = computed(() => route.params.textId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const block = ref<any>(null)
const readingIsRead = ref(false)

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    const data: { block?: any; reading_progress?: { is_read?: boolean } } = await apiClient.request(
      `/api/learning/reading/texts/${textId.value}`
    )
    block.value = data.block ?? null
    readingIsRead.value = !!data.reading_progress?.is_read
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading text'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-text-view { margin: 0; padding: 0; }
</style>

