import { apiClient } from './client'
import { getGrammarCourseCode } from './grammarClient'
import {
  OfflineWordTrainingPack,
  OfflineWordTrainingQueueItem,
  OfflineWordTrainingSession,
  QueuedWordTrainingAttempt,
  clearWordTrainingPack,
  clearWordTrainingSession,
  deleteQueuedWordTrainingAttempt,
  enqueueWordTrainingAttempt,
  getQueuedWordTrainingAttempts,
  getWordTrainingPack,
  getWordTrainingSession,
  packQueueItems,
  removeWordTrainingUserCards,
  setWordTrainingPack,
  setWordTrainingSession,
  wordTrainingQueueCount,
} from './wordTrainingOfflineStore'

export class OfflineWordTrainingUnavailableError extends Error {
  constructor(message = 'Word training is not preloaded for offline use') {
    super(message)
    this.name = 'OfflineWordTrainingUnavailableError'
  }
}

export interface WordTrainingOfflineStatus {
  ready: boolean
  downloadedCards: number
  downloadedAt?: string
  pendingAttempts: number
  legacyPack?: boolean
}

/** Append the active course code so legacy training endpoints scope to the UI-selected
 *  course (the UI is the source of truth, avoiding the stale DB-persisted course race). */
const withCourse = (url: string): string => {
  const code = getGrammarCourseCode()
  if (!code) return url
  const sep = url.includes('?') ? '&' : '?'
  return `${url}${sep}course_code=${encodeURIComponent(code)}`
}

const isBrowserOffline = () => typeof navigator !== 'undefined' && navigator.onLine === false
const isNetworkError = (error: any) => error?.isNetworkError || error?.name === 'TypeError' || String(error?.message || '').includes('Failed to fetch')
const offlineLabel = () => {
  if (typeof localStorage === 'undefined') return 'Offline'
  const locale = localStorage.getItem('locale')
  if (locale === 'ru') return 'Офлайн'
  if (locale === 'es') return 'Offline'
  return 'Offline'
}

function shuffle<T>(items: T[]): T[] {
  const arr = [...items]
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[arr[i], arr[j]] = [arr[j], arr[i]]
  }
  return arr
}

async function offlineFallback<T>(online: () => Promise<T>, offline: () => Promise<T>): Promise<T> {
  if (isBrowserOffline()) return offline()
  try {
    return await online()
  } catch (error) {
    if (!isNetworkError(error)) throw error
    return offline()
  }
}

function isLegacyPack(pack: OfflineWordTrainingPack | null): boolean {
  if (!pack) return false
  return !pack.queue?.length && !!pack.cards?.length
}

async function requirePack(): Promise<OfflineWordTrainingPack> {
  const pack = await getWordTrainingPack()
  const items = packQueueItems(pack)
  if (!pack || items.length === 0) throw new OfflineWordTrainingUnavailableError()
  return pack
}

function queueItemToResponse(session: OfflineWordTrainingSession, item: OfflineWordTrainingQueueItem): any {
  const shownAt = new Date().toISOString()
  session.shown_at = shownAt
  session.options_shown_at = item.type === 'card' ? undefined : shownAt
  void setWordTrainingSession(session)

  const base = {
    type: item.type,
    question: item.question,
    card_index: session.index + 1,
    total_cards: session.queue.length,
    session_id: session.id,
    user_card_id: item.user_card_id,
    training_card_id: item.training_card_id,
    word_card_id: item.word_card_id,
    direction: item.direction,
    word_en: item.word_en,
    word_target: item.word_target,
    word_ru: item.word_ru,
    word_native: item.word_native,
    display_word: item.display_word,
    display_target: item.display_target,
    transcription: item.transcription,
    example_en: item.example_en,
    example_target: item.example_target,
    hint: item.hint,
    word_category: item.word_category,
    morph: item.morph,
    delay_ms: 0,
    offline: true,
    correct_answer: undefined,
    options: undefined,
  }

  if (item.type === 'spell') {
    return {
      ...base,
      prefix: item.prefix,
      letters: item.letters,
    }
  }
  if (item.type === 'type') {
    return {
      ...base,
      prefix: item.prefix,
      hint_first_letter: item.hint_first_letter,
      hint_length: item.hint_length,
    }
  }
  return base
}

