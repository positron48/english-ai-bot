<template>
  <div class="admin-stats-view">
    <div class="admin-stats-header">
      <h1>Statistics</h1>
      <div class="header-controls">
        <select v-model="selectedDays" @change="loadData" class="days-select">
          <option :value="7">Last 7 days</option>
          <option :value="30">Last 30 days</option>
        </select>
        <button @click="refreshData" class="btn-refresh" :disabled="loading" :class="{ 'rotating': loading }">
          <Icon name="refresh" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading statistics...</div>

    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
    </div>

    <div v-else class="admin-stats-content">
      <!-- Users Stats Cards -->
      <div class="stats-section">
        <h2>Users</h2>
        <div class="stats-grid">
          <div class="stat-card stat-card-primary">
            <div class="stat-header">
              <Icon name="users" class="stat-icon" />
              <h3>Total Users</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ stats.users_total || 0 }}</p>
            </div>
          </div>
          <div class="stat-card stat-card-info">
            <div class="stat-header">
              <Icon name="plus" class="stat-icon" />
              <h3>New Users</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.users_new || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.users_new || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.users_new || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
          <div class="stat-card stat-card-success">
            <div class="stat-header">
              <Icon name="target" class="stat-icon" />
              <h3>Active Users</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.users_active || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.users_active || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.users_active || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Activity Stats Cards -->
      <div class="stats-section">
        <h2>Activity & Training</h2>
        <div class="stats-grid">
          <div class="stat-card stat-card-warning">
            <div class="stat-header">
              <Icon name="play" class="stat-icon" />
              <h3>Sessions Started</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.sessions_started || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.sessions_started || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.sessions_started || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
          <div class="stat-card stat-card-success">
            <div class="stat-header">
              <Icon name="shield" class="stat-icon" />
              <h3>Sessions Completed</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.sessions_completed || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.sessions_completed || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.sessions_completed || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
          <div class="stat-card stat-card-info">
            <div class="stat-header">
              <Icon name="chat" class="stat-icon" />
              <h3>Reviews Answered</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.reviews_answered || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.reviews_answered || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.reviews_answered || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
          <div class="stat-card stat-card-primary">
            <div class="stat-header">
              <Icon name="target" class="stat-icon" />
              <h3>Accuracy</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ formatPercent(windows['24h']?.accuracy_percent || 0) }}%</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ formatPercent(windows['7d']?.accuracy_percent || 0) }}%</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ formatPercent(windows['30d']?.accuracy_percent || 0) }}%</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
          <div class="stat-card stat-card-secondary">
            <div class="stat-header">
              <Icon name="plus" class="stat-icon" />
              <h3>Cards Added</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['24h']?.cards_added || 0 }}</p>
              <p class="stat-label">24h</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['7d']?.cards_added || 0 }}</p>
              <p class="stat-label">7d</p>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ windows['30d']?.cards_added || 0 }}</p>
              <p class="stat-label">30d</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Cards State Stats -->
      <div class="stats-section">
        <h2>Cards State (Global)</h2>
        <div class="stats-grid">
          <div class="stat-card stat-card-primary">
            <div class="stat-header">
              <Icon name="dashboard" class="stat-icon" />
              <h3>Total Cards</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ cardsState.total || 0 }}</p>
            </div>
          </div>
          <div class="stat-card stat-card-info">
            <div class="stat-header">
              <Icon name="sparkles" class="stat-icon" />
              <h3>New</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ cardsState.new || 0 }}</p>
            </div>
          </div>
          <div class="stat-card stat-card-warning">
            <div class="stat-header">
              <Icon name="book-open" class="stat-icon" />
              <h3>Learning</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ cardsState.learning || 0 }}</p>
            </div>
          </div>
          <div class="stat-card stat-card-success">
            <div class="stat-header">
              <Icon name="refresh" class="stat-icon" />
              <h3>Review</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ cardsState.review || 0 }}</p>
            </div>
          </div>
          <div class="stat-card stat-card-secondary">
            <div class="stat-header">
              <Icon name="refresh" class="stat-icon" />
              <h3>Due Now</h3>
            </div>
            <div class="stat-value-row">
              <p class="stat-number">{{ cardsState.due_now || 0 }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Charts Section -->
      <div class="charts-section">
        <div class="chart-card">
          <h2>Active Users ({{ selectedDays }} days)</h2>
          <div v-if="dailyStats && dailyStats.length > 0" class="chart-container">
            <canvas ref="activeUsersChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No data available</p>
          </div>
        </div>

        <div class="chart-card">
          <h2>Sessions Started ({{ selectedDays }} days)</h2>
          <div v-if="dailyStats && dailyStats.length > 0" class="chart-container">
            <canvas ref="sessionsChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No data available</p>
          </div>
        </div>

        <div class="chart-card">
          <h2>Reviews Answered ({{ selectedDays }} days)</h2>
          <div v-if="dailyStats && dailyStats.length > 0" class="chart-container">
            <canvas ref="reviewsChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No data available</p>
          </div>
        </div>

        <div class="chart-card">
          <h2>Cards Added ({{ selectedDays }} days)</h2>
          <div v-if="dailyStats && dailyStats.length > 0" class="chart-container">
            <canvas ref="cardsAddedChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No data available</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { Chart, registerables } from 'chart.js'
import { useTheme } from '../composables/useTheme'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

Chart.register(...registerables)

const { theme } = useTheme()
const selectedDays = ref(30)
const loading = ref(true)
const error = ref<string | null>(null)

const stats = ref<any>({})
const windows = ref<any>({})
const cardsState = ref<any>({})
const dailyStats = ref<any[]>([])

const activeUsersChartCanvas = ref<HTMLCanvasElement | null>(null)
const sessionsChartCanvas = ref<HTMLCanvasElement | null>(null)
const reviewsChartCanvas = ref<HTMLCanvasElement | null>(null)
const cardsAddedChartCanvas = ref<HTMLCanvasElement | null>(null)

let activeUsersChart: Chart | null = null
let sessionsChart: Chart | null = null
let reviewsChart: Chart | null = null
let cardsAddedChart: Chart | null = null

const formatPercent = (value: number): string => {
  return value.toFixed(1)
}

const formatDayLabel = (dayString: string): string => {
  const date = new Date(dayString + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  date.setHours(0, 0, 0, 0)
  const diffTime = today.getTime() - date.getTime()
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === 1) {
    return 'Yesterday'
  } else {
    return date.toLocaleDateString('en-US', { weekday: 'short', day: 'numeric' })
  }
}

const getChartColors = () => {
  const root = getComputedStyle(document.documentElement)
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  
  let primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
  let successColor = root.getPropertyValue('--color-success').trim() || '#28a745'
  const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
  let textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
  let borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'
  
  if (!isDark) {
    primaryColor = '#0056b3'
    successColor = '#1e7e34'
    textSecondary = '#444444'
    borderColor = '#cccccc'
  }
  
  const hexToRgba = (hex: string, alpha: number) => {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }
  
  return { primaryColor, successColor, textPrimary, textSecondary, borderColor, isDark, hexToRgba }
}

const updateCharts = () => {
  if (!dailyStats.value || dailyStats.value.length === 0) {
    return
  }
  
  const labels = dailyStats.value.map((d: any) => formatDayLabel(d.day))
  
  const colors = getChartColors()
  
  // Active Users Chart
  if (activeUsersChartCanvas.value) {
    if (activeUsersChart) {
      activeUsersChart.destroy()
    }
    activeUsersChart = new Chart(activeUsersChartCanvas.value, {
      type: 'line',
      data: {
        labels: labels,
        datasets: [{
          label: 'Active Users',
          data: dailyStats.value.map((d: any) => d.active_users || 0),
          borderColor: colors.primaryColor,
          backgroundColor: colors.hexToRgba(colors.primaryColor, 0.1),
          tension: 0.4,
          fill: true
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true, position: 'top' }
        },
        scales: {
          x: { ticks: { color: colors.textSecondary } },
          y: { ticks: { color: colors.textSecondary }, beginAtZero: true }
        }
      }
    })
  }
  
  // Sessions Chart
  if (sessionsChartCanvas.value) {
    if (sessionsChart) {
      sessionsChart.destroy()
    }
    sessionsChart = new Chart(sessionsChartCanvas.value, {
      type: 'bar',
      data: {
        labels: labels,
        datasets: [{
          label: 'Sessions Started',
          data: dailyStats.value.map((d: any) => d.sessions_started || 0),
          backgroundColor: colors.hexToRgba(colors.primaryColor, colors.isDark ? 0.7 : 0.6)
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true, position: 'top' }
        },
        scales: {
          x: { ticks: { color: colors.textSecondary } },
          y: { ticks: { color: colors.textSecondary }, beginAtZero: true }
        }
      }
    })
  }
  
  // Reviews Chart
  if (reviewsChartCanvas.value) {
    if (reviewsChart) {
      reviewsChart.destroy()
    }
    reviewsChart = new Chart(reviewsChartCanvas.value, {
      type: 'bar',
      data: {
        labels: labels,
        datasets: [{
          label: 'Reviews Answered',
          data: dailyStats.value.map((d: any) => d.reviews_answered || 0),
          backgroundColor: colors.hexToRgba(colors.successColor, colors.isDark ? 0.7 : 0.6)
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true, position: 'top' }
        },
        scales: {
          x: { ticks: { color: colors.textSecondary } },
          y: { ticks: { color: colors.textSecondary }, beginAtZero: true }
        }
      }
    })
  }
  
  // Cards Added Chart
  if (cardsAddedChartCanvas.value) {
    if (cardsAddedChart) {
      cardsAddedChart.destroy()
    }
    cardsAddedChart = new Chart(cardsAddedChartCanvas.value, {
      type: 'line',
      data: {
        labels: labels,
        datasets: [{
          label: 'Cards Added',
          data: dailyStats.value.map((d: any) => d.cards_added || 0),
          borderColor: colors.successColor,
          backgroundColor: colors.hexToRgba(colors.successColor, 0.1),
          tension: 0.4,
          fill: true
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: true, position: 'top' }
        },
        scales: {
          x: { ticks: { color: colors.textSecondary } },
          y: { ticks: { color: colors.textSecondary }, beginAtZero: true }
        }
      }
    })
  }
}

watch(() => dailyStats.value, async () => {
  if (dailyStats.value && dailyStats.value.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateCharts()
    }, 150)
  }
}, { deep: true })

