<template>
  <div class="training">
    <h1 v-if="!sessionActive && !loading && !sessionComplete" class="training-title">{{ t('training.title') }}</h1>
    
    <TrainingSessionCompletion
      v-if="sessionComplete && !sessionActive"
      :total-cards="trainingStats.totalCards"
      :correct-cards="trainingStats.correctCards"
      :stats-loaded="statsLoaded"
      :available-for-training="stats.availableForTraining"
      :estimated-time-for-remaining="estimatedTimeForRemaining"
      :show-continue-button="true"
      :sounds-enabled="false"
      @continue="startTraining"
    />

    <div
      v-if="sessionComplete && !sessionActive && !loading && showSpanishVerbFormsTraining"
      class="card training-verb-forms-cta training-verb-forms-cta--compact"
    >
      <p
        v-if="verbFormsTotalCardsPool !== null"
        class="training-verb-forms-cta__count training-verb-forms-cta__count--compact"
      >
        {{ t('verbTraining.totalCardsAvailable', { count: verbFormsTotalCardsPool }) }}
      </p>
      <button
        type="button"
        class="btn btn-primary training-verb-forms-cta__btn"
        @click="openVerbFormsTraining"
      >
        {{ t('verbTraining.openDedicated') }}
      </button>
    </div>

    <div v-if="!sessionActive && !loading && !sessionComplete" class="training-idle-stack">
      <div class="card start-screen">
        <div class="start-screen-content">
        <div class="start-screen-stats" v-if="statsLoaded">
          <div class="start-stat-item">
            <span class="start-stat-label">{{ t('training.availableForTraining') }}</span>
            <span class="start-stat-value">
              {{ stats.availableForTraining }} {{ (t as any)('common.cards', stats.availableForTraining) }}
              <span v-if="estimatedTime">({{ estimatedTime }})</span>
            </span>
          </div>
        </div>
        
        <!-- Upcoming cards chart -->
        <div v-if="statsLoaded && upcomingCardsLoaded" class="upcoming-cards-chart">
          <div class="chart-header">
            <h3 class="chart-title">{{ t('training.upcomingCards') }}</h3>
            <div class="chart-subtitle">{{ t('training.upcomingCardsDescription') }}</div>
          </div>
          <div class="chart-container">
            <canvas ref="upcomingChartCanvas"></canvas>
          </div>
        </div>
        
        <button v-if="statsLoaded && stats.availableForTraining > 0" @click="startTraining" class="btn btn-primary btn-start">
          {{ t('training.startTraining') || 'Start Training' }}
        </button>
        <p v-if="statsLoaded && stats.availableForTraining === 0" class="no-cards-message">
          {{ t('training.noCardsAvailable') }}
        </p>
        </div>
      </div>
      <div v-if="showSpanishVerbFormsTraining" class="card training-verb-forms-cta">
        <h3 class="training-verb-forms-cta__title">{{ t('verbTraining.title') }}</h3>
        <p class="training-verb-forms-cta__text">{{ t('verbTraining.shortBlurb') }}</p>
        <p v-if="verbFormsTotalCardsPool !== null" class="training-verb-forms-cta__count">
          {{ t('verbTraining.totalCardsAvailable', { count: verbFormsTotalCardsPool }) }}
        </p>
        <button
          type="button"
          class="btn btn-primary training-verb-forms-cta__btn"
          @click="openVerbFormsTraining"
        >
          {{ t('verbTraining.openDedicated') }}
        </button>
      </div>
    </div>

    <LgLoader v-if="loading" />

    <!-- Network error notification -->
    <div v-if="networkError" class="network-error-notification">
      <div class="network-error-content">
        <Icon name="warning" class="network-error-icon" />
        <div class="network-error-text">
          <div class="network-error-title">{{ t('training.networkError') }}</div>
          <div class="network-error-message">
            {{ networkErrorRetrying ? t('common.retrying', { attempt: networkErrorAttempt, max: networkErrorMaxAttempts }) : t('common.networkError') }}
          </div>
        </div>
        <button type="button" class="network-error-close" @click="dismissNetworkError">×</button>
      </div>
    </div>

    <div 
      v-if="sessionActive && currentCard" 
      class="card"
      :class="{ 'card-timer-active': waitingDelay }"
      @mousedown="waitingDelay ? handleTimerMouseDown($event) : null"
      @mouseup="waitingDelay ? handleTimerMouseUp($event) : null"
      @mouseleave="waitingDelay ? handleTimerMouseLeave() : null"
      @touchstart="waitingDelay ? handleTimerMouseDown($event) : null"
      @touchend="waitingDelay ? handleTimerMouseUp($event) : null"
      @touchcancel="waitingDelay ? handleTimerMouseLeave() : null"
    >
      <div class="training-progress" v-if="cardIndex > 0 && totalCards > 0">
        <p>{{ t('training.cardOf', { current: cardIndex, total: totalCards }) }}</p>
      </div>

      <div class="question">
        <div v-html="processedQuestion"></div>
        <div v-if="showQuestionMetaRow" class="question-meta-row">
          <div v-if="showMorphInTraining && isTargetLangSide && morphCompactText" class="question-morph-inline">
            <template v-if="morphDisplay.kind === 'noun'">
              <div v-if="nounOppositeWord" class="morph-opposite-line">
                <span class="morph-opposite" :class="morphOppositeGenderClass">({{ nounOppositeWord }})</span>
              </div>
            </template>
            <template v-else>
              {{ morphCompactText }}
            </template>
          </div>

          <div
            v-if="isTargetLangSide && (currentCard?.transcription || pronunciationWord)"
            class="training-pronunciation-row"
          >
            <span v-if="currentCard?.transcription" class="training-transcription">{{ currentCard.transcription }}</span>
            <button
              v-if="pronunciationWord"
              type="button"
              class="btn-pronunciation"
              :disabled="playingPronunciation || !pronunciationWord"
              :aria-label="t('training.listen') || 'Pronounce'"
              @click="playCurrentPronunciation"
            >
              <Icon name="play" />
            </button>
          </div>
        </div>
      </div>
      <div
        v-if="!showQuestionMetaRow && isTargetLangSide && (currentCard?.transcription || pronunciationWord)"
        class="training-pronunciation-row training-pronunciation-row-standalone"
      >
        <span v-if="currentCard?.transcription" class="training-transcription">{{ currentCard.transcription }}</span>
        <button
          v-if="pronunciationWord"
          type="button"
          class="btn-pronunciation"
          :disabled="playingPronunciation || !pronunciationWord"
          :aria-label="t('training.listen') || 'Pronounce'"
          @click="playCurrentPronunciation"
        >
          <Icon name="play" />
        </button>
      </div>
      <!-- Type: type the word; lightbulb shows first letter + underscores -->
      <div v-if="currentCard?.type === 'type'" class="type-block">
        <div class="type-answer-row">
          <span class="type-answer-label">{{ t('training.typeWord') || 'Enter the word:' }}</span>
          <div class="type-input-inline">
            <template v-if="feedback && !feedback.is_correct && feedback.correct_answer">
              <span class="type-input type-reveal-text">{{ typeRevealDisplayText }}</span>
            </template>
            <template v-else>
              <span v-if="currentCard?.prefix" class="type-input-prefix">{{ currentCard.prefix }}</span>
              <input
                ref="typeInputRef"
                v-model.trim="typeAnswerText"
                type="text"
                class="type-input"
                :placeholder="t('training.typeWordPlaceholder') || 'word'"
                :disabled="!!feedback || answering"
                @keydown.enter.prevent="submitTypeAnswer"
              />
              <button
                type="button"
                class="type-submit-inline"
                :disabled="!typeAnswerText || !!feedback || answering"
                :aria-label="t('training.check') || 'Check'"
                @click="submitTypeAnswer"
              >
                <Icon name="check" class="type-submit-icon" />
              </button>
              <VoiceMicButton
                :lang="learning?.target_lang ?? 'en'"
                :disabled="!!feedback || answering"
                :label="t('sentence.voiceInput')"
                @transcript="onTrainingVoiceTranscript"
              />
            </template>
          </div>
        </div>
        <div class="type-actions-row">
          <button
            v-if="!feedback && !answering"
            type="button"
            class="btn btn-secondary type-skip"
            @click="skipTypeAnswer"
          >{{ t('training.skip') || 'Пропустить' }}</button>
        </div>
        <div
          v-if="showTypeHintButton && (currentCard?.hint_first_letter !== undefined && currentCard?.hint_length != null) && !(feedback && !feedback.is_correct)"
          class="type-hint-button-wrapper"
          :class="{ 'type-hint-button-visible': typeHintButtonVisible }"
        >
          <button
            v-if="!typeHintShown && !feedback && !answering"
            type="button"
            class="btn-type-hint-icon"
            :aria-label="t('training.typeHint') || 'Подсказка'"
            @click="typeHintShown = true"
          >
            <Icon name="lightbulb" class="type-hint-icon" />
          </button>
          <div v-else-if="typeHintShown" class="type-hint-text">
            {{ typeHintDisplay }}
          </div>
        </div>
      </div>

      <!-- Spell: compose word from letters -->
      <div
        v-if="currentCard?.type === 'spell' && currentCard?.letters?.length"
        class="spell-block"
        :class="{ 'spell-long': (currentCard?.letters?.length ?? 0) > 6 }"
      >
        <div class="spell-answer-row">
          <span class="spell-answer-label">{{ t('training.composeWord') || 'Your word:' }}</span>
          <div
            ref="spellAnswerLettersContainerRef"
            class="spell-answer-letters"
            :class="{
              'spell-reveal-letters': feedback && !feedback.is_correct && spellRevealLetters.length,
              'spell-autopick-active': spellSkipResultActive
            }"
          >
            <div
              ref="spellAnswerLettersWrapRef"
              class="spell-answer-letters-inner"
              :style="spellAnswerLettersWrapStyle"
            >
              <span v-if="currentCard?.prefix" class="spell-answer-prefix">{{ currentCard.prefix }}</span>
              <template v-if="feedback && !feedback.is_correct && spellRevealLetters.length">
                <TransitionGroup
                  name="spell-reorder"
                  tag="span"
                  class="spell-reorder-group"
                >
                  <span
                    v-for="item in spellRevealLetters"
                    :key="item.key"
                    class="spell-reveal-char"
                  >{{ item.letter }}</span>
                </TransitionGroup>
              </template>
              <template v-else>
                <button
                  v-for="(ch, i) in spellAnswerLetters"
                  :key="`a-${i}`"
                  type="button"
                  class="btn spell-letter-btn spell-answer-char-btn"
                  :disabled="(!!feedback || answering) && !spellSkipResultActive"
                  @click="spellRemoveLetterAt(i)"
                >{{ ch }}</button>
                <span v-if="spellAnswerLetters.length === 0" class="spell-answer-placeholder">...</span>
              </template>
            </div>
          </div>
        </div>
        <div
          v-show="!feedback || spellSkipAutoPickInProgress"
          class="spell-letters"
          :class="{ 'spell-letters-autopick': spellSkipAutoPickInProgress }"
        >
          <button
            v-for="(ch, i) in (currentCard?.letters ?? [])"
            :key="i"
            type="button"
            class="btn spell-letter-btn"
            :class="{ 'spell-letter-used': spellUsedIndices.includes(i) }"
            :disabled="spellUsedIndices.includes(i) || ((!!feedback || answering) && !spellSkipResultActive)"
            @click="spellAddLetterByIndex(i)"
          >{{ ch }}</button>
        </div>
        <div class="spell-actions-row">
          <button
            v-if="!feedback && !answering"
            type="button"
            class="btn btn-secondary spell-skip"
            @click="skipSpellAnswer"
          >{{ t('training.skip') || 'Пропустить' }}</button>
          <VoiceMicButton
            v-if="!feedback && !answering"
            :lang="learning?.target_lang ?? 'en'"
            :label="t('sentence.voiceInput')"
            @transcript="onTrainingVoiceTranscript"
          />
        </div>
        <div
          v-if="showSpellHintButton && spellHintEligible && !(feedback && !feedback.is_correct)"
          class="type-hint-button-wrapper spell-hint-button-wrapper"
          :class="{ 'type-hint-button-visible': spellHintButtonVisible }"
        >
          <button
            v-if="!spellHintShown && !feedback && !answering"
            type="button"
            class="btn-type-hint-icon"
            :aria-label="t('training.typeHint') || 'Подсказка'"
            @click="spellHintShown = true"
          >
            <Icon name="lightbulb" class="type-hint-icon" />
          </button>
          <div v-else-if="spellHintShown" class="type-hint-text">
            {{ spellHintDisplay }}
          </div>
        </div>
      </div>

      <div 
        v-if="optionsShown && currentCard?.type !== 'spell' && currentCard?.type !== 'type'" 
        class="options"
        @mousedown="waitingDelay ? handleTimerMouseDown($event) : null"
        @mouseup="waitingDelay ? handleTimerMouseUp($event) : null"
        @touchstart="waitingDelay ? handleTimerMouseDown($event) : null"
        @touchend="waitingDelay ? handleTimerMouseUp($event) : null"
      >
        <button
          v-for="(option, index) in options"
          :key="`${index}:${option}`"
          @click="!feedback && !answering && submitAnswer(index)"
          :class="[
            'btn',
            'option-btn',
            {
              'option-correct': isCorrectOption(option, index),
              'option-incorrect': isIncorrectChosenOption(index),
              'option-disabled': !!feedback || answering
            }
          ]"
        >
          <span class="option-number">{{ index + 1 }}</span>
          <span class="option-text">{{ option }}</span>
        </button>
      </div>

      <!-- Example button/display - show for English words, appears after options -->
      <div 
        v-if="optionsShown && showExampleButton && isTargetLangSide" 
        class="example-button-wrapper"
        :class="{ 'example-button-visible': showExampleButtonVisible }"
      >
        <button 
          v-if="!exampleUsageShown && !feedback"
          @click="showExampleUsage" 
          class="btn-example-icon"
          aria-label="Показать пример"
        >
          <Icon name="lightbulb" class="example-icon" />
        </button>
        <div v-else-if="exampleUsageShown" class="example example-usage">
          <ClickableText :text="currentCard?.example_target || currentCard?.example_en || feedback?.example_target || feedback?.example || ''" :exclude="hintExcludedWord" subtle-underline />
        </div>
      </div>

      <div v-if="feedback" class="feedback-section">
        <div v-if="feedback.is_correct" class="feedback-badge feedback-success">
          <!-- Success particles -->
          <div class="success-particles">
            <div v-for="i in 12" :key="i" class="success-particle" :style="getSuccessParticleStyle(i)"></div>
          </div>
          <span class="feedback-icon"><Icon name="check" /></span>
          <span class="feedback-text">{{ currentEncouragingPhrase }}</span>
        </div>
        
        <!-- For incorrect answers: spell = letters reorder in block above; type/cards = hint/example -->
        <template v-if="!feedback.is_correct">
          <div v-if="feedback.hint" class="hint"><ClickableText :text="feedback.hint" :exclude="hintExcludedWord" /></div>
          <div v-if="feedback.example" class="example"><ClickableText :text="feedback.example" :exclude="hintExcludedWord" subtle-underline /></div>
          <div class="feedback-badge feedback-error">
            <div 
              v-if="waitingDelay" 
              class="error-progress-wrapper"
              @mousedown="handleTimerMouseDown"
              @mouseup="handleTimerMouseUp"
              @mouseleave="handleTimerMouseLeave"
              @touchstart="handleTimerMouseDown"
              @touchend="handleTimerMouseUp"
              @touchcancel="handleTimerMouseLeave"
            >
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
                  :style="{ 
                    strokeDasharray: errorCircumference, 
                    strokeDashoffset: errorProgressOffset 
                  }"
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
            <span class="feedback-text">{{ currentDisappointingPhrase }}</span>
          </div>
        </template>
        
        <!-- For correct answers: show example after notification -->
        <div v-if="feedback.is_correct && feedback.example" class="example"><ClickableText :text="feedback.example" :exclude="hintExcludedWord" subtle-underline /></div>
        
        <!-- Circular progress for correct answers delay (if any) -->
        <div 
          v-if="waitingDelay && feedback.is_correct" 
          class="waiting-progress"
          @mousedown="handleTimerMouseDown"
          @mouseup="handleTimerMouseUp"
          @mouseleave="handleTimerMouseLeave"
          @touchstart="handleTimerMouseDown"
          @touchend="handleTimerMouseUp"
          @touchcancel="handleTimerMouseLeave"
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

    </div>
    <div v-if="sessionActive && currentCard" class="report-row report-row-outside">
      <button
        v-if="!reportAlreadySent"
        type="button"
        class="report-text-link"
        :disabled="reportSubmitting"
        @click="openWordReportDialog"
      >
        {{ t('training.reportIssue') || 'Пожаловаться' }}
      </button>
      <span v-if="reportMessage" class="report-message">{{ reportMessage }}</span>
    </div>
    <ContentReportDialog
      :open="reportDialogOpen"
      :submitting="reportSubmitting"
      :categories="wordReportCategories"
      :category="reportCategory"
      :details="reportDetails"
      @update:category="reportCategory = $event"
      @update:details="reportDetails = $event"
      @close="closeWordReportDialog"
      @submit="submitWordReport"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick, TransitionGroup } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { contentReportClient } from '../api/contentReportClient'
