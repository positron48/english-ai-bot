<template>
  <nav v-if="breadcrumbs.length > 0" class="breadcrumbs">
    <ol class="breadcrumbs-list">
      <li v-for="(crumb, index) in breadcrumbs" :key="index" class="breadcrumb-item">
        <router-link 
          v-if="index < breadcrumbs.length - 1" 
          :to="crumb.path" 
          class="breadcrumb-link"
        >
          <Icon v-if="index === 0" name="home" class="breadcrumb-icon" />
          <span>{{ crumb.label }}</span>
        </router-link>
        <span v-else class="breadcrumb-current">
          {{ crumb.label }}
        </span>
        <Icon 
          v-if="index < breadcrumbs.length - 1" 
          name="chevron-right" 
          class="breadcrumb-separator" 
        />
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import Icon from './Icon.vue'

const { t } = useI18n()

interface Breadcrumb {
  label: string
  path: string
}

interface Category {
  id: number
  name: string
  parent_id?: number | null
}

interface WordSet {
  id: number
  title: string
  category_id?: number | null
}

const route = useRoute()

const wordSetInfo = ref<{ wordSet: WordSet | null; categories: Category[] }>({
  wordSet: null,
  categories: []
})

const currentCategoryId = ref<number | null>(null)

// Grammar-specific state
const grammarSectionName = ref<string | null>(null)
const grammarSectionId = ref<string | null>(null)
const grammarChapterName = ref<string | null>(null)

// Загружаем все категории для построения иерархии
const loadCategories = async () => {
  try {
    const categoriesData: { categories: Category[] } = await apiClient.request('/api/learning/words/categories?all=true')
    wordSetInfo.value.categories = categoriesData.categories || []
  } catch (error) {
    console.error('Failed to load categories for breadcrumbs:', error)
    wordSetInfo.value.categories = []
  }
}

// Загружаем название секции грамматики
const loadGrammarSectionName = async (sectionId: string) => {
  try {
    const data: any = await apiClient.request('/api/learning/grammar/categories')
    const categories = data.categories || []
    const section = categories.find((s: any) => s.section_id === sectionId)
    
    if (section?.title) {
      grammarSectionName.value = section.title
      grammarSectionId.value = sectionId
    } else {
      // Fallback: format sectionId
      grammarSectionName.value = sectionId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
      grammarSectionId.value = sectionId
      console.warn('Section not found in categories, using formatted ID:', sectionId, 'Available sections:', categories.map((s: any) => s.section_id))
    }
  } catch (error) {
    console.error('Failed to load grammar section name:', error, sectionId)
    // Fallback: format sectionId
    grammarSectionName.value = sectionId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
    grammarSectionId.value = sectionId
  }
}

// Загружаем название главы грамматики и секции
const loadGrammarChapterName = async (chapterId: string) => {
  try {
    const data: any = await apiClient.request(`/api/learning/grammar/chapters/${chapterId}`)
    
    // Extract title - API returns { title: string, chapter: Chapter }
    if (data.title) {
      grammarChapterName.value = data.title
    } else {
      // Fallback: format chapterId
      grammarChapterName.value = chapterId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
    }
    
    // Extract section_id from chapter data
    // Chapter has section_id field (json: "section_id")
    let sectionId: string | null = null
    if (data.chapter?.section_id) {
      sectionId = data.chapter.section_id
    } else {
      // Fallback: extract section_id from chapterId
      // Example: en.grammar.orientation_how_to_read.sentence_modes_statement_question
      // -> en.grammar.orientation_how_to_read
      const sectionMatch = chapterId.match(/^(.+)\.[^.]+$/)
      if (sectionMatch) {
        sectionId = sectionMatch[1]
      }
    }
    
    if (sectionId) {
      grammarSectionId.value = sectionId
      await loadGrammarSectionName(sectionId)
    }
  } catch (error) {
    console.error('Failed to load grammar chapter name:', error, chapterId)
    // Fallback: format chapterId
    grammarChapterName.value = chapterId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
    // Try to extract section_id from chapterId as fallback
    const sectionMatch = chapterId.match(/^(.+)\.[^.]+$/)
    if (sectionMatch) {
      const extractedSectionId = sectionMatch[1]
      grammarSectionId.value = extractedSectionId
      await loadGrammarSectionName(extractedSectionId)
    }
  }
}

// Загружаем информацию о наборе и категориях для страницы деталей набора
const loadWordSetInfo = async (setId: string) => {
  try {
    // Загружаем информацию о наборе
    const setData: { word_set: WordSet } = await apiClient.request(`/api/learning/words/sets/${setId}`)
    wordSetInfo.value.wordSet = setData.word_set
    
    // Загружаем все категории для построения иерархии
    await loadCategories()
  } catch (error) {
    console.error('Failed to load word set info for breadcrumbs:', error)
    wordSetInfo.value.wordSet = null
    wordSetInfo.value.categories = []
  }
}

