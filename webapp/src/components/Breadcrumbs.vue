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
import { apiClient } from '../api/client'
import Icon from './Icon.vue'

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

// Определяем иерархию маршрутов
const routeHierarchy: Record<string, Breadcrumb[]> = {
  '/dashboard': [
    { label: 'Dashboard', path: '/dashboard' }
  ],
  '/vocab': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Vocabulary', path: '/vocab' }
  ],
  '/learning': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' }
  ],
  '/learning/grammar': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' },
    { label: 'Grammar', path: '/learning/grammar' }
  ],
  '/learning/words': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' },
    { label: 'Word Sets', path: '/learning/words' }
  ],
  '/chat': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Chat', path: '/chat' }
  ],
  '/settings': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Settings', path: '/settings' }
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
  '/admin/db-schema': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'DB Schema', path: '/admin/db-schema' }
  ]
}

const breadcrumbs = computed(() => {
  const currentPath = route.path
  
  // Страницы без крошек
  if (currentPath === '/training' || currentPath.match(/^\/learning\/words\/\d+\/study$/)) {
    return []
  }
  
  // Для динамических маршрутов типа /learning/words/:setId
  if (currentPath.match(/^\/learning\/words\/\d+$/) && !currentPath.endsWith('/study')) {
    const setId = currentPath.split('/').pop()
    const result: Breadcrumb[] = [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Learning', path: '/learning' },
      { label: 'Word Sets', path: '/learning/words' }
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
        { label: 'Dashboard', path: '/dashboard' },
        { label: 'Learning', path: '/learning' },
        { label: 'Word Sets', path: '/learning/words' }
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
  return routeHierarchy[currentPath] || []
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
  } else {
    wordSetInfo.value.wordSet = null
    currentCategoryId.value = null
    // Не очищаем categories, они могут понадобиться
  }
}, { immediate: true })

onMounted(() => {
  // Всегда загружаем категории, если их еще нет, чтобы они были доступны для крошек
  if (wordSetInfo.value.categories.length === 0) {
    loadCategories()
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