import { wordTrainingClient } from '../api/wordTrainingClient'
import { showAlert } from '../composables/useDialog'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'
import { useLocale } from '../composables/useLocale'
import { Chart, registerables } from 'chart.js'
import Icon from '../components/Icon.vue'
import ClickableText from '../components/ClickableText.vue'
import VoiceMicButton from '../components/VoiceMicButton.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import TrainingSessionCompletion from '../components/TrainingSessionCompletion.vue'
import ContentReportDialog from '../components/ContentReportDialog.vue'
import {
  WORD_TRAINING_REPORT_CATEGORIES,
  buildReportComment
} from '../constants/contentReportCategories'
import { useLearningConfig } from '../composables/useLearningConfig'
import { useCourse } from '../composables/useCourse'
import { useSpanishVerbFormsPractice } from '../composables/useSpanishVerbFormsPractice'

const { t, tm, locale } = useI18n()
const { ensureLearningLoaded, learning } = useLearningConfig()
const { currentCourseCode } = useCourse()

const isOnline = ref(typeof navigator === 'undefined' ? true : navigator.onLine)
const {
  verbFormsTotalCardsPool,
  showSpanishVerbFormsTraining,
  refreshVerbFormsPoolCount,
  openVerbFormsTraining,
} = useSpanishVerbFormsPractice(isOnline)

watch(currentCourseCode, async () => {
  await ensureLearningLoaded()
  loadStats()
  loadUpcomingCards()
  await refreshVerbFormsPoolCount()
})

function phraseList(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.filter((x): x is string => typeof x === 'string' && x.length > 0)
}

const encouragingPhrasesList = computed(() => phraseList('trainingFeedback.encouragingPhrases'))
const disappointingPhrasesList = computed(() => phraseList('trainingFeedback.disappointingPhrases'))

const motivationalBuckets = computed(() => ({
  excellent: phraseList('trainingFeedback.motivational.excellent'),
  great: phraseList('trainingFeedback.motivational.great'),
  good: phraseList('trainingFeedback.motivational.good'),
  okay: phraseList('trainingFeedback.motivational.okay'),
  needsWork: phraseList('trainingFeedback.motivational.needsWork'),
  poor: phraseList('trainingFeedback.motivational.poor'),
}))
const { currentLocale } = useLocale()

Chart.register(...registerables)

interface Card {
  question: string
  card_index: number
  total_cards: number
  session_id: number
  user_card_id: number
  delay_ms: number
  direction: string
  word_en?: string
  word_target?: string
  transcription?: string
  display_word?: string
  display_target?: string
  example_en?: string
  example_target?: string
  /** Spell challenge: compose word from letters; type: type-the-word (no letters) */
  type?: 'card' | 'spell' | 'type'
  word_ru?: string
  /** Non-editable prefix for spell (e.g. "to " for verbs); user composes the rest */
  prefix?: string
  letters?: string[]
  correct_answer?: string
  /** Type challenge: first letter for hint (on demand) */
  hint_first_letter?: string
  /** Type challenge: word length for hint */
  hint_length?: number
  morph?: MorphInfo
  word_card_id?: number
  training_card_id?: number
  word_category?: string
  offline?: boolean
}

interface MorphVerbForms {
  v1?: string
  v2?: string
  v3?: string
}

interface MorphInfo {
  pos?: string
  noun_gender?: string
  article?: string
  opposite_gender_word?: string
  verb_forms?: MorphVerbForms
}

interface OptionsResponse {
  options: string[]
  user_card_id: number
}

interface Feedback {
  is_correct: boolean
  chosen_option: string
  correct_answer: string
  hint?: string
  example?: string
  example_target?: string
  delay_seconds?: number
}

const sessionActive = ref(false)

// The trained word in the target language: hints/examples must not open its own
// dictionary card (that would spoil the answer), so ClickableText excludes it.
const hintExcludedWord = computed(() => {
  const card = currentCard.value
  return (
    card?.word_target ||
    card?.display_target ||
    card?.word_en ||
    card?.display_word ||
    feedback.value?.correct_answer ||
    ''
  )
})
const loading = ref(false)
const currentCard = ref<Card | null>(null)
const prefetchedCardResponse = ref<any | null>(null)
let prefetchInFlight: Promise<void> | null = null
let syncCurrentInFlight: Promise<void> | null = null
const optionsShown = ref(false)
const options = ref<string[]>([])
const feedback = ref<Feedback | null>(null)
const chosenOptionIndex = ref<number | null>(null)
const currentEncouragingPhrase = ref('')
const currentDisappointingPhrase = ref('')
const answering = ref(false)
const waitingDelay = ref(false)
const delaySeconds = ref(0)
const initialDelaySeconds = ref(0)
const remainingMs = ref(0)
const initialDelayMs = ref(0)
const sessionComplete = ref(false)
const cardsCompleted = ref(0)
const trainingStats = ref({
  totalCards: 0,
  correctCards: 0
})
const stats = ref({
  dueCount: 0,
  availableForTraining: 0,
})
const statsLoaded = ref(false)
const networkError = ref(false)
const networkErrorRetrying = ref(false)
const networkErrorAttempt = ref(0)
const networkErrorMaxAttempts = ref(3)
const animatedPercentage = ref(0)
const percentageAnimationComplete = ref(false)
const upcomingChartCanvas = ref<HTMLCanvasElement | null>(null)
const upcomingCardsLoaded = ref(false)
const upcomingCardsData = ref<Record<string, { date: string; label: string; count: number }>>({})
let upcomingChartInstance: Chart | null = null
let networkErrorHideTimer: ReturnType<typeof setTimeout> | null = null
const showExampleButton = ref(false)
const showExampleButtonVisible = ref(false)
const exampleUsageShown = ref(false)
let exampleButtonTimer: ReturnType<typeof setTimeout> | null = null
const reportSubmitting = ref(false)
const reportMessage = ref('')
const reportSentForCardKey = ref('')
const reportDialogOpen = ref(false)
const reportCategory = ref('')
const reportDetails = ref('')
const wordReportCategories = WORD_TRAINING_REPORT_CATEGORIES

const cardReportKey = (card: Card | null): string => {
  if (!card) return ''
  if (card.user_card_id) return `user:${card.user_card_id}`
  if (card.training_card_id) return `training:${card.training_card_id}`
  if (card.word_card_id) return `word:${card.word_card_id}`
  return `${card.word_en || card.display_word || ''}:${card.direction || ''}`
}

const reportAlreadySent = computed(() => {
  const key = cardReportKey(currentCard.value)
  return !!key && reportSentForCardKey.value === key
})

// Spell (compose word) state
const spellAnswerLetters = ref<string[]>([])
/** Indices into currentCard.letters that are already used (same order as spellAnswerLetters) */
const spellUsedIndices = ref<number[]>([])
const spellAnswerLettersContainerRef = ref<HTMLElement | null>(null)
const spellAnswerLettersWrapRef = ref<HTMLElement | null>(null)
/** Scale factor so that collected letters fit container width (1 = no scaling) */
const spellScale = ref(1)
/** For wrong spell: letters in correct order with stable keys for reorder animation */
const spellRevealLetters = ref<Array<{ letter: string; key: number }>>([])

/** Recompute scale so spell answer letters fit container width */
function updateSpellScale() {
  const container = spellAnswerLettersContainerRef.value
  const inner = spellAnswerLettersWrapRef.value
  if (!container || !inner) {
    spellScale.value = 1
    return
  }
  const containerWidth = container.clientWidth
  const contentWidth = inner.scrollWidth
  if (contentWidth <= 0) {
    spellScale.value = 1
    return
  }
  const scale = containerWidth / contentWidth
  spellScale.value = scale < 1 ? Math.max(0.35, scale) : 1
}

const spellAnswerLettersWrapStyle = computed(() => ({
  transform: `scale(${spellScale.value})`
}))
const spellSkipAutoPickInProgress = ref(false)
const spellSkipResultActive = ref(false)
const spellHintShown = ref(false)
const showSpellHintButton = ref(false)
const spellHintButtonVisible = ref(false)
let spellHintButtonTimer: ReturnType<typeof setTimeout> | null = null
// Type (type the word, no letters) state
const typeAnswerText = ref('')
const typeHintShown = ref(false)
const showTypeHintButton = ref(false)
const typeHintButtonVisible = ref(false)
let typeHintButtonTimer: ReturnType<typeof setTimeout> | null = null
/** For wrong type answer: animated text (erase wrong → type correct) */
const typeRevealDisplayText = ref('')
let typeRevealTimeouts: ReturnType<typeof setTimeout>[] = []
const typeInputRef = ref<HTMLInputElement | null>(null)
const playingPronunciation = ref(false)
const currentPronunciationURL = ref<string | null>(null)
let pronunciationLoadRequestId = 0
let currentCardGeneration = 0

const sameTrainingCard = (a: Card | null | undefined, b: Card | null | undefined): boolean => {
  if (!a || !b) return false
  return a.user_card_id === b.user_card_id && a.card_index === b.card_index && a.session_id === b.session_id
}

// Settings
const { settings, setAutoplayPronunciation } = useSettings()
const { playSuccess, playFail, playVictory, playDefeat, getWordPronunciationURL, playWordPronunciation } = useAudio()

// Target-language side of the card (e.g. EN in RU→EN when direction is en_ru)
const isTargetLangSide = computed(() => {
  return currentCard.value?.direction === 'en_ru'
})

const showMorphInTraining = computed(() => !settings.value.hideMorphInTraining)

const morphDisplay = computed(() => {
  const morph = currentCard.value?.morph
  const nounWord = currentCard.value?.word_target || currentCard.value?.display_target || currentCard.value?.word_en || ''
  if (!morph) return { kind: 'none' as const, article: '', gender: '', opposite: '', word: '' }
  if (morph.pos === 'noun' && morph.noun_gender) {
    return {
      kind: 'noun' as const,
      article: morph.article || '',
      gender: morph.noun_gender,
      opposite: morph.opposite_gender_word || '',
      word: nounWord
    }
  }
  return { kind: 'other' as const, article: '', gender: '', opposite: '', word: '' }
})

const morphGenderClass = computed(() => {
  const g = (morphDisplay.value.gender || '').trim().toLowerCase()
  if (g === 'm' || g === 'masculine' || g === 'masculino') return 'morph-gender-m'
  if (g === 'f' || g === 'feminine' || g === 'femenino') return 'morph-gender-f'
  return ''
})

const morphOppositeGenderClass = computed(() => {
  const g = (morphDisplay.value.gender || '').trim().toLowerCase()
  if (g === 'm' || g === 'masculine' || g === 'masculino') return 'morph-gender-f'
  if (g === 'f' || g === 'feminine' || g === 'femenino') return 'morph-gender-m'
  return ''
})

const nounOppositeWord = computed(() => {
  if (morphDisplay.value.kind !== 'noun') return ''
  const opposite = (morphDisplay.value.opposite || '').trim()
  if (!opposite) return ''
  const currentWord = (morphDisplay.value.word || '').trim().toLowerCase()
  if (opposite.toLowerCase() === currentWord) return ''
  return opposite
})

const morphCompactText = computed(() => {
  const morph = currentCard.value?.morph
  if (!morph) return ''
  if (morph.pos === 'noun' && morph.noun_gender) {
    return nounOppositeWord.value ? `(${nounOppositeWord.value})` : ''
  }
  if ((morph.pos === 'verb' || morph.pos === 'aux') && morph.verb_forms) {
    const forms = [morph.verb_forms.v1, morph.verb_forms.v2, morph.verb_forms.v3].filter(Boolean)
    if (forms.length > 0) return forms.join(', ')
  }
  return ''
})

const showQuestionMetaRow = computed(() => {
  const hasPron = isTargetLangSide.value && (!!currentCard.value?.transcription || !!pronunciationWord.value)
  const hasMorph = showMorphInTraining.value && isTargetLangSide.value && morphCompactText.value.length > 0
  return hasPron || hasMorph
})

const pronunciationWord = computed(() => {
  const card = currentCard.value
  if (!card) return ''
  return (card.word_target || card.word_en || '').trim()
})

const normalizeAnswerOption = (value: string | undefined | null): string =>
  (value || '').trim().replace(/\s+/g, ' ').toLowerCase()

const isCorrectOption = (option: string, index: number): boolean => {
  const resp = feedback.value
  if (!resp) return false
  if (resp.is_correct && index === chosenOptionIndex.value) return true
  return normalizeAnswerOption(option) === normalizeAnswerOption(resp.correct_answer)
}

