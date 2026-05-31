<template>
  <div class="grammar-chapters">
    <div v-if="loading" class="loading">
      <p>{{ t('grammar.loadingChapters') }}</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadChapters" class="btn btn-primary">{{ t('common.retry') }}</button>
    </div>
    
    <div v-else>
      <div class="chapters-header">
        <h1>{{ categoryTitle }}</h1>
      </div>
      
      <div v-if="allChaptersPassed && chapters.length > 0" class="category-test-banner">
        <p>{{ bannerMessage }}</p>
        <div v-if="categoryTestScore !== null" class="category-test-score">
          <span>{{ t('grammar.categoryTestScore') }}: <strong>{{ categoryTestScore }}%</strong></span>
        </div>
        <button @click="startCategoryTest" class="btn btn-primary">
          {{ categoryTestScore !== null ? t('grammar.retakeCategoryTest') : t('grammar.startCategoryTest') }}
        </button>
      </div>
      
      <div v-if="accessError" class="access-error">
        <Icon name="lock" class="error-icon" />
        <p>{{ t('grammar.completePreviousChapterToUnlock') }}</p>
      </div>
      
      <div v-if="chapters.length === 0" class="empty">
        <p>{{ t('grammar.noChaptersAvailable') }}</p>
      </div>
      
      <div v-else class="chapters-list">
        <div
          v-for="(chapter, index) in chapters"
          :key="chapter.chapter_id"
          class="chapter-item"
          :class="{ 'locked': !chapter.can_access && !chapter.passed }"
        >
          <router-link
            v-if="chapter.can_access || chapter.passed"
            :to="`/learning/grammar/chapter/${chapter.chapter_id}`"
            class="chapter-link"
          >
            <div class="chapter-info">
              <h3>{{ getLocalizedTitle(chapter.title, chapter.title_translations) }}</h3>
              <div class="chapter-meta">
                <span v-if="chapter.level" class="chapter-level">{{ chapter.level }}</span>
                <span v-if="chapter.estimated_minutes" class="chapter-time">
                  ~{{ chapter.estimated_minutes }} {{ t('grammar.min') }}
                </span>
              </div>
            </div>
            <div class="chapter-status">
              <div v-if="chapter.best_score > 0" class="status-badge" :class="{ 'passed': chapter.passed, 'attempted': !chapter.passed }">
                {{ chapter.best_score }}%
              </div>
              <div v-else class="status-badge not-started">
                0%
              </div>
              <Icon name="chevron-right" class="chevron" />
            </div>
          </router-link>
          <div v-else class="chapter-link locked-link">
            <div class="chapter-info">
              <h3>{{ getLocalizedTitle(chapter.title, chapter.title_translations) }}</h3>
              <div class="chapter-meta">
                <span v-if="chapter.level" class="chapter-level">{{ chapter.level }}</span>
                <span v-if="chapter.estimated_minutes" class="chapter-time">
                  ~{{ chapter.estimated_minutes }} {{ t('grammar.min') }}
                </span>
              </div>
            </div>
            <div class="chapter-status">
              <div class="status-badge locked">
                <Icon name="lock" />
                {{ t('grammar.locked') }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { grammarClient } from '../api/grammarClient'
import Icon from '../components/Icon.vue'

const { t, locale } = useI18n()
const route = useRoute()

interface Chapter {
  chapter_id: string
  title: string
  title_translations?: Record<string, string>
  title_short?: string
  description?: string
  level?: string
  order: number
  estimated_minutes?: number
  best_score: number
  passed: boolean
  last_attempt_at?: string
  can_access?: boolean
}

const router = useRouter()
const sectionId = computed(() => route.params.sectionId as string)

const chapters = ref<Chapter[]>([])
const categoryTitleBase = ref('')
const categoryTitleTranslations = ref<Record<string, string> | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const accessError = ref(false)
const categoryTestScore = ref<number | null>(null)
const nextCategoryAccessible = ref(false)

const allChaptersPassed = computed(() => {
  return chapters.value.length > 0 && chapters.value.every(ch => ch.passed)
})

const bannerMessage = computed(() => {
  if (nextCategoryAccessible.value) {
    return t('grammar.allChaptersCompletedCanTake')
  }
  return t('grammar.allChaptersCompletedTakeTest')
})

const getLocalizedTitle = (title: string, titleTranslations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && titleTranslations?.[currentLocale]) {
    return titleTranslations[currentLocale]
  }
  return title
}

const categoryTitle = computed(() => {
  return getLocalizedTitle(categoryTitleBase.value, categoryTitleTranslations.value || undefined)
})

