import type { CourseMastery, CourseMasteryLevel, CourseMasteryMetric } from '../api/courseClient'

export type ActivityStatus = 'gray' | 'orange' | 'yellow' | 'green'

/** Map a mastery metric percent to the city-map activity icon colour. */
export function metricPercentToStatus(pct: number): ActivityStatus {
  if (pct <= 0) return 'gray'
  if (pct < 34) return 'orange'
  if (pct < 67) return 'yellow'
  return 'green'
}

export function formatLevelPath(current?: string, next?: string): string {
  const cur = (current || 'A0').toUpperCase()
  const nxt = (next || '').toUpperCase()
  return nxt ? `${cur} → ${nxt}` : cur
}

export function masteryLevelByCode(mastery: CourseMastery | null | undefined, levelCode: string): CourseMasteryLevel | null {
  if (!mastery?.levels?.length) return null
  const code = (levelCode || '').toUpperCase()
  return mastery.levels.find(l => (l.level_code || '').toUpperCase() === code) || null
}

export function metricForLevel(
  mastery: CourseMastery | null | undefined,
  levelCode: string,
  key: keyof CourseMasteryLevel['metrics'] | string,
): CourseMasteryMetric | null {
  const lv = masteryLevelByCode(mastery, levelCode)
  if (!lv?.metrics) return null
  return lv.metrics[key] ?? null
}

export function districtMapLevel(mastery: CourseMastery | null | undefined, levelCode: string): number {
  const lv = masteryLevelByCode(mastery, levelCode)
  if (!lv || !lv.unlocked) return 1
  const pct = lv.mastery_percent || 0
  if (lv.can_open_next || pct >= 75) return 5
  if (pct >= 40) return 4
  if (pct > 0) return 3
  return 2
}

export function districtStatusLabel(
  lv: CourseMasteryLevel | null,
  t: (key: string) => string,
): { pct: number; status: string; fill: string; locked: boolean } {
  const fills = ['#3F6F3F', '#7FAE6A', '#D9A83F', '#E3D8C6']
  if (!lv || !lv.unlocked) {
    return { pct: 0, status: '', fill: fills[3], locked: true }
  }
  const pct = Math.round(lv.mastery_percent || 0)
  let status = ''
  let fill = fills[3]
  if (pct >= 75) { status = t('progress.distExcellent'); fill = fills[0] }
  else if (pct >= 40) { status = t('progress.distGood'); fill = fills[1] }
  else if (pct >= 10) { status = t('progress.distInProgress'); fill = fills[2] }
  else if (pct > 0) { status = t('progress.distJustStarted') }
  return { pct, status, fill, locked: false }
}
