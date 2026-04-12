<template>
  <div class="word-set-detail">
    <div class="detail-header">
      <button @click="goBack" class="btn-back">
        {{ t('wordSets.back') }}
      </button>
      <h1>{{ wordSet?.title }}</h1>
    </div>
    
    <div v-if="loading" class="loading">{{ t('wordSets.loading') }}</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="wordSet">
      <div class="word-set-info">
        <p v-if="wordSet.description" class="description">{{ wordSet.description }}</p>
        <div class="word-set-meta">
          <span class="total-words">{{ t('wordSets.totalWords', { n: wordSet.total_words }) }}</span>
          <span class="progress-info">
            {{ t('wordSets.progress') }}: {{ Math.round(wordSet.progress_percent) }}%
            ({{ wordSet.known_words + wordSet.words_in_vocab }}/{{ wordSet.total_words }})
          </span>
        </div>
      </div>
      
      <div class="words-section">
        <h2>{{ t('wordSets.words') }}</h2>
        <div class="words-list">
          <div 
            v-for="word in words" 
            :key="word.word_card_id" 
            class="word-item"
            :class="`status-${word.status}`"
            @click="openWordCard(word)"
          >
            <span class="word-text">{{ word.display_word || word.word }}</span>
            <span class="status-badge" :class="`badge-${word.status}`">
              {{ word.status === 'known' ? t('wordSets.statusKnown') : word.status === 'in_vocab' ? t('wordSets.statusInVocab') : t('wordSets.statusNew') }}
            </span>
          </div>
        </div>
      </div>
      
      <!-- Word Card Modal -->
      <div v-if="showWordModal" class="modal" @click.self="closeWordModal">
        <div class="modal-content modal-word-card">
          <div class="modal-header">
            <h3>{{ t('wordSets.wordCard') }}</h3>
            <button @click="closeWordModal" class="btn-close">&times;</button>
          </div>
          
          <div v-if="loadingCard" class="loading">{{ t('wordSets.loadingCard') }}</div>
          <div v-else-if="currentTrainingCard" class="word-card-content">
            <div class="word-display">
              <h2>{{ currentTrainingCard.display_target || currentTrainingCard.display_word || currentTrainingCard.word_target || currentTrainingCard.word_en || selectedWord?.display_target || selectedWord?.display_word || selectedWord?.word }}</h2>
              <div v-if="currentTrainingCard.transcription" class="transcription-with-audio">
                <div class="transcription">[{{ currentTrainingCard.transcription }}]</div>
                <button
                  v-if="currentPronunciationURL"
                  class="btn-pronunciation"
                  :disabled="playingPronunciation"
                  :title="t('wordSets.pronounceAria')"
                  :aria-label="t('wordSets.pronounceAria')"
                  @click="playCurrentPronunciation"
                >
                  <Icon name="play" />
                </button>
              </div>
              <div v-if="formatMorph(currentTrainingCard.morph)" class="morph">
                {{ formatMorph(currentTrainingCard.morph) }}
              </div>
              <div v-if="currentTrainingCard.word_native || currentTrainingCard.word_ru" class="translation">
                {{ currentTrainingCard.word_native || currentTrainingCard.word_ru }}
              </div>
              <div v-if="currentTrainingCard.meaning_target || currentTrainingCard.meaning_en" class="meaning">
                {{ currentTrainingCard.meaning_target || currentTrainingCard.meaning_en }}
              </div>
              <div v-if="currentTrainingCard.example_target || currentTrainingCard.example_en" class="example">
                <strong>{{ t('wordSets.example') }}:</strong> {{ currentTrainingCard.example_target || currentTrainingCard.example_en }}
              </div>
              <div v-if="currentTrainingCard.example_native || currentTrainingCard.example_ru" class="example-ru">
                {{ currentTrainingCard.example_native || currentTrainingCard.example_ru }}
              </div>
            </div>
            
            <div class="study-actions">
              <button @click="markKnown" class="btn btn-know" :disabled="processing">
                {{ t('wordSets.know') }}
              </button>
              <button @click="markLearn" class="btn btn-learn" :disabled="processing">
                {{ t('wordSets.learn') }}
              </button>
              <button @click="closeWordModal" class="btn btn-skip" :disabled="processing">
                {{ t('wordSets.close') }}
              </button>
            </div>
          </div>
          <div v-else class="error">{{ t('wordSets.failedLoadCard') }}</div>
        </div>
      </div>
      
      <div class="actions-section" v-if="wordSet.unknown_words > 0">
        <button @click="startStudy" class="btn btn-primary btn-large">
          {{ t('wordSets.startLearning', { n: wordSet.unknown_words }) }}
        </button>
      </div>
      <div v-else class="complete-message">
        <p>{{ t('wordSets.allInVocab') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { showAlert } from '../composables/useDialog'
import { useAudio } from '../composables/useAudio'
import Icon from '../components/Icon.vue'

interface WordSet {
  id: number
  title: string
  description?: string | null
  category_id?: number | null
  total_words: number
  known_words: number
  words_in_vocab: number
  unknown_words: number
  progress_percent: number
}

interface WordInfo {
  word_card_id: number
  word: string
  display_word: string
  display_target?: string
  status: 'unknown' | 'in_vocab' | 'known'
}

interface TrainingCard {
  id: number
  word_card_id: number
  word_en: string
  word_target?: string
  transcription?: string
  sense_index: number
  word_ru?: string
  word_native?: string
  meaning_en?: string
  meaning_target?: string
  example_en?: string
  example_target?: string
  example_ru?: string
  example_native?: string
  display_word?: string
  display_target?: string
  distractors_ru?: string
  distractors_en?: string
  hint?: string
  pos?: string
  display_word?: string
  morph?: MorphInfo
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

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const setId = route.params.setId as string

const loading = ref(true)
const error = ref<string | null>(null)
const wordSet = ref<WordSet | null>(null)
const words = ref<WordInfo[]>([])

const showWordModal = ref(false)
const selectedWord = ref<WordInfo | null>(null)
const currentTrainingCard = ref<TrainingCard | null>(null)
const loadingCard = ref(false)
const processing = ref(false)
const currentPronunciationURL = ref<string | null>(null)
const playingPronunciation = ref(false)
const { getWordPronunciationURL, playWordPronunciation } = useAudio()

let openWordCardGen = 0

const formatMorph = (morph?: MorphInfo): string => {
  if (!morph) return ''
  if (morph.pos === 'noun' && morph.noun_gender) {
    const core = morph.article ? `${morph.article} • ${morph.noun_gender}` : morph.noun_gender
    return morph.opposite_gender_word ? `${core} (${morph.opposite_gender_word})` : core
  }
  if (morph.pos === 'verb' && morph.verb_forms) {
    const forms = [morph.verb_forms.v1, morph.verb_forms.v2, morph.verb_forms.v3].filter(Boolean)
    if (forms.length > 0) return forms.join(', ')
  }
  return ''
}

onMounted(async () => {
  await loadWordSet()
})

const loadWordSet = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { word_set: WordSet; words: WordInfo[] } = 
      await apiClient.request(`/api/learning/words/sets/${setId}`)
    wordSet.value = data.word_set
    words.value = data.words || []
  } catch (error: any) {
    console.error('Failed to load word set:', error)
    error.value = error.message || t('wordSets.loadFailed')
  } finally {
    loading.value = false
  }
}

const openWordCard = async (word: WordInfo) => {
  const targetWordCardId = word.word_card_id
  const gen = ++openWordCardGen
  selectedWord.value = word
  showWordModal.value = true
  loadingCard.value = true
  currentTrainingCard.value = null
  currentPronunciationURL.value = null
  
  try {
    const data: { training_card: TrainingCard } = 
      await apiClient.request(`/api/learning/words/sets/${setId}/study?word_card_id=${word.word_card_id}`)
    if (gen !== openWordCardGen) {
      return
    }
    currentTrainingCard.value = data.training_card
    if (currentTrainingCard.value?.transcription) {
      const pronunciationWord =
        currentTrainingCard.value.word_target ||
        currentTrainingCard.value.word_en ||
        selectedWord.value?.word ||
        ''
      const url = pronunciationWord ? await getWordPronunciationURL(pronunciationWord) : null
      if (gen !== openWordCardGen) {
        return
      }
      currentPronunciationURL.value = url
    }
  } catch (error: any) {
    console.error('Failed to load training card:', error)
    if (gen === openWordCardGen) {
      currentTrainingCard.value = null
    }
  } finally {
    if (gen === openWordCardGen) {
      loadingCard.value = false
    }
  }
}

const closeWordModal = () => {
  openWordCardGen++
  showWordModal.value = false
  selectedWord.value = null
  currentTrainingCard.value = null
  currentPronunciationURL.value = null
  loadingCard.value = false
}

const handleKeydown = (event: KeyboardEvent) => {
  if (showWordModal.value && event.key === 'Escape') {
    event.preventDefault()
    closeWordModal()
  }
}

watch(showWordModal, (isOpen) => {
  if (isOpen) {
    window.addEventListener('keydown', handleKeydown)
  } else {
    window.removeEventListener('keydown', handleKeydown)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

const playCurrentPronunciation = async () => {
  if (playingPronunciation.value || !currentTrainingCard.value) return
  const pronunciationWord =
    currentTrainingCard.value.word_target ||
    currentTrainingCard.value.word_en ||
    selectedWord.value?.word ||
    ''
  if (!pronunciationWord) return
  playingPronunciation.value = true
  try {
    await playWordPronunciation(pronunciationWord)
  } finally {
    playingPronunciation.value = false
  }
}

const markKnown = async () => {
  if (!selectedWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/api/learning/words/sets/${setId}/study/know`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word_card_id: selectedWord.value.word_card_id })
    })
    
    // Update word status in the list
    const wordIndex = words.value.findIndex(w => w.word_card_id === selectedWord.value!.word_card_id)
    if (wordIndex !== -1) {
      words.value[wordIndex].status = 'known'
    }
    
    // Reload word set to update progress
    await loadWordSet()
    
    closeWordModal()
  } catch (error: any) {
    console.error('Failed to mark as known:', error)
    await showAlert(error.message || t('wordSets.markKnownFailed'))
  } finally {
    processing.value = false
  }
}

const markLearn = async () => {
  if (!selectedWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/api/learning/words/sets/${setId}/study/learn`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word_card_id: selectedWord.value.word_card_id })
    })
    
    // Update word status in the list
    const wordIndex = words.value.findIndex(w => w.word_card_id === selectedWord.value!.word_card_id)
    if (wordIndex !== -1) {
      words.value[wordIndex].status = 'in_vocab'
    }
    
    // Reload word set to update progress
    await loadWordSet()
    
    closeWordModal()
  } catch (error: any) {
    console.error('Failed to add to learning:', error)
    await showAlert(error.message || t('wordSets.addLearningFailed'))
  } finally {
    processing.value = false
  }
}

