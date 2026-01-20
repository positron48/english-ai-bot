<template>
  <div class="grammar-categories">
    <h1>Grammar Course</h1>
    
    <!-- Statistics Block -->
    <div v-if="!loading && !error && statistics" class="statistics-block">
      <div class="stats-content">
        <!-- Current Level (Left) -->
        <div class="stat-item level-item">
          <div class="stat-label">Level</div>
          <div class="level-badge-compact" :class="levelBadgeClass">
            {{ statistics.confirmed_level || 'Not started' }}
          </div>
        </div>
        
        <!-- Course Completion Percentage -->
        <div class="stat-item percentage-item">
          <div class="stat-label">Course</div>
          <div class="percentage-wrapper">
            <div class="percentage-circle-small-wrapper">
              <svg class="percentage-circle-small" viewBox="0 0 60 60">
                <circle
                  class="percentage-circle-small-bg"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  stroke="var(--bg-tertiary)"
                  stroke-width="4"
                />
                <circle
                  class="percentage-circle-small-outline"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  :stroke="getPercentageColor(statistics.course_completion_pct)"
                  stroke-width="4"
                  stroke-opacity="0.2"
                />
                <circle
                  class="percentage-circle-small-fill"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  :stroke="getPercentageColor(statistics.course_completion_pct)"
                  stroke-width="4"
                  stroke-linecap="round"
                  :style="{
                    strokeDasharray: smallCircleCircumference,
                    strokeDashoffset: getPercentageOffset(statistics.course_completion_pct)
                  }"
                />
              </svg>
              <div class="percentage-value-small">{{ statistics.course_completion_pct }}%</div>
            </div>
          </div>
        </div>
        
        <!-- Average Test Score -->
        <div class="stat-item percentage-item">
          <div class="stat-label">Tests</div>
          <div class="percentage-wrapper">
            <div class="percentage-circle-small-wrapper">
              <svg class="percentage-circle-small" viewBox="0 0 60 60">
                <circle
                  class="percentage-circle-small-bg"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  stroke="var(--bg-tertiary)"
                  stroke-width="4"
                />
                <circle
                  class="percentage-circle-small-outline"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  :stroke="getPercentageColor(statistics.average_test_score || 0)"
                  stroke-width="4"
                  stroke-opacity="0.2"
                />
                <circle
                  class="percentage-circle-small-fill"
                  cx="30"
                  cy="30"
                  r="26"
                  fill="none"
                  :stroke="getPercentageColor(statistics.average_test_score || 0)"
                  stroke-width="4"
                  stroke-linecap="round"
                  :style="{
                    strokeDasharray: smallCircleCircumference,
                    strokeDashoffset: getPercentageOffset(statistics.average_test_score || 0)
                  }"
                />
              </svg>
              <div class="percentage-value-small">{{ statistics.average_test_score || 0 }}%</div>
            </div>
          </div>
        </div>
        
        <!-- Chapters Progress (Right) -->
        <div class="stat-item chapters-item">
          <div class="stat-label">Chapters</div>
          <div class="chapters-value-compact">
            <span class="chapters-number">{{ passedChapters }}</span>
            <span class="chapters-separator">/</span>
            <span class="chapters-total">{{ totalChapters }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="loading" class="loading">
      <p>Loading categories...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadCategories" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else-if="!categories || categories.length === 0" class="empty">
      <p>No grammar categories available yet.</p>
    </div>
    
    <div v-else-if="categories && categories.length > 0" class="categories-grid">
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
                :style="{ width: `${categoryProgressPercentage(category)}%` }"
              ></div>
            </div>
            <span class="progress-text">
              {{ categoryProgressPercentage(category) }}% complete
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
                :style="{ width: `${categoryProgressPercentage(category)}%` }"
              ></div>
            </div>
            <span class="progress-text">
              {{ categoryProgressPercentage(category) }}% complete
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
import { ref, computed, onMounted } from 'vue'
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
  progress_percentage?: number
  can_access?: boolean
}

const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const statistics = ref<{ confirmed_level: string; course_completion_pct: number; average_test_score?: number } | null>(null)

// Computed properties for statistics
const totalChapters = computed(() => {
  return categories.value.reduce((sum, cat) => sum + cat.total_chapters, 0)
})

const passedChapters = computed(() => {
  return categories.value.reduce((sum, cat) => sum + cat.passed_chapters, 0)
})

