<template>
  <div class="word-sets-view">
    <h1>{{ pageTitle }}</h1>
    
    <div class="content-section">
      <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else-if="categories.length === 0 && wordSets.length === 0" class="empty-state">
        <p>{{ t('common.noItemsFound') || 'No items found.' }}</p>
      </div>
      <div v-else class="items-grid">
        <!-- Categories -->
        <div
          v-for="category in categories"
          :key="`cat-${category.id}`"
          class="category-card"
          :class="{ locked: isCategoryLocked(category) }"
          @click="isCategoryLocked(category) ? null : selectCategory(category.id)"
        >
          <div class="card-header">
            <Icon name="folder" class="card-icon" />
            <h3>{{ category.name }}</h3>
            <div v-if="isCategoryLocked(category)" class="lock-badge">
              <Icon name="lock" />
            </div>
            <div
              v-else-if="(category.total_words ?? 0) > 0"
              class="progress-badge"
              :class="getProgressClass(category.progress_percent ?? 0)"
            >
              {{ Math.round(category.progress_percent ?? 0) }}%
            </div>
          </div>
          <p v-if="category.description" class="card-description">{{ category.description }}</p>
          <div v-if="!isCategoryLocked(category) && (category.total_words ?? 0) > 0" class="word-set-stats">
            <span>{{ (category.known_words ?? 0) + (category.words_in_vocab ?? 0) }}/{{ category.total_words }} {{ (t as any)('common.words', category.total_words ?? 0) }}</span>
            <span v-if="(category.unknown_words ?? 0) > 0" class="unknown-count">
              {{ category.unknown_words }} {{ t('common.new') || 'new' }}
            </span>
          </div>
        </div>
        
        <!-- Word Sets -->
        <div 
          v-for="wordSet in wordSets" 
          :key="`set-${wordSet.id}`"
          class="word-set-card"
          @click="viewWordSet(wordSet.id)"
        >
          <div class="word-set-header">
            <h3>{{ wordSet.title }}</h3>
            <div class="progress-badge" :class="getProgressClass(wordSet.progress_percent)">
              {{ Math.round(wordSet.progress_percent) }}%
            </div>
          </div>
          <div class="word-set-stats">
            <span>{{ wordSet.known_words + wordSet.words_in_vocab }}/{{ wordSet.total_words }} {{ (t as any)('common.words', wordSet.total_words) }}</span>
            <span v-if="wordSet.unknown_words > 0" class="unknown-count">
              {{ wordSet.unknown_words }} {{ t('common.new') || 'new' }}
            </span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="selectedCategoryId !== null" class="breadcrumb">
      <button @click="goBack" class="breadcrumb-back">
        <Icon name="arrow-left" />
        <span>{{ t('common.back') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { grammarClient } from '../api/grammarClient'
import Icon from '../components/Icon.vue'
import { useCourse } from '../composables/useCourse'
import { buildGrammarLevelAccess, isDistrictUnlocked, type GrammarLevelAccess } from '../utils/districtUnlock'

const { t } = useI18n()
const { currentCourseCode, ensureCourseLoaded } = useCourse()

// Append the selected course so word sets always match the UI's chosen course.
const withCourse = (params: URLSearchParams): string => {
  if (currentCourseCode.value) params.append('course_code', currentCourseCode.value)
  return params.toString()
}

interface Category {
  id: number
  name: string
  description?: string | null
  sort_order: number
  parent_id?: number | null
  level_code?: string | null
  children?: Category[]
  total_words?: number
  known_words?: number
  words_in_vocab?: number
  unknown_words?: number
  progress_percent?: number
}

interface WordSet {
  id: number
  title: string
  description?: string | null
  total_words: number
  known_words: number
  words_in_vocab: number
  unknown_words: number
  progress_percent: number
}

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const error = ref<string | null>(null)
const categories = ref<Category[]>([])
const wordSets = ref<WordSet[]>([])
// Grammar level access, keyed by CEFR level — top-level word-set categories stay locked
// until the learner has unlocked the matching grammar level (mirrors the city map gating).
const grammarAccess = ref<GrammarLevelAccess>({})
const selectedCategoryId = ref<number | null>(null)
const currentParentId = ref<number | null>(null)
const categoryHistory = ref<number[]>([]) // Track navigation history
const allCategories = ref<Category[]>([]) // Все категории для получения названия текущей

// Computed для заголовка страницы
const pageTitle = computed(() => {
  if (selectedCategoryId.value === null) {
    return t('learning.words')
  }
  if (selectedCategoryId.value && allCategories.value.length > 0) {
    const currentCategory = allCategories.value.find(cat => cat.id === selectedCategoryId.value)
    if (currentCategory) {
      return currentCategory.name
    }
  }
  return t('learning.words')
})

// Items computed is not needed - we'll render categories and wordSets separately in template

// Resolves a CEFR level (e.g. "A0") to the id of its root word-set category
// ("Уровень A0" with no parent). Used when arriving from the city district page,
// which only knows the district's level, not any category id.
const resolveLevelCategoryId = async (level: string): Promise<number | null> => {
  try {
    const lvlParams = new URLSearchParams()
    lvlParams.append('all', 'true')
    const data: { categories: Category[] } = await apiClient.request(`/api/learning/words/categories?${withCourse(lvlParams)}`)
    allCategories.value = data.categories || []
    const match = allCategories.value.find(
      c => !c.parent_id && (c.level_code || '').toUpperCase() === level.toUpperCase()
    )
    return match ? match.id : null
  } catch (error) {
    console.error('Failed to resolve level category:', error)
    return null
  }
}

onMounted(async () => {
  // Make sure the selected course is known before fetching so we never request
  // word sets for the wrong (default) course on first paint.
  await ensureCourseLoaded()
  categoryHistory.value = []

  // Check if category_id is in query params
  const categoryIdParam = route.query.category_id
  if (categoryIdParam) {
    const categoryId = parseInt(categoryIdParam as string, 10)
    if (!isNaN(categoryId)) {
      // Set the category to show its word sets
      selectedCategoryId.value = categoryId
      // Set currentParentId to categoryId to show subcategories of this category
      currentParentId.value = categoryId
      // Add to history
      categoryHistory.value.push(categoryId)
    }
  } else {
    const levelParam = route.query.level
    const level = typeof levelParam === 'string' ? levelParam : Array.isArray(levelParam) ? levelParam[0] : null
    if (level) {
      const levelCategoryId = await resolveLevelCategoryId(level)
      if (levelCategoryId !== null) {
        selectedCategoryId.value = levelCategoryId
        currentParentId.value = levelCategoryId
        categoryHistory.value.push(levelCategoryId)
      }
    }
  }

  await Promise.all([loadCategories(), loadWordSets(), loadGrammarAccess()])
})

// Loads grammar level access so root categories for not-yet-unlocked levels show as locked.
const loadGrammarAccess = async () => {
  try {
    const { categories: grammarCategories } = await grammarClient.getCategories()
    grammarAccess.value = buildGrammarLevelAccess(grammarCategories || [])
  } catch {
    grammarAccess.value = {}
  }
}

// A top-level (root) category is locked when its CEFR level hasn't been unlocked in grammar.
// Only root categories carry a level_code and only the root list gates access.
const isCategoryLocked = (category: Category): boolean => {
  if (selectedCategoryId.value !== null) return false
  if (!category.level_code) return false
  return !isDistrictUnlocked(category.level_code, grammarAccess.value)
}

// Отслеживаем изменения query параметра category_id
watch(() => route.query.category_id, async (newCategoryId) => {
  if (newCategoryId) {
    const categoryId = typeof newCategoryId === 'string' 
      ? parseInt(newCategoryId, 10) 
      : Array.isArray(newCategoryId) 
        ? parseInt(newCategoryId[0] as string, 10)
        : null
    
    if (categoryId && !isNaN(categoryId) && categoryId !== selectedCategoryId.value) {
      selectedCategoryId.value = categoryId
      currentParentId.value = categoryId
      // Обновляем историю, если нужно
      if (!categoryHistory.value.includes(categoryId)) {
        categoryHistory.value = [categoryId]
      }
      await loadCategories()
      await loadWordSets()
    }
  } else if (selectedCategoryId.value !== null) {
    // Если category_id удален из query, возвращаемся к корню
    selectedCategoryId.value = null
    currentParentId.value = null
    categoryHistory.value = []
    await loadCategories()
    await loadWordSets()
  }
})

const loadCategories = async () => {
  try {
    // Загружаем все категории для получения названия текущей категории
    const allParams = new URLSearchParams()
    allParams.append('all', 'true')
    const allCategoriesData: { categories: Category[] } = await apiClient.request(`/api/learning/words/categories?${withCourse(allParams)}`)
    allCategories.value = allCategoriesData.categories || []

    // Загружаем дочерние категории для текущего уровня
    const params = new URLSearchParams()
    if (currentParentId.value !== null) {
      params.append('parent_id', currentParentId.value.toString())
    }
    const data: { categories: Category[] } = await apiClient.request(`/api/learning/words/categories?${withCourse(params)}`)
    categories.value = data.categories || []
  } catch (error: any) {
    console.error('Failed to load categories:', error)
    categories.value = []
    allCategories.value = []
  }
}

const loadWordSets = async () => {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    if (selectedCategoryId.value !== null) {
      params.append('category_id', selectedCategoryId.value.toString())
    }
    
    const data: { sets: WordSet[] } = await apiClient.request(`/api/learning/words/sets?${withCourse(params)}`)
    wordSets.value = data.sets || []
  } catch (error: any) {
    console.error('Failed to load word sets:', error)
    error.value = error.message || 'Failed to load word sets'
  } finally {
    loading.value = false
  }
}

const selectCategory = (categoryId: number | null) => {
  if (categoryId !== null) {
    // Add to history
    categoryHistory.value.push(categoryId)
    // Update URL with category_id query parameter
    router.push({ path: '/learning/words', query: { category_id: categoryId.toString() } })
  } else {
    // Go to root
    router.push({ path: '/learning/words' })
  }
  // Note: watch on route.query.category_id will handle the actual loading
}

const goBack = () => {
  if (categoryHistory.value.length > 0) {
    // Go back to previous category
    categoryHistory.value.pop()
    const prevCategoryId = categoryHistory.value.length > 0 
      ? categoryHistory.value[categoryHistory.value.length - 1] 
      : null
    
    if (prevCategoryId !== null) {
      router.push({ path: '/learning/words', query: { category_id: prevCategoryId.toString() } })
    } else {
      router.push({ path: '/learning/words' })
    }
    // Note: watch on route.query.category_id will handle the actual loading
  } else {
    // Go to root
    router.push({ path: '/learning/words' })
    // Note: watch on route.query.category_id will handle the actual loading
  }
}

const viewWordSet = (setId: number) => {
  // Передаем category_id в query параметрах, если мы находимся внутри категории
  if (selectedCategoryId.value !== null) {
    router.push({ 
      path: `/learning/words/${setId}`, 
      query: { category_id: selectedCategoryId.value.toString() } 
    })
  } else {
    router.push(`/learning/words/${setId}`)
  }
}

const getProgressClass = (percent: number): string => {
  if (percent >= 80) return 'progress-high'
  if (percent >= 50) return 'progress-medium'
  return 'progress-low'
}
</script>

<style scoped>
.word-sets-view {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  overflow-x: hidden;
  width: 100%;
  box-sizing: border-box;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg-primary) 22%, transparent) 0%, var(--bg-primary) 380px),
    linear-gradient(90deg, color-mix(in srgb, var(--bg-primary) 96%, transparent) 0%, color-mix(in srgb, var(--bg-primary) 78%, transparent) 48%, color-mix(in srgb, var(--bg-primary) 20%, transparent) 100%),
    url('/app/linglow/art/bg-word-cards.jpg') top center / 100% auto no-repeat;
}

