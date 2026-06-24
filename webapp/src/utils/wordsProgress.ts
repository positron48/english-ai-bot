// Single source of truth for the "words" progress percentage shown on the city
// map and on the district page. Both must compute identically: the number of
// mastered words in a district, relative to the number of words required for
// that district's CEFR level.

import type { CourseProgressLocation } from '../api/courseClient'

// Words required to complete each CEFR level (the per-level norm).
export const WORDS_NORM_BY_LEVEL: Record<string, number> = {
  A0: 150,
  A1: 350,
  A2: 700,
  B1: 1300,
  B2: 2500,
  C1: 5000,
}

// Sum of mastered words across a district's word_market locations.
export function masteredWordsInDistrict(
  byLocation: CourseProgressLocation[],
  districtCode: string,
): number {
  return (byLocation || [])
    .filter(l => l.district_code === districtCode && l.location_type === 'word_market')
    .reduce((s, l) => s + (l.mastered_items || 0), 0)
}

// Percentage of mastered words against the level norm, clamped to 0..100.
export function wordsPercentForDistrict(
  byLocation: CourseProgressLocation[],
  districtCode: string,
  cefrLevel: string,
): number {
  const norm = WORDS_NORM_BY_LEVEL[(cefrLevel || '').toUpperCase()]
  if (!norm) return 0
  const mastered = masteredWordsInDistrict(byLocation, districtCode)
  return Math.min(100, Math.round((mastered / norm) * 100))
}
