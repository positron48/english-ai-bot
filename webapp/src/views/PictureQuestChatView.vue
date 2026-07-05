<template>
  <div class="chat-page chat-page--shell">
    <LgPageHeader
      :title="headerTitle"
      :show-back="true"
      @back="goBack"
    >
      <template #right>
        <button v-if="session" class="chat-reset" type="button" @click="resetChat">
          {{ t('chat.reset') }}
        </button>
      </template>
    </LgPageHeader>

    <div v-if="loading" class="chat-loading">{{ t('common.loading') }}</div>
    <template v-else-if="session">
      <div ref="scrollEl" class="chat-thread">
        <!-- the picture is the core content: persistent panel, tap to expand full-screen -->
        <div v-if="session.image_url" class="pq-image" @click="imageExpanded = true">
          <img :src="mediaUrl(session.image_url)" alt="" class="pq-img" />
          <span class="pq-image-hint">{{ t('pictureQuest.tapToExpand') }}</span>
          <button
            v-if="hasHints"
            class="pq-hint-btn"
            type="button"
            :title="t('pictureQuest.hintTitle')"
            :aria-label="t('pictureQuest.hintTitle')"
            @click.stop="hintOpen = true"
          >
            <LgIcon name="lightbulb" :s="18" />
          </button>
        </div>

        <!-- task checklist -->
        <div v-if="tasks.length" class="chat-tasks">
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
            :class="m.role === 'user' ? 'chat-row--user' : 'chat-row--npc'"
          >
            <template v-if="m.role !== 'user'">
              <LgLumi pose="teacher" :size="36" class="pq-lumi" />
              <LgSpeechBubble><ClickableText :text="m.content" subtle-underline /></LgSpeechBubble>
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
          <LgLumi pose="writes" :size="36" class="pq-lumi" />
          <div class="chat-typing">…</div>
        </div>

        <!-- transient send error -->
        <div v-if="sendError" class="chat-error" role="alert">
          <span>{{ t('chat.errorSend') }}</span>
          <button class="chat-error-dismiss" type="button" @click="sendError = false"><LgIcon name="x" :s="14" /></button>
        </div>

        <!-- completion / budget banners -->
        <div v-if="questPassed" class="chat-banner chat-banner--win" :class="{ 'chat-banner--perfect': allTasksDone }">
          <LgLumi :pose="allTasksDone ? 'star' : 'clapping'" :size="48" />
          <div class="chat-banner-text">
            <span v-if="allTasksDone"><LgIcon name="star-filled" :s="14" /> {{ t('chat.questPerfect') }}</span>
            <span v-else><LgIcon name="check" :s="14" /> {{ t('chat.questComplete') }}</span>
          </div>
          <div v-if="status === 'open' && !allTasksDone" class="chat-banner-hint">{{ t('chat.questOptionalHint') }}</div>
          <div v-else-if="status === 'open'" class="chat-banner-hint">{{ t('chat.questCompleteHint') }}</div>
          <LgButton @click="goBack">{{ t('pictureQuest.backToList') }}</LgButton>
        </div>
        <div v-else-if="budgetExhausted" class="chat-banner">
          <div class="chat-banner-text">{{ t('chat.budgetEnded') }}</div>
          <LgButton @click="goBack">{{ t('pictureQuest.backToList') }}</LgButton>
        </div>
      </div>

      <!-- composer -->
      <div v-if="canSend" class="chat-composer">
        <input
          ref="composerInput"
          v-model="input"
          class="chat-input"
          type="text"
          :placeholder="t('pictureQuest.placeholder')"
          @keyup.enter="send"
        />
        <VoiceMicButton
          :lang="targetLang || 'en'"
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

      <!-- full-screen picture overlay -->
      <div v-if="imageExpanded && session.image_url" class="pq-overlay" @click="imageExpanded = false">
        <img :src="mediaUrl(session.image_url)" alt="" class="pq-overlay-img" />
        <button class="pq-overlay-close" type="button"><LgIcon name="x" :s="18" /></button>
      </div>

      <!-- vocabulary hint sheet (tied to the course language, not the UI locale) -->
      <PictureHintSheet
        v-if="hintOpen"
        :target-lang="targetLang"
        :native-lang="nativeLang"
        @close="hintOpen = false"
      />
    </template>
    <div v-else class="chat-loading">{{ loadError || t('chat.notAvailable') }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import {
  courseClient,
  type PictureQuestSessionState,
  type ConversationMessage,
  type ConversationTask,
} from '../api/courseClient'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgSpeechBubble from '../components/linglow/LgSpeechBubble.vue'
import ClickableText from '../components/ClickableText.vue'
import LgButton from '../components/linglow/LgButton.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import PictureHintSheet from '../components/PictureHintSheet.vue'
import VoiceMicButton from '../components/VoiceMicButton.vue'
import { getHintSections } from '../constants/pictureHintPhrasebook'
import { useCourse } from '../composables/useCourse'
import { showAlert, showConfirm } from '../composables/useDialog'
import { mediaUrl } from '../utils/mediaUrl'

const { t } = useI18n()
const { currentCourse, currentCourseCode, ensureCourseLoaded } = useCourse()
const route = useRoute()
const router = useRouter()

const questCode = computed(() => String(route.params.questCode || ''))

const loading = ref(true)
const session = ref<PictureQuestSessionState | null>(null)
const messages = ref<ConversationMessage[]>([])
const tasks = ref<ConversationTask[]>([])
const input = ref('')
const sending = ref(false)
const questPassed = ref(false)
const budgetExhausted = ref(false)
const status = ref('open')
const imageExpanded = ref(false)
const hintOpen = ref(false)
const scrollEl = ref<HTMLElement | null>(null)
const composerInput = ref<HTMLInputElement | null>(null)
const loadError = ref('')
const sendError = ref(false)

function focusComposer() {
  nextTick(() => composerInput.value?.focus())
}

function onVoiceTranscript(text: string) {
  input.value = input.value ? `${input.value.trimEnd()} ${text}` : text
  focusComposer()
}

const headerTitle = computed(() => session.value?.title || t('pictureQuest.title'))
// Keep the composer usable after the quest passes so the learner can finish optional tasks;
// the session only truly ends when status flips to completed or budget runs out.
const canSend = computed(() => status.value === 'open' && !budgetExhausted.value)
const allTasksDone = computed(() => tasks.value.length > 0 && tasks.value.every(t => t.completed))

// Hint sheet is driven by the course languages (target + native), never the UI locale.
const targetLang = computed(() => currentCourse.value?.target_language || '')
const nativeLang = computed(() => currentCourse.value?.native_language || 'ru')
const hasHints = computed(() => getHintSections(targetLang.value, nativeLang.value).length > 0)

function goBack() {
  router.push({ name: 'PictureQuestCategories' })
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
    const res = await courseClient.postPictureQuestMessage(session.value.session_id, text)
    if (res.reply) messages.value.push({ role: 'assistant', content: res.reply, corrections: res.corrections })
    tasks.value = res.tasks
    status.value = res.status
    questPassed.value = res.quest_passed
    budgetExhausted.value = res.budget_exhausted
  } catch {
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
    const s = await courseClient.resetPictureQuestSession(session.value.session_id)
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
    const s = await courseClient.startPictureQuestSession(questCode.value, course)
    session.value = s
    messages.value = [...(s.messages || [])]
    tasks.value = s.tasks || []
    status.value = s.status
    questPassed.value = s.quest_passed
    await scrollToBottom()
  } catch (e: any) {
    const msg = String(e?.message || '')
    if (msg.includes('403')) loadError.value = t('chat.requiresPro')
    else if (msg.includes('404')) loadError.value = t('chat.scenarioMissing')
    else loadError.value = t('chat.notAvailable')
    console.error('Failed to start picture quest:', e)
  } finally {
    loading.value = false
    if (canSend.value) focusComposer()
  }
}

