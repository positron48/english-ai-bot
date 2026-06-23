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

export function useMe() {
  // ensureMe loads /api/me once and caches it for the app session.
  async function ensureMe(force = false): Promise<MeProfile | null> {
    if (force) {
      me.value = null
      loadPromise = null
    }
    if (me.value) return me.value
    if (!loadPromise) {
      loadPromise = apiClient
        .request<MeProfile>('/api/me')
        .then((data) => {
          me.value = data
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