// Функция для построения пути категорий от корня до указанной категории
const buildCategoryPath = (categoryId: number | null | undefined, allCategories: Category[]): Category[] => {
  if (!categoryId) return []
  
  const categoryMap = new Map<number, Category>()
  allCategories.forEach(cat => categoryMap.set(cat.id, cat))
  
  const path: Category[] = []
  let currentId: number | null | undefined = categoryId
  
  while (currentId) {
    const category = categoryMap.get(currentId)
    if (!category) break
    
    path.unshift(category)
    currentId = category.parent_id
  }
  
  return path
}

// Определяем иерархию маршрутов (используем computed для i18n)
const routeHierarchy = computed(() => {
  return {
    '/dashboard': [
      { label: t('navigation.dashboard'), path: '/dashboard' }
    ],
    '/vocab': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.vocab'), path: '/vocab' }
    ],
    '/learning': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.learning'), path: '/learning' }
    ],
    '/learning/grammar': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.learning'), path: '/learning' },
      { label: t('learning.grammar'), path: '/learning/grammar' }
    ],
    // Dynamic grammar routes will be handled in computed
    '/learning/words': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.learning'), path: '/learning' },
      { label: t('learning.words'), path: '/learning/words' }
    ],
    '/chat': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.chat'), path: '/chat' }
    ],
    '/settings': [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.settings'), path: '/settings' }
    ],
    '/admin': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' }
    ],
    '/admin/circuit-breaker': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Circuit Breaker', path: '/admin/circuit-breaker' }
    ],
    '/admin/prompt-tester': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Prompt Tester', path: '/admin/prompt-tester' }
    ],
    '/admin/orphaned-cards': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Orphaned Cards', path: '/admin/orphaned-cards' }
    ],
    '/admin/word-sets': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Word Sets', path: '/admin/word-sets' }
    ],
    '/admin/word-sets/categories': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Word Sets', path: '/admin/word-sets' },
      { label: 'Categories', path: '/admin/word-sets/categories' }
    ],
    '/admin/word-sets/sets': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'Word Sets', path: '/admin/word-sets' },
      { label: 'Sets', path: '/admin/word-sets/sets' }
    ],
    '/admin/db-schema': [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Admin', path: '/admin' },
      { label: 'DB Schema', path: '/admin/db-schema' }
    ]
  } as Record<string, Breadcrumb[]>
})

