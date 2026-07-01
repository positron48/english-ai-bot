import { ref } from 'vue'
import { apiClient } from '../api/client'

export interface MeProfile {
  id: number
  telegram_id: number
  telegram_username: string
  created_at: string
  subscription_tier: string
  features: Record<string, boolean>
}

const me = ref<MeProfile | null>(null)
let loadPromise: Promise<MeProfile | null> | null = null

// /api/me (tier + feature flags) changes rarely, so persist it with a short TTL to avoid
// re-fetching on every page/navigation and every full reload. force=true bypasses the cache.
const ME_CACHE_KEY = 'me_profile_cache_v1'
const ME_CACHE_TTL_MS = 30 * 60 * 1000 // 30 minutes

function readCachedMe(): MeProfile | null {
  try {
    const raw = localStorage.getItem(ME_CACHE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as { at: number; data: MeProfile }
    if (!parsed?.data || typeof parsed.at !== 'number') return null
    if (Date.now() - parsed.at > ME_CACHE_TTL_MS) return null
    return parsed.data
  } catch {
    return null
  }
}

function writeCachedMe(data: MeProfile): void {
  try {
    localStorage.setItem(ME_CACHE_KEY, JSON.stringify({ at: Date.now(), data }))
  } catch {
    /* ignore quota/availability errors */
  }
}

export function clearMeCache(): void {
  me.value = null
  loadPromise = null
  try {
    localStorage.removeItem(ME_CACHE_KEY)
  } catch {
    /* ignore */
  }
}

export function useMe() {
  // ensureMe loads /api/me once and caches it (in-memory + localStorage TTL) for the app session.
  async function ensureMe(force = false): Promise<MeProfile | null> {
    if (force) {
      me.value = null
      loadPromise = null
      try {
        localStorage.removeItem(ME_CACHE_KEY)
      } catch {
        /* ignore */
      }
    }
    if (me.value) return me.value
    if (!force) {
      const cached = readCachedMe()
      if (cached) {
        me.value = cached
        return cached
      }
    }
    if (!loadPromise) {
      loadPromise = apiClient
        .request<MeProfile>('/api/me')
        .then((data) => {
          me.value = data
          writeCachedMe(data)
          return data
        })
        .catch(() => {
          loadPromise = null
          return null
        })
    }
    return loadPromise
  }

  function hasFeature(feature: string): boolean {
    return !!me.value?.features?.[feature]
  }

  return { me, ensureMe, hasFeature }
}