.word-sets-view h1 {
  margin-bottom: 24px;
  word-wrap: break-word;
  overflow-wrap: break-word;
  color: var(--text-primary);
}

.filters-section {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  flex-wrap: wrap;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 200px;
  max-width: 400px;
}

.search-input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 16px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  margin-bottom: 0;
}

.tag-filters {
  display: flex;
  gap: 8px;
}

.tag-select {
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  min-width: 150px;
}

.content-section {
  margin-bottom: 32px;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  width: 100%;
  box-sizing: border-box;
}

.category-card {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 24px;
  background: var(--card-bg);
  cursor: pointer;
  transition: all 0.2s;
  min-height: 120px;
  display: flex;
  flex-direction: column;
}

.category-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.card-header .progress-badge {
  margin-left: auto;
}

.category-card.locked {
  cursor: default;
  opacity: 0.5;
}

.category-card.locked:hover {
  border-color: var(--border-primary);
  box-shadow: none;
  transform: none;
}

.lock-badge {
  margin-left: auto;
  display: flex;
  align-items: center;
  color: var(--text-secondary);
}

.category-card .word-set-stats {
  margin-top: auto;
}

.card-icon {
  font-size: 24px;
  color: var(--color-primary);
}

.category-card h3 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.card-description {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.5;
  margin: 0;
  flex: 1;
}

