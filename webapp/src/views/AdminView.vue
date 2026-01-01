<template>
  <div class="admin">
    <h1>Admin Panel</h1>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else>
      <div class="card">
        <div class="circuit-breaker-header">
          <h2>Circuit Breaker</h2>
          <button v-if="circuitBreaker" @click="resetCircuitBreaker" class="btn btn-primary">Reset</button>
        </div>
        <div v-if="circuitBreaker" class="circuit-breaker-content">
          <div class="circuit-breaker-info">
            <div class="info-row">
              <span class="info-label">State:</span>
              <span class="info-value" :class="{ 'state-open': circuitBreaker.state === 'open', 'state-closed': circuitBreaker.state === 'closed' }">
                {{ circuitBreaker.state || 'closed' }}
              </span>
            </div>
            <div v-if="circuitBreaker.failures !== undefined" class="info-row">
              <span class="info-label">Failures:</span>
              <span class="info-value">{{ circuitBreaker.failures }}</span>
            </div>
            <div v-if="circuitBreaker.last_failure_at" class="info-row">
              <span class="info-label">Last failure at:</span>
              <span class="info-value">{{ formatDate(circuitBreaker.last_failure_at) }}</span>
            </div>
            <div v-if="circuitBreaker.last_failure" class="info-row">
              <span class="info-label">Last failure:</span>
              <span class="info-value">{{ circuitBreaker.last_failure }}</span>
            </div>
            <div v-if="circuitBreaker.last_reset_at" class="info-row">
              <span class="info-label">Last reset at:</span>
              <span class="info-value">{{ formatDate(circuitBreaker.last_reset_at) }}</span>
            </div>
          </div>
        </div>
        <p v-else>No circuit breaker data</p>
      </div>

      <div class="card">
        <h2>Training Cards Management</h2>
        <div class="admin-actions">
          <input
            v-model="wordToManage"
            type="text"
            placeholder="Enter word"
            class="admin-input"
          />
          <button @click="getTrainingData" class="btn btn-primary">Get Training Data</button>
          <button @click="deleteTrainingWord" class="btn btn-danger">Delete Word</button>
          <button @click="deleteAllTraining" class="btn btn-danger">Delete All</button>
        </div>
        
        <div v-if="trainingData" class="training-data">
          <h3>Training Data for "{{ wordToManage }}"</h3>
          <pre>{{ JSON.stringify(trainingData, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface CircuitBreaker {
  state: string
  failures?: number
  last_failure?: string
}

const loading = ref(true)
const circuitBreaker = ref<CircuitBreaker | null>(null)
const wordToManage = ref('')
const trainingData = ref<any>(null)

onMounted(async () => {
  await loadAdminData()
})

const loadAdminData = async () => {
  loading.value = true
  try {
    const data: { circuit_breaker: CircuitBreaker } = await apiClient.request('/app/admin')
    circuitBreaker.value = data.circuit_breaker
  } catch (error) {
    console.error('Failed to load admin data:', error)
  } finally {
    loading.value = false
  }
}

const resetCircuitBreaker = async () => {
  try {
    await apiClient.request('/app/admin/circuit/reset', { method: 'POST' })
    await loadAdminData()
    alert('Circuit breaker reset successfully')
  } catch (error) {
    console.error('Failed to reset circuit breaker:', error)
    alert('Failed to reset circuit breaker')
  }
}

const getTrainingData = async () => {
  if (!wordToManage.value.trim()) {
    alert('Please enter a word')
    return
  }

  try {
    const data = await apiClient.request(`/app/admin/training/${wordToManage.value.trim()}`)
    trainingData.value = data
  } catch (error) {
    console.error('Failed to get training data:', error)
    alert('Failed to get training data')
  }
}

const deleteTrainingWord = async () => {
  if (!wordToManage.value.trim()) {
    alert('Please enter a word')
    return
  }

  if (!confirm(`Are you sure you want to delete all training cards for "${wordToManage.value}"?`)) {
    return
  }

  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/app/admin/training/${wordToManage.value.trim()}/delete`, formData)
    trainingData.value = null
    alert('Training cards deleted successfully')
  } catch (error) {
    console.error('Failed to delete training word:', error)
    alert('Failed to delete training word')
  }
}

const deleteAllTraining = async () => {
  if (!confirm('Are you sure you want to delete ALL training cards? This cannot be undone!')) {
    return
  }

  try {
    const formData = new FormData()
    await apiClient.requestFormData('/app/admin/training/delete_all', formData)
    trainingData.value = null
    alert('All training cards deleted successfully')
  } catch (error) {
    console.error('Failed to delete all training:', error)
    alert('Failed to delete all training')
  }
}

const formatDate = (dateStr: string | null | undefined) => {
  if (!dateStr) return '—'
  
  // Handle SQL datetime format "2006-01-02 15:04:05" (same as dashboard sessions)
  let date: Date
  if (dateStr.includes(' ')) {
    // SQL format: replace space with T for ISO format, assume local timezone
    date = new Date(dateStr.replace(' ', 'T'))
  } else {
    date = new Date(dateStr)
  }
  
  // Check if date is valid
  if (isNaN(date.getTime())) {
    return '—'
  }
  
  // Format same way as VocabView does it
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.admin {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.admin h1 {
  margin-bottom: 24px;
}

.admin .card h2 {
  margin-bottom: 20px;
}

.circuit-breaker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.circuit-breaker-header h2 {
  margin: 0;
}

.circuit-breaker-content {
  margin-top: 0;
}

.circuit-breaker-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  gap: 10px;
  align-items: baseline;
}

.info-label {
  font-weight: 500;
  color: var(--text-secondary);
  min-width: 140px;
}

.info-value {
  color: var(--text-primary);
  word-break: break-word;
}

.info-value.state-open {
  color: var(--color-danger);
  font-weight: 600;
  background: rgba(220, 53, 69, 0.1);
  padding: 4px 12px;
  border-radius: 4px;
  display: inline-block;
  text-transform: uppercase;
  font-size: 0.9em;
  letter-spacing: 0.5px;
}

.info-value.state-closed {
  color: var(--color-success);
  font-weight: 600;
  background: rgba(40, 167, 69, 0.1);
  padding: 4px 12px;
  border-radius: 4px;
  display: inline-block;
  text-transform: uppercase;
  font-size: 0.9em;
  letter-spacing: 0.5px;
}

.admin-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: stretch;
  margin-bottom: 20px;
}

.admin-input {
  flex: 1;
  min-width: 200px;
  height: 40px;
  padding: 10px;
  box-sizing: border-box;
}

.admin-actions .btn {
  height: 40px;
  padding: 10px 20px;
  box-sizing: border-box;
  white-space: nowrap;
}

.training-data {
  margin-top: 20px;
  padding: 15px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  color: var(--text-primary);
}

.training-data pre {
  overflow-x: auto;
  font-size: 12px;
}
</style>

