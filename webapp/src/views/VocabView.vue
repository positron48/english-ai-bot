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
          <div v-for="senseGroup in groupedCards" :key="senseGroup.sense_index" class="sense-group">
            <div class="sense-header">
              <h4>
                Sense {{ senseGroup.sense_index }}
                <span class="word-en">{{ selectedWord }}</span>
                <span class="word-ru">{{ senseGroup.word_ru }}</span>
              </h4>
              <div v-if="isAdmin" class="sense-actions">
                <button 
                  @click="editCard(senseGroup)" 
                  class="btn btn-sm btn-primary"
                  title="Edit card"
                >
                  Edit
                </button>
                <button 
                  @click="confirmDeleteCard(senseGroup)" 
                  class="btn btn-sm btn-danger"
                  title="Delete card"
                >
                  Delete
                </button>
              </div>
            </div>
            
            <!-- Sense information (shown once per sense) -->
            <div class="sense-info">
              <div class="card-row" v-if="senseGroup.pos">
                <span class="label">POS:</span>
                <span>{{ senseGroup.pos }}</span>
              </div>
              <div class="card-row" v-if="senseGroup.meaning_en">
                <span class="label">Meaning:</span>
                <span>{{ senseGroup.meaning_en }}</span>
              </div>
              <div class="card-row" v-if="senseGroup.example_en">
                <span class="label">Example EN:</span>
                <span>{{ senseGroup.example_en }}</span>
              </div>
              <div class="card-row" v-if="senseGroup.example_ru">
                <span class="label">Example RU:</span>
                <span>{{ senseGroup.example_ru }}</span>
              </div>
              <div class="card-row" v-if="senseGroup.transcription">
                <span class="label">Transcription:</span>
                <span class="transcription">{{ senseGroup.transcription }}</span>
              </div>
            </div>
            
            <!-- Directions with training stats -->
            <div class="directions-list">
              <div v-for="directionCard in senseGroup.directions" :key="directionCard.direction" class="direction-item">
                <div class="direction-header">
                  <span class="direction-badge" :class="`direction-${directionCard.direction}`">
                    {{ directionCard.direction === 'ru_en' ? 'RU→EN' : 'EN→RU' }}
                  </span>
                </div>
                <div class="card-stats">
                  <div class="stat-item">
                    <span class="stat-label">State:</span>
                    <span :class="['state-badge', `state-${directionCard.state}`]">{{ directionCard.state }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Reps:</span>
                    <span>{{ directionCard.reps }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Reviews:</span>
                    <span>{{ directionCard.review_count }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">EF:</span>
                    <span>{{ directionCard.ef.toFixed(2) }}</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Interval:</span>
                    <span>{{ directionCard.interval_days }} days</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Lapses:</span>
                    <span>{{ directionCard.lapse_count }}</span>
                  </div>
                  <div class="stat-item" v-if="directionCard.next_due_at">
                    <span class="stat-label">Next Due:</span>
                    <span>{{ formatDate(directionCard.next_due_at) }}</span>
                  </div>
                  <div class="stat-item" v-if="directionCard.last_review_at">
                    <span class="stat-label">Last Review:</span>
                    <span>{{ formatDate(directionCard.last_review_at) }}</span>
                  </div>
                  <div class="stat-item" v-if="directionCard.last_quality !== null">
                    <span class="stat-label">Last Quality:</span>
                    <span>{{ directionCard.last_quality }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit Card Modal -->
    <div v-if="showEditCardModal && cardToEdit" class="modal" @click.self="closeEditCardModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Edit Training Card (Sense {{ cardToEdit.sense_index }})</h3>
          <button @click="closeEditCardModal" class="btn-close">&times;</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveCard" class="edit-form">
            <div class="form-group">
              <label>Word RU:</label>
              <input v-model="editForm.word_ru" type="text" required class="form-input" />
            </div>
            <div class="form-group">
              <label>Meaning EN:</label>
              <textarea v-model="editForm.meaning_en" required class="form-textarea" rows="3"></textarea>
            </div>
            <div class="form-group">
              <label>Example EN:</label>
              <textarea v-model="editForm.example_en" class="form-textarea" rows="2"></textarea>
            </div>
            <div class="form-group">
              <label>Example RU:</label>
              <textarea v-model="editForm.example_ru" class="form-textarea" rows="2"></textarea>
            </div>
            <div class="form-group">
              <label>Transcription:</label>
              <input v-model="editForm.transcription" type="text" class="form-input" />
            </div>
            <div class="form-group">
              <label>Distractors RU:</label>
              <div class="distractors-list">
                <input v-model="editForm.distractors_ru[0]" type="text" class="form-input" placeholder="Option 1" />
                <input v-model="editForm.distractors_ru[1]" type="text" class="form-input" placeholder="Option 2" />
                <input v-model="editForm.distractors_ru[2]" type="text" class="form-input" placeholder="Option 3" />
              </div>
            </div>
            <div class="form-group">
              <label>Distractors EN:</label>
              <div class="distractors-list">
                <input v-model="editForm.distractors_en[0]" type="text" class="form-input" placeholder="Option 1" />
                <input v-model="editForm.distractors_en[1]" type="text" class="form-input" placeholder="Option 2" />
                <input v-model="editForm.distractors_en[2]" type="text" class="form-input" placeholder="Option 3" />
              </div>
            </div>
            <div class="form-group">
              <label>Hint:</label>
              <input v-model="editForm.hint" type="text" class="form-input" />
            </div>
            <div class="modal-actions">
              <button type="submit" class="btn btn-primary">Save</button>
              <button type="button" @click="closeEditCardModal" class="btn btn-secondary">Cancel</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Delete Card Confirmation Modal -->
    <div v-if="showDeleteCardConfirm && cardToDelete" class="modal" @click.self="closeDeleteCardConfirm">
      <div class="modal-content">
        <h3>Confirm Delete</h3>
        <p>Are you sure you want to delete training card for "{{ cardToDelete.word_ru }}" (Sense {{ cardToDelete.sense_index }})?</p>
        <p class="warning-text">This will delete the training card and all associated user cards. This action cannot be undone.</p>
        <div class="modal-actions">
          <button @click="deleteCard" class="btn btn-danger">Delete</button>
          <button @click="closeDeleteCardConfirm" class="btn btn-secondary">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '../api/client'
import { useAuth } from '../composables/useAuth'

const { isAdmin } = useAuth()

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
  pos?: string
  review_count: number
}

interface TrainingCard {
  id: number
  word_card_id: number
  word_en: string
  transcription: string
  sense_index: number
  word_ru: string
  meaning_en: string
  example_en: string
  example_ru: string
  distractors_ru: string
  distractors_en: string
  hint: string
}

const words = ref<VocabWord[]>([])
const loading = ref(true)
const searchQuery = ref('')
const searchTimeout = ref<number | null>(null)
const sortField = ref<string>('word_en')
const sortOrder = ref<'asc' | 'desc'>('asc')
const pagination = ref<Pagination>({
  page: 1,
  limit: 25,
  total: 0,
  total_pages: 0
})

const showDeleteConfirm = ref(false)
const wordToDelete = ref('')

const showCardsModal = ref(false)
const selectedWord = ref('')
const cards = ref<CardDetail[]>([])
const cardsLoading = ref(false)

const showEditCardModal = ref(false)
const showDeleteCardConfirm = ref(false)
const cardToEdit = ref<{ training_card_id: number; sense_index: number; word_ru: string; meaning_en: string; example_en: string; example_ru: string; transcription: string } | null>(null)
const cardToDelete = ref<{ training_card_id: number; sense_index: number; word_ru: string } | null>(null)
const editForm = ref({
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
      limit: 25,
      total: 0,
      total_pages: 0
    }
  } catch (error) {
    console.error('Failed to load vocabulary:', error)
    words.value = []
    pagination.value = {
      page: 1,
      limit: 25,
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

// Group cards by sense_index
interface SenseGroup {
  sense_index: number
  word_ru: string
  meaning_en: string
  example_en: string
  example_ru: string
  transcription: string
  pos?: string
  directions: CardDetail[]
}

const groupedCards = computed((): SenseGroup[] => {
  const groups = new Map<number, SenseGroup>()
  
  for (const card of cards.value) {
    if (!groups.has(card.sense_index)) {
      groups.set(card.sense_index, {
        sense_index: card.sense_index,
        word_ru: card.word_ru,
        meaning_en: card.meaning_en,
        example_en: card.example_en,
        example_ru: card.example_ru,
        transcription: card.transcription,
        pos: card.pos,
        directions: []
      })
    }
    
    const group = groups.get(card.sense_index)!
    group.directions.push(card)
  }
  
  // Sort by sense_index and sort directions within each group
  return Array.from(groups.values())
    .sort((a, b) => a.sense_index - b.sense_index)
    .map(group => ({
      ...group,
      directions: group.directions.sort((a, b) => {
        // Sort directions: EN→RU first, then RU→EN
        if (a.direction === 'en_ru' && b.direction === 'ru_en') return -1
        if (a.direction === 'ru_en' && b.direction === 'en_ru') return 1
        return 0
      })
    }))
})

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const editCard = async (senseGroup: SenseGroup) => {
  // Get training_card_id from first direction card
  if (senseGroup.directions.length === 0) return
  
  const trainingCardId = senseGroup.directions[0].training_card_id
  
  cardToEdit.value = {
    training_card_id: trainingCardId,
    sense_index: senseGroup.sense_index,
    word_ru: senseGroup.word_ru,
    meaning_en: senseGroup.meaning_en,
    example_en: senseGroup.example_en || '',
    example_ru: senseGroup.example_ru || '',
    transcription: senseGroup.transcription || ''
  }
  
  // Load full card data to get distractors and hint
  try {
    const data: { word_en: string, cards: TrainingCard[] } = await apiClient.request(`/app/admin/training/${selectedWord.value}`)
    const card = data.cards.find(c => c.id === trainingCardId)
    
    // Parse distractors from JSON arrays
    let distractorsRU: string[] = ['', '', '']
    let distractorsEN: string[] = ['', '', '']
    
    if (card?.distractors_ru) {
      try {
        // Handle both string and already parsed array
        let parsed: any
        if (typeof card.distractors_ru === 'string') {
          parsed = JSON.parse(card.distractors_ru)
        } else {
          parsed = card.distractors_ru
        }
        if (Array.isArray(parsed)) {
          distractorsRU = [...parsed, '', '', ''].slice(0, 3)
        }
      } catch (e) {
        console.error('Failed to parse distractors_ru:', e, 'Value:', card.distractors_ru)
      }
    }
    
    if (card?.distractors_en) {
      try {
        // Handle both string and already parsed array
        let parsed: any
        if (typeof card.distractors_en === 'string') {
          parsed = JSON.parse(card.distractors_en)
        } else {
          parsed = card.distractors_en
        }
        if (Array.isArray(parsed)) {
          distractorsEN = [...parsed, '', '', ''].slice(0, 3)
        }
      } catch (e) {
        console.error('Failed to parse distractors_en:', e, 'Value:', card.distractors_en)
      }
    }
    
    editForm.value = {
      word_ru: senseGroup.word_ru,
      meaning_en: senseGroup.meaning_en,
      example_en: senseGroup.example_en || '',
      example_ru: senseGroup.example_ru || '',
      transcription: senseGroup.transcription || '',
      distractors_ru: distractorsRU,
      distractors_en: distractorsEN,
      hint: card?.hint || ''
    }
  } catch (error) {
    console.error('Failed to load card details:', error)
    // Use basic data if loading fails
    editForm.value = {
      word_ru: senseGroup.word_ru,
      meaning_en: senseGroup.meaning_en,
      example_en: senseGroup.example_en || '',
      example_ru: senseGroup.example_ru || '',
      transcription: senseGroup.transcription || '',
      distractors_ru: ['', '', ''],
      distractors_en: ['', '', ''],
      hint: ''
    }
  }
  
  showEditCardModal.value = true
}

const confirmDeleteCard = (senseGroup: SenseGroup) => {
  if (senseGroup.directions.length === 0) return
  
  cardToDelete.value = {
    training_card_id: senseGroup.directions[0].training_card_id,
    sense_index: senseGroup.sense_index,
    word_ru: senseGroup.word_ru
  }
  
  showDeleteCardConfirm.value = true
}

const saveCard = async () => {
  if (!cardToEdit.value) return
  
  try {
    // Convert distractors arrays to JSON
    const distractorsRU = JSON.stringify(editForm.value.distractors_ru.filter(v => v.trim() !== ''))
    const distractorsEN = JSON.stringify(editForm.value.distractors_en.filter(v => v.trim() !== ''))
    
    const params = new URLSearchParams()
    params.append('word_ru', editForm.value.word_ru || '')
    params.append('meaning_en', editForm.value.meaning_en || '')
    params.append('example_en', editForm.value.example_en || '')
    params.append('example_ru', editForm.value.example_ru || '')
    params.append('transcription', editForm.value.transcription || '')
    params.append('distractors_ru', distractorsRU)
    params.append('distractors_en', distractorsEN)
    params.append('hint', editForm.value.hint || '')
    
    await apiClient.request(`/app/admin/training/card/${cardToEdit.value.training_card_id}`, { 
      method: 'PUT',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString()
    })
    showEditCardModal.value = false
    cardToEdit.value = null
    await showCards(selectedWord.value) // Reload cards
    await loadVocab() // Reload words list
  } catch (error) {
    console.error('Failed to save card:', error)
    alert('Failed to save card')
  }
}

const deleteCard = async () => {
  if (!cardToDelete.value) return
  
  try {
    await apiClient.request(`/app/admin/training/card/${cardToDelete.value.training_card_id}`, { method: 'DELETE' })
    showDeleteCardConfirm.value = false
    cardToDelete.value = null
    await showCards(selectedWord.value) // Reload cards
    await loadVocab() // Reload words list
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
  justify-content: flex-end;
}

.modal-body {
  margin-top: 20px;
}

.edit-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
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
}

.form-textarea {
  resize: vertical;
  min-height: 60px;
}

.warning-text {
  color: var(--color-danger, #ef4444);
  font-size: 13px;
  margin-top: 8px;
}

.distractors-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
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

.cards-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.sense-group {
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  border-radius: 8px;
  padding: 16px;
  background: var(--card-bg);
}

.sense-header {
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.sense-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.sense-header h4 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  flex: 1;
}

.sense-actions {
  display: flex;
  gap: 8px;
}

.word-en {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 16px;
}

.word-ru {
  font-weight: 600;
  color: var(--color-primary);
  font-size: 18px;
}

.sense-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.directions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.direction-item {
  padding: 12px;
  background: var(--input-bg, rgba(0, 0, 0, 0.02));
  border-radius: 6px;
  border-left: 3px solid var(--color-primary);
}

.direction-header {
  margin-bottom: 12px;
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
