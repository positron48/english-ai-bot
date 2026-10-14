import type { CourseMastery } from '../api/courseClient'

/** Shared city district access, including vocabulary, grammar and placement unlocks. */
export type DistrictLevelAccess = Record<string, { canAccess: boolean }>

export function buildDistrictLevelAccess(mastery?: CourseMastery): DistrictLevelAccess {
  const map: DistrictLevelAccess = {}
  for (const level of mastery?.levels || []) {
    map[level.level_code.toUpperCase()] = { canAccess: level.unlocked }
  }
  return map
}

export function isDistrictUnlocked(levelCode: string, access: DistrictLevelAccess): boolean {
  return !!access[levelCode.toUpperCase()]?.canAccess
}
