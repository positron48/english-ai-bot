import { describe, expect, it } from 'vitest'
import { buildDistrictLevelAccess, isDistrictUnlocked } from './districtUnlock'
import type { CourseMastery } from '../api/courseClient'

describe('district access follows city mastery', () => {
  it('opens B1 through vocabulary even with zero grammar progress', () => {
    const mastery: CourseMastery = {
      current_level_code: 'B1', progress_percent: 0,
      levels: [
        {
          level_code: 'B1', unlocked: true, current: true,
          mastery_percent: 0, can_open_next: false,
          metrics: {
            grammar: { done: 0, total: 10, target: 10, percent: 0, included: true },
            words: { done: 1500, total: 2000, target: 2000, percent: 75, included: true },
          },
        },
        {
          level_code: 'B2', unlocked: false, current: false,
          mastery_percent: 0, can_open_next: false, metrics: {},
        },
      ],
    }
    const access = buildDistrictLevelAccess(mastery)
    expect(isDistrictUnlocked('b1', access)).toBe(true)
    expect(isDistrictUnlocked('B2', access)).toBe(false)
    expect(isDistrictUnlocked('C1', access)).toBe(false)
  })

  it('keeps materials locked when mastery is unavailable', () => {
    expect(isDistrictUnlocked('B1', buildDistrictLevelAccess())).toBe(false)
  })
})
