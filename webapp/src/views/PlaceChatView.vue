<template>
  <div class="chat-page" :class="{ 'chat-page--shell': !!scenarioCode }">
    <LgPageHeader
      :title="headerTitle"
      :show-back="true"
      @back="goBack"
    />

    <!-- NPC LIST -->
    <template v-if="!scenarioCode">
      <div v-if="loading" class="chat-loading">{{ t('common.loading') }}</div>
      <div v-else-if="!npcGroups.length" class="chat-loading">{{ t('chat.noPlaces') }}</div>
      <div v-else class="chat-list">
        <div v-for="g in npcGroups" :key="g.key" class="npc-block">
          <button
            class="chat-card"
            :class="{ 'chat-card--locked': g.locked }"
            type="button"
            @click="onNpcClick(g)"
          >
            <LgActivityIcon
              type="conversation"
              :status="g.allDone ? 'green' : (g.locked ? 'gray' : 'orange')"
              :size="22"
            />
            <div class="chat-card-body">
              <div class="chat-card-title">{{ g.npcName }}</div>
              <div class="chat-card-meta">
                {{ placeLabel(g.placeType) }} · {{ g.level }}
                <span class="chat-tag">{{ g.completedCount }}/{{ g.total }}</span>
                <span v-if="g.locked" class="chat-tag chat-tag--locked">🔒 {{ t('chat.locked') }}</span>
              </div>
              <div v-if="g.total > 1" class="npc-bar-track">
                <div class="npc-bar-fill" :style="{ width: g.pct + '%' }" />
              </div>
            </div>
            <span v-if="g.allDone" class="chat-done">✓</span>
            <svg
              v-else-if="g.total > 1"
              class="npc-chevron"
              :class="{ 'npc-chevron--open': expandedNpc === g.key }"
              width="16" height="16" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
            >
              <path d="M6 9l6 6 6-6" />
            </svg>
          </button>

          <!-- chain steps -->
          <div v-if="g.total > 1 && expandedNpc === g.key" class="npc-steps">
            <button
              v-for="(s, idx) in g.scenarios"
              :key="s.code"
              class="npc-step"
              :class="{ 'npc-step--locked': s.locked, 'npc-step--done': s.session_status === 'completed' }"
              type="button"
              :disabled="s.locked"
              @click="openScenario(s.code)"
            >
              <span class="npc-step-num">{{ idx + 1 }}</span>
              <span class="npc-step-body">
                <span class="npc-step-title">{{ s.title }}</span>
                <span class="npc-step-meta">
                  <span v-if="s.is_quest" class="chat-tag">{{ t('chat.quest') }}</span>
                  <span v-else class="chat-tag chat-tag--free">{{ t('chat.free') }}</span>
                </span>
              </span>
              <span class="npc-step-state">
                <template v-if="s.session_status === 'completed'">✓</template>
                <template v-else-if="s.locked">🔒</template>
                <template v-else>›</template>
              </span>
            </button>
          </div>
        </div>
      </div>
    </template>

    <!-- CHAT -->
    <template v-else>
      <div v-if="loading" class="chat-loading">{{ t('common.loading') }}</div>
      <template v-else-if="session">
        <!-- task checklist -->
        <div v-if="session.is_quest && tasks.length" class="chat-tasks">
          <div class="chat-tasks-title">{{ t('chat.tasksTitle') }}</div>
          <div
            v-for="task in tasks"
            :key="task.code"
            class="chat-task"
            :class="{ 'chat-task--done': task.completed, 'chat-task--optional': !task.required }"
          >
            <span class="chat-task-check">{{ task.completed ? '✓' : '○' }}</span>
            <span class="chat-task-label">{{ task.title }}</span>
            <span v-if="!task.required" class="chat-task-opt">{{ t('chat.optional') }}</span>
          </div>
        </div>

        <!-- messages -->
        <div ref="scrollEl" class="chat-thread">
          <p v-if="session.scene_setup" class="chat-scene">{{ session.scene_setup }}</p>
          <template v-for="(m, i) in messages" :key="i">
            <div
              class="chat-row"
              :class="m.role === 'user' ? 'chat-row--user' : 'chat-row--npc'"
            >
              <LgSpeechBubble v-if="m.role !== 'user'" :text="m.content" />
              <div v-else class="chat-user-bubble">{{ m.content }}</div>
            </div>
            <div v-if="m.role !== 'user' && m.corrections && m.corrections.length" class="chat-corrections">
              <div class="chat-corrections-title">{{ t('chat.correctionsTitle') }}</div>
              <div v-for="(c, ci) in m.corrections" :key="ci" class="chat-correction">
                <div class="chat-correction-line">
                  <span class="chat-correction-bad">{{ c.original }}</span>
                  <span class="chat-correction-arrow">→</span>
                  <span class="chat-correction-good">{{ c.corrected }}</span>
                </div>
                <div v-if="c.explanation" class="chat-correction-expl">{{ c.explanation }}</div>
              </div>
            </div>
          </template>
          <div v-if="sending" class="chat-row chat-row--npc">
            <div class="chat-typing">…</div>
          </div>
        </div>

        <!-- completion / budget banners -->
        <div v-if="questPassed" class="chat-banner chat-banner--win">
          <LgLumi pose="clapping" :size="48" />
          <div class="chat-banner-text">{{ t('chat.questComplete') }}</div>
          <div v-if="status === 'open'" class="chat-banner-hint">{{ t('chat.questCompleteHint') }}</div>
          <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
        </div>
        <div v-else-if="budgetExhausted" class="chat-banner">
          <div class="chat-banner-text">{{ t('chat.budgetEnded') }}</div>
          <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
        </div>

        <!-- composer -->
        <div v-if="canSend" class="chat-composer">
          <input
            ref="composerInput"
            v-model="input"
            class="chat-input"
            type="text"
            :placeholder="t('chat.placeholder')"
            @keyup.enter="send"
          />
          <button class="chat-send" type="button" :disabled="sending || !input.trim()" @click="send">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z" />
            </svg>
          </button>
        </div>
      </template>
      <div v-else class="chat-loading">{{ loadError || t('chat.notAvailable') }}</div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  courseClient,
  type ConversationScenarioSummary,
  type ConversationSessionState,
  type ConversationMessage,
  type ConversationTask,
} from '../api/courseClient'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgSpeechBubble from '../components/linglow/LgSpeechBubble.vue'
import LgButton from '../components/linglow/LgButton.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import { useCourse } from '../composables/useCourse'

