<template>
  <div class="admin-grammar">
    <h1>Grammar Course Management</h1>
    
    <div v-if="loading" class="loading">
      <p>Loading...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadCategories" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else>
      <div class="search-bar">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search categories..."
          class="search-input"
        />
      </div>
      
      <div class="categories-list">
        <div
          v-for="category in filteredCategories"
          :key="category.section_id"
          class="category-item"
        >
          <div class="category-header">
            <div class="category-info">
              <h3>
                <input
                  v-model="category.custom_name"
                  @blur="saveCategoryName(category)"
                  @keyup.enter="saveCategoryName(category)"
                  class="name-input"
                  :placeholder="category.title"
                />
              </h3>
              <div class="category-meta">
                <span class="meta-badge">{{ category.level }}</span>
                <span class="meta-text">
                  {{ category.published_chapters }}/{{ category.total_chapters }} chapters published
                </span>
              </div>
            </div>
            <div class="category-actions">
              <label class="toggle-switch">
                <input
                  type="checkbox"
                  :checked="category.is_published"
                  @change="toggleCategoryPublish(category)"
                />
                <span class="toggle-slider"></span>
              </label>
              <span class="toggle-label">{{ category.is_published ? 'Published' : 'Hidden' }}</span>
              <button
                @click="toggleCategory(category.section_id)"
                class="btn btn-sm"
              >
                {{ category.expanded ? 'Collapse' : 'Expand' }}
              </button>
            </div>
          </div>
          
          <div v-if="category.expanded" class="category-chapters">
            <div v-if="category.chaptersLoading" class="loading-small">
              Loading chapters...
            </div>
            <div v-else class="chapters-list">
              <div
                v-for="chapter in category.chapters"
                :key="chapter.chapter_id"
                class="chapter-item"
              >
                <div class="chapter-info">
                  <input
                    v-model="chapter.custom_name"
                    @blur="saveChapterName(chapter)"
                    @keyup.enter="saveChapterName(chapter)"
                    class="name-input"
                    :placeholder="chapter.title"
                    :class="{ 'error': !chapter.file_exists }"
                  />
                  <span v-if="!chapter.file_exists" class="file-warning">
                    ⚠ File not found
                  </span>
                </div>
                <div class="chapter-actions">
                  <label class="toggle-switch">
                    <input
                      type="checkbox"
                      :checked="chapter.is_published"
                      :disabled="!chapter.file_exists"
                      @change="toggleChapterPublish(chapter)"
                    />
                    <span class="toggle-slider"></span>
                  </label>
                  <span class="toggle-label">{{ chapter.is_published ? 'Published' : 'Hidden' }}</span>
                </div>
              </div>
            </div>
            <div class="category-bulk-actions">
              <button
                @click="bulkPublishChapters(category, true)"
                class="btn btn-sm btn-primary"
              >
                Publish All Chapters
              </button>
              <button
                @click="bulkPublishChapters(category, false)"
                class="btn btn-sm btn-secondary"
              >
                Hide All Chapters
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface Category {
  section_id: string
  title: string
  level: string
  order: number
  is_published: boolean
  custom_name?: string | null
  total_chapters: number
  available_chapters: number
  published_chapters: number
  expanded?: boolean
  chapters?: Chapter[]
  chaptersLoading?: boolean
}

interface Chapter {
  chapter_id: string
  title: string
  is_published: boolean
  custom_name?: string | null
  file_exists: boolean
}

const categories = ref<Category[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const searchQuery = ref('')

const filteredCategories = computed(() => {
  if (!searchQuery.value) return categories.value
  const query = searchQuery.value.toLowerCase()
  return categories.value.filter(cat => 
    cat.title.toLowerCase().includes(query) ||
    cat.section_id.toLowerCase().includes(query) ||
    (cat.custom_name && cat.custom_name.toLowerCase().includes(query))
  )
})

const loadCategories = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { categories: Category[] } = await apiClient.request('/api/admin/grammar/categories')
    categories.value = (data.categories || []).map(cat => ({
      ...cat,
      expanded: false,
      chapters: undefined,
      chaptersLoading: false
    }))
  } catch (err: any) {
    error.value = err.message || 'Failed to load categories'
    console.error('Failed to load grammar categories:', err)
  } finally {
    loading.value = false
  }
}

const toggleCategory = async (sectionId: string) => {
  const category = categories.value.find(c => c.section_id === sectionId)
  if (!category) return
  
  category.expanded = !category.expanded
  
  if (category.expanded && !category.chapters) {
    await loadChapters(sectionId)
  }
}

