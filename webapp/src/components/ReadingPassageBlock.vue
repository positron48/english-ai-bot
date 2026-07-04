<template>
  <div class="screen">
    <main class="reader-shell">
      <header class="header">
        <button type="button" class="back-button" :aria-label="t('common.back')" @click="goBack">
          <Icon name="arrow-left" />
        </button>
        <h1 class="title">{{ block.reading_passage?.title || block.title }}</h1>
        <div class="header-actions">
          <button
            v-if="canDelete"
            type="button"
            class="icon-button icon-button-danger"
            :disabled="deleting"
            :aria-label="t('reading.deleteText')"
            :title="t('reading.deleteText')"
            @click.stop="emit('delete-request')"
          >
            <Icon name="trash" />
          </button>
          <button
            type="button"
            class="icon-button"
            :class="{ 'icon-button--stop': isAutoplaying }"
            :aria-label="isAutoplaying ? 'Остановить текст' : 'Автовоспроизведение'"
            :title="isAutoplaying ? 'Остановить текст' : 'Автовоспроизведение'"
            @click="toggleAutoplay"
          >
            <Icon :name="isAutoplaying ? 'stop' : 'play'" />
          </button>
        </div>
      </header>

      <section class="content">
        <img
          v-if="coverHeroRelPath"
          class="hero"
          :src="readingImageUrl(coverHeroRelPath)"
          alt=""
        />
        <div class="text-flow">
          <div
            v-for="segment in segments"
            :key="segment.segment_id"
            class="sentence-row"
            :class="[{ narrator: isNarrator(segment), active: activeSegmentId === segment.segment_id }]"
            @click="playSingleSegment(segment)"
          >
            <div
              v-if="!isNarrator(segment)"
              class="speaker-icon"
              :style="{ color: speakerIconColor(segment.speaker_id) }"
              :title="speakerLabel(segment.speaker_id)"
              role="img"
              :aria-label="speakerLabel(segment.speaker_id)"
            >
              <Icon name="user" />
            </div>
            <div v-else></div>

            <div>
              <div class="sentence-text">
                <span
                  v-for="(token, tokenIndex) in segment.tokens || []"
                  :key="`${segment.segment_id}-${token.token_idx}`"
                  class="token"
                  :class="{
                    clickable: token.clickable,
                    'word-selected': selectedTokenKey === tokenKey(segment.segment_id, token.token_idx),
                  }"
                  @click.stop="onTokenClick($event, token, segment)"
                >
                  {{ tokenText(segment.tokens || [], tokenIndex) }}
                </span>
              </div>
              <div v-if="segment.text_translation_ru" class="translation" :class="{ hidden: !showTranslation }">
                {{ segment.text_translation_ru }}
              </div>
            </div>

            <button
              v-if="segment.audio_rel_path"
              type="button"
              class="sentence-audio-button"
              :class="{ 'sentence-audio-button--stop': activeSegmentId === segment.segment_id }"
              :aria-label="activeSegmentId === segment.segment_id ? 'Остановить' : 'Озвучить предложение'"
              :title="activeSegmentId === segment.segment_id ? 'Остановить' : 'Озвучить предложение'"
              @click.stop="activeSegmentId === segment.segment_id ? stopPlayback() : playSingleSegment(segment)"
            >
              <Icon :name="activeSegmentId === segment.segment_id ? 'stop' : 'play'" />
            </button>
          </div>
        </div>

        <footer class="footer">
          <button
            type="button"
            class="translation-toggle"
            @click="showTranslation = !showTranslation"
          >
            {{ showTranslation ? t('reading.hideTranslation') : t('reading.showTranslation') }}
          </button>
          <button
            type="button"
            class="mark-read-button"
            :disabled="isRead || markingRead"
            @click="markRead"
          >
            <Icon name="check" />
            <span>{{ isRead ? t('reading.alreadyRead') : t('reading.markRead') }}</span>
          </button>
          <button
            v-if="categoryId && otherUnreadInCategoryCount > 0"
            type="button"
            class="random-unread-footer-button"
            :disabled="randomUnreadNavigating"
            @click="openRandomUnreadInCategory"
          >
            <Icon name="dice" />
            <span>{{ t('reading.anotherRandomUnread') }}</span>
          </button>
        </footer>
      </section>
    </main>

    <div v-if="quizOpen" class="word-modal-overlay" @click.self="closeQuiz">
      <div class="word-modal-panel">
        <h3 style="margin-top:0">{{ t('reading.markRead') }}</h3>
        <p class="translation" style="margin-top:0">{{ t('reading.quizNeedCorrect', { need: passThreshold, total: quizQuestions.length }) }}</p>
        <div v-if="quizQuestions.length" style="margin:12px 0">
          <div class="translation" style="margin-bottom:10px">Вопрос {{ quizIndex + 1 }} из {{ quizQuestions.length }}</div>
          <GrammarQuestion
            :key="`q-${quizIndex}`"
            :question="quizQuestionCurrent"
            :show-answers="quizAnswered[quizIndex] === true"
            :show-explanation="false"
            :show-theory-help-button="false"
            :initial-answer="quizAnswers[quizIndex]"
            @answer="onQuizAnswer"
          />
        </div>
        <div v-if="quizDone" class="feedback-section">
          <div class="feedback-badge" :class="quizCorrectCount >= passThreshold ? 'feedback-success' : 'feedback-error'">
            <span class="feedback-icon"><Icon :name="quizCorrectCount >= passThreshold ? 'check' : 'close'" /></span>
            <span class="feedback-text">{{ t('reading.quizResult', { percent: quizPercent, correct: quizCorrectCount, total: quizQuestions.length }) }}</span>
          </div>
        </div>
        <div v-if="!quizDone && currentWrongExplanation" class="translation" style="margin-top:10px">{{ currentWrongExplanation }}</div>
        <div style="display:flex; gap:8px; margin-top:14px">
          <template v-if="quizDone">
            <button
              v-if="quizCorrectCount >= passThreshold"
              type="button"
              class="word-modal-close-btn"
              :disabled="markingRead"
              @click="backToList"
            >
              {{ t('reading.quizBackToList') }}
            </button>
            <template v-else>
              <button type="button" class="word-modal-close-btn" :disabled="markingRead" @click="retryQuiz">
                {{ t('reading.quizRetry') }}
              </button>
              <button type="button" class="word-modal-close-btn" @click="closeQuiz">{{ t('common.close') }}</button>
            </template>
          </template>
          <button v-else type="button" class="word-modal-close-btn" @click="closeQuiz">{{ t('common.close') }}</button>
        </div>
      </div>
    </div>

    <!-- Полный экран как в «Словарь»: карточки, SRS, формы глаголов; слово в обучение — на сервере при word-lookup -->
    <div v-if="wordModalVisible" class="word-modal-overlay" @click.self="closeWordModal">
      <div class="word-modal-panel">
        <div v-if="wordLookupLoading" class="word-modal-loading">
          {{ wordLookupGenerating ? t('reading.wordGenerating') : t('common.loading') }}
        </div>
        <div v-else-if="wordLookupError" class="word-modal-error">
          <p class="word-modal-error-text">{{ wordLookupError }}</p>
          <button type="button" class="word-modal-close-btn" @click="closeWordModal">{{ t('common.close') }}</button>
        </div>
        <VocabWordCardsDetail
          v-else
          :lemma="modalLemma"
          :preloaded="modalPreloaded"
          @close="closeWordModal"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { getGrammarCourseCode, withCourseCode } from '../api/grammarClient'
