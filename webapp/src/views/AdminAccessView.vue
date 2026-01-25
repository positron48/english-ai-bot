<template>
  <div class="admin-content">
    <h2>Access Control</h2>
    
    <div class="admin-tabs-inner">
      <button 
        @click="switchTab('categories')" 
        :class="['tab-button', { active: activeTab === 'categories' }]"
      >
        Categories
      </button>
      <button 
        @click="switchTab('users')" 
        :class="['tab-button', { active: activeTab === 'users' }]"
      >
        User Assignments
      </button>
    </div>
    
    <!-- Categories Tab -->
    <div v-if="activeTab === 'categories'" class="tab-content">
      <div class="section-header">
        <h2>User Access Categories</h2>
        <button @click="startCreateCategory" class="btn btn-primary">Create Category</button>
      </div>
      
      <div v-if="categoriesLoading" class="loading">Loading categories...</div>
      <div v-else-if="categoriesError" class="error">{{ categoriesError }}</div>
      <div v-else-if="categories.length === 0" class="empty-message">
        <p>No categories found. Create your first category to get started.</p>
      </div>
      <div v-else class="categories-list">
        <div v-for="category in categories" :key="category.id" class="category-item">
          <div class="category-row">
            <div class="category-info">
              <strong>{{ category.name }}</strong>
              <span v-if="category.description" class="category-desc">{{ category.description }}</span>
              <div class="permissions-list">
                <span class="permissions-label">Permissions:</span>
                <span v-if="category.permissions && category.permissions.length > 0" class="permissions-tags">
                  <span v-for="perm in category.permissions" :key="perm" class="permission-tag">{{ perm }}</span>
                </span>
                <span v-else class="no-permissions">No permissions assigned</span>
              </div>
            </div>
            <div class="category-actions">
              <button @click="startEditCategory(category)" class="btn btn-sm btn-primary">Edit</button>
              <button @click="startEditPermissions(category)" class="btn btn-sm btn-secondary">Permissions</button>
              <button @click="confirmDeleteCategory(category)" class="btn btn-sm btn-danger">Delete</button>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Users Tab -->
    <div v-if="activeTab === 'users'" class="tab-content">
      <div class="section-header">
        <h2>User Category Assignments</h2>
      </div>
      
      <div class="user-select-section">
        <label>Select User:</label>
        <select v-model="selectedUserId" @change="loadUserCategories" class="form-select">
          <option :value="null">Choose a user...</option>
          <option v-for="user in users" :key="user.id" :value="user.id">
            {{ user.telegram_username || `User #${user.telegram_id}` }} (ID: {{ user.id }})
          </option>
        </select>
      </div>
      
      <div v-if="selectedUserId && userCategoriesLoading" class="loading">Loading user categories...</div>
      <div v-else-if="selectedUserId && userCategoriesError" class="error">{{ userCategoriesError }}</div>
      <div v-else-if="selectedUserId" class="user-categories-section">
        <h3>Categories for selected user</h3>
        <div class="categories-checkboxes">
          <label v-for="cat in categories" :key="cat.id" class="checkbox-label">
            <input 
              type="checkbox" 
              :value="cat.id" 
              v-model="selectedUserCategories"
            />
            <span>{{ cat.name }}</span>
            <span v-if="cat.permissions && cat.permissions.length > 0" class="permissions-hint">
              ({{ cat.permissions.length }} permission{{ cat.permissions.length !== 1 ? 's' : '' }})
            </span>
          </label>
        </div>
        <div class="form-actions">
          <button @click="saveUserCategories" class="btn btn-primary" :disabled="savingUserCategories">
            {{ savingUserCategories ? 'Saving...' : 'Save' }}
          </button>
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
          <div class="modal-actions">
            <button type="submit" class="btn btn-primary">Save</button>
            <button type="button" @click="closeCategoryModal" class="btn btn-secondary">Cancel</button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- Permissions Modal -->
    <div v-if="showPermissionsModal && categoryForPermissions" class="modal" @click.self="closePermissionsModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Edit Permissions: {{ categoryForPermissions.name }}</h3>
          <button @click="closePermissionsModal" class="btn-close">&times;</button>
        </div>
        <div class="modal-body">
          <div class="permissions-grid">
            <label v-for="perm in availablePermissions" :key="perm" class="checkbox-label">
              <input 
                type="checkbox" 
                :value="perm" 
                v-model="selectedPermissions"
              />
              <span>{{ perm }}</span>
            </label>
          </div>
          <div class="modal-actions">
            <button @click="savePermissions" class="btn btn-primary" :disabled="savingPermissions">
              {{ savingPermissions ? 'Saving...' : 'Save' }}
            </button>
            <button type="button" @click="closePermissionsModal" class="btn btn-secondary">Cancel</button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Delete Confirm Modal -->
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import { useAuth } from '../composables/useAuth'

