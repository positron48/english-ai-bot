<template>
  <div class="grammar-training">
    <div v-if="loading" class="loading">{{ t('common.loading') }}</div>

    <div v-else-if="error" class="card">
      <p>{{ error }}</p>
      <button class="btn btn-primary" @click="init">{{ t('common.retry') }}</button>
    </div>

    <div v-else-if="!available" class="card">
      <h1>{{ t('learning.grammar') }}</h1>
      <p>{{ t('grammar.trainingUnavailable') }}</p>
      <router-link class="btn btn-primary" to="/learning/grammar">{{ t('common.back') }}</router-link>
    </div>

    <div v-else-if="sessionDone" class="grammar-training-done">
      <h1 class="grammar-done-title">{{ t('grammar.trainingTitle') }}</h1>
      <TrainingSessionCompletion
        :total-cards="totalCount"
        :correct-cards="correctCount"
        :stats-loaded="true"
        :available-for-training="0"
        :estimated-time-for-remaining="null"
        :continue-label="t('training.startTraining')"
        :show-continue-button="true"
        :show-continue-without-due-cards="true"
        :sounds-enabled="settings.soundsEnabled"
        :sound-theme="settings.soundTheme"
        @continue="startSession"
      />
      <div class="grammar-done-back-row">
        <router-link class="btn btn-secondary grammar-done-back-link" to="/learning/grammar">
          {{ t('common.back') }}
        </router-link>
      </div>
    </div>

    <div v-else-if="currentQuestion" class="question-stage">
      <div class="card">
        <div class="header">
          <h1>{{ t('grammar.trainingTitle') }}</h1>
          <div class="progress">{{ currentIndex + 1 }} / {{ totalCount }}</div>
        </div>

      <GrammarQuestion
        ref="grammarQuestionRef"
        :key="`${currentQuestion.id}-${currentIndex}`"
        :question="currentQuestion"
        :theory-block="currentTheoryBlock"
        :theory-chapter-context="currentChapterContext"
        :show-answers="!!result"
        :show-explanation="false"
        :show-theory-help-button="false"
        @answer="onAnswer"
      />

        <div v-if="result" class="feedback-section">
        <div v-if="result.correct" class="feedback-badge feedback-success">
          <span class="feedback-icon">✓</span>
          <span class="feedback-text">{{ feedbackHeadline }}</span>
        </div>
        <div v-else class="feedback-badge feedback-error">
          <span class="feedback-icon">✗</span>
          <span class="feedback-text">{{ feedbackHeadline }}</span>
        </div>
        <div v-if="result.explanation" class="feedback-explanation">{{ result.explanation }}</div>

        <div
          v-if="waitingCorrectDelay && result.correct"
          class="waiting-progress"
        >
          <div class="circular-progress">
            <svg class="progress-ring" width="80" height="80">
              <circle
                class="progress-ring-circle-bg"
                stroke="var(--bg-secondary, rgba(0, 0, 0, 0.1))"
                stroke-width="6"
                fill="transparent"
                r="34"
                cx="40"
                cy="40"
              />
              <circle
                class="progress-ring-circle"
                stroke="var(--color-primary)"
                stroke-width="6"
                fill="transparent"
                r="34"
                cx="40"
                cy="40"
                :style="{ strokeDasharray: delayCircumference, strokeDashoffset: strokeDashoffset }"
              />
            </svg>
            <div class="progress-text">{{ delaySeconds }}</div>
          </div>
        </div>
      </div>

        <div class="actions footer-actions">
          <button class="btn btn-primary" :disabled="!result" @click="nextQuestion">
            {{ currentIndex + 1 >= totalCount ? t('common.finish') : t('common.next') }}
          </button>
          <button
            v-if="hasTheoryForCurrentCard"
            type="button"
            class="theory-help-footer-btn"
            :title="t('grammar.theoryBlock')"
            :aria-label="t('grammar.theoryBlock')"
            @click="grammarQuestionRef?.toggleTheoryHelp()"
          >
            <span class="theory-help-footer-icon">i</span>
          </button>
        </div>
      </div>

      <div class="report-footer">
        <button
          v-if="!reportAlreadySent"
          type="button"
          class="report-text-link"
          :disabled="reportSubmitting || !currentQuestion"
          @click="openGrammarReportDialog"
        >
          {{ t('training.reportIssue') }}
        </button>
        <span v-if="reportMessage" class="report-message">{{ reportMessage }}</span>
      </div>
      <div v-if="reportDialogOpen" class="report-modal-backdrop" @click.self="closeGrammarReportDialog">
        <div class="report-modal">
          <h3 class="report-modal-title">{{ t('training.reportIssue') }}</h3>
          <textarea
            v-model.trim="reportComment"
            class="report-modal-textarea"
            :placeholder="t('training.reportCommentPlaceholder') || 'Опишите, что не так с вопросом'"
            rows="5"
            maxlength="1000"
          />
          <div class="report-modal-actions">
            <button type="button" class="report-modal-cancel" @click="closeGrammarReportDialog">
              {{ t('common.cancel') || 'Отмена' }}
            </button>
            <button type="button" class="report-modal-submit" :disabled="reportSubmitting || !reportComment" @click="reportCurrentQuestion">
              {{ t('training.reportSend') || 'Отправить' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import TrainingSessionCompletion from '../components/TrainingSessionCompletion.vue'
import { useSettings } from '../composables/useSettings'

const { settings } = useSettings()

const { t, tm } = useI18n()

function phraseList(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.filter((x): x is string => typeof x === 'string' && x.length > 0)
}

const encouragingPhrasesList = computed(() => phraseList('trainingFeedback.encouragingPhrases'))
const disappointingPhrasesList = computed(() => phraseList('trainingFeedback.disappointingPhrases'))

const generateWeights = (phrases: string[]) => {
  const n = phrases.length
  if (n <= 0) return []
  if (n === 1) return [100]
  const weights: number[] = []
  const maxWeight = 30
  const minWeight = 0.01
  for (let i = 0; i < n; i++) {
    const ratio = i / (n - 1)
    const weight = maxWeight * Math.pow(minWeight / maxWeight, ratio)
    weights.push(weight)
  }
  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map((w) => (w * 100) / sum)
}

const generateCumulativeWeights = (weights: number[]) => {
  const cumulative: number[] = []
  let sum = 0
  for (const weight of weights) {
    sum += weight
    cumulative.push(sum)
  }
  return cumulative
}

const pickWeightedRandom = (phrases: string[]): string => {
  if (!phrases.length) return ''
  if (phrases.length === 1) return phrases[0]
  const weights = generateWeights(phrases)
  const cumulative = generateCumulativeWeights(weights)
  const random = Math.random() * 100
  for (let i = 0; i < cumulative.length; i++) {
    if (random <= cumulative[i]) {
      return phrases[i]
    }
  }
  return phrases[0]
}

const getRandomEncouragingPhrase = () => pickWeightedRandom(encouragingPhrasesList.value)
const getRandomDisappointingPhrase = () => pickWeightedRandom(disappointingPhrasesList.value)

/** Same random headlines as word training (trainingFeedback.*) */
const feedbackHeadline = ref('')

type GrammarQuestionExpose = {
  toggleTheoryHelp: () => void
  closeTheoryHelp: () => void
}

const grammarQuestionRef = ref<GrammarQuestionExpose | null>(null)

const loading = ref(true)
const error = ref<string | null>(null)
const available = ref(false)
const sessionQuestions = ref<any[]>([])
const theoryBlockMap = ref<Record<string, any>>({})
/** Category/chapter/level labels for theory modal (from chapter API + section) */
const chapterContextMap = ref<
  Record<
    string,
    {
      categoryTitle: string
      categoryTitleTranslations?: Record<string, string>
      chapterTitle: string
      chapterTitleTranslations?: Record<string, string>
      level: string
    }
  >
>({})
const currentIndex = ref(0)
const result = ref<any | null>(null)
const correctCount = ref(0)

const currentQuestion = computed(() => sessionQuestions.value[currentIndex.value] || null)
const currentTheoryBlock = computed(() => {
  const q = currentQuestion.value
  if (!q?.chapter_id || !q?.theory_block_id) return null
  return theoryBlockMap.value[`${q.chapter_id}::${q.theory_block_id}`] || null
})

const currentChapterContext = computed(() => {
  const q = currentQuestion.value
  if (!q?.chapter_id) return null
  return chapterContextMap.value[q.chapter_id] ?? null
})

function cacheChapterContextFromApi(chapterId: string, data: any) {
  const ch = data?.chapter
  const sec = data?.section
  const displayChapterTitle = typeof data?.title === 'string' && data.title.length > 0 ? data.title : (ch?.title ?? '')
  const level =
    (typeof ch?.level === 'string' && ch.level.length > 0 ? ch.level : '') ||
    (typeof sec?.level === 'string' && sec.level.length > 0 ? sec.level : '')
  chapterContextMap.value = {
    ...chapterContextMap.value,
    [chapterId]: {
      categoryTitle: typeof sec?.title === 'string' ? sec.title : '',
      categoryTitleTranslations:
        sec?.title_translations && typeof sec.title_translations === 'object' ? sec.title_translations : undefined,
      chapterTitle: displayChapterTitle,
      chapterTitleTranslations:
        ch?.title_translations && typeof ch.title_translations === 'object' ? ch.title_translations : undefined,
      level
    }
  }
}
const totalCount = computed(() => sessionQuestions.value.length)
const sessionDone = computed(() => !loading.value && available.value && totalCount.value > 0 && currentIndex.value >= totalCount.value)

/** Same setting as word training: pause before answer options; used here as delay before auto-advance after a correct answer */
const optionsDelaySeconds = ref(5)

const hasTheoryForCurrentCard = computed(() => {
  const q = currentQuestion.value
  return typeof q?.theory_block_id === 'string' && q.theory_block_id.length > 0
})

let correctAdvanceTimer: ReturnType<typeof setTimeout> | null = null
let correctCountdownRaf: number | null = null
let correctCountdownEnd: number | null = null

/** Обратный отсчёт до автоперехода после верного ответа (как в тренировке слов) */
const waitingCorrectDelay = ref(false)
const remainingMs = ref(0)
const initialDelayMs = ref(0)
const delaySeconds = ref(0)
const reportSubmitting = ref(false)
const reportMessage = ref('')
const reportSentForQuestionID = ref('')
const reportDialogOpen = ref(false)
const reportComment = ref('')
const reportAlreadySent = computed(() => {
  const qid = currentQuestion.value?.id
  return !!qid && reportSentForQuestionID.value === qid
})

const delayCircumference = computed(() => {
  const radius = 34
  return 2 * Math.PI * radius
})

const strokeDashoffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0) {
    return delayCircumference.value
  }
  const progress = remainingMs.value / initialDelayMs.value
  return delayCircumference.value * (1 - progress)
})

