<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Dashboard</h1>
      <button @click="refreshData" class="btn-refresh" :disabled="loading">
        <span v-if="!loading">↻</span>
        <span v-else>⟳</span>
      </button>
    </div>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else class="dashboard-content">
      <!-- Main Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card stat-card-primary stat-card-clickable" @click="goToTraining">
          <div class="stat-header">
            <div class="stat-icon">📚</div>
            <h3>Available for Training</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.availableForTraining }}</p>
            <p class="stat-label">Cards available</p>
          </div>
        </div>

        <div class="stat-card stat-card-info">
          <div class="stat-header">
            <div class="stat-icon">✨</div>
            <h3>New Cards</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.newCount }}</p>
            <p class="stat-label">Not started</p>
          </div>
        </div>

        <div class="stat-card stat-card-success">
          <div class="stat-header">
            <div class="stat-icon">📖</div>
            <h3>In Learning</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.learningCount }}</p>
            <p class="stat-label">Being learned</p>
          </div>
        </div>

        <div class="stat-card stat-card-warning">
          <div class="stat-header">
            <div class="stat-icon">🔄</div>
            <h3>In Review</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.reviewCount }}</p>
            <p class="stat-label">Mastered cards</p>
          </div>
        </div>
      </div>

      <!-- Progress Section -->
      <div class="progress-section">
        <div class="card">
          <h2>Your Progress</h2>
          <div class="progress-grid">
            <div class="progress-item">
              <div class="progress-header">
                <span>Total Cards</span>
                <span class="progress-value">{{ stats.totalCards }}</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: '100%' }"></div>
              </div>
            </div>
            
            <div class="progress-item">
              <div class="progress-header">
                <span>Ready for Review</span>
                <span class="progress-value">{{ stats.dueCount }}</span>
              </div>
              <div class="progress-bar">
                <div 
                  class="progress-fill progress-fill-primary" 
                  :style="{ width: stats.totalCards > 0 ? (stats.dueCount / stats.totalCards * 100) + '%' : '0%' }"
                ></div>
              </div>
            </div>

            <div class="progress-item">
              <div class="progress-header">
                <span>Accuracy (30 days)</span>
                <span class="progress-value">{{ formatPercent(stats.accuracyPercent) }}%</span>
              </div>
              <div class="progress-bar">
                <div 
                  class="progress-fill progress-fill-success" 
                  :style="{ width: stats.accuracyPercent + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Activity Section -->
      <div class="activity-section">
        <div class="activity-grid">
          <div class="card activity-card">
            <h2>Today</h2>
            <div class="activity-stats">
              <div class="activity-stat">
                <span class="activity-label">Sessions</span>
                <span class="activity-value">{{ stats.todaySessions }}</span>
              </div>
              <div class="activity-stat">
                <span class="activity-label">Cards Completed</span>
                <span class="activity-value">{{ stats.todayCardsCompleted }}</span>
              </div>
            </div>
          </div>

          <div class="card activity-card">
            <h2>This Week</h2>
            <div class="activity-stats">
              <div class="activity-stat">
                <span class="activity-label">Sessions</span>
                <span class="activity-value">{{ stats.weekSessions }}</span>
              </div>
              <div class="activity-stat">
                <span class="activity-label">Cards Completed</span>
                <span class="activity-value">{{ stats.weekCardsCompleted }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Weekly Chart -->
      <div class="weekly-chart-section">
        <div class="card">
          <h2>Weekly Activity</h2>
          <div v-if="stats.weeklyStats && stats.weeklyStats.length > 0" class="chart-container">
            <canvas ref="chartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No training data available for the last week</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Chart, registerables } from 'chart.js'
import { useTheme } from '../composables/useTheme'
import { apiClient } from '../api/client'

Chart.register(...registerables)

const router = useRouter()
const { theme } = useTheme()
const chartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

interface WeeklyStat {
  day: string
  cards_completed: number
  cards_correct: number
}

interface DashboardStats {
  dueCount: number
  newCount: number
  learningCount: number
  reviewCount: number
  totalCards: number
  availableForTraining: number
  todaySessions: number
  todayCardsCompleted: number
  weekSessions: number
  weekCardsCompleted: number
  accuracyPercent: number
  weeklyStats: WeeklyStat[]
}

const stats = ref<DashboardStats>({
  dueCount: 0,
  newCount: 0,
  learningCount: 0,
  reviewCount: 0,
  totalCards: 0,
  availableForTraining: 0,
  todaySessions: 0,
  todayCardsCompleted: 0,
  weekSessions: 0,
  weekCardsCompleted: 0,
  accuracyPercent: 0,
  weeklyStats: []
})

const loading = ref(true)

// Watch for changes in weeklyStats and update chart (after stats is initialized)
watch(() => stats.value.weeklyStats, async (newStats) => {
  if (newStats && newStats.length > 0) {
    await nextTick()
    // Wait a bit more to ensure canvas is fully rendered
    setTimeout(() => {
      updateChart()
    }, 150)
  }
}, { deep: true })

// Watch for theme changes and rebuild chart
watch(() => theme.value, async () => {
  if (stats.value.weeklyStats && stats.value.weeklyStats.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateChart()
    }, 100)
  }
})