const { loadPermissions } = useAuth()

interface Category {
  id: number
  name: string
  description?: string | null
  permissions?: string[]
}

interface User {
  id: number
  telegram_id: number
  telegram_username?: string
}

const activeTab = ref<'categories' | 'users'>('categories')

// Categories
const categories = ref<Category[]>([])
const categoriesLoading = ref(false)
const categoriesError = ref<string | null>(null)
const showCategoryModal = ref(false)
const editingCategory = ref<Category | null>(null)
const categoryForm = ref({
  name: '',
  description: ''
})
const showDeleteCategoryConfirm = ref(false)
const categoryToDelete = ref<Category | null>(null)

// Permissions
const showPermissionsModal = ref(false)
const categoryForPermissions = ref<Category | null>(null)
const availablePermissions = ref<string[]>([])
const selectedPermissions = ref<string[]>([])
const savingPermissions = ref(false)

// Users
const users = ref<User[]>([])
const usersLoading = ref(false)
const selectedUserId = ref<number | null>(null)
const userCategories = ref<number[]>([])
const selectedUserCategories = ref<number[]>([])
const userCategoriesLoading = ref(false)
const userCategoriesError = ref<string | null>(null)
const savingUserCategories = ref(false)

const switchTab = (tab: 'categories' | 'users') => {
  activeTab.value = tab
  if (tab === 'users' && users.value.length === 0) {
    loadUsers()
  }
}

onMounted(async () => {
  await loadPermissions()
  await loadAvailablePermissions()
  await loadCategories()
  await loadUsers()
})

const loadAvailablePermissions = async () => {
  try {
    const data: { permissions: string[] } = await apiClient.request('/api/admin/access/available-permissions')
    availablePermissions.value = data.permissions || []
  } catch (error) {
    console.error('Failed to load available permissions:', error)
  }
}

const loadCategories = async () => {
  categoriesLoading.value = true
  categoriesError.value = null
  try {
    const data: { categories: Category[] } = await apiClient.request('/api/admin/access/categories')
    categories.value = (data.categories || []).map(cat => ({
      id: cat.id,
      name: cat.name,
      description: cat.description,
      permissions: cat.permissions || []
    }))
  } catch (error: any) {
    console.error('Failed to load categories:', error)
    categoriesError.value = error.message || 'Failed to load categories'
  } finally {
    categoriesLoading.value = false
  }
}

const loadUsers = async () => {
  usersLoading.value = true
  try {
    const data: { users: User[] } = await apiClient.request('/api/admin/users')
    users.value = data.users || []
  } catch (error) {
    console.error('Failed to load users:', error)
  } finally {
    usersLoading.value = false
  }
}

const loadUserCategories = async () => {
  if (!selectedUserId.value) {
    return
  }
  
  userCategoriesLoading.value = true
  userCategoriesError.value = null
  try {
    const data: { user_id: number; categories: number[] } = await apiClient.request(`/api/admin/access/users/${selectedUserId.value}`)
    selectedUserCategories.value = data.categories || []
  } catch (error: any) {
    console.error('Failed to load user categories:', error)
    userCategoriesError.value = error.message || 'Failed to load user categories'
  } finally {
    userCategoriesLoading.value = false
  }
}

// Category CRUD
const startCreateCategory = () => {
  editingCategory.value = null
  categoryForm.value = {
    name: '',
    description: ''
  }
  showCategoryModal.value = true
}

