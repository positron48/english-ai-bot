<template>
  <div class="admin">
    <h1>Admin Panel</h1>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else>
      <div class="card">
        <div class="circuit-breaker-header">
          <h2>Circuit Breaker</h2>
          <button v-if="circuitBreaker" @click="resetCircuitBreaker" class="btn btn-primary">Reset</button>
        </div>
        <div v-if="circuitBreaker" class="circuit-breaker-content">
          <div class="circuit-breaker-info">
            <div class="info-row">
              <span class="info-label">State:</span>
              <span class="info-value" :class="{ 'state-open': circuitBreaker.state === 'open', 'state-closed': circuitBreaker.state === 'closed' }">
                {{ circuitBreaker.state || 'closed' }}
              </span>
            </div>
            <div v-if="circuitBreaker.failures !== undefined" class="info-row">
              <span class="info-label">Failures:</span>
              <span class="info-value">{{ circuitBreaker.failures }}</span>
            </div>
            <div v-if="circuitBreaker.last_failure_at" class="info-row">
              <span class="info-label">Last failure at:</span>
              <span class="info-value">{{ formatDate(circuitBreaker.last_failure_at) }}</span>
            </div>
            <div v-if="circuitBreaker.last_failure" class="info-row">
              <span class="info-label">Last failure:</span>
              <span class="info-value">{{ circuitBreaker.last_failure }}</span>
            </div>
            <div v-if="circuitBreaker.last_reset_at" class="info-row">
              <span class="info-label">Last reset at:</span>
              <span class="info-value">{{ formatDate(circuitBreaker.last_reset_at) }}</span>
            </div>
          </div>
        </div>
        <p v-else>No circuit breaker data</p>
      </div>

      <div class="card">
        <h2>Training Cards Management</h2>
        <div class="admin-actions">
          <input
            v-model="wordToManage"
            type="text"
            placeholder="Enter word"
            class="admin-input"
          />
          <button @click="getTrainingData" class="btn btn-primary">Get Training Data</button>
          <button @click="deleteTrainingWord" class="btn btn-danger">Delete Word</button>
          <button @click="deleteAllTraining" class="btn btn-danger">Delete All</button>
        </div>
        
        <div v-if="trainingData" class="training-data">
          <h3>Training Data for "{{ wordToManage }}"</h3>
          <pre>{{ JSON.stringify(trainingData, null, 2) }}</pre>
        </div>
      </div>

      <div class="card">
        <h2>Words Management</h2>
        <div class="words-filters">
          <div class="search-box">
            <input
              type="text"
              v-model="wordsSearchQuery"
              @input="onWordsSearchInput"
              placeholder="Search words..."
              class="search-input"
            />
          </div>
          <select v-model="wordsFilterUser" class="admin-select" @change="onFilterChange">
            <option :value="null">All users</option>
            <option v-for="user in users" :key="user.id" :value="user.id">
              {{ user.telegram_username || `User #${user.telegram_id}` }} (ID: {{ user.id }})
            </option>
          </select>
          <label class="checkbox-label">
            <input type="checkbox" v-model="wordsOnlyErrors" @change="onFilterChange" />
            Only with errors
          </label>
          <button @click="loadWords" class="btn btn-primary">Refresh</button>
        </div>

        <div v-if="wordsLoading" class="loading">Loading words...</div>
        <div v-else-if="words.length === 0" class="empty-message">
          <p v-if="wordsSearchQuery">No words found matching "{{ wordsSearchQuery }}".</p>
          <p v-else>No words found</p>
        </div>
        <div v-else class="words-table-container">
          <table class="words-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Word</th>
                <th>Definition</th>
                <th>Has Cards</th>
                <th>Processed At</th>
                <th>Error</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="word in words" :key="word.ID">
                <tr :class="{ 'has-error': word.ProcessingError }">
                  <td>{{ word.ID }}</td>
                  <td><strong>{{ word.Word }}</strong></td>
                  <td>
                    <span v-if="!word.editing">{{ word.Definition }}</span>
                  <div v-else class="edit-form">
                    <textarea v-model="word.editDefinition" rows="10" class="edit-textarea"></textarea>
                      <div class="edit-actions">
                        <button @click="saveWordDefinition(word)" class="btn btn-sm btn-primary">Save</button>
                        <button @click="cancelEditWord(word)" class="btn btn-sm">Cancel</button>
                      </div>
                    </div>
                  </td>
                  <td>
                    <span v-if="!word.HasTrainingCards" :class="{ 'badge': true, 'badge-secondary': true }">
                      No
                    </span>
                    <a 
                      v-else
                      @click.prevent="toggleWordCards(word)"
                      :class="{ 'card-link': true, 'card-link-active': word.showingCards === true }"
                      :style="{ cursor: word.cardsLoading ? 'wait' : 'pointer', opacity: word.cardsLoading ? 0.6 : 1 }"
                    >
                      {{ (word.showingCards === true) ? 'Hide' : 'Yes' }}
                    </a>
                  </td>
                  <td>{{ word.ProcessedAt ? formatDate(word.ProcessedAt) : '—' }}</td>
                  <td>
                    <span v-if="word.ProcessingError" class="error-text">{{ word.ProcessingError }}</span>
                    <span v-else>—</span>
                  </td>
                  <td>
                    <div class="action-buttons">
                      <button 
                        v-if="!word.editing"
                        @click="startEditWord(word)" 
                        class="btn btn-sm btn-primary"
                      >
                        Edit
                      </button>
                      <button 
                        v-if="word.ProcessingError"
                        @click="resetWordError(word)" 
                        class="btn btn-sm btn-warning"
                      >
                        Reset Error
                      </button>
                      <button 
                        @click="deleteWord(word)" 
                        class="btn btn-sm btn-danger"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
                <!-- Cards row -->
                <tr v-if="word && word.showingCards === true && word.HasTrainingCards" class="cards-row">
                <td colspan="7">
                  <div v-if="word.cardsLoading" class="cards-loading">Loading cards...</div>
                  <div v-else-if="word.cards && word.cards.length > 0" class="cards-container">
                    <h4>Training Cards for "{{ word.Word }}"</h4>
                    <div v-for="card in word.cards" :key="card.id" class="card-item">
                      <div class="card-header">
                        <div>
                          <strong>Card #{{ card.sense_index + 1 }}</strong>
                          <span class="card-id">ID: {{ card.id }}</span>
                        </div>
                        <div class="card-actions">
                          <button 
                            @click="editTrainingCard(card, word)" 
                            class="btn btn-sm btn-primary"
                          >
                            Edit
                          </button>
                          <button 
                            @click="confirmDeleteTrainingCard(card, word)" 
                            class="btn btn-sm btn-danger"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                      <div class="card-content">
                        <div class="card-field">
                          <span class="field-label">Word EN:</span>
                          <span class="field-value">{{ card.word_en }}</span>
                          <span v-if="card.transcription" class="transcription">{{ card.transcription }}</span>
                        </div>
                        <div class="card-field">
                          <span class="field-label">Word RU:</span>
                          <span class="field-value">{{ card.word_ru }}</span>
                        </div>
                        <div class="card-field">
                          <span class="field-label">Meaning EN:</span>
                          <span class="field-value">{{ card.meaning_en }}</span>
                        </div>
                        <div v-if="card.example_en" class="card-field">
                          <span class="field-label">Example EN:</span>
                          <span class="field-value">{{ card.example_en }}</span>
                        </div>
                        <div v-if="card.example_ru" class="card-field">
                          <span class="field-label">Example RU:</span>
                          <span class="field-value">{{ card.example_ru }}</span>
                        </div>
                        <div v-if="card.hint" class="card-field">
                          <span class="field-label">Hint:</span>
                          <span class="field-value">{{ card.hint }}</span>
                        </div>
                        <div v-if="card.distractors_ru" class="card-field">
                          <span class="field-label">Distractors RU:</span>
                          <span class="field-value">{{ parseJSONArray(card.distractors_ru)?.join(', ') || card.distractors_ru }}</span>
                        </div>
                        <div v-if="card.distractors_en" class="card-field">
                          <span class="field-label">Distractors EN:</span>
                          <span class="field-value">{{ parseJSONArray(card.distractors_en)?.join(', ') || card.distractors_en }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div v-else class="cards-empty">No cards found</div>
                </td>
              </tr>
              </template>
            </tbody>
          </table>
        </div>
        <div class="pagination" v-if="wordsPagination.total_pages > 1">
          <button 
            @click="goToWordsPage(wordsPagination.page - 1)" 
            :disabled="wordsPagination.page <= 1"
            class="btn btn-secondary"
          >
            Previous
          </button>
          <span class="page-info">
            Page {{ wordsPagination.page }} of {{ wordsPagination.total_pages }} 
            ({{ wordsPagination.total }} total)
          </span>
          <button 
            @click="goToWordsPage(wordsPagination.page + 1)" 
            :disabled="wordsPagination.page >= wordsPagination.total_pages"
            class="btn btn-secondary"
          >
            Next
          </button>
        </div>
      </div>

      <!-- Edit Training Card Modal -->
      <div v-if="showEditCardModal && cardToEdit" class="modal" @click.self="closeEditCardModal">
        <div class="modal-content modal-large">
          <div class="modal-header">
            <h3>Edit Training Card (Sense {{ cardToEdit.sense_index }})</h3>
            <button @click="closeEditCardModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveTrainingCard" class="edit-form">
              <div class="form-group">
                <label>Word RU:</label>
                <input v-model="editCardForm.word_ru" type="text" required class="form-input" />
              </div>
              <div class="form-group">
                <label>Meaning EN:</label>
                <textarea v-model="editCardForm.meaning_en" required class="form-textarea" rows="3"></textarea>
              </div>
              <div class="form-group">
                <label>Example EN:</label>
                <textarea v-model="editCardForm.example_en" class="form-textarea" rows="2"></textarea>
              </div>
              <div class="form-group">
                <label>Example RU:</label>
                <textarea v-model="editCardForm.example_ru" class="form-textarea" rows="2"></textarea>
              </div>
              <div class="form-group">
                <label>Transcription:</label>
                <input v-model="editCardForm.transcription" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label>Distractors RU:</label>
                <div class="distractors-list">
                  <input v-model="editCardForm.distractors_ru[0]" type="text" class="form-input" placeholder="Option 1" />
                  <input v-model="editCardForm.distractors_ru[1]" type="text" class="form-input" placeholder="Option 2" />
                  <input v-model="editCardForm.distractors_ru[2]" type="text" class="form-input" placeholder="Option 3" />
                </div>
              </div>
              <div class="form-group">
                <label>Distractors EN:</label>
                <div class="distractors-list">
                  <input v-model="editCardForm.distractors_en[0]" type="text" class="form-input" placeholder="Option 1" />
                  <input v-model="editCardForm.distractors_en[1]" type="text" class="form-input" placeholder="Option 2" />
                  <input v-model="editCardForm.distractors_en[2]" type="text" class="form-input" placeholder="Option 3" />
                </div>
              </div>
              <div class="form-group">
                <label>Hint:</label>
                <input v-model="editCardForm.hint" type="text" class="form-input" />
              </div>
              <div class="modal-actions">
                <button type="submit" class="btn btn-primary">Save</button>
                <button type="button" @click="closeEditCardModal" class="btn btn-secondary">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Delete Training Card Confirmation Modal -->
      <div v-if="showDeleteCardConfirm && cardToDelete" class="modal" @click.self="closeDeleteCardConfirm">
        <div class="modal-content">
          <h3>Confirm Delete</h3>
          <p>Are you sure you want to delete training card #{{ cardToDelete.sense_index + 1 }} (ID: {{ cardToDelete.id }})?</p>
          <p class="warning-text">This will delete the training card and all associated user cards. This action cannot be undone.</p>
          <div class="modal-actions">
            <button @click="deleteTrainingCard" class="btn btn-danger">Delete</button>
            <button @click="closeDeleteCardConfirm" class="btn btn-secondary">Cancel</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface CircuitBreaker {
  state: string
  failures?: number
  last_failure?: string
}

interface User {
  id: number
  telegram_id: number
  telegram_username?: string
}

interface TrainingCard {
  id: number
  word_card_id: number
  word_en: string
  transcription?: string
  sense_index: number
  word_ru: string
  meaning_en: string
  example_en?: string
  example_ru?: string
  distractors_ru?: string
  distractors_en?: string
  hint?: string
  created_at?: string
}

interface WordCard {
  ID: number
  Word: string
  Definition: string
  ProcessedAt?: string | null
  ProcessingError?: string | null
  HasTrainingCards: boolean
  RequestingUsers?: number[]
  editing?: boolean
  editDefinition?: string
  showingCards?: boolean
  cards?: TrainingCard[]
  cardsLoading?: boolean
}

const loading = ref(true)
const circuitBreaker = ref<CircuitBreaker | null>(null)
const wordToManage = ref('')
const trainingData = ref<any>(null)

// Words management
const words = ref<WordCard[]>([])
const wordsLoading = ref(false)
const wordsFilterUser = ref<number | null>(null)
const wordsOnlyErrors = ref(false)
const wordsSearchQuery = ref('')
const wordsSearchTimeout = ref<number | null>(null)
const wordsPagination = ref({
  page: 1,
  limit: 50,
  total: 0,
  total_pages: 0
})
const users = ref<User[]>([])

// Training card editing
const showEditCardModal = ref(false)
const showDeleteCardConfirm = ref(false)
const cardToEdit = ref<TrainingCard | null>(null)
const cardToDelete = ref<{ card: TrainingCard; word: WordCard } | null>(null)
const editCardForm = ref({
  word_ru: '',
  meaning_en: '',
  example_en: '',
  example_ru: '',
  transcription: '',
  distractors_ru: ['', '', ''],
  distractors_en: ['', '', ''],
  hint: ''
})

onMounted(async () => {
  await loadAdminData()
  await loadUsers()
  await loadWords()
})

const loadAdminData = async () => {
  loading.value = true
  try {
    const data: { circuit_breaker: CircuitBreaker } = await apiClient.request('/app/admin')
    circuitBreaker.value = data.circuit_breaker
  } catch (error) {
    console.error('Failed to load admin data:', error)
  } finally {
    loading.value = false
  }
}

const resetCircuitBreaker = async () => {
  try {
    await apiClient.request('/app/admin/circuit/reset', { method: 'POST' })
    await loadAdminData()
    alert('Circuit breaker reset successfully')
  } catch (error) {
    console.error('Failed to reset circuit breaker:', error)
    alert('Failed to reset circuit breaker')
  }
}

const getTrainingData = async () => {
  if (!wordToManage.value.trim()) {
    alert('Please enter a word')
    return
  }

  try {
    const data = await apiClient.request(`/app/admin/training/${wordToManage.value.trim()}`)
    trainingData.value = data
  } catch (error) {
    console.error('Failed to get training data:', error)
    alert('Failed to get training data')
  }
}

const deleteTrainingWord = async () => {
  if (!wordToManage.value.trim()) {
    alert('Please enter a word')
    return
  }

  if (!confirm(`Are you sure you want to delete all training cards for "${wordToManage.value}"?`)) {
    return
  }

  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/app/admin/training/${wordToManage.value.trim()}/delete`, formData)
    trainingData.value = null
    alert('Training cards deleted successfully')
  } catch (error) {
    console.error('Failed to delete training word:', error)
    alert('Failed to delete training word')
  }
}

const deleteAllTraining = async () => {
  if (!confirm('Are you sure you want to delete ALL training cards? This cannot be undone!')) {
    return
  }

  try {
    const formData = new FormData()
    await apiClient.requestFormData('/app/admin/training/delete_all', formData)
    trainingData.value = null
    alert('All training cards deleted successfully')
  } catch (error) {
    console.error('Failed to delete all training:', error)
    alert('Failed to delete all training')
  }
}

const loadUsers = async () => {
  try {
    const data: { users: User[] } = await apiClient.request('/app/admin/users')
    users.value = data.users
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const loadWords = async () => {
  wordsLoading.value = true
  try {
    const params = new URLSearchParams()
    if (wordsFilterUser.value !== null) {
      params.append('user_id', wordsFilterUser.value.toString())
    }
    if (wordsOnlyErrors.value) {
      params.append('only_errors', '1')
    }
    if (wordsSearchQuery.value) {
      params.append('search', wordsSearchQuery.value)
    }
    const offset = (wordsPagination.value.page - 1) * wordsPagination.value.limit
    params.append('limit', wordsPagination.value.limit.toString())
    params.append('offset', offset.toString())

    const data: { words: WordCard[]; pagination: { page: number; limit: number; total: number; total_pages: number } } = await apiClient.request(`/app/admin/words?${params.toString()}`)
    words.value = data.words.map(w => ({ 
      ...w, 
      editing: false, 
      showingCards: false,
      cardsLoading: false,
      cards: undefined
    }))
    if (data.pagination) {
      wordsPagination.value = data.pagination
    }
  } catch (error) {
    console.error('Failed to load words:', error)
    alert('Failed to load words')
  } finally {
    wordsLoading.value = false
  }
}

const onWordsSearchInput = () => {
  if (wordsSearchTimeout.value) {
    clearTimeout(wordsSearchTimeout.value)
  }
  wordsSearchTimeout.value = window.setTimeout(() => {
    wordsPagination.value.page = 1
    loadWords()
  }, 500)
}

const goToWordsPage = (page: number) => {
  if (page >= 1 && page <= wordsPagination.value.total_pages) {
    wordsPagination.value.page = page
    loadWords()
  }
}

const onFilterChange = () => {
  wordsPagination.value.page = 1
  loadWords()
}

const toggleWordCards = async (word: WordCard) => {
  // Ensure showingCards is initialized
  if (word.showingCards === undefined) {
    word.showingCards = false
  }
  
  if (word.showingCards === true) {
    // Hide cards
    word.showingCards = false
    word.cards = undefined
  } else {
    // Show cards - load them
    word.showingCards = true
    word.cardsLoading = true
    try {
      const data: { word_en: string; cards: TrainingCard[] } = await apiClient.request(`/app/admin/training/${word.Word}`)
      word.cards = data.cards || []
    } catch (error) {
      console.error('Failed to load training cards:', error)
      alert('Failed to load training cards')
      word.cards = []
    } finally {
      word.cardsLoading = false
    }
  }
}

const parseJSONArray = (jsonStr: string | undefined): string[] | null => {
  if (!jsonStr) return null
  try {
    const parsed = JSON.parse(jsonStr)
    if (Array.isArray(parsed)) {
      return parsed
    }
    return null
  } catch {
    return null
  }
}

const editTrainingCard = async (card: TrainingCard, word: WordCard) => {
  cardToEdit.value = card
  
  // Parse distractors from JSON arrays
  let distractorsRU: string[] = ['', '', '']
  let distractorsEN: string[] = ['', '', '']
  
  if (card.distractors_ru) {
    try {
      const parsed = parseJSONArray(card.distractors_ru)
      if (parsed && Array.isArray(parsed)) {
        distractorsRU = [...parsed, '', '', ''].slice(0, 3)
      }
    } catch (e) {
      console.error('Failed to parse distractors_ru:', e)
    }
  }
  
  if (card.distractors_en) {
    try {
      const parsed = parseJSONArray(card.distractors_en)
      if (parsed && Array.isArray(parsed)) {
        distractorsEN = [...parsed, '', '', ''].slice(0, 3)
      }
    } catch (e) {
      console.error('Failed to parse distractors_en:', e)
    }
  }
  
  editCardForm.value = {
    word_ru: card.word_ru || '',
    meaning_en: card.meaning_en || '',
    example_en: card.example_en || '',
    example_ru: card.example_ru || '',
    transcription: card.transcription || '',
    distractors_ru: distractorsRU,
    distractors_en: distractorsEN,
    hint: card.hint || ''
  }
  
  showEditCardModal.value = true
}

const confirmDeleteTrainingCard = (card: TrainingCard, word: WordCard) => {
  cardToDelete.value = { card, word }
  showDeleteCardConfirm.value = true
}

const reloadWordCards = async (word: WordCard) => {
  if (!word.showingCards) return
  
  word.cardsLoading = true
  try {
    const data: { word_en: string; cards: TrainingCard[] } = await apiClient.request(`/app/admin/training/${word.Word}`)
    word.cards = data.cards || []
  } catch (error) {
    console.error('Failed to reload training cards:', error)
    word.cards = []
  } finally {
    word.cardsLoading = false
  }
}

const saveTrainingCard = async () => {
  if (!cardToEdit.value) return
  
  try {
    // Convert distractors arrays to JSON
    const distractorsRU = JSON.stringify(editCardForm.value.distractors_ru.filter(v => v.trim() !== ''))
    const distractorsEN = JSON.stringify(editCardForm.value.distractors_en.filter(v => v.trim() !== ''))
    
    const params = new URLSearchParams()
    params.append('word_ru', editCardForm.value.word_ru || '')
    params.append('meaning_en', editCardForm.value.meaning_en || '')
    params.append('example_en', editCardForm.value.example_en || '')
    params.append('example_ru', editCardForm.value.example_ru || '')
    params.append('transcription', editCardForm.value.transcription || '')
    params.append('distractors_ru', distractorsRU)
    params.append('distractors_en', distractorsEN)
    params.append('hint', editCardForm.value.hint || '')
    
    await apiClient.request(`/app/admin/training/card/${cardToEdit.value.id}`, { 
      method: 'PUT',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString()
    })
    
    showEditCardModal.value = false
    const word = words.value.find(w => w.showingCards === true && w.cards?.some(c => c.id === cardToEdit.value!.id))
    cardToEdit.value = null
    
    // Reload cards for the word
    if (word) {
      await reloadWordCards(word)
    }
    
    alert('Card updated successfully')
  } catch (error) {
    console.error('Failed to save card:', error)
    alert('Failed to save card')
  }
}

const deleteTrainingCard = async () => {
  if (!cardToDelete.value) return
  
  try {
    await apiClient.request(`/app/admin/training/card/${cardToDelete.value.card.id}`, { method: 'DELETE' })
    showDeleteCardConfirm.value = false
    
    const word = cardToDelete.value.word
    const cardId = cardToDelete.value.card.id
    cardToDelete.value = null
    
    // Reload cards
    if (word.showingCards) {
      await reloadWordCards(word)
      
      // Update HasTrainingCards status if no cards left
      if (word.cards && word.cards.length === 0) {
        word.HasTrainingCards = false
        // Also update in the list
        const wordInList = words.value.find(w => w.ID === word.ID)
        if (wordInList) {
          wordInList.HasTrainingCards = false
        }
      }
    }
    
    alert('Card deleted successfully')
  } catch (error) {
    console.error('Failed to delete card:', error)
    alert('Failed to delete card')
  }
}

const closeEditCardModal = () => {
  showEditCardModal.value = false
  cardToEdit.value = null
}

const closeDeleteCardConfirm = () => {
  showDeleteCardConfirm.value = false
  cardToDelete.value = null
}

const startEditWord = (word: WordCard) => {
  word.editing = true
  word.editDefinition = word.Definition
}

const cancelEditWord = (word: WordCard) => {
  word.editing = false
  word.editDefinition = undefined
}

const saveWordDefinition = async (word: WordCard) => {
  if (!word.editDefinition || word.editDefinition.trim() === '') {
    alert('Definition cannot be empty')
    return
  }

  try {
    await apiClient.request(`/app/admin/words/${word.ID}`, {
      method: 'PUT',
      body: JSON.stringify({ definition: word.editDefinition.trim() })
    })
    word.Definition = word.editDefinition.trim()
    word.editing = false
    word.editDefinition = undefined
    alert('Definition updated successfully')
  } catch (error) {
    console.error('Failed to update word definition:', error)
    alert('Failed to update definition')
  }
}

const resetWordError = async (word: WordCard) => {
  if (!confirm(`Reset error for word "${word.Word}"? This will allow the worker to process it again.`)) {
    return
  }

  try {
    await apiClient.request(`/app/admin/words/${word.ID}/reset`, { method: 'POST' })
    word.ProcessingError = null
    word.ProcessedAt = null
    alert('Error reset successfully. Word will be processed again by the worker.')
  } catch (error) {
    console.error('Failed to reset word error:', error)
    alert('Failed to reset error')
  }
}

const deleteWord = async (word: WordCard) => {
  if (!confirm(`Are you sure you want to delete word "${word.Word}"?\n\nThis will delete:\n- The word itself\n- All training cards\n- All user cards\n- All request history\n\nThis action cannot be undone!`)) {
    return
  }

  try {
    await apiClient.request(`/app/admin/words/${word.ID}`, { method: 'DELETE' })
    // Remove word from list
    const index = words.value.findIndex(w => w.ID === word.ID)
    if (index !== -1) {
      words.value.splice(index, 1)
    }
    alert('Word and all related data deleted successfully')
  } catch (error) {
    console.error('Failed to delete word:', error)
    alert('Failed to delete word')
  }
}

const formatDate = (dateStr: string | null | undefined) => {
  if (!dateStr) return '—'
  
  // Handle SQL datetime format "2006-01-02 15:04:05" (same as dashboard sessions)
  let date: Date
  if (dateStr.includes(' ')) {
    // SQL format: replace space with T for ISO format, assume local timezone
    date = new Date(dateStr.replace(' ', 'T'))
  } else {
    date = new Date(dateStr)
  }
  
  // Check if date is valid
  if (isNaN(date.getTime())) {
    return '—'
  }
  
  // Format same way as VocabView does it
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<style scoped>
.admin {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.admin h1 {
  margin-bottom: 24px;
}

.admin .card h2 {
  margin-bottom: 20px;
}

.circuit-breaker-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.circuit-breaker-header h2 {
  margin: 0;
}

.circuit-breaker-content {
  margin-top: 0;
}

.circuit-breaker-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  gap: 10px;
  align-items: baseline;
}

.info-label {
  font-weight: 500;
  color: var(--text-secondary);
  min-width: 140px;
}

.info-value {
  color: var(--text-primary);
  word-break: break-word;
}

.info-value.state-open {
  color: var(--color-danger);
  font-weight: 600;
  background: rgba(220, 53, 69, 0.1);
  padding: 4px 12px;
  border-radius: 4px;
  display: inline-block;
  text-transform: uppercase;
  font-size: 0.9em;
  letter-spacing: 0.5px;
}

.info-value.state-closed {
  color: var(--color-success);
  font-weight: 600;
  background: rgba(40, 167, 69, 0.1);
  padding: 4px 12px;
  border-radius: 4px;
  display: inline-block;
  text-transform: uppercase;
  font-size: 0.9em;
  letter-spacing: 0.5px;
}

.admin-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: stretch;
  margin-bottom: 20px;
}

.admin-input {
  flex: 1;
  min-width: 200px;
  height: 40px;
  padding: 10px;
  box-sizing: border-box;
}

.admin-actions .btn {
  height: 40px;
  padding: 10px 20px;
  box-sizing: border-box;
  white-space: nowrap;
}

.training-data {
  margin-top: 20px;
  padding: 15px;
  background: var(--bg-tertiary);
  border-radius: 4px;
  color: var(--text-primary);
}

.training-data pre {
  overflow-x: auto;
  font-size: 12px;
}

.words-filters {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
}

.search-box {
  flex: 1;
  min-width: 200px;
  max-width: 400px;
}

.search-input {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 16px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  box-sizing: border-box;
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

.admin-select {
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background: var(--input-bg);
  color: var(--text-primary);
  min-width: 200px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  cursor: pointer;
}

.words-table-container {
  overflow-x: auto;
  margin-top: 20px;
}

.words-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.words-table th,
.words-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-primary);
}

.words-table th {
  background: var(--bg-secondary);
  font-weight: 600;
  color: var(--text-primary);
  position: sticky;
  top: 0;
}

.words-table tbody tr:hover {
  background: var(--bg-secondary);
}

.words-table tbody tr.has-error {
  background: rgba(220, 53, 69, 0.05);
}

.words-table tbody tr.has-error:hover {
  background: rgba(220, 53, 69, 0.1);
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.edit-textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  background: var(--input-bg);
  color: var(--text-primary);
  font-family: inherit;
  resize: vertical;
}

.edit-actions {
  display: flex;
  gap: 8px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.badge-success {
  background: rgba(40, 167, 69, 0.1);
  color: var(--color-success);
}

.badge-secondary {
  background: rgba(108, 117, 125, 0.1);
  color: var(--text-secondary);
}

.badge-clickable {
  cursor: pointer;
  transition: opacity 0.2s;
}

.badge-clickable:hover:not(:disabled) {
  opacity: 0.8;
}

.badge-clickable:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.card-link {
  color: var(--color-success);
  text-decoration: none;
  border-bottom: 1px dashed var(--color-success);
  transition: all 0.2s;
  font-weight: 500;
  padding: 2px 0;
}

.card-link:hover {
  color: var(--color-success);
  border-bottom-style: solid;
  opacity: 0.8;
}

.card-link-active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.card-link-active:hover {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.error-text {
  color: var(--color-danger);
  font-size: 12px;
  max-width: 300px;
  word-break: break-word;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.btn-warning {
  background: var(--color-warning, #f59e0b);
  color: white;
  border: none;
}

.btn-warning:hover {
  background: var(--color-warning, #d97706);
  opacity: 0.9;
}

.empty-message {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

.cards-row {
  background: var(--bg-secondary);
}

.cards-row td {
  padding: 20px;
  border-top: 2px solid var(--border-primary);
}

.cards-loading {
  padding: 20px;
  text-align: center;
  color: var(--text-secondary);
}

.cards-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.cards-container h4 {
  margin: 0 0 15px 0;
  color: var(--text-primary);
  font-size: 16px;
}

.card-item {
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 15px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border-primary);
}

.card-header strong {
  color: var(--text-primary);
  font-size: 14px;
}

.card-id {
  color: var(--text-secondary);
  font-size: 12px;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.card-field {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: baseline;
}

.field-label {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 12px;
  min-width: 120px;
}

.field-value {
  color: var(--text-primary);
  flex: 1;
  word-break: break-word;
}

.transcription {
  color: var(--text-secondary);
  font-style: italic;
  font-size: 12px;
  margin-left: 8px;
}

.cards-empty {
  padding: 20px;
  text-align: center;
  color: var(--text-secondary);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.card-actions {
  display: flex;
  gap: 8px;
}

/* Modal styles */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
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

.modal-header h3 {
  margin: 0;
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
  background-color: rgba(0, 0, 0, 0.1);
}

.modal-body {
  margin-top: 20px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  justify-content: flex-end;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.form-group label {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 14px;
}

.form-input,
.form-textarea {
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  font-family: inherit;
  width: 100%;
}

.form-textarea {
  resize: vertical;
  min-height: 60px;
}

.distractors-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.warning-text {
  color: var(--color-danger);
  font-size: 13px;
  margin-top: 8px;
}

@media (min-width: 768px) {
  .distractors-list {
    flex-direction: row;
    gap: 12px;
  }
  
  .distractors-list input {
    flex: 1;
    min-width: 0;
  }
}
</style>