function toCardResponse(session: OfflineWordTrainingSession): any {
  if (session.index >= session.queue.length) {
    const response = {
      complete: true,
      cards_completed: session.queue.length,
      total_cards: session.queue.length,
      correct_cards: session.correct_count,
      offline: true,
    }
    void clearWordTrainingSession()
    return response
  }
  const item = session.queue[session.index]
  return queueItemToResponse(session, item)
}

function createID(prefix: string): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function queueCardAttempt(session: OfflineWordTrainingSession, item: OfflineWordTrainingQueueItem, optionIndex: number): Promise<any> {
  const options = item.options || []
  const optionsShownAt = session.options_shown_at || new Date().toISOString()
  const answeredAt = new Date().toISOString()
  const shownAt = session.shown_at || optionsShownAt
  const chosenOption = options[optionIndex] || ''
  const isCorrect = chosenOption === item.correct_answer
  const answerTimeMS = Math.max(0, new Date(answeredAt).getTime() - new Date(optionsShownAt).getTime())
  const tDelayMS = Math.max(0, new Date(optionsShownAt).getTime() - new Date(shownAt).getTime())
  const attempt: QueuedWordTrainingAttempt = {
    client_attempt_id: createID('offline-word-training'),
    user_card_id: item.user_card_id,
    training_card_id: item.training_card_id || 0,
    direction: item.direction,
    mode: 'card',
    shown_at: shownAt,
    options_shown_at: optionsShownAt,
    answered_at: answeredAt,
    t_delay_ms: tDelayMS,
    answer_time_ms: answerTimeMS,
    early_reveal: false,
    options,
    chosen_option: chosenOption,
    correct_answer: item.correct_answer,
  }
  await enqueueWordTrainingAttempt(attempt)
  await removeWordTrainingUserCards([item.user_card_id])
  if (isCorrect) session.correct_count++
  session.index++
  await setWordTrainingSession(session)
  return {
    is_correct: isCorrect,
    chosen_option: chosenOption,
    correct_answer: item.correct_answer,
    hint: isCorrect ? undefined : item.hint,
    example: isCorrect ? undefined : item.example_en,
    example_target: isCorrect ? undefined : item.example_target,
    delay_seconds: isCorrect ? undefined : 2,
    offline: true,
    queued: true,
  }
}

async function queueSpellTypeAttempt(session: OfflineWordTrainingSession, item: OfflineWordTrainingQueueItem, answerText: string): Promise<any> {
  const shownAt = session.shown_at || new Date().toISOString()
  const answeredAt = new Date().toISOString()
  const normalizedAnswer = answerText.trim().toLowerCase()
  const normalizedCorrect = item.correct_answer.trim().toLowerCase()
  const isCorrect = normalizedAnswer !== '' && normalizedAnswer === normalizedCorrect
  const answerTimeMS = Math.max(0, new Date(answeredAt).getTime() - new Date(shownAt).getTime())
  const attempt: QueuedWordTrainingAttempt = {
    client_attempt_id: createID('offline-word-training'),
    user_card_id: item.user_card_id,
    training_card_id: item.training_card_id || 0,
    direction: item.direction,
    mode: item.type,
    shown_at: shownAt,
    options_shown_at: shownAt,
    answered_at: answeredAt,
    t_delay_ms: 0,
    answer_time_ms: answerTimeMS,
    early_reveal: false,
    options: [],
    answer_text: answerText,
    correct_answer: item.correct_answer,
  }
  await enqueueWordTrainingAttempt(attempt)
  if (item.user_card_id > 0) {
    await removeWordTrainingUserCards([item.user_card_id])
  }
  if (isCorrect) session.correct_count++
  session.index++
  await setWordTrainingSession(session)
  return {
    is_correct: isCorrect,
    chosen_option: answerText,
    correct_answer: item.correct_answer,
    delay_seconds: isCorrect ? undefined : 2,
    offline: true,
    queued: true,
  }
}