const isIncorrectChosenOption = (index: number): boolean => {
  const resp = feedback.value
  return !!resp && !resp.is_correct && index === chosenOptionIndex.value
}

const shouldAutoplayPronunciationOnCardShown = computed(() => {
  const card = currentCard.value
  if (!card) return false
  if (!settings.value.autoplayPronunciation) return false
  // Trigger A: foreign-side card is shown (en_ru).
  return card.direction === 'en_ru'
})

watch(currentCard, async (card) => {
  const reqId = ++pronunciationLoadRequestId
  if (!card) {
    currentPronunciationURL.value = null
    return
  }
  const word = (card.word_target || card.word_en || '').trim()
  if (!word) {
    currentPronunciationURL.value = null
    return
  }
  const url = await getWordPronunciationURL(word)
  if (reqId !== pronunciationLoadRequestId || card !== currentCard.value) return
  currentPronunciationURL.value = url
  if (!url || !shouldAutoplayPronunciationOnCardShown.value || playingPronunciation.value) return
  playingPronunciation.value = true
  try {
    await playWordPronunciation(word)
  } finally {
    playingPronunciation.value = false
  }
})

const normalizePronunciationAnswerWord = (answer: string, prefix?: string): string => {
  const raw = (answer || '').trim()
  if (!raw) return ''
  const p = (prefix || '').trim()
  if (!p) return raw
  const lowerRaw = raw.toLowerCase()
  const lowerPrefix = p.toLowerCase()
  if (lowerRaw.startsWith(lowerPrefix)) {
    const cut = raw.slice(p.length).trim()
    if (cut) return cut
  }
  return raw
}

const autoplayPronunciationAfterAnswer = async (resp?: Feedback) => {
  const card = currentCard.value
  if (!card) return
  if (!settings.value.autoplayPronunciation) return
  // Trigger B: native prompt with target-language answer => play after answer feedback.
  // For classic cards this is ru_en; for special challenges backend uses direction=spell/type.
  const expectsTargetAnswer = card.direction === 'ru_en' || card.type === 'spell' || card.type === 'type'
  if (!expectsTargetAnswer) return
  const fromAnswer = normalizePronunciationAnswerWord(resp?.correct_answer || '', card.prefix)
  const word = fromAnswer || pronunciationWord.value
  if (!word || playingPronunciation.value) return
  playingPronunciation.value = true
  try {
    await playWordPronunciation(word)
  } finally {
    playingPronunciation.value = false
  }
}

const playCurrentPronunciation = async () => {
  const word = pronunciationWord.value
  if (!word || playingPronunciation.value) return
  playingPronunciation.value = true
  try {
    await playWordPronunciation(word)
  } finally {
    playingPronunciation.value = false
  }
}

// Same mask as type: prefix + first letter of stem + underscores (stem = answer after prefix)
function typeChallengeStemFromAnswer(fullAnswer: string | undefined, prefix: string | undefined): string {
  const full = (fullAnswer ?? '').trim()
  const pre = prefix ?? ''
  if (pre && full.startsWith(pre)) return full.slice(pre.length)
  return full
}

function maskedFirstLetterHint(prefix: string, firstLetter: string, stemRuneCount: number): string {
  if (!firstLetter || stemRuneCount <= 0) return prefix
  const rest = stemRuneCount > 1 ? ' ' + '_'.repeat(stemRuneCount - 1) : ''
  return prefix + firstLetter + rest
}

// Type challenge hint: optional prefix + first letter + masked rest (e.g. "to s ___")
const typeHintDisplay = computed(() => {
  const card = currentCard.value
  const prefix = card?.prefix ?? ''
  const first = card?.hint_first_letter ?? ''
  const len = card?.hint_length ?? 0
  return maskedFirstLetterHint(prefix, first, len)
})

const spellHintEligible = computed(() => {
  const card = currentCard.value
  if (card?.type !== 'spell' || !card.correct_answer?.trim()) return false
  const stem = typeChallengeStemFromAnswer(card.correct_answer, card.prefix)
  return [...stem].length > 0
})

const spellHintDisplay = computed(() => {
  const card = currentCard.value
  const prefix = card?.prefix ?? ''
  const stem = typeChallengeStemFromAnswer(card?.correct_answer, card?.prefix)
  const runes = [...stem]
  if (runes.length === 0) return prefix
  return maskedFirstLetterHint(prefix, runes[0], runes.length)
})

// For spell keyboard: first unused index with this letter
// Voice answer for type/spell challenges: the transcript fills the text input, or picks the
// spell letters in order (auto-submitting when all letters are placed, same as tapping them).
const onTrainingVoiceTranscript = (raw: string) => {
  const card = currentCard.value
  if (!card || feedback.value || answering.value) return
  let word = raw.trim().toLowerCase()
  const prefix = (card.prefix ?? '').trim().toLowerCase()
  if (prefix && word.startsWith(`${prefix} `)) word = word.slice(prefix.length).trim()
  if (!word) return
  if (card.type === 'type') {
    typeAnswerText.value = word
    return
  }
  if (card.type === 'spell' && card.letters?.length) {
    while (spellAnswerLetters.value.length > 0) {
      spellRemoveLetterAt(spellAnswerLetters.value.length - 1)
    }
    for (const ch of word.replace(/\s+/g, '')) {
      const idx = spellFirstUnusedIndexForLetter(ch)
      if (idx < 0) break
      spellAddLetterByIndex(idx)
      if (feedback.value || answering.value) break
    }
  }
}

function spellFirstUnusedIndexForLetter(ch: string): number {
  const letters = currentCard.value?.letters ?? []
  const used = spellUsedIndices.value
  for (let i = 0; i < letters.length; i++) {
    if (used.includes(i)) continue
    if (letters[i].toLowerCase() === ch.toLowerCase()) return i
  }
  return -1
}

const estimatedTime = computed(() => {
  const cards = stats.value.availableForTraining
  if (cards === 0) return null
  
  // Average 15 seconds per card (same as notification service)
  const avgSecondsPerCard = 15
  const totalSeconds = cards * avgSecondsPerCard
  const minutes = Math.floor(totalSeconds / 60)
  
  if (minutes < 1) {
    return t('training.lessThanMinute')
  } else if (minutes === 1) {
    return t('training.oneMinute')
  } else {
    return t('training.minutes', { minutes })
  }
})

const estimatedTimeForRemaining = computed(() => {
  const cards = stats.value.availableForTraining
  if (cards === 0) return null
  
  // Average 15 seconds per card (same as notification service)
  const avgSecondsPerCard = 15
  const totalSeconds = cards * avgSecondsPerCard
  const minutes = Math.floor(totalSeconds / 60)
  
  if (minutes < 1) {
    return t('training.oneMin')
  } else if (minutes === 1) {
    return t('training.oneMin')
  } else {
    return t('training.min', { minutes })
  }
})

// Calculate accuracy percentage
const accuracyPercentage = computed(() => {
  if (trainingStats.value.totalCards === 0) return 0
  return Math.round((trainingStats.value.correctCards / trainingStats.value.totalCards) * 100)
})

// Percentage circle calculations
const circumference = computed(() => 2 * Math.PI * 54)
const animatedPercentageOffset = ref(0)
const percentageOffset = computed(() => {
  // Use animated offset during animation, computed offset after
  if (!percentageAnimationComplete.value) {
    return animatedPercentageOffset.value
  }
  const progress = animatedPercentage.value / 100
  return circumference.value * (1 - progress)
})

// Color based on percentage
const percentageColor = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return '#10b981' // green
  if (percent >= 70) return '#3b82f6' // blue
  if (percent >= 50) return '#f59e0b' // orange
  return '#ef4444' // red
})

// Generate weights for motivational messages (first 30%, last 1%)
const generateMessageWeights = (count: number) => {
  if (count <= 0) return []
  if (count === 1) return [100]
  const weights: number[] = []
  const maxWeight = 30 // 30%
  const minWeight = 1 // 1%

  for (let i = 0; i < count; i++) {
    const ratio = i / (count - 1)
    const weight = maxWeight * Math.pow(minWeight / maxWeight, ratio)
    weights.push(weight)
  }

  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map(w => w * 100 / sum)
}

const pickNonEmpty = (preferred: string[], fallback: string[]): string[] =>
  preferred.length > 0 ? preferred : fallback

const motivationalMessage = computed(() => {
  const percent = accuracyPercentage.value
  const b = motivationalBuckets.value
  let messages: string[]

  if (percent >= 95) {
    messages = pickNonEmpty(b.excellent, b.great)
  } else if (percent >= 90) {
    messages = pickNonEmpty(b.great, b.good)
  } else if (percent >= 80) {
    messages = pickNonEmpty(b.good, b.okay)
  } else if (percent >= 70) {
    messages = pickNonEmpty(b.okay, b.needsWork)
  } else if (percent >= 50) {
    messages = pickNonEmpty(b.needsWork, b.poor)
  } else {
    messages = pickNonEmpty(b.poor, b.needsWork)
  }

  if (!messages.length) return ''
  return getWeightedMessage(messages)
})

const messageClass = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return 'message-excellent'
  if (percent >= 70) return 'message-good'
  if (percent >= 50) return 'message-okay'
  return 'message-needs-improvement'
})

// Confetti styles - start from random points on circle edge
const getConfettiStyle = (index: number) => {
  const colors = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
  const color = colors[index % colors.length]
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: viewBox 120x120, wrapper 200x200px, so scale = 200/120 = 1.67
  // Radius in viewBox = 54, so real radius = 54 * 1.67 ≈ 90px
  const circleRadius = 90
  // Start position on circle edge
  const startX = Math.cos(startAngleRad) * circleRadius
  const startY = Math.sin(startAngleRad) * circleRadius
  // Random direction outward
  const directionAngle = startAngle + (Math.random() - 0.5) * 40 // ±20 degrees variation
  const directionAngleRad = (directionAngle * Math.PI) / 180
  // Distance to travel
  const distance = 150 + Math.random() * 100 // 150-250px
  const endX = Math.cos(directionAngleRad) * distance
  const endY = Math.sin(directionAngleRad) * distance
  // Random delay and duration
  const delay = Math.random() * 0.5
  const duration = 1.5 + Math.random() * 1
  return {
    '--confetti-color': color,
    '--confetti-start-x': `${startX}px`,
    '--confetti-start-y': `${startY}px`,
    '--confetti-end-x': `${endX}px`,
    '--confetti-end-y': `${endY}px`,
    '--confetti-delay': `${delay}s`,
    '--confetti-duration': `${duration}s`
  }
}

// Failure item styles - falling poop from top
const getFailureItemStyle = (index: number) => {
  // Start from random position at top (well above visible area)
  const startX = (Math.random() - 0.5) * 400 // -200 to 200px from center
  const startY = -200 - Math.random() * 150 // Start well above screen
  
  // End position - fall down and slightly outward
  const endX = startX + (Math.random() - 0.5) * 100 // Some horizontal drift
  const endY = 250 + Math.random() * 150 // Fall well below center
  
  // Rotation and scale
  const rotation = (Math.random() - 0.5) * 720 // Random rotation
  const scale = 0.6 + Math.random() * 0.4 // 0.6-1.0
  
  // Timing - staggered fall, ensure they start invisible and well above
  const delay = 0.2 + Math.random() * 0.6 // Start after 0.2s to avoid center flash
  const duration = 1.8 + Math.random() * 0.7 // 1.8-2.5s
  
  return {
    '--start-x': `${startX}px`,
    '--start-y': `${startY}px`,
    '--end-x': `${endX}px`,
    '--end-y': `${endY}px`,
    '--rotation': `${rotation}deg`,
    '--scale': scale,
    '--delay': `${delay}s`,
    '--duration': `${duration}s`
  }
}

// Success particle styles - explode outward from center
const getSuccessParticleStyle = (index: number) => {
  // Random angle (0-360 degrees)
  const angle = (index * 30) + Math.random() * 15 // Spread evenly with some randomness
  const angleRad = (angle * Math.PI) / 180
  // Distance to travel
  const distance = 60 + Math.random() * 40 // 60-100px
  const endX = Math.cos(angleRad) * distance
  const endY = Math.sin(angleRad) * distance
  // Random size
  const size = 4 + Math.random() * 4 // 4-8px
  // Random delay
  const delay = Math.random() * 0.2
  return {
    '--particle-end-x': `${endX}px`,
    '--particle-end-y': `${endY}px`,
    '--particle-size': `${size}px`,
    '--particle-delay': `${delay}s`
  }
}

// Firework styles - start from random points on circle edge
const getFireworkStyle = (index: number) => {
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: viewBox 120x120, wrapper 200x200px, so scale = 200/120 = 1.67
  // Radius in viewBox = 54, so real radius = 54 * 1.67 ≈ 90px
  const circleRadius = 90
  // Start position on circle edge
  const startX = Math.cos(startAngleRad) * circleRadius
  const startY = Math.sin(startAngleRad) * circleRadius
  // Random delay for staggered explosion
  const delay = Math.random() * 1.2
  // Random size variation
  const size = 0.8 + Math.random() * 0.4
  return {
    '--firework-x': `${startX}px`,
    '--firework-y': `${startY}px`,
    '--delay': `${delay}s`,
    '--firework-size': size
  }
}

// Animate percentage when training completes
watch(() => sessionComplete.value, (complete) => {
  if (complete) {
    animatedPercentage.value = 0
    animatedPercentageOffset.value = circumference.value // Start from full (empty circle)
    percentageAnimationComplete.value = false
    const target = accuracyPercentage.value
    const duration = 1500 // 1.5 seconds
    const startTime = Date.now()
    
    const animate = () => {
      const elapsed = Date.now() - startTime
      const progress = Math.min(elapsed / duration, 1)
      // Easing function (ease-out cubic) - matches CSS cubic-bezier(0.4, 0, 0.2, 1)
      // CSS cubic-bezier(0.4, 0, 0.2, 1) approximates to ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3)
      animatedPercentage.value = Math.round(target * eased)
      // Animate circle offset simultaneously
      animatedPercentageOffset.value = circumference.value * (1 - eased * target / 100)
      
      if (progress < 1) {
        requestAnimationFrame(animate)
      } else {
        animatedPercentage.value = target
        animatedPercentageOffset.value = circumference.value * (1 - target / 100)
        percentageAnimationComplete.value = true
      }
    }
    
    requestAnimationFrame(animate)
  }
})

// Play victory/defeat melodies when animations start
// Watch for locale changes and rebuild chart to update labels
watch(() => currentLocale.value, () => {
  if (upcomingCardsLoaded.value && upcomingCardsData.value && Object.keys(upcomingCardsData.value).length > 0) {
    updateUpcomingChart()
  }
})

