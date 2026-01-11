<template>
  <div class="word-set-detail">
    <div class="detail-header">
      <button @click="goBack" class="btn-back">
        ← Back
      </button>
      <h1>{{ wordSet?.title }}</h1>
    </div>
    
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="wordSet">
      <div class="word-set-info">
        <p v-if="wordSet.description" class="description">{{ wordSet.description }}</p>
        <div class="word-set-meta">
          <span class="total-words">{{ wordSet.total_words }} words</span>
          <span class="progress-info">
            Progress: {{ Math.round(wordSet.progress_percent) }}%
            ({{ wordSet.known_words + wordSet.words_in_vocab }}/{{ wordSet.total_words }})
          </span>
        </div>
      </div>
      
      <div class="words-section">
        <h2>Words</h2>
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
              {{ word.status === 'known' ? 'Known' : word.status === 'in_vocab' ? 'In Vocab' : 'New' }}
            </span>
          </div>
        </div>
      </div>
      
      <!-- Word Card Modal -->
      <div v-if="showWordModal" class="modal" @click.self="closeWordModal">
        <div class="modal-content modal-word-card">
          <div class="modal-header">
            <h3>Word Card</h3>
            <button @click="closeWordModal" class="btn-close">&times;</button>
          </div>
          
          <div v-if="loadingCard" class="loading">Loading card...</div>
          <div v-else-if="currentTrainingCard" class="word-card-content">
            <div class="word-display">
              <h2>{{ currentTrainingCard.word_en || selectedWord?.display_word || selectedWord?.word }}</h2>
              <div v-if="currentTrainingCard.transcription" class="transcription">
                [{{ currentTrainingCard.transcription }}]
              </div>
              <div v-if="currentTrainingCard.word_ru" class="translation">
                {{ currentTrainingCard.word_ru }}
              </div>
              <div v-if="currentTrainingCard.meaning_en" class="meaning">
                {{ currentTrainingCard.meaning_en }}
              </div>
              <div v-if="currentTrainingCard.example_en" class="example">
                <strong>Example:</strong> {{ currentTrainingCard.example_en }}
              </div>
              <div v-if="currentTrainingCard.example_ru" class="example-ru">
                {{ currentTrainingCard.example_ru }}
              </div>
            </div>
            
            <div class="study-actions">
              <button @click="markKnown" class="btn btn-know" :disabled="processing">
                Know
              </button>
              <button @click="markLearn" class="btn btn-learn" :disabled="processing">
                Learn
              </button>
              <button @click="closeWordModal" class="btn btn-skip" :disabled="processing">
                Close
              </button>
            </div>
          </div>
          <div v-else class="error">Failed to load word card</div>
        </div>
      </div>
      
      <div class="actions-section" v-if="wordSet.unknown_words > 0">
        <button @click="startStudy" class="btn btn-primary btn-large">
          Start Learning ({{ wordSet.unknown_words }} new words)
        </button>
      </div>
      <div v-else class="complete-message">
        <p>All words in this set are already in your vocabulary!</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { showAlert } from '../composables/useDialog'

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
  status: 'unknown' | 'in_vocab' | 'known'
}

interface TrainingCard {
  id: number
  word_card_id: number
  word_en: string
  transcription?: string
  sense_index: number
  word_ru?: string
  meaning_en?: string
  example_en?: string
  example_ru?: string
  distractors_ru?: string
  distractors_en?: string
  hint?: string
  pos?: string
  display_word?: string
}

const route = useRoute()
const router = useRouter()
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

onMounted(async () => {
  await loadWordSet()
})

const loadWordSet = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { word_set: WordSet; words: WordInfo[] } = 
      await apiClient.request(`/app/learning/words/sets/${setId}`)
    wordSet.value = data.word_set
    words.value = data.words || []
  } catch (error: any) {
    console.error('Failed to load word set:', error)
    error.value = error.message || 'Failed to load word set'
  } finally {
    loading.value = false
  }
}

const openWordCard = async (word: WordInfo) => {
  selectedWord.value = word
  showWordModal.value = true
  loadingCard.value = true
  currentTrainingCard.value = null
  
  try {
    const data: { training_card: TrainingCard } = 
      await apiClient.request(`/app/learning/words/sets/${setId}/study?word_card_id=${word.word_card_id}`)
    currentTrainingCard.value = data.training_card
  } catch (error: any) {
    console.error('Failed to load training card:', error)
    currentTrainingCard.value = null
  } finally {
    loadingCard.value = false
  }
}

const closeWordModal = () => {
  showWordModal.value = false
  selectedWord.value = null
  currentTrainingCard.value = null
}

const markKnown = async () => {
  if (!selectedWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/app/learning/words/sets/${setId}/study/know`, {
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
    await showAlert(error.message || 'Failed to mark as known')
  } finally {
    processing.value = false
  }
}

const markLearn = async () => {
  if (!selectedWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/app/learning/words/sets/${setId}/study/learn`, {
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
    await showAlert(error.message || 'Failed to add to learning')
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
}

.word-display h2 {
  font-size: 36px;
  margin: 0;
  color: var(--text-primary);
}

.transcription {
  font-size: 18px;
  color: var(--text-secondary);
  margin-top: 8px;
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif;
  font-style: italic;
}

.translation {
  font-size: 24px;
  color: var(--text-primary);
  margin-top: 16px;
  font-weight: 600;
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