const { t } = useI18n()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const route = useRoute()
const router = useRouter()

const districtCode = computed(() => String(route.params.districtCode || ''))
const scenarioCode = computed(() => String(route.params.scenarioCode || ''))

const loading = ref(true)
const scenarios = ref<ConversationScenarioSummary[]>([])
const expandedNpc = ref('')

interface NpcGroup {
  key: string
  npcName: string
  placeType: string
  level: string
  scenarios: ConversationScenarioSummary[]
  total: number
  completedCount: number
  pct: number
  allDone: boolean
  locked: boolean
}

// Group scenarios by NPC (npc_code). Standalone scenarios (no npc_code) each form their own group.
// Backend already returns scenarios ordered by sort_order, so chain order is preserved.
const npcGroups = computed<NpcGroup[]>(() => {
  const groups: NpcGroup[] = []
  const byKey = new Map<string, NpcGroup>()
  for (const s of scenarios.value) {
    const key = s.npc_code || `_solo_${s.code}`
    let g = byKey.get(key)
    if (!g) {
      g = {
        key, npcName: s.npc_name, placeType: s.place_type, level: s.cefr_level,
        scenarios: [], total: 0, completedCount: 0, pct: 0, allDone: false, locked: false,
      }
      byKey.set(key, g)
      groups.push(g)
    }
    g.scenarios.push(s)
  }
  for (const g of groups) {
    g.total = g.scenarios.length
    g.completedCount = g.scenarios.filter(s => s.session_status === 'completed').length
    g.pct = g.total > 0 ? Math.round((g.completedCount / g.total) * 100) : 0
    g.allDone = g.total > 0 && g.completedCount === g.total
    // The NPC is "locked" only when every one of its scenarios is locked (nothing to start yet).
    g.locked = g.scenarios.every(s => s.locked)
  }
  return groups
})