const clearCorrectAdvanceTimer = () => {
  if (correctAdvanceTimer) {
    clearTimeout(correctAdvanceTimer)
    correctAdvanceTimer = null
  }
  if (correctCountdownRaf != null) {
    cancelAnimationFrame(correctCountdownRaf)
    correctCountdownRaf = null
  }
  correctCountdownEnd = null
  waitingCorrectDelay.value = false
  remainingMs.value = 0
  initialDelayMs.value = 0
  delaySeconds.value = 0
}

const loadTrainingDelaySetting = async () => {
  try {
    const data = await apiClient.request<{ settings?: { options_delay_seconds?: number } }>('/api/settings')
    const v = data.settings?.options_delay_seconds
    if (typeof v === 'number' && !Number.isNaN(v)) {
      optionsDelaySeconds.value = Math.max(0, Math.min(10, v))
    }
  } catch {
    // keep default
  }
}

const init = async () => {
  loading.value = true
  error.value = null
  try {
    await loadTrainingDelaySetting()
    const data: any = await apiClient.request('/api/learning/grammar/training/availability')
    available.value = !!data?.grammar_training?.available
    if (available.value) {
      await startSession()
    }
  } catch (e: any) {
    error.value = e?.message || 'Failed to load grammar training'
  } finally {
    loading.value = false
  }
}

