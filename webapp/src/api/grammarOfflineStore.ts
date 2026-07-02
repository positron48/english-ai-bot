const DB_NAME = 'qantrix-grammar-offline'
const DB_VERSION = 2
const META_KEY = 'bundle'
const DEFAULT_SCOPE = 'default'

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
  opened_by_placement?: boolean
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
  source_chapter_id?: string
  course_code?: string
  payload: any
}

export interface QueuedGrammarAttempt {
  client_attempt_id: string
  course_code?: string
  scope: 'chapter' | 'category'
  scope_id: string
  answers: Array<{ question_id: string; chapter_id?: string; answer: any }>
  course_version?: string
  created_at: string
  result: any
}

export interface QueuedGrammarTrainingAttempt {
  client_attempt_id: string
  course_code?: string
  question_id: string
  answer: any
  created_at: string
  result: any
}

type StoreName = 'meta' | 'chapters' | 'queue' | 'training' | 'training_queue'

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
      if (!db.objectStoreNames.contains('training')) db.createObjectStore('training')
      if (!db.objectStoreNames.contains('training_queue')) db.createObjectStore('training_queue', { keyPath: 'client_attempt_id' })
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

function scopeKey(courseCode?: string): string {
  const code = (courseCode || '').trim().toLowerCase()
  return code || DEFAULT_SCOPE
}

function metaKey(courseCode?: string): string {
  const scope = scopeKey(courseCode)
  return scope === DEFAULT_SCOPE ? META_KEY : `${META_KEY}:${scope}`
}

function chapterKey(courseCode: string | undefined, chapterID: string): string {
  return `${scopeKey(courseCode)}:${chapterID}`
}

function trainingKey(courseCode?: string): string {
  return `questions:${scopeKey(courseCode)}`
}

function matchesScope(rowCourseCode: string | undefined, courseCode?: string): boolean {
  return scopeKey(rowCourseCode) === scopeKey(courseCode)
}

export async function getOfflineMeta(courseCode?: string): Promise<OfflineGrammarMeta | null> {
  return (await tx<OfflineGrammarMeta>( 'meta', 'readonly', (store) => store.get(metaKey(courseCode)))) || null
}

export async function setOfflineMeta(meta: OfflineGrammarMeta, courseCode?: string): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.put(meta, metaKey(courseCode)))
}

export async function getStoredChapter(chapterID: string, courseCode?: string): Promise<any | null> {
  const row = await tx<StoredOfflineChapter>('chapters', 'readonly', (store) => store.get(chapterKey(courseCode, chapterID)))
  return row?.payload || null
}

export async function getStoredChapters(courseCode?: string): Promise<any[]> {
  const rows = await getAllFromStore<StoredOfflineChapter>('chapters')
  return rows.filter((row) => matchesScope(row.course_code, courseCode)).map((row) => row.payload).filter(Boolean)
}

export async function setStoredChapter(chapterID: string, payload: any, courseCode?: string): Promise<void> {
  await tx('chapters', 'readwrite', (store) => store.put({
    chapter_id: chapterKey(courseCode, chapterID),
    source_chapter_id: chapterID,
    course_code: scopeKey(courseCode),
    payload,
  }))
}

export async function countStoredChapters(courseCode?: string): Promise<number> {
  return (await getStoredChapters(courseCode)).length
}

export async function clearOfflineGrammar(courseCode?: string): Promise<void> {
  const scope = scopeKey(courseCode)
  if (scope === DEFAULT_SCOPE) {
    await tx('meta', 'readwrite', (store) => store.clear())
    await tx('chapters', 'readwrite', (store) => store.clear())
    await tx('queue', 'readwrite', (store) => store.clear())
    await tx('training', 'readwrite', (store) => store.clear())
    await tx('training_queue', 'readwrite', (store) => store.clear())
    return
  }
  await tx('meta', 'readwrite', (store) => store.delete(metaKey(courseCode)))
  const chapters = await getAllFromStore<StoredOfflineChapter>('chapters')
  await tx('chapters', 'readwrite', (store) => {
    for (const row of chapters) if (matchesScope(row.course_code, courseCode)) store.delete(row.chapter_id)
  })
  const attempts = await getAllFromStore<QueuedGrammarAttempt>('queue')
  await tx('queue', 'readwrite', (store) => {
    for (const row of attempts) if (matchesScope(row.course_code, courseCode)) store.delete(row.client_attempt_id)
  })
  await tx('training', 'readwrite', (store) => store.delete(trainingKey(courseCode)))
  const trainingAttempts = await getAllFromStore<QueuedGrammarTrainingAttempt>('training_queue')
  await tx('training_queue', 'readwrite', (store) => {
    for (const row of trainingAttempts) if (matchesScope(row.course_code, courseCode)) store.delete(row.client_attempt_id)
  })
}

export async function enqueueAttempt(attempt: QueuedGrammarAttempt): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.put(attempt))
}

export async function getQueuedAttempts(courseCode?: string): Promise<QueuedGrammarAttempt[]> {
  return (await getAllFromStore<QueuedGrammarAttempt>('queue')).filter((row) => matchesScope(row.course_code, courseCode))
}

export async function deleteQueuedAttempt(clientAttemptID: string): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.delete(clientAttemptID))
}

export async function queueCount(courseCode?: string): Promise<number> {
  return (await getQueuedAttempts(courseCode)).length
}

export async function setTrainingQuestions(questions: any[], courseCode?: string): Promise<void> {
  await tx('training', 'readwrite', (store) => store.put(questions, trainingKey(courseCode)))
}

export async function getTrainingQuestions(courseCode?: string): Promise<any[]> {
  return (await tx<any[]>('training', 'readonly', (store) => store.get(trainingKey(courseCode)))) || []
}

export async function enqueueTrainingAttempt(attempt: QueuedGrammarTrainingAttempt): Promise<void> {
  await tx('training_queue', 'readwrite', (store) => store.put(attempt))
}

export async function getQueuedTrainingAttempts(courseCode?: string): Promise<QueuedGrammarTrainingAttempt[]> {
  return (await getAllFromStore<QueuedGrammarTrainingAttempt>('training_queue')).filter((row) => matchesScope(row.course_code, courseCode))
}

export async function deleteQueuedTrainingAttempt(clientAttemptID: string): Promise<void> {
  await tx('training_queue', 'readwrite', (store) => store.delete(clientAttemptID))
}

export async function trainingQueueCount(courseCode?: string): Promise<number> {
  return (await getQueuedTrainingAttempts(courseCode)).length
}
