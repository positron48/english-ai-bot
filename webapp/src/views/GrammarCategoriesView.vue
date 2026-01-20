<template>
  <div class="grammar-categories">
    <h1>Grammar Course</h1>
    
    <div v-if="loading" class="loading">
      <p>Loading categories...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadCategories" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else-if="categories.length === 0" class="empty">
      <p>No grammar categories available yet.</p>
    </div>
    
    <div v-else class="categories-grid">
      <router-link
        v-for="category in categories"
        :key="category.section_id"
        :to="`/learning/grammar/${category.section_id}`"
        class="category-card"
      >
        <div class="category-header">
          <h2>{{ category.title }}</h2>
          <span class="category-level">{{ category.level }}</span>
        </div>
        <p class="category-description">{{ category.title }}</p>
        <div class="category-progress">
          <div class="progress-bar">
            <div 
              class="progress-fill" 
              :style="{ width: category.total_chapters > 0 ? `${(category.passed_chapters / category.total_chapters) * 100}%` : '0%' }"
            ></div>
          </div>
          <span class="progress-text">
            {{ category.passed_chapters }} / {{ category.total_chapters }} chapters
          </span>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface Category {
  section_id: string
  title: string
  level: string
  order: number
  published_chapters: number
  passed_chapters: number
  total_chapters: number
}

const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadCategories = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { categories: Category[] } = await apiClient.request('/api/learning/grammar/categories')
    categories.value = data.categories || []
  } catch (err: any) {
    error.value = err.message || 'Failed to load categories'
    console.error('Failed to load grammar categories:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.grammar-categories {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.grammar-categories h1 {
  margin-bottom: 32px;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger);
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
}

.category-card {
  display: flex;
  flex-direction: column;
  padding: 24px;
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  text-decoration: none;
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.category-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.category-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.category-header h2 {
  margin: 0;
  font-size: 20px;
  flex: 1;
}

.category-level {
  padding: 4px 8px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.category-description {
  color: var(--text-secondary);
  font-size: 14px;
  margin: 0 0 16px 0;
}

.category-progress {
  margin-top: auto;
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 8px;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  transition: width 0.3s ease;
}

.progress-text {
  font-size: 12px;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .categories-grid {
    grid-template-columns: 1fr;
  }
}
</style>
