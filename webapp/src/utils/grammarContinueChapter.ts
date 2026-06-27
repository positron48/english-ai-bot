export interface GrammarContinueChapter {
  id: string
  title: string
  url: string
}

export interface GrammarContinueChapterAPI {
  chapter_id: string
  title: string
  title_translations?: Record<string, string>
  url: string
}

export function grammarContinueStorageKey(courseCode?: string): string {
  const code = (courseCode || '').trim()
  return code ? `linglow:last-grammar-chapter:${code}` : 'linglow:last-grammar-chapter'
}

export function getLocalizedGrammarTitle(
  title: string,
  titleTranslations: Record<string, string> | undefined,
  locale: string,
): string {
  const base = (title || '').trim()
  if (locale && locale !== 'en' && titleTranslations?.[locale]) {
    return titleTranslations[locale]
  }
  return base
}

export function toGrammarContinueChapter(
  apiChapter: GrammarContinueChapterAPI,
  locale: string,
): GrammarContinueChapter | null {
  const id = (apiChapter.chapter_id || '').trim()
  const url = (apiChapter.url || '').trim()
  const title = getLocalizedGrammarTitle(apiChapter.title, apiChapter.title_translations, locale)
  if (!id || !url.startsWith('/learning/grammar/chapter/') || !title) {
    return null
  }
  return { id, title, url }
}

export function readStoredGrammarContinueChapter(courseCode?: string): GrammarContinueChapter | null {
  try {
    const raw = localStorage.getItem(grammarContinueStorageKey(courseCode))
    if (!raw) return null
    const parsed = JSON.parse(raw) as { id?: unknown; title?: unknown; url?: unknown }
    const title = typeof parsed.title === 'string' ? parsed.title.trim() : ''
    const url = typeof parsed.url === 'string' ? parsed.url.trim() : ''
    const id = typeof parsed.id === 'string' ? parsed.id : ''
    return title && url.startsWith('/learning/grammar/chapter/') ? { id, title, url } : null
  } catch {
    return null
  }
}

export function writeStoredGrammarContinueChapter(
  chapter: GrammarContinueChapter,
  courseCode?: string,
): void {
  try {
    localStorage.setItem(grammarContinueStorageKey(courseCode), JSON.stringify(chapter))
  } catch {
    // ignore storage errors
  }
}
