<template>
  <div
    id="verb-forms-training"
    class="verb-forms-root"
    :class="{ 'verb-forms-root--embedded': embedded }"
  >
    <div v-if="autoStarting" class="card verb-forms-loading">
      <p class="verb-forms-loading-text">{{ t('common.loading') }}</p>
    </div>

    <section v-else-if="showIdleChrome" class="card verb-forms-pre">
      <component :is="headingTag" class="verb-forms-pre__title">
        {{ t('verbTraining.title') }}
      </component>
      <p class="verb-forms-pre__intro">{{ t('verbTraining.intro') }}</p>
      <button type="button" class="btn btn-primary" @click="start">{{ t('verbTraining.start') }}</button>
      <p v-if="error" class="verb-forms-pre__error">{{ error }}</p>
    </section>

    <div v-else-if="active" class="card verb-forms-training-card">
      <div v-if="finished" class="verb-forms-finished">
        <p>{{ t('verbTraining.sessionFinished') }}</p>
        <button type="button" class="btn btn-primary" @click="start">{{ t('verbTraining.newSession') }}</button>
      </div>

      <template v-else-if="card">
        <div v-if="card.total_cards" class="training-progress">
          <p>{{ t('verbTraining.cardProgress', { current: card.card_index, total: card.total_cards }) }}</p>
        </div>

        <div class="question verb-forms-question">
          <div>{{ card.prompt?.question }}</div>
          <p v-if="card.prompt?.example_translation" class="verb-forms-ru-line">{{ card.prompt.example_translation }}</p>
        </div>

        <div v-if="inputMode === 'choice' && card.options?.length" class="options">
          <button
            v-for="(option, index) in card.options"
            :key="`${index}-${option}`"
            type="button"
            :class="[
              'btn',
              'option-btn',
              {
                'option-correct': verbFeedback && option === verbFeedback.correct_answer,
                'option-incorrect': verbFeedback && !verbFeedback.is_correct && option === verbFeedback.chosen_option,
                'option-disabled': !!verbFeedback || answeringLocal,
              },
            ]"
            @click="!verbFeedback && !answeringLocal && submitAnswer(option)"
          >
            <span class="option-number">{{ index + 1 }}</span>
            <span class="option-text">{{ option }}</span>
          </button>
        </div>

        <div v-else-if="inputMode === 'typed' && card" class="type-block">
          <div class="type-answer-row">
            <span class="type-answer-label">{{ t('verbTraining.typeFormPlaceholder') }}</span>
            <div class="type-input-inline">
              <input
                v-model.trim="typedAnswer"
                type="text"
                class="type-input"
                :placeholder="t('training.typeWordPlaceholder') || ''"
                :disabled="!!verbFeedback || answeringLocal"
                @keydown.enter.prevent="submitTyped"
              />
              <button
                type="button"
                class="type-submit-inline"
                :disabled="!typedAnswer || !!verbFeedback || answeringLocal"
                :aria-label="t('training.check')"
                @click="submitTyped"
              >
                <Icon name="check" class="type-submit-icon" />
              </button>
            </div>
          </div>
        </div>

        <div v-if="!verbFeedback && !answeringLocal" class="type-actions-row">
          <button type="button" class="btn btn-secondary type-skip" @click="submitSkip">{{ t('training.skip') }}</button>
        </div>

        <div v-if="verbFeedback" class="feedback-section">
          <div v-if="verbFeedback.is_correct" class="feedback-badge feedback-success">
            <div class="success-particles">
              <div
                v-for="i in 12"
                :key="i"
                class="success-particle"
                :style="getSuccessParticleStyle(i)"
              ></div>
            </div>
            <span class="feedback-icon">✓</span>
            <span class="feedback-text">{{ encouragingPhrase }}</span>
          </div>
          <template v-else>
            <div class="feedback-badge feedback-error">
              <div v-if="waitingDelay" class="error-progress-wrapper">
                <div class="error-progress-pulse"></div>
                <svg class="error-progress-ring" width="40" height="40">
                  <circle
                    class="error-progress-circle-bg"
                    stroke="rgba(255, 255, 255, 0.2)"
                    stroke-width="2.5"
                    fill="transparent"
                    r="16"
                    cx="20"
                    cy="20"
                  />
                  <circle
                    class="error-progress-circle"
                    stroke="white"
                    stroke-width="2.5"
                    fill="transparent"
                    r="16"
                    cx="20"
                    cy="20"
                    :style="{ strokeDasharray: errorCircumference, strokeDashoffset: errorProgressOffset }"
                  />
                </svg>
                <svg class="error-icon-svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                  <line x1="18" y1="6" x2="6" y2="18"></line>
                  <line x1="6" y1="6" x2="18" y2="18"></line>
                </svg>
              </div>
              <svg v-else class="error-icon-svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
              <span class="feedback-text">{{ disappointingPhrase }}</span>
            </div>
          </template>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from './Icon.vue'
