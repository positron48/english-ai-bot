import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { grammarClient, setGrammarCourse } from './grammarClient'
import { getOfflineMeta, getStoredChapter, getStoredChapters, getTrainingQuestions } from './grammarOfflineStore'

vi.mock('./grammarOfflineStore', async importOriginal => ({
  ...await importOriginal<typeof import('./grammarOfflineStore')>(),
  getOfflineMeta: vi.fn(), getStoredChapter: vi.fn(), getStoredChapters: vi.fn(), getTrainingQuestions: vi.fn(),
}))

function chapterPayload() {
  return { chapter: {
    id: 'ch', ui_language: 'ru',
    blocks: [
      { id: 'b1', type: 'theory', theory: { examples: [{ text: 'I work.', translation: 'Я работаю.' }] } },
      { id: 'quiz', type: 'quiz_inline', quiz_inline: { question_ids: ['translated', 'missing'] } },
    ],
    question_bank: { questions: [
      { id: 'translated', type: 'reorder', theory_block_id: 'b1', correct_answer: 'I work.' },
      { id: 'missing', type: 'reorder', theory_block_id: 'b1', correct_answer: 'I work here.' },
      { id: 'choice', type: 'mcq_single', theory_block_id: 'b2', correct_answer: 'a' },
    ] },
    chapter_test: { num_questions: 10, pool_question_ids: ['translated', 'missing', 'choice'] },
  } }
}

describe('offline grammar reorder delivery', () => {
  beforeEach(() => {
    vi.spyOn(navigator, 'onLine', 'get').mockReturnValue(false)
    setGrammarCourse('en_ru')
    vi.mocked(getOfflineMeta).mockResolvedValue({ sections: [{ section_id: 'sec', chapters: [{ chapter_id: 'ch', passed: true }] }] } as any)
    vi.mocked(getStoredChapter).mockResolvedValue(chapterPayload())
    vi.mocked(getStoredChapters).mockResolvedValue([chapterPayload()])
    vi.mocked(getTrainingQuestions).mockResolvedValue([])
  })
  afterEach(() => vi.restoreAllMocks())

  it('filters old cached chapter and category pools before selection and counts only available questions', async () => {
    for (const result of [await grammarClient.getChapterTest('ch'), await grammarClient.getCategoryTest('sec')]) {
      expect(result.total).toBe(2)
      expect(result.questions.map(q => q.id).sort()).toEqual(['choice', 'translated'])
      expect(result.questions.find(q => q.id === 'translated')).toMatchObject({ translation_ru: 'Я работаю.', correct_answer: 'I work.' })
      expect(result.questions.find(q => q.id === 'choice')).not.toHaveProperty('correct_answer')
    }
    const payload = await grammarClient.getChapter('ch')
    expect(payload.chapter.blocks[1].quiz_inline.question_ids).toEqual(['translated'])
  })

  it('retains word tiles and translation in fallback training', async () => {
    const session = await grammarClient.startTrainingSession(10)
    expect(session.items).toHaveLength(2)
    expect(session.items.find((item: any) => item.question.type === 'reorder').question).toMatchObject({
      translation_ru: 'Я работаю.', correct_answer: 'I work.',
    })
  })

  it('excludes untranslated questions from a previously cached training pack', async () => {
    vi.mocked(getTrainingQuestions).mockResolvedValue([{ id: 'old', chapter_id: 'ch', type: 'reorder', correct_answer: 'I work.' }])
    expect((await grammarClient.startTrainingSession()).items).toEqual([])
    expect((await grammarClient.getTrainingAvailability()).grammar_training.available).toBe(false)
  })
})
