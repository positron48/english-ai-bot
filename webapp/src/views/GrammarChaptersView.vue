<template>
  <div class="grammar-chapters">
    <div v-if="loading" class="loading">
      <p>Loading chapters...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadChapters" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else>
      <div class="chapters-header">
        <h1>{{ categoryTitle }}</h1>
        <button 
          v-if="allChaptersPassed && chapters.length > 0"
          @click="startCategoryTest"
          class="btn btn-primary"
        >
          Category Test
        </button>
      </div>
      
      <div v-if="chapters.length === 0" class="empty">
        <p>No chapters available in this category.</p>
      </div>
      
      <div v-else class="chapters-list">
        <div
          v-for="chapter in chapters"
          :key="chapter.chapter_id"
          class="chapter-item"
        >
          <router-link
            :to="`/learning/grammar/chapter/${chapter.chapter_id}`"
            class="chapter-link"
          >
            <div class="chapter-info">
              <h3>{{ chapter.title }}</h3>
              <div class="chapter-meta">
                <span v-if="chapter.level" class="chapter-level">{{ chapter.level }}</span>
                <span v-if="chapter.estimated_minutes" class="chapter-time">
                  ~{{ chapter.estimated_minutes }} min
                </span>
              </div>
            </div>
            <div class="chapter-status">
              <div v-if="chapter.passed" class="status-badge passed">
                <Icon name="check" />
                Passed
              </div>
              <div v-else-if="chapter.best_score > 0" class="status-badge attempted">
                Score: {{ chapter.best_score }}%
              </div>
              <Icon name="chevron-right" class="chevron" />
            </div>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

interface Chapter {
  chapter_id: string
  title: string
  title_short?: string
  description?: string
  level?: string
  order: number
  estimated_minutes?: number
  best_score: number
  passed: boolean
  last_attempt_at?: string
}

const route = useRoute()
const router = useRouter()
const sectionId = computed(() => route.params.sectionId as string)

const chapters = ref<Chapter[]>([])
const categoryTitle = ref('')
const loading = ref(true)
const error = ref<string | null>(null)

const allChaptersPassed = computed(() => {
  return chapters.value.length > 0 && chapters.value.every(ch => ch.passed)
})

const loadChapters = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { chapters: Chapter[] } = await apiClient.request(
      `/api/learning/grammar/categories/${sectionId.value}/chapters`
    )
    chapters.value = data.chapters || []
    
    // Try to get category title from first chapter or use section ID
    if (chapters.value.length > 0) {
      // Extract section title from section_id or use a default
      categoryTitle.value = sectionId.value.split('.').pop()?.replace(/_/g, ' ') || sectionId.value
    } else {
      categoryTitle.value = sectionId.value
    }
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

.chevron {
  color: var(--text-tertiary);
}

@media (max-width: 768px) {
  .chapters-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
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
