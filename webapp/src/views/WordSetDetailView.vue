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
          >
            <span class="word-text">{{ word.display_word || word.word }}</span>
            <span class="status-badge" :class="`badge-${word.status}`">
              {{ word.status === 'known' ? 'Known' : word.status === 'in_vocab' ? 'In Vocab' : 'New' }}
            </span>
          </div>
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
import Icon from '../components/Icon.vue'

interface WordSet {
  id: number
  title: string
  description?: string | null
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

const route = useRoute()
const router = useRouter()
const setId = route.params.setId as string

const loading = ref(true)
const error = ref<string | null>(null)
const wordSet = ref<WordSet | null>(null)
const words = ref<WordInfo[]>([])

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

const goBack = () => {
  router.push('/learning/words')
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
  align-items: center;
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
}

.btn-back:hover {
  background: var(--bg-hover);
  border-color: var(--color-primary);
}

.detail-header h1 {
  margin: 0;
  flex: 1;
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

@media (max-width: 768px) {
  .words-list {
    grid-template-columns: 1fr;
  }
  
  .word-set-meta {
    flex-direction: column;
    gap: 8px;
  }
}
</style>
