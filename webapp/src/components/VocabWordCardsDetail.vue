<template>
  <div class="vocab-detail-root">
    <div class="modal-header">
      <div class="word-header-info">
        <div class="word-title-row">
          <h3>{{ selectedWordDisplay }}</h3>
          <div v-if="selectedTranscription" class="transcription-with-audio">
            <div class="transcription">{{ selectedTranscription }}</div>
            <button
              v-if="selectedPronunciationURL"
              class="btn-pronunciation"
              :disabled="playingPronunciation"
              type="button"
              :title="t('wordSets.pronounceAria')"
              :aria-label="t('wordSets.pronounceAria')"
              @click="playSelectedPronunciation"
            >
              <Icon name="play" />
            </button>
          </div>
        </div>
        <div class="word-summary">
          <span v-if="selectedMorphText">{{ selectedMorphText }}</span>
          <span v-if="listMasteringScore !== null && listMasteringScore !== undefined" class="mastering-score-inline" :title="t('vocab.scoreLabel') + ' 0–100'">
            <span class="mastery-dot-inline" :style="{ backgroundColor: masteryColor(listMasteringScore) }" />
            {{ listMasteringScore }}
          </span>
          <span>{{ t('vocab.cards', totalCards, { n: totalCards }) }}</span>
          <span v-if="totalDue > 0">{{ t('vocab.due', totalDue, { n: totalDue }) }}</span>
          <span v-if="lastReview" :title="formatDateAbsolute(lastReview)">{{ t('vocab.last') }} {{ formatDateRelative(lastReview) }}</span>
        </div>
      </div>
      <button type="button" class="btn-close" @click="onCloseClick">&times;</button>
    </div>

    <div v-if="cardsLoading" class="loading">{{ t('vocab.loadingCards') }}</div>
    <div v-else-if="cards.length === 0" class="no-cards">{{ t('vocab.noCardsFound') }}</div>
    <div v-else>
      <div class="cards-list-simple">
        <div v-for="senseGroup in groupedCards" :key="senseGroup.sense_index" class="sense-group-simple">
          <div class="sense-header-simple">
            <h4>
              <span class="word-ru">{{ senseGroup.word_native || senseGroup.word_ru }}</span>
              <span v-if="senseGroup.pos" class="pos-badge">{{ senseGroup.pos }}</span>
            </h4>
          </div>
          <div class="sense-info-simple">
            <div v-if="senseGroup.meaning_target || senseGroup.meaning_en" class="meaning">
              {{ senseGroup.meaning_target || senseGroup.meaning_en }}
            </div>
            <div v-if="senseGroup.example_target || senseGroup.example_en" class="example">
              <strong>{{ t('vocab.example') }}:</strong> {{ senseGroup.example_target || senseGroup.example_en }}
            </div>
          </div>
          <div class="directions-simple">
            <div v-for="directionCard in senseGroup.directions" :key="directionCard.direction" class="direction-item-simple">
              <div class="direction-header-simple">
                <span class="direction-badge" :class="`direction-${directionCard.direction}`">
                  {{ directionCard.direction === 'ru_en' ? 'RU→EN' : 'EN→RU' }}
                </span>
                <span :class="['state-badge', `state-${directionCard.state}`]">{{ directionCard.state }}</span>
                <span
                  class="srs-info-wrap"
                  @click.prevent
                  @mouseenter="(e: MouseEvent) => showSrsTooltip(e, directionCard)"
                  @mouseleave="hideSrsTooltip"
                >
                  <Icon name="info" class="srs-info-icon" />
                </span>
              </div>
              <div class="direction-stats-simple">
                <span v-if="directionCard.reps > 0" :title="t('vocab.reps')">{{ t('vocab.reps') }} {{ directionCard.reps }}</span>
                <span v-else-if="directionCard.review_count > 0" :title="t('vocab.reviews')">{{ t('vocab.reviews') }} {{ directionCard.review_count }}</span>
                <span v-if="directionCard.next_due_at" :title="`${t('vocab.due')}: ${formatDateAbsolute(directionCard.next_due_at)}`">{{ t('vocab.due') }}: {{ formatDateRelative(directionCard.next_due_at) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="isVerbLikePOS && verbForms" class="verb-forms-section">
        <h4>{{ t('vocab.verbForms') }}</h4>
        <div class="verb-forms-list">
          <div v-if="verbForms.v1" class="verb-form-item">
            <span class="verb-form-label">{{ t('vocab.v1Base') }}</span>
            <span class="verb-form-value">{{ verbForms.v1 }}</span>
          </div>
          <div v-if="verbForms.v2" class="verb-form-item">
            <span class="verb-form-label">{{ t('vocab.v2PastSimple') }}</span>
            <span class="verb-form-value">{{ verbForms.v2 }}</span>
          </div>
          <div v-if="verbForms.v3" class="verb-form-item">
            <span class="verb-form-label">{{ t('vocab.v3PastParticiple') }}</span>
            <span class="verb-form-value">{{ verbForms.v3 }}</span>
          </div>
          <div v-if="verbForms.gerund" class="verb-form-item">
            <span class="verb-form-label">{{ t('vocab.gerund') }}</span>
            <span class="verb-form-value">{{ verbForms.gerund }}</span>
          </div>
          <div v-if="verbForms.third_person" class="verb-form-item">
            <span class="verb-form-label">{{ t('vocab.thirdPerson') }}</span>
            <span class="verb-form-value">{{ verbForms.third_person }}</span>
          </div>
        </div>
      </div>
      <div v-if="isVerbLikePOS && fullVerbForms.length > 0" class="verb-forms-section">
        <h4>{{ t('vocab.allVerbForms') }}</h4>
        <div class="verb-forms-table-wrap">
          <table class="verb-forms-table">
            <thead>
              <tr>
                <th>{{ t('vocab.verbFormTense') }}</th>
                <th>{{ t('vocab.verbFormMood') }}</th>
                <th>{{ t('vocab.verbFormSubject') }}</th>
                <th>{{ t('vocab.verbFormSurface') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in fullVerbForms" :key="`${item.tense}-${item.mood}-${item.person}-${item.number}-${item.surface_form}`">
                <td>{{ item.tense }}</td>
                <td>{{ item.mood }}</td>
                <td class="verb-forms-pronoun">{{ spanishVerbSubjectPronoun(item.person, item.number) }}</td>
                <td>{{ item.surface_form }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="modal-footer">
      <div class="footer-actions">
        <button v-if="hasUserCards" type="button" class="btn btn-primary" :disabled="processingAction" @click="markKnown">
          {{ t('vocab.moveToKnown') }}
        </button>
        <button v-if="!hasUserCards && isKnown" type="button" class="btn btn-primary" :disabled="processingAction" @click="moveToTraining">
          {{ t('vocab.moveToTraining') }}
        </button>
        <button type="button" class="btn btn-danger" :disabled="processingAction" @click="confirmDelete">
          {{ t('vocab.removeFromVocabulary') }}
        </button>
      </div>
    </div>

    <Teleport to="body">
      <div
        v-if="srsTooltipCard"
        class="srs-tooltip srs-tooltip-fixed"
        :style="srsTooltipStyle"
        @mouseenter="keepSrsTooltip"
        @mouseleave="hideSrsTooltip(true)"
      >
        <div class="srs-tooltip-title">{{ t('vocab.srsTooltipTitle') }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsState') }}:</span> {{ srsTooltipCard.state }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsEf') }}:</span> {{ formatSrsNumber(srsTooltipCard.ef) }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsReps') }}:</span> {{ srsTooltipCard.reps }} <span v-if="srsTooltipCard.state === 'learning'" class="srs-tooltip-hint" :title="t('vocab.srsRepsNote')">(?)</span></div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsIntervalDays') }}:</span> {{ srsTooltipCard.interval_days }}<template v-if="srsTooltipCard.state === 'learning'"> → <span class="srs-tooltip-step">{{ t('vocab.srsStepInterval') }}: {{ getStepIntervalDays(srsTooltipCard.direction, srsTooltipCard.learning_step) }} {{ getStepIntervalDays(srsTooltipCard.direction, srsTooltipCard.learning_step) === 1 ? t('vocab.srsDay') : t('vocab.srsDays') }}</span></template></div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsLearningStep') }}:</span> {{ srsTooltipCard.learning_step }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsLapseCount') }}:</span> {{ srsTooltipCard.lapse_count }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsNextDueAt') }}:</span> {{ formatDateAbsolute(srsTooltipCard.next_due_at) }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsLastReviewAt') }}:</span> {{ formatDateAbsolute(srsTooltipCard.last_review_at) }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsLastQuality') }}:</span> {{ srsTooltipCard.last_quality != null ? srsTooltipCard.last_quality : '—' }}</div>
        <div v-if="srsTooltipCard.last_quality === 0 || srsTooltipCard.last_quality === 1" class="srs-tooltip-reason">{{ t('vocab.srsQualityHardReason') }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsCreatedAt') }}:</span> {{ formatDateAbsolute(srsTooltipCard.created_at ?? null) }}</div>
        <div class="srs-tooltip-row"><span>{{ t('vocab.srsUpdatedAt') }}:</span> {{ formatDateAbsolute(srsTooltipCard.updated_at ?? null) }}</div>
      </div>
    </Teleport>

    <div v-if="showDeleteConfirm" class="modal modal-nested" @click.self="showDeleteConfirm = false">
      <div class="modal-content modal-small">
        <h3>{{ t('vocab.removeFromVocabularyTitle') }}</h3>
        <p>{{ t('vocab.removeConfirm', { word: wordToDelete }) }}</p>
        <p class="warning-text">{{ t('vocab.removeWarning') }}</p>
        <div class="modal-actions">
          <button type="button" class="btn btn-danger" @click="deleteWord">{{ t('vocab.remove') }}</button>
          <button type="button" class="btn btn-secondary" @click="showDeleteConfirm = false">{{ t('common.cancel') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { withCourseCode } from '../api/grammarClient'
import { spanishVerbSubjectPronoun } from '../utils/spanishVerbPronouns'
import { showAlert } from '../composables/useDialog'
import { useAudio } from '../composables/useAudio'
import Icon from './Icon.vue'

export interface MorphVerbForms {
  v1?: string
  v2?: string
  v3?: string
}

export interface MorphInfo {
  pos?: string
  noun_gender?: string
  article?: string
  opposite_gender_word?: string
  verb_forms?: MorphVerbForms
}

export interface CardDetail {
  id: number
  training_card_id: number
  direction: string
  state: string
  ef: number
  reps: number
  interval_days: number
  learning_step: number
  lapse_count: number
  next_due_at: string | null
  last_review_at: string | null
  last_quality: number | null
  created_at?: string
  updated_at?: string
  word_ru: string
  word_native?: string
  meaning_en: string
  meaning_target?: string
  example_en: string
  example_target?: string
  example_ru: string
  example_native?: string
  transcription: string
  sense_index: number
  pos?: string
  review_count: number
}

export interface VerbFormRow {
  word_card_id: number
  lemma: string
  mood: string
  tense: string
  person: string
  number: string
  surface_form: string
  is_irregular: boolean
}

export interface VocabCardsAPIResponse {
  lemma: string
  word_card_id: number
  cards: CardDetail[]
  verb_forms?: Record<string, unknown>
  pos?: string
  noun_gender?: string
  opposite_gender_word?: string
  morph?: MorphInfo
  has_user_cards?: boolean
  is_known?: boolean
}

interface SenseGroup {
  sense_index: number
  word_ru: string
  word_native?: string
  meaning_en: string
  meaning_target?: string
  example_en: string
  example_target?: string
  example_ru: string
  example_native?: string
  transcription: string
  pos?: string
  directions: CardDetail[]
}

const props = defineProps<{
  lemma: string
  /** From vocabulary list (mastering dot). Omit in reading. */
  listMasteringScore?: number | null
  /** When set (e.g. from `/api/reading/word-lookup`), skip duplicate `/api/vocab/.../cards` fetch. */
  preloaded?: VocabCardsAPIResponse | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'vocabChanged'): void
}>()

const { t } = useI18n()
const { getWordPronunciationURL, playWordPronunciation } = useAudio()

const cleanLemma = (lemma: string): string => lemma.replace(/^to\s+/i, '')

const cardsLoading = ref(true)
const cards = ref<CardDetail[]>([])
const verbForms = ref<Record<string, unknown> | null>(null)
const fullVerbForms = ref<VerbFormRow[]>([])
const wordPOS = ref<string | null>(null)
const selectedWordCardID = ref<number | null>(null)
const nounGender = ref<string | null>(null)
const wordMorph = ref<MorphInfo | null>(null)
const hasUserCards = ref(false)
const isKnown = ref(false)
const processingAction = ref(false)

const selectedWordDisplay = ref('')
const selectedTranscription = ref('')
const selectedPronunciationWord = ref('')
const selectedPronunciationURL = ref<string | null>(null)
const playingPronunciation = ref(false)

const showDeleteConfirm = ref(false)
const wordToDelete = ref('')

const isVerbLikePOS = computed(() => {
  const p = (wordPOS.value || '').toLowerCase()
  return p === 'verb' || p === 'aux'
})

const srsTooltipCard = ref<CardDetail | null>(null)
const srsTooltipPosition = ref<{ top: number; left: number } | null>(null)
let srsTooltipHideTimeout: ReturnType<typeof setTimeout> | null = null

const srsTooltipStyle = computed(() => {
  const pos = srsTooltipPosition.value
  if (!pos) return {}
  return { top: `${pos.top}px`, left: `${pos.left}px` }
})

function showSrsTooltip(e: MouseEvent, card: CardDetail) {
  if (srsTooltipHideTimeout) {
    clearTimeout(srsTooltipHideTimeout)
    srsTooltipHideTimeout = null
  }
  const el = e.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  srsTooltipCard.value = card
  srsTooltipPosition.value = { top: rect.top - 8, left: rect.left + rect.width / 2 }
}

function keepSrsTooltip() {
  if (srsTooltipHideTimeout) {
    clearTimeout(srsTooltipHideTimeout)
    srsTooltipHideTimeout = null
  }
}

function hideSrsTooltip(immediate = false) {
  if (immediate) {
    if (srsTooltipHideTimeout) {
      clearTimeout(srsTooltipHideTimeout)
      srsTooltipHideTimeout = null
    }
    srsTooltipCard.value = null
    srsTooltipPosition.value = null
    return
  }
  srsTooltipHideTimeout = setTimeout(() => {
    srsTooltipHideTimeout = null
    srsTooltipCard.value = null
    srsTooltipPosition.value = null
  }, 150)
}

let loadGen = 0

function applyPayload(data: VocabCardsAPIResponse) {
  cards.value = data.cards || []
  verbForms.value = (data.verb_forms as Record<string, unknown>) || null
  wordPOS.value = data.pos || null
  selectedWordCardID.value = data.word_card_id || null
  nounGender.value = data.noun_gender || null
  wordMorph.value = data.morph || null
  hasUserCards.value = data.has_user_cards || false
  isKnown.value = data.is_known || false
  fullVerbForms.value = []
}

async function loadCards() {
  const gen = ++loadGen
  const lemma = props.lemma?.trim()
  if (!lemma) return

  cardsLoading.value = true
  cards.value = []
  selectedTranscription.value = ''
  selectedPronunciationWord.value = ''
  selectedPronunciationURL.value = null

  try {
    const pre = props.preloaded
    const preLemma = pre?.lemma?.trim().toLowerCase()
    const wantLemma = lemma.toLowerCase()
    if (pre && preLemma === wantLemma) {
      applyPayload(pre)
    } else {
      const data: VocabCardsAPIResponse = await apiClient.request(
        withCourseCode(`/api/vocab/${encodeURIComponent(lemma)}/cards`),
      )
      if (gen !== loadGen || props.lemma !== lemma) return
      applyPayload(data)
    }

    if (gen !== loadGen || props.lemma !== lemma) return

    if (isVerbLikePOS.value && selectedWordCardID.value) {
      try {
        const formsResp: { forms?: VerbFormRow[] } = await apiClient.request(`/api/vocab/${selectedWordCardID.value}/verb-forms`)
        if (gen === loadGen) fullVerbForms.value = formsResp.forms || []
      } catch (e) {
        console.warn('Failed to load full verb forms', e)
      }
    }

    selectedWordDisplay.value = cleanLemma(lemma)
    selectedPronunciationWord.value = lemma
    if (cards.value.length > 0) {
      selectedTranscription.value = cards.value[0].transcription || ''
      if (selectedTranscription.value) {
        const url = await getWordPronunciationURL(selectedPronunciationWord.value)
        if (gen !== loadGen || props.lemma !== lemma) return
        selectedPronunciationURL.value = url
      }
    }
  } catch (error) {
    console.error('Failed to load word cards detail:', error)
    if (gen === loadGen) await showAlert(t('vocab.alertLoadCardsFailed'))
  } finally {
    if (gen === loadGen) cardsLoading.value = false
  }
}

watch(
  () => [props.lemma, props.preloaded] as const,
  () => {
    hideSrsTooltip(true)
    loadCards()
  },
  { immediate: true },
)

const groupedCards = computed((): SenseGroup[] => {
  const groups = new Map<number, SenseGroup>()
  for (const card of cards.value) {
    if (!groups.has(card.sense_index)) {
      groups.set(card.sense_index, {
        sense_index: card.sense_index,
        word_ru: card.word_ru,
        word_native: card.word_native,
        meaning_en: card.meaning_en,
        meaning_target: card.meaning_target,
        example_en: card.example_en,
        example_target: card.example_target,
        example_ru: card.example_ru,
        example_native: card.example_native,
        transcription: card.transcription,
        pos: card.pos,
        directions: [],
      })
    }
    groups.get(card.sense_index)!.directions.push(card)
  }
  return Array.from(groups.values())
    .sort((a, b) => a.sense_index - b.sense_index)
    .map((group) => ({
      ...group,
      directions: group.directions.sort((a, b) => {
        if (a.direction === 'en_ru' && b.direction === 'ru_en') return -1
        if (a.direction === 'ru_en' && b.direction === 'en_ru') return 1
        return 0
      }),
    }))
})

const selectedMorphText = computed(() => {
  if (wordMorph.value) {
    const m = wordMorph.value
    if (m.pos === 'noun' && m.noun_gender) {
      const core = m.article ? `${m.article} • ${m.noun_gender}` : m.noun_gender
      return m.opposite_gender_word ? `${core} (${m.opposite_gender_word})` : core
    }
    if ((m.pos === 'verb' || m.pos === 'aux') && m.verb_forms) {
      const forms = [m.verb_forms.v1, m.verb_forms.v2, m.verb_forms.v3].filter(Boolean)
      if (forms.length > 0) return forms.join(', ')
    }
  }
  if (wordPOS.value === 'noun' && nounGender.value) {
    return nounGender.value
  }
  return ''
})

const totalCards = computed(() => cards.value.length)
const totalDue = computed(() => cards.value.filter((c) => c.next_due_at && new Date(c.next_due_at) <= new Date()).length)
const lastReview = computed(() => {
  const reviews = cards.value
    .map((c) => c.last_review_at)
    .filter((d): d is string => d !== null)
    .sort()
    .reverse()
  return reviews.length > 0 ? reviews[0] : null
})

const masteryColor = (score: number): string => {
  const s = Math.max(0, Math.min(100, score ?? 0))
  const hue = (120 * s) / 100
  return `hsl(${hue}, 72%, 42%)`
}

const formatDateAbsolute = (dateStr: string | null): string => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '—'
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  return `${day}.${month}.${year}`
}

const formatDateRelative = (dateStr: string | null): string => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '—'
  const now = new Date()
  const diffTime = now.getTime() - date.getTime()
  const diffDays = Math.floor(Math.abs(diffTime) / (1000 * 60 * 60 * 24))
  const diffHours = Math.floor(Math.abs(diffTime) / (1000 * 60 * 60))
  const diffMinutes = Math.floor(Math.abs(diffTime) / (1000 * 60))
  const isFuture = diffTime < 0
  if (diffDays === 0) {
    if (diffHours === 0) {
      if (diffMinutes < 1) return t('vocab.justNow')
      if (isFuture) return t('vocab.inMinutes', diffMinutes, { n: diffMinutes })
      return t('vocab.minutesAgo', diffMinutes, { n: diffMinutes })
    }
    if (isFuture) return t('vocab.inHours', diffHours, { n: diffHours })
    return t('vocab.hoursAgo', diffHours, { n: diffHours })
  }
  if (diffDays === 1) return isFuture ? t('vocab.tomorrow') : t('vocab.yesterday')
  if (diffDays < 7) return isFuture ? t('vocab.inDays', diffDays, { n: diffDays }) : t('vocab.daysAgo', diffDays, { n: diffDays })
  const diffWeeks = Math.floor(diffDays / 7)
  if (diffWeeks < 4) {
    return isFuture ? t('vocab.inWeeks', diffWeeks, { n: diffWeeks }) : t('vocab.weeksAgo', diffWeeks, { n: diffWeeks })
  }
  const diffMonths = Math.floor(diffDays / 30)
  if (diffMonths < 12) {
    return isFuture ? t('vocab.inMonths', diffMonths, { n: diffMonths }) : t('vocab.monthsAgo', diffMonths, { n: diffMonths })
  }
  const diffYears = Math.floor(diffDays / 365)
  return isFuture ? t('vocab.inYears', diffYears, { n: diffYears }) : t('vocab.yearsAgo', diffYears, { n: diffYears })
}

const formatSrsNumber = (n: number): string => {
  if (typeof n !== 'number' || Number.isNaN(n)) return '—'
  return Number.isInteger(n) ? String(n) : n.toFixed(2)
}

const LEARNING_STEPS_EN_RU = [1, 3, 7]
const LEARNING_STEPS_RU_EN = [1, 3, 7, 14]

function getStepIntervalDays(direction: string, learningStep: number): number {
  const steps = direction === 'ru_en' ? LEARNING_STEPS_RU_EN : LEARNING_STEPS_EN_RU
  if (learningStep < 0 || learningStep >= steps.length) return 0
  return steps[learningStep]
}

const playSelectedPronunciation = async () => {
  if (!selectedPronunciationWord.value || playingPronunciation.value) return
  playingPronunciation.value = true
  try {
    await playWordPronunciation(selectedPronunciationWord.value)
  } finally {
    playingPronunciation.value = false
  }
}

function onCloseClick() {
  hideSrsTooltip(true)
  emit('close')
}

const markKnown = async () => {
  if (processingAction.value || !props.lemma) return
  processingAction.value = true
  try {
    const formData = new FormData()
    await apiClient.requestFormData(withCourseCode(`/api/vocab/${encodeURIComponent(props.lemma)}/mark_known`), formData)
    emit('close')
    emit('vocabChanged')
  } catch (error) {
    console.error('Failed to mark as known:', error)
    await showAlert(t('vocab.alertMarkKnownFailed'))
  } finally {
    processingAction.value = false
  }
}

const moveToTraining = async () => {
  if (processingAction.value || !props.lemma) return
  processingAction.value = true
  try {
    const formData = new FormData()
    await apiClient.requestFormData(withCourseCode(`/api/vocab/${encodeURIComponent(props.lemma)}/move_to_training`), formData)
    emit('close')
    emit('vocabChanged')
  } catch (error) {
    console.error('Failed to move to training:', error)
    await showAlert(t('vocab.alertMoveTrainingFailed'))
  } finally {
    processingAction.value = false
  }
}

const confirmDelete = () => {
  wordToDelete.value = selectedWordDisplay.value
  showDeleteConfirm.value = true
}

const deleteWord = async () => {
  if (!props.lemma) return
  try {
    const formData = new FormData()
    await apiClient.requestFormData(withCourseCode(`/api/vocab/${encodeURIComponent(props.lemma)}/delete`), formData)
    showDeleteConfirm.value = false
    emit('close')
    emit('vocabChanged')
  } catch (error) {
    console.error('Failed to delete word:', error)
    await showAlert(t('vocab.alertRemoveFailed'))
  }
}

onUnmounted(() => {
  hideSrsTooltip(true)
})
</script>

<style scoped>
.vocab-detail-root {
  color: var(--text-primary);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 16px;
}

.word-header-info {
  flex: 1;
}

.word-title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  flex-wrap: nowrap;
  overflow-x: auto;
  overflow-y: visible;
  min-width: 0;
}