const startSession = async () => {
  clearCorrectAdvanceTimer()
  result.value = null
  currentIndex.value = 0
  correctCount.value = 0
  theoryBlockMap.value = {}
  chapterContextMap.value = {}
  const data: any = await apiClient.request('/api/learning/grammar/training/session/start', {
    method: 'POST',
    body: JSON.stringify({ limit: 20 })
  })
  sessionQuestions.value = (data?.items || []).map((it: any) => it.question).filter(Boolean)
  await hydrateTheoryBlocksForSession()
}

const hydrateTheoryBlocksForSession = async () => {
  const chapterIds = Array.from(new Set(
    sessionQuestions.value
      .map((q: any) => q?.chapter_id)
      .filter((id: any) => typeof id === 'string' && id.length > 0)
  ))
  if (chapterIds.length === 0) return

  for (const chapterId of chapterIds) {
    try {
      const data: any = await apiClient.request(`/api/learning/grammar/chapters/${chapterId}`)
      cacheChapterContextFromApi(chapterId, data)
      const blocks = data?.chapter?.blocks
      if (!Array.isArray(blocks)) continue
      for (const block of blocks) {
        if (block?.type !== 'theory' || !block?.id) continue
        const key = `${chapterId}::${block.id}`
        theoryBlockMap.value[key] = block
      }
    } catch {
      // Не блокируем тренировку, если по главе не удалось получить теорию.
    }
  }
}

