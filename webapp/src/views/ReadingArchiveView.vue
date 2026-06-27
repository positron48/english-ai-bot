<template>
  <div class="reading-archive">
    <div class="page-heading">
      <button
        type="button"
        class="back-btn"
        :aria-label="t('common.back')"
        @click="router.push(`/learning/reading/category/${categoryId}`)"
      >
        <Icon name="arrow-left" />
        <span>{{ t('common.back') }}</span>
      </button>
      <h1 class="page-title">{{ t('reading.archiveTitle') }}</h1>
    </div>
    <LgLoader v-if="loading" />
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="total === 0" class="empty">{{ t('reading.archiveEmpty') }}</div>
    <div v-else class="list">
      <router-link
        v-for="text in pagedTexts"
        :key="text.text_id"
        :to="`/learning/reading/text/${text.text_id}`"
        class="item"
      >
        <div class="item-main">
          <span class="item-title">{{ getLocalizedTitle(text.title, text.title_translations) }}</span>
          <span class="level-pill">{{ text.level }}</span>
        </div>
        <img
          v-if="text.cover_thumb_rel_path"
          class="item-thumb"
          :src="readingImageUrl(text.cover_thumb_rel_path)"
          alt=""
          loading="lazy"
        />
      </router-link>
    </div>
    <div v-if="!loading && !error && hasMore" class="footer-row">
      <button
        type="button"
        class="load-more-btn"
        :disabled="loadingMore"
        @click="loadMore"
      >
        {{ loadingMore ? t('common.loading') : t('reading.loadMore') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import { getGrammarCourseCode } from '../api/grammarClient'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const categoryId = computed(() => route.params.categoryId as string)
const loading = ref(true)
const loadingMore = ref(false)
const error = ref<string | null>(null)
const pagedTexts = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const perPage = 20

const hasMore = computed(() => pagedTexts.value.length < total.value)

const getLocalizedTitle = (value: string, translations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && translations?.[currentLocale]) {
    return translations[currentLocale]
  }
  return value
}

const readingImageUrl = (relPath: string) => {
  const courseCode = getGrammarCourseCode()
  const courseParam = courseCode ? `&course_code=${encodeURIComponent(courseCode)}` : ''
  return `/api/learning/reading/image?path=${encodeURIComponent(relPath)}${courseParam}`
}

const fetchPage = async (pg: number) => {
  const courseCode = getGrammarCourseCode()
  const params = new URLSearchParams({ page: String(pg), per_page: String(perPage), archive: 'true' })
  if (courseCode) params.set('course_code', courseCode)
  const data: { texts?: any[]; total?: number } = await apiClient.request(
    `/api/learning/reading/categories/${categoryId.value}/texts?${params}`
  )
  return data
}

onMounted(async () => {
  loading.value = true
  try {
    const data = await fetchPage(1)
    pagedTexts.value = data.texts || []
    total.value = data.total ?? pagedTexts.value.length
    page.value = 1
  } catch (e: any) {
    error.value = e?.message || 'Failed to load archive'
  } finally {
    loading.value = false
  }
})

const loadMore = async () => {
  if (loadingMore.value || !hasMore.value) return
  loadingMore.value = true
  try {
    const nextPage = page.value + 1
    const data = await fetchPage(nextPage)
    pagedTexts.value = [...pagedTexts.value, ...(data.texts || [])]
    total.value = data.total ?? total.value
    page.value = nextPage
  } catch (e: any) {
    error.value = e?.message || 'Failed to load more'
  } finally {
    loadingMore.value = false
  }
}
</script>

<style scoped>
.reading-archive { max-width: 900px; margin: 0 auto; padding: 20px; }
.page-heading {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 1rem;
}
.page-title { margin: 0; flex: 1; }
.back-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 42px;
  padding: 0 12px;
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  background: var(--card-bg);
  color: var(--text-primary);
  cursor: pointer;
  font: inherit;
}
.back-btn:hover {
  border-color: var(--accent-primary, #3b82f6);
  color: var(--accent-primary, #3b82f6);
}
.list { display: flex; flex-direction: column; gap: 10px; }
.item {
  border: 2px dashed var(--border-primary);
  border-radius: 8px;
  padding: 14px;
  text-decoration: none;
  color: var(--text-primary);
  background: var(--card-bg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  opacity: 0.85;
}
.item-main {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}
.item-title { font-weight: 600; }
.item-thumb {
  width: 100px;
  height: 72px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
  background: var(--bg-tertiary);
}
.level-pill {
  flex-shrink: 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
}
.footer-row { display: flex; justify-content: center; margin-top: 20px; }
.load-more-btn {
  padding: 10px 24px;
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  background: var(--card-bg);
  color: var(--text-primary);
  cursor: pointer;
  font: inherit;
  font-size: 0.95rem;
}
.load-more-btn:hover:not(:disabled) {
  border-color: var(--accent-primary, #3b82f6);
  color: var(--accent-primary, #3b82f6);
}
.load-more-btn:disabled { opacity: 0.5; cursor: default; }
.empty { color: var(--text-secondary); }
</style>
