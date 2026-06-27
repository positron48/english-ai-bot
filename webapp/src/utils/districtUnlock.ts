/** Grammar access map keyed by CEFR level (e.g. A0, B1). Mirrors CityMapView unlock logic. */
export type GrammarLevelAccess = Record<string, { canAccess: boolean }>

export function buildGrammarLevelAccess(
  categories: Array<{ level?: string; can_access?: boolean }>,
): GrammarLevelAccess {
  const map: GrammarLevelAccess = {}
  for (const cat of categories) {
    const lv = (cat.level || '').toUpperCase()
    if (!lv) continue
    if (!map[lv]) map[lv] = { canAccess: false }
    if (cat.can_access) map[lv].canAccess = true
  }
  return map
}

/** District is unlocked when the learner has grammar access for its CEFR level. */
export function isDistrictUnlocked(levelCode: string, grammar: GrammarLevelAccess): boolean {
  const lv = levelCode.toUpperCase()
  return !!grammar[lv]?.canAccess
}
