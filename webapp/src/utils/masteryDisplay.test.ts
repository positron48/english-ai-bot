import { describe, expect, it } from 'vitest'
import type { CourseMasteryLevel } from '../api/courseClient'
import { districtCoreMetricsComplete, districtMapLevel } from './masteryDisplay'

function level(partial: Partial<CourseMasteryLevel> & Pick<CourseMasteryLevel, 'metrics'>): CourseMasteryLevel {
  return {
    level_code: 'A1',
    unlocked: true,
    current: false,
    mastery_percent: 0,
    can_open_next: false,
    ...partial,
  }
}

function metric(percent: number, included = true) {
  return { done: 0, total: 0, target: 0, percent, included }
}

describe('districtCoreMetricsComplete', () => {
  it('requires grammar, words and reading at 100% when all are included', () => {
    expect(districtCoreMetricsComplete(level({
      metrics: {
        grammar: metric(100),
        words: metric(100),
        reading: metric(100),
      },
    }))).toBe(true)

    expect(districtCoreMetricsComplete(level({
      metrics: {
        grammar: metric(100),
        words: metric(100),
        reading: metric(67),
      },
    }))).toBe(false)
  })

  it('ignores reading when it is not part of the level target', () => {
    expect(districtCoreMetricsComplete(level({
      metrics: {
        grammar: metric(100),
        words: metric(100),
        reading: metric(0, false),
      },
    }))).toBe(true)
  })
})

describe('districtMapLevel', () => {
  it('uses tier 5 only when core metrics hit their level targets', () => {
    const mastery = {
      current_level_code: 'A1',
      progress_percent: 80,
      levels: [level({
        level_code: 'A1',
        mastery_percent: 82,
        can_open_next: true,
        metrics: {
          grammar: metric(100),
          words: metric(100),
          reading: metric(80),
        },
      })],
    }

    expect(districtMapLevel(mastery, 'A1')).toBe(4)
  })

  it('reaches tier 5 when grammar, words and reading are all at 100%', () => {
    const mastery = {
      current_level_code: 'A1',
      progress_percent: 100,
      levels: [level({
        level_code: 'A1',
        mastery_percent: 100,
        can_open_next: true,
        metrics: {
          grammar: metric(100),
          words: metric(100),
          reading: metric(100),
        },
      })],
    }

    expect(districtMapLevel(mastery, 'A1')).toBe(5)
  })
})
