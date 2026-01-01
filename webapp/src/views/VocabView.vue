<template>
  <div class="vocab">
    <div class="vocab-header">
      <h1>Vocabulary</h1>
      <div class="search-box">
        <input
          type="text"
          v-model="searchQuery"
          @input="onSearchInput"
          placeholder="Search words..."
          class="search-input"
        />
      </div>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else>
      <div v-if="words.length === 0" class="card">
        <p v-if="searchQuery">
          No words found matching "{{ searchQuery }}".
        </p>
        <p v-else>
          No words in your vocabulary yet.
        </p>
      </div>
      
      <div v-else>
        <div class="table-container">
          <table class="vocab-table">
            <thead>
              <tr>
                <th @click="sortBy('word_en')" class="sortable">
                  Word
                  <span v-if="sortField === 'word_en'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('total_cards')" class="sortable">
                  Cards
                  <span v-if="sortField === 'total_cards'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('mastery_level')" class="sortable">
                  Mastery
                  <span v-if="sortField === 'mastery_level'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('total_reps')" class="sortable">
                  Reps
                  <span v-if="sortField === 'total_reps'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('review_count')" class="sortable">
                  Reviews
                  <span v-if="sortField === 'review_count'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('due_count')" class="sortable">
                  Due
                  <span v-if="sortField === 'due_count'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('added_at')" class="sortable">
                  Added
                  <span v-if="sortField === 'added_at'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th @click="sortBy('last_review')" class="sortable">
                  Last Review
                  <span v-if="sortField === 'last_review'" class="sort-indicator">
                    {{ sortOrder === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="word in words" :key="word.word_en">
                <td class="word-cell">{{ word.word_en }}</td>
                <td>
                  <button 
                    @click="showCards(word.word_en)" 
                    class="link-button"
                    :title="`View ${word.total_cards} cards`"
                  >
                    {{ word.total_cards }}
                  </button>
                </td>
                <td>
                  <span :class="['mastery-badge', `mastery-${word.mastery_level}`]">
                    {{ word.mastery_level }}
                  </span>
                </td>
                <td>{{ word.total_reps }}</td>
                <td>{{ word.review_count }}</td>
                <td>{{ word.due_count }}</td>
                <td>{{ formatDate(word.added_at) }}</td>
                <td>{{ formatDate(word.last_review) }}</td>
                <td>
                  <button 
                    @click="confirmDelete(word.word_en)" 
                    class="btn btn-danger btn-sm"
                    title="Delete word"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="pagination" v-if="pagination.total_pages > 1">
          <button 
            @click="goToPage(pagination.page - 1)" 
            :disabled="pagination.page <= 1"
            class="btn btn-secondary"
          >
            Previous
          </button>
          <span class="page-info">
            Page {{ pagination.page }} of {{ pagination.total_pages }} 
            ({{ pagination.total }} total)
          </span>
          <button 
            @click="goToPage(pagination.page + 1)" 
            :disabled="pagination.page >= pagination.total_pages"
            class="btn btn-secondary"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="modal" @click.self="showDeleteConfirm = false">
      <div class="modal-content">
        <h3>Confirm Delete</h3>
        <p>Are you sure you want to delete "{{ wordToDelete }}" from your vocabulary?</p>
        <div class="modal-actions">
          <button @click="deleteWord" class="btn btn-danger">Delete</button>
          <button @click="showDeleteConfirm = false" class="btn btn-secondary">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Cards Detail Modal -->
    <div v-if="showCardsModal" class="modal" @click.self="closeCardsModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Cards for "{{ selectedWord }}"</h3>
          <button @click="closeCardsModal" class="btn-close">&times;</button>
        </div>
        <div v-if="cardsLoading" class="loading">Loading cards...</div>
        <div v-else-if="cards.length === 0" class="no-cards">No cards found.</div>
        <div v-else class="cards-list">
          <div v-for="card in cards" :key="`${card.training_card_id}-${card.direction}`" class="card-item">
            <div class="card-header">
              <h4>
                {{ card.word_ru }} 
                <span class="direction-badge" :class="`direction-${card.direction}`">
                  {{ card.direction === 'ru_en' ? 'RU→EN' : 'EN→RU' }}
                </span>
                <span class="sense-badge">Sense {{ card.sense_index }}</span>
              </h4>
            </div>
            <div class="card-body">
              <div class="card-row">
                <span class="label">Meaning:</span>
                <span>{{ card.meaning_en }}</span>
              </div>
              <div class="card-row" v-if="card.example_en">
                <span class="label">Example EN:</span>
                <span>{{ card.example_en }}</span>
              </div>
              <div class="card-row" v-if="card.example_ru">
                <span class="label">Example RU:</span>
                <span>{{ card.example_ru }}</span>
              </div>
              <div class="card-row" v-if="card.transcription">
                <span class="label">Transcription:</span>
                <span class="transcription">{{ card.transcription }}</span>
              </div>
              <div class="card-stats">
                <div class="stat-item">
                  <span class="stat-label">State:</span>
                  <span :class="['state-badge', `state-${card.state}`]">{{ card.state }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Reps:</span>
                  <span>{{ card.reps }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Reviews:</span>
                  <span>{{ card.review_count }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">EF:</span>
                  <span>{{ card.ef.toFixed(2) }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Interval:</span>
                  <span>{{ card.interval_days }} days</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">Lapses:</span>
                  <span>{{ card.lapse_count }}</span>
                </div>
                <div class="stat-item" v-if="card.next_due_at">
                  <span class="stat-label">Next Due:</span>
                  <span>{{ formatDate(card.next_due_at) }}</span>
                </div>
                <div class="stat-item" v-if="card.last_review_at">
                  <span class="stat-label">Last Review:</span>
                  <span>{{ formatDate(card.last_review_at) }}</span>
                </div>
                <div class="stat-item" v-if="card.last_quality !== null">
                  <span class="stat-label">Last Quality:</span>
                  <span>{{ card.last_quality }}</span>
                </div>
              </div>
            </div>
          </div>
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
  total_reps: number
  added_at: string | null
  mastery_level: string
  review_count: number
}

interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
}

interface CardDetail {
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
  created_at: string
  updated_at: string
  word_ru: string
  meaning_en: string
  example_en: string
  example_ru: string
  transcription: string
  sense_index: number
  review_count: number
}

const words = ref<VocabWord[]>([])
const loading = ref(true)
const searchQuery = ref('')
const searchTimeout = ref<number | null>(null)
const sortField = ref<string>('word_en')
const sortOrder = ref<'asc' | 'desc'>('asc')
const pagination = ref<Pagination>({
  page: 1,
  limit: 100,
  total: 0,
  total_pages: 0
})

const showDeleteConfirm = ref(false)
const wordToDelete = ref('')

const showCardsModal = ref(false)
const selectedWord = ref('')
const cards = ref<CardDetail[]>([])
const cardsLoading = ref(false)

onMounted(async () => {
  await loadVocab()
})

const onSearchInput = () => {
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }
  searchTimeout.value = window.setTimeout(() => {
    pagination.value.page = 1
    loadVocab()
  }, 500)
}

const sortBy = (field: string) => {
  if (sortField.value === field) {
    // Toggle sort order if clicking the same field
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    // New field, default to ascending
    sortField.value = field
    sortOrder.value = 'asc'
  }
  pagination.value.page = 1
  loadVocab()
}

const loadVocab = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: pagination.value.page.toString(),
      limit: pagination.value.limit.toString(),
      sort_by: sortField.value,
      sort_order: sortOrder.value
    })
    if (searchQuery.value) {
      params.append('search', searchQuery.value)
    }
    
    const data: { words: VocabWord[], pagination: Pagination } = await apiClient.request(`/app/vocab?${params.toString()}`)
    words.value = data.words || []
    pagination.value = data.pagination || {
      page: 1,
      limit: 100,
      total: 0,
      total_pages: 0
    }
  } catch (error) {
    console.error('Failed to load vocabulary:', error)
    words.value = []
    pagination.value = {
      page: 1,
      limit: 100,
      total: 0,
      total_pages: 0
    }
  } finally {
    loading.value = false
  }
}

