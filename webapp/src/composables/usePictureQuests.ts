import { courseClient } from '../api/courseClient'
import type { PictureQuestSummary } from '../api/courseClient'
import { grammarClient } from '../api/grammarClient'
import { buildGrammarLevelAccess, isDistrictUnlocked } from '../utils/districtUnlock'

export interface PictureQuestLoad {
  isPro: boolean
  quests: PictureQuestSummary[]
}

/**
 * loadPictureQuestsFlat resolves the flat list of picture quests across every unlocked
 * district for the current course. `archive` selects passed quests (true) or the active,
 * not-yet-passed ones (false) — the backend filters so the archive is never loaded with
 * the main list.
 */
export async function loadPictureQuestsFlat(
  currentCourseCode: string | undefined,
  hasPictureFeature: boolean,
  archive: boolean,
): Promise<PictureQuestLoad> {
  if (!hasPictureFeature) return { isPro: false, quests: [] }

  const courseCode = currentCourseCode || undefined
  const [map, grammarData] = await Promise.all([
    courseClient.getCourseMap(courseCode),
    grammarClient.getCategories().catch(() => ({ categories: [] })),
  ])
  const grammarAccess = buildGrammarLevelAccess(grammarData.categories || [])
  const openDistricts = (map.districts || []).filter(d => isDistrictUnlocked(d.level_code, grammarAccess))

  const results = await Promise.all(
    openDistricts.map(d => courseClient.listPictureQuests(d.code, courseCode, archive)),
  )
  const quests = results.flatMap(r => r.quests || [])
  return { isPro: true, quests }
}
