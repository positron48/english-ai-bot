import { describe, expect, it } from 'vitest'
import { buildGrammarLevelAccess, isDistrictUnlocked } from './districtUnlock'

describe('buildGrammarLevelAccess', () => {
  it('skips categories without level', () => {
    expect(buildGrammarLevelAccess([{ can_access: true }])).toEqual({})
  })

  it('normalizes level keys and merges can_access across categories', () => {
    const map = buildGrammarLevelAccess([
      { level: 'a1', can_access: false },
      { level: 'A1', can_access: true },
      { level: 'b1' },
    ])

    expect(map).toEqual({
      A1: { canAccess: true },
      B1: { canAccess: false },
    })
  })
})

describe('isDistrictUnlocked', () => {
  it('returns true when grammar access exists for the level', () => {
    const grammar = buildGrammarLevelAccess([{ level: 'A2', can_access: true }])
    expect(isDistrictUnlocked('a2', grammar)).toBe(true)
  })

  it('returns false when level is missing or not accessible', () => {
    const grammar = buildGrammarLevelAccess([{ level: 'A2', can_access: false }])
    expect(isDistrictUnlocked('B1', grammar)).toBe(false)
    expect(isDistrictUnlocked('A2', grammar)).toBe(false)
  })
})
