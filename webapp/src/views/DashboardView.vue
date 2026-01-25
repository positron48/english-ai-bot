<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Dashboard</h1>
      <button @click="refreshData" class="btn-refresh" :disabled="loading" :class="{ 'rotating': loading }">
        <Icon name="refresh" />
      </button>
    </div>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else class="dashboard-content">
      <!-- Main Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card stat-card-primary stat-card-clickable" @click="goToTraining">
          <div class="stat-header">
            <Icon name="book" class="stat-icon" />
            <h3>Available for Training</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.availableForTraining }}</p>
            <p class="stat-label">Cards available</p>
          </div>
        </div>

        <div class="stat-card stat-card-info">
          <div class="stat-header">
            <Icon name="sparkles" class="stat-icon" />
            <h3>New Cards</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.newCount }}</p>
            <p class="stat-label">Not started</p>
          </div>
        </div>

        <div class="stat-card stat-card-success">
          <div class="stat-header">
            <Icon name="book-open" class="stat-icon" />
            <h3>In Learning</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.learningCount }}</p>
            <p class="stat-label">Being learned</p>
          </div>
        </div>

        <div class="stat-card stat-card-warning">
          <div class="stat-header">
            <Icon name="refresh" class="stat-icon" />
            <h3>In Review</h3>
          </div>
          <div class="stat-value-row">
            <p class="stat-number">{{ stats.reviewCount }}</p>
            <p class="stat-label">Mastered cards</p>
          </div>
        </div>
      </div>

      <!-- Grammar Statistics Section -->
      <div v-if="stats.grammarStats" class="grammar-stats-section">
        <router-link to="/learning/grammar" class="grammar-stats-link">
          <div class="statistics-block">
            <div class="stats-content">
              <!-- Current Level (Left) -->
              <div class="stat-item level-item">
                <div class="stat-label">Level</div>
                <div class="level-badge-compact" :class="grammarLevelBadgeClass">
                  {{ stats.grammarStats.confirmed_level || 'Not started' }}
                </div>
              </div>
              
              <!-- Course Completion Percentage -->
              <div class="stat-item percentage-item">
                <div class="stat-label">Course</div>
                <div class="percentage-wrapper">
                  <div class="percentage-circle-small-wrapper">
                    <svg class="percentage-circle-small" viewBox="0 0 60 60">
                      <circle
                        class="percentage-circle-small-bg"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        stroke="var(--bg-tertiary)"
                        stroke-width="4"
                      />
                      <circle
                        class="percentage-circle-small-outline"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        :stroke="getGrammarPercentageColor(stats.grammarStats.course_completion_pct || 0)"
                        stroke-width="4"
                        stroke-opacity="0.2"
                      />
                      <circle
                        class="percentage-circle-small-fill"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        :stroke="getGrammarPercentageColor(stats.grammarStats.course_completion_pct || 0)"
                        stroke-width="4"
                        stroke-linecap="round"
                        :style="{
                          strokeDasharray: grammarSmallCircleCircumference,
                          strokeDashoffset: getGrammarPercentageOffset(stats.grammarStats.course_completion_pct || 0)
                        }"
                      />
                    </svg>
                    <div class="percentage-value-small">{{ stats.grammarStats.course_completion_pct || 0 }}%</div>
                  </div>
                </div>
              </div>
              
              <!-- Average Test Score -->
              <div class="stat-item percentage-item">
                <div class="stat-label">Test (avg.)</div>
                <div class="percentage-wrapper">
                  <div class="percentage-circle-small-wrapper">
                    <svg class="percentage-circle-small" viewBox="0 0 60 60">
                      <circle
                        class="percentage-circle-small-bg"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        stroke="var(--bg-tertiary)"
                        stroke-width="4"
                      />
                      <circle
                        class="percentage-circle-small-outline"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        :stroke="getGrammarPercentageColor(stats.grammarStats.average_test_score || 0)"
                        stroke-width="4"
                        stroke-opacity="0.2"
                      />
                      <circle
                        class="percentage-circle-small-fill"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        :stroke="getGrammarPercentageColor(stats.grammarStats.average_test_score || 0)"
                        stroke-width="4"
                        stroke-linecap="round"
                        :style="{
                          strokeDasharray: grammarSmallCircleCircumference,
                          strokeDashoffset: getGrammarPercentageOffset(stats.grammarStats.average_test_score || 0)
                        }"
                      />
                    </svg>
                    <div class="percentage-value-small">{{ stats.grammarStats.average_test_score || 0 }}%</div>
                  </div>
                </div>
              </div>
              
              <!-- Chapters Progress (Right) -->
              <div class="stat-item chapters-item">
                <div class="stat-label">Chapters</div>
                <div class="chapters-value-compact">
                  <span class="chapters-number">{{ stats.grammarStats.passed_chapters || 0 }}</span>
                  <span class="chapters-separator">/</span>
                  <span class="chapters-total">{{ stats.grammarStats.total_chapters || 0 }}</span>
                </div>
              </div>
            </div>
          </div>
        </router-link>
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

      <!-- Words Added Chart -->
      <div class="weekly-chart-section">
        <div class="card">
          <h2>Cards Added (7 days)</h2>
          <div v-if="stats.wordsAddedStats && stats.wordsAddedStats.length > 0" class="chart-container">
            <canvas ref="wordsChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>No cards added in the last week</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { Chart, registerables } from 'chart.js'
