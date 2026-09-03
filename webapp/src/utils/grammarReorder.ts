// Mirrors repository/grammar_reorder.go for previously downloaded raw chapters.
// Never use an explanation or a partial sentence match as a translation.
const normalizeSentence = (text: string): string => text.trim().replace(/\s+/g, ' ')

export function grammarQuestionAvailable(question: any): boolean {
  return question?.type !== 'reorder' || (
    typeof question.translation_ru === 'string' && question.translation_ru.trim().length > 0
  )
}

export function chapterQuestionsWithTranslations(chapter: any): any[] {
  const questions = chapter?.question_bank?.questions || []
  if (chapter?.ui_language !== 'ru') return questions
  const translations = new Map<string, Map<string, string>>()
  for (const block of chapter.blocks || []) {
    const byText = new Map<string, string>()
    for (const example of block.theory?.examples || []) {
      if (typeof example.text !== 'string' || typeof example.translation !== 'string') continue
      const text = normalizeSentence(example.text)
      const translation = example.translation.trim()
      if (!text || !translation) continue
      byText.set(text, byText.has(text) && byText.get(text) !== translation ? '' : translation)
    }
    translations.set(block.id, byText)
  }
  return questions.map((question: any) => {
    if (grammarQuestionAvailable(question) || typeof question.correct_answer !== 'string') return question
    const translation = translations.get(question.theory_block_id)?.get(normalizeSentence(question.correct_answer))
    return translation ? { ...question, translation_ru: translation } : question
  })
}

export function chapterForQuestionDisplay(chapter: any): any {
  if (!chapter) return chapter
  const questions = chapterQuestionsWithTranslations(chapter).filter(grammarQuestionAvailable)
  const ids = new Set(questions.map((question: any) => question.id))
  const blocks = (chapter.blocks || []).flatMap((block: any) => {
    if (block.type !== 'quiz_inline') return [block]
    const questionIDs = (block.quiz_inline?.question_ids || []).filter((id: string) => ids.has(id))
    return questionIDs.length ? [{ ...block, quiz_inline: { ...block.quiz_inline, question_ids: questionIDs } }] : []
  })
  return { ...chapter, blocks, question_bank: { ...chapter.question_bank, questions } }
}
