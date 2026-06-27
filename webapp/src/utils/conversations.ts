import type { ConversationScenarioSummary } from '../api/courseClient'

export interface NpcGroup {
  key: string
  npcName: string
  npcRole: string
  placeType: string
  level: string
  questScenarios: ConversationScenarioSummary[]
  freeScenario: ConversationScenarioSummary | null
  questTotal: number
  completedCount: number
  pct: number
  allDone: boolean
  hasCompletedQuests: boolean
  hasIncompleteQuests: boolean
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
        questScenarios: [],
        freeScenario: null,
        questTotal: 0,
        completedCount: 0,
        pct: 0,
        allDone: false,
        hasCompletedQuests: false,
        hasIncompleteQuests: false,
        locked: false,
        expandable: false,
      }
      byKey.set(key, g)
      groups.push(g)
    }
    if (s.is_quest) g.questScenarios.push(s)
    else if (!g.freeScenario) g.freeScenario = s
  }

  for (const g of groups) {
    g.questTotal = g.questScenarios.length
    g.completedCount = g.questScenarios.filter(s => s.session_status === 'completed').length
    g.pct = g.questTotal > 0 ? Math.round((g.completedCount / g.questTotal) * 100) : 0
    g.allDone = g.questTotal > 0 && g.completedCount === g.questTotal
    g.hasCompletedQuests = g.completedCount > 0
    g.hasIncompleteQuests = g.questScenarios.some(s => s.session_status !== 'completed')
    // Locked only when there is nothing to start: every quest is locked and free chat is absent
    // or still gated by a prerequisite.
    g.locked = (!g.freeScenario || g.freeScenario.locked) && g.questScenarios.every(s => s.locked)
    g.expandable = (g.questTotal + (g.freeScenario ? 1 : 0)) > 1
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