watch([() => percentageAnimationComplete.value, () => accuracyPercentage.value], ([complete, percentage]) => {
  if (!complete || !settings.value.soundsEnabled) return
  
  if (percentage > 90) {
    // Victory - play when fireworks animation starts
    playVictory(settings.value.soundTheme)
  } else if (percentage < 10) {
    // Defeat - play when failure animation starts
    playDefeat(settings.value.soundTheme)
  }
})

// Calculate progress for circular progress bar (delay timer)
const delayCircumference = computed(() => {
  const radius = 34
  return 2 * Math.PI * radius
})

const strokeDashoffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0) {
    return delayCircumference.value
  }
  // Calculate progress based on remaining milliseconds for precision
  const progress = remainingMs.value / initialDelayMs.value
  return delayCircumference.value * (1 - progress)
})

// Calculate progress for error circular progress bar (5 seconds countdown)
const errorCircumference = computed(() => {
  const radius = 16
  return 2 * Math.PI * radius
})

const errorProgressOffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0 || feedback.value?.is_correct) {
    return 0
  }
  // For incorrect answers: progress goes from 0 to full (reverse countdown)
  const progress = remainingMs.value / initialDelayMs.value
  return errorCircumference.value * (1 - progress)
})

const cardIndex = ref(0)
const totalCards = ref(0)
const userCardId = ref(0)

// Generate weights helper function
const generateWeights = (phrases: string[]) => {
  const n = phrases.length
  if (n <= 0) return []
  if (n === 1) return [100]
  const weights: number[] = []
  const maxWeight = 30 // 30%
  const minWeight = 0.01 // 0.01%

  for (let i = 0; i < n; i++) {
    const ratio = i / (n - 1)
    const weight = maxWeight * Math.pow(minWeight / maxWeight, ratio)
    weights.push(weight)
  }

  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map(w => w * 100 / sum)
}

// Cumulative distribution helper function
const generateCumulativeWeights = (weights: number[]) => {
  const cumulative: number[] = []
  let sum = 0
  for (const weight of weights) {
    sum += weight
    cumulative.push(sum)
  }
  return cumulative
}

// Get weighted random message for motivational messages
const getWeightedMessage = (messages: string[]): string => {
  if (!messages.length) return ''
  if (messages.length === 1) return messages[0]
  const weights = generateMessageWeights(messages.length)
  const cumulative = generateCumulativeWeights(weights)

  const random = Math.random() * 100
  for (let i = 0; i < cumulative.length; i++) {
    if (random <= cumulative[i]) {
      return messages[i]
    }
  }
  return messages[0]
}

// Get random encouraging phrase based on weighted distribution
const getRandomEncouragingPhrase = (): string => {
  const phrases = encouragingPhrasesList.value
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

// Get random disappointing phrase based on weighted distribution
const getRandomDisappointingPhrase = (): string => {
  const phrases = disappointingPhrasesList.value
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

watch(
  [encouragingPhrasesList, disappointingPhrasesList, locale],
  () => {
    currentEncouragingPhrase.value = getRandomEncouragingPhrase()
    currentDisappointingPhrase.value = getRandomDisappointingPhrase()
  },
  { immediate: true }
)

// Timer for automatic options reveal
let autoRevealTimer: ReturnType<typeof setTimeout> | null = null
// Timer for automatic next card transition
let autoNextCardTimer: ReturnType<typeof setTimeout> | null = null
const cardShownAt = ref<Date | null>(null)

// Timer pause state
const timerPaused = ref(false)
let timerPauseStartTime: number | null = null
let timerPausedRemainingMs: number | null = null
let countdownAnimationFrameId: number | null = null
let timerEndTime: number | null = null
let autoNextCardTimerStartTime: number | null = null
let autoNextCardTimerDelayMs: number | null = null
let spellAnswerLettersResizeObserver: ResizeObserver | null = null

// Process question to wrap transcription in span if not already wrapped
const processedQuestion = computed(() => {
  if (!currentCard.value?.question) return ''
  
  let question = currentCard.value.question
  
  // Pattern to match transcription: /.../ after </strong>
  // Match: </strong> /.../
  if (!question.includes('<span class="transcription">')) {
    const transcriptionPattern = /(<\/strong>)\s*(\/[^\/]+\/)/g
    question = question.replace(transcriptionPattern, '$1 <span class="transcription">$2</span>')
  }

  if (
    showMorphInTraining.value &&
    isTargetLangSide.value &&
    morphDisplay.value.kind === 'noun' &&
    morphDisplay.value.article
  ) {
    const article = morphDisplay.value.article.trim()
    question = question.replace(/<strong>(.*?)<\/strong>/, (_, inner: string) => {
      const raw = (inner || '').trim()
      if (!raw) return `<strong>${article}</strong>`
      const lower = raw.toLowerCase()
      if (lower.startsWith(`${article.toLowerCase()} `)) return `<strong>${raw}</strong>`
      return `<strong>${article} ${raw}</strong>`
    })
  }
  
  return question
})

const handleKeyPress = (event: KeyboardEvent) => {
  if (!sessionActive.value || feedback.value || answering.value) return

  // Spell: only keys from available letters, or Backspace
  if (currentCard.value?.type === 'spell' && currentCard.value?.letters?.length) {
    const key = event.key
    if (key === 'Backspace') {
      if (spellAnswerLetters.value.length > 0) {
        event.preventDefault()
        spellRemoveLetterAt(spellAnswerLetters.value.length - 1)
      }
      return
    }
    if (key.length === 1) {
      const idx = spellFirstUnusedIndexForLetter(key)
      if (idx >= 0) {
        event.preventDefault()
        spellAddLetterByIndex(idx)
      }
    }
    return
  }

  // Type: Enter to submit (already on input)
  if (currentCard.value?.type === 'type') return

  // Options: number keys 1-4
  if (!optionsShown.value) return
  const key = event.key
  if (key >= '1' && key <= '4') {
    const optionIndex = parseInt(key) - 1
    if (optionIndex >= 0 && optionIndex < options.value.length) {
      event.preventDefault()
      submitAnswer(optionIndex)
    }
  }
}

interface TrainingSettingsResponse {
  settings?: {
    autoplay_pronunciation?: boolean
  }
}

const loadTrainingUISettings = async () => {
  try {
    const data = await apiClient.request<TrainingSettingsResponse>('/api/settings')
    const autoplay = data.settings?.autoplay_pronunciation
    setAutoplayPronunciation(autoplay === undefined ? true : autoplay)
  } catch (error) {
    console.error('Failed to load training UI settings:', error)
  }
}

const dismissNetworkError = () => {
  networkError.value = false
  networkErrorRetrying.value = false
  if (networkErrorHideTimer) {
    clearTimeout(networkErrorHideTimer)
    networkErrorHideTimer = null
  }
}

const handleNetworkChange = () => {
  isOnline.value = typeof navigator === 'undefined' ? true : navigator.onLine
}

onMounted(async () => {
  // Set up network error callback
  apiClient.setNetworkErrorCallback((isRetrying: boolean, attempt: number, maxAttempts: number) => {
    if (typeof navigator !== 'undefined' && navigator.onLine === false) {
      dismissNetworkError()
      return
    }
    networkError.value = true
    networkErrorRetrying.value = isRetrying
    networkErrorAttempt.value = attempt
    networkErrorMaxAttempts.value = maxAttempts
    if (networkErrorHideTimer) clearTimeout(networkErrorHideTimer)
    networkErrorHideTimer = setTimeout(dismissNetworkError, isRetrying ? 7000 : 4500)
  })
  
  // Set up network success callback to hide error notification
  apiClient.setNetworkSuccessCallback(() => {
    dismissNetworkError()
  })
  
  // Add keyboard event listener
  window.addEventListener('keydown', handleKeyPress)
  window.addEventListener('online', handleNetworkChange)
  window.addEventListener('offline', handleNetworkChange)
  
  await Promise.all([ensureLearningLoaded(), loadTrainingUISettings(), loadStats(), loadUpcomingCards(), checkCurrentSession()])

  // Spell: scale collected letters to fit container width
  watch(
    () => spellAnswerLettersContainerRef.value,
    (el) => {
      if (spellAnswerLettersResizeObserver) {
        spellAnswerLettersResizeObserver.disconnect()
        spellAnswerLettersResizeObserver = null
      }
      if (el) {
        spellAnswerLettersResizeObserver = new ResizeObserver(() => updateSpellScale())
        spellAnswerLettersResizeObserver.observe(el)
        nextTick().then(updateSpellScale)
      }
    },
    { immediate: true }
  )
  watch(
    () => [spellAnswerLetters.value.length, spellRevealLetters.value.length, currentCard.value?.type],
    () => {
      if (currentCard.value?.type === 'spell') nextTick().then(updateSpellScale)
    },
    { deep: true }
  )
})

const loadStats = async () => {
  try {
    const data: {
      due_count: number
      available_for_training?: number
    } = await wordTrainingClient.getDashboard()
    stats.value.dueCount = data.due_count || 0
    stats.value.availableForTraining = data.available_for_training || data.due_count || 0
    statsLoaded.value = true
  } catch (error) {
    console.error('Failed to load stats:', error)
    statsLoaded.value = true // Mark as loaded even on error to avoid infinite loading state
  }
  await refreshVerbFormsPoolCount()
}

const loadUpcomingCards = async () => {
  try {
    const data = await wordTrainingClient.getUpcoming()
    console.log('Upcoming cards data:', data)
    
    // Ensure data is in correct format
    if (data && typeof data === 'object') {
      upcomingCardsData.value = data
    } else {
      console.warn('Invalid data format:', data)
      upcomingCardsData.value = {}
    }
    
    upcomingCardsLoaded.value = true
    await nextTick()
    setTimeout(() => {
      updateUpcomingChart()
    }, 150)
  } catch (error) {
    console.error('Failed to load upcoming cards:', error)
    upcomingCardsLoaded.value = true // Mark as loaded even on error
  }
}

const updateUpcomingChart = () => {
  if (!upcomingChartCanvas.value) {
    setTimeout(() => {
      if (upcomingChartCanvas.value) {
        updateUpcomingChart()
      }
    }, 100)
    return
  }
  
  if (!upcomingCardsData.value || Object.keys(upcomingCardsData.value).length === 0) {
    return
  }
  
  // Destroy existing chart if it exists
  if (upcomingChartInstance) {
    upcomingChartInstance.destroy()
    upcomingChartInstance = null
  }
  
  // Prepare data - ensure we process dates in order
  const dates = Object.keys(upcomingCardsData.value).sort()
  const labels: string[] = []
  const counts: number[] = []
  
  console.log('Processing upcoming cards data:', upcomingCardsData.value)
  console.log('Sorted dates:', dates)
  
  dates.forEach(date => {
    const item = upcomingCardsData.value[date]
    console.log(`Processing date ${date}:`, item)
    if (item && typeof item === 'object' && 'label' in item && 'count' in item) {
      labels.push(item.label)
      counts.push(item.count)
      console.log(`Added: label="${item.label}", count=${item.count}`)
    } else {
      // Fallback if data structure is different
      console.warn('Unexpected data structure for date:', date, item)
      // Try to extract date part for display
      const datePart = date.split('T')[0] || date
      labels.push(datePart)
      counts.push(0)
    }
  })
  
  console.log('Final chart data - labels:', labels, 'counts:', counts)
  
  // Get theme colors
  const root = getComputedStyle(document.documentElement)
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  const primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
  const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
  const textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
  const borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'
  
  // Convert hex to rgba
  const hexToRgba = (hex: string, alpha: number) => {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }
  
  // Create bar chart
  upcomingChartInstance = new Chart(upcomingChartCanvas.value, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [{
        label: (t as any)('common.cards', 2),
        data: counts,
        backgroundColor: hexToRgba(primaryColor, isDark ? 0.7 : 0.6),
        borderColor: primaryColor,
        borderWidth: 1
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          titleColor: '#fff',
          bodyColor: '#fff',
          borderColor: borderColor,
          borderWidth: 1,
          padding: 12,
          callbacks: {
            label: function(context) {
              const value = context.parsed.y || 0
              return t('training.chartCardsTooltip', value, { count: value })
            }
          }
        }
      },
      scales: {
        x: {
          ticks: {
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            }
          },
          grid: {
            color: borderColor,
            display: false
          }
        },
        y: {
          type: 'linear',
          display: true,
          beginAtZero: true,
          ticks: {
            stepSize: 1,
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            },
            callback: function(value) {
              return Number.isInteger(value) ? value : ''
            }
          },
          grid: {
            color: isDark ? borderColor : '#e0e0e0'
          }
        }
      }
    }
  })
}

onUnmounted(() => {
  // Remove keyboard event listener
  window.removeEventListener('keydown', handleKeyPress)
  window.removeEventListener('online', handleNetworkChange)
  window.removeEventListener('offline', handleNetworkChange)
  
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (exampleButtonTimer) {
    clearTimeout(exampleButtonTimer)
    exampleButtonTimer = null
  }
  if (typeHintButtonTimer) {
    clearTimeout(typeHintButtonTimer)
    typeHintButtonTimer = null
  }
  if (spellHintButtonTimer) {
    clearTimeout(spellHintButtonTimer)
    spellHintButtonTimer = null
  }
  typeRevealTimeouts.forEach(clearTimeout)
  typeRevealTimeouts = []
  if (spellAnswerLettersResizeObserver) {
    spellAnswerLettersResizeObserver.disconnect()
    spellAnswerLettersResizeObserver = null
  }
  dismissNetworkError()

  // Destroy chart
  if (upcomingChartInstance) {
    upcomingChartInstance.destroy()
    upcomingChartInstance = null
  }
})

const checkCurrentSession = async () => {
  try {
    const response = await wordTrainingClient.current()
    
    // No active session (HTTP 200)
    if (response && typeof response === 'object' && 'active' in response && (response as any).active === false) {
      sessionActive.value = false
      currentCardGeneration++
      currentCard.value = null
      return
    }
    
    // Training complete response (HTTP 200)
    if (response && typeof response === 'object' && 'complete' in response) {
      sessionComplete.value = true
      sessionActive.value = false
      currentCardGeneration++
      currentCard.value = null
      await loadStats()
      return
    }
    
    const card = response as Card
    sessionActive.value = true
    setupCard(card)
    void prefetchNextTrainingCard()
  } catch (error: any) {
    console.error('Failed to check session:', error)
  }
}

const isTrainingCompleteResponse = (response: any): response is { complete: boolean; cards_completed: number; total_cards?: number; correct_cards?: number } =>
  !!response && typeof response === 'object' && 'complete' in response

const isTrainingInactiveResponse = (response: any): response is { active: false } =>
  !!response && typeof response === 'object' && 'active' in response && response.active === false

