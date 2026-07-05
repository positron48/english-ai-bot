<template>
  <div class="chat-page" :class="{ 'chat-page--shell': !!scenarioCode }">
    <LgPageHeader
      :title="headerTitle"
      :show-back="true"
      @back="goBack"
    >
      <template #right>
        <button v-if="scenarioCode && session" class="chat-reset" type="button" @click="resetChat">
          {{ t('chat.reset') }}
        </button>
      </template>
    </LgPageHeader>

    <!-- NPC LIST -->
    <template v-if="!scenarioCode">
      <div v-if="loading" class="chat-loading">{{ t('common.loading') }}</div>
      <div v-else-if="!npcGroups.length" class="chat-loading">{{ t('chat.noPlaces') }}</div>
      <div v-else class="chat-list">
        <div v-for="g in npcGroups" :key="g.key" class="npc-block">
          <div class="npc-card-row">
            <button
              class="chat-card"
              :class="{ 'chat-card--locked': g.locked }"
              type="button"
              @click="onNpcClick(g)"
            >
              <img
                v-if="g.npcImageUrl"
                :src="mediaUrl(g.npcImageUrl)"
                class="npc-avatar"
                alt=""
              />
              <LgActivityIcon
                v-else
                type="conversation"
                :status="g.allDone ? 'green' : (g.locked ? 'gray' : 'orange')"
                :size="22"
              />
              <div class="chat-card-body">
                <div class="chat-card-title">
                  {{ g.npcName }}
                  <span v-if="g.npcRole" class="npc-role">, {{ g.npcRole }}</span>
                  <span v-if="g.hasAvailableIncompleteQuests" class="npc-bang" :title="t('chat.newQuests')">!</span>
                </div>
                <div class="chat-card-meta">
                  {{ placeLabel(g.placeType) }} · {{ g.level }}
                  <span v-if="g.allDone" class="chat-tag chat-tag--perfect"><LgIcon name="star-filled" :s="11" /> {{ t('chat.completed100') }}</span>
                  <span v-if="g.locked" class="chat-tag chat-tag--locked"><LgIcon name="lock" :s="11" /> {{ t('chat.locked') }}</span>
                </div>
              </div>
              <span v-if="g.allDone" class="chat-done chat-done--perfect"><LgIcon name="star-filled" :s="16" /></span>
              <span v-else-if="g.allPassed || g.hasCompletedQuests" class="chat-done"><LgIcon name="check" :s="16" /></span>
              <svg
                v-if="g.expandable"
                class="npc-chevron"
                :class="{ 'npc-chevron--open': expandedNpc === g.key }"
                width="16" height="16" viewBox="0 0 24 24" fill="none"
                stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
              >
                <path d="M6 9l6 6 6-6" />
              </svg>
            </button>
            <!-- free chat button: shown only when all quests are complete -->
            <button
              v-if="g.freeChatAvailable && g.freeScenario"
              class="npc-free-btn"
              type="button"
              :title="t('chat.freeChat')"
              @click="openScenario(g.freeScenario.code)"
            ><LgIcon name="message-circle" :s="18" /></button>
          </div>

          <!-- quest chain -->
          <div v-if="g.expandable && expandedNpc === g.key" class="npc-steps">
            <button
              v-for="(s, idx) in g.questScenarios"
              :key="s.code"
              class="npc-step"
              :class="{
                'npc-step--locked': s.locked,
                'npc-step--done': scenarioQuestPassed(s),
                'npc-step--perfect': scenarioQuestPerfect(s),
              }"
              type="button"
              :disabled="s.locked && !scenarioQuestPassed(s)"
              @click="openScenario(s.code)"
            >
              <img v-if="s.image_url" :src="mediaUrl(s.image_url)" class="npc-step-thumb" alt="" />
              <span v-else class="npc-step-num">{{ idx + 1 }}</span>
              <span class="npc-step-body">
                <span class="npc-step-title">{{ s.title }}</span>
                <span class="npc-step-meta">
                  <span class="chat-tag">{{ t('chat.quest') }}</span>
                  <span v-if="s.locked && s.cooldown_until" class="npc-cooldown"><LgIcon name="clock" :s="11" /> {{ formatCooldown(s.cooldown_until) }}</span>
                </span>
              </span>
              <span class="npc-step-state">
                <template v-if="scenarioQuestPerfect(s)"><LgIcon name="star-filled" :s="14" /></template>
                <template v-else-if="scenarioQuestPassed(s)"><LgIcon name="check" :s="14" /></template>
                <template v-else-if="s.locked"><LgIcon name="lock" :s="14" /></template>
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
        <!-- messages (task checklist, quest image and completion banners scroll along with the
             conversation instead of pinning to the top/bottom, so the thread keeps usable height
             when the keyboard is open and the composer never gets pushed off-screen). -->
        <div
          ref="scrollEl"
          class="chat-thread"
          :class="{ 'chat-thread--scene': !!session.image_url }"
          :style="session.image_url ? { backgroundImage: `url(${mediaUrl(session.image_url)})` } : undefined"
        >
          <div v-if="session.image_url" class="chat-quest-scene">
            <div v-if="session.is_quest && tasks.length" class="chat-tasks chat-tasks--scene">
              <div class="chat-tasks-title">{{ t('chat.tasksTitle') }}</div>
              <div
                v-for="task in tasks"
                :key="task.code"
                class="chat-task"
                :class="{ 'chat-task--done': task.completed, 'chat-task--optional': !task.required }"
              >
                <span class="chat-task-check"><LgIcon :name="task.completed ? 'check' : 'circle'" :s="14" /></span>
                <span class="chat-task-label">{{ task.title }}</span>
                <span v-if="!task.required" class="chat-task-opt">{{ t('chat.optional') }}</span>
              </div>
            </div>
          </div>

          <!-- task checklist -->
          <div v-else-if="session.is_quest && tasks.length" class="chat-tasks">
            <div class="chat-tasks-title">{{ t('chat.tasksTitle') }}</div>
            <div
              v-for="task in tasks"
              :key="task.code"
              class="chat-task"
              :class="{ 'chat-task--done': task.completed, 'chat-task--optional': !task.required }"
            >
              <span class="chat-task-check"><LgIcon :name="task.completed ? 'check' : 'circle'" :s="14" /></span>
              <span class="chat-task-label">{{ task.title }}</span>
              <span v-if="!task.required" class="chat-task-opt">{{ t('chat.optional') }}</span>
            </div>
          </div>

          <template v-for="(m, i) in messages" :key="i">
            <div
              class="chat-row"
              :class="[
                m.role === 'user' ? 'chat-row--user' : 'chat-row--npc',
                session.image_url && i === 0 && m.role !== 'user' ? 'chat-row--scene-opening' : '',
              ]"
            >
              <template v-if="m.role !== 'user'">
                <img
                  v-if="session.npc_image_url"
                  :src="mediaUrl(session.npc_image_url)"
                  class="chat-npc-avatar"
                  alt=""
                />
                <span v-else class="chat-npc-avatar chat-npc-avatar--fallback">{{ npcInitial }}</span>
                <LgSpeechBubble :text="m.content" />
              </template>
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
            <img
              v-if="session.npc_image_url"
              :src="mediaUrl(session.npc_image_url)"
              class="chat-npc-avatar"
              alt=""
            />
            <span v-else class="chat-npc-avatar chat-npc-avatar--fallback">{{ npcInitial }}</span>
            <div class="chat-typing">…</div>
          </div>

          <!-- transient send error (styled notice, not a chat bubble) -->
          <div v-if="sendError" class="chat-error" role="alert">
            <span>{{ t('chat.errorSend') }}</span>
            <button class="chat-error-dismiss" type="button" @click="sendError = false"><LgIcon name="x" :s="14" /></button>
          </div>

          <!-- completion / budget banners -->
          <div v-if="questPassed" class="chat-banner chat-banner--win" :class="{ 'chat-banner--perfect': allTasksDone }">
            <LgLumi pose="clapping" :size="48" />
            <div class="chat-banner-text">
              <span v-if="allTasksDone"><LgIcon name="star-filled" :s="14" /> {{ t('chat.questPerfect') }}</span>
              <span v-else><LgIcon name="check" :s="14" /> {{ t('chat.questComplete') }}</span>
            </div>
            <div v-if="status === 'open' && !allTasksDone" class="chat-banner-hint">{{ t('chat.questOptionalHint') }}</div>
            <div v-else-if="status === 'open'" class="chat-banner-hint">{{ t('chat.questCompleteHint') }}</div>
            <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
          </div>
          <div v-else-if="budgetExhausted" class="chat-banner">
            <div class="chat-banner-text">{{ t('chat.budgetEnded') }}</div>
            <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
          </div>
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
          <VoiceMicButton
            :lang="learning?.target_lang ?? 'en'"
            :disabled="sending"
            :label="t('sentence.voiceInput')"
            @transcript="onVoiceTranscript"
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
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  courseClient,
  type ConversationScenarioSummary,
  type ConversationSessionState,
  type ConversationMessage,
  type ConversationTask,
} from '../api/courseClient'
import { buildNpcGroups, scenarioQuestPassed, scenarioQuestPerfect } from '../utils/conversations'
import type { NpcGroup } from '../utils/conversations'
import { mediaUrl } from '../utils/mediaUrl'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgSpeechBubble from '../components/linglow/LgSpeechBubble.vue'
import LgButton from '../components/linglow/LgButton.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import VoiceMicButton from '../components/VoiceMicButton.vue'
import { useCourse } from '../composables/useCourse'
import { useLearningConfig } from '../composables/useLearningConfig'
import { showAlert, showConfirm } from '../composables/useDialog'

