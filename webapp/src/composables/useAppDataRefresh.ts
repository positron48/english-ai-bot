import { apiClient } from '../api/client'
import { scheduleOfflineSync } from '../api/offlineSyncRunner'
import {
  getCachedScreen,
  resolveUserScopeFromStorage,
  setCachedScreen,
  type AppDataScreenKey,
} from '../api/appDataCache'
import { emitAppDataEvent, registerAppDataRefreshHandler } from '../api/cacheInvalidation'
import { courseClient } from '../api/courseClient'

type OverviewScreenKey = Exclude<AppDataScreenKey, 'daily-route' | 'review' | 'city-district'>

const OVERVIEW_ENDPOINTS: Record<OverviewScreenKey, string> = {
  dashboard: '/api/overview/dashboard',
  city: '/api/overview/city',
  learning: '/api/overview/learning',
  progress: '/api/overview/progress',
}

const inflight = new Map<string, Promise<void>>()

function notifyScreensRefreshed(screens: AppDataScreenKey[], courseCode: string): void {
  if (typeof window === 'undefined') return
  window.dispatchEvent(new CustomEvent('linglow-app-data-refresh', {
    detail: { screens, courseCode },
  }))
}

function courseQuery(courseCode: string): string {
  return courseCode ? `?course_code=${encodeURIComponent(courseCode)}` : ''
}

async function fetchOverviewScreen(screenKey: OverviewScreenKey, courseCode: string): Promise<unknown> {
  const endpoint = `${OVERVIEW_ENDPOINTS[screenKey]}${courseQuery(courseCode)}`
  return apiClient.request(endpoint)
}

async function fetchCityDistrictBundle(courseCode: string): Promise<unknown> {
  const [map, prog] = await Promise.all([
    courseClient.getCourseMap(courseCode),
    courseClient.getProgress(courseCode),
  ])
  return { course_map: map, progress: prog }
}

async function fetchDailyRouteBundle(courseCode: string): Promise<unknown> {
  const [route, review, progress] = await Promise.all([
    courseClient.getDailyRoute(16, courseCode),
    courseClient.getReviewQueue(16, courseCode),
    courseClient.getProgress(courseCode),
  ])
  return { route, review, progress }
}

export async function refreshAppData(options: {
  courseCode: string
  locale?: string
  reason?: string
  screens?: AppDataScreenKey[]
  scheduleSync?: boolean
}): Promise<void> {
  const {
    courseCode,
    locale = 'ru',
    reason = 'manual-refresh',
    screens = ['dashboard', 'learning', 'city', 'progress', 'daily-route', 'review'],
    scheduleSync = reason === 'manual-refresh',
  } = options
  if (!courseCode) return
  if (scheduleSync && typeof navigator !== 'undefined' && navigator.onLine !== false) {
    scheduleOfflineSync()
  }
  const key = `${courseCode}:${screens.join(',')}:${reason}`
  const existing = inflight.get(key)
  if (existing) return existing

  const job = (async () => {
    const userScope = resolveUserScopeFromStorage()
    await Promise.all(screens.map(async (screenKey) => {
      try {
        if (screenKey === 'daily-route' || screenKey === 'review') {
          const bundle = await fetchDailyRouteBundle(courseCode)
          await setCachedScreen('daily-route', courseCode, bundle, { userScope, locale })
          if (screenKey === 'review') {
            await setCachedScreen('review', courseCode, (bundle as any).review, { userScope, locale })
          }
          return
        }
        if (screenKey === 'city-district') {
          const bundle = await fetchCityDistrictBundle(courseCode)
          await setCachedScreen('city-district', courseCode, bundle, { userScope, locale })
          return
        }
        const payload = await fetchOverviewScreen(screenKey, courseCode)
        await setCachedScreen(screenKey, courseCode, payload, { userScope, locale })
      } catch (error) {
        console.warn(`[app-data] refresh failed for ${screenKey}:`, error)
      }
    }))
    if (reason === 'manual-refresh') {
      emitAppDataEvent('manual-refresh', courseCode)
    }
    notifyScreensRefreshed(screens, courseCode)
  })().finally(() => inflight.delete(key))

  inflight.set(key, job)
  return job
}

export async function prefetchAppData(courseCode: string, locale = 'ru'): Promise<void> {
  if (!courseCode) return
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  const screens: AppDataScreenKey[] = ['dashboard', 'learning', 'city', 'progress', 'daily-route', 'review']
  const missing: AppDataScreenKey[] = []
  const userScope = resolveUserScopeFromStorage()
  for (const screen of screens) {
    const cached = await getCachedScreen(screen, courseCode, userScope, locale)
    if (!cached) missing.push(screen)
  }
  if (missing.length === 0) return
  await refreshAppData({ courseCode, locale, screens: missing, reason: 'prefetch', scheduleSync: false })
}

export function registerBackgroundRefresh(
  onScreens: (screens: AppDataScreenKey[], courseCode: string) => void,
): void {
  registerAppDataRefreshHandler((screens, courseCode) => onScreens(screens, courseCode))
}
