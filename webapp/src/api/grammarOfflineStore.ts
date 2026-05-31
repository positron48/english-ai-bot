const DB_NAME = 'qantrix-grammar-offline'
const DB_VERSION = 1
const META_KEY = 'bundle'

export interface OfflineChapterManifest {
  chapter_id: string
  title: string
  title_translations?: Record<string, string>
  title_short?: string
  description?: string
  level?: string
  order: number
  estimated_minutes?: number
  best_score: number
  passed: boolean
  can_access?: boolean
  download_url: string
  approx_bytes?: number
}

export interface OfflineSectionManifest {
  section_id: string
  title: string
  title_translations?: Record<string, string>
  level: string
  order: number
  published_chapters: number
  passed_chapters: number
  total_chapters: number
  progress_percentage: number
  can_access: boolean
  category_test_score?: number
  chapters: OfflineChapterManifest[]
}

export interface OfflineGrammarMeta {
  app_code: string
  bundle_id: string
  native_lang: string
  target_lang: string
  course_version: string
  generated_at: string
  version_hash: string
  approx_bytes: number
  total_chapters: number
  downloaded_from?: string
  downloaded_at: string
  sections: OfflineSectionManifest[]
}

export interface StoredOfflineChapter {
  chapter_id: string
  payload: any
}

export interface QueuedGrammarAttempt {
  client_attempt_id: string
  scope: 'chapter' | 'category'
  scope_id: string
  answers: Array<{ question_id: string; chapter_id?: string; answer: any }>
  course_version?: string
  created_at: string
  result: any
}

type StoreName = 'meta' | 'chapters' | 'queue'

let dbPromise: Promise<IDBDatabase> | null = null

function openDB(): Promise<IDBDatabase> {
  if (dbPromise) return dbPromise
  dbPromise = new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains('meta')) db.createObjectStore('meta')
      if (!db.objectStoreNames.contains('chapters')) db.createObjectStore('chapters', { keyPath: 'chapter_id' })
      if (!db.objectStoreNames.contains('queue')) db.createObjectStore('queue', { keyPath: 'client_attempt_id' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
  return dbPromise
}

async function tx<T>(storeName: StoreName, mode: IDBTransactionMode, fn: (store: IDBObjectStore) => IDBRequest<T> | void): Promise<T | void> {
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

function getAllFromStore<T>(storeName: StoreName): Promise<T[]> {
  return tx<T[]>(storeName, 'readonly', (store) => store.getAll()) as Promise<T[]>
}

export async function getOfflineMeta(): Promise<OfflineGrammarMeta | null> {
  return (await tx<OfflineGrammarMeta>( 'meta', 'readonly', (store) => store.get(META_KEY))) || null
}

export async function setOfflineMeta(meta: OfflineGrammarMeta): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.put(meta, META_KEY))
}

export async function getStoredChapter(chapterID: string): Promise<any | null> {
  const row = await tx<StoredOfflineChapter>('chapters', 'readonly', (store) => store.get(chapterID))
  return row?.payload || null
}

export async function setStoredChapter(chapterID: string, payload: any): Promise<void> {
  await tx('chapters', 'readwrite', (store) => store.put({ chapter_id: chapterID, payload }))
}

export async function countStoredChapters(): Promise<number> {
  return (await tx<number>('chapters', 'readonly', (store) => store.count())) || 0
}

export async function clearOfflineGrammar(): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.clear())
  await tx('chapters', 'readwrite', (store) => store.clear())
  await tx('queue', 'readwrite', (store) => store.clear())
}

export async function enqueueAttempt(attempt: QueuedGrammarAttempt): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.put(attempt))
}

export async function getQueuedAttempts(): Promise<QueuedGrammarAttempt[]> {
  return getAllFromStore<QueuedGrammarAttempt>('queue')
}

export async function deleteQueuedAttempt(clientAttemptID: string): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.delete(clientAttemptID))
}

export async function queueCount(): Promise<number> {
  return (await tx<number>('queue', 'readonly', (store) => store.count())) || 0
}