const startEditCategory = (category: Category) => {
  editingCategory.value = category
  categoryForm.value = {
    name: category.name,
    description: category.description || ''
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
      ? `/api/admin/access/categories/${editingCategory.value.id}`
      : '/api/admin/access/categories'
    
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
    await apiClient.request(`/api/admin/access/categories/${categoryToDelete.value.id}`, {
      method: 'DELETE'
    })
    
    closeDeleteCategoryConfirm()
    await loadCategories()
    await showAlert('Category deleted successfully')
  } catch (error: any) {
    console.error('Failed to delete category:', error)
    await showAlert(error.message || 'Failed to delete category')
  }
}

// Permissions
const startEditPermissions = (category: Category) => {
  categoryForPermissions.value = category
  selectedPermissions.value = [...(category.permissions || [])]
  showPermissionsModal.value = true
}

const closePermissionsModal = () => {
  showPermissionsModal.value = false
  categoryForPermissions.value = null
  selectedPermissions.value = []
}

const savePermissions = async () => {
  if (!categoryForPermissions.value) return
  
  savingPermissions.value = true
  try {
    await apiClient.request(`/api/admin/access/categories/${categoryForPermissions.value.id}/permissions`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ permissions: selectedPermissions.value })
    })
    
    closePermissionsModal()
    await loadCategories()
    await showAlert('Permissions saved successfully')
  } catch (error: any) {
    console.error('Failed to save permissions:', error)
    await showAlert(error.message || 'Failed to save permissions')
  } finally {
    savingPermissions.value = false
  }
}

// User categories
const saveUserCategories = async () => {
  if (!selectedUserId.value) return
  
  savingUserCategories.value = true
  try {
    await apiClient.request(`/api/admin/access/users/${selectedUserId.value}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ category_ids: selectedUserCategories.value })
    })
    
    await loadUserCategories()
    await showAlert('User categories saved successfully')
  } catch (error: any) {
    console.error('Failed to save user categories:', error)
    await showAlert(error.message || 'Failed to save user categories')
  } finally {
    savingUserCategories.value = false
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    if (showCategoryModal.value) {
      event.preventDefault()
      closeCategoryModal()
    } else if (showPermissionsModal.value) {
      event.preventDefault()
      closePermissionsModal()
    } else if (showDeleteCategoryConfirm.value) {
      event.preventDefault()
      closeDeleteCategoryConfirm()
    }
  }
}

watch([showCategoryModal, showPermissionsModal, showDeleteCategoryConfirm], 
  ([catOpen, permsOpen, delOpen]) => {
    if (catOpen || permsOpen || delOpen) {
      window.addEventListener('keydown', handleKeydown)
    } else {
      window.removeEventListener('keydown', handleKeydown)
    }
  }
)

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.admin-content {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  font-size: 16px;
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

.categories-list {
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
  gap: 8px;
}

.category-desc {
  color: var(--text-secondary);
  font-size: 14px;
}

.permissions-list {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.permissions-label {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 14px;
}

.permissions-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.permission-tag {
  padding: 2px 8px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  font-size: 12px;
  color: var(--text-primary);
}

.no-permissions {
  color: var(--text-secondary);
  font-size: 12px;
  font-style: italic;
}

.category-actions {
  display: flex;
  gap: 8px;
}

.user-select-section {
  margin-bottom: 24px;
}

.user-select-section label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--text-primary);
}

.user-categories-section {
  margin-top: 24px;
}

.user-categories-section h3 {
  margin-bottom: 16px;
}

.categories-checkboxes {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.permissions-hint {
  color: var(--text-secondary);
  font-size: 12px;
}

.form-select,
.form-input,
.form-textarea {
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
}

.form-select {
  min-width: 300px;
}

.form-textarea {
  width: 100%;
  resize: vertical;
}

.form-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
  margin-bottom: 20px;
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

.btn-primary:hover:not(:disabled) {
  background-color: var(--color-primary-hover, #2563eb);
}

.btn-secondary {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: var(--color-secondary-hover, #4b5563);
}

.btn-danger {
  background-color: var(--color-danger, #ef4444);
  color: white;
}

.btn-danger:hover:not(:disabled) {
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
  .admin-content {
    margin-top: 0 !important;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .category-row {
    flex-wrap: wrap;
    gap: 8px;
  }
  
  .category-actions {
    width: 100%;
    justify-content: flex-start;
  }
  
  .permissions-grid {
    grid-template-columns: 1fr;
  }
  
  .form-select {
    width: 100%;
  }
}
</style>