import { useTheme } from '../composables/useTheme'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

Chart.register(...registerables)

const router = useRouter()
const { theme } = useTheme()
const { isAuthenticated } = useAuth()
const chartCanvas = ref<HTMLCanvasElement | null>(null)
const wordsChartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null
let wordsChartInstance: Chart | null = null

interface WeeklyStat {
  day: string
  cards_completed: number
  cards_correct: number
}

interface WordsAddedStat {
  day: string
  words_added: number
}

interface DashboardStats {
  dueCount: number
  newCount: number
  learningCount: number
  reviewCount: number
  totalCards: number
  availableForTraining: number
  accuracyPercent: number
  weeklyStats: WeeklyStat[]
  wordsAddedStats: WordsAddedStat[]
}

const stats = ref<DashboardStats>({
  dueCount: 0,
  newCount: 0,
  learningCount: 0,
  reviewCount: 0,
  totalCards: 0,
  availableForTraining: 0,
  accuracyPercent: 0,
  weeklyStats: [],
  wordsAddedStats: []
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

// Watch for changes in wordsAddedStats and update chart
watch(() => stats.value.wordsAddedStats, async (newStats) => {
  if (newStats && newStats.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateWordsChart()
    }, 150)
  }
}, { deep: true })

// Watch for theme changes and rebuild charts
watch(() => theme.value, async () => {
  if (stats.value.weeklyStats && stats.value.weeklyStats.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateChart()
    }, 100)
  }
  if (stats.value.wordsAddedStats && stats.value.wordsAddedStats.length > 0) {
    await nextTick()
    setTimeout(() => {
      updateWordsChart()
    }, 100)
  }
})

