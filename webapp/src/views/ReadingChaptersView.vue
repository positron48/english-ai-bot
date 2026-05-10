<template>
  <div class="reading-chapters">
    <h1>{{ title }}</h1>
    <div v-if="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="texts.length === 0" class="empty">{{ t('reading.noTexts') }}</div>
    <div v-else class="list">
      <router-link
        v-for="text in texts"
        :key="text.text_id"
        :to="`/learning/reading/text/${text.text_id}`"
        class="item"
      >
        <strong>{{ getLocalizedTitle(text.title, text.title_translations) }}</strong>
        <span>{{ text.level }}</span>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { apiClient } from '../api/client'

const { t, locale } = useI18n()
const route = useRoute()
const categoryId = computed(() => route.params.categoryId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const texts = ref<any[]>([])
const title = ref<string>('Reading')

const getLocalizedTitle = (value: string, translations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && translations?.[currentLocale]) {
    return translations[currentLocale]
  }
  return value
}

onMounted(async () => {
  loading.value = true
  try {
    const data: { category?: any; texts?: any[] } = await apiClient.request(
      `/api/learning/reading/categories/${categoryId.value}/texts`
    )
    texts.value = data.texts || []
    if (data.category?.title) {
      title.value = getLocalizedTitle(data.category.title, data.category.title_translations)
    }
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading texts'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-chapters { max-width: 900px; margin: 0 auto; padding: 20px; }
.list { display: flex; flex-direction: column; gap: 10px; }
.item { border: 2px solid var(--border-primary); border-radius: 8px; padding: 14px; text-decoration: none; color: var(--text-primary); background: var(--card-bg); display: flex; justify-content: space-between; gap: 10px; }
.empty { color: var(--text-secondary); }
</style>

