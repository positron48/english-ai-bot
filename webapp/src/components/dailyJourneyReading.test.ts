import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import ReadingPassageBlock from './ReadingPassageBlock.vue'
import type { ReadingJourneyExercise } from '../api/dailyJourneyClient'
const { request } = vi.hoisted(() => ({ request: vi.fn(async (_url: string) => ({ settings: {} })) }))
vi.mock('../api/client', () => ({ apiClient: { request } }))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key, locale: ref('en') }) }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push: vi.fn() }) }))
vi.mock('../composables/useSettings', () => ({ useSettings: () => ({ settings: ref({ soundsEnabled: false, vibrationEnabled: false }) }) }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: ref('en_ru') }) }))
vi.mock('../composables/useAudio', () => ({ useAudio: () => ({ playSuccess: vi.fn(), playFail: vi.fn() }) }))
const state = (): ReadingJourneyExercise => ({ document: { id: 'text-a', category_id: 'a0', title: 'Story', reading_passage: { segments: [] } }, questions: [{ id: 'q1', type: 'true_false', prompt: 'A question' }], answers: {}, correct: {}, index: 0, quiz_open: true, feedback: false, done: false, passed: false })
afterEach(() => { vi.useRealTimers() })
describe('dailyJourney in the existing reading block', () => {
 it('sends answers to the parent and never marks the text read locally', async () => {
  const view = mount(ReadingPassageBlock, {
   props: { block: { title: 'Story', reading_passage: { segments: [] } }, textId: 'text-a', isRead: false, journeyQuiz: state() },
   global: { stubs: { GrammarQuestion: { emits: ['answer'], template: '<button data-test="answer" @click="$emit(\'answer\', \'false\')">Answer</button>' } } },
  })
  await flushPromises()
  await view.get('[data-test="answer"]').trigger('click')
  expect(view.emitted('journey-answer')?.[0]).toEqual(['false'])
  expect(view.emitted('marked-read')).toBeUndefined()
  expect(request.mock.calls.some(call => String(call[0]).includes('mark-read'))).toBe(false)
  view.unmount()
 })
 it('shows the ordinary failed result and allows retrying the daily quiz', async () => {
  const quiz = state()
  quiz.answers = { 0: 'false' }; quiz.correct = { 0: false }; quiz.feedback = true; quiz.done = true
  quiz.questions[0].correct_answer = 'true'
  quiz.questions[0].explanation = 'Server explanation'
  const view = mount(ReadingPassageBlock, { props: { block: { title: 'Story', reading_passage: { segments: [] } }, isRead: false, journeyQuiz: quiz }, global: { stubs: { GrammarQuestion: true } } })
  const next = view.findAll('button').find(button => button.text() === 'reading.quizRetry')!
  await next.trigger('click')
  expect(view.emitted('journey-open-quiz')).toHaveLength(1)
  expect(view.emitted('marked-read')).toBeUndefined()
  view.unmount()
 })
 it('automatically advances after confirmed feedback and keeps the question mounted while loading', async () => {
  vi.useFakeTimers()
  const quiz = state()
  quiz.questions.push({ id: 'q2', type: 'true_false', prompt: 'Next question' })
  const view = mount(ReadingPassageBlock, { props: { isRead: false, block: { title: 'Story', reading_passage: { segments: [] } }, journeyQuiz: quiz }, global: { stubs: { GrammarQuestion: true } } })
  await view.setProps({ journeyBusy: true })
  expect(view.findComponent({ name: 'GrammarQuestion' }).exists()).toBe(true)
  await view.setProps({ journeyQuiz: { ...quiz, feedback: true, answers: { 0: 'true' }, correct: { 0: true } }, journeyBusy: false })
  expect(view.text()).not.toContain('dailyJourney.continue')
  await vi.advanceTimersByTimeAsync(699)
  expect(view.emitted('journey-next')).toBeUndefined()
  await vi.advanceTimersByTimeAsync(1)
  expect(view.emitted('journey-next')).toHaveLength(1)
  view.unmount()
 })
 it('retains the passed result until the user returns to the route', async () => {
  vi.useFakeTimers()
  const quiz = { ...state(), feedback: true, done: true, passed: true, answers: { 0: 'true' }, correct: { 0: true } }
  const view = mount(ReadingPassageBlock, { props: { isRead: false, block: { title: 'Story', reading_passage: { segments: [] } }, journeyQuiz: quiz }, global: { stubs: { GrammarQuestion: true } } })
  await vi.advanceTimersByTimeAsync(2000)
  expect(view.emitted('journey-next')).toBeUndefined()
  await view.findAll('button').find(button => button.text() === 'dailyJourney.returnToRoute')!.trigger('click')
  expect(view.emitted('journey-back')).toHaveLength(1)
  view.unmount()
 })

 it('renders the same true/false labels as standalone reading', async () => {
  const question = { id: 'q1', type: 'true_false', prompt: 'A question', correct_answer: 'true' }
  const block = { title: 'Story', reading_passage: { segments: [], comprehension_questions: [question] } }
  const standalone = mount(ReadingPassageBlock, { props: { block, isRead: false } })
  await standalone.findAll('button').find(button => button.text() === 'reading.markRead')!.trigger('click')
  const daily = mount(ReadingPassageBlock, { props: { block, isRead: false, journeyQuiz: state() } })
  const labels = (view: typeof daily) => view.getComponent({ name: 'GrammarQuestion' }).findAll('.choice-btn').map(button => button.text())
  expect(labels(daily)).toEqual(labels(standalone))
  expect(labels(daily)).toHaveLength(2)
  standalone.unmount(); daily.unmount()
 })

})