const PLACE_LABELS: Record<string, string> = {
  cafe: 'Кафе', shop: 'Магазин', police_station: 'Полиция',
  pharmacy: 'Аптека', hotel: 'Отель', restaurant: 'Ресторан', market: 'Рынок',
}
function placeLabel(placeType: string): string {
  return PLACE_LABELS[placeType] || placeType
}

function onNpcClick(g: NpcGroup) {
  if (g.locked) return
  if (g.total === 1) {
    openScenario(g.scenarios[0].code)
    return
  }
  // Multi-scenario NPC: toggle the chain steps.
  expandedNpc.value = expandedNpc.value === g.key ? '' : g.key
}

const session = ref<ConversationSessionState | null>(null)
const messages = ref<ConversationMessage[]>([])
const tasks = ref<ConversationTask[]>([])
const input = ref('')
const sending = ref(false)
const questPassed = ref(false)
const budgetExhausted = ref(false)
const status = ref('open')
const scrollEl = ref<HTMLElement | null>(null)
const composerInput = ref<HTMLInputElement | null>(null)
const loadError = ref('')

function focusComposer() {
  // Keep the cursor in the input after sending so the learner can keep typing without re-tapping.
  nextTick(() => composerInput.value?.focus())
}

const headerTitle = computed(() => {
  if (scenarioCode.value && session.value) return session.value.title
  return t('chat.npcTitle')
})
// Keep the composer usable after the quest passes so the learner can say goodbye / finish
// optional tasks; the session only truly ends when status flips to completed or budget runs out.
const canSend = computed(() => status.value === 'open' && !budgetExhausted.value)

function goBack() {
  router.push({ name: 'CityDistrict', params: { districtCode: districtCode.value } })
}

function openScenario(code: string) {
  router.push({ name: 'PlaceChat', params: { districtCode: districtCode.value, scenarioCode: code } })
}

async function scrollToBottom() {
  await nextTick()
  if (scrollEl.value) scrollEl.value.scrollTop = scrollEl.value.scrollHeight
}

async function send() {
  const text = input.value.trim()
  if (!text || sending.value || !session.value) return
  input.value = ''
  messages.value.push({ role: 'user', content: text })
  sending.value = true
  await scrollToBottom()
  try {
    const res = await courseClient.postConversationMessage(session.value.session_id, text)
    if (res.reply) messages.value.push({ role: 'assistant', content: res.reply, corrections: res.corrections })
    tasks.value = res.tasks
    status.value = res.status
    questPassed.value = res.quest_passed
    budgetExhausted.value = res.budget_exhausted
  } catch {
    messages.value.push({ role: 'assistant', content: t('chat.errorSend') })
  } finally {
    sending.value = false
    await scrollToBottom()
    if (canSend.value) focusComposer()
  }
}

// loadForRoute (re)loads the list or the active session for the current route. It must run on
// every scenarioCode change, not just onMounted: PublicLayout keys <router-view> by course code,
// so navigating list -> scenario reuses this same component instance and onMounted does not re-fire.
async function loadForRoute() {
  loading.value = true
  loadError.value = ''
  session.value = null
  messages.value = []
  tasks.value = []
  questPassed.value = false
  budgetExhausted.value = false
  status.value = 'open'
  try {
    await ensureCourseLoaded()
    const course = currentCourseCode.value || undefined
    if (scenarioCode.value) {
      const s = await courseClient.startConversationSession(scenarioCode.value, course)
      session.value = s
      messages.value = [...(s.messages || [])]
      tasks.value = s.tasks || []
      status.value = s.status
      questPassed.value = s.quest_passed
      await scrollToBottom()
    } else {
      const r = await courseClient.listConversationScenarios(districtCode.value, course)
      scenarios.value = r.scenarios || []
    }
  } catch (e: any) {
    // Surface the real reason instead of a blanket "not available".
    const msg = String(e?.message || '')
    if (msg.includes('403')) loadError.value = t('chat.requiresPro')
    else if (msg.includes('404')) loadError.value = t('chat.scenarioMissing')
    else loadError.value = t('chat.notAvailable')
    console.error('Failed to start conversation:', e)
  } finally {
    loading.value = false
  }
}