import { useAudio } from '../composables/useAudio'
import { useTrainingAnswerDelay } from '../composables/useTrainingAnswerDelay'

const props = withDefaults(
  defineProps<{
    embedded?: boolean
    autoStart?: boolean
  }>(),
  { embedded: true, autoStart: false }
)

const router = useRouter()
const { t, tm } = useI18n()
const { playSuccess, playFail } = useAudio()
const { waitingDelay, remainingMs, initialDelayMs, runWrongAnswerDelay, clearAll } = useTrainingAnswerDelay()

const headingTag = computed(() => (props.embedded ? 'h3' : 'h1'))

interface VerbFeedback {
  is_correct: boolean
  chosen_option: string
  correct_answer: string
  delay_seconds?: number
}

interface VerbCard {
  input_mode?: string
  typed_min_reps?: number
  options?: string[]
  prompt?: { question?: string; example_translation?: string }
  user_verb_card_id?: number
  card_index?: number
  total_cards?: number
}

const active = ref(false)
const finished = ref(false)
const error = ref('')
const autoStarting = ref(false)

/** Title + intro only before the first successful start (idle). Hidden during load and whole session. */
const showIdleChrome = computed(() => !active.value && !autoStarting.value)
const typedAnswer = ref('')
const card = ref<VerbCard | null>(null)
const verbFeedback = ref<VerbFeedback | null>(null)
const answeringLocal = ref(false)
const encouragingPhrase = ref('')
const disappointingPhrase = ref('')

const inputMode = computed(() => {
  const c = card.value
  if (!c) return 'choice'
  if (c.input_mode === 'typed') return 'typed'
  if (c.input_mode === 'choice') return 'choice'
  return c.options && c.options.length > 0 ? 'choice' : 'typed'
})

const errorCircumference = 2 * Math.PI * 16
const errorProgressOffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0 || verbFeedback.value?.is_correct) {
    return 0
  }
  const progress = remainingMs.value / initialDelayMs.value
  return errorCircumference * (1 - progress)
})

function phraseList(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.filter((x): x is string => typeof x === 'string' && x.length > 0)
}

function pickRandom(list: string[]): string {
  if (!list.length) return ''
  return list[Math.floor(Math.random() * list.length)]
}

function getSuccessParticleStyle(index: number) {
  const angle = index * 30 + 7
  const angleRad = (angle * Math.PI) / 180
  const distance = 72 + (index % 5) * 8
  const endX = Math.cos(angleRad) * distance
  const endY = Math.sin(angleRad) * distance
  const size = 5 + (index % 3)
  const delay = (index % 5) * 0.04
  return {
    '--particle-end-x': `${endX}px`,
    '--particle-end-y': `${endY}px`,
    '--particle-size': `${size}px`,
    '--particle-delay': `${delay}s`,
  }
}

const start = async () => {
  error.value = ''
  verbFeedback.value = null
  typedAnswer.value = ''
  finished.value = false
  clearAll()
  try {
    const data = await apiClient.request<VerbCard>('/api/verb-training/start', { method: 'POST' })
    active.value = true
    card.value = data
  } catch (e: any) {
    active.value = false
    if (e?.code === 'verb_training_disabled') {
      error.value = t('verbTraining.featureDisabledHint')
    } else {
      error.value = e?.message || t('verbTraining.startFailed')
    }
  }
}

