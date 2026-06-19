import { computed, ref } from 'vue'
import { statsClient, type LinglowStats } from '../api/statsClient'
import { useCourse } from './useCourse'

const stats = ref<LinglowStats | null>(null)
let loadedForCourse = ''
let inflight: Promise<void> | null = null

// Session-wide stats cache (streak badge, today counters). refreshStats() after
// finishing a training session to bump the badge without a reload.
export function useStats() {
  const { currentCourseCode } = useCourse()

  const load = async (force = false) => {
    const code = currentCourseCode.value || ''
    if (!force && stats.value && loadedForCourse === code) return
    if (inflight) return inflight
    inflight = statsClient.getStats({ courseCode: code || undefined })
      .then((s) => {
        stats.value = s
        loadedForCourse = code
      })
      .catch(() => { /* keep previous */ })
      .finally(() => { inflight = null })
    return inflight
  }

  const refreshStats = () => load(true)

  const streakDays = computed(() => stats.value?.streak.current_days ?? 0)

  return { stats, streakDays, ensureStatsLoaded: load, refreshStats }
}