const goToPage = (page: number) => {
  if (page >= 1 && page <= pagination.value.total_pages) {
    pagination.value.page = page
    loadVocab()
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

const showCards = async (word: string) => {
  selectedWord.value = word
  showCardsModal.value = true
  cardsLoading.value = true
  cards.value = []
  
  try {
    const data: { word_en: string, cards: CardDetail[] } = await apiClient.request(`/app/vocab/${word}/cards`)
    cards.value = data.cards
  } catch (error) {
    console.error('Failed to load cards:', error)
    alert('Failed to load cards')
  } finally {
    cardsLoading.value = false
  }
}

const closeCardsModal = () => {
  showCardsModal.value = false
  selectedWord.value = ''
  cards.value = []
}

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.vocab {
  padding: 20px;
}

.vocab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  gap: 20px;
}

.vocab-header h1 {
  margin: 0;
}

.search-box {
  max-width: 400px;
  flex-shrink: 0;
}

.search-input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 16px;
  background-color: var(--input-bg);
  color: var(--text-primary);
}

.table-container {
  overflow-x: auto;
  margin-bottom: 20px;
}

.vocab-table {
  width: 100%;
  border-collapse: collapse;
  background: var(--card-bg);
  border-radius: 8px;
  overflow: hidden;
}

.vocab-table thead {
  background-color: var(--table-header-bg, rgba(0, 0, 0, 0.1));
}

.vocab-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid var(--table-border, rgba(0, 0, 0, 0.1));
  color: var(--text-primary);
}

.vocab-table th.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px;
}

.vocab-table th.sortable:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.05));
}

.sort-indicator {
  position: absolute;
  right: 8px;
  font-size: 14px;
  color: var(--color-primary);
}

.vocab-table td {
  padding: 12px;
  border-bottom: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  color: var(--text-primary);
}

.vocab-table tbody tr:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.05));
}

.word-cell {
  font-weight: 500;
}

.link-button {
  background: none;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
  font-size: inherit;
}

.link-button:hover {
  color: var(--color-primary-hover);
}

.btn-sm {
  padding: 6px 12px;
  font-size: 14px;
}

.mastery-badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}

.mastery-mastered {
  background-color: var(--color-success, #10b981);
  color: white;
}

.mastery-learning {
  background-color: var(--color-warning, #f59e0b);
  color: white;
}

.mastery-new {
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  margin-top: 20px;
}

.page-info {
  color: var(--text-secondary);
}

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
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  color: var(--text-primary);
}

.modal-large {
  max-width: 900px;
  width: 95%;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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
}

.btn-close:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.1));
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.cards-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.card-item {
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  border-radius: 8px;
  padding: 16px;
  background: var(--card-bg);
}

.card-header {
  margin-bottom: 12px;
}

.card-header h4 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
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

.sense-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  background-color: var(--color-secondary, #6b7280);
  color: white;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-row {
  display: flex;
  gap: 8px;
}

.card-row .label {
  font-weight: 600;
  min-width: 120px;
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.stat-item {
  display: flex;
  gap: 8px;
}

.stat-label {
  font-weight: 600;
  color: var(--text-secondary);
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

.no-cards {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.transcription {
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif;
  font-style: italic;
  letter-spacing: 0.5px;
}
</style>