watch(() => theme.value, async () => {
  if (dailyStats.value && dailyStats.value.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateCharts()
    }, 100)
  }
})

const loadData = async () => {
  loading.value = true
  error.value = null
  
  try {
    const data = await apiClient.request(`/api/admin/stats?days=${selectedDays.value}`)
    stats.value = { users_total: data.users_total || 0 }
    windows.value = data.windows || {}
    cardsState.value = data.cards_state || {}
    dailyStats.value = data.daily || []
    
    await nextTick()
    updateCharts()
  } catch (err: any) {
    console.error('Failed to load stats:', err)
    error.value = err.message || 'Failed to load statistics'
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.admin-stats-view {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.admin-stats-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.admin-stats-header h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-controls {
  display: flex;
  gap: 12px;
  align-items: center;
}

.days-select {
  padding: 8px 12px;
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 14px;
  cursor: pointer;
}

.btn-refresh {
  padding: 8px;
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s;
}

.btn-refresh:hover {
  background: var(--bg-hover);
}

.btn-refresh.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-error, #d32f2f);
}

.stats-section {
  margin-bottom: 32px;
}

.stats-section h2 {
  margin: 0 0 16px 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stat-card {
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 16px;
}

.stat-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.stat-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.stat-icon {
  width: 20px;
  height: 20px;
  color: var(--text-secondary);
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
}

.stat-number {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
}

.stat-label {
  margin: 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.charts-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 24px;
  margin-top: 32px;
}

.chart-card {
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
}

.chart-card h2 {
  margin: 0 0 16px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.chart-container {
  position: relative;
  height: 300px;
}

.chart-empty {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .admin-stats-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .charts-section {
    grid-template-columns: 1fr;
  }
}
</style>
