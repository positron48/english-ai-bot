import { journeyMessage } from '../components/dailyJourney/messages'
import { apiClient } from '../api/client'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { useDailyJourneyExercise } from './useDailyJourneyExercise'
export interface JourneyVerbSession { session_id: number; card_id?: number; card_index: number; assisted?: boolean; feedback?: unknown; forms?: any[] }
export function useDailyJourneyVerbClient<T extends JourneyVerbSession>() {
 const daily = useDailyJourneyExercise<T>('verb_forms')
 async function current(): Promise<T> {
  if (!daily.isJourney.value) return apiClient.request<T>('/api/verb-training/v2/current')
  try {
   const state = await daily.ready()
   if (!daily.finished.value) return state
  } catch (error) { if (!daily.finished.value) throw error }
  if (daily.stepStatus?.value && daily.stepStatus.value !== 'completed') throw new Error(journeyMessage('stepUnavailable'))
  const details = daily.resultDetails?.value
  return (details?.practice || { completed: true, session_id: 0, card_index: Number(details?.answered || 0), total_cards: Number(details?.answered || 0), results: [], summary: { correct: Number(details?.correct || 0), total: Number(details?.answered || 0) } }) as T
 }
 async function act(action: string, body: Record<string, unknown>): Promise<T> {
  if (!daily.isJourney.value) return apiClient.request<T>('/api/verb-training/v2/' + action, { method: 'POST', body })
  const state = await daily.ready()
  if (action === 'start') return state
  if (body.session_id !== state.session_id || body.card_id !== state.card_id) throw new Error('Карточка маршрута изменилась. Обновите задание.')
  const payload = { index: state.card_index - 1, card_id: state.card_id }
  if (action === 'help') await daily.act('start', { ...payload, help: true })
  else if (action === 'answer') await daily.act('answer', { ...payload, answer: body.answer || '', skip: body.skip === true })
  else if (action === 'advance') await daily.act('ack_feedback')
  else throw new Error('Это действие недоступно в маршруте дня.')
  if (daily.error.value) throw new Error(daily.error.value)
  if (action === 'answer') emitAppDataEvent('verb-training-recorded', daily.context.value?.courseCode)
  return current()
 }
 async function forms(lemma: string): Promise<{ forms: any[] }> {
  if (!daily.isJourney.value) return apiClient.request('/api/verb-training/forms-by-lemma?lemma=' + encodeURIComponent(lemma))
  const state = await daily.ready()
  if (!state.assisted && !state.feedback) throw new Error('Сначала откройте подсказку.')
  return { forms: state.forms || [] }
 }
 return { daily, current, act, forms }
}
