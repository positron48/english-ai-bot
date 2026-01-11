<template>
  <div class="admin-word-sets">
    <AdminMenu />
    
    <h2>Word Sets Management</h2>
    
    <div class="admin-tabs-inner">
      <button 
        @click="activeTab = 'categories'" 
        :class="['tab-button', { active: activeTab === 'categories' }]"
      >
        Categories
      </button>
      <button 
        @click="activeTab = 'sets'" 
        :class="['tab-button', { active: activeTab === 'sets' }]"
      >
        Word Sets
      </button>
    </div>
    
    <!-- Categories Tab -->
    <div v-if="activeTab === 'categories'" class="tab-content">
      <div class="section-header">
        <h2>Categories</h2>
        <button @click="startCreateCategory" class="btn btn-primary">Create Category</button>
      </div>
      
      <div v-if="categoriesLoading" class="loading">Loading categories...</div>
      <div v-else-if="categoriesError" class="error">{{ categoriesError }}</div>
      <div v-else-if="categoriesTree.length === 0" class="empty-message">
        <p>No categories found. Create your first category to get started.</p>
      </div>
      <div v-else class="categories-tree">
        <div v-for="category in categoriesTree" :key="category.id" class="category-item">
          <div class="category-row">
            <div class="category-info">
              <strong>{{ category.name }}</strong>
              <span v-if="category.description" class="category-desc">{{ category.description }}</span>
              <span v-if="category.parent_id" class="category-parent">Parent: {{ getCategoryName(category.parent_id) }}</span>
            </div>
            <div class="category-actions">
              <button @click="startEditCategory(category)" class="btn btn-sm btn-primary">Edit</button>
              <button @click="confirmDeleteCategory(category)" class="btn btn-sm btn-danger">Delete</button>
            </div>
          </div>
          <div v-if="category.children && category.children.length > 0" class="category-children">
            <div v-for="child in category.children" :key="child.id" class="category-item category-child">
              <div class="category-row">
                <div class="category-info">
                  <strong>{{ child.name }}</strong>
                  <span v-if="child.description" class="category-desc">{{ child.description }}</span>
                </div>
                <div class="category-actions">
                  <button @click="startEditCategory(child)" class="btn btn-sm btn-primary">Edit</button>
                  <button @click="confirmDeleteCategory(child)" class="btn btn-sm btn-danger">Delete</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Word Sets Tab -->
    <div v-if="activeTab === 'sets'" class="tab-content">
      <div class="section-header">
        <h2>Word Sets</h2>
        <button @click="startCreateWordSet" class="btn btn-primary">Create Word Set</button>
      </div>
      
      <div class="filters">
        <select v-model="selectedCategory" @change="loadWordSets" class="form-select">
          <option :value="null">All Categories</option>
          <option v-for="cat in allCategoriesFlat" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>
      </div>
      
      <div v-if="wordSetsLoading" class="loading">Loading word sets...</div>
      <div v-else-if="wordSetsError" class="error">{{ wordSetsError }}</div>
      <div v-else class="word-sets-list">
        <div v-for="wordSet in wordSets" :key="wordSet.id" class="word-set-item">
          <div class="word-set-header">
            <div class="word-set-info">
              <h3>{{ wordSet.title }}</h3>
              <p v-if="wordSet.description">{{ wordSet.description }}</p>
              <div class="word-set-meta">
                <span>Category: {{ getCategoryName(wordSet.category_id) }}</span>
                <span :class="['published-badge', wordSet.is_published ? 'published' : 'unpublished']">
                  {{ wordSet.is_published ? 'Published' : 'Unpublished' }}
                </span>
              </div>
            </div>
            <div class="word-set-actions">
              <button @click="viewWordSet(wordSet)" class="btn btn-sm btn-secondary">View</button>
              <button @click="startEditWordSet(wordSet)" class="btn btn-sm btn-primary">Edit</button>
              <button @click="confirmDeleteWordSet(wordSet)" class="btn btn-sm btn-danger">Delete</button>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Category Modal -->
    <div v-if="showCategoryModal" class="modal" @click.self="closeCategoryModal">
      <div class="modal-content">
        <div class="modal-header">
          <h3>{{ editingCategory ? 'Edit Category' : 'Create Category' }}</h3>
          <button @click="closeCategoryModal" class="btn-close">&times;</button>
        </div>
        <form @submit.prevent="saveCategory" class="modal-body">
          <div class="form-group">
            <label>Name *</label>
            <input v-model="categoryForm.name" type="text" required class="form-input" />
          </div>
          <div class="form-group">
            <label>Description</label>
            <textarea v-model="categoryForm.description" class="form-textarea" rows="3"></textarea>
          </div>
          <div class="form-group">
            <label>Parent Category</label>
            <select v-model="categoryForm.parent_id" class="form-select">
              <option :value="null">None (Root)</option>
              <option v-for="cat in allCategoriesFlat" :key="cat.id" :value="cat.id">
                {{ cat.name }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Sort Order</label>
            <input v-model.number="categoryForm.sort_order" type="number" class="form-input" />
          </div>
          <div class="modal-actions">
            <button type="submit" class="btn btn-primary">Save</button>
            <button type="button" @click="closeCategoryModal" class="btn btn-secondary">Cancel</button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- Word Set Modal -->
    <div v-if="showWordSetModal" class="modal" @click.self="closeWordSetModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>{{ editingWordSet ? 'Edit Word Set' : 'Create Word Set' }}</h3>
          <button @click="closeWordSetModal" class="btn-close">&times;</button>
        </div>
        <form @submit.prevent="saveWordSet" class="modal-body">
          <div class="form-group">
            <label>Title *</label>
            <input v-model="wordSetForm.title" type="text" required class="form-input" />
          </div>
          <div class="form-group">
            <label>Description</label>
            <textarea v-model="wordSetForm.description" class="form-textarea" rows="3"></textarea>
          </div>
          <div class="form-group">
            <label>Category</label>
            <select v-model="wordSetForm.category_id" class="form-select">
              <option :value="null">None</option>
              <option v-for="cat in allCategoriesFlat" :key="cat.id" :value="cat.id">
                {{ cat.name }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>Sort Order</label>
            <input v-model.number="wordSetForm.sort_order" type="number" class="form-input" />
            <p class="form-hint">Sets are sorted by this value within their category (lower numbers first)</p>
          </div>
          <div class="form-group">
            <label>
              <input v-model="wordSetForm.is_published" type="checkbox" />
              Published
            </label>
          </div>
          <div class="form-group">
            <label>
              Words (comma-separated)
              <span class="word-count">({{ wordSetWordsCount }} words)</span>
            </label>
            <textarea 
              v-model="wordSetForm.words" 
              class="form-textarea" 
              rows="6" 
              placeholder="word1, word2, word3, ..."
            ></textarea>
            <p class="form-hint">Enter words separated by commas. Words will be normalized and duplicates removed.</p>
          </div>
          <div class="modal-actions">
            <button type="submit" class="btn btn-primary" :disabled="wordSetLoading">
              {{ wordSetLoading ? 'Saving...' : 'Save' }}
            </button>
            <button type="button" @click="closeWordSetModal" class="btn btn-secondary" :disabled="wordSetLoading">Cancel</button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- Word Set Items Modal -->
    <div v-if="showItemsModal && editingWordSet" class="modal" @click.self="closeItemsModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Edit Words: {{ editingWordSet.title }}</h3>
          <button @click="closeItemsModal" class="btn-close">&times;</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>
              Words (comma-separated) *
              <span class="word-count">({{ itemsWordsCount }} words)</span>
            </label>
            <textarea 
              v-model="itemsForm.words" 
              class="form-textarea" 
              rows="10" 
              placeholder="word1, word2, word3, ..."
              required
            ></textarea>
            <p class="form-hint">Enter words separated by commas. Words will be normalized and duplicates removed.</p>
          </div>
          <div class="modal-actions">
            <button @click="saveWordSetItems" class="btn btn-primary" :disabled="itemsLoading">
              {{ itemsLoading ? 'Processing...' : 'Save Words' }}
            </button>
            <button type="button" @click="closeItemsModal" class="btn btn-secondary">Cancel</button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Delete Confirm Modals -->
    <div v-if="showDeleteCategoryConfirm" class="modal" @click.self="closeDeleteCategoryConfirm">
      <div class="modal-content">
        <h3>Delete Category</h3>
        <p>Are you sure you want to delete category "{{ categoryToDelete?.name }}"?</p>
        <p class="warning-text">This action cannot be undone.</p>
        <div class="modal-actions">
          <button @click="deleteCategory" class="btn btn-danger">Delete</button>
          <button @click="closeDeleteCategoryConfirm" class="btn btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
    
    <div v-if="showDeleteWordSetConfirm" class="modal" @click.self="closeDeleteWordSetConfirm">
      <div class="modal-content">
        <h3>Delete Word Set</h3>
        <p>Are you sure you want to delete word set "{{ wordSetToDelete?.title }}"?</p>
        <p class="warning-text">This will delete the set and all its items. This action cannot be undone.</p>
        <div class="modal-actions">
          <button @click="deleteWordSet" class="btn btn-danger">Delete</button>
          <button @click="closeDeleteWordSetConfirm" class="btn btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import AdminMenu from '../components/AdminMenu.vue'

interface Category {
  id: number
  parent_id?: number | null
  name: string
  description?: string | null
  sort_order: number
  children?: Category[]
}

interface WordSet {
  id: number
  category_id?: number | null
  title: string
  description?: string | null
  is_published: boolean
  sort_order?: number
  created_at: string
  updated_at: string
}

const activeTab = ref<'categories' | 'sets'>('categories')

// Categories
const categories = ref<Category[]>([])
const categoriesLoading = ref(false)
const categoriesError = ref<string | null>(null)
const showCategoryModal = ref(false)
const editingCategory = ref<Category | null>(null)
const categoryForm = ref({
  name: '',
  description: '',
  parent_id: null as number | null,
  sort_order: 0
})
const showDeleteCategoryConfirm = ref(false)
const categoryToDelete = ref<Category | null>(null)

// Word Sets
const wordSets = ref<WordSet[]>([])
const wordSetsLoading = ref(false)
const wordSetsError = ref<string | null>(null)
const selectedCategory = ref<number | null>(null)
const showWordSetModal = ref(false)
const editingWordSet = ref<WordSet | null>(null)
const wordSetForm = ref({
  title: '',
  description: '',
  category_id: null as number | null,
  is_published: true,
  sort_order: 0,
  words: ''
})
const showDeleteWordSetConfirm = ref(false)
const wordSetToDelete = ref<WordSet | null>(null)
const showItemsModal = ref(false)
const itemsForm = ref({
  words: ''
})
const itemsLoading = ref(false)
const wordSetLoading = ref(false)

// Computed
const categoriesTree = computed(() => {
  const tree: Category[] = []
  const categoryMap = new Map<number, Category>()
  
  // First pass: create map
  categories.value.forEach(cat => {
    categoryMap.set(cat.id, { ...cat, children: [] })
  })
  
  // Second pass: build tree
  categories.value.forEach(cat => {
    const category = categoryMap.get(cat.id)!
    if (cat.parent_id == null) {
      tree.push(category)
    } else {
      const parent = categoryMap.get(cat.parent_id)
      if (parent) {
        if (!parent.children) {
          parent.children = []
        }
        parent.children.push(category)
      }
    }
  })
  
  return tree
})

const allCategoriesFlat = computed(() => {
  const flatten = (cats: Category[]): Category[] => {
    const result: Category[] = []
    cats.forEach(cat => {
      result.push(cat)
      if (cat.children) {
        result.push(...flatten(cat.children))
      }
    })
    return result
  }
  return flatten(categoriesTree.value)
})

const wordSetWordsCount = computed(() => {
  if (!wordSetForm.value.words || !wordSetForm.value.words.trim()) {
    return 0
  }
  return wordSetForm.value.words
    .split(',')
    .map(w => w.trim())
    .filter(w => w.length > 0).length
})

const itemsWordsCount = computed(() => {
  if (!itemsForm.value.words || !itemsForm.value.words.trim()) {
    return 0
  }
  return itemsForm.value.words
    .split(',')
    .map(w => w.trim())
    .filter(w => w.length > 0).length
})

onMounted(async () => {
  await loadCategories()
  await loadWordSets()
})

const loadCategories = async () => {
  categoriesLoading.value = true
  categoriesError.value = null
  try {
    const data: any = await apiClient.request('/app/admin/word-set-categories')
    // Handle both snake_case and PascalCase field names
    categories.value = (data.categories || []).map((cat: any) => ({
      id: cat.id || cat.ID,
      parent_id: cat.parent_id !== undefined ? cat.parent_id : (cat.ParentID !== undefined ? cat.ParentID : null),
      name: cat.name || cat.Name || '',
      description: cat.description !== undefined ? cat.description : (cat.Description !== undefined ? cat.Description : null),
      sort_order: cat.sort_order !== undefined ? cat.sort_order : (cat.SortOrder !== undefined ? cat.SortOrder : 0)
    }))
  } catch (error: any) {
    console.error('Failed to load categories:', error)
    categoriesError.value = error.message || 'Failed to load categories'
  } finally {
    categoriesLoading.value = false
  }
}

const loadWordSets = async () => {
  wordSetsLoading.value = true
  wordSetsError.value = null
  try {
    const params = new URLSearchParams()
    if (selectedCategory.value !== null) {
      params.append('category_id', selectedCategory.value.toString())
    }
    
    const data: { word_sets: WordSet[] } = await apiClient.request(`/app/admin/word-sets?${params.toString()}`)
    wordSets.value = data.word_sets || []
  } catch (error: any) {
    console.error('Failed to load word sets:', error)
    wordSetsError.value = error.message || 'Failed to load word sets'
  } finally {
    wordSetsLoading.value = false
  }
}

const getCategoryName = (categoryId: number | null | undefined): string => {
  if (!categoryId) return 'None'
  const cat = allCategoriesFlat.value.find(c => c.id === categoryId)
  return cat?.name || `Category #${categoryId}`
}

// Category CRUD
const startCreateCategory = () => {
  editingCategory.value = null
  categoryForm.value = {
    name: '',
    description: '',
    parent_id: null,
    sort_order: 0
  }
  showCategoryModal.value = true
}

const startEditCategory = (category: Category) => {
  editingCategory.value = category
  categoryForm.value = {
    name: category.name,
    description: category.description || '',
    parent_id: category.parent_id || null,
    sort_order: category.sort_order
  }
  showCategoryModal.value = true
}

const closeCategoryModal = () => {
  showCategoryModal.value = false
  editingCategory.value = null
}

const saveCategory = async () => {
  try {
    const url = editingCategory.value
      ? `/app/admin/word-set-categories/${editingCategory.value.id}`
      : '/app/admin/word-set-categories'
    
    const method = editingCategory.value ? 'PUT' : 'POST'
    
    await apiClient.request(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(categoryForm.value)
    })
    
    closeCategoryModal()
    await loadCategories()
    await showAlert('Category saved successfully')
  } catch (error: any) {
    console.error('Failed to save category:', error)
    await showAlert(error.message || 'Failed to save category')
  }
}

const confirmDeleteCategory = (category: Category) => {
  categoryToDelete.value = category
  showDeleteCategoryConfirm.value = true
}

const closeDeleteCategoryConfirm = () => {
  showDeleteCategoryConfirm.value = false
  categoryToDelete.value = null
}

const deleteCategory = async () => {
  if (!categoryToDelete.value) return
  
  try {
    await apiClient.request(`/app/admin/word-set-categories/${categoryToDelete.value.id}`, {
      method: 'DELETE'
    })
    
    closeDeleteCategoryConfirm()
    await loadCategories()
  } catch (error: any) {
    console.error('Failed to delete category:', error)
    await showAlert(error.message || 'Failed to delete category')
  }
}

// Word Set CRUD
const startCreateWordSet = () => {
  editingWordSet.value = null
  wordSetForm.value = {
    title: '',
    description: '',
    category_id: null,
    is_published: true,
    sort_order: 0,
    words: ''
  }
  showWordSetModal.value = true
}

const startEditWordSet = async (wordSet: WordSet) => {
  editingWordSet.value = wordSet
  
  wordSetForm.value = {
    title: wordSet.title,
    description: wordSet.description || '',
    category_id: wordSet.category_id || null,
    is_published: wordSet.is_published,
    sort_order: wordSet.sort_order || 0,
    words: ''
  }
  
  // Load words for editing (same as in viewWordSet)
  try {
    const data: any = await apiClient.request(`/app/admin/word-sets/${wordSet.id}`)
    const words = data.words || []
    wordSetForm.value.words = Array.isArray(words) && words.length > 0 
      ? words.map((w: any) => w.word).join(', ')
      : ''
  } catch (error: any) {
    console.error('Failed to load words for editing:', error)
    // Continue anyway - words field will be empty
  }
  
  showWordSetModal.value = true
}

const closeWordSetModal = () => {
  showWordSetModal.value = false
  editingWordSet.value = null
}

const saveWordSet = async () => {
  if (wordSetLoading.value) return
  
  wordSetLoading.value = true
  try {
    const url = editingWordSet.value
      ? `/app/admin/word-sets/${editingWordSet.value.id}`
      : '/app/admin/word-sets'
    
    const method = editingWordSet.value ? 'PUT' : 'POST'
    
    // Prepare data without words field
    const { words, ...wordSetData } = wordSetForm.value
    const dataToSend = {
      ...wordSetData
    }
    
    const response: any = await apiClient.request(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(dataToSend)
    })
    
    // Get the word set ID (either from response or existing)
    const wordSetId = response.id || editingWordSet.value?.id
    
    // If words were provided, save them
    if (words && words.trim()) {
      if (wordSetId) {
        await apiClient.request(`/app/admin/word-sets/${wordSetId}/items`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ words: words.trim() })
        })
      }
    }
    
    closeWordSetModal()
    await loadWordSets()
    await showAlert('Word set saved successfully')
  } catch (error: any) {
    console.error('Failed to save word set:', error)
    await showAlert(error.message || 'Failed to save word set')
  } finally {
    wordSetLoading.value = false
  }
}