const breadcrumbs = computed(() => {
  const currentPath = route.path
  
  // Страницы без крошек
  if (currentPath === '/training' || currentPath.match(/^\/learning\/words\/\d+\/study$/)) {
    return []
  }
  
  // Для grammar routes
  if (currentPath.startsWith('/learning/grammar/')) {
    const parts = currentPath.split('/').filter(p => p)
    const result: Breadcrumb[] = [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.learning'), path: '/learning' },
      { label: t('learning.grammar'), path: '/learning/grammar' }
    ]
    
    // /learning/grammar/:sectionId
    if (parts.length === 3 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] !== 'chapter') {
      const sectionId = parts[2]
      const sectionLabel = grammarSectionName.value || sectionId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
      result.push({ label: sectionLabel, path: currentPath })
    }
    // /learning/grammar/chapter/:chapterId
    else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter') {
      const chapterId = parts[3]
      // Format chapter label: remove en.grammar prefix and replace underscores
      const chapterLabel = grammarChapterName.value || chapterId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
      // Add section name if available
      if (grammarSectionName.value && grammarSectionId.value) {
        result.push({ label: grammarSectionName.value, path: `/learning/grammar/${grammarSectionId.value}` })
      } else {
        // Try to show section ID as fallback
        const sectionMatch = chapterId.match(/^(.+)\.[^.]+$/)
        if (sectionMatch) {
          const sectionId = sectionMatch[1]
          const sectionLabel = sectionId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
          result.push({ label: sectionLabel, path: `/learning/grammar/${sectionId}` })
        }
      }
      result.push({ label: chapterLabel, path: currentPath })
    }
    // /learning/grammar/:sectionId/test
    else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[3] === 'test') {
      const sectionId = parts[2]
      const sectionLabel = grammarSectionName.value || sectionId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
      result.push(
        { label: sectionLabel, path: `/learning/grammar/${sectionId}` },
        { label: t('common.test') || 'Test', path: currentPath }
      )
    }
    // /learning/grammar/chapter/:chapterId/test
    else if (parts.length === 5 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter' && parts[4] === 'test') {
      const chapterId = parts[3]
      const chapterLabel = grammarChapterName.value || chapterId.replace(/^en\.grammar\./, '').replace(/_/g, ' ')
      // Add section name if available
      if (grammarSectionName.value && grammarSectionId.value) {
        result.push({ label: grammarSectionName.value, path: `/learning/grammar/${grammarSectionId.value}` })
      }
      result.push(
        { label: chapterLabel, path: `/learning/grammar/chapter/${chapterId}` },
        { label: t('common.test') || 'Test', path: currentPath }
      )
    }
    
    return result
  }
  
  // Для динамических маршрутов типа /learning/words/:setId
  if (currentPath.match(/^\/learning\/words\/\d+$/) && !currentPath.endsWith('/study')) {
    const setId = currentPath.split('/').pop()
    const result: Breadcrumb[] = [
      { label: t('navigation.dashboard'), path: '/dashboard' },
      { label: t('navigation.learning'), path: '/learning' },
      { label: t('learning.words'), path: '/learning/words' }
    ]
    
    // Определяем category_id: сначала из query параметра (если перешли из категории),
    // потом из загруженной информации о наборе
    let categoryId: number | null | undefined = null
    
    // Проверяем query параметр (если перешли из категории)
    if (route.query.category_id) {
      const categoryIdParam = route.query.category_id
      categoryId = typeof categoryIdParam === 'string' 
        ? parseInt(categoryIdParam, 10) 
        : Array.isArray(categoryIdParam) 
          ? parseInt(categoryIdParam[0] as string, 10)
          : null
      if (categoryId && isNaN(categoryId)) {
        categoryId = null
      }
    }
    
    // Если нет в query, используем из загруженной информации о наборе
    if (!categoryId && wordSetInfo.value.wordSet?.category_id) {
      categoryId = wordSetInfo.value.wordSet.category_id
    }
    
    // Добавляем иерархию категорий, если категории загружены и есть category_id
    if (categoryId && wordSetInfo.value.categories.length > 0) {
      const categoriesPath = buildCategoryPath(categoryId, wordSetInfo.value.categories)
      
      // Добавляем категории в хлебные крошки
      categoriesPath.forEach((category) => {
        result.push({
          label: category.name,
          path: `/learning/words?category_id=${category.id}`
        })
      })
    }
    
    // Добавляем название набора в конец
    // Не показываем название пока оно не загружено, чтобы избежать "прыжков"
    // Если данные еще не загружены, просто не добавляем название набора
    if (wordSetInfo.value.wordSet?.title) {
      result.push({
        label: wordSetInfo.value.wordSet.title,
        path: currentPath
      })
    } else if (wordSetInfo.value.wordSet === null) {
      // Данные еще загружаются - не показываем название, чтобы избежать прыжков
      // Можно показать просто "Loading..." или ничего не показывать
      // Но лучше показать базовую структуру без названия набора
    } else {
      // wordSetInfo.value.wordSet существует, но title еще не загружен
      // Это маловероятно, но на всякий случай
    }
    
    return result
  }
  
  // Для маршрута /learning/words с query параметром category_id
  if (currentPath === '/learning/words' && route.query.category_id) {
    const categoryIdParam = route.query.category_id
    const categoryId = typeof categoryIdParam === 'string' 
      ? parseInt(categoryIdParam, 10) 
      : Array.isArray(categoryIdParam) 
        ? parseInt(categoryIdParam[0] as string, 10)
        : null
    
    if (categoryId && !isNaN(categoryId)) {
      const result: Breadcrumb[] = [
        { label: t('navigation.dashboard'), path: '/dashboard' },
        { label: t('navigation.learning'), path: '/learning' },
        { label: t('learning.words'), path: '/learning/words' }
      ]
      
      // Добавляем иерархию категорий, если они загружены
      if (wordSetInfo.value.categories.length > 0) {
        const categoriesPath = buildCategoryPath(categoryId, wordSetInfo.value.categories)
        categoriesPath.forEach((category) => {
          result.push({
            label: category.name,
            path: `/learning/words?category_id=${category.id}`
          })
        })
      }
      
      return result
    }
  }
  
  // Для статических маршрутов
  return routeHierarchy.value[currentPath] || []
})

