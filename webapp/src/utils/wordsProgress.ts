// Single source of truth for the "words" progress percentage shown on the city
// map and on the district page. Both compute identically from per-CEFR-level word
// coverage (mastered vs. total words for that level), served by /api/linglow/word-progress,
// which reads the legacy vocab (the same data the word-set progress bars use).

export interface WordLevelProgress {
  total: number
  mastered: number
}

export type WordLevelProgressMap = Record<string, WordLevelProgress>

// Percentage of mastered words for a district's CEFR level, 0..100.
export function wordsPercentForLevel(
  levels: WordLevelProgressMap | null | undefined,
  cefrLevel: string,
): number {
  const lvl = levels?.[(cefrLevel || '').toUpperCase()]
  if (!lvl || lvl.total <= 0) return 0
  return Math.min(100, Math.round((lvl.mastered / lvl.total) * 100))
}
