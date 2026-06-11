<template>
  <div class="dashboard lg-page">
    <!-- Brand header -->
    <div class="lg-home-header">
      <div>
        <div class="lg-home-brand">
          <span class="lg-home-logo">Linglow</span>
          <span class="lg-home-spark">✦</span>
        </div>
        <div class="lg-home-course">{{ targetLangDisplay }}</div>
      </div>
      <button @click="refreshData" class="lg-refresh-btn" :disabled="loading" :class="{ 'rotating': loading }">
        <LgIcon name="refresh" :s="18" c="var(--text)" />
      </button>
    </div>

    <div v-if="loading" class="lg-loading">{{ t('common.loading') }}</div>

    <div v-else class="dashboard-content">
      <!-- Ciudad Luminaria card -->
      <router-link to="/city" class="lg-city-card">
        <div class="lg-city-card-bg" />
        <div class="lg-city-card-overlay" />
        <div class="lg-city-card-content">
          <div>
            <div class="lg-city-card-title">{{ linglowProgress?.course?.city_name || t('navigation.city') }}</div>
            <div class="lg-city-card-kicker">{{ t('city.kicker') }}</div>
          </div>
          <div v-if="linglowProgress" class="lg-city-card-progress">
            <div class="lg-city-card-progress-label">
              {{ Math.round(linglowProgress.summary.progress_percent) }}% · {{ linglowProgress.summary.attempted_items }} {{ t('city.items') }}
            </div>
            <LgProgressBar :pct="linglowProgress.summary.progress_percent" :h="4" />
          </div>
        </div>
      </router-link>

      <!-- Твой путь сегодня -->
      <div class="lg-path-wrap">
        <div class="lg-path-lumi"><LgLumi :size="68" /></div>
        <div class="lg-path-card">
          <div class="lg-path-head">
            <span class="lg-path-title">{{ t('lg.todayPath') }}</span>
            <span class="lg-path-arrow">←</span>
          </div>
          <router-link v-if="stats.availableForTraining > 0" to="/training" class="lg-path-row">
            <div class="lg-icon-box">📖</div>
            <div class="lg-path-row-text">
              <div class="lg-list-row-title">{{ t('lg.repeatWords', { n: stats.availableForTraining }) }}</div>
              <div class="lg-list-row-sub">{{ t('lg.repeatWordsSub') }}</div>
            </div>
            <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
          </router-link>
          <router-link to="/city/daily-route" class="lg-path-row">
            <div class="lg-icon-box">🗺️</div>
            <div class="lg-path-row-text">
              <div class="lg-list-row-title">{{ t('lg.dailyRoute') }}</div>
              <div class="lg-list-row-sub">{{ t('lg.dailyRouteSub') }}</div>
            </div>
            <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
          </router-link>
          <router-link to="/learning/reading" class="lg-path-row">
            <div class="lg-icon-box">📚</div>
            <div class="lg-path-row-text">
              <div class="lg-list-row-title">{{ t('lg.readText') }}</div>
              <div class="lg-list-row-sub">{{ t('lg.readTextSub') }}</div>
            </div>
            <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
          </router-link>
          <router-link to="/chat" class="lg-path-row">
            <div class="lg-icon-box">💬</div>
            <div class="lg-path-row-text">
              <div class="lg-list-row-title">{{ t('lg.practiceChat') }}</div>
              <div class="lg-list-row-sub">{{ t('lg.practiceChatSub') }}</div>
            </div>
            <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
          </router-link>
        </div>
      </div>

      <!-- N слов пора повторить -->
      <div v-if="stats.availableForTraining > 0" class="lg-due-card">
        <div class="lg-due-counter">
          <span class="lg-due-num">{{ stats.availableForTraining }}</span>
          <span class="lg-due-unit">{{ t('common.words') }}</span>
        </div>
        <div class="lg-due-text">
          <div class="lg-list-row-title">{{ t('lg.wordsDueTitle', { n: stats.availableForTraining }) }}</div>
          <div class="lg-list-row-sub">{{ t('lg.wordsDueSub') }}</div>
        </div>
        <button class="lg-due-btn" @click="goToTraining">{{ t('lg.start') }}</button>
      </div>

      <!-- Offline notice -->
      <div v-if="offlineDashboard" class="lg-card lg-section-gap">
        <strong>{{ t('offline.modeTitle') }}</strong>
        <p class="lg-muted" style="margin: 6px 0 0">{{ t('offline.dashboardDescription') }}</p>
      </div>

      <!-- Grammar Statistics Section -->
      <div v-if="stats.grammarStats" class="grammar-stats-section">
        <router-link to="/learning/grammar" class="grammar-stats-link">
          <div class="statistics-block">
            <div class="stats-content">
              <!-- Course: circle = progress (цвет) + неопубликованная часть (другой цвет) -->
              <div class="stat-item percentage-item">
                <div class="stat-label">{{ t('dashboard.course') }}</div>
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
                        :stroke="getGrammarPercentageColor(stats.grammarStats.whole_course_completion_pct ?? 0)"
                        stroke-width="4"
                        stroke-opacity="0.2"
                      />
                      <circle
                        v-if="grammarUnpublishedSegmentLength > 0"
                        class="percentage-circle-small-fill percentage-circle-unpublished"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        stroke="var(--color-unpublished-segment, #94a3b8)"
                        stroke-width="4"
                        stroke-linecap="round"
                        :style="{
                          strokeDasharray: `${grammarUnpublishedSegmentLength} ${grammarSmallCircleCircumference}`,
                          strokeDashoffset: grammarUnpublishedSegmentDashOffset
                        }"
                      />
                      <circle
                        class="percentage-circle-small-fill"
                        cx="30"
                        cy="30"
                        r="26"
                        fill="none"
                        :stroke="getGrammarPercentageColor(stats.grammarStats.whole_course_completion_pct ?? 0)"
                        stroke-width="4"
                        stroke-linecap="round"
                        :style="{
                          strokeDasharray: grammarSmallCircleCircumference,
                          strokeDashoffset: getGrammarPercentageOffset(stats.grammarStats.whole_course_completion_pct ?? 0)
                        }"
                      />
                    </svg>
                    <div class="percentage-value-small">{{ stats.grammarStats.whole_course_completion_pct ?? 0 }}%</div>
                  </div>
                </div>
              </div>

              <!-- Current Level -->
              <div class="stat-item level-item">
                <div class="stat-label">{{ t('dashboard.level') }}</div>
                <div class="level-badge-compact" :class="grammarLevelBadgeClass">
                  {{ stats.grammarStats.confirmed_level || t('common.notStarted') }}
                </div>
              </div>

              <!-- Average Test Score -->
              <div class="stat-item percentage-item">
                <div class="stat-label">{{ t('dashboard.testAvg') }}</div>
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
              
              <!-- Chapters Progress -->
              <div class="stat-item chapters-item">
                <div class="stat-label">{{ t('dashboard.chapters') }}</div>
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
          <h2>{{ t('dashboard.yourProgress') }}</h2>
          <div class="progress-grid">
            <div class="progress-item">
              <div class="progress-header">
                <span>{{ t('dashboard.totalCards') }}</span>
                <span class="progress-value">{{ stats.totalCards }}</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: '100%' }"></div>
              </div>
            </div>

            <div class="progress-item">
              <div class="progress-header">
                <span>{{ t('dashboard.accuracy30Days') }}</span>
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
          <h2>{{ t('dashboard.weeklyActivity') }}</h2>
          <div v-if="stats.weeklyStats && stats.weeklyStats.length > 0" class="chart-container">
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
          <div v-if="stats.wordsAddedStats && stats.wordsAddedStats.length > 0" class="chart-container">
            <canvas ref="wordsChartCanvas"></canvas>
          </div>
          <div v-else class="chart-empty">
            <p>{{ t('dashboard.noCardsAdded') }}</p>
          </div>
        </div>
      </div>

      <!-- Совет от Lumi -->
      <LgLumiFact />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Chart, registerables } from 'chart.js'
