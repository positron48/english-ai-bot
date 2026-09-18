import { ref, watch, onUnmounted, type Ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart, registerables } from 'chart.js'
import { wordTrainingClient } from '../api/wordTrainingClient'

Chart.register(...registerables)

export function useTrainingUpcoming(upcomingChartCanvas: Ref<HTMLCanvasElement | null>) {
  const { t, locale } = useI18n()
  const upcomingCardsLoaded = ref(false)
  const upcomingCardsData = ref<Record<string, { date: string; label: string; count: number }>>({})
  let upcomingChartInstance: Chart | null = null
  let loadGeneration = 0
  const loadUpcomingCards = async () => {
    const generation = ++loadGeneration
    try {
      const data = await wordTrainingClient.getUpcoming()
      if (generation !== loadGeneration) return

      // Ensure data is in correct format
      if (data && typeof data === 'object') {
        upcomingCardsData.value = data
      } else {
        console.warn('Invalid data format:', data)
        upcomingCardsData.value = {}
      }

      upcomingCardsLoaded.value = true
    } catch (error) {
      console.error('Failed to load upcoming cards:', error)
      if (generation === loadGeneration) upcomingCardsLoaded.value = true // Stop loading after a failed request
    }
  }

  const updateUpcomingChart = () => {

    // Destroy existing chart if it exists
    if (upcomingChartInstance) {
      upcomingChartInstance.destroy()
      upcomingChartInstance = null
    }

    if (!upcomingChartCanvas.value || Object.keys(upcomingCardsData.value).length === 0) return

    // Prepare data - ensure we process dates in order
    const dates = Object.keys(upcomingCardsData.value).sort()
    const labels: string[] = []
    const counts: number[] = []

    dates.forEach(date => {
      const item = upcomingCardsData.value[date]

      if (item && typeof item === 'object' && 'label' in item && 'count' in item) {
        labels.push(item.label)
        counts.push(item.count)

      } else {
        // Fallback if data structure is different
        console.warn('Unexpected data structure for date:', date, item)
        // Try to extract date part for display
        const datePart = date.split('T')[0] || date
        labels.push(datePart)
        counts.push(0)
      }
    })

    // Get theme colors
    const root = getComputedStyle(document.documentElement)
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
    const primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
    const textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
    const borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'

    // Convert hex to rgba
    const hexToRgba = (hex: string, alpha: number) => {
      const r = parseInt(hex.slice(1, 3), 16)
      const g = parseInt(hex.slice(3, 5), 16)
      const b = parseInt(hex.slice(5, 7), 16)
      return `rgba(${r}, ${g}, ${b}, ${alpha})`
    }

    // Create bar chart
    upcomingChartInstance = new Chart(upcomingChartCanvas.value, {
      type: 'bar',
      data: {
        labels: labels,
        datasets: [{
          label: t('common.cards', 2),
          data: counts,
          backgroundColor: hexToRgba(primaryColor, isDark ? 0.7 : 0.6),
          borderColor: primaryColor,
          borderWidth: 1
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
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
                return t('training.chartCardsTooltip', { count: value }, value)
              }
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

  watch([upcomingChartCanvas, upcomingCardsData, locale], updateUpcomingChart, { flush: 'post' })
  onUnmounted(() => {
    loadGeneration++
    upcomingChartInstance?.destroy()
    upcomingChartInstance = null
  })
  return { upcomingCardsLoaded, loadUpcomingCards }
}