const applyTrainingSessionResponse = async (response: any): Promise<boolean> => {
  if (isTrainingCompleteResponse(response)) {
    sessionComplete.value = true
    cardsCompleted.value = response.cards_completed || 0
    trainingStats.value = {
      totalCards: response.total_cards || response.cards_completed || 0,
      correctCards: response.correct_cards || 0,
    }
    sessionActive.value = false
    currentCardGeneration++
    currentCard.value = null
    prefetchedCardResponse.value = null
    await loadStats()
    return false
  }

  if (isTrainingInactiveResponse(response)) {
    sessionActive.value = false
    currentCardGeneration++
    currentCard.value = null
    prefetchedCardResponse.value = null
    await loadStats()
    await showAlert(t('training.noActiveSession'))
    return false
  }

  const card = response as Card
  setupCard(card)
  void prefetchNextTrainingCard()

  if (card.card_index > card.total_cards) {
    sessionComplete.value = true
    cardsCompleted.value = card.card_index - 1
    try {
      const statsResponse = await wordTrainingClient.current()
      if (isTrainingCompleteResponse(statsResponse)) {
        trainingStats.value = {
          totalCards: statsResponse.total_cards || card.card_index - 1,
          correctCards: statsResponse.correct_cards || 0,
        }
      } else {
        trainingStats.value = {
          totalCards: card.card_index - 1,
          correctCards: 0,
        }
      }
    } catch {
      trainingStats.value = {
        totalCards: card.card_index - 1,
        correctCards: 0,
      }
    }
    sessionActive.value = false
    currentCardGeneration++
    currentCard.value = null
    prefetchedCardResponse.value = null
    await loadStats()
    return false
  }

  return true
}

const prefetchNextTrainingCard = () => {
  if (!sessionActive.value || prefetchInFlight) return
  const generation = currentCardGeneration
  prefetchInFlight = (async () => {
    try {
      const response = await wordTrainingClient.prefetchNext()
      if (!sessionActive.value || generation !== currentCardGeneration) return
      prefetchedCardResponse.value = response
    } catch (error) {
      console.error('Failed to prefetch next training card:', error)
    } finally {
      prefetchInFlight = null
    }
  })()
}

const syncCurrentCardState = (): Promise<void> => {
  if (syncCurrentInFlight) return syncCurrentInFlight
  syncCurrentInFlight = (async () => {
    try {
      await wordTrainingClient.current()
    } catch (error) {
      console.error('Failed to sync current training card state:', error)
    } finally {
      syncCurrentInFlight = null
    }
  })()
  return syncCurrentInFlight
}

const setupCard = (card: Card) => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }
  if (exampleButtonTimer) {
    clearTimeout(exampleButtonTimer)
    exampleButtonTimer = null
  }
  if (typeHintButtonTimer) {
    clearTimeout(typeHintButtonTimer)
    typeHintButtonTimer = null
  }
  if (spellHintButtonTimer) {
    clearTimeout(spellHintButtonTimer)
    spellHintButtonTimer = null
  }

  currentCardGeneration++
  currentCard.value = card
  cardIndex.value = card.card_index
  totalCards.value = card.total_cards
  userCardId.value = card.user_card_id ?? 0
  optionsShown.value = false
  options.value = []
  feedback.value = null
  chosenOptionIndex.value = null
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  timerEndTime = null
  autoNextCardTimerStartTime = null
  autoNextCardTimerDelayMs = null
  cardShownAt.value = new Date()
  showExampleButton.value = false
  showExampleButtonVisible.value = false
  exampleUsageShown.value = false
  spellAnswerLetters.value = []
  spellUsedIndices.value = []
  spellRevealLetters.value = []
  spellSkipAutoPickInProgress.value = false
  spellSkipResultActive.value = false
  spellHintShown.value = false
  showSpellHintButton.value = false
  spellHintButtonVisible.value = false
  typeAnswerText.value = ''
  typeHintShown.value = false
  showTypeHintButton.value = false
  typeHintButtonVisible.value = false
  typeRevealDisplayText.value = ''
  typeRevealTimeouts.forEach(clearTimeout)
  typeRevealTimeouts = []

  // Spell cards: no options delay, letters shown immediately
  if (card.type === 'spell') {
    optionsShown.value = true
    const stem = typeChallengeStemFromAnswer(card.correct_answer, card.prefix)
    if ([...stem].length > 0) {
      spellHintButtonTimer = setTimeout(() => {
        showSpellHintButton.value = true
        setTimeout(() => {
          spellHintButtonVisible.value = true
        }, 50)
      }, 2000)
    }
    return
  }
  // Type cards: no options, input shown immediately; hint button with delay like example
  if (card.type === 'type') {
    optionsShown.value = true
    if (currentCard.value?.hint_first_letter !== undefined && currentCard.value?.hint_length != null) {
      typeHintButtonTimer = setTimeout(() => {
        showTypeHintButton.value = true
        setTimeout(() => {
          typeHintButtonVisible.value = true
        }, 50)
      }, 2000)
    }
    nextTick(() => {
      typeInputRef.value?.focus()
    })
    return
  }

  // Schedule automatic options reveal (or show immediately if delay is 0)
  if (card.delay_ms > 0) {
    autoRevealTimer = setTimeout(() => {
      if (!optionsShown.value) {
        revealOptions(false) // false = not early reveal
      }
    }, card.delay_ms)
  } else {
    // Delay 0: show options on next tick so card is rendered first
    autoRevealTimer = setTimeout(() => {
      if (!optionsShown.value) {
        revealOptions(false)
      }
    }, 0)
  }
}

const startTraining = async () => {
  loading.value = true
  try {
    const card: Card = await wordTrainingClient.start()
    sessionActive.value = true
    setupCard(card)
    void prefetchNextTrainingCard()
    sessionComplete.value = false
    // Card is on screen — hide the loader before refreshing stats so the
    // spinner doesn't keep spinning over the first card.
    loading.value = false
    // Update stats in the background (non-blocking).
    void loadStats()
  } catch (error: any) {
    if (error.message?.includes('No cards available')) {
      await showAlert(t('training.noCardsAvailable'))
    } else {
      console.error('Failed to start training:', error)
      await showAlert(t('training.failedStartTraining'))
    }
  } finally {
    loading.value = false
  }
}

const submitWordReport = async () => {
  if (!currentCard.value || reportSubmitting.value || reportAlreadySent.value) return
  const comment = buildReportComment(
    reportCategory.value,
    reportDetails.value,
    t(`training.reportCategories.${reportCategory.value}`)
  )
  if (!comment) return
  reportMessage.value = ''
  const key = cardReportKey(currentCard.value)
  reportSubmitting.value = true
  try {
    const extra: Record<string, unknown> = {
      question: currentCard.value.question,
      card_type: currentCard.value.type || 'card'
    }
    if (currentCard.value.word_en) extra.word_en = currentCard.value.word_en
    if (currentCard.value.word_ru) extra.word_ru = currentCard.value.word_ru
    const submitResult = await contentReportClient.submit({
      sourceType: 'word_training',
      reportCategory: reportCategory.value,
      comment,
      userCardID: currentCard.value.user_card_id || 0,
      word: currentCard.value.word_en || currentCard.value.display_word || '',
      direction: currentCard.value.direction || '',
      wordCardID: currentCard.value.word_card_id,
      trainingCardID: currentCard.value.training_card_id,
      wordCategory: currentCard.value.word_category || '',
      payload: extra,
    })
    reportSentForCardKey.value = key
    reportMessage.value = submitResult.queued
      ? (t('training.reportQueued') || 'Жалоба сохранена и будет отправлена при появлении сети.')
      : (t('training.reportThanks') || 'Спасибо, жалоба отправлена.')
    reportDialogOpen.value = false
    reportCategory.value = ''
    reportDetails.value = ''
  } catch (error) {
    console.error('Failed to submit training report:', error)
    reportMessage.value = t('training.reportFailed') || 'Не удалось отправить жалобу'
  } finally {
    reportSubmitting.value = false
  }
}

const openWordReportDialog = () => {
  if (!currentCard.value || reportSubmitting.value || reportAlreadySent.value) return
  reportCategory.value = ''
  reportDetails.value = ''
  reportDialogOpen.value = true
}

const closeWordReportDialog = () => {
  if (reportSubmitting.value) return
  reportDialogOpen.value = false
}

watch(() => cardReportKey(currentCard.value), () => {
  reportMessage.value = ''
  reportDialogOpen.value = false
  reportCategory.value = ''
  reportDetails.value = ''
})

const revealOptions = async (isEarly: boolean = false) => {
  // Clear timer if it exists
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  
  // Clear example button timer if it exists
  if (exampleButtonTimer) {
    clearTimeout(exampleButtonTimer)
    exampleButtonTimer = null
  }

  // If already shown, don't do anything
  if (optionsShown.value) {
    return
  }

  try {
    if (syncCurrentInFlight) {
      await syncCurrentInFlight
    }
    const data: OptionsResponse = await wordTrainingClient.reveal()
    options.value = data.options
    optionsShown.value = true
    
    // Reset example button state
    showExampleButton.value = false
    showExampleButtonVisible.value = false
    exampleUsageShown.value = false
    
    // Show example button after 2 seconds if word is English
    if (isTargetLangSide.value) {
      exampleButtonTimer = setTimeout(() => {
        showExampleButton.value = true
        // Trigger visibility animation after a brief delay
        setTimeout(() => {
          showExampleButtonVisible.value = true
        }, 50)
      }, 2000)
    }
  } catch (error: any) {
    console.error('Failed to reveal options:', error)
    // Network error is already handled by callback, but we should handle other errors
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert(t('training.failedLoadOptions'))
    }
  }
}

const showExampleUsage = () => {
  if (exampleUsageShown.value) return
  
  // Show example if available from current card or feedback
  if (currentCard.value?.example_target || currentCard.value?.example_en || feedback.value?.example_target || feedback.value?.example) {
    exampleUsageShown.value = true
  }
}

// Sound functions using new audio system
const playCorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playSuccess(settings.value.soundTheme)
}

const playIncorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playFail(settings.value.soundTheme)
}

// Haptic feedback helper function
const triggerHapticFeedback = (isCorrect: boolean) => {
  if (!settings.value.vibrationEnabled) return
  
  const tg = (window as any).Telegram?.WebApp
  
  // Try Telegram Web App API first
  if (tg?.HapticFeedback) {
    try {
      const haptic = tg.HapticFeedback
      if (isCorrect) {
        // Success feedback - try notificationOccurred first, then impactOccurred
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('success')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('medium')
        }
      } else {
        // Error feedback - try notificationOccurred first, then impactOccurred
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('error')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('heavy')
        }
      }
      return
    } catch (error) {
      console.warn('Telegram haptic feedback failed:', error)
    }
  }
  
  // Fallback to native Vibration API
  if ('vibrate' in navigator && typeof navigator.vibrate === 'function') {
    try {
      if (isCorrect) {
        // Short, pleasant vibration for correct answer
        navigator.vibrate(50)
      } else {
        // Longer, more noticeable vibration for incorrect answer
        navigator.vibrate([100, 50, 100])
      }
    } catch (error) {
      console.warn('Native vibration failed:', error)
    }
  }
}

const spellAddLetterByIndex = (letterIndex: number) => {
  if (feedback.value || answering.value) return
  if (spellUsedIndices.value.includes(letterIndex)) return
  const letters = currentCard.value?.letters ?? []
  if (letterIndex < 0 || letterIndex >= letters.length) return
  const ch = letters[letterIndex]
  spellUsedIndices.value = [...spellUsedIndices.value, letterIndex]
  const next = [...spellAnswerLetters.value, ch]
  spellAnswerLetters.value = next
  const expectedLen = letters.length
  if (expectedLen > 0 && next.length === expectedLen) {
    submitSpellAnswer()
  }
}

const spellRemoveLetterAt = (answerPosition: number) => {
  if (feedback.value || answering.value) return
  spellAnswerLetters.value = spellAnswerLetters.value.filter((_, i) => i !== answerPosition)
  spellUsedIndices.value = spellUsedIndices.value.filter((_, i) => i !== answerPosition)
}

const skipSpellAnswer = () => {
  if (feedback.value || answering.value) return
  spellAnswerLetters.value = []
  spellUsedIndices.value = []
  spellRevealLetters.value = []
  submitSpellAnswerAs('', true)
}

const submitSpellAnswer = () => {
  const prefix = currentCard.value?.prefix ?? ''
  submitSpellAnswerAs(prefix + spellAnswerLetters.value.join(''))
}

/** Build correct-order list with keys from wrong-order indices for TransitionGroup move */
function buildSpellRevealLetters(correctAnswer: string, wrongOrder: string[], prefix: string): Array<{ letter: string; key: number }> {
  const afterPrefix = prefix ? correctAnswer.slice(prefix.length) : correctAnswer
  const correctLetters = Array.from(afterPrefix)
  const wrong = [...wrongOrder]
  const used = new Set<number>()
  return correctLetters.map((letter) => {
    const idx = wrong.findIndex((c, i) => (c === letter || c.toLowerCase() === letter.toLowerCase()) && !used.has(i))
    if (idx >= 0) used.add(idx)
    const key = idx >= 0 ? idx : used.size
    return { letter, key }
  })
}

async function animateSpellSkipAutoPick(correctAnswer: string, prefix: string) {
  const afterPrefix = prefix ? correctAnswer.slice(prefix.length) : correctAnswer
  const targetLetters = Array.from(afterPrefix)
  const sourceLetters = currentCard.value?.letters ?? []
  const usedSource = new Set<number>()
  spellSkipAutoPickInProgress.value = true
  spellAnswerLetters.value = []
  spellUsedIndices.value = []
  for (const targetLetter of targetLetters) {
    const sourceIndex = sourceLetters.findIndex((c, idx) => (c === targetLetter || c.toLowerCase() === targetLetter.toLowerCase()) && !usedSource.has(idx))
    if (sourceIndex >= 0) {
      usedSource.add(sourceIndex)
      spellUsedIndices.value = [...spellUsedIndices.value, sourceIndex]
    }
    spellAnswerLetters.value = [...spellAnswerLetters.value, targetLetter]
    await new Promise((resolve) => setTimeout(resolve, 55))
  }
  spellSkipAutoPickInProgress.value = false
}