const viewWordSet = async (wordSet: WordSet) => {
  try {
    const data: any = await apiClient.request(`/app/admin/word-sets/${wordSet.id}`)
    
    editingWordSet.value = data.word_set
    // Handle both null and empty array cases
    const words = data.words || []
    itemsForm.value.words = Array.isArray(words) && words.length > 0 
      ? words.map((w: any) => w.word).join(', ')
      : ''
    showItemsModal.value = true
  } catch (error: any) {
    console.error('Failed to load word set:', error)
    await showAlert(error.message || 'Failed to load word set')
  }
}

const closeItemsModal = () => {
  showItemsModal.value = false
  editingWordSet.value = null
  itemsForm.value.words = ''
}

const saveWordSetItems = async () => {
  if (!editingWordSet.value) return
  
  itemsLoading.value = true
  try {
    await apiClient.request(`/app/admin/word-sets/${editingWordSet.value.id}/items`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ words: itemsForm.value.words })
    })
    
    closeItemsModal()
    await showAlert('Words saved successfully')
  } catch (error: any) {
    console.error('Failed to save words:', error)
    await showAlert(error.message || 'Failed to save words')
  } finally {
    itemsLoading.value = false
  }
}

const confirmDeleteWordSet = (wordSet: WordSet) => {
  wordSetToDelete.value = wordSet
  showDeleteWordSetConfirm.value = true
}

