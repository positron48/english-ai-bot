import 'fake-indexeddb/auto'
import { beforeEach, expect, it } from 'vitest'
import { enqueueAttempt, getQueuedAttempts, clearOfflineGrammar } from './grammarOfflineStore'
import { getWordTrainingPack, setWordTrainingPack, clearWordTrainingPack, enqueueWordTrainingAttempt, getQueuedWordTrainingAttempts } from './wordTrainingOfflineStore'
import { setActiveCourseCodeForInvalidation } from './cacheInvalidation'
import { currentUserScope } from './sessionScope'

let user = 100
function login(id: number, suffix = '') {
  localStorage.setItem('access_token', `header.${btoa(JSON.stringify({ user_id: id }))}.${suffix}`)
}
beforeEach(() => { localStorage.clear(); login(++user); setActiveCourseCodeForInvalidation('en_ru') })

it('keeps grammar answers with their account across logout/login and token refresh', async () => {
  await enqueueAttempt({ client_attempt_id: 'same-id', course_code: 'en_ru', scope: 'chapter', scope_id: 'one', answers: [], created_at: '2026-09-18', result: {} })
  login(user + 1000)
  expect(await getQueuedAttempts('en_ru')).toEqual([])
  localStorage.setItem('me_profile_cache_v1', JSON.stringify({ data: { id: user } }))
  expect(currentUserScope()).toBe(`user:${user + 1000}`)
  login(user, 'refreshed')
  expect(await getQueuedAttempts('en_ru')).toHaveLength(1)
  localStorage.removeItem('access_token')
  expect(await getQueuedAttempts('en_ru')).toEqual([])
})

it('keeps word packs per course and preserves pending answers on pack removal', async () => {
  const pack = { app_code: 'linglow', native_lang: 'ru', target_lang: 'en', generated_at: '', algo_version: '', total_cards: 1, available_count: 1, downloaded_at: '', cards: [] }
  await setWordTrainingPack(pack)
  setActiveCourseCodeForInvalidation('es_ru')
  expect(await getWordTrainingPack()).toBeNull()
  await setWordTrainingPack({ ...pack, target_lang: 'es' })
  setActiveCourseCodeForInvalidation('en_ru')
  expect((await getWordTrainingPack())?.target_lang).toBe('en')
  await enqueueWordTrainingAttempt({ client_attempt_id: 'answer', user_card_id: 1, training_card_id: 1, direction: 'ru_en', mode: 'card', shown_at: '', options_shown_at: '', answered_at: '', t_delay_ms: 0, answer_time_ms: 0, early_reveal: false, options: [], correct_answer: 'answer' })
  await clearWordTrainingPack()
  expect(await getQueuedWordTrainingAttempts()).toHaveLength(1)
  setActiveCourseCodeForInvalidation('es_ru')
  expect((await getWordTrainingPack())?.target_lang).toBe('es')
  login(user + 1000)
  expect(await getWordTrainingPack()).toBeNull()
  expect(await getQueuedWordTrainingAttempts()).toEqual([])
})

it('rejects an operation when the account changes while IndexedDB is opening', async () => {
  const pending = enqueueAttempt({ client_attempt_id: 'opening', course_code: 'en_ru', scope: 'chapter', scope_id: 'one', answers: [], created_at: '', result: {} })
  login(user + 1000)
  await expect(pending).rejects.toThrow('Session changed')
  expect(await getQueuedAttempts('en_ru')).toEqual([])
})

it('captures the course key before asynchronous IndexedDB work', async () => {
  const pack = { app_code: 'linglow', native_lang: 'ru', target_lang: 'en', generated_at: '', algo_version: '', total_cards: 1, available_count: 1, downloaded_at: '', cards: [] }
  const pending = setWordTrainingPack(pack)
  setActiveCourseCodeForInvalidation('es_ru')
  await pending
  expect(await getWordTrainingPack()).toBeNull()
  setActiveCourseCodeForInvalidation('en_ru')
  expect((await getWordTrainingPack())?.target_lang).toBe('en')
})

it('preserves pending grammar attempts when clearing all downloaded content', async () => {
  await enqueueAttempt({ client_attempt_id: 'pending', scope: 'chapter', scope_id: 'one', answers: [], created_at: '', result: {} })
  await clearOfflineGrammar()
  expect(await getQueuedAttempts()).toHaveLength(1)
})