import { useAudio } from '../composables/useAudio'
import { useSettings } from '../composables/useSettings'
import Icon from './Icon.vue'
import GrammarQuestion from './GrammarQuestion.vue'
import VocabWordCardsDetail, { type VocabCardsAPIResponse } from './VocabWordCardsDetail.vue'

const props = defineProps<{
  block: any
  chapterId?: string
  textId?: string
  categoryId?: string
  isRead: boolean
  canDelete?: boolean
  deleting?: boolean
  coverHeroRelPath?: string
}>()

const emit = defineEmits<{
  (e: 'marked-read'): void
  (e: 'delete-request'): void
}>()

const { t } = useI18n()
const router = useRouter()
const { settings } = useSettings()
const { playSuccess, playFail } = useAudio()

const showTranslation = ref(false)
const wordModalVisible = ref(false)
const wordLookupLoading = ref(false)
// True once a lookup has been pending longer than typical DB round-trip time — at that point the
// backend is almost certainly generating a brand-new word card via AI rather than just fetching
// an existing one, so we swap the generic "loading" message for one that explains the wait.
const wordLookupGenerating = ref(false)
let wordLookupGeneratingTimer: ReturnType<typeof setTimeout> | null = null
const wordLookupError = ref('')
const modalLemma = ref('')
const modalPreloaded = ref<VocabCardsAPIResponse | null>(null)
const markingRead = ref(false)
const otherUnreadInCategoryCount = ref(0)
const randomUnreadNavigating = ref(false)
const quizOpen = ref(false)
const quizAnswers = ref<Record<number, string>>({})
const quizIndex = ref(0)
const quizAnswered = ref<Record<number, boolean>>({})
const quizCorrectMap = ref<Record<number, boolean>>({})
const quizDone = ref(false)
const quizPercent = ref(0)