const loadData = async () => {
  // Ensure tokens are loaded before making request
  apiClient.loadTokens()
  
  // Don't make request if not authenticated
  if (!isAuthenticated.value) {
    console.warn('Not authenticated, skipping dashboard data load')
    return
  }
  
  try {
    loading.value = true
    const data = await apiClient.request('/api/dashboard')
    stats.value = {
      dueCount: data.due_count || 0,
      newCount: data.new_count || 0,
      learningCount: data.learning_count || 0,
      reviewCount: data.review_count || 0,
      totalCards: data.total_cards || 0,
      availableForTraining: data.available_for_training || 0,
      accuracyPercent: data.accuracy_percent || 0,
      weeklyStats: data.weekly_stats || [],
      wordsAddedStats: data.words_added_stats || [],
      grammarStats: data.grammar_stats || null
    }
    await nextTick()
    updateChart()
    updateWordsChart()
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

// Grammar statistics computed properties
const grammarLevelBadgeClass = computed(() => {
  const level = stats.value.grammarStats?.confirmed_level || ''
  if (level.startsWith('C')) return 'badge-c2'
  if (level.startsWith('B')) return 'badge-b'
  if (level.startsWith('A')) return 'badge-a'
  return 'badge-none'
})

const grammarSmallCircleCircumference = computed(() => 2 * Math.PI * 26)

const getGrammarPercentageOffset = (percent: number): number => {
  const progress = Math.max(0, Math.min(100, percent)) / 100
  return grammarSmallCircleCircumference.value * (1 - progress)
}

const getGrammarPercentageColor = (percent: number): string => {
  if (percent >= 90) return '#10b981' // green
  if (percent >= 70) return '#3b82f6' // blue
  if (percent >= 50) return '#f59e0b' // orange
  if (percent >= 25) return '#f97316' // orange-red
  return '#ef4444' // red
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

// Format date to YYYY-MM-DD in local timezone (not UTC)
const formatDateLocal = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatDayLabel = (dayString: string): string => {
  const date = new Date(dayString + 'T00:00:00') // Parse as local date
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
  
  // Generate last 7 days in local timezone
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    date.setHours(0, 0, 0, 0) // Reset to start of day in local time
    const dayStr = formatDateLocal(date)
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

const updateWordsChart = () => {
  if (!wordsChartCanvas.value) {
    // Retry after a short delay
    setTimeout(() => {
      if (wordsChartCanvas.value) {
        updateWordsChart()
      }
    }, 100)
    return
  }
  
  if (!stats.value.wordsAddedStats || stats.value.wordsAddedStats.length === 0) {
    return
  }
  
  // Destroy existing chart if it exists
  if (wordsChartInstance) {
    wordsChartInstance.destroy()
    wordsChartInstance = null
  }
  
  // Prepare data for last 7 days
  const today = new Date()
  const days: string[] = []
  const wordsAddedData: number[] = []
  
  // Create a map of existing data
  const dataMap = new Map<string, number>()
  stats.value.wordsAddedStats.forEach((stat: WordsAddedStat) => {
    dataMap.set(stat.day, stat.words_added)
  })
  
  // Generate last 7 days in local timezone
  for (let i = 6; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    date.setHours(0, 0, 0, 0) // Reset to start of day in local time
    const dayStr = formatDateLocal(date)
    days.push(dayStr)
    
    const count = dataMap.get(dayStr) || 0
    wordsAddedData.push(count)
  }
  
  const labels = days.map(formatDayLabel)
  
  // Get theme colors from CSS variables
  const root = getComputedStyle(document.documentElement)
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  
  // Use darker, more contrasting colors for light theme
  let primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
  const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
  let textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
  let borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'
  
  // Adjust colors for light theme for better contrast
  if (!isDark) {
    // Use darker colors for better visibility on light background
    primaryColor = '#0056b3' // Darker blue
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
  
  // Create bar chart for words added
  wordsChartInstance = new Chart(wordsChartCanvas.value, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [
        {
          label: 'Cards Added',
          data: wordsAddedData,
          backgroundColor: hexToRgba(primaryColor, isDark ? 0.8 : 0.7),
          borderColor: primaryColor,
          borderWidth: 1
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
              const value = context.parsed.y || 0
              return `Cards added: ${value}`
            },
          }
        }
      },
      scales: {
        x: {
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

// Watch for authentication state and load data when authenticated
watch(isAuthenticated, (authenticated) => {
  if (authenticated) {
    loadData()
  }
}, { immediate: true })

onMounted(() => {
  // Also try to load on mount if already authenticated
  // (watch will handle it, but this ensures it happens)
  if (isAuthenticated.value) {
    loadData()
  }
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
}

.btn-refresh.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
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
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
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

/* Grammar Statistics Section */
.grammar-stats-section {
  margin-top: 8px;
}

.grammar-stats-link {
  text-decoration: none;
  color: inherit;
  display: block;
}

.statistics-block {
  padding: 12px 16px;
  background: linear-gradient(135deg, var(--card-bg) 0%, rgba(var(--color-primary-rgb, 59, 130, 246), 0.05) 100%);
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.2s ease;
  cursor: pointer;
}

.grammar-stats-link:hover .statistics-block {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.stats-content {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  align-items: center;
}

.stat-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
}

.stat-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  white-space: nowrap;
  min-width: 50px;
}

/* Level Item */
.level-item {
  justify-content: flex-start;
}

.level-badge-compact {
  padding: 6px 12px;
  font-size: 20px;
  font-weight: 700;
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.level-badge-compact.badge-a {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.level-badge-compact.badge-b {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.3);
}

.level-badge-compact.badge-c2 {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
}

.level-badge-compact.badge-none {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

/* Percentage Items */
.percentage-item {
  justify-content: flex-start;
}

.percentage-wrapper {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.percentage-circle-small-wrapper {
  position: relative;
  width: 60px;
  height: 60px;
}

.percentage-circle-small {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.percentage-circle-small-bg {
  opacity: 0.2;
}

.percentage-circle-small-outline {
  transition: stroke 0.3s ease;
}

.percentage-circle-small-fill {
  transition: stroke-dashoffset 0.6s ease, stroke 0.3s ease;
}

.percentage-value-small {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  text-align: center;
}

/* Chapters Item */
.chapters-item {
  justify-content: flex-start;
}

.chapters-value-compact {
  display: flex;
  align-items: baseline;
  gap: 4px;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.chapters-value-compact .chapters-number {
  color: var(--color-primary);
}

.chapters-value-compact .chapters-separator {
  color: var(--text-secondary);
  font-weight: 400;
}

.chapters-value-compact .chapters-total {
  color: var(--text-secondary);
  font-weight: 500;
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
    padding: 12px 8px;
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
  
  /* Compact Grammar Stats Section */
  .grammar-stats-section {
    margin-top: 0;
  }
  
  .statistics-block {
    padding: 10px 12px;
  }
  
  .stats-content {
    grid-template-columns: repeat(2, 1fr);
    gap: 12px;
  }
  
  .percentage-circle-small-wrapper {
    width: 50px;
    height: 50px;
  }
  
  .percentage-value-small {
    font-size: 12px;
  }
  
  .chapters-value-compact {
    font-size: 18px;
  }
  
  .level-badge-compact {
    font-size: 18px;
    padding: 5px 10px;
  }
  
  .stat-label {
    font-size: 10px;
    min-width: 40px;
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

