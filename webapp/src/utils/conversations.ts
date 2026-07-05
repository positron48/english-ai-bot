import type { ConversationScenarioSummary } from '../api/courseClient'

/** Mandatory tasks done (quest passed). */
export function scenarioQuestPassed(s: ConversationScenarioSummary): boolean {
  if (s.quest_passed) return true
  return s.session_status === 'passed' || s.session_status === 'completed'
}

/** All tasks done including optional (100% / golden star). */
export function scenarioQuestPerfect(s: ConversationScenarioSummary): boolean {
  if (s.all_tasks_done) return true
  return s.session_status === 'completed'
}

export interface NpcGroup {
  key: string
  npcName: string
  npcRole: string
  placeType: string
  level: string
  npcImageUrl: string
  questScenarios: ConversationScenarioSummary[]
  visibleQuestScenarios: ConversationScenarioSummary[]
  freeScenario: ConversationScenarioSummary | null
  freeChatAvailable: boolean
  cooldownUntil: string | null
  questTotal: number
  completedCount: number
  pct: number
  allDone: boolean
  allPassed: boolean
  hasCompletedQuests: boolean
  hasIncompleteQuests: boolean
  hasAvailableIncompleteQuests: boolean
  locked: boolean
  expandable: boolean
}

// Group scenarios by NPC (npc_code). Quests form the ordered chain; a single is_quest=false
// scenario (if any) is offered as a separate "free chat". Standalone scenarios (no npc_code)
// each form their own group. Backend already orders scenarios by sort_order, so chain order is preserved.
export function buildNpcGroups(scenarios: ConversationScenarioSummary[], courseCode = ''): NpcGroup[] {
  const groups: NpcGroup[] = []
  const byKey = new Map<string, NpcGroup>()
  for (const s of scenarios) {
    const key = s.npc_code || `_solo_${s.code}`
    let g = byKey.get(key)
    if (!g) {
      g = {
        key,
        npcName: s.npc_name,
        npcRole: npcRoleLabel(s, courseCode),
        placeType: s.place_type,
        level: s.cefr_level,
        npcImageUrl: '',
        questScenarios: [],
        visibleQuestScenarios: [],
        freeScenario: null,
        freeChatAvailable: false,
        cooldownUntil: null,
        questTotal: 0,
        completedCount: 0,
        pct: 0,
        allDone: false,
        allPassed: false,
        hasCompletedQuests: false,
        hasIncompleteQuests: false,
        hasAvailableIncompleteQuests: false,
        locked: false,
        expandable: false,
      }
      byKey.set(key, g)
      groups.push(g)
    }
    // Pick up NPC image from any scenario in the group.
    if (!g.npcImageUrl && s.npc_image_url) g.npcImageUrl = s.npc_image_url
    if (s.is_quest) g.questScenarios.push(s)
    else if (!g.freeScenario) g.freeScenario = s
  }

  for (const g of groups) {
    g.questTotal = g.questScenarios.length
    const passedCount = g.questScenarios.filter(scenarioQuestPassed).length
    const perfectCount = g.questScenarios.filter(scenarioQuestPerfect).length
    g.completedCount = passedCount
    g.pct = g.questTotal > 0 ? Math.round((perfectCount / g.questTotal) * 100) : 0
    g.allDone = g.questTotal > 0 && perfectCount === g.questTotal
    g.allPassed = g.questTotal > 0 && passedCount === g.questTotal
    g.hasCompletedQuests = passedCount > 0
    g.hasIncompleteQuests = g.questScenarios.some(s => !scenarioQuestPassed(s))
    g.visibleQuestScenarios = g.questScenarios.filter(s => !s.locked && !scenarioQuestPassed(s)).slice(0, 1)
    g.hasAvailableIncompleteQuests = g.visibleQuestScenarios.length > 0
    // Locked only when there is nothing to start: every quest is locked and free chat is absent
    // or still gated by a prerequisite.
    g.locked = (!g.freeScenario || g.freeScenario.locked) && g.questScenarios.every(s => s.locked)
    // Free chat is only available once all quests in the chain are passed (mandatory tasks done).
    g.freeChatAvailable = g.allPassed && !!g.freeScenario && !g.freeScenario.locked
    // Cooldown: surface the unlock time of the next locked-but-cooldown quest so the UI can show a timer.
    const nextCooldown = g.questScenarios.find(s => s.locked && s.cooldown_until)
    g.cooldownUntil = nextCooldown?.cooldown_until ?? null
    g.expandable = g.questTotal > 1 && g.visibleQuestScenarios.length > 0
  }
  return groups
}

const NPC_ROLE_LABELS: Record<string, Record<'en' | 'es', string>> = {
  mara_barista: { en: 'barista', es: 'barista de cafetería' },
  sam_shop: { en: 'shop assistant', es: 'dependiente' },
  park_police: { en: 'police officer', es: 'policía' },
}

const PLACE_ROLE_LABELS: Record<string, Record<'en' | 'es', string>> = {
  cafe: { en: 'barista', es: 'barista de cafetería' },
  shop: { en: 'shop assistant', es: 'dependiente' },
  police_station: { en: 'police officer', es: 'policía' },
  pharmacy: { en: 'pharmacist', es: 'farmacéutico' },
  hotel: { en: 'receptionist', es: 'recepcionista' },
  restaurant: { en: 'waiter', es: 'camarero' },
  market: { en: 'market seller', es: 'vendedor de mercado' },
}

function npcRoleLabel(s: ConversationScenarioSummary, courseCode: string): string {
  const lang = courseCode === 'es_ru' ? 'es' : 'en'
  return NPC_ROLE_LABELS[s.npc_code]?.[lang] || PLACE_ROLE_LABELS[s.place_type]?.[lang] || ''
}