const MAX_QUIZ_QUESTIONS = 3

const segments = computed(() => props.block?.reading_passage?.segments || [])
const quizPool = computed(() => {
  const raw = props.block?.reading_passage?.comprehension_questions
  return Array.isArray(raw) ? raw : []
})
// Subset (max 3, random order) chosen each time the quiz opens.
const quizQuestions = ref<any[]>([])
// Strictly more than 50% correct is required to mark the text as read.
const passThreshold = computed(() => {
  const n = quizQuestions.value.length
  if (n <= 0) return 0
  return Math.floor(n / 2) + 1
})

function shuffle<T>(arr: T[]): T[] {
  const a = arr.slice()
  for (let i = a.length - 1; i > 0; i -= 1) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}

function pickQuizQuestions(): any[] {
  return shuffle(quizPool.value).slice(0, MAX_QUIZ_QUESTIONS)
}
const quizQuestionCurrent = computed(() => {
  const q = quizQuestions.value[quizIndex.value] || {}
  if (String(q?.type || '').toLowerCase() === 'true_false' || !Array.isArray(q?.choices) || !q.choices.length) {
    return { ...q, type: 'true_false' as const }
  }
  return { ...q, type: 'mcq_single' as const }
})
const quizCorrectCount = computed(() => Object.values(quizCorrectMap.value).filter(Boolean).length)
const currentWrongExplanation = computed(() => {
  const idx = quizIndex.value
  if (!quizAnswered.value[idx]) return ''
  if (quizCorrectMap.value[idx]) return ''
  const exp = quizQuestions.value[idx]?.explanation
  return typeof exp === 'string' ? exp.trim() : ''
})
const activeSegmentId = ref<string | null>(null)
const selectedTokenKey = ref('')
const isAutoplaying = ref(false)
let currentAudio: HTMLAudioElement | null = null
let currentAudioFinish: (() => void) | null = null
let autoplayRun = 0
let playbackDisposed = false

/** Distinct hues for dialogue speakers (first appearance order). */
const SPEAKER_ICON_PALETTE = [
  '#5b9cff',
  '#ffb020',
  '#3dd68c',
  '#f472b6',
  '#a78bfa',
  '#fb923c',
  '#38bdf8',
  '#f87171',
  '#4ade80',
  '#e879f9',
] as const