// Загружаем информацию о наборе при изменении маршрута
watch(() => [route.path, route.query.category_id], ([newPath, categoryId]) => {
  const match = newPath.match(/^\/learning\/words\/(\d+)$/)
  if (match && !newPath.endsWith('/study')) {
    // Загружаем категории сразу, чтобы крошки могли их использовать
    // даже пока загружается информация о наборе
    if (wordSetInfo.value.categories.length === 0) {
      loadCategories()
    }
    loadWordSetInfo(match[1])
  } else if (newPath === '/learning/words') {
    // Загружаем категории для страницы списка категорий (всегда, чтобы они были доступны для крошек)
    loadCategories()
    if (categoryId) {
      const catId = typeof categoryId === 'string' 
        ? parseInt(categoryId, 10) 
        : Array.isArray(categoryId) 
          ? parseInt(categoryId[0] as string, 10)
          : null
      currentCategoryId.value = catId && !isNaN(catId) ? catId : null
    } else {
      currentCategoryId.value = null
    }
    wordSetInfo.value.wordSet = null
  } else if (newPath.startsWith('/learning/grammar/')) {
    // Загружаем названия для грамматики
    const parts = newPath.split('/').filter(p => p)
    // parts[0] = 'learning', parts[1] = 'grammar', parts[2] = 'chapter' or sectionId, etc.
    if (parts.length === 3 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] !== 'chapter') {
      // Section page: /learning/grammar/:sectionId
      const sectionId = parts[2]
      grammarChapterName.value = null
      loadGrammarSectionName(sectionId)
    } else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter') {
      // Chapter page: /learning/grammar/chapter/:chapterId
      const chapterId = parts[3]
      // Load chapter name first, which will also load section name
      loadGrammarChapterName(chapterId)
    } else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[3] === 'test') {
      // Section test: /learning/grammar/:sectionId/test
      const sectionId = parts[2]
      grammarChapterName.value = null
      loadGrammarSectionName(sectionId)
    } else if (parts.length === 5 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter' && parts[4] === 'test') {
      // Chapter test: /learning/grammar/chapter/:chapterId/test
      const chapterId = parts[3]
      // Load chapter name first, which will also load section name
      loadGrammarChapterName(chapterId)
    } else {
      grammarSectionName.value = null
      grammarSectionId.value = null
      grammarChapterName.value = null
    }
  } else {
    wordSetInfo.value.wordSet = null
    currentCategoryId.value = null
    grammarSectionName.value = null
    grammarChapterName.value = null
    // Не очищаем categories, они могут понадобиться
  }
}, { immediate: true })

onMounted(() => {
  // Всегда загружаем категории, если их еще нет, чтобы они были доступны для крошек
  if (wordSetInfo.value.categories.length === 0) {
    loadCategories()
  }
  
  // Load grammar names if on grammar route (watch with immediate: true should handle this, but ensure it happens)
  if (route.path.startsWith('/learning/grammar/')) {
    const parts = route.path.split('/').filter(p => p)
    // parts[0] = 'learning', parts[1] = 'grammar', parts[2] = 'chapter' or sectionId, etc.
    if (parts.length === 3 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] !== 'chapter') {
      // Section page: /learning/grammar/:sectionId
      const sectionId = parts[2]
      loadGrammarSectionName(sectionId)
    } else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter') {
      // Chapter page: /learning/grammar/chapter/:chapterId
      const chapterId = parts[3]
      loadGrammarChapterName(chapterId)
    } else if (parts.length === 4 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[3] === 'test') {
      // Section test: /learning/grammar/:sectionId/test
      const sectionId = parts[2]
      loadGrammarSectionName(sectionId)
    } else if (parts.length === 5 && parts[0] === 'learning' && parts[1] === 'grammar' && parts[2] === 'chapter' && parts[4] === 'test') {
      // Chapter test: /learning/grammar/chapter/:chapterId/test
      const chapterId = parts[3]
      loadGrammarChapterName(chapterId)
    }
  }
  
  const match = route.path.match(/^\/learning\/words\/(\d+)$/)
  if (match && !route.path.endsWith('/study')) {
    loadWordSetInfo(match[1])
  } else if (route.path === '/learning/words') {
    if (route.query.category_id) {
      const categoryIdParam = route.query.category_id
      const categoryId = typeof categoryIdParam === 'string' 
        ? parseInt(categoryIdParam, 10) 
        : Array.isArray(categoryIdParam) 
          ? parseInt(categoryIdParam[0] as string, 10)
          : null
      currentCategoryId.value = categoryId && !isNaN(categoryId) ? categoryId : null
    }
  }
})
</script>

<style scoped>
.breadcrumbs {
  padding: 12px 0;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border-primary);
}

.breadcrumbs-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  list-style: none;
  margin: 0;
  padding: 0;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
  padding: 4px 0;
}

.breadcrumb-link:hover {
  color: var(--color-primary);
}

.breadcrumb-icon {
  font-size: 16px;
  line-height: 1;
}

.breadcrumb-current {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  padding: 4px 0;
}

.breadcrumb-separator {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.6;
  line-height: 1;
}

@media (max-width: 768px) {
  .breadcrumbs {
    display: none;
  }
}
</style>