const loadData = async () => {
  try {
    loading.value = true
    const data = await apiClient.request('/app/dashboard')
    stats.value = {
      dueCount: data.due_count || 0,
      newCount: data.new_count || 0,
      learningCount: data.learning_count || 0,
      reviewCount: data.review_count || 0,
      totalCards: data.total_cards || 0,
      availableForTraining: data.available_for_training || 0,
      todaySessions: data.today_sessions || 0,
      todayCardsCompleted: data.today_cards_completed || 0,
      weekSessions: data.week_sessions || 0,
      weekCardsCompleted: data.week_cards_completed || 0,
      accuracyPercent: data.accuracy_percent || 0,
      weeklyStats: data.weekly_stats || []
    }
    await nextTick()
    updateChart()
  } catch (error) {
    console.error('Failed to load dashboard:', error)
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadData()
}

const formatPercent = (value: number): string => {
  return value.toFixed(1)
}

const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diffTime = Math.abs(now.getTime() - date.getTime())
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === 1) {
    return 'Yesterday'
  } else if (diffDays < 7) {
    return `${diffDays} days ago`
  } else {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }
}

const goToTraining = () => {
  router.push('/training')
}

const formatDayLabel = (dayString: string): string => {
  const date = new Date(dayString)
  const today = new Date()
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

const updateChart = () => {
  if (!chartCanvas.value) {
    // Retry after a short delay
    setTimeout(() => {
      if (chartCanvas.value) {
        updateChart()
      }
    }, 100)
    return
  }
  
  if (!stats.value.weeklyStats || stats.value.weeklyStats.length === 0) {
    return
  }
  
  // Destroy existing chart if it exists
  if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
  
  // Prepare data for last 7 days
  const today = new Date()
  const days: string[] = []
  const cardsTotalData: number[] = []
  const cardsCorrectData: number[] = []
  
  // Create a map of existing data
  const dataMap = new Map<string, { total: number; correct: number }>()
  stats.value.weeklyStats.forEach((stat: WeeklyStat) => {
    dataMap.set(stat.day, { total: stat.cards_completed, correct: stat.cards_correct || 0 })
  })
  
  // Generate last 7 days
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    const dayStr = date.toISOString().split('T')[0]
    days.push(dayStr)
    
    const data = dataMap.get(dayStr) || { total: 0, correct: 0 }
    cardsTotalData.push(data.total)
    cardsCorrectData.push(data.correct)
  }
  
  const labels = days.map(formatDayLabel)
  
  // Get theme colors from CSS variables
  const root = getComputedStyle(document.documentElement)
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  
  // Use darker, more contrasting colors for light theme
  let primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
  let successColor = root.getPropertyValue('--color-success').trim() || '#28a745'
  const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
  let textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
  let borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'
  
  // Adjust colors for light theme for better contrast
  if (!isDark) {
    // Use darker colors for better visibility on light background
    primaryColor = '#0056b3' // Darker blue
    successColor = '#1e7e34' // Darker green
    textSecondary = '#444444' // Darker grey for better readability
    borderColor = '#cccccc' // Darker border
  }
  
  // Convert hex to rgba
  const hexToRgba = (hex: string, alpha: number) => {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }
  
  // Create stacked bar chart: blue = total cards, green inside = correct cards
  chartInstance = new Chart(chartCanvas.value, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Accuracy',
          data: cardsCorrectData,
          backgroundColor: hexToRgba(successColor, isDark ? 0.9 : 0.8),
          borderWidth: 0
        },
        {
          label: 'Cards',
          data: cardsTotalData.map((total, idx) => {
            const correct = cardsCorrectData[idx] || 0
            return Math.max(0, total - correct)
          }),
          backgroundColor: hexToRgba(primaryColor, isDark ? 0.7 : 0.6),
          borderWidth: 0
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: 'index',
        intersect: false
      },
      plugins: {
        legend: {
          display: true,
          position: 'top',
          labels: {
            color: isDark ? textPrimary : '#222222',
            usePointStyle: true,
            padding: 15,
            font: {
              size: 12,
              weight: '500'
            }
          }
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          titleColor: '#fff',
          bodyColor: '#fff',
          borderColor: borderColor,
          borderWidth: 1,
          padding: 12,
          callbacks: {
            label: function(context) {
              const datasetLabel = context.dataset.label || ''
              const value = context.parsed.y || 0
              const total = cardsTotalData[context.dataIndex] || 0
              
              if (datasetLabel === 'Accuracy') {
                const percent = total > 0 ? ((value / total) * 100).toFixed(1) : '0'
                return `Accuracy: ${percent}% (${value} correct)`
              } else if (datasetLabel === 'Cards') {
                return `Cards: ${total} trained`
              }
              return `${datasetLabel}: ${value}`
            },
          }
        }
      },
      scales: {
        x: {
          stacked: true,
          ticks: {
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            }
          },
          grid: {
            color: borderColor,
            display: false
          }
        },
        y: {
          stacked: true,
          type: 'linear',
          display: true,
          beginAtZero: true,
          ticks: {
            stepSize: 1,
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            },
            callback: function(value) {
              return Number.isInteger(value) ? value : ''
            }
          },
          grid: {
            color: isDark ? borderColor : '#e0e0e0'
          }
        }
      }
    }
  })
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.dashboard-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.btn-refresh {
  background: var(--card-bg);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 20px;
  transition: all 0.2s;
  color: var(--text-primary);
}

