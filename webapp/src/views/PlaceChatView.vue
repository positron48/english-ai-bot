<template>
  <div class="chat-page">
    <LgPageHeader
      :title="headerTitle"
      :show-back="true"
      @back="goBack"
    />

    <!-- SCENARIO LIST -->
    <template v-if="!scenarioCode">
      <div v-if="loading" class="chat-loading">{{ t('common.loading') }}</div>
      <div v-else-if="!scenarios.length" class="chat-loading">{{ t('chat.noPlaces') }}</div>
      <div v-else class="chat-list">
        <button
          v-for="s in scenarios"
          :key="s.code"
          class="chat-card"
          type="button"
          @click="openScenario(s.code)"
        >
          <LgActivityIcon type="conversation" :status="s.session_status === 'completed' ? 'green' : 'gray'" :size="22" />
          <div class="chat-card-body">
            <div class="chat-card-title">{{ s.title }}</div>
            <div class="chat-card-meta">
              {{ s.npc_name }} · {{ s.cefr_level }}
              <span v-if="s.is_quest" class="chat-tag">{{ t('chat.quest') }}</span>
              <span v-else class="chat-tag chat-tag--free">{{ t('chat.free') }}</span>
            </div>
          </div>
          <span v-if="s.session_status === 'completed'" class="chat-done">✓</span>
        </button>
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
          <div
            v-for="(m, i) in messages"
            :key="i"
            class="chat-row"
            :class="m.role === 'user' ? 'chat-row--user' : 'chat-row--npc'"
          >
            <LgSpeechBubble v-if="m.role !== 'user'" :text="m.content" />
            <div v-else class="chat-user-bubble">{{ m.content }}</div>
          </div>
          <div v-if="sending" class="chat-row chat-row--npc">
            <div class="chat-typing">…</div>
          </div>
        </div>

        <!-- completion / budget banners -->
        <div v-if="questPassed" class="chat-banner chat-banner--win">
          <LgLumi pose="clapping" :size="48" />
          <div class="chat-banner-text">{{ t('chat.questComplete') }}</div>
          <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
        </div>
        <div v-else-if="budgetExhausted" class="chat-banner">
          <div class="chat-banner-text">{{ t('chat.budgetEnded') }}</div>
          <LgButton @click="goBack">{{ t('chat.backToDistrict') }}</LgButton>
        </div>

        <!-- composer -->
        <div v-if="canSend" class="chat-composer">
          <input
            v-model="input"
            class="chat-input"
            type="text"
            :placeholder="t('chat.placeholder')"
            :disabled="sending"
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
import { computed, nextTick, onMounted, ref } from 'vue'
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

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const districtCode = computed(() => String(route.params.districtCode || ''))
const scenarioCode = computed(() => String(route.params.scenarioCode || ''))

const loading = ref(true)
const scenarios = ref<ConversationScenarioSummary[]>([])

const session = ref<ConversationSessionState | null>(null)
const messages = ref<ConversationMessage[]>([])
const tasks = ref<ConversationTask[]>([])
const input = ref('')
const sending = ref(false)
const questPassed = ref(false)
const budgetExhausted = ref(false)
const status = ref('open')
const scrollEl = ref<HTMLElement | null>(null)
const loadError = ref('')

const headerTitle = computed(() => {
  if (scenarioCode.value && session.value) return session.value.title
  return t('chat.placesTitle')
})
const canSend = computed(() => status.value === 'open' && !questPassed.value && !budgetExhausted.value)

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
    if (res.reply) messages.value.push({ role: 'assistant', content: res.reply })
    tasks.value = res.tasks
    status.value = res.status
    questPassed.value = res.quest_passed
    budgetExhausted.value = res.budget_exhausted
  } catch {
    messages.value.push({ role: 'assistant', content: t('chat.errorSend') })
  } finally {
    sending.value = false
    await scrollToBottom()
  }
}

