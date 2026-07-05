import { ref, onMounted, onUnmounted, type Ref } from 'vue'
import {
  entryNeedsRefresh,
  getCachedScreen,
  resolveUserScopeFromStorage,
  setCachedScreen,
  type AppDataScreenKey,
} from '../api/appDataCache'

export interface CachedOverviewOptions<T> {
  screenKey: AppDataScreenKey
  courseCode: Ref<string>
  locale?: Ref<string>
  fetcher: () => Promise<T>
  applyPayload: (payload: T, meta: { fromCache: boolean }) => void
  offlineFetcher?: () => Promise<T | null>
}

export function useCachedOverviewScreen<T>(options: CachedOverviewOptions<T>) {
  const loadingInitial = ref(true)
  const refreshing = ref(false)
  const hasCache = ref(false)
  const error = ref('')
  const lastFetchedAt = ref<string | null>(null)
  let loadToken = 0

  async function hydrateFromCache(): Promise<boolean> {
    const code = options.courseCode.value
    if (!code) return false
    const locale = options.locale?.value || 'ru'
    const cached = await getCachedScreen<T>(options.screenKey, code, resolveUserScopeFromStorage(), locale)
    if (!cached) return false
    options.applyPayload(cached.payload, { fromCache: true })
    hasCache.value = true
    loadingInitial.value = false
    lastFetchedAt.value = cached.fetchedAt
    return true
  }

  async function refresh(force = false): Promise<void> {
    const code = options.courseCode.value
    if (!code) return
    const token = ++loadToken
    const locale = options.locale?.value || 'ru'
    const cached = await getCachedScreen<T>(options.screenKey, code, resolveUserScopeFromStorage(), locale)
    const shouldFetch = force || entryNeedsRefresh(cached)
    if (!shouldFetch) {
      if (!hasCache.value && cached) {
        options.applyPayload(cached.payload, { fromCache: true })
        hasCache.value = true
        loadingInitial.value = false
        lastFetchedAt.value = cached.fetchedAt
      }
      return
    }

    const isOffline = typeof navigator !== 'undefined' && navigator.onLine === false
    if (isOffline) {
      if (options.offlineFetcher) {
        try {
          const offlinePayload = await options.offlineFetcher()
          if (offlinePayload && token === loadToken) {
            options.applyPayload(offlinePayload, { fromCache: true })
            hasCache.value = true
          }
        } catch { /* ignore */ }
      }
      loadingInitial.value = false
      return
    }

    if (hasCache.value || cached) refreshing.value = true
    else loadingInitial.value = true
    error.value = ''

    try {
      const payload = await options.fetcher()
      if (token !== loadToken) return
      options.applyPayload(payload, { fromCache: false })
      const saved = await setCachedScreen(options.screenKey, code, payload, {
        locale,
        preserveDirtyTags: false,
      })
      hasCache.value = true
      lastFetchedAt.value = saved.fetchedAt
    } catch (err: any) {
      if (token !== loadToken) return
      error.value = err?.message || 'network error'
      if (!hasCache.value && cached) {
        options.applyPayload(cached.payload, { fromCache: true })
        hasCache.value = true
      }
    } finally {
      if (token === loadToken) {
        loadingInitial.value = false
        refreshing.value = false
      }
    }
  }

  async function load(force = false): Promise<void> {
    if (!force) await hydrateFromCache()
    await refresh(force)
  }

  function onRefreshEvent(event: Event) {
    const detail = (event as CustomEvent<{ screens: AppDataScreenKey[]; courseCode: string }>).detail
    if (!detail) return
    if (detail.courseCode !== options.courseCode.value) return
    if (!detail.screens.includes(options.screenKey)) return
    void refresh(true)
  }

  onMounted(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('linglow-app-data-refresh', onRefreshEvent)
    }
  })
  onUnmounted(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('linglow-app-data-refresh', onRefreshEvent)
    }
  })

  return {
    loadingInitial,
    refreshing,
    hasCache,
    error,
    lastFetchedAt,
    load,
    refresh,
    hydrateFromCache,
  }
}
