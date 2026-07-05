<template>
  <div class="reading-categories">
    <h1 class="page-title">{{ t('reading.title') }}</h1>
    <LgLoader v-if="loading" />
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="categories.length === 0" class="empty">{{ t('reading.noTexts') }}</div>
    <div v-else class="grid">
      <button
        v-for="category in categories"
        :key="category.category_id"
        class="card"
        :class="{ 'card--locked': isCategoryLocked(category) }"
        type="button"
        :disabled="isCategoryLocked(category)"
        @click="openCategory(category)"
      >
        <div class="card-head">
          <h3 class="level-label">{{ categoryLabel(category) }}</h3>
          <LgIcon v-if="isCategoryLocked(category)" name="lock" :s="16" class="lock-icon" />
        </div>
        <p class="meta">
          {{ t('reading.textsCount', category.text_count, { n: category.text_count }) }}
          <span v-if="!isCategoryLocked(category) && category.text_count > 0" class="meta-pct">
            · {{ Math.round((category.read_count || 0) / category.text_count * 100) }}%
          </span>
        </p>
      </button>
    </div>
    <LgLumiFact :lumi-size="44" context="reading" style="margin-top: 14px" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { getGrammarCourseCode, grammarClient } from '../api/grammarClient'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import { buildGrammarLevelAccess, isDistrictUnlocked, type GrammarLevelAccess } from '../utils/districtUnlock'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const error = ref<string | null>(null)
const allCategories = ref<any[]>([])
const grammarAccess = ref<GrammarLevelAccess>({})

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

const isCategoryLocked = (category: { level?: string }) => {
  const level = String(category.level || '').trim()
  if (!level) return false
  return !isDistrictUnlocked(level, grammarAccess.value)
}

const openCategory = (category: { category_id: string; level?: string }) => {
  if (isCategoryLocked(category)) return
  router.push({ name: 'ReadingChapters', params: { categoryId: category.category_id } })
}

onMounted(async () => {
  loading.value = true
  try {
    const courseCode = getGrammarCourseCode()
    const courseQuery = courseCode ? `?course_code=${encodeURIComponent(courseCode)}` : ''
    const [data, grammarData] = await Promise.all([
      apiClient.request<{ categories: any[] }>(`/api/learning/reading/categories${courseQuery}`),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    allCategories.value = data.categories || []
    grammarAccess.value = buildGrammarLevelAccess(grammarData.categories || [])

    // When coming from a district, skip the category list and go directly into the matching category
    if (districtLevel.value) {
      const match = allCategories.value.filter(c => String(c.level || '').toUpperCase() === districtLevel.value)
      if (match.length === 1 && !isCategoryLocked(match[0])) {
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
.reading-categories {
  max-width: 1100px;
  margin: 0 auto;
  padding: 20px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-primary) 22%, transparent) 0%, var(--bg-primary) 380px),
    url('/app/linglow/art/bg-reading.jpg') top center / 100% auto no-repeat;
}
.page-title { margin: 0 0 1rem; color: var(--text-primary); }
.grid { display: grid; gap: 16px; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); }
.card { text-align: left; color: var(--text-primary); border: 2px solid var(--border-primary); border-radius: 10px; padding: 16px; background: var(--card-bg); cursor: pointer; font: inherit; }
.card--locked { opacity: 0.55; cursor: default; }
.card-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.lock-icon { color: var(--text-secondary); flex-shrink: 0; }
.level-label { margin: 0 0 8px; font-size: 1.75rem; font-weight: 700; letter-spacing: 0.02em; line-height: 1.2; }
.meta { margin: 0; font-size: 0.9rem; color: var(--text-secondary); }
.empty { color: var(--text-secondary); }
</style>
