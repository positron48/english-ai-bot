import { describe, expect, it } from 'vitest'
import type { ConversationScenarioSummary } from '../api/courseClient'
import {
  buildNpcGroups,
  scenarioQuestPassed,
  scenarioQuestPerfect,
} from './conversations'

function scenario(
  overrides: Partial<ConversationScenarioSummary> = {},
): ConversationScenarioSummary {
  return {
    code: 'q1',
    title: 'Quest',
    npc_name: 'Mara',
    npc_code: 'mara_barista',
    place_type: 'cafe',
    cefr_level: 'A1',
    is_quest: true,
    prerequisite_code: '',
    image_url: '',
    npc_image_url: '/img/mara.png',
    cooldown_until: null,
    locked: false,
    tasks: [],
    session_status: 'none',
    quest_passed: false,
    all_tasks_done: false,
    ...overrides,
  }
}

describe('scenarioQuestPassed', () => {
  it('returns true when quest_passed flag is set', () => {
    expect(scenarioQuestPassed(scenario({ quest_passed: true }))).toBe(true)
  })

  it('returns true for passed or completed session status', () => {
    expect(scenarioQuestPassed(scenario({ session_status: 'passed' }))).toBe(true)
    expect(scenarioQuestPassed(scenario({ session_status: 'completed' }))).toBe(true)
  })

  it('returns false when quest is not passed', () => {
    expect(scenarioQuestPassed(scenario())).toBe(false)
  })
})

describe('scenarioQuestPerfect', () => {
  it('returns true when all tasks are done', () => {
    expect(scenarioQuestPerfect(scenario({ all_tasks_done: true }))).toBe(true)
  })

  it('returns true for completed session status', () => {
    expect(scenarioQuestPerfect(scenario({ session_status: 'completed' }))).toBe(true)
  })

  it('returns false when only mandatory tasks are done', () => {
    expect(scenarioQuestPerfect(scenario({ session_status: 'passed' }))).toBe(false)
  })
})

describe('buildNpcGroups', () => {
  it('groups quests by npc and computes progress', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', quest_passed: true, all_tasks_done: true }),
      scenario({ code: 'q2', session_status: 'passed' }),
    ], 'en_ru')

    expect(groups).toHaveLength(1)
    const g = groups[0]
    expect(g.npcRole).toBe('barista')
    expect(g.npcImageUrl).toBe('/img/mara.png')
    expect(g.questTotal).toBe(2)
    expect(g.completedCount).toBe(2)
    expect(g.pct).toBe(50)
    expect(g.allPassed).toBe(true)
    expect(g.allDone).toBe(false)
    expect(g.hasIncompleteQuests).toBe(false)
    expect(g.visibleQuestScenarios).toHaveLength(0)
    expect(g.hasAvailableIncompleteQuests).toBe(false)
    expect(g.expandable).toBe(false)
    expect(g.locked).toBe(false)
  })

  it('shows only the next available unfinished quest in a chain', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', quest_passed: true }),
      scenario({ code: 'q2', title: 'Visible next quest' }),
      scenario({ code: 'q3', title: 'Future quest' }),
    ])

    expect(groups[0].visibleQuestScenarios.map(s => s.code)).toEqual(['q2'])
    expect(groups[0].hasAvailableIncompleteQuests).toBe(true)
    expect(groups[0].expandable).toBe(true)
  })

  it('does not mark cooldown quests as newly available', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', quest_passed: true }),
      scenario({ code: 'q2', locked: true, cooldown_until: '2026-07-01T12:00:00Z' }),
    ])

    expect(groups[0].visibleQuestScenarios).toHaveLength(0)
    expect(groups[0].hasAvailableIncompleteQuests).toBe(false)
    expect(groups[0].cooldownUntil).toBe('2026-07-01T12:00:00Z')
  })

  it('enables free chat when all quests passed and free scenario is unlocked', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', quest_passed: true }),
      scenario({
        code: 'free',
        is_quest: false,
        locked: false,
        npc_code: 'mara_barista',
      }),
    ])

    expect(groups[0].freeChatAvailable).toBe(true)
    expect(groups[0].freeScenario?.code).toBe('free')
  })

  it('marks group locked when every quest and free chat are locked', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', locked: true }),
      scenario({ code: 'free', is_quest: false, locked: true, npc_code: 'mara_barista' }),
    ])

    expect(groups[0].locked).toBe(true)
    expect(groups[0].freeChatAvailable).toBe(false)
  })

  it('uses Spanish role labels for es_ru course', () => {
    const groups = buildNpcGroups([scenario()], 'es_ru')
    expect(groups[0].npcRole).toBe('barista de cafetería')
  })

  it('creates solo groups for scenarios without npc_code', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'solo', npc_code: '', is_quest: false }),
    ])
    expect(groups[0].key).toBe('_solo_solo')
  })

  it('surfaces cooldown from next locked quest', () => {
    const groups = buildNpcGroups([
      scenario({ code: 'q1', locked: true, cooldown_until: '2026-07-01T12:00:00Z' }),
    ])
    expect(groups[0].cooldownUntil).toBe('2026-07-01T12:00:00Z')
  })
})