const loadChapters = async (sectionId: string) => {
  const category = categories.value.find(c => c.section_id === sectionId)
  if (!category) return
  
  category.chaptersLoading = true
  try {
    const data: { chapters: Chapter[] } = await apiClient.request(
      `/api/admin/grammar/chapters?section_id=${sectionId}`
    )
    category.chapters = data.chapters || []
  } catch (err: any) {
    console.error('Failed to load chapters:', err)
    category.chapters = []
  } finally {
    category.chaptersLoading = false
  }
}

const toggleCategoryPublish = async (category: Category) => {
  try {
    await apiClient.request(`/api/admin/grammar/categories/${category.section_id}/publish`, {
      method: 'POST',
      body: {
        is_published: !category.is_published,
        cascade: false
      }
    })
    category.is_published = !category.is_published
  } catch (err: any) {
    console.error('Failed to toggle category publish:', err)
    error.value = err.message || 'Failed to update category'
  }
}

const toggleChapterPublish = async (chapter: Chapter) => {
  try {
    await apiClient.request(`/api/admin/grammar/chapters/${chapter.chapter_id}/publish`, {
      method: 'POST',
      body: {
        is_published: !chapter.is_published
      }
    })
    chapter.is_published = !chapter.is_published
  } catch (err: any) {
    console.error('Failed to toggle chapter publish:', err)
    error.value = err.message || 'Failed to update chapter'
  }
}

const saveCategoryName = async (category: Category) => {
  try {
    await apiClient.request(`/api/admin/grammar/items/section/${category.section_id}/rename`, {
      method: 'POST',
      body: {
        name: category.custom_name || null
      }
    })
  } catch (err: any) {
    console.error('Failed to save category name:', err)
    error.value = err.message || 'Failed to save name'
  }
}

const saveChapterName = async (chapter: Chapter) => {
  try {
    await apiClient.request(`/api/admin/grammar/items/chapter/${chapter.chapter_id}/rename`, {
      method: 'POST',
      body: {
        name: chapter.custom_name || null
      }
    })
  } catch (err: any) {
    console.error('Failed to save chapter name:', err)
    error.value = err.message || 'Failed to save name'
  }
}

const bulkPublishChapters = async (category: Category, publish: boolean) => {
  if (!category.chapters) return
  
  try {
    await apiClient.request(`/api/admin/grammar/categories/${category.section_id}/publish`, {
      method: 'POST',
      body: {
        is_published: publish,
        cascade: true
      }
    })
    
    // Update local state
    category.chapters.forEach(ch => {
      ch.is_published = publish
    })
    category.published_chapters = publish ? category.total_chapters : 0
  } catch (err: any) {
    console.error('Failed to bulk publish chapters:', err)
    error.value = err.message || 'Failed to update chapters'
  }
}

onMounted(() => {
  loadCategories()
})
</script>

<style scoped>
.admin-grammar {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error {
  text-align: center;
  padding: 40px 20px;
}

.error {
  color: var(--color-danger);
}

.search-bar {
  margin-bottom: 24px;
}

.search-input {
  width: 100%;
  max-width: 400px;
  padding: 12px;
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  background: var(--input-bg);
  color: var(--text-primary);
  font-size: 16px;
}

.categories-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.category-item {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
}

.category-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
}

.category-info {
  flex: 1;
}

.category-info h3 {
  margin: 0 0 8px 0;
}

.name-input {
  width: 100%;
  padding: 8px 12px;
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  background: var(--input-bg);
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
}

.name-input.error {
  border-color: var(--color-danger);
}

.category-meta {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-top: 8px;
}

.meta-badge {
  padding: 4px 8px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.meta-text {
  font-size: 14px;
  color: var(--text-secondary);
}

.category-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--border-primary);
  transition: 0.3s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

.toggle-switch input:checked + .toggle-slider {
  background-color: var(--color-primary);
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(26px);
}

.toggle-switch input:disabled + .toggle-slider {
  opacity: 0.5;
  cursor: not-allowed;
}

.toggle-label {
  font-size: 14px;
  color: var(--text-secondary);
  min-width: 80px;
}

.category-chapters {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 2px solid var(--border-primary);
}

.loading-small {
  text-align: center;
  padding: 20px;
  color: var(--text-secondary);
}

.chapters-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.chapter-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
}

.chapter-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
}

.chapter-info .name-input {
  font-size: 14px;
  font-weight: normal;
}

.file-warning {
  color: var(--color-danger);
  font-size: 12px;
}

.chapter-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-bulk-actions {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 14px;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 2px solid var(--border-primary);
}

.btn-secondary:hover {
  border-color: var(--color-primary);
}
</style>
