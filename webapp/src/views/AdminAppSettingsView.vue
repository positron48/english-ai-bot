<template>
  <div class="admin-content">
    <div class="card">
      <h2>App Settings</h2>
      
      <div v-if="loading" class="loading">
        <p>Loading settings...</p>
      </div>
      
      <div v-else-if="error" class="error">
        <p>{{ error }}</p>
        <button @click="loadSettings" class="btn btn-primary">Retry</button>
      </div>
      
      <div v-else class="settings-form">
        <div class="form-group">
          <label class="form-label">
            <input
              type="checkbox"
              v-model="hidePlacementTestButton"
              @change="saveSettings"
              class="form-checkbox"
            />
            <span class="form-label-text">Hide Placement Test Button</span>
          </label>
          <p class="form-hint">
            When enabled, the "Take Placement Test" button will be hidden in the public grammar categories view.
          </p>
        </div>
        
        <div v-if="saving" class="saving-indicator">
          <span>Saving...</span>
        </div>
        <div v-if="saveSuccess" class="success-message">
          <span>Settings saved successfully!</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

const loading = ref(true)
const error = ref<string | null>(null)
const hidePlacementTestButton = ref(false)
const saving = ref(false)
const saveSuccess = ref(false)

const loadSettings = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { hide_placement_test_button: boolean } = await apiClient.request('/api/admin/app-settings')
    hidePlacementTestButton.value = data.hide_placement_test_button || false
  } catch (err: any) {
    error.value = err.message || 'Failed to load settings'
    console.error('Failed to load app settings:', err)
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  saveSuccess.value = false
  try {
    await apiClient.request('/api/admin/app-settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        hide_placement_test_button: hidePlacementTestButton.value
      })
    })
    saveSuccess.value = true
    setTimeout(() => {
      saveSuccess.value = false
    }, 3000)
  } catch (err: any) {
    error.value = err.message || 'Failed to save settings'
    console.error('Failed to save app settings:', err)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})
</script>

<style scoped>
.admin-content {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.card {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  padding: 24px;
}

.card h2 {
  margin: 0 0 24px 0;
  font-size: 24px;
  color: var(--text-primary);
}

.loading, .error {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger);
}

.settings-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  user-select: none;
}

.form-checkbox {
  width: 20px;
  height: 20px;
  cursor: pointer;
  accent-color: var(--color-primary);
}

.form-label-text {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary);
}

.form-hint {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
  margin-left: 32px;
}

.saving-indicator {
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.success-message {
  padding: 12px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.3);
  border-radius: 8px;
  color: #10b981;
  font-size: 14px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
  opacity: 0.9;
}
</style>