const speakerColorById = computed(() => {
  const map: Record<string, string> = {}
  let idx = 0
  for (const seg of segments.value) {
    const id = String(seg?.speaker_id || '').trim().toLowerCase()
    if (!id || id === 'narrator') continue
    if (map[id]) continue
    map[id] = SPEAKER_ICON_PALETTE[idx % SPEAKER_ICON_PALETTE.length]
    idx += 1
  }
  return map
})

function speakerIconColor(speakerId: string): string {
  const id = String(speakerId || '').trim().toLowerCase()
  return speakerColorById.value[id] ?? '#94a3b8'
}

function speakerLabel(id: unknown): string {
  const s = String(id ?? '').trim()
  return s || 'speaker'
}

const noSpaceBefore = new Set(['.', ',', '!', '?', ';', ':', ')', ']', '}', '%', '…'])
const noSpaceAfter = new Set(['(', '[', '{', '¿', '¡', '«'])

const tokenText = (tokens: any[], index: number) => {
  const current = String(tokens?.[index]?.surface ?? '')
  const next = String(tokens?.[index + 1]?.surface ?? '')
  if (!current) return ''
  if (!next) return current
  if (noSpaceAfter.has(current) || noSpaceBefore.has(next)) return current
  return `${current} `
}

const waitMs = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms))

const goBack = async () => {
  const cid = String(props.categoryId || '').trim()
  if (cid) {
    await router.push({ name: 'ReadingChapters', params: { categoryId: cid } })
    return
  }
  await router.push({ name: 'ReadingCategories' })
}

const tokenKey = (segmentId: string, tokenIdx: number) => `${segmentId}-${tokenIdx}`

const readingImageUrl = (relPath: string) => {
  const courseCode = getGrammarCourseCode()
  const courseParam = courseCode ? `&course_code=${encodeURIComponent(courseCode)}` : ''
  return `/api/learning/reading/image?path=${encodeURIComponent(relPath)}${courseParam}`
}

const isNarrator = (segment: any) => String(segment?.speaker_id || '').toLowerCase() === 'narrator'

const stopCurrentAudio = (clearActive = true) => {
  const audio = currentAudio
  if (audio) {
    currentAudio = null
    audio.pause()
    audio.currentTime = 0
  }
  const finish = currentAudioFinish
  currentAudioFinish = null
  finish?.()
  if (clearActive) {
    activeSegmentId.value = null
  }
}

const stopPlayback = () => {
  isAutoplaying.value = false
  autoplayRun += 1
  stopCurrentAudio()
}

const playSegmentAudio = async (audioRelPath: string) => {
  const courseCode = getGrammarCourseCode()
  const courseParam = courseCode ? `&course_code=${encodeURIComponent(courseCode)}` : ''
  const url = `/api/learning/reading/audio?path=${encodeURIComponent(audioRelPath)}${courseParam}`
  stopCurrentAudio(false)
  const audio = new Audio(url)
  currentAudio = audio
  await new Promise<void>((resolve) => {
    const finish = () => {
      audio.removeEventListener('ended', finish)
      audio.removeEventListener('error', finish)
      if (currentAudio === audio) currentAudio = null
      if (currentAudioFinish === finish) currentAudioFinish = null
      resolve()
    }
    currentAudioFinish = finish
    audio.addEventListener('ended', finish, { once: true })
    audio.addEventListener('error', finish, { once: true })
    audio.play().catch((error) => {
      console.error('Failed to play segment audio', error)
      finish()
    })
  })
}

const playSingleSegment = async (segment: any) => {
  if (!segment?.audio_rel_path) return
  isAutoplaying.value = false
  const runId = ++autoplayRun
  activeSegmentId.value = segment.segment_id
  await playSegmentAudio(segment.audio_rel_path)
  if (runId === autoplayRun && !isAutoplaying.value) {
    activeSegmentId.value = null
  }
}