const closeDeleteWordSetConfirm = () => {
  showDeleteWordSetConfirm.value = false
  wordSetToDelete.value = null
}

const deleteWordSet = async () => {
  if (!wordSetToDelete.value) return
  
  try {
    await apiClient.request(`/app/admin/word-sets/${wordSetToDelete.value.id}`, {
      method: 'DELETE'
    })
    
    closeDeleteWordSetConfirm()
    await loadWordSets()
  } catch (error: any) {
    console.error('Failed to delete word set:', error)
    await showAlert(error.message || 'Failed to delete word set')
  }
}
</script>

<style scoped>
.admin-word-sets {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  overflow-x: hidden;
  width: 100%;
  box-sizing: border-box;
}

.admin-tabs-inner {
  display: flex;
  gap: 0;
  margin-bottom: 24px;
  border-bottom: 2px solid var(--border-primary);
}

.tab-button {
  padding: 12px 24px;
  background: transparent;
  border: none;
  border-bottom: 3px solid transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
  position: relative;
  bottom: -2px;
}

.tab-button:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.tab-button.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
  font-weight: 500;
}

.tab-content {
  margin-top: 24px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filters {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.form-select,
.form-input {
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
}

.form-select {
  min-width: 200px;
}

.form-input {
  flex: 1;
  min-width: 200px;
}

.categories-tree {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.category-item {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 16px;
  background: var(--card-bg);
}

.category-child {
  margin-left: 32px;
  background: var(--bg-secondary);
}

.category-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.category-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.category-desc {
  color: var(--text-secondary);
  font-size: 14px;
}

.category-parent {
  color: var(--text-secondary);
  font-size: 12px;
}

.category-actions {
  display: flex;
  gap: 8px;
}

.category-children {
  margin-top: 12px;
  padding-left: 16px;
  border-left: 2px solid var(--border-primary);
}

.word-sets-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.word-set-item {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 16px;
  background: var(--card-bg);
}

.word-set-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  min-width: 0;
  flex-wrap: wrap;
}

.word-set-info {
  flex: 1;
  min-width: 0;
}

.word-set-info h3 {
  margin: 0 0 8px 0;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.word-set-info p {
  margin: 0 0 8px 0;
  color: var(--text-secondary);
}

.word-set-meta {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
}

.published-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.published-badge.published {
  background: var(--color-success, #10b981);
  color: white;
}

.published-badge.unpublished {
  background: var(--color-secondary, #6b7280);
  color: white;
}

.word-set-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  flex-wrap: wrap;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-bg);
  padding: 30px;
  border-radius: 8px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-large {
  max-width: 800px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.btn-close {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: var(--text-primary);
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.btn-close:hover {
  background-color: var(--bg-hover);
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  color: var(--text-primary);
}

.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  font-family: inherit;
  resize: vertical;
}

.form-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}

.word-count {
  font-weight: normal;
  color: var(--text-secondary);
  font-size: 0.9em;
  margin-left: 8px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  justify-content: flex-end;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background-color: var(--color-primary, #3b82f6);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover, #2563eb);
}

.btn-secondary {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.btn-secondary:hover {
  background-color: var(--color-secondary-hover, #4b5563);
}

.btn-danger {
  background-color: var(--color-danger, #ef4444);
  color: white;
}

.btn-danger:hover {
  background-color: var(--color-danger-hover, #dc2626);
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger, #ef4444);
}

.empty-message {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.warning-text {
  color: var(--color-danger, #ef4444);
  font-size: 13px;
  margin-top: 8px;
}

@media (max-width: 768px) {
  .admin-word-sets {
    padding: 12px;
  }
  
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .section-header h2 {
    font-size: 20px;
  }
  
  .word-set-header {
    flex-direction: column;
    gap: 12px;
  }
  
  .word-set-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .word-set-actions .btn {
    flex: 1;
    min-width: 0;
  }
  
  .word-set-meta {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
  
  .word-set-info h3 {
    font-size: 16px;
  }
  
  .word-set-info p {
    font-size: 13px;
  }
  
  .modal-content {
    padding: 20px;
    width: 95%;
    max-width: 95%;
  }
  
  .modal-large {
    max-width: 95%;
  }
  
  .categories-tree {
    padding-left: 0;
  }
  
  .category-item {
    padding: 12px;
  }
  
  .category-child {
    margin-left: 16px;
  }
  
  .category-row {
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .category-actions {
    width: 100%;
    justify-content: flex-start;
  }
}

@media (max-width: 480px) {
  .admin-word-sets {
    padding: 8px;
  }
  
  .word-set-actions {
    flex-direction: column;
  }
  
  .word-set-actions .btn {
    width: 100%;
  }
  
  .section-header .btn {
    width: 100%;
  }
  
  .filters {
    width: 100%;
  }
  
  .form-select {
    width: 100%;
  }
}

</style>
