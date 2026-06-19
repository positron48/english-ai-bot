<template>
  <div class="reading-categories">
    <h1 class="page-title">{{ t('reading.title') }}</h1>
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
        <h3 class="level-label">{{ categoryLabel(category) }}</h3>
        <p class="meta">
          {{ t('reading.textsCount', category.text_count, { n: category.text_count }) }}
        </p>
      </router-link>
    </div>
    <LgLumiFact :lumi-size="44" context="reading" style="margin-top: 14px" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const allCategories = ref<any[]>([])

const districtLevel = computed(() => {
  const lv = route.query.level
  return typeof lv === 'string' ? lv.toUpperCase() : ''
})

const categories = computed(() => {
  if (!districtLevel.value) return allCategories.value
  return allCategories.value.filter(c => String(c.level || '').toUpperCase() === districtLevel.value)
})

const categoryLabel = (category: { level?: string; category_id: string }) => {
  const lv = String(category.level || '').trim()
  if (lv) return lv
  return category.category_id
}

onMounted(async () => {
  loading.value = true
  try {
    const data: { categories: any[] } = await apiClient.request('/api/learning/reading/categories')
    allCategories.value = data.categories || []

    // When coming from a district, skip the category list and go directly into the matching category
    if (districtLevel.value) {
      const match = allCategories.value.filter(c => String(c.level || '').toUpperCase() === districtLevel.value)
      if (match.length === 1) {
        router.replace({
          name: 'ReadingChapters',
          params: { categoryId: match[0].category_id },
          query: { from_district: '1' },
        })
        return
      }
    }
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading categories'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-categories { max-width: 1100px; margin: 0 auto; padding: 20px; }
.page-title { margin: 0 0 1rem; }
.grid { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); }
.card { text-decoration: none; color: var(--text-primary); border: 2px solid var(--border-primary); border-radius: 10px; padding: 16px; background: var(--card-bg); }
.level-label { margin: 0 0 8px; font-size: 1.75rem; font-weight: 700; letter-spacing: 0.02em; line-height: 1.2; }
.meta { margin: 0; font-size: 0.9rem; color: var(--text-secondary); }
.empty { color: var(--text-secondary); }
</style>

