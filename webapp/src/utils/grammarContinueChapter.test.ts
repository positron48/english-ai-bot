import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  getLocalizedGrammarTitle,
  grammarContinueStorageKey,
  readStoredGrammarContinueChapter,
  toGrammarContinueChapter,
  writeStoredGrammarContinueChapter,
} from './grammarContinueChapter'

const validChapter = {
  id: 'ch-1',
  title: 'Present Simple',
  url: '/learning/grammar/chapter/ch-1',
}

describe('grammarContinueStorageKey', () => {
  it('uses course-scoped key when course code is provided', () => {
    expect(grammarContinueStorageKey('es_ru')).toBe('linglow:last-grammar-chapter:es_ru')
  })

  it('falls back to default key without course code', () => {
    expect(grammarContinueStorageKey()).toBe('linglow:last-grammar-chapter')
    expect(grammarContinueStorageKey('  ')).toBe('linglow:last-grammar-chapter')
  })
})

describe('getLocalizedGrammarTitle', () => {
  it('returns translation for non-English locale when available', () => {
    expect(
      getLocalizedGrammarTitle('Hello', { ru: 'Привет' }, 'ru'),
    ).toBe('Привет')
  })

  it('returns base title for English or missing translation', () => {
    expect(getLocalizedGrammarTitle('Hello', { ru: 'Привет' }, 'en')).toBe('Hello')
    expect(getLocalizedGrammarTitle('Hello', undefined, 'ru')).toBe('Hello')
  })
})

describe('toGrammarContinueChapter', () => {
  it('maps valid API chapter to stored shape', () => {
    expect(
      toGrammarContinueChapter(
        {
          chapter_id: ' ch-1 ',
          title: 'Present Simple',
          title_translations: { ru: 'Настоящее простое' },
          url: ' /learning/grammar/chapter/ch-1 ',
        },
        'ru',
      ),
    ).toEqual({
      id: 'ch-1',
      title: 'Настоящее простое',
      url: '/learning/grammar/chapter/ch-1',
    })
  })

  it('returns null for invalid id, url, or title', () => {
    expect(
      toGrammarContinueChapter(
        { chapter_id: '', title: 'T', url: '/learning/grammar/chapter/x' },
        'en',
      ),
    ).toBeNull()
    expect(
      toGrammarContinueChapter(
        { chapter_id: 'x', title: 'T', url: '/wrong/path' },
        'en',
      ),
    ).toBeNull()
    expect(
      toGrammarContinueChapter(
        { chapter_id: 'x', title: '  ', url: '/learning/grammar/chapter/x' },
        'en',
      ),
    ).toBeNull()
  })
})

describe('localStorage persistence', () => {
  afterEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('reads and writes valid chapter JSON', () => {
    writeStoredGrammarContinueChapter(validChapter, 'en_ru')
    expect(readStoredGrammarContinueChapter('en_ru')).toEqual(validChapter)
  })

  it('returns null for missing, invalid JSON, or malformed entries', () => {
    expect(readStoredGrammarContinueChapter()).toBeNull()
    localStorage.setItem(grammarContinueStorageKey(), '{not json')
    expect(readStoredGrammarContinueChapter()).toBeNull()
    localStorage.setItem(
      grammarContinueStorageKey(),
      JSON.stringify({ id: 'x', title: '', url: '/learning/grammar/chapter/x' }),
    )
    expect(readStoredGrammarContinueChapter()).toBeNull()
  })

  it('ignores storage write errors', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota exceeded')
    })
    expect(() => writeStoredGrammarContinueChapter(validChapter)).not.toThrow()
  })
})
