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
          <span v-if="displayMasteringScore !== null && displayMasteringScore !== undefined" class="mastering-score-inline" :title="t('vocab.scoreLabel') + ' 0–100'">
            <span class="mastery-dot-inline" :style="{ backgroundColor: masteryColor(displayMasteringScore) }" />
            {{ displayMasteringScore }}
          </span>
          <span v-if="masteryLevel">{{ t(`vocab.${masteryLevel}`) }}</span>
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
  mastery_level?: string
  mastering_score?: number
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
const masteryLevel = ref('')
const payloadMasteringScore = ref<number | null>(null)
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
  masteryLevel.value = data.mastery_level || ''
  payloadMasteringScore.value = typeof data.mastering_score === 'number' ? data.mastering_score : null
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
      })
    }
  }
  return Array.from(groups.values())
    .sort((a, b) => a.sense_index - b.sense_index)
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

const displayMasteringScore = computed(() => (
  props.listMasteringScore !== null && props.listMasteringScore !== undefined
    ? props.listMasteringScore
    : payloadMasteringScore.value
))

const masteryColor = (score: number): string => {
  const s = Math.max(0, Math.min(100, score ?? 0))
  const hue = (120 * s) / 100
  return `hsl(${hue}, 72%, 42%)`
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
  .word-summary {
    flex-direction: column;
    gap: 4px;
  }
}
</style>