import { useTheme } from '../composables/useTheme'
import { useAuth } from '../composables/useAuth'
import { useLocale } from '../composables/useLocale'
import { apiClient } from '../api/client'
import { grammarClient } from '../api/grammarClient'
import { wordTrainingClient } from '../api/wordTrainingClient'
import { courseClient, CourseProgress, LinglowHistory } from '../api/courseClient'
import { useCourse } from '../composables/useCourse'
import { useLearningConfig } from '../composables/useLearningConfig'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import LgProgressBar from '../components/linglow/LgProgressBar.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'

const { t } = useI18n()
const { targetLangDisplay, ensureLearningLoaded } = useLearningConfig()
ensureLearningLoaded()
const { currentLocale } = useLocale()
const { currentCourseCode } = useCourse()

const linglowProgress = ref<CourseProgress | null>(null)

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
  grammarStats?: any
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
const offlineDashboard = ref(false)

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

// Watch for locale changes and rebuild charts to update labels
watch(() => currentLocale.value, async () => {
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
    offlineDashboard.value = false
    let data: any
    if (typeof navigator !== 'undefined' && navigator.onLine === false) {
      const [wordStats, grammarStats] = await Promise.all([
        wordTrainingClient.getDashboard().catch(() => null),
        grammarClient.getStatistics().catch(() => null),
      ])
      data = {
        due_count: wordStats?.due_count || 0,
        total_cards: wordStats?.total_cards || 0,
        available_for_training: wordStats?.available_for_training || 0,
        grammar_stats: grammarStats,
      }
      offlineDashboard.value = true
    } else {
      let history: LinglowHistory | null = null;
      [data] = await Promise.all([
        apiClient.request('/api/dashboard'),
        courseClient.getProgress(currentCourseCode.value || undefined).then(p => { linglowProgress.value = p }).catch(() => {}),
        courseClient.getHistory({ courseCode: currentCourseCode.value || undefined, days: 7 }).then(h => { history = h }).catch(() => {}),
      ])
      // On the unified Linglow DB the legacy charts are empty; fall back to canonical history.
      const legacyWeekly = data.weekly_stats || []
      const legacyWordsAdded = data.words_added_stats || []
      if (history) {
        if (legacyWeekly.length === 0 && (history as LinglowHistory).weekly_stats?.length > 0) {
          data.weekly_stats = (history as LinglowHistory).weekly_stats
        }
        if (legacyWordsAdded.length === 0 && (history as LinglowHistory).words_added_stats?.length > 0) {
          data.words_added_stats = (history as LinglowHistory).words_added_stats
        }
        if (!(data.accuracy_percent > 0) && (history as LinglowHistory).accuracy_percent > 0) {
          data.accuracy_percent = (history as LinglowHistory).accuracy_percent
        }
      }
    }
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
    // charts are updated via watchers on weeklyStats / wordsAddedStats
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

const grammarUnpublishedSegmentLength = computed(() => {
  const g = stats.value?.grammarStats
  const total = Number(g?.total_chapters_in_course)
  const published = Number(g?.total_chapters)
  if (!Number.isFinite(total) || !Number.isFinite(published) || total <= 0) return 0
  const unpublishedPct = ((total - published) / total) * 100
  return (unpublishedPct / 100) * grammarSmallCircleCircumference.value
})

// Неопубликованный сегмент в конце круга: отрицательный offset = сегмент в конце пути (кроссбраузерно)
const grammarUnpublishedSegmentDashOffset = computed(() => {
  const len = grammarUnpublishedSegmentLength.value
  if (len <= 0) return 0
  return len - grammarSmallCircleCircumference.value
})

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
    return t('common.today')
  } else if (diffDays === 1) {
    return t('common.yesterday')
  } else {
    // Use locale-aware formatting
    const locale = currentLocale.value === 'ru' ? 'ru-RU' : 'en-US'
    return date.toLocaleDateString(locale, { weekday: 'short', day: 'numeric' })
  }
}

