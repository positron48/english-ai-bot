const DB_NAME = 'qantrix-word-training-offline'
const DB_VERSION = 1
const PACK_KEY = 'pack'
const SESSION_KEY = 'active_session'

export interface OfflineWordTrainingCard {
  question: string
  user_card_id: number
  training_card_id: number
  word_card_id: number
  direction: string
  word_en?: string
  word_target?: string
  word_ru?: string
  word_native?: string
  display_word?: string
  display_target?: string
  transcription?: string
  example_en?: string
  example_target?: string
  hint?: string
  word_category?: string
  morph?: any
  options: string[]
  correct_answer: string
  srs?: Record<string, unknown>
}

export interface OfflineWordTrainingPack {
  app_code: string
  native_lang: string
  target_lang: string
  generated_at: string
  algo_version: string
  total_cards: number
  available_count: number
  downloaded_at: string
  cards: OfflineWordTrainingCard[]
}

export interface OfflineWordTrainingSession {
  id: number
  started_at: string
  index: number
  correct_count: number
  queue: OfflineWordTrainingCard[]
  shown_at?: string
  options_shown_at?: string
}

export interface QueuedWordTrainingAttempt {
  client_attempt_id: string
  user_card_id: number
  training_card_id: number
  direction: string
  mode: 'card'
  shown_at: string
  options_shown_at: string
  answered_at: string
  t_delay_ms: number
  answer_time_ms: number
  early_reveal: boolean
  options: string[]
  chosen_option: string
  correct_answer: string
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

export async function getWordTrainingPack(): Promise<OfflineWordTrainingPack | null> {
  return (await tx<OfflineWordTrainingPack>('meta', 'readonly', (store) => store.get(PACK_KEY))) || null
}

export async function setWordTrainingPack(pack: OfflineWordTrainingPack): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.put(pack, PACK_KEY))
}

export async function removeWordTrainingCards(userCardIDs: number[]): Promise<void> {
  if (userCardIDs.length === 0) return
  const pack = await getWordTrainingPack()
  if (!pack) return
  const ids = new Set(userCardIDs)
  await setWordTrainingPack({ ...pack, cards: pack.cards.filter((card) => !ids.has(card.user_card_id)) })
}

export async function clearWordTrainingPack(): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.delete(PACK_KEY))
  await tx('meta', 'readwrite', (store) => store.delete(SESSION_KEY))
  await tx('queue', 'readwrite', (store) => store.clear())
}

export async function getWordTrainingSession(): Promise<OfflineWordTrainingSession | null> {
  return (await tx<OfflineWordTrainingSession>('meta', 'readonly', (store) => store.get(SESSION_KEY))) || null
}

export async function setWordTrainingSession(session: OfflineWordTrainingSession): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.put(session, SESSION_KEY))
}

export async function clearWordTrainingSession(): Promise<void> {
  await tx('meta', 'readwrite', (store) => store.delete(SESSION_KEY))
}

export async function enqueueWordTrainingAttempt(attempt: QueuedWordTrainingAttempt): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.put(attempt))
}

export async function getQueuedWordTrainingAttempts(): Promise<QueuedWordTrainingAttempt[]> {
  return (await tx<QueuedWordTrainingAttempt[]>('queue', 'readonly', (store) => store.getAll())) || []
}

export async function deleteQueuedWordTrainingAttempt(clientAttemptID: string): Promise<void> {
  await tx('queue', 'readwrite', (store) => store.delete(clientAttemptID))
}

export async function wordTrainingQueueCount(): Promise<number> {
  return (await tx<number>('queue', 'readonly', (store) => store.count())) || 0
}