const toggleAutoplay = async () => {
  if (isAutoplaying.value) {
    stopPlayback()
    return
  }

  const runId = ++autoplayRun
  isAutoplaying.value = true
  for (const segment of segments.value) {
    if (!isAutoplaying.value || runId !== autoplayRun) break
    if (!segment?.audio_rel_path) continue
    activeSegmentId.value = segment.segment_id
    await playSegmentAudio(segment.audio_rel_path)
    if (!isAutoplaying.value || runId !== autoplayRun) break
    await waitMs(500)
  }
  if (runId === autoplayRun) {
    isAutoplaying.value = false
    activeSegmentId.value = null
  }
}

const closeWordModal = () => {
  wordModalVisible.value = false
  wordLookupLoading.value = false
  wordLookupError.value = ''
  modalLemma.value = ''
  modalPreloaded.value = null
  selectedTokenKey.value = ''
}

const wordModalKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && wordModalVisible.value) {
    event.preventDefault()
    closeWordModal()
  }
}

watch(wordModalVisible, (open) => {
  if (open) {
    window.addEventListener('keydown', wordModalKeydown)
  } else {
    window.removeEventListener('keydown', wordModalKeydown)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', wordModalKeydown)
  if (wordLookupGeneratingTimer) clearTimeout(wordLookupGeneratingTimer)
})

const onTokenClick = async (event: MouseEvent, token: any, segment: any) => {
  if (!token?.clickable || !token?.lemma) return

  selectedTokenKey.value = tokenKey(segment.segment_id, token.token_idx)

  wordModalVisible.value = true
  wordLookupLoading.value = true
  wordLookupGenerating.value = false
  wordLookupError.value = ''
  modalPreloaded.value = null
  modalLemma.value = token.lemma

  if (wordLookupGeneratingTimer) clearTimeout(wordLookupGeneratingTimer)
  wordLookupGeneratingTimer = setTimeout(() => {
    wordLookupGenerating.value = true
  }, 600)

  try {
    const data: VocabCardsAPIResponse = await apiClient.request(
      withCourseCode(`/api/reading/word-lookup?lemma=${encodeURIComponent(token.lemma)}`),
    )
    modalLemma.value = data.lemma || token.lemma
    modalPreloaded.value = data
  } catch (error: any) {
    console.error('Word lookup failed', error)
    modalPreloaded.value = null
    const status = typeof error?.status === 'number' ? error.status : 0
    if (error?.isNetworkError) {
      wordLookupError.value = t('reading.wordLookupNetwork')
    } else if (status === 404) {
      wordLookupError.value = t('reading.wordNotFound')
    } else if (status >= 500) {
      wordLookupError.value = t('reading.wordLookupServerError')
    } else {
      wordLookupError.value = t('reading.wordLookupFailed')
    }
  } finally {
    if (wordLookupGeneratingTimer) {
      clearTimeout(wordLookupGeneratingTimer)
      wordLookupGeneratingTimer = null
    }
    wordLookupLoading.value = false
    wordLookupGenerating.value = false
  }
}

const loadAutoplayPreference = async () => {
  let enabled = !!settings.value.autoplayPronunciation
  try {
    const data = await apiClient.request<{ settings?: { autoplay_pronunciation?: boolean } }>('/api/settings')
    if (typeof data?.settings?.autoplay_pronunciation === 'boolean') {
      enabled = data.settings.autoplay_pronunciation
    }
  } catch (error) {
    console.error('Failed to load autoplay setting for reading mode:', error)
  }
  if (playbackDisposed) return
  if (enabled && segments.value.length > 0) {
    void toggleAutoplay()
  }
}

onMounted(() => {
  playbackDisposed = false
  void loadAutoplayPreference()
})

onBeforeUnmount(() => {
  playbackDisposed = true
  stopPlayback()
})

const markRead = async () => {
  if (!isReadQuestionGateEnabled()) {
    await markReadDirect()
    return
  }
  quizQuestions.value = pickQuizQuestions()
  resetQuizState()
  quizOpen.value = true
}

function resetQuizState() {
  quizAnswers.value = {}
  quizIndex.value = 0
  quizAnswered.value = {}
  quizCorrectMap.value = {}
  quizDone.value = false
  quizPercent.value = 0
}

function retryQuiz() {
  if (markingRead.value) return
  quizQuestions.value = pickQuizQuestions()
  resetQuizState()
}

const closeQuiz = () => {
  if (markingRead.value) return
  quizOpen.value = false
}

function isReadQuestionGateEnabled(): boolean {
  return quizPool.value.length > 0
}

function normalizeAnswer(v: unknown): string {
  return String(v ?? '').trim().toLowerCase()
}
function onQuizAnswer(answer: any) {
  const idx = quizIndex.value
  const got = normalizeAnswer(answer)
  const want = normalizeAnswer(quizQuestions.value[idx]?.correct_answer)
  const ok = !!got && !!want && got === want
  quizAnswers.value[idx] = got
  quizAnswered.value[idx] = true
  quizCorrectMap.value[idx] = ok
  triggerHapticFeedback(ok)
  if (ok) playCorrectSound()
  else playIncorrectSound()
  setTimeout(() => {
    if (quizDone.value || !quizOpen.value) return
    if (idx !== quizIndex.value) return
    // Advance regardless of correctness; auto-finish after the last question.
    if (idx < quizQuestions.value.length - 1) {
      quizIndex.value += 1
    } else {
      void finishQuiz()
    }
  }, ok ? 700 : 1000)
}

const finishQuiz = async () => {
  const total = quizQuestions.value.length
  const correct = quizCorrectCount.value
  quizDone.value = true
  quizPercent.value = total > 0 ? Math.round((correct * 100) / total) : 0
  const passed = correct >= passThreshold.value
  triggerHapticFeedback(passed)
  // Pass = strictly more than 50% correct -> mark read immediately.
  if (passed && !props.isRead) {
    await markReadDirect()
  }
}

const backToList = async () => {
  quizOpen.value = false
  await goBack()
}

const playCorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playSuccess(settings.value.soundTheme)
}
const playIncorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playFail(settings.value.soundTheme)
}
const triggerHapticFeedback = (isCorrect: boolean) => {
  if (!settings.value.vibrationEnabled) return
  const tg = (window as any).Telegram?.WebApp
  if (tg?.HapticFeedback?.notificationOccurred) {
    try { tg.HapticFeedback.notificationOccurred(isCorrect ? 'success' : 'error'); return } catch {}
  }
  if ('vibrate' in navigator && typeof navigator.vibrate === 'function') navigator.vibrate(isCorrect ? 50 : [100, 50, 100])
}