const updateChart = () => {
  try {
  if (!chartCanvas.value) {
    setTimeout(() => { if (chartCanvas.value) updateChart() }, 100)
    return
  }

  if (!stats.value.weeklyStats || stats.value.weeklyStats.length === 0) {
    return
  }

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
  chartInstance = new Chart(chartCanvas.value!, {
    type: 'bar',
    data: {
      labels: labels,
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
              
              if (datasetLabel === t('dashboard.chartAccuracy')) {
                const percent = total > 0 ? ((value / total) * 100).toFixed(1) : '0'
                return t('dashboard.chartAccuracyTooltip', { percent, correct: value })
              } else if (datasetLabel === t('dashboard.chartCards')) {
                return t('dashboard.chartCardsTooltip', total, { count: total })
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

  if (!stats.value.wordsAddedStats || stats.value.wordsAddedStats.length === 0) {
    return
  }

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
              return t('dashboard.chartCardsAddedTooltip', { count: value })
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
  } catch (e) {
    console.warn('updateWordsChart failed:', e)
    wordsChartInstance = null
  }
}

watch(currentCourseCode, () => {
  if (isAuthenticated.value) loadData()
})

// Watch for authentication state and load data when authenticated
watch(isAuthenticated, (authenticated) => {
  if (authenticated) {
    loadData()
  }
}, { immediate: true })

onMounted(() => {
  if (isAuthenticated.value) loadData()
})
</script>

<style scoped>
/* Linglow progress block */
.linglow-progress-section { margin-top: 4px; }
.linglow-progress-link { text-decoration: none; color: inherit; display: block; }
.linglow-progress-block {
  border: 2px solid var(--color-primary, #3b82f6);
  border-radius: 10px;
  padding: 14px 18px;
  background: linear-gradient(135deg, var(--card-bg) 0%, rgba(59, 130, 246, 0.05) 100%);
  transition: box-shadow 0.2s;
}
.linglow-progress-link:hover .linglow-progress-block { box-shadow: 0 4px 14px rgba(59, 130, 246, 0.15); }
.linglow-progress-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  text-align: center;
}
.linglow-metric-value { display: block; font-size: 1.5rem; font-weight: 800; color: var(--text-primary); }
.linglow-metric-label { display: block; font-size: 0.78rem; color: var(--text-secondary); font-weight: 600; text-transform: uppercase; }
.linglow-progress-hint { margin: 10px 0 0; font-size: 0.82rem; color: var(--text-secondary); text-align: center; }
@media (max-width: 480px) {
  .linglow-progress-row { grid-template-columns: repeat(2, 1fr); }
}

.dashboard {
  width: min(880px, 100%);
  margin: 0 auto;
  box-sizing: border-box;
}

/* ─── New home header ─── */
.lg-home-header {
  padding: 22px 4px 0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}
.lg-home-brand {
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.lg-home-logo {
  font-family: 'Lora', serif;
  font-size: 36px;
  font-weight: 600;
  color: var(--text);
  letter-spacing: -0.02em;
  line-height: 1;
}
.lg-home-spark {
  color: var(--dorado);
  font-size: 16px;
  line-height: 1;
  padding-bottom: 2px;
}
.lg-home-course {
  font-size: 14px;
  color: var(--subtext);
  margin-top: 4px;
}
.lg-refresh-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
.lg-refresh-btn.rotating {
  animation: rotate 1s linear infinite;
}
.lg-refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ─── City card ─── */
.lg-city-card {
  display: block;
  position: relative;
  width: 100%;
  min-height: 118px;
  border-radius: 22px;
  overflow: hidden;
  border: 1px solid var(--border);
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  text-decoration: none;
  margin-bottom: 10px;
}
.lg-city-card-bg {
  position: absolute;
  inset: 0;
  background-image: url('../assets/linglow/map_city.jpg');
  background-size: cover;
  background-position: center;
}
.lg-city-card-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(255,249,237,0.98) 0%, rgba(255,249,237,0.80) 50%, rgba(255,249,237,0.05) 100%);
}
:root[data-theme="dark"] .lg-city-card-overlay {
  background: linear-gradient(90deg, rgba(16,28,21,0.97) 0%, rgba(16,28,21,0.78) 50%, rgba(16,28,21,0.12) 100%);
}
.lg-city-card-content {
  position: relative;
  z-index: 2;
  padding: 18px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 118px;
}
.lg-city-card-title {
  font-family: 'Lora', serif;
  font-size: 24px;
  font-weight: 600;
  color: var(--text);
  line-height: 1.15;
  margin-bottom: 5px;
}
.lg-city-card-kicker {
  font-size: 12px;
  color: var(--subtext);
}
.lg-city-card-progress {
  margin-top: 14px;
  max-width: 55%;
}
.lg-city-card-progress-label {
  font-size: 11px;
  color: var(--subtext);
  margin-bottom: 5px;
}

/* ─── Today path ─── */
.lg-path-wrap {
  position: relative;
  margin-bottom: 10px;
}
.lg-path-lumi {
  position: absolute;
  top: -22px;
  right: 8px;
  z-index: 10;
  pointer-events: none;
}
.lg-path-card {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: var(--shadow-card);
}
.lg-path-head {
  padding: 16px 80px 14px 18px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.lg-path-title {
  font-family: 'Lora', serif;
  font-size: 20px;
  font-weight: 600;
  color: var(--text);
  line-height: 1;
}
.lg-path-arrow {
  color: var(--dorado);
  font-size: 16px;
  line-height: 1;
}
.lg-path-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 13px 18px;
  border-top: 1px solid var(--border);
  text-decoration: none;
  cursor: pointer;
}
.lg-path-row-text { flex: 1; }

/* ─── Words due card ─── */
.lg-due-card {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 22px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  box-shadow: var(--shadow-card);
  margin-bottom: 10px;
}
.lg-due-counter {
  width: 58px;
  height: 60px;
  border-radius: 14px;
  flex-shrink: 0;
  background: var(--surface-2);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.lg-due-num {
  font-family: 'Lora', serif;
  font-size: 24px;
  font-weight: 600;
  color: var(--text);
  line-height: 1;
}
.lg-due-unit {
  font-size: 9px;
  color: var(--subtext);
  margin-top: 2px;
}
.lg-due-text { flex: 1; }
.lg-due-btn {
  padding: 12px 20px;
  border-radius: 14px;
  border: 1px solid var(--btn-border);
  background: var(--btn-gradient);
  color: white;
  flex-shrink: 0;
  font-weight: 600;
  font-size: 15px;
  cursor: pointer;
  white-space: nowrap;
  box-shadow: var(--btn-shadow);
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
  justify-items: center;
}

.stat-item {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: center;
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
  
  /* Compact Grammar Stats Section */
  .grammar-stats-section {
    margin-top: 0;
  }
  
  .statistics-block {
    padding: 10px 12px;
  }
  
  .stats-content {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
    justify-items: stretch;
    padding: 0 4px;
    max-width: 387px;
    margin: auto;
  }

  .stat-item {
    width: 100%;
    min-width: 0;
    flex-direction: row-reverse;
    justify-content: flex-end;
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
    min-width: auto;
    text-align: left;
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

.dashboard-offline-card {
  border: 1px solid rgba(180, 83, 9, 0.35);
  background: linear-gradient(135deg, var(--card-bg), rgba(180, 83, 9, 0.08));
}

.dashboard-offline-card p {
  margin: 6px 0 0;
  color: var(--text-secondary);
}
</style>
