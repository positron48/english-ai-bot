import { describe, expect, it, vi } from 'vitest'
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
 it('restores the saved wrong answer and leaves progression to an explicit acknowledgment', async () => {
  const quiz = state()
  quiz.answers = { 0: 'false' }; quiz.correct = { 0: false }; quiz.feedback = true; quiz.done = true
  quiz.questions[0].correct_answer = 'true'
  quiz.questions[0].explanation = 'Server explanation'
  const view = mount(ReadingPassageBlock, { props: { block: { title: 'Story', reading_passage: { segments: [] } }, isRead: false, journeyQuiz: quiz }, global: { stubs: { GrammarQuestion: true } } })
  const next = view.findAll('button').find(button => button.text() === 'dailyJourney.returnToRoute')!
  await next.trigger('click')
  expect(view.emitted('journey-next')).toHaveLength(1)
  expect(view.emitted('marked-read')).toBeUndefined()
  view.unmount()
 })
})