const { t } = useI18n()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { learning, ensureLearningLoaded } = useLearningConfig()
ensureLearningLoaded()

// Voice input: dictate a chat reply in the target language.
const onVoiceTranscript = (text: string) => {
  input.value = input.value ? `${input.value.trimEnd()} ${text}` : text
  focusComposer()
}
const route = useRoute()
const router = useRouter()

const districtCode = computed(() => String(route.params.districtCode || ''))
const scenarioCode = computed(() => String(route.params.scenarioCode || ''))

const loading = ref(true)
const scenarios = ref<ConversationScenarioSummary[]>([])
const expandedNpc = ref('')

const npcGroups = computed<NpcGroup[]>(() => buildNpcGroups(scenarios.value, currentCourseCode.value || ''))

const PLACE_LABELS: Record<string, string> = {
  cafe: 'Кафе', shop: 'Магазин', police_station: 'Полиция',
  pharmacy: 'Аптека', hotel: 'Отель', restaurant: 'Ресторан', market: 'Рынок',
}
function placeLabel(placeType: string): string {
  return PLACE_LABELS[placeType] || placeType
}

function onNpcClick(g: NpcGroup) {
  if (g.locked) return
  if (!g.expandable) {
    // Single quest or free chat only — open directly (including replay of a passed quest).
    const quest = g.questScenarios[0]
    if (quest && (!quest.locked || scenarioQuestPassed(quest))) {
      openScenario(quest.code)
      return
    }
    if (g.freeScenario && !g.freeScenario.locked) openScenario(g.freeScenario.code)
    return
  }
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
const sendError = ref(false)

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
const allTasksDone = computed(() => tasks.value.length > 0 && tasks.value.every(t => t.completed))
const npcInitial = computed(() => (session.value?.npc_name || '?').trim().slice(0, 1).toUpperCase())

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
  sendError.value = false
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
    // Surface the failure as a styled notice (not a fake NPC message) and restore the draft.
    sendError.value = true
    input.value = text
    if (messages.value[messages.value.length - 1]?.role === 'user') messages.value.pop()
  } finally {
    sending.value = false
    await scrollToBottom()
    if (canSend.value) focusComposer()
  }
}

