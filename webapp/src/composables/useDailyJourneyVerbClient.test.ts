import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useDailyJourneyVerbClient } from './useDailyJourneyVerbClient'
const { request, source, emit } = vi.hoisted(() => ({ request: vi.fn(), source: vi.fn(), emit: vi.fn() }))
vi.mock('../api/client', () => ({ apiClient: { request } }))
vi.mock('./useDailyJourneyExercise', () => ({ useDailyJourneyExercise: source }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: emit }))
let daily: any, state: any
beforeEach(() => {
 vi.clearAllMocks()
 state = { session_id: -42, card_id: 9, card_index: 2, assisted: false, forms: [{ surface_form: 'hablo' }] }
 daily = { isJourney: ref(true), context: ref({ courseCode: 'es_ru' }), error: ref(''), finished: ref(false), ready: vi.fn(async () => state), act: vi.fn(), back: vi.fn() }
 source.mockReturnValue(daily)
})
describe('verb source for the ordinary practice panel', () => {
 it('restores the bound session and persists help before showing forms', async () => {
  const client = useDailyJourneyVerbClient()
  expect(await client.current()).toEqual(state)
  await expect(client.forms('hablar')).rejects.toThrow('подсказку')
  daily.act.mockImplementation(() => { state.assisted = true })
  await client.act('help', { session_id: -42, card_id: 9 })
  expect(daily.act).toHaveBeenCalledWith('start', { index: 1, card_id: 9, help: true })
  expect(await client.forms('hablar')).toEqual({ forms: state.forms })
  expect(request).not.toHaveBeenCalled()
 })
 it('sends answers to the confirmed card and rejects stale panels', async () => {
  const client = useDailyJourneyVerbClient()
  await expect(client.act('answer', { session_id: -42, card_id: 8, answer: 'hablo' })).rejects.toThrow('изменилась')
  expect(daily.act).not.toHaveBeenCalled()
  await client.act('answer', { session_id: -42, card_id: 9, answer: 'hablo' })
  expect(daily.act).toHaveBeenCalledExactlyOnceWith('answer', { index: 1, card_id: 9, answer: 'hablo', skip: false })
  expect(emit).toHaveBeenCalledWith('verb-training-recorded', 'es_ru')
  expect(request).not.toHaveBeenCalled()
 })
 it('keeps feedback until acknowledgment and never starts a repeat session', async () => {
  const client = useDailyJourneyVerbClient()
  await expect(client.act('repeat', { session_id: -42, card_id: 9 })).rejects.toThrow('недоступно')
  daily.act.mockImplementation(() => { daily.finished.value = true })
  await client.act('advance', { session_id: -42, card_id: 9 })
  expect(daily.act).toHaveBeenCalledWith('ack_feedback')
  expect(daily.back).toHaveBeenCalledOnce()
  expect(request).not.toHaveBeenCalled()
 })
 it('keeps ordinary APIs unchanged and never uses them as a daily error fallback', async () => {
  const client = useDailyJourneyVerbClient()
  daily.ready.mockRejectedValue(new Error('Network failure'))
  await expect(client.current()).rejects.toThrow('Network failure')
  expect(request).not.toHaveBeenCalled()
  daily.isJourney.value = false
  await client.current(); await client.act('start', {}); await client.forms('hablar')
  expect(request).toHaveBeenCalledWith('/api/verb-training/v2/current')
  expect(request).toHaveBeenCalledWith('/api/verb-training/v2/start', { method: 'POST', body: {} })
  expect(request).toHaveBeenCalledWith('/api/verb-training/forms-by-lemma?lemma=hablar')
 })
})
