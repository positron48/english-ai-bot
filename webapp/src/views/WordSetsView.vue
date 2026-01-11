<template>
  <div class="word-sets-view">
    <h1>Word Sets Library</h1>
    
    <div class="content-section">
      <div v-if="loading" class="loading">Loading...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else-if="categories.length === 0 && wordSets.length === 0" class="empty-state">
        <p>No items found.</p>
      </div>
      <div v-else class="items-grid">
        <!-- Categories -->
        <div 
          v-for="category in categories" 
          :key="`cat-${category.id}`"
          class="category-card"
          @click="selectCategory(category.id)"
        >
          <div class="card-header">
            <Icon name="folder" class="card-icon" />
            <h3>{{ category.name }}</h3>
          </div>
          <p v-if="category.description" class="card-description">{{ category.description }}</p>
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
          <p v-if="wordSet.description" class="word-set-description">{{ wordSet.description }}</p>
          <div class="word-set-stats">
            <span>{{ wordSet.known_words + wordSet.words_in_vocab }}/{{ wordSet.total_words }} words</span>
            <span v-if="wordSet.unknown_words > 0" class="unknown-count">
              {{ wordSet.unknown_words }} new
            </span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="selectedCategoryId !== null" class="breadcrumb">
      <button @click="goBack" class="breadcrumb-back">
        <Icon name="arrow-left" />
        <span>Back</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

interface Category {
  id: number
  name: string
  description?: string | null
  sort_order: number
  children?: Category[]
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
const selectedCategoryId = ref<number | null>(null)
const currentParentId = ref<number | null>(null)
const categoryHistory = ref<number[]>([]) // Track navigation history

// Items computed is not needed - we'll render categories and wordSets separately in template

onMounted(async () => {
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
  }
  
  await loadCategories()
  await loadWordSets()
})

const loadCategories = async () => {
  try {
    const params = new URLSearchParams()
    if (currentParentId.value !== null) {
      params.append('parent_id', currentParentId.value.toString())
    }
    const data: { categories: Category[] } = await apiClient.request(`/app/learning/words/categories?${params.toString()}`)
    categories.value = data.categories || []
  } catch (error: any) {
    console.error('Failed to load categories:', error)
    categories.value = []
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
    
    const data: { sets: WordSet[] } = await apiClient.request(`/app/learning/words/sets?${params.toString()}`)
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
  }
  selectedCategoryId.value = categoryId
  currentParentId.value = categoryId
  loadCategories()
  loadWordSets()
}

const goBack = () => {
  if (categoryHistory.value.length > 0) {
    // Go back to previous category
    categoryHistory.value.pop()
    const prevCategoryId = categoryHistory.value.length > 0 
      ? categoryHistory.value[categoryHistory.value.length - 1] 
      : null
    selectedCategoryId.value = prevCategoryId
    currentParentId.value = prevCategoryId
  } else {
    // Go to root
    selectedCategoryId.value = null
    currentParentId.value = null
  }
  loadCategories()
  loadWordSets()
}

const viewWordSet = (setId: number) => {
  router.push(`/learning/words/${setId}`)
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
}

.word-sets-view h1 {
  margin-bottom: 24px;
  word-wrap: break-word;
  overflow-wrap: break-word;
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