onMounted(async () => {
  if (!props.autoStart) return
  autoStarting.value = true
  try {
    await start()
    if (active.value && !error.value) {
      await router.replace({ path: '/training/verbs' })
    }
  } finally {
    autoStarting.value = false
  }
})

async function postAnswer(body: { answer?: string; skip?: boolean }) {
  if (!card.value?.user_verb_card_id) return
  answeringLocal.value = true
  try {
    const data = await apiClient.request<VerbFeedback & { next?: boolean }>('/api/verb-training/answer', {
      method: 'POST',
      body: {
        user_verb_card_id: card.value.user_verb_card_id,
        ...body,
      },
    })
    verbFeedback.value = {
      is_correct: !!data.is_correct,
      chosen_option: data.chosen_option ?? '',
      correct_answer: data.correct_answer ?? '',
      delay_seconds: data.delay_seconds,
    }
    if (data.is_correct) {
      playSuccess()
      encouragingPhrase.value = pickRandom(phraseList('trainingFeedback.encouragingPhrases')) || t('verbTraining.correct')
    } else {
      playFail()
      disappointingPhrase.value = pickRandom(phraseList('trainingFeedback.disappointingPhrases')) || t('verbTraining.incorrect', { expected: data.correct_answer })
    }

    const advance = async () => {
      verbFeedback.value = null
      encouragingPhrase.value = ''
      disappointingPhrase.value = ''
      typedAnswer.value = ''
      clearAll()
      if (!data.next) {
        finished.value = true
        card.value = null
        return
      }
      card.value = await apiClient.request<VerbCard>('/api/verb-training/current')
    }

    if (data.is_correct) {
      await new Promise((r) => setTimeout(r, 1000))
      await advance()
    } else {
      runWrongAnswerDelay(data.delay_seconds ?? 0, () => {
        void advance()
      })
    }
  } catch (e) {
    console.error(e)
    verbFeedback.value = null
  } finally {
    answeringLocal.value = false
  }
}

const submitAnswer = (option: string) => {
  void postAnswer({ answer: option })
}

const submitTyped = () => {
  if (!typedAnswer.value) return
  void postAnswer({ answer: typedAnswer.value })
}

const submitSkip = () => {
  void postAnswer({ skip: true })
}
</script>