const goBack = () => {
  if (wordSet.value?.category_id) {
    router.push(`/learning/words?category_id=${wordSet.value.category_id}`)
  } else {
    router.push('/learning/words')
  }
}

const startStudy = () => {
  router.push(`/learning/words/${setId}/study`)
}
</script>

<style scoped>
.word-set-detail {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.detail-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 24px;
}

.btn-back {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
  flex-shrink: 0;
  white-space: nowrap;
}

.btn-back:hover {
  background: var(--bg-hover);
  border-color: var(--color-primary);
}

.detail-header h1 {
  margin: 0;
  flex: 1;
  min-width: 0;
  word-wrap: break-word;
  overflow-wrap: break-word;
  line-height: 1.3;
}

.word-set-info {
  margin-bottom: 32px;
}

.description {
  font-size: 16px;
  color: var(--text-secondary);
  margin-bottom: 16px;
  line-height: 1.6;
}

.word-set-meta {
  display: flex;
  gap: 24px;
  font-size: 14px;
  color: var(--text-secondary);
}

.words-section {
  margin-bottom: 32px;
}

.words-section h2 {
  margin-bottom: 16px;
}

.words-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.word-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  background: var(--card-bg);
  gap: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.word-item:hover {
  border-color: var(--color-primary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.word-text {
  font-weight: 600;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  white-space: nowrap;
}

.badge-known {
  background: var(--color-success, #10b981);
  color: white;
}

.badge-in_vocab {
  background: var(--color-primary, #3b82f6);
  color: white;
}

.badge-unknown {
  background: var(--color-secondary, #6b7280);
  color: white;
}

.actions-section {
  text-align: center;
  padding: 32px 0;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
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

.btn-large {
  padding: 16px 32px;
  font-size: 18px;
}

.complete-message {
  text-align: center;
  padding: 32px;
  color: var(--text-secondary);
  font-size: 16px;
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger, #ef4444);
}

/* Modal Styles */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-modal-overlay, rgba(0, 0, 0, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-bg);
  padding: 30px;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  color: var(--text-primary);
}

.modal-word-card {
  max-width: 700px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  margin: 0;
  font-size: 24px;
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

.word-card-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.word-display {
  text-align: center;
  overflow-x: auto;
  overflow-y: visible;
  width: 100%;
}

.word-display h2 {
  font-size: 36px;
  margin: 0;
  color: var(--text-primary);
  white-space: nowrap;
}

.transcription {
  font-size: 18px;
  color: var(--text-secondary);
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif;
  font-style: italic;
  white-space: nowrap;
}

.transcription-with-audio {
  margin-top: 8px;
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

.translation {
  font-size: 24px;
  color: var(--text-primary);
  margin-top: 16px;
  font-weight: 600;
}

.morph {
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.meaning {
  font-size: 16px;
  color: var(--text-secondary);
  margin-top: 16px;
  font-style: italic;
}

.example {
  font-size: 16px;
  color: var(--text-primary);
  margin-top: 24px;
  padding: 16px;
  background: var(--bg-hover, rgba(0, 0, 0, 0.05));
  border-radius: 8px;
}

.example-ru {
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 8px;
  padding: 0 16px;
}

.study-actions {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  justify-content: center;
}

.btn-know {
  background-color: var(--color-success, #10b981);
  color: white;
}

.btn-know:hover:not(:disabled) {
  background-color: #059669;
}

.btn-learn {
  background-color: var(--color-primary, #3b82f6);
  color: white;
}

.btn-learn:hover:not(:disabled) {
  background-color: #2563eb;
}

.btn-skip {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.btn-skip:hover:not(:disabled) {
  background-color: #4b5563;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .word-set-detail {
    padding: 12px;
  }
  
  .detail-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }
  
  .btn-back {
    align-self: flex-start;
    padding: 8px 12px;
    font-size: 13px;
  }
  
  .detail-header h1 {
    font-size: 22px;
    margin: 0;
    width: 100%;
  }
  
  .words-list {
    grid-template-columns: 1fr;
  }
  
  .word-set-meta {
    flex-direction: column;
    gap: 8px;
  }
  
  .modal-content {
    padding: 20px;
    width: 95%;
  }
  
  .word-display h2 {
    font-size: 28px;
  }
  
  .study-actions {
    flex-direction: column;
    width: 100%;
  }
  
  .study-actions .btn {
    width: 100%;
  }
}
</style>
