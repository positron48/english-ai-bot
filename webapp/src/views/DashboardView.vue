<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else class="dashboard-stats">
      <div class="card">
        <h2>Cards Ready for Review</h2>
        <p class="stat-number">{{ dueCount }}</p>
        <router-link to="/training" class="btn btn-primary">Start Training</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

const dueCount = ref(0)
const loading = ref(true)

onMounted(async () => {
  try {
    const data: { due_count: number } = await apiClient.request('/app/dashboard')
    dueCount.value = data.due_count
  } catch (error) {
    console.error('Failed to load dashboard:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dashboard-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.stat-number {
  font-size: 48px;
  font-weight: bold;
  color: #007bff;
  margin: 20px 0;
}

h2 {
  margin-bottom: 10px;
}
</style>

