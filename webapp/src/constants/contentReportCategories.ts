export const WORD_TRAINING_REPORT_CATEGORIES = [
  'wrong_translation',
  'wrong_example',
  'wrong_distractors',
  'typo',
  'bad_audio',
  'unclear_question',
  'other'
] as const

export const GRAMMAR_TRAINING_REPORT_CATEGORIES = [
  'wrong_answer',
  'ambiguous',
  'wrong_explanation',
  'theory_mismatch',
  'typo',
  'too_hard',
  'other'
] as const

export const GRAMMAR_CHAPTER_REPORT_CATEGORIES = [
  'wrong_theory',
  'wrong_example',
  'typo',
  'unclear_explanation',
  'other'
] as const

export const GRAMMAR_TEST_REPORT_CATEGORIES = [
  'wrong_answer',
  'ambiguous',
  'wrong_explanation',
  'typo',
  'too_hard',
  'other'
] as const

export type WordReportCategory = (typeof WORD_TRAINING_REPORT_CATEGORIES)[number]
export type GrammarReportCategory = (typeof GRAMMAR_TRAINING_REPORT_CATEGORIES)[number]
export type GrammarChapterReportCategory = (typeof GRAMMAR_CHAPTER_REPORT_CATEGORIES)[number]
export type GrammarTestReportCategory = (typeof GRAMMAR_TEST_REPORT_CATEGORIES)[number]

export function buildReportComment(
  category: string,
  details: string,
  categoryLabel: string
): string {
  const d = details.trim()
  if (d) {
    if (category === 'other') return d
    return `${categoryLabel}: ${d}`
  }
  if (category && category !== 'other') return categoryLabel
  return ''
}

export function canSubmitReport(category: string, details: string): boolean {
  const c = category.trim()
  const d = details.trim()
  if (!c) return false
  if (c === 'other') return d.length > 0
  return true
}