const availableCategories = computed(() => {
  return categories.value.filter(cat => cat.can_access).length
})


const levelBadgeClass = computed(() => {
  const level = statistics.value?.confirmed_level || ''
  if (level.startsWith('C')) return 'badge-c2'
  if (level.startsWith('B')) return 'badge-b'
  if (level.startsWith('A')) return 'badge-a'
  return 'badge-none'
})

const levelDescription = computed(() => {
  const level = statistics.value?.confirmed_level || ''
  if (!level || level === 'Not started') return 'Start learning to unlock your level'
  if (level.startsWith('C')) return 'Advanced proficiency'
  if (level.startsWith('B')) return 'Intermediate level'
  if (level.startsWith('A')) return 'Beginner level'
  return 'Keep learning!'
})

// Small circle calculations
const smallCircleCircumference = computed(() => 2 * Math.PI * 26)

const getPercentageOffset = (percent: number): number => {
  const progress = Math.max(0, Math.min(100, percent)) / 100
  return smallCircleCircumference.value * (1 - progress)
}

const getPercentageColor = (percent: number): string => {
  if (percent >= 90) return '#10b981' // green
  if (percent >= 70) return '#3b82f6' // blue
  if (percent >= 50) return '#f59e0b' // orange
  if (percent >= 25) return '#f97316' // orange-red
  return '#ef4444' // red
}

// Calculate category progress percentage
// Use progress_percentage from backend if available, otherwise fallback to passed_chapters calculation
const categoryProgressPercentage = (category: Category): number => {
  if (category.progress_percentage !== undefined) {
    return category.progress_percentage
  }
  // Fallback: calculate based on passed chapters
  if (category.total_chapters === 0) return 0
  return Math.round((category.passed_chapters / category.total_chapters) * 100)
}

const loadCategories = async () => {
  loading.value = true
  error.value = null
  try {
    const [categoriesData, statsData] = await Promise.all([
      apiClient.request('/api/learning/grammar/categories'),
      apiClient.request('/api/learning/grammar/statistics')
    ])
    
    const loadedCategories = (categoriesData as { categories: Category[] }).categories || []
    
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
    statistics.value = statsData as { confirmed_level: string; course_completion_pct: number; average_test_score?: number }
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

.statistics-block {
  margin-bottom: 24px;
  padding: 12px 16px;
  background: linear-gradient(135deg, var(--card-bg) 0%, rgba(var(--color-primary-rgb, 59, 130, 246), 0.05) 100%);
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.stats-content {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  align-items: center;
}

.stat-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.stat-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
  min-width: 50px;
}

/* Level Item */
.level-item {
  justify-content: flex-start;
}

.level-badge-compact {
  padding: 6px 12px;
  font-size: 20px;
  font-weight: 700;
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.level-badge-compact.badge-a {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.level-badge-compact.badge-b {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.level-badge-compact.badge-c2 {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.level-badge-compact.badge-none {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

/* Percentage Items */
.percentage-item {
  justify-content: flex-start;
}

.percentage-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.percentage-circle-small-wrapper {
  position: relative;
  width: 60px;
  height: 60px;
}

.percentage-circle-small {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.percentage-circle-small-bg {
  opacity: 0.2;
}

.percentage-circle-small-outline {
  transition: stroke 0.3s ease;
}

.percentage-circle-small-fill {
  transition: stroke-dashoffset 0.6s ease, stroke 0.3s ease;
}

.percentage-value-small {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  text-align: center;
}

/* Chapters Item */
.chapters-item {
  justify-content: flex-start;
}

.chapters-value-compact {
  display: flex;
  align-items: baseline;
  gap: 4px;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.chapters-value-compact .chapters-number {
  color: var(--color-primary);
}

.chapters-value-compact .chapters-separator {
  color: var(--text-secondary);
  font-weight: 400;
}

.chapters-value-compact .chapters-total {
  color: var(--text-secondary);
  font-weight: 500;
}

@media (max-width: 768px) {
  .stats-content {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  
  .percentage-circle-small-wrapper {
    width: 50px;
    height: 50px;
  }
  
  .percentage-value-small {
    font-size: 12px;
  }
  
  .chapters-value-compact {
    font-size: 18px;
  }
  
  .level-badge-compact {
    font-size: 18px;
    padding: 5px 10px;
  }
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
