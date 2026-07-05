import {
  APP_DATA_CACHE_DB_NAME,
  APP_DATA_CACHE_DB_VERSION,
  APP_DATA_CACHE_SCHEMA_VERSION,
  type AppDataDirtyTag,
  type AppDataScreenKey,
  type CachedScreenPayload,
  buildScreenStorageKey,
  computeStaleAt,
  entryNeedsRefresh,
  hashTokenScope,
  screensForTags,
  SCREEN_DIRTY_TAGS,
} from './appDataCacheLogic'
import { getInitialLocale } from '../i18n'

export * from './appDataCacheLogic'

export function getAppDataLocale(): string {
  return getInitialLocale()
}

const META_SCHEMA_KEY = 'schema'
const APP_VERSION = 'linglow-web'

type StoreName = 'screens' | 'meta'

let dbPromise: Promise<IDBDatabase> | null = null

function openDB(): Promise<IDBDatabase> {
  if (dbPromise) return dbPromise
  if (typeof indexedDB === 'undefined') {
    return Promise.reject(new Error('indexedDB unavailable'))
  }
  dbPromise = new Promise((resolve, reject) => {
    const request = indexedDB.open(APP_DATA_CACHE_DB_NAME, APP_DATA_CACHE_DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains('screens')) {
        db.createObjectStore('screens')
      }
      if (!db.objectStoreNames.contains('meta')) {
        db.createObjectStore('meta')
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
  return dbPromise
}

async function tx<T>(
  storeName: StoreName,
  mode: IDBTransactionMode,
  fn: (store: IDBObjectStore) => IDBRequest<T> | void,
): Promise<T | void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const transaction = db.transaction(storeName, mode)
    const store = transaction.objectStore(storeName)
    const request = fn(store)
    let result: T | void
    if (request) {
      request.onsuccess = () => { result = request.result }
      request.onerror = () => reject(request.error)
    }
    transaction.oncomplete = () => resolve(result)
    transaction.onerror = () => reject(transaction.error)
  })
}

export function getAppDataVersion(): string {
  return APP_VERSION
}

export function resolveUserScopeFromStorage(): string {
  if (typeof localStorage === 'undefined') return 'anon'
  try {
    const raw = localStorage.getItem('me_profile_cache_v1')
    if (raw) {
      const parsed = JSON.parse(raw) as { data?: { id?: number } }
      if (parsed?.data?.id) return `user:${parsed.data.id}`
    }
  } catch { /* ignore */ }
  const token = localStorage.getItem('access_token')
  if (token) return hashTokenScope(token)
  return 'anon'
}

export async function getCachedScreen<T>(
  screenKey: AppDataScreenKey,
  courseCode: string,
  userScope = resolveUserScopeFromStorage(),
  locale = getAppDataLocale(),
): Promise<CachedScreenPayload<T> | null> {
  if (!courseCode) return null
  try {
    const key = buildScreenStorageKey(userScope, courseCode, screenKey, locale)
    const entry = await tx<CachedScreenPayload<T>>('screens', 'readonly', (store) => store.get(key))
    if (!entry || entry.key !== screenKey) return null
    if (entry.userScope !== userScope || entry.courseCode !== courseCode) return null
    return entry
  } catch {
    return null
  }
}

export async function setCachedScreen<T>(
  screenKey: AppDataScreenKey,
  courseCode: string,
  payload: T,
  options: {
    userScope?: string
    locale?: string
    dirtyTags?: AppDataDirtyTag[]
    pendingLocalMutations?: number
    preserveDirtyTags?: boolean
  } = {},
): Promise<CachedScreenPayload<T>> {
  const userScope = options.userScope ?? resolveUserScopeFromStorage()
  const locale = options.locale ?? getAppDataLocale()
  const fetchedAt = new Date().toISOString()
  const existing = await getCachedScreen<T>(screenKey, courseCode, userScope, locale)
  const entry: CachedScreenPayload<T> = {
    key: screenKey,
    userScope,
    courseCode,
    locale,
    appVersion: getAppDataVersion(),
    dataVersion: APP_DATA_CACHE_SCHEMA_VERSION,
    payload,
    fetchedAt,
    staleAt: computeStaleAt(fetchedAt, screenKey),
    dirtyTags: options.preserveDirtyTags ? (existing?.dirtyTags || []) : (options.dirtyTags || []),
    pendingLocalMutations: options.pendingLocalMutations ?? existing?.pendingLocalMutations ?? 0,
  }
  const key = buildScreenStorageKey(userScope, courseCode, screenKey, locale)
  await tx('screens', 'readwrite', (store) => store.put(entry, key))
  await tx('meta', 'readwrite', (store) => store.put({
    schemaVersion: APP_DATA_CACHE_SCHEMA_VERSION,
    appVersion: getAppDataVersion(),
    lastRefreshAt: fetchedAt,
    userScope,
    courseCode,
  }, META_SCHEMA_KEY))
  return entry
}

