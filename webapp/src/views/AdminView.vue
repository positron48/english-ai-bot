<template>
  <div class="admin">
    <h1>Admin Panel</h1>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else>
      <div class="card">
        <h2>Circuit Breaker</h2>
        <div v-if="circuitBreaker">
          <p>State: {{ circuitBreaker.state }}</p>
          <p v-if="circuitBreaker.failures">Failures: {{ circuitBreaker.failures }}</p>
          <p v-if="circuitBreaker.last_failure">Last failure: {{ formatDate(circuitBreaker.last_failure) }}</p>
          <button @click="resetCircuitBreaker" class="btn btn-primary">Reset Circuit Breaker</button>
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

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}
</script>

<style scoped>
.admin-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}

.admin-input {
  flex: 1;
  min-width: 200px;
}

.training-data {
  margin-top: 20px;
  padding: 15px;
  background: #f5f5f5;
  border-radius: 4px;
}

.training-data pre {
  overflow-x: auto;
  font-size: 12px;
}
</style>

