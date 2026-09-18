import { getActiveCourseCodeForInvalidation } from './cacheInvalidation'
import { captureUserScope } from './sessionScope'
import { createOfflineStore } from './offlineStore'
const packKey = (course = getActiveCourseCodeForInvalidation()) => `pack:${course || 'default'}`
const sessionKey = (course = getActiveCourseCodeForInvalidation()) => `session:${course || 'default'}`

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

export interface OfflineWordTrainingQueueItem extends Partial<OfflineWordTrainingCard> {
  type: 'card' | 'spell' | 'type'
  user_card_id: number
  direction: string
  correct_answer: string
  prefix?: string
  letters?: string[]
  hint_first_letter?: string
  hint_length?: number
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
  queue?: OfflineWordTrainingQueueItem[]
}

export interface OfflineWordTrainingSession {
  id: number
  started_at: string
  index: number
  correct_count: number
  queue: OfflineWordTrainingQueueItem[]
  shown_at?: string
  options_shown_at?: string
}

export interface QueuedWordTrainingAttempt {
  client_attempt_id: string
  user_card_id: number
  training_card_id: number
  direction: string
  mode: 'card' | 'spell' | 'type'
  shown_at: string
  options_shown_at: string
  answered_at: string
  t_delay_ms: number
  answer_time_ms: number
  early_reveal: boolean
  options: string[]
  chosen_option?: string
  answer_text?: string
  correct_answer: string
}

type StoreName = 'meta' | 'queue'

const tx = createOfflineStore<StoreName>('wordTrainingOfflineStore', { meta: undefined, queue: { keyPath: 'client_attempt_id' } })

export async function getWordTrainingPack(): Promise<OfflineWordTrainingPack | null> {
  const key = packKey()
  return (await tx<OfflineWordTrainingPack>('meta', 'readonly', (store) => store.get(key))) || null
}

export async function setWordTrainingPack(pack: OfflineWordTrainingPack): Promise<void> {
  const key = packKey()
  await tx('meta', 'readwrite', (store) => store.put(pack, key))
}

export function packQueueItems(pack: OfflineWordTrainingPack | null): OfflineWordTrainingQueueItem[] {
  if (!pack) return []
  if (pack.queue?.length) return pack.queue
  return (pack.cards || []).map((card) => ({ ...card, type: 'card' as const }))
}

export async function removeWordTrainingUserCards(userCardIDs: number[]): Promise<void> {
  if (userCardIDs.length === 0) return
  const checkUser = captureUserScope()
  const course = getActiveCourseCodeForInvalidation()
  const pack = await getWordTrainingPack()
  checkUser()
  if (getActiveCourseCodeForInvalidation() !== course) throw new Error('Course changed')
  if (!pack) return
  const ids = new Set(userCardIDs)
  await setWordTrainingPack({
    ...pack,
    cards: (pack.cards || []).filter((card) => !ids.has(card.user_card_id)),
    queue: (pack.queue || []).filter((item) => !ids.has(item.user_card_id)),
  })
}

/** @deprecated use removeWordTrainingUserCards */
export async function removeWordTrainingCards(userCardIDs: number[]): Promise<void> {
  return removeWordTrainingUserCards(userCardIDs)
}

export async function clearWordTrainingPack(): Promise<void> {
  const pack = packKey()
  const session = sessionKey()
  await tx('meta', 'readwrite', (store) => { store.delete(pack); store.delete(session) })
}

export async function getWordTrainingSession(): Promise<OfflineWordTrainingSession | null> {
  const key = sessionKey()
  return (await tx<OfflineWordTrainingSession>('meta', 'readonly', (store) => store.get(key))) || null
}

export async function setWordTrainingSession(session: OfflineWordTrainingSession): Promise<void> {
  const key = sessionKey()
  await tx('meta', 'readwrite', (store) => store.put(session, key))
}

export async function clearWordTrainingSession(): Promise<void> {
  const key = sessionKey()
  await tx('meta', 'readwrite', (store) => store.delete(key))
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
