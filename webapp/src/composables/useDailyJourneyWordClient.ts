import { wordTrainingClient as standalone } from '../api/wordTrainingClient'
import type { WordsJourneyExercise } from '../api/dailyJourneyClient'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { useDailyJourneyExercise } from './useDailyJourneyExercise'

// Scoped to this page instance: no global client mutation and no offline fallback
// for journey attempts. All answer/next operations retain the daily action key.
export function useDailyJourneyWordClient() {
  const daily = useDailyJourneyExercise<WordsJourneyExercise>('words')
  async function current() {
    const data = await daily.ready()
    if (data.phase !== 'train' || !data.card) throw new Error('Сначала добавьте слова в изучение.')
    return { ...data.card, journey_feedback: data.feedback, journey_options: data.options }
  }
  const client = {
    ...standalone,
    start: async () => daily.isJourney.value ? current() : standalone.start(),
    current: async () => daily.isJourney.value ? current() : standalone.current(),
    prefetchNext: async () => daily.isJourney.value ? null : standalone.prefetchNext(),
    reveal: async () => {
      if (!daily.isJourney.value) return standalone.reveal()
      await daily.act('start', { reveal: true })
      if (daily.error.value) throw new Error(daily.error.value)
      const data = await daily.ready()
      return { options: data.options || [], ...data.card }
    },
    answer: async (form: FormData) => {
      if (!daily.isJourney.value) return standalone.answer(form)
      const state = await daily.ready()
      if (!state.card || Number(form.get('session_id')) !== state.card.session_id || Number(form.get('card_index')) !== state.card.card_index) throw new Error('Карточка маршрута изменилась.')
      const payload: Record<string, unknown> = { card_index: Number(form.get('card_index')), user_card_id: form.has('user_card_id') ? Number(form.get('user_card_id')) : state.card.user_card_id }
      if (form.has('answer_text')) payload.answer_text = String(form.get('answer_text') || '')
      else payload.option_index = Number(form.get('option_index'))
      await daily.act('answer', payload)
      if (daily.error.value) throw new Error(daily.error.value)
      const data = await daily.ready()
      const course = daily.context.value?.courseCode
      if (course) emitAppDataEvent('word-review-recorded', course)
      return data.feedback
    },
    next: async () => {
      if (!daily.isJourney.value) return standalone.current()
      await daily.act('ack_feedback')
      if (daily.error.value) throw new Error(daily.error.value)
      if (daily.finished.value) { await daily.back(); return { complete: true } }
      return current()
    },
  }
  return { daily, client }
}
