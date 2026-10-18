import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useDailyJourneyGrammar } from './useDailyJourneyGrammar'
const { useExercise, emit } = vi.hoisted(() => ({ useExercise: vi.fn(), emit: vi.fn() }))
vi.mock('./useDailyJourneyExercise', () => ({ useDailyJourneyExercise: useExercise }))
vi.mock('../api/grammarClient', () => ({ clearCategoriesCache: vi.fn() }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: emit }))
let daily: any
beforeEach(() => {
 vi.clearAllMocks()
 daily = { data: ref({ kind: 'block', questions: [{ id: 'q1', prompt: 'Question' }], index: 0, answers: {}, results: [], feedback: false }), error: ref(''), busy: ref(false), context: ref({ courseCode: 'es_ru' }), ready: vi.fn(async () => daily.data.value), act: vi.fn() }
 useExercise.mockReturnValue(daily)
})
describe('grammar source for the existing chapter/test screens', () => {
 it('saves the bound question and waits for explicit acknowledgment of feedback', async () => {
  const grammar = useDailyJourneyGrammar()
  await grammar.answer('true')
  expect(useExercise).toHaveBeenCalledWith('grammar')
  expect(daily.act).toHaveBeenCalledExactlyOnceWith('answer', { index: 0, answer: 'true' })
  expect(emit).not.toHaveBeenCalled()
  daily.data.value = { ...daily.data.value, answers: { 0: 'false' }, results: [{ correct: false, correct_answer: 'true', explanation: 'Saved explanation' }], feedback: true }
  expect(grammar.question.value).toMatchObject({ correct_answer: 'true', explanation: 'Saved explanation' })
  await grammar.answer('true')
  expect(daily.act).toHaveBeenCalledTimes(1)
  await grammar.acknowledge()
  expect(daily.act).toHaveBeenLastCalledWith('ack_feedback')
 })
 it('invalidates only the explicit course after a confirmed full-test result', async () => {
  daily.data.value.kind = 'category'
  daily.act.mockImplementation(() => { daily.data.value = { ...daily.data.value, result: { score: 50, passed: true } } })
  await useDailyJourneyGrammar().answer('A')
  expect(emit).toHaveBeenCalledExactlyOnceWith('grammar-test-submitted', 'es_ru')
 })
 it('uses the SRS source and invalidates its own course after an answer', async () => {
  await useDailyJourneyGrammar('grammar_srs').answer('a')
  expect(useExercise).toHaveBeenCalledWith('grammar_srs')
  expect(emit).toHaveBeenCalledExactlyOnceWith('grammar-training-recorded', 'es_ru')
 })
 it('keeps failures visible without awarding a result or falling back', async () => {
  daily.act.mockImplementation(() => { daily.error.value = 'Pending retry' })
  await expect(useDailyJourneyGrammar().answer('A')).rejects.toThrow('Pending retry')
  expect(emit).not.toHaveBeenCalled()
 })
})
