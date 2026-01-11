<template>
  <div class="word-set-study">
    <div class="study-header">
      <button @click="exitStudy" class="btn-exit">
        <Icon name="close" />
      </button>
      <div class="study-progress">
        <span>{{ currentIndex + 1 }} / {{ totalWords }}</span>
      </div>
    </div>
    
    <div v-if="loading" class="loading">Loading words...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="currentWord" class="study-card">
      <div v-if="loadingCard" class="loading">Loading card...</div>
      <div v-else-if="currentTrainingCard" class="word-display">
        <h2>{{ currentTrainingCard.word_en || currentWord.display_word || currentWord.word }}</h2>
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
      <div v-else class="word-display">
        <h2>{{ currentWord.display_word || currentWord.word }}</h2>
      </div>
      
      <div class="study-actions">
        <button @click="markKnown" class="btn btn-know" :disabled="processing">
          Know
        </button>
        <button @click="markLearn" class="btn btn-learn" :disabled="processing">
          Learn
        </button>
        <button @click="skipWord" class="btn btn-skip" :disabled="processing">
          Skip
        </button>
      </div>
    </div>
    <div v-else class="complete-message">
      <h2>Study Complete!</h2>
      <p>You've reviewed all words in this set.</p>
      <button @click="goBack" class="btn btn-primary">Back to Set</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import Icon from '../components/Icon.vue'

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
const words = ref<WordInfo[]>([])
const currentIndex = ref(0)
const processing = ref(false)
const loadingCard = ref(false)
const currentTrainingCard = ref<TrainingCard | null>(null)

const currentWord = computed(() => {
  if (currentIndex.value >= words.value.length) {
    return null
  }
  return words.value[currentIndex.value]
})

const totalWords = computed(() => {
  return words.value.length
})

// Watch for current word changes and load training card
watch(currentWord, async (newWord) => {
  if (newWord) {
    await loadTrainingCard(newWord.word_card_id)
  } else {
    currentTrainingCard.value = null
  }
})

onMounted(async () => {
  await loadWords()
})

const loadWords = async () => {
  loading.value = true
  error.value = null
  try {
    const data: { word_set: any; words: WordInfo[] } = 
      await apiClient.request(`/app/learning/words/sets/${setId}`)
    // Filter to only unknown words for study
    words.value = (data.words || []).filter(w => w.status === 'unknown')
    currentIndex.value = 0
  } catch (error: any) {
    console.error('Failed to load words:', error)
    error.value = error.message || 'Failed to load words'
  } finally {
    loading.value = false
  }
}

const loadTrainingCard = async (wordCardId: number) => {
  loadingCard.value = true
  currentTrainingCard.value = null
  try {
    const data: { training_card: TrainingCard } = 
      await apiClient.request(`/app/learning/words/sets/${setId}/study?word_card_id=${wordCardId}`)
    currentTrainingCard.value = data.training_card
  } catch (error: any) {
    console.error('Failed to load training card:', error)
    // Don't show error to user, just continue without card
    currentTrainingCard.value = null
  } finally {
    loadingCard.value = false
  }
}

const markKnown = async () => {
  if (!currentWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/app/learning/words/sets/${setId}/study/know`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word_card_id: currentWord.value.word_card_id })
    })
    
    // Remove word from list (it's now known, so not in unknown list)
    words.value = words.value.filter(w => w.word_card_id !== currentWord.value!.word_card_id)
    
    // Move to next word (index stays the same, but list is shorter)
    if (currentIndex.value >= words.value.length) {
      currentIndex.value = words.value.length
    }
  } catch (error: any) {
    console.error('Failed to mark as known:', error)
    await showAlert(error.message || 'Failed to mark as known')
  } finally {
    processing.value = false
  }
}

const markLearn = async () => {
  if (!currentWord.value || processing.value) return
  
  processing.value = true
  try {
    await apiClient.request(`/app/learning/words/sets/${setId}/study/learn`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ word_card_id: currentWord.value.word_card_id })
    })
    
    // Remove word from list (it's now in vocab, so not in unknown list)
    words.value = words.value.filter(w => w.word_card_id !== currentWord.value!.word_card_id)
    
    // Move to next word (index stays the same, but list is shorter)
    if (currentIndex.value >= words.value.length) {
      currentIndex.value = words.value.length
    }
  } catch (error: any) {
    console.error('Failed to add to learning:', error)
    await showAlert(error.message || 'Failed to add to learning')
  } finally {
    processing.value = false
  }
}

const skipWord = () => {
  nextWord()
}

const nextWord = () => {
  if (currentIndex.value < words.value.length - 1) {
    currentIndex.value++
  } else {
    // All words reviewed
    currentIndex.value = words.value.length
  }
}

const exitStudy = async () => {
  const confirmed = await showConfirm('Are you sure you want to exit? Your progress will be saved.')
  if (confirmed) {
    goBack()
  }
}

const goBack = () => {
  router.push(`/learning/words/${setId}`)
}
</script>

<style scoped>
.word-set-study {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
  min-height: 60vh;
  display: flex;
  flex-direction: column;
}

.study-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.btn-exit {
  background: transparent;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--text-primary);
  padding: 8px;
  border-radius: 4px;
  transition: background 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
}

.btn-exit:hover {
  background: var(--bg-hover);
}

.study-progress {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-secondary);
}

.study-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 48px;
}

.word-display {
  text-align: center;
}

.word-display h2 {
  font-size: 48px;
  margin: 0;
  color: var(--text-primary);
}

.transcription {
  font-size: 18px;
  color: var(--text-secondary);
  margin-top: 8px;
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

.btn {
  padding: 16px 32px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 120px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.complete-message {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  gap: 24px;
}

.complete-message h2 {
  margin: 0;
  font-size: 32px;
}

.complete-message p {
  font-size: 18px;
  color: var(--text-secondary);
}

.loading, .error {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.error {
  color: var(--color-danger, #ef4444);
}

@media (max-width: 768px) {
  .word-display h2 {
    font-size: 36px;
  }
  
  .study-actions {
    flex-direction: column;
    width: 100%;
  }
  
  .btn {
    width: 100%;
  }
}
</style>
