<template>
  <div class="admin-content">
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'
import { showAlert } from '../composables/useDialog'

interface CircuitBreaker {
  state: string
  failures?: number
  last_failure?: string
  last_failure_at?: string
  last_reset_at?: string
}

const loading = ref(true)
const circuitBreaker = ref<CircuitBreaker | null>(null)

onMounted(async () => {
  await loadAdminData()
})

const loadAdminData = async () => {
  loading.value = true
  try {
    const data: { circuit_breaker: CircuitBreaker } = await apiClient.request('/api/admin')
    circuitBreaker.value = data.circuit_breaker
  } catch (error) {
    console.error('Failed to load admin data:', error)
  } finally {
    loading.value = false
  }
}

const resetCircuitBreaker = async () => {
  try {
    await apiClient.request('/api/admin/circuit/reset', { method: 'POST' })
    await loadAdminData()
    await showAlert('Circuit breaker reset successfully')
  } catch (error) {
    console.error('Failed to reset circuit breaker:', error)
    await showAlert('Failed to reset circuit breaker')
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
.admin-content {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  font-size: 16px;
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

.card {
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  opacity: 0.9;
}

.loading {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .admin-content {
    margin-top: 0 !important;
  }

  .admin-tab {
    padding: 10px 16px;
  }
}

@media (max-width: 480px) {
  .admin-tab {
    padding: 8px 12px;
  }
}
</style>