onMounted(async () => {
  try {
    if (scenarioCode.value) {
      const s = await courseClient.startConversationSession(scenarioCode.value)
      session.value = s
      messages.value = [...(s.messages || [])]
      tasks.value = s.tasks || []
      status.value = s.status
      questPassed.value = s.quest_passed
      await scrollToBottom()
    } else {
      const r = await courseClient.listConversationScenarios(districtCode.value)
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
})
</script>

<style scoped>
.chat-page { display: flex; flex-direction: column; min-height: 100%; padding-bottom: 16px; }
.chat-loading { padding: 40px 16px; text-align: center; color: var(--subtext); }

/* scenario list */
.chat-list { padding: 8px 16px; display: flex; flex-direction: column; gap: 10px; }
.chat-card {
  display: flex; align-items: center; gap: 12px; width: 100%;
  padding: 14px; border-radius: 14px; border: 1px solid var(--border, rgba(0,0,0,0.08));
  background: var(--card, #fff); cursor: pointer; text-align: left;
}
.chat-card-body { flex: 1; min-width: 0; }
.chat-card-title { font-family: 'Inter', sans-serif; font-size: 15px; font-weight: 600; color: var(--text); }
.chat-card-meta { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); margin-top: 2px; }
.chat-tag {
  display: inline-block; margin-left: 6px; padding: 1px 7px; border-radius: 10px;
  font-size: 10px; font-weight: 600; background: rgba(45,107,58,0.12); color: #2d6b3a;
}
.chat-tag--free { background: rgba(200,168,75,0.18); color: #9a7b1e; }
.chat-done { color: #2d6b3a; font-weight: 700; font-size: 18px; }

/* task checklist */
.chat-tasks { margin: 8px 16px; padding: 12px 14px; border-radius: 14px; background: var(--card, #fff); border: 1px solid var(--border, rgba(0,0,0,0.08)); }
.chat-tasks-title { font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 700; color: var(--subtext); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 8px; }
.chat-task { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-family: 'Inter', sans-serif; font-size: 13px; color: var(--text); transition: opacity 0.25s; }
.chat-task-check { width: 18px; text-align: center; color: var(--subtext); }
.chat-task--done { color: #2d6b3a; }
.chat-task--done .chat-task-check { color: #2d6b3a; }
.chat-task--optional { opacity: 0.85; }
.chat-task-opt { font-size: 10px; color: var(--subtext); }

/* thread */
.chat-thread { flex: 1; overflow-y: auto; padding: 12px 16px; display: flex; flex-direction: column; gap: 10px; }
.chat-scene { font-family: 'Inter', sans-serif; font-size: 12px; font-style: italic; color: var(--subtext); text-align: center; margin: 0 0 6px; }
.chat-row { display: flex; }
.chat-row--user { justify-content: flex-end; }
.chat-row--npc { justify-content: flex-start; }
.chat-user-bubble {
  max-width: 78%; padding: 10px 14px; border-radius: 16px 16px 4px 16px;
  background: #2d6b3a; color: #fff; font-family: 'Inter', sans-serif; font-size: 14px; line-height: 1.4;
}
.chat-typing { padding: 8px 14px; color: var(--subtext); font-size: 18px; letter-spacing: 2px; }

/* banners */
.chat-banner { margin: 10px 16px; padding: 16px; border-radius: 14px; background: var(--card, #fff); border: 1px solid var(--border, rgba(0,0,0,0.08)); display: flex; flex-direction: column; align-items: center; gap: 10px; }
.chat-banner--win { background: rgba(45,107,58,0.08); border-color: rgba(45,107,58,0.3); }
.chat-banner-text { font-family: 'Inter', sans-serif; font-size: 15px; font-weight: 600; color: var(--text); text-align: center; }

/* composer */
.chat-composer { display: flex; gap: 8px; padding: 10px 16px; }
.chat-input {
  flex: 1; padding: 11px 14px; border-radius: 22px; border: 1px solid var(--border, rgba(0,0,0,0.12));
  background: var(--card, #fff); color: var(--text); font-family: 'Inter', sans-serif; font-size: 14px; outline: none;
}
.chat-send {
  flex-shrink: 0; width: 44px; height: 44px; border-radius: 50%; border: none;
  background: #2d6b3a; color: #fff; display: flex; align-items: center; justify-content: center; cursor: pointer;
}
.chat-send:disabled { opacity: 0.5; cursor: default; }
</style>