watch(scenarioCode, () => { loadForRoute() })
onMounted(loadForRoute)
</script>

<style scoped>
.chat-page { display: flex; flex-direction: column; min-height: 100%; padding-bottom: 16px; }

/* Chat (scenario) mode: a fixed app-shell pinned between the status bar and the bottom nav so
   only the message thread scrolls and the composer stays put at the bottom of the screen. */
.chat-page--shell {
  position: fixed;
  left: 0;
  right: 0;
  top: max(env(safe-area-inset-top, 0px), var(--android-inset-top, 0px));
  bottom: calc(60px + max(env(safe-area-inset-bottom, 0px), var(--android-inset-bottom, 0px)));
  min-height: 0;
  padding-bottom: 0;
  background: var(--bg);
  z-index: 50;
}
@media (min-width: 900px) {
  /* Desktop has a 220px side nav and no bottom nav; align the shell with the centered content column. */
  .chat-page--shell {
    left: 220px;
    top: 0;
    bottom: 0;
    max-width: 880px;
    margin: 0 auto;
  }
}
.chat-loading { padding: 40px 16px; text-align: center; color: var(--subtext); }

/* scenario list */
.chat-list { padding: 8px 16px; display: flex; flex-direction: column; gap: 10px; }
.chat-card {
  display: flex; align-items: center; gap: 12px; width: 100%;
  padding: 14px; border-radius: 14px; border: 1px solid var(--border, rgba(0,0,0,0.08));
  background: var(--card-bg); cursor: pointer; text-align: left;
}
.chat-card-body { flex: 1; min-width: 0; }
.chat-card-title { font-family: 'Inter', sans-serif; font-size: 15px; font-weight: 600; color: var(--text); }
.chat-card-meta { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); margin-top: 2px; }
.chat-tag {
  display: inline-block; margin-left: 6px; padding: 1px 7px; border-radius: 10px;
  font-size: 10px; font-weight: 600; background: rgba(45,107,58,0.12); color: #2d6b3a;
}
.chat-tag--free { background: rgba(200,168,75,0.18); color: #9a7b1e; }
.chat-tag--locked { background: rgba(120,120,120,0.16); color: var(--subtext); }
.chat-done { color: #2d6b3a; font-weight: 700; font-size: 18px; }

/* NPC groups + chain steps */
.npc-block { display: flex; flex-direction: column; }
.chat-card--locked { opacity: 0.6; cursor: default; }
.npc-chevron { color: var(--subtext); flex-shrink: 0; transition: transform 0.2s; }
.npc-chevron--open { transform: rotate(180deg); }
.npc-bar-track { margin-top: 6px; height: 3px; border-radius: 999px; background: var(--progress-track, rgba(0,0,0,0.08)); overflow: hidden; }
.npc-bar-fill { height: 100%; border-radius: 999px; background: #2d6b3a; transition: width 0.4s ease; }

.npc-steps { display: flex; flex-direction: column; gap: 6px; margin: 6px 0 2px 14px; padding-left: 14px; border-left: 2px solid var(--border, rgba(0,0,0,0.08)); }
.npc-step {
  display: flex; align-items: center; gap: 10px; width: 100%;
  padding: 10px 12px; border-radius: 12px; border: 1px solid var(--border, rgba(0,0,0,0.08));
  background: var(--card-bg); cursor: pointer; text-align: left;
}
.npc-step:disabled { cursor: default; opacity: 0.55; }
.npc-step--done { background: rgba(45,107,58,0.06); border-color: rgba(45,107,58,0.25); }
.npc-step-num {
  width: 22px; height: 22px; flex-shrink: 0; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 700; color: var(--subtext);
  background: var(--chip-bg, rgba(0,0,0,0.05));
}
.npc-step--done .npc-step-num { background: #2d6b3a; color: #fff; }
.npc-step-body { flex: 1; min-width: 0; }
.npc-step-title { display: block; font-family: 'Inter', sans-serif; font-size: 14px; font-weight: 600; color: var(--text); }
.npc-step-meta { margin-top: 2px; }
.npc-step-state { flex-shrink: 0; color: var(--subtext); font-size: 16px; font-weight: 700; }
.npc-step--done .npc-step-state { color: #2d6b3a; }

/* task checklist */
.chat-tasks { margin: 8px 16px; padding: 12px 14px; border-radius: 14px; background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.08)); }
.chat-tasks-title { font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 700; color: var(--subtext); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 8px; }
.chat-task { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-family: 'Inter', sans-serif; font-size: 13px; color: var(--text); transition: opacity 0.25s; }
.chat-task-check { width: 18px; text-align: center; color: var(--subtext); }
.chat-task--done { color: #2d6b3a; }
.chat-task--done .chat-task-check { color: #2d6b3a; }
.chat-task--optional { opacity: 0.85; }
.chat-task-opt { font-size: 10px; color: var(--subtext); }

/* thread */
.chat-thread { flex: 1; min-height: 0; overflow-y: auto; padding: 12px 16px; display: flex; flex-direction: column; gap: 10px; }
/* NPC reply bubbles in chat are wider than the default Lumi speech bubble. */
.chat-thread :deep(.lg-bubble) { max-width: 400px; font-size: 14px; }
.chat-scene { font-family: 'Inter', sans-serif; font-size: 12px; font-style: italic; color: var(--subtext); text-align: center; margin: 0 0 6px; }
.chat-row { display: flex; }
.chat-row--user { justify-content: flex-end; }
.chat-row--npc { justify-content: flex-start; }
.chat-user-bubble {
  max-width: 78%; padding: 10px 14px; border-radius: 16px 16px 4px 16px;
  background: #2d6b3a; color: #fff; font-family: 'Inter', sans-serif; font-size: 14px; line-height: 1.4;
}
.chat-typing { padding: 8px 14px; color: var(--subtext); font-size: 18px; letter-spacing: 2px; }

/* error corrections block (shown under the NPC reply) */
.chat-corrections {
  align-self: flex-start; max-width: 88%; margin: -2px 0 2px;
  padding: 10px 12px; border-radius: 12px;
  background: rgba(200, 80, 60, 0.07); border: 1px solid rgba(200, 80, 60, 0.22);
}
:root[data-theme="dark"] .chat-corrections {
  background: rgba(200, 80, 60, 0.14); border-color: rgba(200, 80, 60, 0.32);
}
.chat-corrections-title {
  font-family: 'Inter', sans-serif; font-size: 11px; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.04em; color: #b3503c; margin-bottom: 6px;
}
.chat-correction { padding: 3px 0; }
.chat-correction + .chat-correction { border-top: 1px solid rgba(200, 80, 60, 0.15); margin-top: 4px; padding-top: 6px; }
.chat-correction-line { font-family: 'Inter', sans-serif; font-size: 13px; line-height: 1.5; }
.chat-correction-bad { color: #b3503c; text-decoration: line-through; }
.chat-correction-arrow { margin: 0 6px; color: var(--subtext); }
.chat-correction-good { color: #2d6b3a; font-weight: 600; }
.chat-correction-expl { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); margin-top: 2px; }

/* banners */
.chat-banner { margin: 10px 16px; padding: 16px; border-radius: 14px; background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.08)); display: flex; flex-direction: column; align-items: center; gap: 10px; }
.chat-banner--win { background: rgba(45,107,58,0.08); border-color: rgba(45,107,58,0.3); }
.chat-banner-text { font-family: 'Inter', sans-serif; font-size: 15px; font-weight: 600; color: var(--text); text-align: center; }
.chat-banner-hint { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); text-align: center; margin-top: -4px; }

/* composer */
.chat-composer {
  display: flex; gap: 8px; padding: 10px 16px;
  flex-shrink: 0;
  background: var(--card-bg);
  border-top: 1px solid var(--border);
}
.chat-input {
  flex: 1; padding: 11px 14px; border-radius: 22px; border: 1px solid var(--border, rgba(0,0,0,0.12));
  background: var(--card-bg); color: var(--text); font-family: 'Inter', sans-serif; font-size: 14px; outline: none;
}
.chat-send {
  flex-shrink: 0; width: 44px; height: 44px; border-radius: 50%; border: none;
  background: #2d6b3a; color: #fff; display: flex; align-items: center; justify-content: center; cursor: pointer;
}
.chat-send:disabled { opacity: 0.5; cursor: default; }
</style>
