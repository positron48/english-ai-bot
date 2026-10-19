import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useDailyJourneyWordClient } from './useDailyJourneyWordClient'
const { standalone, useExercise } = vi.hoisted(() => ({
 standalone: { start: vi.fn(), current: vi.fn(), prefetchNext: vi.fn(), reveal: vi.fn(), answer: vi.fn(), getDashboard: vi.fn() },
 useExercise: vi.fn(),
}))
vi.mock('../api/wordTrainingClient', () => ({ wordTrainingClient: standalone }))
vi.mock('./useDailyJourneyExercise', () => ({ useDailyJourneyExercise: useExercise }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: vi.fn() }))
let state: any
let daily: any
beforeEach(() => {
 vi.clearAllMocks()
 state = { phase: 'train', card: { session_id: -42, card_index: 1, user_card_id: 77 }, options: ['book', 'table'], feedback: { is_correct: false, correct_answer: 'book' } }
 daily = { isJourney: ref(true), error: ref(''), context: ref({ courseCode: 'en_ru' }), resultDetails: ref({ reviewed_words: 5, correct: 4 }), finished: ref(false), ready: vi.fn(async () => state), act: vi.fn(), back: vi.fn() }
 useExercise.mockReturnValue(daily)
})
describe('dailyJourney word source on the existing screen', () => {
 it('restores feedback without advancing a question or fetching the normal session', async () => {
  const { client } = useDailyJourneyWordClient()
  expect(await client.current()).toMatchObject({ session_id: -42, journey_feedback: { is_correct: false } })
  expect(await client.prefetchNext()).toBeNull()
  expect(daily.act).not.toHaveBeenCalled()
  expect(standalone.current).not.toHaveBeenCalled()
  expect(standalone.prefetchNext).not.toHaveBeenCalled()
 })
 it('sends spell/type answers with the bound user card, without offline/global fallback', async () => {
  const { client } = useDailyJourneyWordClient()
  const form = new FormData()
  form.set('session_id', '-42'); form.set('card_index', '1'); form.set('answer_text', 'book')
  await client.answer(form)
  expect(daily.act).toHaveBeenCalledWith('answer', { card_index: 1, user_card_id: 77, answer_text: 'book' })
  expect(standalone.answer).not.toHaveBeenCalled()
  daily.error.value = 'Network failure'
  await expect(client.answer(form)).rejects.toThrow('Network failure')
  expect(standalone.answer).not.toHaveBeenCalled()
 })
 it('returns ordinary session statistics and keeps the completion screen open', async () => {
  daily.act.mockImplementation(() => { daily.finished.value = true })
  const { client } = useDailyJourneyWordClient()
  expect(await client.next()).toEqual({ complete: true, cards_completed: 5, total_cards: 5, correct_cards: 4 })
  expect(daily.act).toHaveBeenCalledWith('ack_feedback')
  expect(daily.back).not.toHaveBeenCalled()
  expect(standalone.current).not.toHaveBeenCalled()
 })
 it('restores statistics when a completed route step is opened again', async () => {
  daily.ready.mockImplementation(async () => { daily.finished.value = true; throw new Error('No active card') })
  const { client } = useDailyJourneyWordClient()
  expect(await client.current()).toMatchObject({ complete: true, total_cards: 5, correct_cards: 4 })
  expect(daily.back).not.toHaveBeenCalled()
 })
 it('preserves standalone behavior when there is no journey context', async () => {
  daily.isJourney.value = false
  const { client } = useDailyJourneyWordClient()
  await client.current(); await client.start(); await client.reveal()
  const form = new FormData(); await client.answer(form)
  expect(standalone.current).toHaveBeenCalledOnce()
  expect(standalone.start).toHaveBeenCalledOnce()
  expect(standalone.answer).toHaveBeenCalledWith(form)
  expect(daily.act).not.toHaveBeenCalled()
 })
})
