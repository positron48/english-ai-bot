export type AppDataScreenKey =
  | 'dashboard'
  | 'city'
  | 'city-district'
  | 'learning'
  | 'progress'
  | 'daily-route'
  | 'review'

export type AppDataDirtyTag =
  | 'srs'
  | 'words'
  | 'grammar'
  | 'reading'
  | 'speaking'
  | 'conversation'
  | 'picture'
  | 'stats'
  | 'progress'
  | 'city'
  | 'daily-route'
  | 'course'

export const APP_DATA_CACHE_SCHEMA_VERSION = 1
export const APP_DATA_CACHE_DB_NAME = 'linglow-app-data-cache'
export const APP_DATA_CACHE_DB_VERSION = 1

export const SCREEN_TTL_MS: Record<AppDataScreenKey, number> = {
  dashboard: 5 * 60_000,
  learning: 5 * 60_000,
  city: 15 * 60_000,
  'city-district': 15 * 60_000,
  progress: 10 * 60_000,
  'daily-route': 2 * 60_000,
  review: 2 * 60_000,
}

export const SCREEN_DIRTY_TAGS: Record<AppDataScreenKey, AppDataDirtyTag[]> = {
  dashboard: ['words', 'srs', 'stats', 'progress', 'grammar', 'reading', 'daily-route'],
  city: ['city', 'grammar', 'reading', 'conversation', 'picture', 'progress', 'words'],
  'city-district': ['city', 'grammar', 'reading', 'conversation', 'picture', 'progress', 'words'],
  learning: ['grammar', 'words', 'reading', 'speaking'],
  progress: ['stats', 'progress', 'grammar', 'reading', 'words'],
  'daily-route': ['srs', 'daily-route', 'reading', 'grammar', 'words'],
  review: ['srs'],
}

export const EVENT_DIRTY_TAGS: Record<string, AppDataDirtyTag[]> = {
  'word-review-recorded': ['words', 'srs', 'stats', 'progress', 'daily-route'],
  'word-review-session-completed': ['words', 'srs', 'stats', 'progress', 'daily-route'],
  'grammar-test-submitted': ['grammar', 'stats', 'progress', 'city'],
  'grammar-training-recorded': ['grammar', 'srs', 'stats', 'progress', 'daily-route'],
  'reading-marked-read': ['reading', 'stats', 'progress', 'city', 'daily-route'],
  'speaking-attempt-submitted': ['speaking', 'stats', 'progress', 'city', 'daily-route'],
  'conversation-progressed': ['conversation', 'stats', 'progress', 'city', 'daily-route'],
  'picture-quest-progressed': ['picture', 'stats', 'progress', 'city', 'daily-route'],
  'word-set-updated': ['words', 'stats', 'progress', 'city', 'daily-route', 'srs'],
  'activity-recorded': ['stats'],
  'offline-sync-completed': ['words', 'srs', 'grammar', 'stats', 'progress', 'city', 'daily-route'],
  'course-selected': ['course'],
  'manual-refresh': [],
}

export interface CachedScreenPayload<T = unknown> {
  key: AppDataScreenKey
  userScope: string
  courseCode: string
  locale: string
  appVersion: string
  dataVersion: number
  payload: T
  fetchedAt: string
  staleAt: string
  dirtyTags: AppDataDirtyTag[]
  pendingLocalMutations: number
}

export function buildScreenStorageKey(
  userScope: string,
  courseCode: string,
  screenKey: AppDataScreenKey,
  locale = 'ru',
): string {
  return `${userScope}|${courseCode}|${locale}|${screenKey}`
}

export function computeStaleAt(fetchedAt: string, screenKey: AppDataScreenKey, now = Date.now()): string {
  const base = Date.parse(fetchedAt)
  const at = Number.isFinite(base) ? base : now
  return new Date(at + SCREEN_TTL_MS[screenKey]).toISOString()
}

export function isEntryStale(entry: Pick<CachedScreenPayload, 'staleAt'>, now = Date.now()): boolean {
  const staleAt = Date.parse(entry.staleAt)
  return !Number.isFinite(staleAt) || staleAt <= now
}

export function entryNeedsRefresh(entry: CachedScreenPayload | null, now = Date.now()): boolean {
  if (!entry) return true
  if (entry.dirtyTags.length > 0) return true
  return isEntryStale(entry, now)
}

export function tagsAffectScreen(screenKey: AppDataScreenKey, tags: Iterable<AppDataDirtyTag>): boolean {
  const affinity = SCREEN_DIRTY_TAGS[screenKey]
  for (const tag of tags) {
    if (affinity.includes(tag)) return true
  }
  return false
}

export function screensForTags(tags: Iterable<AppDataDirtyTag>): AppDataScreenKey[] {
  const tagList = [...tags]
  if (tagList.includes('course')) {
    return Object.keys(SCREEN_DIRTY_TAGS) as AppDataScreenKey[]
  }
  return (Object.keys(SCREEN_DIRTY_TAGS) as AppDataScreenKey[]).filter((screen) => tagsAffectScreen(screen, tagList))
}

export function hashTokenScope(token: string): string {
  let hash = 2166136261
  for (let i = 0; i < token.length; i++) {
    hash ^= token.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return `tok:${(hash >>> 0).toString(16)}`
}
