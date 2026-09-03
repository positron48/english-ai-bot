import { describe, expect, it } from 'vitest'
import { chapterForQuestionDisplay, chapterQuestionsWithTranslations, grammarQuestionAvailable } from './grammarReorder'

const reorderChapterFixture = () => ({
  id: 'ch', ui_language: 'ru', blocks: [
    { id: 'b1', type: 'theory', theory: { examples: [
      { text: 'I work at home.', translation: 'Я работаю дома.' },
      { text: 'I stay.', translation: 'Я остаюсь.' },
      { text: 'I stay.', translation: 'Я останусь.' },
      { text: 'I stay.', translation: 'Я остаюсь.' },
    ] } },
    { id: 'quiz', type: 'quiz_inline', quiz_inline: { question_ids: ['exact', 'partial', 'choice'] } },
    { id: 'empty', type: 'quiz_inline', quiz_inline: { question_ids: ['partial'] } },
  ],
  question_bank: { questions: [
    { id: 'exact', type: 'reorder', theory_block_id: 'b1', correct_answer: ' I  work at home. ' },
    { id: 'explicit', type: 'reorder', translation_ru: 'Я работаю из дома.', correct_answer: 'I work at home.' },
    { id: 'partial', type: 'reorder', theory_block_id: 'b1', correct_answer: 'I work.' },
    { id: 'wrong-block', type: 'reorder', theory_block_id: 'b2', correct_answer: 'I work at home.' },
    { id: 'ambiguous', type: 'reorder', theory_block_id: 'b1', correct_answer: 'I stay.' },
    { id: 'choice', type: 'mcq_single', correct_answer: 'a' },
  ] },
})

describe('grammar reorder translations', () => {
  it('reuses only an exact, unambiguous example in the same Russian theory block', () => {
    const questions = chapterQuestionsWithTranslations(reorderChapterFixture())
    expect(questions[0].translation_ru).toBe('Я работаю дома.')
    expect(questions[1].translation_ru).toBe('Я работаю из дома.')
    expect(questions.filter(grammarQuestionAvailable).map(q => q.id)).toEqual(['exact', 'explicit', 'choice'])
    expect(questions).toHaveLength(6)
  })

  it('removes unavailable quiz references and empty quizzes without mutating cached content', () => {
    const chapter = reorderChapterFixture()
    const before = JSON.stringify(chapter)
    const display = chapterForQuestionDisplay(chapter)
    expect(display.question_bank.questions.map((q: any) => q.id)).toEqual(['exact', 'explicit', 'choice'])
    expect(display.blocks).toHaveLength(2)
    expect(display.blocks[1].quiz_inline.question_ids).toEqual(['exact', 'choice'])
    expect(JSON.stringify(chapter)).toBe(before)
  })

  it('does not use another language or blank/invalid translations', () => {
    const chapter = reorderChapterFixture()
    chapter.ui_language = 'en'
    expect(grammarQuestionAvailable(chapterQuestionsWithTranslations(chapter)[0])).toBe(false)
    for (const translation_ru of [undefined, '', ' \n\t', 42]) {
      expect(grammarQuestionAvailable({ type: 'reorder', translation_ru })).toBe(false)
    }
  })
})
