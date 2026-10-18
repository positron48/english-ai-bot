import { computed } from 'vue'
import { clearCategoriesCache } from '../api/grammarClient'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { useDailyJourneyExercise } from './useDailyJourneyExercise'

export interface JourneyGrammarData {
  kind: 'block' | 'chapter' | 'category' | 'srs'
  chapters?: Record<string, any>
  chapter?: any
  questions: any[]
  index: number
  answers: Record<string, any>
  feedback: boolean
  results: any[]
  result?: { score: number; passed: boolean; correct: number; total: number; results: any[] } | null
}

export function useDailyJourneyGrammar(mode: 'grammar' | 'grammar_srs' = 'grammar') {
  const exercise = useDailyJourneyExercise<JourneyGrammarData>(mode)
  const question = computed(() => {
    const state = exercise.data.value
    const current = state?.questions[state.index]
    if (!current) return null
    if (!state.feedback) return current
    const result = state.results[state.index]
    return { ...current, correct_answer: result?.correct ? state.answers[state.index] : result?.correct_answer, explanation: result?.explanation }
  })
  async function answer(answer: unknown) {
    if (exercise.busy.value) return
    const state = await exercise.ready()
    if (state.feedback || state.result) return
    await exercise.act('answer', { index: state.index, answer })
    if (exercise.error.value) throw new Error(exercise.error.value)
    if (mode === 'grammar_srs') emitAppDataEvent('grammar-training-recorded', exercise.context.value?.courseCode)
    else if (exercise.data.value?.result) { clearCategoriesCache(); emitAppDataEvent('grammar-test-submitted', exercise.context.value?.courseCode) }
  }
  return { ...exercise, question, answer, acknowledge: () => exercise.act('ack_feedback') }
}
