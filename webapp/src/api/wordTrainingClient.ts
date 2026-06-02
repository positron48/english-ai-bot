import { apiClient } from './client'
import {
  OfflineWordTrainingCard,
  OfflineWordTrainingPack,
  OfflineWordTrainingSession,
  QueuedWordTrainingAttempt,
  clearWordTrainingPack,
  clearWordTrainingSession,
  deleteQueuedWordTrainingAttempt,
  enqueueWordTrainingAttempt,
  getQueuedWordTrainingAttempts,
  getWordTrainingPack,
  getWordTrainingSession,
  removeWordTrainingCards,
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
}

const isBrowserOffline = () => typeof navigator !== 'undefined' && navigator.onLine === false
const isNetworkError = (error: any) => error?.isNetworkError || error?.name === 'TypeError' || String(error?.message || '').includes('Failed to fetch')

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

async function requirePack(): Promise<OfflineWordTrainingPack> {
  const pack = await getWordTrainingPack()
  if (!pack || !Array.isArray(pack.cards) || pack.cards.length === 0) throw new OfflineWordTrainingUnavailableError()
  return pack
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
  const shownAt = new Date().toISOString()
  session.shown_at = shownAt
  session.options_shown_at = undefined
  void setWordTrainingSession(session)
  return {
    ...item,
    type: 'card',
    card_index: session.index + 1,
    total_cards: session.queue.length,
    session_id: session.id,
    delay_ms: 0,
    offline: true,
    options: undefined,
    correct_answer: undefined,
  }
}

function createID(prefix: string): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

async function queueAttempt(session: OfflineWordTrainingSession, item: OfflineWordTrainingCard, optionIndex: number): Promise<any> {
  const optionsShownAt = session.options_shown_at || new Date().toISOString()
  const answeredAt = new Date().toISOString()
  const shownAt = session.shown_at || optionsShownAt
  const chosenOption = item.options[optionIndex] || ''
  const isCorrect = chosenOption === item.correct_answer
  const answerTimeMS = Math.max(0, new Date(answeredAt).getTime() - new Date(optionsShownAt).getTime())
  const tDelayMS = Math.max(0, new Date(optionsShownAt).getTime() - new Date(shownAt).getTime())
  const attempt: QueuedWordTrainingAttempt = {
    client_attempt_id: createID('offline-word-training'),
    user_card_id: item.user_card_id,
    training_card_id: item.training_card_id,
    direction: item.direction,
    mode: 'card',
    shown_at: shownAt,
    options_shown_at: optionsShownAt,
    answered_at: answeredAt,
    t_delay_ms: tDelayMS,
    answer_time_ms: answerTimeMS,
    early_reveal: false,
    options: item.options,
    chosen_option: chosenOption,
    correct_answer: item.correct_answer,
  }
  await enqueueWordTrainingAttempt(attempt)
  await removeWordTrainingCards([item.user_card_id])
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

export const wordTrainingClient = {
  async getOfflineStatus(): Promise<WordTrainingOfflineStatus> {
    const pack = await getWordTrainingPack()
    const pendingAttempts = await wordTrainingQueueCount()
    return {
      ready: !!pack && (pack.cards?.length || 0) > 0,
      downloadedCards: pack?.cards?.length || 0,
      downloadedAt: pack?.downloaded_at,
      pendingAttempts,
    }
  },

  async preload(): Promise<WordTrainingOfflineStatus> {
    const pack = await apiClient.request<OfflineWordTrainingPack>('/api/training/offline/pack')
    await setWordTrainingPack({ ...pack, downloaded_at: new Date().toISOString(), cards: pack.cards || [] })
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
      () => apiClient.request('/api/dashboard'),
      async () => {
        const pack = await requirePack()
        return { due_count: pack.cards.length, total_cards: pack.cards.length, available_for_training: pack.cards.length, offline: true }
      },
    )
  },

  async getUpcoming(): Promise<any> {
    return offlineFallback(
      () => apiClient.request('/api/training/upcoming'),
      async () => {
        const pack = await requirePack()
        const today = new Date().toISOString().slice(0, 10)
        return { [today]: { date: today, label: 'Offline', count: pack.cards.length } }
      },
    )
  },

  async start(): Promise<any> {
    return offlineFallback(
      () => apiClient.request('/api/training/start', { method: 'POST' }),
      async () => {
        const pack = await requirePack()
        const queue = shuffle(pack.cards).slice(0, Math.min(30, pack.cards.length))
        if (queue.length === 0) throw new OfflineWordTrainingUnavailableError('No preloaded cards available')
        const session: OfflineWordTrainingSession = { id: Date.now(), started_at: new Date().toISOString(), index: 0, correct_count: 0, queue }
        await setWordTrainingSession(session)
        return toCardResponse(session)
      },
    )
  },

  async current(): Promise<any> {
    return offlineFallback(
      () => apiClient.request('/api/training/current'),
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
    if (formData.has('answer_text')) throw new OfflineWordTrainingUnavailableError('Spell/type word training is not available offline yet')
    const item = session.queue[session.index]
    const optionIndex = Number(formData.get('option_index') || '-1')
    if (!Number.isInteger(optionIndex) || optionIndex < 0 || optionIndex >= item.options.length) throw new Error('Invalid option index')
    return queueAttempt(session, item, optionIndex)
  },
}