<style scoped>
.verb-forms-root {
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.verb-forms-root--embedded {
  padding: 0;
}

.verb-forms-loading {
  padding: 28px 16px;
  text-align: center;
}

.verb-forms-loading-text {
  margin: 0;
  font-size: 1rem;
  color: var(--text-secondary);
}

.verb-forms-pre {
  text-align: center;
  padding: 24px 20px;
}

.verb-forms-pre__title {
  margin: 0 0 10px;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
}

.verb-forms-root--embedded .verb-forms-pre__title {
  font-size: 1.15rem;
}

.verb-forms-pre__intro {
  margin: 0 0 18px;
  font-size: 0.95rem;
  line-height: 1.45;
  color: var(--text-secondary);
}

.verb-forms-pre__error {
  margin: 14px 0 0;
  color: #c62828;
  font-size: 0.9rem;
}

.verb-forms-finished {
  text-align: center;
  padding: 20px 12px;
}

.verb-forms-finished p {
  margin: 0 0 16px;
}

/* —— session card: mirror TrainingView (word training) —— */
.verb-forms-training-card .training-progress {
  margin-bottom: 20px;
  text-align: center;
}

.verb-forms-training-card .verb-forms-question {
  font-size: clamp(1.35rem, 3.2vw, 1.85rem);
  line-height: 1.35;
  margin: 24px 0 28px;
  text-align: center;
  font-weight: 600;
  word-wrap: break-word;
  overflow-wrap: break-word;
  hyphens: auto;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

.verb-forms-ru-line {
  margin: 14px 0 0;
  padding: 0;
  font-size: 0.95rem;
  font-weight: 400;
  line-height: 1.45;
  text-align: center;
  color: var(--text-secondary, #888);
}

.type-block {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.type-answer-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  width: 100%;
}
.type-input-inline {
  display: flex;
  align-items: center;
  width: 100%;
  max-width: 320px;
  border: 2px solid var(--border-color, #ddd);
  border-radius: 10px;
  background: var(--bg-primary);
  overflow: hidden;
  transition: border-color 0.2s ease;
}
.type-input-inline:focus-within {
  border-color: var(--primary, #4a90d9);
}
.type-answer-label {
  font-size: 0.95rem;
  color: var(--text-secondary, #666);
}
.type-input {
  flex: 1;
  min-width: 0;
  height: 44px;
  padding: 0 12px 0 10px;
  margin-bottom: 0;
  font-size: 1.1rem;
  border: none;
  background: transparent;
  color: var(--text-primary);
  text-align: center;
  box-sizing: border-box;
}
.type-input:focus {
  outline: none;
}
.type-input::placeholder {
  color: var(--text-tertiary, #999);
}
.type-submit-inline {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  align-self: stretch;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s ease, background 0.2s ease;
}
.type-submit-inline:hover:not(:disabled) {
  color: var(--primary, #4a90d9);
  background: var(--bg-secondary, rgba(0, 0, 0, 0.05));
}
.type-submit-inline:disabled {
  opacity: 0.4;
  cursor: default;
}
.type-submit-icon {
  width: 22px;
  height: 22px;
}
.type-actions-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-top: 12px;
}
.type-skip {
  margin: 0;
}

.options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 20px 0;
}

@media (min-width: 768px) {
  .options {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

.option-btn {
  min-height: 60px;
  font-size: 16px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
}

[data-theme='dark'] .option-btn {
  background-color: var(--bg-tertiary);
  border-color: var(--border-secondary);
}

.option-btn:hover:not(.option-disabled) {
  background-color: var(--bg-hover);
  border-color: var(--border-focus);
}

.option-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
}

.option-btn.option-correct .option-number,
.option-btn.option-incorrect .option-number {
  background: rgba(255, 255, 255, 0.3);
}

.option-text {
  flex: 1;
  text-align: left;
}

.option-btn.option-disabled {
  cursor: not-allowed;
  opacity: 0.7;
  background-color: var(--bg-secondary);
  border-color: var(--border-primary);
}

.option-btn.option-correct {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: 1px solid #10b981;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  animation: vft-correct-success 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
  overflow: hidden;
}

.option-btn.option-correct::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(
    45deg,
    transparent 30%,
    rgba(255, 255, 255, 0.3) 50%,
    transparent 70%
  );
  animation: vft-correct-shine 0.8s ease-out;
  pointer-events: none;
}

.option-btn.option-incorrect {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: 1px solid #ef4444;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  animation: vft-incorrect-fail 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
  position: relative;
}

.option-btn.option-incorrect::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  border-radius: 8px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  transform: translate(-50%, -50%) scale(0);
  animation: vft-incorrect-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes vft-incorrect-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 0.6;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.2);
    opacity: 0.3;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

@keyframes vft-correct-success {
  0% {
    transform: scale(1);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  30% {
    transform: scale(1.15) rotate(2deg);
    box-shadow: 0 8px 24px rgba(16, 185, 129, 0.5);
  }
  60% {
    transform: scale(1.08) rotate(-1deg);
    box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
  }
  100% {
    transform: scale(1);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
}

@keyframes vft-correct-shine {
  0% {
    transform: translateX(-100%) translateY(-100%) rotate(45deg);
  }
  100% {
    transform: translateX(100%) translateY(100%) rotate(45deg);
  }
}

@keyframes vft-incorrect-fail {
  0%,
  100% {
    transform: translateX(0) scale(1);
  }
  10% {
    transform: translateX(-12px) scale(0.95) rotate(-3deg);
  }
  20% {
    transform: translateX(12px) scale(0.95) rotate(3deg);
  }
  30% {
    transform: translateX(-10px) scale(0.97) rotate(-2deg);
  }
  40% {
    transform: translateX(10px) scale(0.97) rotate(2deg);
  }
  50% {
    transform: translateX(-8px) scale(0.98) rotate(-1deg);
  }
  60% {
    transform: translateX(8px) scale(0.98) rotate(1deg);
  }
  70% {
    transform: translateX(-4px) scale(0.99);
  }
  80% {
    transform: translateX(4px) scale(0.99);
  }
  90% {
    transform: translateX(-2px) scale(1);
  }
}

.feedback-section {
  margin-top: 30px;
  text-align: center;
}

.feedback-badge {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 16px 32px;
  border-radius: 12px;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  position: relative;
  overflow: hidden;
}

.feedback-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  animation: vft-feedback-success-appear 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feedback-success::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  animation: vft-feedback-success-glow 1.5s ease-out;
  pointer-events: none;
}

.feedback-error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  animation: vft-feedback-error-appear 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.feedback-error::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  border-radius: 12px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.2) 0%, transparent 70%);
  transform: translate(-50%, -50%) scale(0);
  animation: vft-feedback-error-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes vft-feedback-error-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 0.8;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.3);
    opacity: 0.4;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.6);
    opacity: 0;
  }
}