.btn-refresh:hover:not(:disabled) {
  background: var(--input-bg);
  transform: rotate(180deg);
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dashboard-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.stat-card {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px var(--card-shadow);
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
  position: relative;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--card-shadow);
}

.stat-card-clickable {
  cursor: pointer;
}

.stat-card-clickable:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--card-shadow);
}

.stat-card-primary {
  border-left: 4px solid var(--color-primary);
}

.stat-card-info {
  border-left: 4px solid #3498db;
}

.stat-card-success {
  border-left: 4px solid var(--color-success);
}

.stat-card-warning {
  border-left: 4px solid #f39c12;
}

.stat-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-icon {
  font-size: 24px;
  line-height: 1;
}

.stat-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  margin: 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  flex: 1;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.stat-number {
  font-size: 42px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

/* Progress Section */
.progress-section {
  margin-top: 8px;
}

.progress-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 20px;
}

.progress-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: var(--text-primary);
}

.progress-value {
  font-weight: 600;
  color: var(--color-primary);
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--input-bg);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-fill-primary {
  background: var(--color-primary);
}

.progress-fill-success {
  background: var(--color-success);
}

/* Activity Section */
.activity-section {
  margin-top: 8px;
}

.activity-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.activity-card h2 {
  margin-bottom: 20px;
  font-size: 20px;
}

.activity-stats {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--input-bg);
  border-radius: 6px;
}

.activity-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.activity-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-primary);
}

/* Weekly Chart */
.weekly-chart-section {
  margin-top: 8px;
}

.weekly-chart-section h2 {
  margin-bottom: 20px;
  font-size: 20px;
}

.chart-container {
  position: relative;
  height: 300px;
  width: 100%;
}

.chart-empty {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .dashboard {
    padding: 12px;
  }
  
  .dashboard-header {
    margin-bottom: 16px;
  }
  
  .dashboard-header h1 {
    font-size: 24px;
  }
  
  .btn-refresh {
    padding: 6px 12px;
    font-size: 18px;
  }
  
  .dashboard-content {
    gap: 12px;
  }
  
  /* Compact Stats Grid - 2 columns */
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 6px;
  }
  
  .stat-card {
    padding: 8px;
    gap: 6px;
    border-radius: 6px;
  }
  
  .stat-header {
    gap: 6px;
  }
  
  .stat-icon {
    font-size: 18px;
  }
  
  .stat-header h3 {
    font-size: 10px;
    letter-spacing: 0.3px;
  }
  
  .stat-value-row {
    gap: 6px;
  }
  
  .stat-number {
    font-size: 28px;
    line-height: 1;
  }
  
  .stat-label {
    font-size: 10px;
  }
  
  /* Compact Progress Section */
  .progress-section {
    margin-top: 0;
  }
  
  .progress-section .card {
    padding: 12px;
  }
  
  .progress-section h2 {
    font-size: 16px;
    margin-bottom: 12px;
  }
  
  .progress-grid {
    gap: 12px;
    margin-top: 12px;
  }
  
  .progress-header {
    font-size: 12px;
  }
  
  .progress-value {
    font-size: 12px;
  }
  
  .progress-bar {
    height: 6px;
  }
  
  /* Compact Activity Section */
  .activity-section {
    margin-top: 0;
  }
  
  .activity-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 6px;
  }
  
  .activity-card {
    padding: 8px;
  }
  
  .activity-card h2 {
    font-size: 13px;
    margin-bottom: 8px;
    font-weight: 600;
  }
  
  .activity-stats {
    gap: 6px;
  }
  
  .activity-stat {
    padding: 8px;
    border-radius: 4px;
  }
  
  .activity-label {
    font-size: 11px;
  }
  
  .activity-value {
    font-size: 18px;
  }
  
  /* Compact Weekly Chart */
  .weekly-chart-section {
    margin-top: 0;
  }
  
  .weekly-chart-section .card {
    padding: 12px;
  }
  
  .weekly-chart-section h2 {
    font-size: 16px;
    margin-bottom: 12px;
  }
  
  .chart-container {
    height: 250px;
  }
}
</style>

