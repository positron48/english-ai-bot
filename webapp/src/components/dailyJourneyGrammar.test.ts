import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { computed, ref } from 'vue'
import GrammarChapterView from '../views/GrammarChapterView.vue'
import GrammarTrainingView from '../views/GrammarTrainingView.vue'
import GrammarQuestion from './GrammarQuestion.vue'
import GrammarTestView from '../views/GrammarTestView.vue'
const { getChapter, getChapterTest, getCategoryTest, submitTest, getTrainingAvailability, startTrainingSession, submitTrainingAnswer, getChapterForTheory, useSource, push } = vi.hoisted(() => ({ getChapter: vi.fn(), getChapterTest: vi.fn(), getCategoryTest: vi.fn(), submitTest: vi.fn(), getTrainingAvailability: vi.fn(), startTrainingSession: vi.fn(), submitTrainingAnswer: vi.fn(), getChapterForTheory: vi.fn(), useSource: vi.fn(), push: vi.fn() }))
vi.mock('../api/grammarClient', () => ({ grammarClient: { getChapter, getChapterTest, getCategoryTest, submitTest, getTrainingAvailability, startTrainingSession, submitTrainingAnswer, getChapterForTheory }, getGrammarCourseCode: () => 'en_ru', clearCategoriesCache: vi.fn() }))
vi.mock('../composables/useDailyJourneyGrammar', () => ({ useDailyJourneyGrammar: useSource }))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key, tm: () => [], locale: ref('ru') }) }))
vi.mock('vue-router', () => ({ useRoute: () => ({ path: '/learning/grammar/chapter/ch1/test', params: { chapterId: 'ch1' }, query: { journey_id: 'j1' } }), useRouter: () => ({ push, back: push }), onBeforeRouteUpdate: vi.fn() }))
vi.mock('../composables/useSettings', () => ({ useSettings: () => ({ settings: ref({ soundsEnabled: false, vibrationEnabled: false }) }) }))
vi.mock('../composables/useAudio', () => ({ useAudio: () => ({ playSuccess: vi.fn(), playFail: vi.fn(), playVictory: vi.fn(), playDefeat: vi.fn() }) }))
let daily: any
beforeEach(() => {
 vi.clearAllMocks()
 daily = { isJourney: ref(true), data: ref({ kind: 'block', chapter: { id: 'ch1', title: 'Bound chapter', blocks: [{ id: 'b1', type: 'theory', theory: { content_md: 'Bound theory' } }] }, questions: [{ id: 'q1', prompt: 'Bound question', type: 'true_false' }], index: 0, answers: {}, results: [], feedback: false }), context: ref({ courseCode: 'es_ru' }), error: ref(''), busy: ref(false), finished: ref(false), ready: vi.fn(async () => daily.data.value), answer: vi.fn(), acknowledge: vi.fn(), back: vi.fn(), retry: vi.fn() }
 daily.question = computed(() => daily.data.value.questions[daily.data.value.index])
 useSource.mockReturnValue(daily)
})
const global = { stubs: { LgPageHeader: true, GrammarTheoryExamples: true, ContentReportDialog: true, ConfirmModal: true, GrammarQuestion: { props: ['question'], emits: ['answer'], template: '<button data-test="answer" @click="$emit(\'answer\', \'true\')">{{ question.prompt }}</button>' } } }
describe('daily grammar in the ordinary pages', () => {
 it('renders the bound theory and sends inline answers to the server source', async () => {
  const view = mount(GrammarChapterView, { global })
  await flushPromises()
  expect(view.text()).toContain('Bound theory')
  await view.get('[data-test="answer"]').trigger('click')
  expect(daily.answer).toHaveBeenCalledWith('true')
  expect(getChapter).not.toHaveBeenCalled()
  expect(view.text()).not.toContain('grammar.startChapterTest')
  view.unmount()
 })
 it('saves a test answer only on explicit next/submit, without the normal submit API', async () => {
  daily.data.value.kind = 'chapter'
  const view = mount(GrammarTestView, { global })
  await flushPromises()
  await view.get('[data-test="answer"]').trigger('click')
  expect(daily.answer).not.toHaveBeenCalled()
  await view.findAll('button').find(b => b.text() === 'grammar.submitTest')!.trigger('click')
  await flushPromises()
  expect(daily.answer).toHaveBeenCalledWith('true')
  expect(getChapterTest).not.toHaveBeenCalled()
  expect(getCategoryTest).not.toHaveBeenCalled()
  expect(submitTest).not.toHaveBeenCalled()
  view.unmount()
 })
 it('restores SRS feedback on the ordinary screen without starting a standalone session', async () => {
  daily.data.value = { ...daily.data.value, kind: 'srs', chapters: {}, answers: { 0: 'false' }, feedback: true, results: [{ correct: false, correct_answer: 'true', explanation: 'Saved SRS feedback' }] }
  const view = mount(GrammarTrainingView, { global })
  await flushPromises()
  expect(useSource).toHaveBeenCalledWith('grammar_srs')
  expect(view.text()).toContain('Saved SRS feedback')
  expect(getTrainingAvailability).not.toHaveBeenCalled()
  expect(startTrainingSession).not.toHaveBeenCalled()
  expect(getChapterForTheory).not.toHaveBeenCalled()
  await view.findAll('button').find(b => b.text() === 'common.finish')!.trigger('click')
  expect(daily.acknowledge).toHaveBeenCalledOnce()
  view.unmount()
 })
 it('renders and restores reorder words without exposing the correct sentence', async () => {
  const question = { id: 'r1', type: 'reorder' as const, prompt: 'Reorder', reorder_words: ['I', 'am', 'here'], reorder_punctuation: '.' }
  const view = mount(GrammarQuestion, { props: { question, initialAnswer: 'I am here.', showAnswers: false } })
  await flushPromises()
  await new Promise(resolve => setTimeout(resolve, 5))
  expect(view.findAll('.sentence-word').map(word => word.text())).toEqual(['I', 'am', 'here'])
  expect(view.findAll('.available-word')).toHaveLength(0)
  expect(view.find('.correct-answer-display').exists()).toBe(false)
  view.unmount()
 })

})

