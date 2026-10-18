import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import VerbFormsTrainingPanel from './VerbFormsTrainingPanel.vue'
const { useClient, current, act, forms, back } = vi.hoisted(() => ({ useClient: vi.fn(), current: vi.fn(), act: vi.fn(), forms: vi.fn(), back: vi.fn() }))
vi.mock('../composables/useDailyJourneyVerbClient', () => ({ useDailyJourneyVerbClient: useClient }))
vi.mock('../api/sessionScope', () => ({ currentUserScope: () => 'user-a' }))
vi.mock('vue-i18n', async original => ({ ...await original<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key, te: () => false }) }))
let state: any, daily: any
beforeEach(() => {
 vi.clearAllMocks(); sessionStorage.clear()
 state = { session_id: -42, card_id: 9, card_index: 1, total_cards: 2, input_mode: 'typed', prompt: { lemma: 'hablar', question: 'Yo ____', person: '1', number: 'singular' }, results: [] }
 daily = { isJourney: ref(true), context: ref({ courseCode: 'es_ru', journeyId: 'j1', stepId: 's1' }), error: ref(''), finished: ref(false), retry: vi.fn(), back }
 current.mockImplementation(async () => state);act.mockImplementation(async () => state)
 useClient.mockReturnValue({ daily, current, act, forms })
})
const global = { stubs: { ClickableText: { props: ['text', 'courseCode'], template: '<span>{{ text }}</span>' }, LgIcon: true } }
describe('daily verbs on the existing panel', () => {
 it('restores saved feedback and waits for explicit continuation', async () => {
  state.feedback = { outcome: 'incorrect', correct_answer: 'hablo', chosen_option: 'hablas', sentence: 'Yo hablo', translation: 'Я говорю' }
  const view = mount(VerbFormsTrainingPanel, { global });await flushPromises()
  expect(view.text()).toContain('Yo hablo')
  expect(act).not.toHaveBeenCalled()
  await view.findAll('button').find(button => button.text() === 'verbPractice.next')!.trigger('click');await flushPromises()
  expect(act).toHaveBeenCalledExactlyOnceWith('advance', { session_id: -42, card_id: 9 })
  view.unmount()
 })
 it('restores a draft only for the same user/course/journey/card and clears it after confirmation', async () => {
  const key = 'daily-journey:verb-draft:user-a:es_ru:j1:s1:9'
  sessionStorage.setItem(key, 'habl')
  const view = mount(VerbFormsTrainingPanel, { global });await flushPromises()
  expect((view.get('input').element as HTMLInputElement).value).toBe('habl')
  await view.get('input').setValue('hablo')
  expect(sessionStorage.getItem(key)).toBe('hablo')
  act.mockImplementation(async () => ({ ...state, feedback: { outcome: 'correct', correct_answer: 'hablo' } }))
  await view.get('form').trigger('submit');await flushPromises()
  expect(sessionStorage.getItem(key)).toBeNull()
  expect(act).toHaveBeenCalledWith('answer', { session_id: -42, card_id: 9, answer: 'hablo' })
  view.unmount()
 })
})