const markReadDirect = async () => {
  markingRead.value = true
  try {
    const resourceId = props.textId || props.chapterId
    if (!resourceId) throw new Error('text id is required')
    await apiClient.request(`/api/learning/reading/texts/${resourceId}/mark-read`, { method: 'POST' })
    emit('marked-read')
  } catch (error) {
    console.error('Failed to mark reading text as read', error)
  } finally {
    markingRead.value = false
  }
}

const currentTextId = computed(() => String(props.textId || props.chapterId || '').trim())

async function refreshOtherUnreadInCategory() {
  const cat = String(props.categoryId || '').trim()
  const tid = currentTextId.value
  otherUnreadInCategoryCount.value = 0
  if (!cat || !tid) return
  try {
    const data: { texts?: { text_id: string; is_read: boolean }[] } = await apiClient.request(
      `/api/learning/reading/categories/${encodeURIComponent(cat)}/texts`
    )
    const texts = data.texts || []
    otherUnreadInCategoryCount.value = texts.filter((x) => !x.is_read && x.text_id !== tid).length
  } catch {
    otherUnreadInCategoryCount.value = 0
  }
}

watch(
  () => [props.categoryId, props.textId, props.chapterId, props.isRead] as const,
  () => {
    void refreshOtherUnreadInCategory()
  },
  { immediate: true }
)

