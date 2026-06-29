import { onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { apiClient } from '../api/client'
import { useAuth } from './useAuth'
import { useCourse } from './useCourse'

const QUEUE_KEY = 'lg-activity-queue'
const FLUSH_INTERVAL_MS = 60_000
const IDLE_LIMIT_MS = 60_000

interface ActivityBatch {
  course_code: string
  client_day: string
  seconds: number
  mode: string
}

function routeMode(path: string, navTab?: string): string {
  if (path.includes('sentence')) return 'sentences'
  if (path.startsWith('/chat')) return 'chat'
  if (path.includes('speaking')) return 'speaking'
  if (path.includes('reading')) return 'reading'
  if (path.includes('grammar')) return 'grammar'
  if (path.startsWith('/training') || path.includes('words') || path.includes('word-sets') || path.startsWith('/vocab')) return 'words'
  void navTab
  return 'other'
}

function localDay(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function readQueue(): ActivityBatch[] {
  try { return JSON.parse(localStorage.getItem(QUEUE_KEY) || '[]') } catch { return [] }
}

function writeQueue(q: ActivityBatch[]) {
  try { localStorage.setItem(QUEUE_KEY, JSON.stringify(q.slice(-200))) } catch { /* ignore */ }
}

// Tracks active study time and reports it to /api/linglow/activity.
// Mount once in the public layout.
export function useActivityTracker() {
  const route = useRoute()
  const { isAuthenticated } = useAuth()
  const { currentCourseCode } = useCourse()

  let secondsByMode: Record<string, number> = {}
  // 0 until the user actually interacts: a passive page load must not accrue active time.
  let lastInputAt = 0
  let tickTimer: number | undefined
  let flushTimer: number | undefined

  const markInput = () => { lastInputAt = Date.now() }

  const tick = () => {
    if (!isAuthenticated.value) return
    if (document.visibilityState !== 'visible') return
    if (Date.now() - lastInputAt > IDLE_LIMIT_MS) return
    const mode = routeMode(route.path, route.meta?.navTab as string | undefined)
    secondsByMode[mode] = (secondsByMode[mode] || 0) + 1
  }

  const drainBatches = (): ActivityBatch[] => {
    const course = currentCourseCode.value || ''
    // Never attribute activity to an unknown course: the backend would fall
    // back to the default course (English), so Spanish time leaked there.
    // Hold the accrued seconds until the current course is known.
    if (!course) return []
    const day = localDay()
    const batches: ActivityBatch[] = []
    for (const [mode, seconds] of Object.entries(secondsByMode)) {
      if (seconds > 0) batches.push({ course_code: course, client_day: day, seconds, mode })
    }
    secondsByMode = {}
    return batches
  }

  const sendBatch = async (batch: ActivityBatch): Promise<boolean> => {
    try {
      await apiClient.request('/api/linglow/activity', {
        method: 'POST',
        body: JSON.stringify(batch),
      })
      return true
    } catch {
      return false
    }
  }

  const flush = async () => {
    if (!isAuthenticated.value) return
    const pending = [...readQueue(), ...drainBatches()]
    if (pending.length === 0) return
    writeQueue([])
    const failed: ActivityBatch[] = []
    for (const batch of pending) {
      if (navigator.onLine === false || !(await sendBatch(batch))) failed.push(batch)
    }
    if (failed.length > 0) writeQueue([...readQueue(), ...failed])
  }

  const flushOnHide = () => {
    if (document.visibilityState === 'visible') return
    // Persist immediately; sendBeacon has no auth header, so queue + best-effort flush next session
    const batches = drainBatches()
    if (batches.length > 0) writeQueue([...readQueue(), ...batches])
    void flush()
  }

  onMounted(() => {
    for (const ev of ['pointerdown', 'keydown', 'scroll', 'touchstart']) {
      window.addEventListener(ev, markInput, { passive: true, capture: true })
    }
    document.addEventListener('visibilitychange', flushOnHide)
    window.addEventListener('pagehide', flushOnHide)
    window.addEventListener('online', () => { void flush() })
    tickTimer = window.setInterval(tick, 1000)
    flushTimer = window.setInterval(() => { void flush() }, FLUSH_INTERVAL_MS)
    void flush() // drain queue left over from a previous session
  })

  onUnmounted(() => {
    for (const ev of ['pointerdown', 'keydown', 'scroll', 'touchstart']) {
      window.removeEventListener(ev, markInput, { capture: true } as EventListenerOptions)
    }
    document.removeEventListener('visibilitychange', flushOnHide)
    window.removeEventListener('pagehide', flushOnHide)
    if (tickTimer) window.clearInterval(tickTimer)
    if (flushTimer) window.clearInterval(flushTimer)
  })
}
