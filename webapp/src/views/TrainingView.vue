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
import DOMPurify from 'dompurify'
import type { Card, OptionsResponse, Feedback } from '../api/trainingTypes'
import { ref, computed, onMounted, onUnmounted, watch, nextTick, TransitionGroup } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { contentReportClient } from '../api/contentReportClient'
import { wordTrainingClient } from '../api/wordTrainingClient'
import { showAlert } from '../composables/useDialog'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'
import { useTrainingUpcoming } from '../composables/useTrainingUpcoming'
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
const { upcomingCardsLoaded, loadUpcomingCards } = useTrainingUpcoming(upcomingChartCanvas)
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
let revealingGeneration: number | null = null

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
  
  return DOMPurify.sanitize(question, { USE_PROFILES: { html: true } })
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

onUnmounted(() => {
  sessionActive.value = false
  currentCardGeneration++
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
      await applyTrainingSessionResponse(response)
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
  const card = currentCard.value
  prefetchInFlight = (async () => {
    try {
      const response = await wordTrainingClient.prefetchNext()
      if (!sessionActive.value || generation !== currentCardGeneration) return
      if (!response || response.complete || response.active === false || !response.card_index || !response.total_cards) return
      // The server may have advanced while this request was in flight.
      if (!card || response.session_id !== card.session_id || response.card_index !== card.card_index + 1) return
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

const revealOptions = async (_isEarly: boolean = false) => {
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
  if (!sessionActive.value || !currentCard.value || optionsShown.value || revealingGeneration === currentCardGeneration) {
    return
  }

  const card = currentCard.value
  const generation = currentCardGeneration
  revealingGeneration = generation
  const isCurrent = () => sessionActive.value && generation === currentCardGeneration && sameTrainingCard(card, currentCard.value)
  try {
    if (syncCurrentInFlight) {
      await syncCurrentInFlight
    }
    if (!isCurrent()) return
    const data: OptionsResponse = await wordTrainingClient.reveal()
    if (!isCurrent()) return
    if (data.user_card_id !== card.user_card_id ||
        (data.session_id != null && data.session_id !== card.session_id) ||
        (data.card_index != null && data.card_index !== card.card_index)) {
      // Recover from a session that advanced elsewhere without ever displaying
      // another card's options alongside this question.
      const response = await wordTrainingClient.current()
      if (isCurrent()) await applyTrainingSessionResponse(response)
      return
    }
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
    if (!isCurrent()) return
    console.error('Failed to reveal options:', error)
    // Network error is already handled by callback, but we should handle other errors
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert(t('training.failedLoadOptions'))
    }
  } finally {
    if (revealingGeneration === generation) revealingGeneration = null
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
    formData.append('session_id', String(cardAtSubmit?.session_id || 0))
    formData.append('card_index', String(cardAtSubmit?.card_index || 0))
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
    formData.append('session_id', String(cardAtSubmit?.session_id || 0))
    formData.append('card_index', String(cardAtSubmit?.card_index || 0))
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
    formData.append('session_id', String(cardAtSubmit?.session_id || 0))
    formData.append('card_index', String(cardAtSubmit?.card_index || 0))
    
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
    if (cached && currentCard.value && cached.session_id === currentCard.value.session_id &&
        cached.card_index === currentCard.value.card_index + 1) {
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

<style scoped src="../styles/training.css"></style>