const openRandomUnreadInCategory = async () => {
  const cat = String(props.categoryId || '').trim()
  const tid = currentTextId.value
  if (!cat || !tid || randomUnreadNavigating.value) return
  randomUnreadNavigating.value = true
  try {
    const data: { texts?: { text_id: string; is_read: boolean }[] } = await apiClient.request(
      `/api/learning/reading/categories/${encodeURIComponent(cat)}/texts`
    )
    const pool = (data.texts || []).filter((x) => !x.is_read && x.text_id !== tid)
    if (!pool.length) {
      otherUnreadInCategoryCount.value = 0
      return
    }
    const pick = pool[Math.floor(Math.random() * pool.length)]
    await router.push(`/learning/reading/text/${pick.text_id}`)
  } finally {
    randomUnreadNavigating.value = false
  }
}
</script>

<style scoped>
.screen {
  min-height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  display: flex;
  justify-content: center;
  padding: 16px;
}

.reader-shell {
  width: 100%;
  max-width: 960px;
  min-height: calc(100vh - 32px);
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 24px 80px var(--card-shadow);
}

.header {
  min-height: 84px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-secondary);
}

.back-button {
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border: 0;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 22px;
}

.title {
  flex: 1;
  font-size: 22px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-button {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: grid;
  place-items: center;
  cursor: pointer;
  font-size: 16px;
}

.icon-button--stop {
  background: color-mix(in srgb, var(--error, #C04A2B) 16%, var(--bg-tertiary));
  color: var(--error, #C04A2B);
  border-color: color-mix(in srgb, var(--error, #C04A2B) 45%, var(--border-primary));
}

.icon-button:hover {
  background: var(--bg-hover);
}

.icon-button--stop:hover {
  background: color-mix(in srgb, var(--error, #C04A2B) 22%, var(--bg-hover));
}

.icon-button-danger {
  color: #ef4444;
  border-color: color-mix(in srgb, #ef4444 40%, var(--border-primary));
}

.icon-button-danger:hover:not(:disabled) {
  background: color-mix(in srgb, #ef4444 16%, var(--bg-hover));
}

.icon-button:disabled {
  opacity: 0.55;
  cursor: default;
}

.content {
  padding: 22px 16px 24px;
  position: relative;
}

.hero {
  display: block;
  width: 100%;
  height: auto;
  max-height: 320px;
  margin: 0 0 18px;
  border-radius: 14px;
  object-fit: contain;
  object-position: center;
  background: var(--bg-tertiary, rgba(0, 0, 0, 0.04));
}

.text-flow {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sentence-row {
  position: relative;
  display: grid;
  grid-template-columns: 30px 1fr 40px;
  column-gap: 8px;
  align-items: start;
  min-height: 48px;
  padding: 8px 8px;
  border-radius: 14px;
  cursor: pointer;
}

.sentence-row.narrator {
  grid-template-columns: 0 1fr 40px;
}

.sentence-row:hover {
  background: var(--bg-hover);
}

.sentence-row.active {
  background: linear-gradient(90deg, rgba(44, 116, 255, 0.2), rgba(44, 116, 255, 0.08));
  box-shadow: inset 0 0 0 1px rgba(84, 145, 255, 0.12);
}

.speaker-icon {
  width: 24px;
  height: 24px;
  margin-top: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 22px;
  opacity: 1;
}

.sentence-text {
  font-size: 21px;
  line-height: 1.28;
  font-weight: 500;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.translation {
  margin-top: 6px;
  font-size: 14px;
  line-height: 1.35;
  font-weight: 400;
  color: var(--text-secondary);
}

.translation.hidden {
  display: none;
}

.token {
  white-space: pre-wrap;
}

.token.clickable {
  cursor: pointer;
}

.word-selected {
  background: #3b82f6;
  color: #ffffff;
  border-radius: 6px;
  padding: 1px 4px 3px;
}

.sentence-audio-button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  display: grid;
  place-items: center;
  margin-top: 2px;
  cursor: pointer;
}

.sentence-audio-button--stop {
  background: var(--error, #C04A2B);
  color: #fff;
  border-color: transparent;
  box-shadow: 0 8px 18px color-mix(in srgb, var(--error, #C04A2B) 28%, transparent);
}

.sentence-audio-button--stop:hover {
  background: color-mix(in srgb, var(--error, #C04A2B) 86%, #000);
}

.word-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: var(--bg-modal-overlay, rgba(0, 0, 0, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.word-modal-panel {
  background: var(--card-bg);
  border-radius: 8px;
  max-width: 800px;
  width: 100%;
  max-height: min(90vh, 900px);
  overflow-y: auto;
  padding: 24px 28px;
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
  box-shadow: 0 24px 64px var(--card-shadow);
}

.word-modal-loading {
  text-align: center;
  padding: 48px 16px;
  font-size: 16px;
  color: var(--text-secondary);
}

.word-modal-error {
  padding: 24px 8px;
  text-align: center;
}

.word-modal-error-text {
  margin: 0 0 20px;
  font-size: 15px;
  line-height: 1.45;
  color: var(--text-secondary);
}

.word-modal-close-btn {
  padding: 10px 20px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.word-modal-close-btn:hover {
  background: var(--bg-hover);
}

.footer {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.translation-toggle {
  align-self: center;
  border: 0;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.2;
  padding: 4px 8px;
  cursor: pointer;
  text-decoration: none;
}

.translation-toggle:hover {
  color: var(--text-primary);
}

.mark-read-button {
  width: 100%;
  min-height: 58px;
  border-radius: 12px;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.mark-read-button:hover {
  background: var(--bg-hover);
}

.mark-read-button:disabled {
  opacity: 0.6;
  cursor: default;
}

.random-unread-footer-button {
  width: 100%;
  min-height: 58px;
  border-radius: 12px;
  border: 1px dashed var(--border-primary);
  background: var(--bg-secondary);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.random-unread-footer-button:hover:not(:disabled) {
  background: var(--bg-hover);
}

.random-unread-footer-button:disabled {
  opacity: 0.6;
  cursor: default;
}

.random-unread-footer-button :deep(.icon) {
  font-size: 22px;
}

.feedback-section {
  margin-top: 12px;
}
.feedback-badge {
  display: flex;
  align-items: center;
  gap: 10px;
  border-radius: 12px;
  padding: 12px 14px;
  border: 1px solid var(--border-primary);
}
.feedback-success {
  background: color-mix(in srgb, #10b981 18%, var(--bg-tertiary));
  animation: feedback-success-appear 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.feedback-error {
  background: color-mix(in srgb, #ef4444 14%, var(--bg-tertiary));
  animation: feedback-error-appear 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}
.feedback-icon {
  font-size: 20px;
  font-weight: 700;
}
.feedback-text {
  font-size: 16px;
  font-weight: 600;
}
@keyframes feedback-success-appear {
  from { transform: scale(0.92); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
@keyframes feedback-error-appear {
  from { transform: translateX(-8px); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

@media (max-width: 768px) {
  .screen {
    padding: 8px;
  }

  .reader-shell {
    min-height: calc(100vh - 16px);
    border-radius: 16px;
  }

  .header {
    min-height: 74px;
    padding: 0 12px;
    gap: 10px;
  }

  .title {
    font-size: 18px;
  }

  .icon-button {
    width: 40px;
    height: 40px;
  }

  .content {
    padding: 14px 10px 20px;
  }

  .sentence-text {
    font-size: 19px;
  }

  .translation {
    font-size: 13px;
  }
}
</style>
