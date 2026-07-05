const DB_NAME = 'qantrix-content-reports-offline'
const DB_VERSION = 1

export type ContentReportSourceType =
  | 'word_training'
  | 'grammar_training'
  | 'grammar_chapter'
  | 'grammar_test'
  | 'reading_text'

export interface QueuedContentReport {
  client_report_id: string
  source_type: ContentReportSourceType
  report_category: string
  comment: string
  created_at: string
  word?: string
  direction?: string
  word_card_id?: number
  training_card_id?: number
  user_card_id?: number
  word_category?: string
  grammar_chapter_id?: string
  theory_block_id?: string
  grammar_question_id?: string
  reading_text_id?: string
  reading_category_id?: string
  payload: Record<string, unknown>
}

type StoreName = 'meta' | 'queue'

let dbPromise: Promise<IDBDatabase> | null = null

function openDB(): Promise<IDBDatabase> {
  if (dbPromise) return dbPromise
  dbPromise = new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains('meta')) db.createObjectStore('meta')
      if (!db.objectStoreNames.contains('queue')) db.createObjectStore('queue', { keyPath: 'client_report_id' })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
  return dbPromise
}

async function withStore<T>(storeName: StoreName, mode: IDBTransactionMode, fn: (store: IDBObjectStore) => IDBRequest<T> | void): Promise<T | void> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, mode)
    const store = tx.objectStore(storeName)
    const request = fn(store)
    if (!request) {
      tx.oncomplete = () => resolve(undefined)
      tx.onerror = () => reject(tx.error)
      return
    }
    request.onsuccess = () => resolve(request.result as T)
    request.onerror = () => reject(request.error)
  })
}

export async function enqueueContentReport(report: QueuedContentReport): Promise<void> {
  await withStore('queue', 'readwrite', (store) => store.put(report))
}

export async function getQueuedContentReports(): Promise<QueuedContentReport[]> {
  const db = await openDB()
  return new Promise((resolve, reject) => {
    const tx = db.transaction('queue', 'readonly')
    const store = tx.objectStore('queue')
    const request = store.getAll()
    request.onsuccess = () => resolve((request.result || []) as QueuedContentReport[])
    request.onerror = () => reject(request.error)
  })
}

export async function deleteQueuedContentReport(clientReportID: string): Promise<void> {
  await withStore('queue', 'readwrite', (store) => store.delete(clientReportID))
}

export async function contentReportQueueCount(): Promise<number> {
  const items = await getQueuedContentReports()
  return items.length
}
