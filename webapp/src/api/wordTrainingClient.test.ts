import 'fake-indexeddb/auto'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { wordTrainingClient } from './wordTrainingClient'
import { setWordTrainingPack, getQueuedWordTrainingAttempts, getWordTrainingSession } from './wordTrainingOfflineStore'
import { setGrammarCourse } from './grammarClient'
import { setActiveCourseCodeForInvalidation } from './cacheInvalidation'

let user = 7000
beforeEach(async () => {
  localStorage.clear()
  localStorage.setItem('access_token', `header.${btoa(JSON.stringify({ user_id: ++user }))}.sig`)
  setGrammarCourse('en_ru')
  setActiveCourseCodeForInvalidation('en_ru')
  vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false)
  await setWordTrainingPack({ app_code: 'english', native_lang: 'ru', target_lang: 'en', generated_at: '', algo_version: 'v2', total_cards: 2, available_count: 2, downloaded_at: '', cards: [], queue: [1, 2].map(id => ({ type: 'card', user_card_id: id, training_card_id: id, direction: 'ru_en', correct_answer: 'cat', options: ['cat', 'dog'] })) })
})
afterEach(() => { vi.restoreAllMocks() })

it('keeps the offline session after reconnection through reveal, answer and next card', async () => {
  const first = await wordTrainingClient.start()
  vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(true)
  const fetchMock = vi.fn().mockRejectedValue(new Error('must not use the server session'))
  vi.spyOn(globalThis, 'fetch').mockImplementation(fetchMock)
  const current = await wordTrainingClient.current()
  expect(current.session_id).toBe(first.session_id)
  expect((await wordTrainingClient.reveal()).options).toEqual(['cat', 'dog'])
  expect(await wordTrainingClient.prefetchNext()).toBeNull()
  const form = new FormData()
  form.set('session_id', String(first.session_id))
  form.set('card_index', '1')
  form.set('option_index', '0')
  await wordTrainingClient.answer(form)
  expect(await getQueuedWordTrainingAttempts()).toHaveLength(1)
  const next = await wordTrainingClient.current()
  expect(next.session_id).toBe(first.session_id)
  expect(next.card_index).toBe(2)
  expect(next.offline).toBe(true)
  expect(fetchMock).not.toHaveBeenCalled()
})

it('removes the old local session only after a successful explicit online start', async () => {
  await wordTrainingClient.start()
  vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(true)
  vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({ session_id: 123, card_index: 1 }), { status: 200 }))
  expect((await wordTrainingClient.start()).session_id).toBe(123)
  expect(await getWordTrainingSession()).toBeNull()
})
