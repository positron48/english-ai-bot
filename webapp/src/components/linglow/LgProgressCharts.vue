<template>
  <div class="progress-charts">
    <!-- Progress Section -->
    <div class="progress-section">
      <div class="card">
        <h2>{{ t('dashboard.yourProgress') }}</h2>
        <div class="progress-grid">
          <div class="progress-item">
            <div class="progress-header">
              <span>{{ t('dashboard.totalCards') }}</span>
              <span class="progress-value">{{ totalCards }}</span>
            </div>
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: '100%' }"></div>
            </div>
          </div>

          <div class="progress-item">
            <div class="progress-header">
              <span>{{ t('dashboard.accuracy30Days') }}</span>
              <span class="progress-value">{{ formatPercent(accuracyPercent) }}%</span>
            </div>
            <div class="progress-bar">
              <div
                class="progress-fill progress-fill-success"
                :style="{ width: accuracyPercent + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Weekly Chart -->
    <div class="weekly-chart-section">
      <div class="card">
        <h2>{{ t('dashboard.weeklyActivity') }}</h2>
        <div v-if="weeklyStats.length > 0" class="chart-container">
          <canvas ref="chartCanvas"></canvas>
        </div>
        <div v-else class="chart-empty">
          <p>{{ t('dashboard.noTrainingData') }}</p>
        </div>
      </div>
    </div>

    <!-- Words Added Chart -->
    <div class="weekly-chart-section">
      <div class="card">
        <h2>{{ t('dashboard.cardsAdded7Days') }}</h2>
        <div v-if="wordsAddedStats.length > 0" class="chart-container">
          <canvas ref="wordsChartCanvas"></canvas>
        </div>
        <div v-else class="chart-empty">
          <p>{{ t('dashboard.noCardsAdded') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart, registerables } from 'chart.js'
import { useTheme } from '../../composables/useTheme'
import { useLocale } from '../../composables/useLocale'
import { useCourse } from '../../composables/useCourse'
import { apiClient } from '../../api/client'
import { courseClient, LinglowHistory } from '../../api/courseClient'

Chart.register(...registerables)

const { t } = useI18n()
const { theme } = useTheme()
const { currentLocale } = useLocale()
const { currentCourseCode } = useCourse()

interface WeeklyStat {
  day: string
  cards_completed: number
  cards_correct: number
}

interface WordsAddedStat {
  day: string
  words_added: number
}

const totalCards = ref(0)
const accuracyPercent = ref(0)
const weeklyStats = ref<WeeklyStat[]>([])
const wordsAddedStats = ref<WordsAddedStat[]>([])

const chartCanvas = ref<HTMLCanvasElement | null>(null)
const wordsChartCanvas = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null
let wordsChartInstance: Chart | null = null

const props = defineProps<{
  // When the parent screen already fetched these via an aggregate endpoint, pass them in to
  // avoid a duplicate /api/dashboard + /history round trip. Omitted → self-fetch (standalone use).
  dashboard?: any
  history?: LinglowHistory | null
  // When true, the parent owns data loading (props arrive asynchronously): skip the self-fetch
  // entirely and just render whatever the `dashboard` watch pushes in.
  external?: boolean
}>()

const formatPercent = (value: number): string => value.toFixed(1)

// applyData folds a dashboard payload plus optional canonical history into the chart refs.
const applyData = (data: any, history: LinglowHistory | null) => {
  if (!data) return
  // On the unified Linglow DB the legacy charts are empty; fall back to canonical history.
  let weekly = data.weekly_stats || []
    let wordsAdded = data.words_added_stats || []
    let accuracy = data.accuracy_percent || 0
    if (history) {
      if (weekly.length === 0 && (history as LinglowHistory).weekly_stats?.length > 0) {
        weekly = (history as LinglowHistory).weekly_stats
      }
      if (wordsAdded.length === 0 && (history as LinglowHistory).words_added_stats?.length > 0) {
        wordsAdded = (history as LinglowHistory).words_added_stats
      }
      if (!(accuracy > 0) && (history as LinglowHistory).accuracy_percent > 0) {
        accuracy = (history as LinglowHistory).accuracy_percent
      }
    }
  totalCards.value = data.total_cards || 0
  accuracyPercent.value = accuracy
  weeklyStats.value = weekly
  wordsAddedStats.value = wordsAdded
}

const loadData = async () => {
  // Parent owns loading (aggregate endpoint) — never self-fetch; the props watch applies data.
  if (props.external) {
    if (props.dashboard) applyData(props.dashboard, props.history ?? null)
    return
  }
  apiClient.loadTokens()
  try {
    let history: LinglowHistory | null = null
    const [data] = await Promise.all([
      apiClient.request('/api/dashboard') as Promise<any>,
      courseClient.getHistory({ courseCode: currentCourseCode.value || undefined, days: 7 }).then(h => { history = h }).catch(() => {}),
    ])
    applyData(data, history)
  } catch (error) {
    console.error('Failed to load progress charts:', error)
  }
}

// Re-apply when the parent pushes fresh aggregate data.
watch(() => props.dashboard, (d) => { if (d) applyData(d, props.history ?? null) })

watch(() => weeklyStats.value, async (newStats) => {
  if (newStats && newStats.length > 0) {
    await nextTick()
    setTimeout(() => { updateChart() }, 150)
  }
}, { deep: true })

watch(() => wordsAddedStats.value, async (newStats) => {
  if (newStats && newStats.length > 0) {
    await nextTick()
    setTimeout(() => { updateWordsChart() }, 150)
  }
}, { deep: true })

watch(() => theme.value, async () => {
  if (weeklyStats.value.length > 0) { await nextTick(); setTimeout(() => updateChart(), 100) }
  if (wordsAddedStats.value.length > 0) { await nextTick(); setTimeout(() => updateWordsChart(), 100) }
})

watch(() => currentLocale.value, async () => {
  if (weeklyStats.value.length > 0) { await nextTick(); setTimeout(() => updateChart(), 100) }
  if (wordsAddedStats.value.length > 0) { await nextTick(); setTimeout(() => updateWordsChart(), 100) }
})

watch(currentCourseCode, () => { loadData() })

const formatDateLocal = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatDayLabel = (dayString: string): string => {
  const date = new Date(dayString + 'T00:00:00')
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  date.setHours(0, 0, 0, 0)
  const diffTime = today.getTime() - date.getTime()
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  if (diffDays === 0) return t('common.today')
  if (diffDays === 1) return t('common.yesterday')
  const locale = currentLocale.value === 'ru' ? 'ru-RU' : 'en-US'
  return date.toLocaleDateString(locale, { weekday: 'short', day: 'numeric' })
}

const updateChart = () => {
  try {
    if (!chartCanvas.value) {
      setTimeout(() => { if (chartCanvas.value) updateChart() }, 100)
      return
    }
    if (!weeklyStats.value || weeklyStats.value.length === 0) return

    if (chartInstance) { chartInstance.destroy(); chartInstance = null }

    const today = new Date()
    const days: string[] = []
    const cardsTotalData: number[] = []
    const cardsCorrectData: number[] = []

    const dataMap = new Map<string, { total: number; correct: number }>()
    weeklyStats.value.forEach((stat: WeeklyStat) => {
      dataMap.set(stat.day, { total: stat.cards_completed, correct: stat.cards_correct || 0 })
    })

    for (let i = 6; i >= 0; i--) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      date.setHours(0, 0, 0, 0)
      const dayStr = formatDateLocal(date)
      days.push(dayStr)
      const data = dataMap.get(dayStr) || { total: 0, correct: 0 }
      cardsTotalData.push(data.total)
      cardsCorrectData.push(data.correct)
    }

    const labels = days.map(formatDayLabel)
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

    chartInstance = new Chart(chartCanvas.value!, {
      type: 'bar',
      data: {
        labels,
        datasets: [
          {
            label: t('dashboard.chartAccuracy'),
            data: cardsCorrectData,
            backgroundColor: hexToRgba(successColor, isDark ? 0.9 : 0.8),
            borderWidth: 0
          },
          {
            label: t('dashboard.chartCards'),
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
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: {
            display: true,
            position: 'top',
            labels: {
              color: isDark ? textPrimary : '#222222',
              usePointStyle: true,
              padding: 15,
              font: { size: 12, weight: 500 }
            }
          },
          tooltip: {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor,
            borderWidth: 1,
            padding: 12,
            callbacks: {
              label: function (context) {
                const datasetLabel = context.dataset.label || ''
                const value = context.parsed.y || 0
                const total = cardsTotalData[context.dataIndex] || 0
                if (datasetLabel === t('dashboard.chartAccuracy')) {
                  const percent = total > 0 ? ((value / total) * 100).toFixed(1) : '0'
                  return t('dashboard.chartAccuracyTooltip', { percent, correct: value })
                } else if (datasetLabel === t('dashboard.chartCards')) {
                  return (t as any)('dashboard.chartCardsTooltip', total, { count: total })
                }
                return `${datasetLabel}: ${value}`
              },
            }
          }
        },
        scales: {
          x: {
            stacked: true,
            ticks: { color: isDark ? textSecondary : '#555555', font: { size: 11 } },
            grid: { color: borderColor, display: false }
          },
          y: {
            stacked: true,
            type: 'linear',
            display: true,
            beginAtZero: true,
            ticks: {
              stepSize: 1,
              color: isDark ? textSecondary : '#555555',
              font: { size: 11 },
              callback: function (value) { return Number.isInteger(value) ? value : '' }
            },
            grid: { color: isDark ? borderColor : '#e0e0e0' }
          }
        }
      }
    })
  } catch (e) {
    console.warn('updateChart failed:', e)
    chartInstance = null
  }
}

const updateWordsChart = () => {
  try {
    if (!wordsChartCanvas.value) {
      setTimeout(() => { if (wordsChartCanvas.value) updateWordsChart() }, 100)
      return
    }
    if (!wordsAddedStats.value || wordsAddedStats.value.length === 0) return

    if (wordsChartInstance) { wordsChartInstance.destroy(); wordsChartInstance = null }

    const today = new Date()
    const days: string[] = []
    const wordsAddedData: number[] = []

    const dataMap = new Map<string, number>()
    wordsAddedStats.value.forEach((stat: WordsAddedStat) => {
      dataMap.set(stat.day, stat.words_added)
    })

    for (let i = 6; i >= 0; i--) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      date.setHours(0, 0, 0, 0)
      const dayStr = formatDateLocal(date)
      days.push(dayStr)
      wordsAddedData.push(dataMap.get(dayStr) || 0)
    }

    const labels = days.map(formatDayLabel)
    const root = getComputedStyle(document.documentElement)
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark'

    let primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
    const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
    let textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
    let borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'

    if (!isDark) {
      primaryColor = '#0056b3'
      textSecondary = '#444444'
      borderColor = '#cccccc'
    }

    const hexToRgba = (hex: string, alpha: number) => {
      const r = parseInt(hex.slice(1, 3), 16)
      const g = parseInt(hex.slice(3, 5), 16)
      const b = parseInt(hex.slice(5, 7), 16)
      return `rgba(${r}, ${g}, ${b}, ${alpha})`
    }

    wordsChartInstance = new Chart(wordsChartCanvas.value, {
      type: 'bar',
      data: {
        labels,
        datasets: [
          {
            label: t('dashboard.chartCardsAdded'),
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
        interaction: { mode: 'index', intersect: false },
        plugins: {
          legend: {
            display: true,
            position: 'top',
            labels: {
              color: isDark ? textPrimary : '#222222',
              usePointStyle: true,
              padding: 15,
              font: { size: 12, weight: 500 }
            }
          },
          tooltip: {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: '#fff',
            bodyColor: '#fff',
            borderColor,
            borderWidth: 1,
            padding: 12,
            callbacks: {
              label: function (context) {
                const value = context.parsed.y || 0
                return t('dashboard.chartCardsAddedTooltip', { count: value })
              },
            }
          }
        },
        scales: {
          x: {
            ticks: { color: isDark ? textSecondary : '#555555', font: { size: 11 } },
            grid: { color: borderColor, display: false }
          },
          y: {
            type: 'linear',
            display: true,
            beginAtZero: true,
            ticks: {
              stepSize: 1,
              color: isDark ? textSecondary : '#555555',
              font: { size: 11 },
              callback: function (value) { return Number.isInteger(value) ? value : '' }
            },
            grid: { color: isDark ? borderColor : '#e0e0e0' }
          }
        }
      }
    })
  } catch (e) {
    console.warn('updateWordsChart failed:', e)
    wordsChartInstance = null
  }
}

onMounted(loadData)
</script>

<style scoped>
.progress-charts {
  display: flex;
  flex-direction: column;
  gap: 16px;
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

.progress-fill-success {
  background: var(--color-success);
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
  .progress-section .card,
  .weekly-chart-section .card {
    padding: 12px;
  }
  .progress-section h2,
  .weekly-chart-section h2 {
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
  .chart-container {
    height: 250px;
  }
}
</style>