async function resetChat() {
  if (!session.value) return
  if (!await showConfirm(t('chat.resetConfirm'))) return
  loading.value = true
  sendError.value = false
  try {
    const s = await courseClient.resetConversationSession(session.value.session_id)
    session.value = s
    messages.value = [...(s.messages || [])]
    tasks.value = s.tasks || []
    status.value = s.status
    questPassed.value = s.quest_passed
    budgetExhausted.value = false
    await scrollToBottom()
  } catch (e: any) {
    await showAlert(e?.message || t('chat.notAvailable'))
  } finally {
    loading.value = false
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
    if (scenarioCode.value && canSend.value) focusComposer()
  }
}

// Cooldown countdown timer.
const nowMs = ref(Date.now())
let cooldownTimer: ReturnType<typeof setInterval> | null = null
onMounted(() => { cooldownTimer = setInterval(() => { nowMs.value = Date.now() }, 1000) })
onUnmounted(() => { if (cooldownTimer) clearInterval(cooldownTimer) })

function formatCooldown(isoUntil: string | null): string {
  if (!isoUntil) return ''
  const diff = Math.max(0, new Date(isoUntil).getTime() - nowMs.value)
  if (diff === 0) return ''
  const h = Math.floor(diff / 3600000)
  const m = Math.floor((diff % 3600000) / 60000)
  const s = Math.floor((diff % 60000) / 1000)
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
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
.npc-role { color: var(--subtext); font-weight: 500; }
.chat-card-meta { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); margin-top: 2px; }
.chat-tag {
  display: inline-block; margin-left: 6px; padding: 1px 7px; border-radius: 10px;
  font-size: 10px; font-weight: 600; background: rgba(45,107,58,0.12); color: #2d6b3a;
}
.chat-tag--free { background: rgba(200,168,75,0.18); color: #9a7b1e; }
.chat-tag--locked { background: rgba(120,120,120,0.16); color: var(--subtext); }
.chat-tag--perfect { background: rgba(200,168,75,0.22); color: #9a7b1e; }
.chat-done { color: #2d6b3a; font-weight: 700; font-size: 18px; }
.chat-done--perfect { color: #c8a84b; text-shadow: 0 1px 4px rgba(200,168,75,0.35); }

/* "!" marker on NPCs that still have unfinished quests */
.npc-bang {
  display: inline-flex; align-items: center; justify-content: center;
  width: 18px; height: 18px; margin-left: 6px; border-radius: 50%;
  background: #d97706; color: #fff; font-size: 12px; font-weight: 800; line-height: 1;
  vertical-align: middle;
}

/* reset button in the header */
.chat-reset {
  border: 1px solid var(--border); border-radius: 16px; padding: 5px 12px;
  background: var(--card-bg); color: var(--text); font-size: 12px; font-weight: 600; cursor: pointer;
  font-family: 'Inter', sans-serif;
}
.chat-reset:active { opacity: 0.7; }

/* transient send error notice */
.chat-error {
  display: flex; align-items: center; justify-content: space-between; gap: 10px;
  flex-shrink: 0; padding: 10px 14px; border-radius: 12px;
  background: rgba(200,80,60,0.10); border: 1px solid rgba(200,80,60,0.30);
  color: #b3503c; font-family: 'Inter', sans-serif; font-size: 13px;
}
:root[data-theme="dark"] .chat-error { background: rgba(200,80,60,0.18); border-color: rgba(200,80,60,0.4); }
.chat-error-dismiss { border: none; background: none; color: inherit; cursor: pointer; font-size: 14px; }

/* NPC avatar image in card */
.npc-avatar { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }

/* NPC card row wraps the card + optional free-chat button */
.npc-card-row { display: flex; align-items: stretch; gap: 6px; }
.npc-card-row .chat-card { flex: 1; }
.npc-free-btn {
  flex-shrink: 0; width: 48px; border-radius: 14px;
  border: 1px solid var(--border, rgba(0,0,0,0.08)); background: var(--card-bg);
  font-size: 20px; cursor: pointer; display: flex; align-items: center; justify-content: center;
}
.npc-free-btn:active { opacity: 0.7; }

/* cooldown countdown label inside a quest step */
.npc-cooldown { margin-left: 6px; font-size: 11px; color: #d97706; font-weight: 600; }

/* quest scene background in dialog mode */
.chat-thread {
  position: relative;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.chat-thread--scene {
  background-color: var(--bg);
  background-size: cover;
  background-position: center top;
  background-repeat: no-repeat;
}
.chat-thread--scene::before {
  content: "";
  position: absolute;
  inset: 0;
  background:
    linear-gradient(to bottom, rgba(0,0,0,0.18) 0%, rgba(0,0,0,0.10) 26%, rgba(0,0,0,0.42) 68%, var(--bg) 100%),
    linear-gradient(to right, rgba(0,0,0,0.26), rgba(0,0,0,0.05) 48%, rgba(0,0,0,0.24));
  pointer-events: none;
  z-index: 0;
}
.chat-thread--scene > * {
  position: relative;
  z-index: 1;
}

/* quest task panel starts the dialog over the scene background */
.chat-quest-scene {
  position: relative;
  flex-shrink: 0;
  min-height: min(42vh, 360px);
  margin: -12px -16px 0;
  padding: 14px 16px 34px;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
}
@media (min-width: 640px) {
  .chat-quest-scene { min-height: min(48vh, 460px); }
}

/* NPC groups + chain steps */
.npc-block { display: flex; flex-direction: column; }
.chat-card--locked { opacity: 0.6; cursor: default; }
.npc-chevron { color: var(--subtext); flex-shrink: 0; transition: transform 0.2s; }
.npc-chevron--open { transform: rotate(180deg); }

.npc-steps { display: flex; flex-direction: column; gap: 6px; margin: 6px 0 2px 14px; padding-left: 14px; border-left: 2px solid var(--border, rgba(0,0,0,0.08)); }
.npc-step {
  display: flex; align-items: center; gap: 10px; width: 100%;
  padding: 10px 12px; border-radius: 12px; border: 1px solid var(--border, rgba(0,0,0,0.08));
  background: var(--card-bg); cursor: pointer; text-align: left;
}
.npc-step:disabled { cursor: default; opacity: 0.55; }
.npc-step--done { background: rgba(45,107,58,0.06); border-color: rgba(45,107,58,0.25); }
.npc-step--perfect { background: rgba(200,168,75,0.08); border-color: rgba(200,168,75,0.35); }
.npc-step-num {
  width: 22px; height: 22px; flex-shrink: 0; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 12px; font-weight: 700; color: var(--subtext);
  background: var(--chip-bg, rgba(0,0,0,0.05));
}
.npc-step-thumb {
  width: 46px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 8px;
  object-fit: cover;
  border: 1px solid var(--border, rgba(0,0,0,0.08));
}
.npc-step--done .npc-step-num { background: #2d6b3a; color: #fff; }
.npc-step-body { flex: 1; min-width: 0; }
.npc-step-title { display: block; font-family: 'Inter', sans-serif; font-size: 14px; font-weight: 600; color: var(--text); }
.npc-step-meta { margin-top: 2px; }
.npc-step-state { flex-shrink: 0; color: var(--subtext); font-size: 16px; font-weight: 700; }
.npc-step--done .npc-step-state { color: #2d6b3a; }
.npc-step--perfect .npc-step-state { color: #c8a84b; }

/* task checklist — scrolls with the thread instead of pinning to the top of the screen, so it
   doesn't eat into the conversation's visible height when the mobile keyboard is open. */
.chat-tasks { flex-shrink: 0; padding: 12px 14px; border-radius: 14px; background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.08)); }
.chat-tasks--scene {
  position: relative;
  z-index: 1;
  width: min(100%, 430px);
  background: rgba(255,255,255,0.88);
  border-color: rgba(255,255,255,0.62);
  box-shadow: 0 14px 34px rgba(0,0,0,0.28);
  backdrop-filter: blur(14px) saturate(1.12);
}
:root[data-theme="dark"] .chat-tasks--scene {
  background: rgba(24,28,24,0.84);
  border-color: rgba(255,255,255,0.16);
}
.chat-tasks-title { font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 700; color: var(--subtext); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 8px; }
.chat-task { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-family: 'Inter', sans-serif; font-size: 13px; color: var(--text); transition: opacity 0.25s; }
.chat-tasks--scene .chat-tasks-title { color: rgba(31,41,32,0.72); }
.chat-tasks--scene .chat-task { color: #172017; text-shadow: 0 1px 0 rgba(255,255,255,0.35); }
.chat-tasks--scene .chat-task-opt { color: rgba(31,41,32,0.66); }
:root[data-theme="dark"] .chat-tasks--scene .chat-tasks-title,
:root[data-theme="dark"] .chat-tasks--scene .chat-task-opt { color: rgba(235,238,230,0.72); }
:root[data-theme="dark"] .chat-tasks--scene .chat-task { color: #f4f7ef; text-shadow: 0 1px 2px rgba(0,0,0,0.55); }
.chat-task-check { width: 18px; text-align: center; color: var(--subtext); }
.chat-task--done { color: #2d6b3a; }
.chat-task--done .chat-task-check { color: #2d6b3a; }
.chat-task--optional { opacity: 0.85; }
.chat-task-opt { font-size: 10px; color: var(--subtext); }

/* NPC reply bubbles in chat are wider than the default Lumi speech bubble. */
.chat-thread :deep(.lg-bubble) { max-width: 400px; font-size: 14px; }
.chat-scene { font-family: 'Inter', sans-serif; font-size: 12px; font-style: italic; color: var(--subtext); text-align: center; margin: 0 0 6px; }
.chat-row { display: flex; }
.chat-row--user { justify-content: flex-end; }
.chat-row--npc { justify-content: flex-start; align-items: flex-end; gap: 8px; }
.chat-row--scene-opening {
  margin-top: -20px;
  position: relative;
  z-index: 1;
}
.chat-row--scene-opening :deep(.lg-bubble) {
  background: rgba(255,255,255,0.92);
  border: 1px solid rgba(255,255,255,0.7);
  box-shadow: 0 12px 30px rgba(0,0,0,0.18);
  backdrop-filter: blur(12px) saturate(1.08);
}
:root[data-theme="dark"] .chat-row--scene-opening :deep(.lg-bubble) {
  background: rgba(24,28,24,0.88);
  border-color: rgba(255,255,255,0.14);
}
.chat-npc-avatar {
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid var(--border, rgba(0,0,0,0.08));
  background: var(--card-bg);
}
.chat-npc-avatar--fallback {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #2d6b3a;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 800;
  background: rgba(45,107,58,0.12);
}
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

/* banners — scroll inside the thread (not a fixed block) so they never push the composer
   off-screen, even on small screens or when there are no optional tasks to pad the layout out. */
.chat-banner { flex-shrink: 0; padding: 16px; border-radius: 14px; background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.08)); display: flex; flex-direction: column; align-items: center; gap: 10px; }
.chat-banner--win { background: rgba(45,107,58,0.08); border-color: rgba(45,107,58,0.3); }
.chat-banner--perfect { background: rgba(200,168,75,0.10); border-color: rgba(200,168,75,0.35); }
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