const submitSpellAnswerAs = async (answerText: string, isSkip = false) => {
  if (feedback.value || answering.value) return
  const cardAtSubmit = currentCard.value
  const generation = currentCardGeneration
  answering.value = true
  spellSkipResultActive.value = isSkip
  try {
    const formData = new FormData()
    formData.append('answer_text', answerText)
    const data: Feedback = await wordTrainingClient.answer(formData)
    if (generation !== currentCardGeneration || !sameTrainingCard(cardAtSubmit, currentCard.value)) return
    feedback.value = data
    if (!data.is_correct && currentCard.value?.type === 'spell' && data.correct_answer) {
      const prefix = currentCard.value?.prefix ?? ''
      if (isSkip) {
        spellRevealLetters.value = []
        await animateSpellSkipAutoPick(data.correct_answer, prefix)
      } else {
        const wrongOrder = [...spellAnswerLetters.value]
        spellRevealLetters.value = wrongOrder.map((letter, i) => ({ letter, key: i }))
        nextTick(() => {
          spellRevealLetters.value = buildSpellRevealLetters(data.correct_answer, wrongOrder, prefix)
        })
      }
    }
    triggerHapticFeedback(data.is_correct)
    if (data.is_correct) {
      playCorrectSound()
      currentEncouragingPhrase.value = getRandomEncouragingPhrase()
    } else {
      playIncorrectSound()
      currentDisappointingPhrase.value = getRandomDisappointingPhrase()
    }
    void autoplayPronunciationAfterAnswer(data)
    const nextDelayMs = data.is_correct ? 1000 : (data.delay_seconds ?? 0) * 1000
    const isCorrectSpell = data.is_correct && currentCard.value?.type === 'spell'
    if (isCorrectSpell) {
      // При правильном spell не показываем прогрессбар — просто ждём и переходим дальше
      autoNextCardTimer = setTimeout(() => nextCard(), nextDelayMs)
    } else if (nextDelayMs > 0) {
      const totalSeconds = data.is_correct ? 1 : (data.delay_seconds ?? 0)
      initialDelaySeconds.value = totalSeconds
      initialDelayMs.value = nextDelayMs
      delaySeconds.value = totalSeconds
      remainingMs.value = nextDelayMs
      waitingDelay.value = true
      timerPaused.value = false
      timerPauseStartTime = null
      timerPausedRemainingMs = null
      const startTime = Date.now()
      timerEndTime = startTime + nextDelayMs
      const updateCountdown = () => {
        if (!timerEndTime) return
        if (timerPaused.value) {
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
          return
        }
        const now = Date.now()
        const currentRemainingMs = Math.max(0, timerEndTime - now)
        const currentRemainingSeconds = Math.ceil(currentRemainingMs / 1000)
        remainingMs.value = currentRemainingMs
        delaySeconds.value = currentRemainingSeconds
        if (currentRemainingMs > 0) {
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
        } else {
          delaySeconds.value = 0
          remainingMs.value = 0
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
          if (countdownAnimationFrameId) {
            cancelAnimationFrame(countdownAnimationFrameId)
            countdownAnimationFrameId = null
          }
          nextCard()
        }
      }
      countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
      autoNextCardTimer = setTimeout(() => {
        if (countdownAnimationFrameId) {
          cancelAnimationFrame(countdownAnimationFrameId)
          countdownAnimationFrameId = null
        }
        if (waitingDelay.value) {
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          delaySeconds.value = 0
          remainingMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
        }
        nextCard()
      }, nextDelayMs)
    } else {
      autoNextCardTimer = setTimeout(() => nextCard(), data.is_correct ? 1000 : 150)
    }
  } catch (error: any) {
    console.error('Failed to submit spell answer:', error)
    if (!error.isNetworkError) {
      await showAlert(t('training.failedSubmitAnswer'))
    }
  } finally {
    if (generation === currentCardGeneration && sameTrainingCard(cardAtSubmit, currentCard.value)) {
      answering.value = false
    }
  }
}

const submitTypeAnswerAs = async (answerText: string) => {
  if (feedback.value || answering.value) return
  const cardAtSubmit = currentCard.value
  const generation = currentCardGeneration
  answering.value = true
  try {
    const formData = new FormData()
    formData.append('answer_text', answerText)
    const data: Feedback = await wordTrainingClient.answer(formData)
    if (generation !== currentCardGeneration || !sameTrainingCard(cardAtSubmit, currentCard.value)) return
    feedback.value = data
    triggerHapticFeedback(data.is_correct)
    if (data.is_correct) {
      playCorrectSound()
      currentEncouragingPhrase.value = getRandomEncouragingPhrase()
    } else {
      playIncorrectSound()
      currentDisappointingPhrase.value = getRandomDisappointingPhrase()
    }
    void autoplayPronunciationAfterAnswer(data)
    const nextDelayMs = data.is_correct ? 1000 : (data.delay_seconds ?? 0) * 1000
    const startCountdownOrNext = () => {
      if (nextDelayMs > 0) {
        const totalSeconds = data.is_correct ? 1 : (data.delay_seconds ?? 0)
        initialDelaySeconds.value = totalSeconds
        initialDelayMs.value = nextDelayMs
        delaySeconds.value = totalSeconds
        remainingMs.value = nextDelayMs
        waitingDelay.value = true
        timerPaused.value = false
        timerPauseStartTime = null
        timerPausedRemainingMs = null
        const startTime = Date.now()
        timerEndTime = startTime + nextDelayMs
        const updateCountdown = () => {
          if (!timerEndTime) return
          if (timerPaused.value) {
            countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
            return
          }
          const now = Date.now()
          const currentRemainingMs = Math.max(0, timerEndTime - now)
          const currentRemainingSeconds = Math.ceil(currentRemainingMs / 1000)
          remainingMs.value = currentRemainingMs
          delaySeconds.value = currentRemainingSeconds
          if (currentRemainingMs > 0) {
            countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
          } else {
            delaySeconds.value = 0
            remainingMs.value = 0
            waitingDelay.value = false
            initialDelaySeconds.value = 0
            initialDelayMs.value = 0
            timerPaused.value = false
            timerPauseStartTime = null
            timerPausedRemainingMs = null
            timerEndTime = null
            if (countdownAnimationFrameId) {
              cancelAnimationFrame(countdownAnimationFrameId)
              countdownAnimationFrameId = null
            }
            nextCard()
          }
        }
        countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
        autoNextCardTimer = setTimeout(() => {
          if (countdownAnimationFrameId) {
            cancelAnimationFrame(countdownAnimationFrameId)
            countdownAnimationFrameId = null
          }
          if (waitingDelay.value) {
            waitingDelay.value = false
            initialDelaySeconds.value = 0
            initialDelayMs.value = 0
            delaySeconds.value = 0
            remainingMs.value = 0
            timerPaused.value = false
            timerPauseStartTime = null
            timerPausedRemainingMs = null
            timerEndTime = null
          }
          nextCard()
        }, nextDelayMs)
      } else {
        autoNextCardTimer = setTimeout(() => nextCard(), data.is_correct ? 1000 : 150)
      }
    }
    if (!data.is_correct && currentCard.value?.type === 'type' && data.correct_answer) {
      nextTick(() => {
        startTypeRevealAnimation(data.chosen_option ?? '', data.correct_answer, startCountdownOrNext)
      })
    } else {
      startCountdownOrNext()
    }
  } catch (error: any) {
    console.error('Failed to submit type answer:', error)
    if (!error.isNetworkError) {
      await showAlert(t('training.failedSubmitAnswer'))
    }
  } finally {
    if (generation === currentCardGeneration && sameTrainingCard(cardAtSubmit, currentCard.value)) {
      answering.value = false
    }
  }
}

const submitTypeAnswer = async () => {
  if (!typeAnswerText.value) return
  const prefix = currentCard.value?.prefix ?? ''
  await submitTypeAnswerAs(prefix + typeAnswerText.value)
}

const skipTypeAnswer = () => {
  submitTypeAnswerAs('')
}

function startTypeRevealAnimation(wrongAnswer: string, correctAnswer: string, onComplete?: () => void) {
  typeRevealTimeouts.forEach(clearTimeout)
  typeRevealTimeouts = []
  const wrong = wrongAnswer ?? ''
  const correct = correctAnswer ?? ''
  typeRevealDisplayText.value = wrong
  if (wrong.length === 0 && correct.length === 0) {
    onComplete?.()
    return
  }
  const eraseInterval = 50
  const typeInterval = 80
  const pauseAfterErase = 250
  let eraseStep = 0
  let typeStep = 0
  function run() {
    if (wrong.length > 0 && eraseStep <= wrong.length) {
      if (eraseStep === 0) {
        eraseStep = 1
        typeRevealTimeouts.push(setTimeout(run, eraseInterval))
        return
      }
      typeRevealDisplayText.value = wrong.slice(0, wrong.length - eraseStep)
      eraseStep++
      if (eraseStep <= wrong.length) {
        typeRevealTimeouts.push(setTimeout(run, eraseInterval))
      } else {
        typeRevealTimeouts.push(setTimeout(run, pauseAfterErase))
      }
      return
    }
    if (typeStep < correct.length) {
      typeRevealDisplayText.value = correct.slice(0, typeStep + 1)
      typeStep++
      if (typeStep < correct.length) {
        typeRevealTimeouts.push(setTimeout(run, typeInterval))
      } else {
        onComplete?.()
      }
    }
  }
  typeRevealTimeouts.push(setTimeout(run, wrong.length > 0 ? 400 : 200))
}

const submitAnswer = async (optionIndex: number) => {
  if (feedback.value || answering.value || optionIndex < 0 || optionIndex >= options.value.length) return
  const cardAtSubmit = currentCard.value
  const generation = currentCardGeneration
  answering.value = true
  try {
    const formData = new FormData()
    formData.append('option_index', optionIndex.toString())
    formData.append('user_card_id', userCardId.value.toString())
    
    const data: Feedback = await wordTrainingClient.answer(formData)
    if (generation !== currentCardGeneration || !sameTrainingCard(cardAtSubmit, currentCard.value)) return
    chosenOptionIndex.value = optionIndex
    feedback.value = data
    
    // Hide example if it was shown before answer (to avoid duplicate display in feedback)
    exampleUsageShown.value = false
    
    // Trigger haptic feedback based on answer correctness
    triggerHapticFeedback(data.is_correct)
    
    // Play sound based on answer correctness
    if (data.is_correct) {
      playCorrectSound()
    } else {
      playIncorrectSound()
    }
    void autoplayPronunciationAfterAnswer(data)
    
    // Generate random phrase based on answer correctness
    if (data.is_correct) {
      currentEncouragingPhrase.value = getRandomEncouragingPhrase()
    } else {
      currentDisappointingPhrase.value = getRandomDisappointingPhrase()
    }
    
    // Schedule automatic transition to next card
    const delayMs = data.delay_seconds ? data.delay_seconds * 1000 : 0
    
    if (delayMs > 0) {
      const totalSeconds = data.delay_seconds!
      initialDelaySeconds.value = totalSeconds
      initialDelayMs.value = delayMs
      delaySeconds.value = totalSeconds
      remainingMs.value = delayMs
      waitingDelay.value = true
      
      // Reset pause state
      timerPaused.value = false
      timerPauseStartTime = null
      timerPausedRemainingMs = null
      
      const startTime = Date.now()
      timerEndTime = startTime + delayMs
      
      // Update countdown with precise timing using requestAnimationFrame
      const updateCountdown = () => {
        if (!timerEndTime) {
          return
        }
        
        if (timerPaused.value) {
          // Timer is paused, don't update but keep the loop running
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
          return
        }
        
        const now = Date.now()
        const currentRemainingMs = Math.max(0, timerEndTime! - now)
        const currentRemainingSeconds = Math.ceil(currentRemainingMs / 1000)
        
        remainingMs.value = currentRemainingMs
        delaySeconds.value = currentRemainingSeconds
        
        if (currentRemainingMs > 0) {
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
        } else {
          delaySeconds.value = 0
          remainingMs.value = 0
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
          autoNextCardTimerStartTime = null
          autoNextCardTimerDelayMs = null
          if (countdownAnimationFrameId) {
            cancelAnimationFrame(countdownAnimationFrameId)
            countdownAnimationFrameId = null
          }
          nextCard()
        }
      }
      
      // Start updating immediately
      countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
      
      // Schedule automatic next card as backup
      autoNextCardTimerStartTime = Date.now()
      autoNextCardTimerDelayMs = delayMs
      autoNextCardTimer = setTimeout(() => {
        if (countdownAnimationFrameId) {
          cancelAnimationFrame(countdownAnimationFrameId)
          countdownAnimationFrameId = null
        }
        if (waitingDelay.value) {
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          delaySeconds.value = 0
          remainingMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
        }
        autoNextCardTimerStartTime = null
        autoNextCardTimerDelayMs = null
        nextCard()
      }, delayMs)
    } else {
      // No delay from server: correct answer — ~1s to see success; wrong + delay 0 — minimal pause
      const delayWhenCorrectMs = 1000
      const delayWhenWrongMs = 150
      const nextDelayMs = data.is_correct ? delayWhenCorrectMs : delayWhenWrongMs
      autoNextCardTimerStartTime = Date.now()
      autoNextCardTimerDelayMs = nextDelayMs
      autoNextCardTimer = setTimeout(() => {
        autoNextCardTimerStartTime = null
        autoNextCardTimerDelayMs = null
        nextCard()
      }, nextDelayMs)
    }
  } catch (error: any) {
    console.error('Failed to submit answer:', error)
    // Network error is already handled by callback
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert(t('training.failedSubmitAnswer'))
    }
  } finally {
    if (generation === currentCardGeneration && sameTrainingCard(cardAtSubmit, currentCard.value)) {
      answering.value = false
    }
  }
}

const nextCard = async () => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }
  if (exampleButtonTimer) {
    clearTimeout(exampleButtonTimer)
    exampleButtonTimer = null
  }
  if (typeHintButtonTimer) {
    clearTimeout(typeHintButtonTimer)
    typeHintButtonTimer = null
  }
  if (spellHintButtonTimer) {
    clearTimeout(spellHintButtonTimer)
    spellHintButtonTimer = null
  }
  autoNextCardTimerStartTime = null
  autoNextCardTimerDelayMs = null

  feedback.value = null
  chosenOptionIndex.value = null
  optionsShown.value = false
  options.value = []
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  initialDelaySeconds.value = 0
  cardShownAt.value = null
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  showExampleButton.value = false
  showExampleButtonVisible.value = false
  exampleUsageShown.value = false
  spellAnswerLetters.value = []
  spellUsedIndices.value = []
  spellRevealLetters.value = []
  spellSkipAutoPickInProgress.value = false
  spellSkipResultActive.value = false
  typeAnswerText.value = ''

  try {
    const cached = prefetchedCardResponse.value
    prefetchedCardResponse.value = null
    if (cached) {
      // Show the prefetched card immediately; sync backend session state in the
      // background so reveal/answer use options for the active card, not the previous one.
      void syncCurrentCardState()
      await applyTrainingSessionResponse(cached)
      return
    }

    const response = await wordTrainingClient.current()
    await applyTrainingSessionResponse(response)
  } catch (error: any) {
    console.error('Failed to get next card:', error)
    // Network error is already handled by callback
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert(t('training.failedNextCard'))
    }
  }
}

const resetSession = async () => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }

  sessionActive.value = false
  currentCardGeneration++
  currentCard.value = null
  prefetchedCardResponse.value = null
  syncCurrentInFlight = null
  optionsShown.value = false
  options.value = []
  feedback.value = null
  chosenOptionIndex.value = null
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  sessionComplete.value = false
  cardsCompleted.value = 0
  trainingStats.value = {
    totalCards: 0,
    correctCards: 0
  }
  cardShownAt.value = null
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  timerEndTime = null
  autoNextCardTimerStartTime = null
  autoNextCardTimerDelayMs = null
  
  // Refresh stats and upcoming cards
  await loadStats()
  await loadUpcomingCards()
}