export async function patchCachedScreen<T>(
  screenKey: AppDataScreenKey,
  courseCode: string,
  patcher: (payload: T) => T,
  options: { userScope?: string; locale?: string; addDirtyTags?: AppDataDirtyTag[] } = {},
): Promise<CachedScreenPayload<T> | null> {
  const userScope = options.userScope ?? resolveUserScopeFromStorage()
  const locale = options.locale ?? getAppDataLocale()
  const existing = await getCachedScreen<T>(screenKey, courseCode, userScope, locale)
  if (!existing) return null
  const dirtyTags = new Set(existing.dirtyTags)
  for (const tag of options.addDirtyTags || []) dirtyTags.add(tag)
  return setCachedScreen(screenKey, courseCode, patcher(existing.payload), {
    userScope,
    locale,
    dirtyTags: [...dirtyTags],
    pendingLocalMutations: (existing.pendingLocalMutations || 0) + 1,
  })
}

export async function markScreensDirty(
  courseCode: string,
  tags: AppDataDirtyTag[],
  userScope = resolveUserScopeFromStorage(),
  locale = getAppDataLocale(),
): Promise<void> {
  if (!courseCode || tags.length === 0) return
  const screens = screensForTags(tags)
  await Promise.all(screens.map(async (screenKey) => {
    const existing = await getCachedScreen(screenKey, courseCode, userScope, locale)
    if (!existing) return
    const dirtyTags = new Set([...existing.dirtyTags, ...tags.filter((tag) => SCREEN_DIRTY_TAGS[screenKey].includes(tag))])
    const key = buildScreenStorageKey(userScope, courseCode, screenKey, locale)
    const next = { ...existing, dirtyTags: [...dirtyTags] }
    await tx('screens', 'readwrite', (store) => store.put(next, key))
  }))
}

export async function hasCachedScreen(
  screenKey: AppDataScreenKey,
  courseCode: string,
  userScope = resolveUserScopeFromStorage(),
  locale = getAppDataLocale(),
): Promise<boolean> {
  const entry = await getCachedScreen(screenKey, courseCode, userScope, locale)
  return entry != null
}

export async function clearAppDataCacheForUser(userScope = resolveUserScopeFromStorage()): Promise<void> {
  try {
    const db = await openDB()
    const keys = await new Promise<string[]>((resolve, reject) => {
      const transaction = db.transaction('screens', 'readonly')
      const store = transaction.objectStore('screens')
      const request = store.getAllKeys()
      request.onsuccess = () => resolve((request.result as string[]) || [])
      request.onerror = () => reject(request.error)
    })
    const prefix = `${userScope}|`
    await Promise.all(keys.filter((key) => key.startsWith(prefix)).map((key) =>
      tx('screens', 'readwrite', (store) => store.delete(key)),
    ))
  } catch { /* ignore */ }
}

export async function listCachedScreensDebug(
  courseCode?: string,
  userScope = resolveUserScopeFromStorage(),
): Promise<Array<Pick<CachedScreenPayload, 'key' | 'fetchedAt' | 'staleAt' | 'dirtyTags' | 'pendingLocalMutations'>>> {
  try {
    const db = await openDB()
    const entries = await new Promise<CachedScreenPayload[]>((resolve, reject) => {
      const transaction = db.transaction('screens', 'readonly')
      const store = transaction.objectStore('screens')
      const request = store.getAll()
      request.onsuccess = () => resolve((request.result as CachedScreenPayload[]) || [])
      request.onerror = () => reject(request.error)
    })
    return entries
      .filter((entry) => entry.userScope === userScope && (!courseCode || entry.courseCode === courseCode))
      .map((entry) => ({
        key: entry.key,
        fetchedAt: entry.fetchedAt,
        staleAt: entry.staleAt,
        dirtyTags: entry.dirtyTags,
        pendingLocalMutations: entry.pendingLocalMutations,
      }))
  } catch {
    return []
  }
}

export { entryNeedsRefresh }