.word-header-info h3 {
  margin: 0;
  font-size: 24px;
  white-space: nowrap;
  flex-shrink: 0;
}

.transcription {
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif;
  font-style: italic;
  letter-spacing: 0.5px;
  color: var(--text-secondary);
  font-size: 18px;
  white-space: nowrap;
  flex-shrink: 0;
}

.transcription-with-audio {
  display: inline-flex;
  align-items: center;
  gap: 8px;
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

.word-summary {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
  align-items: center;
}

.mastering-score-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: var(--text-primary);
}

.mastery-dot-inline {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.btn-close {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: var(--text-primary);
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  flex-shrink: 0;
}

.btn-close:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.1));
}

.modal-footer {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.footer-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  flex-wrap: wrap;
}

.modal-nested {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-modal-overlay, rgba(0, 0, 0, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1200;
}

.modal-small {
  max-width: 440px;
  width: 90%;
  padding: 24px;
  background: var(--card-bg);
  border-radius: 8px;
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  justify-content: flex-end;
}

.warning-text {
  color: var(--color-danger, #ef4444);
  font-size: 13px;
  margin-top: 8px;
}

.verb-forms-section {
  margin-bottom: 24px;
  padding: 16px;
  background: var(--input-bg, rgba(0, 0, 0, 0.02));
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  border-radius: 8px;
}

.verb-forms-section h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.verb-forms-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.verb-form-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.verb-form-label {
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 140px;
}

.verb-form-value {
  color: var(--text-primary);
  font-weight: 500;
}

.verb-forms-table-wrap {
  overflow-x: auto;
  margin-top: 4px;
  border: 1px solid var(--border-primary, var(--table-border, rgba(0, 0, 0, 0.12)));
  border-radius: 8px;
  background: var(--card-bg, var(--input-bg, rgba(0, 0, 0, 0.02)));
}

.verb-forms-table {
  width: 100%;
  min-width: 520px;
  border-collapse: collapse;
  font-size: 13px;
  line-height: 1.35;
}

.verb-forms-table thead {
  background: var(--bg-tertiary, rgba(0, 0, 0, 0.06));
}

.verb-forms-table th,
.verb-forms-table td {
  border: 1px solid var(--border-primary, var(--table-border, rgba(0, 0, 0, 0.1)));
  padding: 8px 10px;
  text-align: left;
  vertical-align: middle;
}

.verb-forms-table th {
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
}

.verb-forms-table td {
  color: var(--text-primary);
}

.verb-forms-table tbody tr:nth-child(even) {
  background: var(--bg-secondary, rgba(0, 0, 0, 0.03));
}

.verb-forms-table tbody tr:hover {
  background: var(--hover-bg, rgba(0, 0, 0, 0.06));
}

.verb-forms-table td.verb-forms-pronoun {
  white-space: normal;
  max-width: 14rem;
  line-height: 1.35;
}

.verb-forms-table td:last-child {
  font-weight: 600;
}

.cards-list-simple {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.cards-list-simple + .verb-forms-section {
  margin-top: 24px;
}

.sense-group-simple {
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  border-radius: 8px;
  padding: 16px;
  background: var(--input-bg, rgba(0, 0, 0, 0.02));
}

.sense-header-simple {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.sense-header-simple h4 {
  margin: 0;
  font-size: 18px;
}

.word-ru {
  font-weight: 600;
  color: var(--color-primary);
}

.pos-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  background-color: var(--color-secondary, #6b7280);
  color: white;
  font-weight: 600;
  margin-left: 8px;
}

.sense-info-simple {
  margin-bottom: 16px;
}

.meaning {
  font-size: 15px;
  color: var(--text-primary);
  margin-bottom: 8px;
  line-height: 1.5;
}

.example {
  font-size: 14px;
  color: var(--text-secondary);
  font-style: italic;
  line-height: 1.5;
}

.directions-simple {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.direction-item-simple {
  padding: 12px;
  background: var(--card-bg);
  border-radius: 6px;
  border-left: 3px solid var(--color-primary);
}

.direction-header-simple {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.srs-info-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin-left: auto;
  cursor: help;
}

.srs-info-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
  color: var(--text-secondary);
}

.srs-info-wrap:hover .srs-info-icon {
  opacity: 1;
  color: var(--color-primary);
}

.srs-tooltip-fixed {
  position: fixed;
  transform: translate(-50%, -100%);
  min-width: 240px;
  max-width: 320px;
  padding: 10px 12px;
  background: var(--card-bg);
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.15));
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-primary);
  white-space: nowrap;
  z-index: 1100;
  pointer-events: auto;
}

.srs-tooltip-title {
  font-weight: 600;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.srs-tooltip-row {
  margin-bottom: 2px;
}

.srs-tooltip-row:last-child {
  margin-bottom: 0;
}

.srs-tooltip-row span {
  color: var(--text-secondary);
  margin-right: 6px;
}

.srs-tooltip-hint {
  cursor: help;
  opacity: 0.8;
}

.srs-tooltip-step {
  color: var(--color-primary);
  margin-right: 0;
}

.srs-tooltip-reason {
  margin-top: 8px;
  padding-top: 6px;
  border-top: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  font-size: 11px;
  line-height: 1.4;
  color: var(--text-secondary);
  white-space: normal;
}

.direction-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.direction-ru_en {
  background-color: var(--color-primary, #3b82f6);
  color: white;
}

.direction-en_ru {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.state-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: capitalize;
}

.state-review {
  background-color: var(--color-success, #10b981);
  color: white;
}

.state-learning {
  background-color: var(--color-warning, #f59e0b);
  color: white;
}

.state-new {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.direction-stats-simple {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-secondary);
  flex-wrap: wrap;
}

.no-cards {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.loading {
  text-align: center;
  padding: 20px;
  color: var(--text-primary);
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background-color: var(--color-primary, #3b82f6);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover, #2563eb);
}

.btn-secondary {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.btn-secondary:hover {
  background-color: var(--color-secondary-hover, #4b5563);
}

.btn-danger {
  background-color: var(--color-danger, #ef4444);
  color: white;
}

.btn-danger:hover {
  background-color: var(--color-danger-hover, #dc2626);
}

@media (max-width: 768px) {
  .directions-simple {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }
  .word-summary {
    flex-direction: column;
    gap: 4px;
  }
}

@media (max-width: 400px) {
  .directions-simple {
    grid-template-columns: 1fr;
  }
}
</style>