watch(questCode, () => { loadForRoute() })
onMounted(loadForRoute)
</script>

<style scoped>
.chat-page { display: flex; flex-direction: column; min-height: 100%; padding-bottom: 16px; }

/* fixed app-shell pinned between status bar and bottom nav (same as PlaceChatView) */
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
  .chat-page--shell {
    left: 220px;
    top: 0;
    bottom: 0;
    max-width: 880px;
    margin: 0 auto;
  }
}
.chat-loading { padding: 40px 16px; text-align: center; color: var(--subtext); }

.chat-reset {
  border: 1px solid var(--border); border-radius: 16px; padding: 5px 12px;
  background: var(--card-bg); color: var(--text); font-size: 12px; font-weight: 600; cursor: pointer;
  font-family: 'Inter', sans-serif;
}
.chat-reset:active { opacity: 0.7; }

/* picture panel */
/* Show the whole picture: fill the block width for large images, or its natural
   1:1 size when it's narrower than the block — never crop to a strip. */
.pq-image { position: relative; flex-shrink: 0; cursor: zoom-in; width: fit-content; max-width: 100%; margin: 0 auto; }
.pq-img { max-width: 100%; height: auto; border-radius: 14px; display: block; }
.pq-image-hint {
  position: absolute; right: 8px; bottom: 8px;
  padding: 2px 8px; border-radius: 10px;
  background: rgba(0,0,0,0.55); color: #fff;
  font-family: 'Inter', sans-serif; font-size: 10px;
}
.pq-hint-btn {
  position: absolute; right: 8px; top: 8px;
  width: 34px; height: 34px; border-radius: 50%; border: none;
  display: inline-flex; align-items: center; justify-content: center;
  background: rgba(255,255,255,0.92); color: #b8860b; cursor: pointer;
  box-shadow: 0 1px 6px rgba(0,0,0,0.25);
}
.pq-hint-btn:active { transform: scale(0.94); }
:root[data-theme="dark"] .pq-hint-btn { background: rgba(40,40,40,0.9); color: #e8c15a; }

/* full-screen overlay */
.pq-overlay {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0,0,0,0.88);
  display: flex; align-items: center; justify-content: center;
  cursor: zoom-out;
}
.pq-overlay-img { max-width: 100%; max-height: 100%; object-fit: contain; }
.pq-overlay-close {
  position: absolute; top: max(env(safe-area-inset-top, 0px), 12px); right: 14px;
  width: 36px; height: 36px; border-radius: 50%; border: none;
  background: rgba(255,255,255,0.15); color: #fff; font-size: 16px; cursor: pointer;
}