it('restores a fill draft without submitting or locking it, then clears on confirmation', async () => {
 const key = 'journey:test:fill'
 sessionStorage.setItem(key, JSON.stringify({ answer: 'par', selected: [], available: [] }))
 const view = mount(GrammarQuestion, { props: { question: { id: 'draft-fill', type: 'fill_blank', prompt: '___' }, draftKey: key }, global: { stubs: { Icon: true } } })
 await flushPromises()
 expect((view.get('input').element as HTMLInputElement).value).toBe('par')
 expect(view.get('input').attributes('disabled')).toBeUndefined()
 expect(view.emitted('answer')).toBeUndefined()
 await view.get('input').setValue('partial answer')
 await flushPromises()
 expect(JSON.parse(sessionStorage.getItem(key)!).answer).toBe('partial answer')
 await view.setProps({ initialAnswer: 'confirmed answer' })
 await flushPromises()
 expect(sessionStorage.getItem(key)).toBeNull()
 view.unmount()
})

it('restores a partial reorder only if its word multiset still matches the question', async () => {
 const key = 'journey:test:reorder'
 sessionStorage.setItem(key, JSON.stringify({ answer: null, selected: ['I'], available: ['am', 'here'] }))
 const view = mount(GrammarQuestion, { props: { question: { id: 'draft-reorder', type: 'reorder', prompt: 'Arrange', reorder_words: ['I','am','here'], reorder_punctuation: '.' }, draftKey: key }, global: { stubs: { Icon: true } } })
 await flushPromises()
 expect(view.findAll('.sentence-word').map(word => word.text())).toEqual(['I'])
 expect(view.emitted('answer')).toBeUndefined()
 view.unmount()
})

it('keeps the SRS card mounted and its options stable during a delayed answer and feedback reload', async () => {
 const question = { id: 'srs-stable', type: 'mcq_single', prompt: 'Choose', chapter_id: 'ch1', theory_block_id: 'b1', choices: [{ id: 'a', text: 'One' }, { id: 'b', text: 'Two' }, { id: 'c', text: 'Three' }] }
 daily.data.value = { ...daily.data.value, kind: 'srs', chapters: { ch1: daily.data.value.chapter }, questions: [question] }
 let resolve!: () => void
 const pending = new Promise<void>(done => { resolve = done })
 daily.answer.mockImplementation(async (answer: string) => {
  daily.busy.value = true
  await pending
  daily.data.value = { ...daily.data.value, questions: [{ ...question, correct_answer: answer }], feedback: true, answers: { 0: answer }, results: [{ correct: true, correct_answer: answer }] }
  daily.busy.value = false
 })
 const random = vi.spyOn(Math, 'random').mockReturnValue(0)
 const view = mount(GrammarTrainingView, { global: { stubs: { ...global.stubs, GrammarQuestion: false } } })
 try {
  await flushPromises()
  const component = view.getComponent(GrammarQuestion)
  expect(component.props('theoryBlock')?.id).toBe('b1')
  const element = component.element
  const labels = () => component.findAll('.choice-content').map(choice => choice.text())
  const before = labels()
  random.mockReturnValue(0.999)
  await component.findAll('.choice-btn')[0].trigger('click')
  await flushPromises()
  expect(view.getComponent(GrammarQuestion).element).toBe(element)
  expect(labels()).toEqual(before)
  expect(view.get('.card').attributes('inert')).toBeDefined()
  resolve()
  await flushPromises()
  expect(view.getComponent(GrammarQuestion).element).toBe(element)
  expect(labels()).toEqual(before)
  expect(component.findAll('.choice-btn')[0].classes()).toContain('correct')
  expect(getChapterForTheory).not.toHaveBeenCalled()
 } finally { resolve(); view.unmount(); random.mockRestore() }
})