const onAnswer = async (answer: any) => {
  if (!currentQuestion.value || result.value) return
  try {
    const data = await apiClient.request('/api/learning/grammar/training/session/answer', {
      method: 'POST',
      body: JSON.stringify({ question_id: currentQuestion.value.id, answer })
    })
    result.value = data
    if (data?.correct) correctCount.value++

    if (data?.correct) {
      feedbackHeadline.value = getRandomEncouragingPhrase() || t('grammar.correct')
    } else {
      feedbackHeadline.value = getRandomDisappointingPhrase() || t('grammar.wrong')
    }

    clearCorrectAdvanceTimer()
    if (data?.correct) {
      const ms = optionsDelaySeconds.value * 1000
      if (ms <= 0) {
        advanceAfterCorrect()
        return
      }
      initialDelayMs.value = ms
      remainingMs.value = ms
      delaySeconds.value = Math.ceil(ms / 1000)
      waitingCorrectDelay.value = true
      correctCountdownEnd = Date.now() + ms

      const tick = () => {
        if (correctCountdownEnd === null) return
        const left = Math.max(0, correctCountdownEnd - Date.now())
        remainingMs.value = left
        delaySeconds.value = Math.max(0, Math.ceil(left / 1000))
        if (left > 0) {
          correctCountdownRaf = requestAnimationFrame(tick)
        } else {
          correctCountdownRaf = null
        }
      }
      correctCountdownRaf = requestAnimationFrame(tick)

      correctAdvanceTimer = setTimeout(() => {
        correctAdvanceTimer = null
        clearCorrectAdvanceTimer()
        advanceAfterCorrect()
      }, ms)
    }
  } catch (e: any) {
    error.value = e?.message || 'Failed to submit answer'
  }
}

/** Next card after correct answer without clearing manual-next path */
const advanceAfterCorrect = () => {
  result.value = null
  currentIndex.value++
}

const nextQuestion = () => {
  clearCorrectAdvanceTimer()
  result.value = null
  currentIndex.value++
}

const reportCurrentQuestion = async () => {
  if (!currentQuestion.value || reportSubmitting.value || !reportComment.value) return
  reportMessage.value = ''
  reportSubmitting.value = true
  try {
    await apiClient.request('/api/learning/grammar/training/report', {
      method: 'POST',
      body: JSON.stringify({
        question_id: currentQuestion.value.id,
        chapter_id: currentQuestion.value.chapter_id || '',
        theory_block_id: currentQuestion.value.theory_block_id || '',
        comment: reportComment.value,
        question_data: currentQuestion.value
      })
    })
    reportSentForQuestionID.value = currentQuestion.value.id
    reportMessage.value = t('training.reportThanks')
    reportDialogOpen.value = false
    reportComment.value = ''
  } catch (e) {
    console.error('Failed to report grammar question:', e)
    reportMessage.value = t('training.reportFailed')
  } finally {
    reportSubmitting.value = false
  }
}

const openGrammarReportDialog = () => {
  if (!currentQuestion.value || reportSubmitting.value || reportAlreadySent.value) return
  reportComment.value = ''
  reportDialogOpen.value = true
}

const closeGrammarReportDialog = () => {
  if (reportSubmitting.value) return
  reportDialogOpen.value = false
}

onMounted(init)

onBeforeUnmount(() => {
  clearCorrectAdvanceTimer()
})

watch(currentQuestion, async (q) => {
  reportMessage.value = ''
  reportDialogOpen.value = false
  reportComment.value = ''
  if (!q?.chapter_id || !q?.theory_block_id) return
  const key = `${q.chapter_id}::${q.theory_block_id}`
  if (theoryBlockMap.value[key]) return
  try {
    const data: any = await apiClient.request(`/api/learning/grammar/chapters/${q.chapter_id}`)
    cacheChapterContextFromApi(q.chapter_id, data)
    const blocks = data?.chapter?.blocks
    if (!Array.isArray(blocks)) return
    for (const block of blocks) {
      if (block?.type !== 'theory' || !block?.id) continue
      theoryBlockMap.value[`${q.chapter_id}::${block.id}`] = block
    }
  } catch {
    // best-effort
  }
})
</script>

