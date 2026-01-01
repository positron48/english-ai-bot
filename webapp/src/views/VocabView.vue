<template>
  <div class="vocab">
    <h1>Vocabulary</h1>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else>
      <div v-if="words.length === 0" class="card">
        <p>No words in your vocabulary yet.</p>
      </div>
      
      <div v-else>
        <div v-for="word in words" :key="word.word_en" class="card vocab-item">
          <div class="vocab-header">
            <h2>{{ word.word_en }}</h2>
            <button @click="confirmDelete(word.word_en)" class="btn btn-danger">Delete</button>
          </div>
          <div class="vocab-stats">
            <p>Total cards: {{ word.total_cards }}</p>
            <p>Due: {{ word.due_count }}</p>
            <p v-if="word.last_review">Last review: {{ formatDate(word.last_review) }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showDeleteConfirm" class="modal">
      <div class="modal-content">
        <h3>Confirm Delete</h3>
        <p>Are you sure you want to delete "{{ wordToDelete }}" from your vocabulary?</p>
        <div class="modal-actions">
          <button @click="deleteWord" class="btn btn-danger">Delete</button>
          <button @click="showDeleteConfirm = false" class="btn btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface VocabWord {
  word_en: string
  total_cards: number
  due_count: number
  last_review: string | null
}

const words = ref<VocabWord[]>([])
const loading = ref(true)
const showDeleteConfirm = ref(false)
const wordToDelete = ref('')

onMounted(async () => {
  await loadVocab()
})

const loadVocab = async () => {
  loading.value = true
  try {
    const data: { words: VocabWord[] } = await apiClient.request('/app/vocab')
    words.value = data.words
  } catch (error) {
    console.error('Failed to load vocabulary:', error)
  } finally {
    loading.value = false
  }
}

const confirmDelete = (word: string) => {
  wordToDelete.value = word
  showDeleteConfirm.value = true
}

const deleteWord = async () => {
  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/app/vocab/${wordToDelete.value}/delete`, formData)
    showDeleteConfirm.value = false
    await loadVocab()
  } catch (error) {
    console.error('Failed to delete word:', error)
    alert('Failed to delete word')
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString()
}
</script>

<style scoped>
.vocab-item {
  margin-bottom: 20px;
}

.vocab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.vocab-stats {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-modal-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--card-bg);
  padding: 30px;
  border-radius: 8px;
  max-width: 400px;
  width: 90%;
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}
</style>