export const wordTrainingClient = {
  async getOfflineStatus(): Promise<WordTrainingOfflineStatus> {
    const pack = await getWordTrainingPack()
    const pendingAttempts = await wordTrainingQueueCount()
    const items = packQueueItems(pack)
    return {
      ready: items.length > 0,
      downloadedCards: items.length,
      downloadedAt: pack?.downloaded_at,
      pendingAttempts,
      legacyPack: isLegacyPack(pack),
    }
  },

  async preload(): Promise<WordTrainingOfflineStatus> {
    const pack = await apiClient.request<OfflineWordTrainingPack>(withCourse('/api/training/offline/pack'))
    const queue = pack.queue?.length ? pack.queue : (pack.cards || []).map((card) => ({ ...card, type: 'card' as const }))
    await setWordTrainingPack({
      ...pack,
      downloaded_at: new Date().toISOString(),
      cards: pack.cards || [],
      queue,
    })
    await clearWordTrainingSession()
    return this.getOfflineStatus()
  },

  async clear(): Promise<void> {
    await clearWordTrainingPack()
  },

  async syncQueuedAttempts(): Promise<number> {
    if (isBrowserOffline()) return 0
    const attempts = await getQueuedWordTrainingAttempts()
    if (attempts.length === 0) return 0
    let response: any
    try {
      response = await apiClient.request('/api/training/offline/sync-attempts', {
        method: 'POST',
        body: { attempts } as any,
      })
    } catch (error) {
      if (isNetworkError(error)) return 0
      throw error
    }
    let synced = 0
    for (const item of response.results || []) {
      if (item.synced && item.client_attempt_id) {
        await deleteQueuedWordTrainingAttempt(item.client_attempt_id)
        synced++
      }
    }
    return synced
  },

  async getDashboard(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(withCourse('/api/dashboard')),
      async () => {
        const pack = await requirePack()
        const count = packQueueItems(pack).length
        return {
          due_count: count,
          total_cards: pack.total_cards || count,
          available_for_training: count,
          new_count: count,
          learning_count: 0,
          review_count: 0,
          known_count: 0,
          offline: true,
        }
      },
    )
  },

  async getUpcoming(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(withCourse('/api/training/upcoming')),
      async () => {
        const pack = await requirePack()
        const count = packQueueItems(pack).length
        const today = new Date().toISOString().slice(0, 10)
        return { [today]: { date: today, label: offlineLabel(), count } }
      },
    )
  },

  async start(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(withCourse('/api/training/start'), { method: 'POST' }),
      async () => {
        const pack = await requirePack()
        const source = packQueueItems(pack)
        const queue = shuffle(source).slice(0, Math.min(30, source.length))
        if (queue.length === 0) throw new OfflineWordTrainingUnavailableError('No preloaded cards available')
        const session: OfflineWordTrainingSession = { id: Date.now(), started_at: new Date().toISOString(), index: 0, correct_count: 0, queue }
        await setWordTrainingSession(session)
        return toCardResponse(session)
      },
    )
  },

  async current(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(withCourse('/api/training/current')),
      async () => {
        const session = await getWordTrainingSession()
        if (!session) return { active: false, message: 'No active offline session', offline: true }
        return toCardResponse(session)
      },
    )
  },

  async reveal(): Promise<any> {
    return offlineFallback(
      () => apiClient.request('/api/training/reveal', { method: 'POST', headers: { 'Content-Type': 'application/json' } }),
      async () => {
        const session = await getWordTrainingSession()
        if (!session || session.index >= session.queue.length) throw new OfflineWordTrainingUnavailableError('No active offline session')
        const item = session.queue[session.index]
        if (item.type !== 'card') throw new OfflineWordTrainingUnavailableError('Reveal is only available for card mode offline')
        session.options_shown_at = new Date().toISOString()
        await setWordTrainingSession(session)
        return { options: item.options, user_card_id: item.user_card_id, offline: true }
      },
    )
  },

  async answer(formData: FormData): Promise<any> {
    if (!isBrowserOffline()) {
      try {
        return await apiClient.requestFormData('/api/training/answer', formData)
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }
    const session = await getWordTrainingSession()
    if (!session || session.index >= session.queue.length) throw new OfflineWordTrainingUnavailableError('No active offline session')
    const item = session.queue[session.index]
    if (formData.has('answer_text')) {
      if (isLegacyPack(await getWordTrainingPack())) {
        throw new OfflineWordTrainingUnavailableError('Spell/type word training is not available offline yet. Please update preload.')
      }
      return queueSpellTypeAttempt(session, item, String(formData.get('answer_text') || ''))
    }
    if (item.type !== 'card') throw new OfflineWordTrainingUnavailableError('Use text answer for spell/type offline cards')
    const optionIndex = Number(formData.get('option_index') || '-1')
    if (!Number.isInteger(optionIndex) || optionIndex < 0 || optionIndex >= (item.options?.length || 0)) throw new Error('Invalid option index')
    return queueCardAttempt(session, item, optionIndex)
  },
}
