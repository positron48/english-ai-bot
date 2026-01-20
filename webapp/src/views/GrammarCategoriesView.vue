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
      <div
        v-for="category in categories"
        :key="category.section_id"
        class="category-card"
        :class="{ 'locked': !category.can_access }"
      >
        <router-link
          v-if="category.can_access"
          :to="`/learning/grammar/${category.section_id}`"
          class="category-link"
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
        <div v-else class="category-link locked-link">
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
          <div class="locked-overlay">
            <Icon name="lock" />
            <span>Complete previous chapter to unlock</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

const route = useRoute()

interface Category {
  section_id: string
  title: string
  level: string
  order: number
  published_chapters: number
  passed_chapters: number
  total_chapters: number
  can_access?: boolean
}

const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadCategories = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { categories: Category[] } = await apiClient.request('/api/learning/grammar/categories')
    const loadedCategories = data.categories || []
    
    // Check access for each category
    for (let i = 0; i < loadedCategories.length; i++) {
      const category = loadedCategories[i]
      // First category is always accessible
      if (i === 0) {
        category.can_access = true
      } else {
        // Category is accessible if all chapters in previous category were passed
        const previousCategory = loadedCategories[i - 1]
        category.can_access = previousCategory.passed_chapters === previousCategory.total_chapters && previousCategory.total_chapters > 0
      }
    }
    
    categories.value = loadedCategories
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
  color: var(--text-primary);
  transition: all 0.3s ease;
  position: relative;
}

.category-card.locked {
  opacity: 0.6;
}

.category-link {
  text-decoration: none;
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
  flex: 1;
}

.category-card:not(.locked):hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.category-card.locked .category-link,
.category-card .locked-link {
  pointer-events: none;
  cursor: not-allowed;
}

.locked-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: rgba(0, 0, 0, 0.7);
  border-radius: 12px;
  color: white;
  font-weight: 600;
}

.locked-overlay Icon {
  font-size: 24px;
  margin-bottom: -2px;
}

.locked-overlay span {
  margin-top: -2px;
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