// Timer pause/resume handlers
const pauseTimer = () => {
  if (!waitingDelay.value || timerPaused.value || !timerEndTime) return
  
  timerPaused.value = true
  timerPauseStartTime = Date.now()
  timerPausedRemainingMs = remainingMs.value
  
  // Pause autoNextCardTimer if it exists
  if (autoNextCardTimer && autoNextCardTimerStartTime !== null && autoNextCardTimerDelayMs !== null) {
    const elapsed = Date.now() - autoNextCardTimerStartTime
    const remaining = Math.max(0, autoNextCardTimerDelayMs - elapsed)
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
    // Update delay to remaining time
    autoNextCardTimerDelayMs = remaining
  }
}

const resumeTimer = () => {
  if (!waitingDelay.value || !timerPaused.value || timerPauseStartTime === null || timerPausedRemainingMs === null || !timerEndTime) return
  
  // Calculate how long the timer was paused
  const pauseDuration = Date.now() - timerPauseStartTime
  
  // Adjust the end time by adding the pause duration
  timerEndTime = timerEndTime + pauseDuration
  
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  
  // Resume autoNextCardTimer if it was paused
  if (autoNextCardTimerDelayMs !== null && autoNextCardTimerDelayMs > 0) {
    autoNextCardTimerStartTime = Date.now()
    autoNextCardTimer = setTimeout(() => {
      if (countdownAnimationFrameId) {
        cancelAnimationFrame(countdownAnimationFrameId)
        countdownAnimationFrameId = null
      }
      if (waitingDelay.value) {
        waitingDelay.value = false
        initialDelaySeconds.value = 0
        initialDelayMs.value = 0
        delaySeconds.value = 0
        remainingMs.value = 0
        timerPaused.value = false
        timerPauseStartTime = null
        timerPausedRemainingMs = null
        timerEndTime = null
      }
      autoNextCardTimerStartTime = null
      autoNextCardTimerDelayMs = null
      nextCard()
    }, autoNextCardTimerDelayMs)
  }
}

// Handle mouse/touch events for timer pause
const handleTimerMouseDown = (event: MouseEvent | TouchEvent) => {
  // Only handle timer pause when waitingDelay is active
  if (!waitingDelay.value) {
    return
  }
  
  // When waitingDelay is active, all elements in the card should pause the timer
  // This includes disabled buttons - events are caught on parent container
  // Only prevent if it's an enabled link
  const target = event.target as HTMLElement
  const link = target?.closest('a')
  if (link && !link.hasAttribute('disabled') && !link.hasAttribute('aria-disabled')) {
    return
  }
  
  // Stop propagation to prevent triggering click on disabled buttons
  event.stopPropagation()
  pauseTimer()
}

const handleTimerMouseUp = (event: MouseEvent | TouchEvent) => {
  // Only handle timer resume when waitingDelay is active
  if (!waitingDelay.value) {
    return
  }
  
  // When waitingDelay is active, all elements in the card should resume the timer
  // This includes disabled buttons - events are caught on parent container
  // Only prevent if it's an enabled link
  const target = event.target as HTMLElement
  const link = target?.closest('a')
  if (link && !link.hasAttribute('disabled') && !link.hasAttribute('aria-disabled')) {
    return
  }
  
  // Stop propagation to prevent triggering click on disabled buttons
  event.stopPropagation()
  resumeTimer()
}

const handleTimerMouseLeave = () => {
  // Resume if mouse leaves while button might be pressed
  if (timerPaused.value) {
    resumeTimer()
  }
}
</script>

<style scoped>
.training {
  max-width: 1200px;
  margin: 0 auto;
}

.training h1 {
  margin-bottom: 24px;
}
.training-progress {
  margin-bottom: 20px;
  text-align: center;
}

.card-timer-active {
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}

/* Don't block pointer events - let all elements receive events for timer pause */
/* Buttons will handle their own click events when not disabled */

.question {
  font-size: clamp(17px, 4.6vw, 24px);
  margin: 20px 0;
  text-align: center;
  overflow-x: visible;
  overflow-y: visible;
  word-wrap: break-word;
  overflow-wrap: break-word;
  hyphens: auto;
}