.breadcrumb {
  margin-top: 24px;
  margin-bottom: 16px;
}

.breadcrumb-back {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-primary);
  font-size: 14px;
}

.breadcrumb-back:hover {
  background: var(--bg-hover);
  border-color: var(--color-primary);
}

.word-set-card {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 24px;
  background: var(--card-bg);
  cursor: pointer;
  transition: all 0.2s;
  min-height: 120px;
  display: flex;
  flex-direction: column;
}

.word-set-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.word-set-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
  gap: 12px;
  min-width: 0;
}

.word-set-header h3 {
  margin: 0;
  font-size: 18px;
  flex: 1;
  min-width: 0;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.progress-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.progress-badge.progress-high {
  background: var(--color-success, #10b981);
  color: white;
}

.progress-badge.progress-medium {
  background: var(--color-warning, #f59e0b);
  color: white;
}

.progress-badge.progress-low {
  background: var(--color-secondary, #6b7280);
  color: white;
}

.word-set-description {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.word-set-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
  gap: 8px;
}

.unknown-count {
  color: var(--color-primary);
  font-weight: 600;
}

.loading, .error, .empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger, #ef4444);
}

@media (max-width: 768px) {
  .word-sets-view {
    padding: 12px;
  }
  
  .word-sets-view h1 {
    font-size: 24px;
    margin-bottom: 16px;
  }
  
  .items-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
  
  .category-card,
  .word-set-card {
    padding: 16px;
    min-height: auto;
  }
  
  .category-card h3,
  .word-set-header h3 {
    font-size: 16px;
  }
  
  .word-set-header {
    flex-wrap: wrap;
  }
  
  .progress-badge {
    font-size: 11px;
    padding: 3px 10px;
  }
  
  .word-set-stats {
    font-size: 13px;
    flex-direction: column;
    align-items: flex-start;
  }
  
  .card-header {
    gap: 8px;
  }
  
  .card-icon {
    font-size: 20px;
    flex-shrink: 0;
  }
}

@media (max-width: 480px) {
  .word-sets-view {
    padding: 8px;
  }
  
  .items-grid {
    gap: 12px;
  }
  
  .category-card,
  .word-set-card {
    padding: 12px;
  }
  
  .word-sets-view h1 {
    font-size: 20px;
  }
}
</style>
