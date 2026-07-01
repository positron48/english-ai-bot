<template>
  <div class="reading-text-view">
    <LgLoader v-if="loading" />
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="!block">{{ t('reading.noTexts') }}</div>
    <ReadingPassageBlock
      v-else
      :block="block"
      :text-id="textId"
      :category-id="categoryId"
      :is-read="readingIsRead"
      :cover-hero-rel-path="coverHeroRelPath"
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
import LgLoader from '../components/linglow/LgLoader.vue'

const route = useRoute()
const { t } = useI18n()

const textId = computed(() => route.params.textId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const block = ref<any>(null)
const coverHeroRelPath = ref('')
const readingIsRead = ref(false)
const categoryId = ref('')

onMounted(async () => {
  loading.value = true
  error.value = null
  try {
    const data: {
      block?: any
      category_id?: string
      cover_hero_rel_path?: string
      reading_progress?: { is_read?: boolean }
    } = await apiClient.request(`/api/learning/reading/texts/${textId.value}`)
    block.value = data.block ?? null
    coverHeroRelPath.value = String(data.cover_hero_rel_path || '').trim()
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
</style>