.report-row {
  margin-top: 8px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
}
.report-row-outside {
  margin-top: 10px;
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

.report-message {
  font-size: 12px;
  color: var(--text-secondary);
}

.report-modal-backdrop {
  position: fixed;
  inset: 0;
  background: var(--bg-modal-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10050;
}

.report-modal {
  width: min(92vw, 460px);
  background: var(--card-bg);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  box-shadow: 0 12px 40px var(--card-shadow);
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
  border: 1px solid var(--input-border);
  border-radius: 8px;
  padding: 10px;
  font: inherit;
  color: var(--text-primary);
  background: var(--input-bg);
  box-sizing: border-box;
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
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
}

.report-modal-submit {
  background: var(--color-primary);
  color: var(--text-inverse);
}

.report-modal-submit:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.report-modal-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Make strong element start on a new line */
.question :deep(strong) {
  display: block;
  margin-top: 8px;
  word-break: break-word;
  overflow-wrap: break-word;
}

/* Prevent transcription from breaking into multiple lines */
.question :deep(.transcription),
.question :deep(span.transcription) {
  white-space: nowrap;
  display: none;
}

/* Style transcription in question HTML content */
.question :deep(.transcription),
.question :deep(span.transcription) {
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif !important;
  font-style: italic !important;
  letter-spacing: 0.5px !important;
  font-size: 0.9em !important;
}

/* Style any text that looks like transcription (starts and ends with /) */
.question {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

.question :deep(strong) {
  font-family: inherit;
}

/* Target text nodes after strong that contain transcription pattern */
.question :deep(*) {
  font-family: inherit;
}

.question-meta-row {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 13px;
}

.training-pronunciation-row {
  margin-top: 0;
  margin-bottom: 0;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  line-height: 1;
}

.training-pronunciation-row-standalone {
  margin-top: -8px;
  margin-bottom: 12px;
  width: 100%;
}

.training-transcription {
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif;
  font-style: italic;
  letter-spacing: 0.5px;
  font-size: 0.9em;
  color: var(--text-secondary);
  white-space: nowrap;
}

.question-meta-row .training-transcription {
  font-size: inherit;
  line-height: 1;
}

.question-morph-inline {
  margin-top: 0;
  text-align: center;
  color: var(--text-secondary);
  font-size: inherit;
  line-height: 1;
}

.morph-opposite-line {
  margin-top: 0;
  line-height: 1;
}

.morph-opposite {
  color: var(--text-muted, var(--text-secondary));
}

.morph-gender-m {
  color: #4f7fb5;
}

.morph-gender-f {
  color: #ad5f88;
}

[data-theme="dark"] .morph-gender-m {
  color: #8ab7ee;
}

[data-theme="dark"] .morph-gender-f {
  color: #df9fc2;
}

.btn-pronunciation {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  padding: 0;
}

.btn-pronunciation:hover:not(:disabled) {
  background: var(--bg-hover, rgba(0, 0, 0, 0.06));
}

.btn-pronunciation:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spell-block {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.spell-answer-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  width: 100%;
}
.spell-answer-label {
  font-weight: 600;
  color: var(--text-secondary);
}
.spell-answer-prefix {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  flex-shrink: 0;
  padding: 0 2px;
}
.spell-answer-letters {
  min-height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
  flex-wrap: nowrap;
  padding: 8px 12px;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  box-sizing: border-box;
  overflow: hidden;
}
.spell-answer-letters-inner {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: inherit;
  flex-wrap: nowrap;
  transform-origin: center;
  min-width: min-content;
}
.spell-answer-char-btn {
  flex: 0 1 auto;
  min-width: 0;
  max-width: 2.5em;
  min-height: 44px;
  font-size: 1.1rem;
  font-weight: 600;
  padding: 6px 4px;
  box-sizing: border-box;
}
.spell-autopick-active .spell-answer-char-btn:disabled {
  opacity: 1;
  color: var(--text-primary);
}
.spell-letters-autopick {
  pointer-events: none;
}
.spell-letters-autopick .spell-letter-btn {
  opacity: 1;
  color: var(--text-primary);
}
/* Long word: smaller cap so more fit before shrinking */
.spell-long .spell-answer-letters {
  min-height: 52px;
  gap: 1px;
  padding: 6px 8px;
}
.spell-long .spell-answer-char-btn {
  min-height: 36px;
  max-width: 2.2em;
  font-size: 0.95rem;
  padding: 4px 2px;
}
.spell-answer-placeholder {
  color: var(--text-tertiary, #999);
  font-size: 0.95rem;
}
.spell-letters {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
  justify-content: center;
  max-width: 100%;
}
.spell-letter-btn {
  min-width: 44px;
  min-height: 44px;
  font-size: 1.1rem;
  font-weight: 600;
  padding: 8px 12px;
}
.spell-letter-btn.spell-letter-used {
  visibility: hidden;
}
.spell-long .spell-letter-btn {
  min-width: 34px;
  min-height: 34px;
  font-size: 0.95rem;
  padding: 5px 8px;
}
.spell-skip {
  margin-top: 8px;
}
.spell-actions-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.spell-actions-row .voice-mic {
  margin-top: 8px;
  width: 36px;
  height: 36px;
}
.spell-hint-button-wrapper {
  margin-top: 10px;
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
.type-input-prefix {
  flex-shrink: 0;
  padding-left: 14px;
  font-size: 1.1rem;
  color: var(--text-secondary, #666);
  user-select: none;
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
.type-input-inline:not(:has(.type-input-prefix)) .type-input {
  padding-left: 16px;
}
.type-input:focus {
  outline: none;
}
.type-input::placeholder {
  color: var(--text-tertiary, #999);
}
.type-reveal-text {
  flex: 1;
  min-width: 0;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  color: var(--text-primary);
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
  background: var(--bg-secondary, rgba(0,0,0,0.05));
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
.type-hint-button-wrapper {
  margin-top: 16px;
  text-align: center;
  opacity: 0;
  transform: translateY(10px);
  transition: opacity 0.4s ease, transform 0.4s ease;
}
.type-hint-button-wrapper.type-hint-button-visible {
  opacity: 1;
  transform: translateY(0);
}
.btn-type-hint-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.btn-type-hint-icon:hover {
  opacity: 0.8;
  transform: scale(1.1);
}
.btn-type-hint-icon:active {
  transform: scale(0.95);
}
.type-hint-icon {
  width: 24px;
  height: 24px;
  color: var(--text-secondary);
  opacity: 0.7;
  transition: opacity 0.2s ease, color 0.2s ease;
}
.btn-type-hint-icon:hover .type-hint-icon {
  opacity: 1;
  color: var(--text-primary);
}
[data-theme="dark"] .type-hint-icon {
  color: var(--text-secondary);
  opacity: 0.6;
}
[data-theme="dark"] .btn-type-hint-icon:hover .type-hint-icon {
  color: var(--text-primary);
  opacity: 0.9;
}
.type-hint-text {
  margin-top: 10px;
  font-size: 1.1rem;
  letter-spacing: 0.15em;
  color: var(--text-secondary, #666);
  font-family: var(--font-mono, monospace);
}

.options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 20px 0;
}

@media (min-width: 768px) {
  .options {
    grid-template-columns: repeat(2, minmax(0, 1fr));
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
  min-width: 0;
  white-space: normal;
  overflow-wrap: anywhere;
  word-break: break-word;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
}

[data-theme="dark"] .option-btn {
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
  min-width: 0;
  text-align: left;
  overflow-wrap: anywhere;
  word-break: break-word;
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
  animation: correct-success 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
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
  animation: correct-shine 0.8s ease-out;
  pointer-events: none;
}

.option-btn.option-incorrect {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: 1px solid #ef4444;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  animation: incorrect-fail 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
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
  animation: incorrect-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes incorrect-pulse {
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

@keyframes correct-success {
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

@keyframes correct-shine {
  0% {
    transform: translateX(-100%) translateY(-100%) rotate(45deg);
  }
  100% {
    transform: translateX(100%) translateY(100%) rotate(45deg);
  }
}

@keyframes incorrect-fail {
  0%, 100% {
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
  animation: feedback-success-appear 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feedback-success::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  animation: feedback-success-glow 1.5s ease-out;
  pointer-events: none;
}

.feedback-error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  animation: feedback-error-appear 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
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
  animation: feedback-error-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes feedback-error-pulse {
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

@keyframes feedback-success-appear {
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

@keyframes feedback-success-glow {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

@keyframes feedback-error-appear {
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
  animation: feedback-icon-spin 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feedback-error .feedback-icon {
  animation: feedback-icon-shake 0.5s ease-out;
}

@keyframes feedback-icon-spin {
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

@keyframes feedback-icon-shake {
  0%, 100% {
    transform: translateX(0) rotate(0deg);
  }
  25% {
    transform: translateX(-8px) rotate(-10deg);
  }
  75% {
    transform: translateX(8px) rotate(10deg);
  }
}

/* Success particles animation */
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
  animation: success-particle-fly 1s ease-out var(--particle-delay, 0s) forwards;
}

@keyframes success-particle-fly {
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
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
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
  animation: error-progress-pulse 1s linear 0.5s forwards;
  pointer-events: none;
}

@keyframes error-progress-pulse {
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

.spell-reveal-row {
  margin: 16px 0 8px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}
.spell-reveal-row .spell-answer-label {
  margin: 0;
}
.spell-reveal-letters {
  min-height: 44px;
}
.spell-reorder-group {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex-wrap: nowrap;
}
.spell-reveal-char {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  font-size: 1.1rem;
  font-weight: 600;
  padding: 6px 4px;
  background: var(--bg-secondary, rgba(0,0,0,0.08));
  border-radius: 8px;
  border: 1px solid var(--border-primary, rgba(0,0,0,0.12));
  color: var(--text-primary);
  box-sizing: border-box;
}
.spell-reorder-move {
  transition: transform 0.35s ease;
}
.spell-long .spell-reveal-char {
  min-width: 34px;
  min-height: 36px;
  font-size: 0.95rem;
  padding: 4px 2px;
}

.hint {
  margin: 20px 0 10px 0;
  padding: 12px 18px;
  background: var(--hint-bg, rgba(245, 158, 11, 0.1));
  border-radius: 8px;
  color: var(--text-primary);
  border-left: 4px solid #f59e0b;
  text-align: left;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
  font-size: 14px;
  line-height: 1.5;
}

.example {
  font-style: italic;
  margin: 10px 0 20px 0;
  padding: 15px 20px;
  background: var(--example-bg, rgba(59, 130, 246, 0.1));
  border-radius: 8px;
  color: var(--text-primary);
  border-left: 4px solid var(--color-primary);
  text-align: left;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.waiting-progress {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
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

.training-idle-stack {
  display: flex;
  flex-direction: column;
  gap: 22px;
  width: 100%;
  max-width: 720px;
  margin: 0 auto;
}

.training-verb-forms-cta {
  width: 100%;
  max-width: 720px;
  margin: 0 auto;
  padding: 18px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.training-verb-forms-cta--compact {
  margin-top: 16px;
  padding: 12px 16px;
}

.training-verb-forms-cta__btn {
  text-decoration: none;
  font-weight: 600;
  min-width: min(100%, 280px);
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.35);
}

.training-verb-forms-cta__btn:hover {
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.45);
}

.training-verb-forms-cta__title {
  margin: 0 0 8px;
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text-primary);
  text-align: center;
  width: 100%;
}

.training-verb-forms-cta__text {
  margin: 0 0 14px;
  font-size: 0.9rem;
  line-height: 1.45;
  color: var(--text-secondary);
  text-align: center;
  max-width: 42rem;
}

.training-verb-forms-cta__count {
  margin: 0 0 12px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
}

.training-verb-forms-cta__count--compact {
  margin: 0 0 10px;
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-secondary);
}

.start-screen {
  text-align: center;
  padding: 40px 20px;
}

.start-screen-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 28px;
  max-width: 600px;
  margin: 0 auto;
}

.start-screen-stats {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.start-stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 24px;
  background: linear-gradient(135deg, var(--bg-secondary, rgba(0, 0, 0, 0.05)) 0%, var(--bg-secondary, rgba(0, 0, 0, 0.08)) 100%);
  border-radius: 16px;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.start-stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

.start-stat-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
  opacity: 0.8;
}

.start-stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--color-primary);
  display: inline-block;
  line-height: 1.2;
}

.start-stat-value span:last-child {
  font-size: 18px;
  font-weight: 500;
  color: var(--text-secondary);
  opacity: 0.7;
  margin-left: 6px;
}

.upcoming-cards-chart {
  width: 100%;
  margin-top: 28px;
  padding: 24px;
  background: linear-gradient(135deg, var(--bg-secondary, rgba(0, 0, 0, 0.05)) 0%, var(--bg-secondary, rgba(0, 0, 0, 0.08)) 100%);
  border-radius: 16px;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.upcoming-cards-chart:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

.chart-header {
  margin-bottom: 20px;
  text-align: center;
}

.chart-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 6px 0;
  letter-spacing: -0.3px;
}

.chart-subtitle {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-secondary);
  opacity: 0.7;
  margin: 0;
}

.chart-container {
  position: relative;
  height: 220px;
  width: 100%;
  margin-top: 8px;
}

.btn-start {
  width: 100%;
  max-width: 320px;
  padding: 16px 32px;
  font-size: 17px;
  font-weight: 600;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  letter-spacing: 0.3px;
}

.btn-start:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}

.btn-start:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
}

.stat-item:last-child {
  border-bottom: none;
}

.stat-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--color-primary);
}

.no-cards-message {
  margin-top: 15px;
  color: var(--text-secondary);
  font-style: italic;
}

.completion-screen {
  text-align: center;
  position: relative;
  overflow: hidden;
  padding: 40px 20px;
}

.completion-percentage {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
}

.percentage-circle-wrapper {
  position: relative;
  width: 200px;
  height: 200px;
  flex-shrink: 0;
}

.percentage-circle {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.percentage-circle-bg {
  opacity: 0.2;
}

.percentage-circle-fill {
  transition: stroke 0.3s ease;
  /* No transition for stroke-dashoffset - animated via JavaScript for perfect sync */
}

.percentage-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.percentage-number {
  font-size: 48px;
  font-weight: 700;
  color: var(--text-primary);
  transition: color 0.3s ease;
  line-height: 1;
}

.percentage-ratio {
  font-size: 20px;
  font-weight: 500;
  color: var(--text-secondary);
  opacity: 0.85;
  line-height: 1;
}

.motivational-message {
  padding: 18px 28px;
  border-radius: 12px;
  max-width: 600px;
  width: 100%;
  animation: messageAppear 0.6s ease-out 0.3s both;
  text-align: center;
}

.message-text {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  line-height: 1.5;
}

.message-excellent {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(5, 150, 105, 0.15) 100%);
  color: var(--color-success, #10b981);
  border: 2px solid rgba(16, 185, 129, 0.3);
}

.message-good {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(37, 99, 235, 0.15) 100%);
  color: #3b82f6;
  border: 2px solid rgba(59, 130, 246, 0.3);
}

.message-okay {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(217, 119, 6, 0.15) 100%);
  color: #f59e0b;
  border: 2px solid rgba(245, 158, 11, 0.3);
}

.message-needs-improvement {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(220, 38, 38, 0.15) 100%);
  color: #ef4444;
  border: 2px solid rgba(239, 68, 68, 0.3);
}

@keyframes messageAppear {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Celebration animations */
.celebration-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  overflow: visible;
  z-index: 2;
  width: 0;
  height: 0;
}

.fireworks {
  position: absolute;
  width: 100%;
  height: 100%;
}

.firework {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  top: 0;
  left: 0;
  transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(var(--firework-size, 1));
  animation: firework-explode 2s ease-out var(--delay, 0s) forwards;
}

.firework-core {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: radial-gradient(circle, #fff 0%, #ffd700 50%, transparent 100%);
  box-shadow: 0 0 10px rgba(255, 215, 0, 0.8), 0 0 20px rgba(255, 215, 0, 0.5);
  animation: firework-core-pulse 0.3s ease-out var(--delay, 0s) forwards;
  transform: translate(-50%, -50%);
  top: 50%;
  left: 50%;
}

.firework-particle {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--particle-color, #10b981);
  top: 0;
  left: 0;
  transform: translate(-50%, -50%);
  box-shadow: 0 0 6px var(--particle-color, #10b981);
  animation: particle-fly 2s ease-out var(--delay, 0s) forwards;
}

.firework:nth-child(3n+1) .firework-particle {
  --particle-color: #3b82f6;
}

.firework:nth-child(3n+2) .firework-particle {
  --particle-color: #f59e0b;
}

.firework:nth-child(3n+3) .firework-particle {
  --particle-color: #ec4899;
}

.firework:nth-child(4n+1) .firework-particle {
  --particle-color: #10b981;
}

.firework:nth-child(4n+2) .firework-particle {
  --particle-color: #8b5cf6;
}

@keyframes firework-explode {
  0% {
    opacity: 1;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(var(--firework-size, 1));
  }
  15% {
    opacity: 1;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(calc(var(--firework-size, 1) * 1.5));
  }
  100% {
    opacity: 0;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(0);
  }
}

@keyframes firework-core-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 1;
  }
  50% {
    transform: translate(-50%, -50%) scale(2);
    opacity: 0.8;
  }
  100% {
    transform: translate(-50%, -50%) scale(3);
    opacity: 0;
  }
}

@keyframes particle-fly {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1) rotate(0deg);
  }
  50% {
    opacity: 1;
    transform: translate(-50%, -50%) translateX(calc(var(--offset-x, 0) * 0.5)) translateY(calc(var(--offset-y, 0) * 0.5)) scale(1.2) rotate(180deg);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) translateX(var(--offset-x, 0)) translateY(var(--offset-y, 0)) scale(0) rotate(360deg);
  }
}

/* Calculate offsets for 12 particles in a circle (30 degrees each) */
.firework-particle[data-particle-index="1"] { --offset-x: 100px; --offset-y: 0px; }
.firework-particle[data-particle-index="2"] { --offset-x: 86.6px; --offset-y: -50px; }
.firework-particle[data-particle-index="3"] { --offset-x: 50px; --offset-y: -86.6px; }
.firework-particle[data-particle-index="4"] { --offset-x: 0px; --offset-y: -100px; }
.firework-particle[data-particle-index="5"] { --offset-x: -50px; --offset-y: -86.6px; }
.firework-particle[data-particle-index="6"] { --offset-x: -86.6px; --offset-y: -50px; }
.firework-particle[data-particle-index="7"] { --offset-x: -100px; --offset-y: 0px; }
.firework-particle[data-particle-index="8"] { --offset-x: -86.6px; --offset-y: 50px; }
.firework-particle[data-particle-index="9"] { --offset-x: -50px; --offset-y: 86.6px; }
.firework-particle[data-particle-index="10"] { --offset-x: 0px; --offset-y: 100px; }
.firework-particle[data-particle-index="11"] { --offset-x: 50px; --offset-y: 86.6px; }
.firework-particle[data-particle-index="12"] { --offset-x: 86.6px; --offset-y: 50px; }

.confetti {
  position: absolute;
  width: 0;
  height: 0;
  top: 0;
  left: 0;
}

.confetti-piece {
  position: absolute;
  width: 12px;
  height: 12px;
  background: var(--confetti-color, #10b981);
  top: 0;
  left: 0;
  animation: confetti-fly var(--confetti-duration, 3s) 0s ease-out forwards;
  border-radius: 2px;
  box-shadow: 0 0 4px var(--confetti-color, #10b981);
}

.confetti-piece:nth-child(3n+1) {
  --confetti-color: #10b981;
  clip-path: polygon(50% 0%, 0% 100%, 100% 100%);
}

.confetti-piece:nth-child(3n+2) {
  --confetti-color: #3b82f6;
  border-radius: 50%;
}

.confetti-piece:nth-child(3n+3) {
  --confetti-color: #f59e0b;
  clip-path: polygon(25% 0%, 100% 0%, 75% 100%, 0% 100%);
}

.confetti-piece:nth-child(4n+1) {
  --confetti-color: #ec4899;
}

.confetti-piece:nth-child(4n+2) {
  --confetti-color: #8b5cf6;
}

.confetti-piece:nth-child(4n+3) {
  --confetti-color: #06b6d4;
}

@keyframes confetti-fly {
  0% {
    transform: translate(var(--confetti-start-x, 0), var(--confetti-start-y, 0)) rotate(0deg) scale(1);
    opacity: 1;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translate(var(--confetti-end-x, 0), var(--confetti-end-y, 0)) rotate(1080deg) scale(0);
    opacity: 0;
  }
}

/* Failure animation for <10% */
.failure-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  overflow: visible;
  z-index: 2;
  width: 0;
  height: 0;
}

.failure-rain {
  position: absolute;
  width: 0;
  height: 0;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.failure-item {
  position: absolute;
  top: 50%;
  left: 50%;
  font-size: 48px;
  line-height: 1;
  animation: failure-fall var(--duration, 2s) var(--delay, 0s) ease-in forwards;
  transform-origin: center;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
  opacity: 0; /* Start completely invisible */
  visibility: hidden; /* Hidden until animation starts */
}

.failure-emoji {
  display: block;
  filter: grayscale(0.4) brightness(0.8);
  user-select: none;
}

@keyframes failure-fall {
  0% {
    opacity: 0;
    visibility: hidden;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0);
  }
  3% {
    opacity: 0;
    visibility: visible;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0);
  }
  6% {
    opacity: 1;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0.5);
  }
  90% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) translate(var(--end-x, 0), var(--end-y, 0)) rotate(var(--rotation, 0deg)) scale(var(--scale, 1));
  }
}


.completion-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  width: 100%;
  max-width: 400px;
}

.remaining-cards-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 24px;
  background: var(--bg-secondary, rgba(0, 0, 0, 0.05));
  border-radius: 10px;
  width: 100%;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
}

.remaining-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-primary);
  text-align: center;
  line-height: 1.4;
}

.remaining-text .remaining-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-right: 6px;
}

.btn-continue {
  width: 100%;
  max-width: 300px;
  padding: 14px 28px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 10px;
}


.network-error-notification {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  padding: 16px 24px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
  animation: slideDown 0.3s ease-out;
  max-width: 90%;
  width: auto;
  min-width: 300px;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.network-error-content {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.network-error-close {
  margin-left: auto;
  border: 0;
  border-radius: 50%;
  width: 26px;
  height: 26px;
  background: rgba(255, 255, 255, 0.22);
  color: #fff;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
}

.network-error-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.network-error-text {
  flex: 1;
}

.network-error-title {
  font-weight: 600;
  font-size: 16px;
  margin-bottom: 4px;
}

.network-error-message {
  font-size: 14px;
  opacity: 0.95;
}

/* Example button styles */
.example-button-wrapper {
  margin-top: 16px;
  text-align: center;
  opacity: 0;
  transform: translateY(10px);
  transition: opacity 0.4s ease, transform 0.4s ease;
}

.example-button-wrapper.example-button-visible {
  opacity: 1;
  transform: translateY(0);
}

.btn-example-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  border: none;
  background: none;
  cursor: pointer;
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.btn-example-icon:hover {
  opacity: 0.8;
  transform: scale(1.1);
}

.btn-example-icon:active {
  transform: scale(0.95);
}

.example-icon {
  width: 24px;
  height: 24px;
  color: var(--text-secondary);
  opacity: 0.7;
  transition: opacity 0.2s ease, color 0.2s ease;
}

.btn-example-icon:hover .example-icon {
  opacity: 1;
  color: var(--text-primary);
}

[data-theme="dark"] .example-icon {
  color: var(--text-secondary);
  opacity: 0.6;
}

[data-theme="dark"] .btn-example-icon:hover .example-icon {
  color: var(--text-primary);
  opacity: 0.9;
}

.example-usage {
  margin-top: 0;
}

/* Hide option number on mobile */
@media (max-width: 768px) {
  .spell-answer-prefix {
    font-size: 1rem;
  }

  .option-number {
    display: none;
  }
  
  /* Compact completion screen on mobile */
  .completion-screen {
    padding: 30px 8px;
  }
  
  .completion-percentage {
    gap: 20px;
  }
  
  .percentage-circle-wrapper {
    width: 180px;
    height: 180px;
  }
  
  .percentage-number {
    font-size: 42px;
  }
  
  .percentage-ratio {
    font-size: 18px;
  }
  
  .motivational-message {
    padding: 14px 8px;
    font-size: 16px;
  }
  
  .completion-actions {
    max-width: 100%;
  }
  
  .remaining-cards-info {
    padding: 14px 8px;
  }
  
  .remaining-text {
    font-size: 16px;
  }
  
  .btn-continue {
    max-width: 100%;
  }
  
  .start-screen {
    padding: 30px 12px;
  }
  
  .start-screen-content {
    gap: 24px;
  }
  
  .start-stat-item {
    padding: 20px;
    border-radius: 14px;
  }
  
  .start-stat-value {
    font-size: 28px;
  }
  
  .start-stat-value span:last-child {
    font-size: 16px;
  }
  
  .upcoming-cards-chart {
    padding: 20px;
    border-radius: 14px;
    margin-top: 24px;
  }
  
  .chart-title {
    font-size: 16px;
  }
  
  .chart-subtitle {
    font-size: 12px;
  }
  
  .chart-container {
    height: 200px;
  }
  
  .btn-start {
    max-width: 100%;
    padding: 14px 28px;
    font-size: 16px;
  }
}
</style>