@keyframes vft-feedback-success-appear {
  0% {
    opacity: 0;
    transform: scale(0.3) translateY(-30px) rotate(-10deg);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  50% {
    transform: scale(1.1) translateY(0) rotate(5deg);
    box-shadow: 0 12px 32px rgba(16, 185, 129, 0.5);
  }
  70% {
    transform: scale(0.95) translateY(0) rotate(-2deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0) rotate(0deg);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
}

@keyframes vft-feedback-success-glow {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

@keyframes vft-feedback-error-appear {
  0% {
    opacity: 0;
    transform: scale(0.5) translateY(20px) rotate(10deg);
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  }
  30% {
    transform: scale(1.15) translateY(-5px) rotate(-5deg);
    box-shadow: 0 8px 24px rgba(239, 68, 68, 0.5);
  }
  60% {
    transform: scale(0.9) translateY(2px) rotate(2deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0) rotate(0deg);
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  }
}

.feedback-icon {
  font-size: 28px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  flex-shrink: 0;
  animation: vft-feedback-icon-spin 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes vft-feedback-icon-spin {
  0% {
    transform: scale(0) rotate(-180deg);
  }
  60% {
    transform: scale(1.3) rotate(10deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}

.success-particles {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 0;
  height: 0;
  pointer-events: none;
  z-index: 1;
}

.success-particle {
  position: absolute;
  top: 0;
  left: 0;
  width: var(--particle-size, 6px);
  height: var(--particle-size, 6px);
  background: radial-gradient(circle, rgba(255, 255, 255, 0.9) 0%, rgba(16, 185, 129, 0.8) 100%);
  border-radius: 50%;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
  animation: vft-success-particle-fly 1s ease-out var(--particle-delay, 0s) forwards;
}

@keyframes vft-success-particle-fly {
  0% {
    opacity: 1;
    transform: translate(0, 0) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(var(--particle-end-x, 0), var(--particle-end-y, 0)) scale(0);
  }
}

.feedback-text {
  letter-spacing: 0.5px;
}

.error-progress-wrapper {
  position: relative;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.error-progress-pulse {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.6);
  transform: translate(-50%, -50%);
  animation: vft-error-progress-pulse 1s linear 0.5s forwards;
  pointer-events: none;
}

@keyframes vft-error-progress-pulse {
  0% {
    transform: translate(-50%, -50%) scale(1);
    opacity: 0.8;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.3);
    opacity: 0.4;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.6);
    opacity: 0;
  }
}

.error-progress-ring {
  position: absolute;
  width: 40px;
  height: 40px;
  transform: rotate(-90deg);
}

.error-progress-circle-bg {
  opacity: 0.3;
}

.error-progress-circle {
  transition: stroke-dashoffset 0.1s linear;
  stroke-linecap: round;
}

.error-icon-svg {
  position: relative;
  z-index: 1;
  color: white;
  flex-shrink: 0;
}
</style>
