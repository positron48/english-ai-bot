<template>
  <div class="reading-categories">
    <h1>{{ t('reading.title') }}</h1>
    <div v-if="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="categories.length === 0" class="empty">{{ t('reading.noTexts') }}</div>
    <div v-else class="grid">
      <router-link
        v-for="category in categories"
        :key="category.category_id"
        :to="`/learning/reading/category/${category.category_id}`"
        class="card"
      >
        <h3>{{ getLocalizedTitle(category.title, category.title_translations) }}</h3>
        <p>{{ category.level }}</p>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'

const { t, locale } = useI18n()
const loading = ref(true)
const error = ref<string | null>(null)
const categories = ref<any[]>([])

const getLocalizedTitle = (title: string, titleTranslations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && titleTranslations?.[currentLocale]) {
    return titleTranslations[currentLocale]
  }
  return title
}

onMounted(async () => {
  loading.value = true
  try {
    const data: { categories: any[] } = await apiClient.request('/api/learning/reading/categories')
    categories.value = data.categories || []
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading categories'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-categories { max-width: 1100px; margin: 0 auto; padding: 20px; }
.grid { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); }
.card { text-decoration: none; color: var(--text-primary); border: 2px solid var(--border-primary); border-radius: 10px; padding: 16px; background: var(--card-bg); }
.empty { color: var(--text-secondary); }
</style>