<style scoped>
.grammar-training { max-width: 980px; margin: 0 auto; padding: 20px; }

.grammar-training-done {
  max-width: 720px;
  margin: 0 auto;
  padding: 12px 0 24px;
}

.grammar-done-title {
  text-align: center;
  margin: 0 0 20px;
  font-size: 1.5rem;
  color: var(--text-primary);
}

.grammar-done-back-row {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.grammar-done-back-link {
  text-decoration: none;
  min-width: 200px;
  text-align: center;
}
.question-stage {
  display: flex;
  flex-direction: column;
}
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.progress { color: var(--text-secondary); font-weight: 600; }
.actions { margin-top: 16px; display: flex; gap: 10px; }
.actions.footer-actions {
  justify-content: flex-start;
  align-items: center;
  flex-wrap: wrap;
  width: 100%;
  gap: 0;
  padding-left: 0;
  margin-left: 0;
  box-sizing: border-box;
}

.actions.footer-actions .theory-help-footer-btn {
  margin-left: auto;
}

.theory-help-footer-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  width: 40px;
  height: 40px;
  border: 1px solid var(--border-primary);
  border-radius: 50%;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
  flex-shrink: 0;
}

.theory-help-footer-btn:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.theory-help-footer-icon {
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
  color: var(--text-secondary);
}
.feedback-section {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
}

.report-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px 10px;
  font-size: 12px;
}

.report-btn:hover {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.report-btn.subtle {
  opacity: 0.75;
}

.report-btn-icon {
  font-weight: 700;
}
.report-footer {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 6px;
}
.report-message {
  font-size: 12px;
  color: var(--text-secondary);
}
.report-text-link {
  border: 0;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.2;
  padding: 0;
  margin: 0;
  cursor: pointer;
  text-decoration: none;
}
.report-text-link:hover:not(:disabled) {
  color: var(--text-primary);
}
.report-text-link:disabled {
  cursor: default;
  opacity: 0.8;
}

.report-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.report-modal {
  width: min(92vw, 460px);
  background: var(--background-color);
  color: var(--text-primary);
  border-radius: 12px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
  padding: 16px;
}

.report-modal-title {
  margin: 0 0 10px 0;
  font-size: 16px;
}

.report-modal-textarea {
  width: 100%;
  min-height: 110px;
  resize: vertical;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px;
  font: inherit;
  color: var(--text-primary);
  background: var(--background-color);
}

.report-modal-actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.report-modal-cancel,
.report-modal-submit {
  border: 0;
  border-radius: 8px;
  padding: 8px 12px;
  font: inherit;
  cursor: pointer;
}

.report-modal-cancel {
  background: var(--button-secondary-bg);
  color: var(--button-secondary-text);
}

.report-modal-submit {
  background: var(--button-primary-bg);
  color: var(--button-primary-text);
}

.report-modal-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
.feedback-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-radius: 10px;
  font-weight: 700;
  margin-bottom: 10px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}
.feedback-success {
  background: rgba(40, 167, 69, 0.12);
  border: 1px solid rgba(40, 167, 69, 0.45);
  color: var(--color-success);
}
.feedback-error {
  background: rgba(220, 53, 69, 0.12);
  border: 1px solid rgba(220, 53, 69, 0.45);
  color: var(--color-danger);
}
.feedback-icon {
  font-size: 20px;
  font-weight: 800;
  line-height: 1;
}
.feedback-text {
  line-height: 1.2;
}
.feedback-explanation {
  margin-top: 8px;
  padding: 12px 14px;
  border-left: 3px solid var(--color-primary);
  background: var(--bg-secondary);
  border-radius: 8px;
  color: var(--text-secondary);
  line-height: 1.5;
  width: 100%;
  max-width: 42rem;
  text-align: left;
  box-sizing: border-box;
}

/* Как в TrainingView: круговой таймер до следующей карточки */
.waiting-progress {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  user-select: none;
  -webkit-user-select: none;
}

.circular-progress {
  position: relative;
  width: 80px;
  height: 80px;
}

.progress-ring {
  transform: rotate(-90deg);
  transition: none;
}

.progress-ring-circle {
  transition: none;
  stroke-linecap: round;
}

.progress-ring-circle-bg {
  opacity: 0.3;
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 24px;
  font-weight: bold;
  color: var(--color-primary);
  user-select: none;
}
</style>

