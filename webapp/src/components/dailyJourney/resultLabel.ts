import type { JourneyStep } from './types'
type Translate = (key: string, values?: Record<string, unknown>) => string
export function journeyResultLabel(step: JourneyStep, t: Translate): string {
  const key = 'dailyJourney.results.'
  if (step.status === 'skipped') return t(key + 'skipped')
  if (step.status === 'unavailable') return t(key + 'unavailable')
  const result = step.resultDetails
  if (result) {
    if (step.mode === 'words' && typeof result.reviewed_words === 'number') return t(key + 'words', { reviewed: result.reviewed_words, added: result.added_words || 0 })
    if (step.mode === 'reading' && typeof result.total === 'number') return t(key + 'reading', { correct: result.correct || 0, total: result.total })
    if (step.mode === 'grammar' && result.kind === 'block') return t(key + 'block', { answered: result.answered || 0 })
    if (step.mode === 'grammar' && typeof result.score === 'number') return t(key + (result.passed ? 'testPassed' : 'testFailed'), { score: result.score })
    if ((step.mode === 'picture' || step.mode === 'conversation') && result.attempt_finished === true) return t(key + (result.passed ? 'questPassed' : 'questFailed'), { turns: result.turns || 0 })
    if (typeof result.answered === 'number') return t(key + 'practice', { correct: result.correct || 0, total: result.answered })
  }
  return step.result || t(key + (step.status === 'completed' ? 'completed' : 'unfinished'))
}