/* Lumi avatar next to assistant bubbles */
.pq-lumi { flex-shrink: 0; margin-right: 6px; align-self: flex-end; }

.chat-error {
  display: flex; align-items: center; justify-content: space-between; gap: 10px;
  flex-shrink: 0; padding: 10px 14px; border-radius: 12px;
  background: rgba(200,80,60,0.10); border: 1px solid rgba(200,80,60,0.30);
  color: #b3503c; font-family: 'Inter', sans-serif; font-size: 13px;
}
:root[data-theme="dark"] .chat-error { background: rgba(200,80,60,0.18); border-color: rgba(200,80,60,0.4); }
.chat-error-dismiss { border: none; background: none; color: inherit; cursor: pointer; font-size: 14px; }

/* task checklist */
.chat-tasks { flex-shrink: 0; padding: 12px 14px; border-radius: 14px; background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.08)); }
.chat-tasks-title { font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 700; color: var(--subtext); text-transform: uppercase; letter-spacing: 0.04em; margin-bottom: 8px; }
.chat-task { display: flex; align-items: center; gap: 8px; padding: 4px 0; font-family: 'Inter', sans-serif; font-size: 13px; color: var(--text); transition: opacity 0.25s; }
.chat-task-check { width: 18px; text-align: center; color: var(--subtext); }
.chat-task--done { color: #2d6b3a; }
.chat-task--done .chat-task-check { color: #2d6b3a; }
.chat-task--optional { opacity: 0.85; }
.chat-task-opt { font-size: 10px; color: var(--subtext); }

/* thread */
.chat-thread { flex: 1; min-height: 0; overflow-y: auto; padding: 12px 16px; display: flex; flex-direction: column; gap: 10px; }
.chat-thread :deep(.lg-bubble) { max-width: 400px; font-size: 14px; }
.chat-row { display: flex; }
.chat-row--user { justify-content: flex-end; }
.chat-row--npc { justify-content: flex-start; align-items: flex-end; }
.chat-user-bubble {
  max-width: 78%; padding: 10px 14px; border-radius: 16px 16px 4px 16px;
  background: #2d6b3a; color: #fff; font-family: 'Inter', sans-serif; font-size: 14px; line-height: 1.4;
}
.chat-typing { padding: 8px 14px; color: var(--subtext); font-size: 18px; letter-spacing: 2px; }

/* error corrections block */
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