const loadCategoryTitle = async () => {
  try {
    const data: any = await grammarClient.getCategories()
    const categories = data.categories || []
    const section = categories.find((s: any) => s.section_id === sectionId.value)
    
    if (section?.title) {
      categoryTitleBase.value = section.title
      categoryTitleTranslations.value = section.title_translations || null
    } else {
      // Fallback: format sectionId with proper capitalization
      const formatted = sectionId.value
        .replace(/^en\.grammar\./, '')
        .replace(/_/g, ' ')
        .split(' ')
        .map((word: string) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ')
      categoryTitleBase.value = formatted
      categoryTitleTranslations.value = null
    }
    
    // Check if next category is accessible
    const currentIndex = categories.findIndex((s: any) => s.section_id === sectionId.value)
    if (currentIndex >= 0 && currentIndex < categories.length - 1) {
      const nextCategory = categories[currentIndex + 1]
      nextCategoryAccessible.value = nextCategory.can_access || false
    }
    
    // Load category test score from current section
    if (section?.category_test_score !== undefined && section.category_test_score !== null) {
      categoryTestScore.value = section.category_test_score
    } else {
      categoryTestScore.value = null
    }
  } catch (error) {
    console.error('Failed to load category title:', error)
    // Fallback: format sectionId with proper capitalization
    const formatted = sectionId.value
      .replace(/^en\.grammar\./, '')
      .replace(/_/g, ' ')
      .split(' ')
      .map((word: string) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ')
    categoryTitleBase.value = formatted
    categoryTitleTranslations.value = null
  }
}

const loadChapters = async () => {
  loading.value = true
  error.value = null
  accessError.value = false
  
  // Check if we were redirected due to access error
  const route = useRoute()
  if (route.query.error === 'previous_chapter_not_passed') {
    accessError.value = true
  }
  
  try {
    // Load category title and test score first
    await loadCategoryTitle()
    
    const data: { chapters: Chapter[] } = await grammarClient.getChapters(sectionId.value)
    const loadedChapters = data.chapters || []
    // can_access from API: if section is unlocked (placement/previous) then all chapters; else first + after prev passed
    chapters.value = loadedChapters
  } catch (err: any) {
    error.value = err.message || 'Failed to load chapters'
    console.error('Failed to load grammar chapters:', err)
  } finally {
    loading.value = false
  }
}

const startCategoryTest = () => {
  router.push(`/learning/grammar/${sectionId.value}/test`)
}

onMounted(() => {
  loadChapters()
})
</script>

<style scoped>
.grammar-chapters {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger);
}

.chapters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.chapters-header h1 {
  margin: 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.chapters-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.chapter-item {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.chapter-item:hover {
  border-color: var(--color-primary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chapter-link {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  text-decoration: none;
  color: var(--text-primary);
}

.chapter-info {
  flex: 1;
}

.chapter-info h3 {
  margin: 0 0 8px 0;
  font-size: 18px;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.chapter-meta {
  display: flex;
  gap: 12px;
  font-size: 14px;
  color: var(--text-secondary);
}

.chapter-level {
  padding: 2px 6px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.chapter-time {
  font-size: 12px;
}

.chapter-status {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.passed {
  background: var(--color-success-light, rgba(40, 167, 69, 0.1));
  color: var(--color-success);
}

.status-badge.attempted {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.status-badge.not-started {
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
  opacity: 0.6;
}

.chevron {
  color: var(--text-tertiary);
}

.access-error {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  margin-bottom: 24px;
  background: var(--color-danger-light, rgba(239, 68, 68, 0.1));
  border: 2px solid var(--color-danger);
  border-radius: 8px;
  color: var(--color-danger);
}

.access-error .error-icon {
  flex-shrink: 0;
}

.access-error p {
  margin: 0;
  font-size: 14px;
}

.chapter-item.locked {
  opacity: 0.6;
  cursor: not-allowed;
}

.chapter-item.locked .chapter-link,
.chapter-item .locked-link {
  pointer-events: none;
  cursor: not-allowed;
}

.status-badge.locked {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.category-test-banner {
  padding: 20px;
  margin-bottom: 24px;
  background: var(--color-primary-light, rgba(59, 130, 246, 0.1));
  border: 2px solid var(--color-primary);
  border-radius: 8px;
  text-align: center;
}

.category-test-banner p {
  margin: 0 0 16px 0;
  color: var(--text-primary);
  font-size: 16px;
}

.category-test-score {
  margin-bottom: 16px;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border-radius: 6px;
  font-size: 14px;
  color: var(--text-secondary);
}

.category-test-score strong {
  color: var(--text-primary);
  font-size: 16px;
}

@media (max-width: 768px) {
  .chapters-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }

  .chapters-header h1,
  .chapter-info h3 {
    line-height: 1.25;
    hyphens: auto;
  }
  
  .chapter-link {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .chapter-status {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
