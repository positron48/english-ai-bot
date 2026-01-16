<template>
  <div class="admin-content">
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
        <button 
          @click="toggleOnlyErrors" 
          :class="['btn', 'btn-toggle', { 'btn-toggle-active': wordsOnlyErrors }]"
          type="button"
        >
          Only with errors
        </button>
        <button @click="loadWords" class="btn btn-primary">Refresh</button>
      </div>

      <div class="words-content">
        <div v-if="wordsError && !wordsLoading" class="empty-message">
          <p>{{ wordsError }}</p>
        </div>
        <div v-else-if="words.length === 0 && !wordsLoading" class="empty-message">
          <p v-if="wordsSearchQuery">No words found matching "{{ wordsSearchQuery }}".</p>
          <p v-else>No words found</p>
        </div>
        <div v-else class="words-table-container" :class="{ 'loading-overlay': wordsLoading }">
          <div v-if="wordsLoading" class="loading-overlay-content">
            <div class="loading">Loading words...</div>
          </div>
          <table class="words-table">
            <thead>
              <tr>
                <th 
                  class="desktop-only sortable" 
                  :class="{ 'sort-asc': sortColumn === 'ID' && sortDirection === 'asc', 'sort-desc': sortColumn === 'ID' && sortDirection === 'desc' }"
                  @click="handleSort('ID')"
                >
                  ID
                  <span class="sort-indicator" v-if="sortColumn === 'ID'">
                    {{ sortDirection === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th 
                  class="sortable"
                  :class="{ 'sort-asc': sortColumn === 'Word' && sortDirection === 'asc', 'sort-desc': sortColumn === 'Word' && sortDirection === 'desc' }"
                  @click="handleSort('Word')"
                >
                  Word
                  <span class="sort-indicator" v-if="sortColumn === 'Word'">
                    {{ sortDirection === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th 
                  class="sortable"
                  :class="{ 'sort-asc': sortColumn === 'POS' && sortDirection === 'asc', 'sort-desc': sortColumn === 'POS' && sortDirection === 'desc' }"
                  @click="handleSort('POS')"
                >
                  POS
                  <span class="sort-indicator" v-if="sortColumn === 'POS'">
                    {{ sortDirection === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th 
                  class="sortable"
                  :class="{ 'sort-asc': sortColumn === 'HasCards' && sortDirection === 'asc', 'sort-desc': sortColumn === 'HasCards' && sortDirection === 'desc' }"
                  @click="handleSort('HasCards')"
                >
                  Has Cards
                  <span class="sort-indicator" v-if="sortColumn === 'HasCards'">
                    {{ sortDirection === 'asc' ? '↑' : '↓' }}
                  </span>
                </th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-for="word in words" :key="word.ID">
                <tr :class="{ 'has-error': word.ProcessingError }">
                  <td class="desktop-only">{{ word.ID }}</td>
                  <td><strong>{{ word.Word }}</strong></td>
                  <td>{{ word.POS || '—' }}</td>
                  <td>
                    <span v-if="word.ProcessingError" 
                      @click.prevent="showErrorDetails(word)"
                      class="card-link card-link-error"
                      :style="{ cursor: 'pointer' }"
                    >
                      Error
                    </span>
                    <span v-else-if="!word.HasTrainingCards" :class="{ 'badge': true, 'badge-secondary': true }">
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
                  <td>
                    <div class="action-buttons">
                      <button
                        @click="startEditWord(word)"
                        class="btn btn-sm btn-primary"
                        title="Редактировать"
                      >
                        <Icon name="edit" />
                      </button>
                      <button 
                        v-if="!word.ProcessingError"
                        @click="generateAdditionalCard(word)" 
                        class="btn btn-sm btn-secondary"
                        title="Generate additional card"
                      >
                        <Icon name="magic" />
                      </button>
                      <button 
                        v-if="!word.ProcessingError"
                        @click="createTrainingCard(word)" 
                        class="btn btn-sm btn-primary"
                        title="Add card"
                      >
                        <Icon name="plus" />
                      </button>
                      <button 
                        v-if="word.ProcessingError"
                        @click="resetWordError(word)" 
                        class="btn btn-sm btn-warning"
                        title="Сбросить ошибку"
                      >
                        <Icon name="refresh" />
                      </button>
                      <button 
                        @click="deleteWord(word)" 
                        class="btn btn-sm btn-danger"
                        title="Удалить"
                      >
                        <Icon name="trash" />
                      </button>
                    </div>
                  </td>
                </tr>
                <!-- Cards row -->
                <tr v-if="word && word.showingCards === true && word.HasTrainingCards" class="cards-row">
                <td colspan="6">
                  <div v-if="word.cardsLoading" class="cards-loading">Loading cards...</div>
                  <div v-else class="cards-container">
                    <div class="cards-header">
                      <h4>Training Cards for "{{ word.Word }}"</h4>
                      <div class="cards-header-actions">
                        <button 
                          @click="generateAdditionalCard(word)" 
                          class="btn btn-sm btn-secondary"
                          title="Generate additional card"
                        >
                          <Icon name="magic" />
                        </button>
                        <button 
                          @click="createTrainingCard(word)" 
                          class="btn btn-sm btn-primary"
                          title="Add card"
                        >
                          <Icon name="plus" />
                        </button>
                      </div>
                    </div>
                    <div v-if="word.cards && word.cards.length > 0">
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
                            title="Редактировать"
                          >
                            <Icon name="edit" />
                          </button>
                          <button 
                            @click="confirmDeleteTrainingCard(card, word)" 
                            class="btn btn-sm btn-danger"
                            title="Удалить"
                          >
                            <Icon name="trash" />
                          </button>
                        </div>
                      </div>
                      <div class="card-content">
                        <div class="card-field">
                          <span class="field-label">Word EN:</span>
                          <span class="field-value">{{ card.word_en }}</span>
                          <span v-if="card.transcription" class="transcription">{{ card.transcription }}</span>
                        </div>
                        <div v-if="card.pos" class="card-field">
                          <span class="field-label">POS:</span>
                          <span class="field-value">{{ card.pos }}</span>
                        </div>
                        <div v-if="card.display_word" class="card-field">
                          <span class="field-label">Display Word:</span>
                          <span class="field-value">{{ card.display_word }}</span>
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
                  </div>
                </td>
              </tr>
              </template>
            </tbody>
          </table>
          </div>
        </div>
        <div class="pagination" v-if="wordsPagination.total_pages > 1 && !wordsLoading">
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

      <!-- Edit Word Modal -->
      <div v-if="showEditWordModal && wordToEdit" class="modal">
        <div class="modal-content modal-large">
          <div class="modal-header">
            <h3>Edit Word: {{ wordToEdit.Word }}</h3>
            <button @click="closeEditWordModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveWord" class="edit-form">
              <div class="form-group">
                <label>Word (Lemma):</label>
                <input v-model="editWordForm.word" type="text" required class="form-input" />
              </div>
              <div class="form-group">
                <label>Part of Speech (POS):</label>
                <input v-model="editWordForm.pos" type="text" class="form-input" placeholder="noun, verb, adjective, etc." />
              </div>
              <div class="form-group">
                <label>Transcription (IPA):</label>
                <input v-model="editWordForm.transcription" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label>Definition RU:</label>
                <textarea v-model="editWordForm.definition_ru" class="form-textarea" rows="3"></textarea>
              </div>
              <div class="form-group">
                <label>Display EN:</label>
                <input v-model="editWordForm.display_en" type="text" class="form-input" placeholder="e.g., 'spy' or 'to spy' for verbs" />
              </div>
              <div class="form-group">
                <label>Examples (JSON):</label>
                <textarea v-model="editWordForm.examples_json" class="form-textarea" rows="5" placeholder='[{"example_en": "...", "gloss_ru": "..."}]'></textarea>
              </div>
              <div class="form-group">
                <label>Verb Forms (JSON):</label>
                <textarea v-model="editWordForm.verb_forms_json" class="form-textarea" rows="4" placeholder='{"v1": "...", "v2": "...", "v3": "..."}'></textarea>
              </div>
              <div class="form-group">
                <label>Definition (Legacy):</label>
                <textarea v-model="editWordForm.definition" class="form-textarea" rows="5"></textarea>
              </div>
              <div class="form-actions">
                <button type="button" @click="generateWordCardData" class="btn btn-secondary" :disabled="generatingWordData">
                  {{ generatingWordData ? 'Generating...' : 'AI Fill' }}
                </button>
                <button type="submit" class="btn btn-primary">Save</button>
                <button type="button" @click="closeEditWordModal" class="btn">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Generate Additional Card Modal -->
      <div v-if="showGenerateCardModal && wordToGenerateCard" class="modal" @click.self="closeGenerateCardModal">
        <div class="modal-content">
          <div class="modal-header">
            <h3>Generate Additional Card for "{{ wordToGenerateCard.Word }}"</h3>
            <button @click="closeGenerateCardModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <div class="form-group">
              <label>Constraints (optional):</label>
              <textarea 
                ref="generateCardConstraintsTextarea"
                v-model="generateCardConstraints" 
                class="form-textarea" 
                rows="5"
                placeholder="For example: specific meaning 'bank as financial institution', or part of speech 'verb', or 'meaning related to water'"
              ></textarea>
              <p class="form-hint">Describe what meaning or part of speech the card should have. Leave empty for random generation.</p>
            </div>
            <div class="modal-actions">
              <button 
                @click="doGenerateCard" 
                class="btn btn-primary"
                :disabled="generatingCard"
              >
                {{ generatingCard ? 'Generating...' : 'Generate' }}
              </button>
              <button 
                type="button" 
                @click="closeGenerateCardModal" 
                class="btn btn-secondary"
                :disabled="generatingCard"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Create Training Card Modal -->
      <div v-if="showCreateCardModal && wordToCreateCard" class="modal">
        <div class="modal-content modal-large">
          <div class="modal-header">
            <h3>Create Training Card for "{{ wordToCreateCard.Word }}"</h3>
            <button @click="closeCreateCardModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveNewTrainingCard" class="edit-form">
              <div class="form-group">
                <label>Word EN:</label>
                <input v-model="createCardForm.word_en" type="text" required class="form-input" readonly />
              </div>
              <div class="form-group">
                <label>Part of Speech (POS):</label>
                <input v-model="createCardForm.pos" type="text" class="form-input" placeholder="noun, verb, adjective, etc." />
              </div>
              <div class="form-group">
                <label>Display Word:</label>
                <input v-model="createCardForm.display_word" type="text" class="form-input" placeholder="e.g., 'spy' or 'to spy' for verbs" />
              </div>
              <div class="form-group">
                <label>Transcription (IPA):</label>
                <input v-model="createCardForm.transcription" type="text" class="form-input" />
              </div>
              <div class="form-group">
                <label>Word RU:</label>
                <input v-model="createCardForm.word_ru" type="text" required class="form-input" />
              </div>
              <div class="form-group">
                <label>Meaning EN:</label>
                <textarea v-model="createCardForm.meaning_en" required class="form-textarea" rows="3"></textarea>
              </div>
              <div class="form-group">
                <label>Example EN:</label>
                <textarea v-model="createCardForm.example_en" class="form-textarea" rows="2"></textarea>
              </div>
              <div class="form-group">
                <label>Example RU:</label>
                <textarea v-model="createCardForm.example_ru" class="form-textarea" rows="2"></textarea>
              </div>
              <div class="form-group">
                <label>Distractors RU:</label>
                <div class="distractors-list">
                  <input v-model="createCardForm.distractors_ru[0]" type="text" class="form-input" placeholder="Option 1" />
                  <input v-model="createCardForm.distractors_ru[1]" type="text" class="form-input" placeholder="Option 2" />
                  <input v-model="createCardForm.distractors_ru[2]" type="text" class="form-input" placeholder="Option 3" />
                </div>
              </div>
              <div class="form-group">
                <label>Distractors EN:</label>
                <div class="distractors-list">
                  <input v-model="createCardForm.distractors_en[0]" type="text" class="form-input" placeholder="Option 1" />
                  <input v-model="createCardForm.distractors_en[1]" type="text" class="form-input" placeholder="Option 2" />
                  <input v-model="createCardForm.distractors_en[2]" type="text" class="form-input" placeholder="Option 3" />
                </div>
              </div>
              <div class="form-group">
                <label>Hint:</label>
                <input v-model="createCardForm.hint" type="text" class="form-input" />
              </div>
              <div class="modal-actions">
                <button type="submit" class="btn btn-primary">Create</button>
                <button type="button" @click="closeCreateCardModal" class="btn btn-secondary">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Edit Training Card Modal -->
      <div v-if="showEditCardModal && cardToEdit" class="modal">
        <div class="modal-content modal-large">
          <div class="modal-header">
            <h3>Edit Training Card (Sense {{ cardToEdit.sense_index }})</h3>
            <button @click="closeEditCardModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveTrainingCard" class="edit-form">
              <div class="form-group">
                <label>Word EN:</label>
                <input v-model="editCardForm.word_en" type="text" required class="form-input" />
              </div>
              <div class="form-group">
                <label>Part of Speech (POS):</label>
                <input v-model="editCardForm.pos" type="text" class="form-input" placeholder="noun, verb, adjective, etc." />
              </div>
              <div class="form-group">
                <label>Display Word:</label>
                <input v-model="editCardForm.display_word" type="text" class="form-input" placeholder="e.g., 'spy' or 'to spy' for verbs" />
              </div>
              <div class="form-group">
                <label>Transcription (IPA):</label>
                <input v-model="editCardForm.transcription" type="text" class="form-input" />
              </div>
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

      <!-- Reset Word Error Confirmation Modal -->
      <div v-if="showResetErrorConfirm && wordToResetError" class="modal" @click.self="closeResetErrorConfirm">
        <div class="modal-content">
          <div class="modal-header">
            <h3>Reset Error</h3>
            <button @click="closeResetErrorConfirm" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <p>Reset error for word "<strong>{{ wordToResetError.Word }}</strong>"?</p>
            <p>This will allow the worker to process it again.</p>
          </div>
          <div class="modal-actions">
            <button @click="confirmResetError" class="btn btn-warning">Reset Error</button>
            <button @click="closeResetErrorConfirm" class="btn btn-secondary">Cancel</button>
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

      <!-- Error Details Modal -->
      <div v-if="showErrorDetailsModal && wordWithError" class="modal" @click.self="closeErrorDetailsModal">
        <div class="modal-content">
          <div class="modal-header">
            <h3>Error Details: {{ wordWithError.Word }}</h3>
            <button @click="closeErrorDetailsModal" class="btn-close">&times;</button>
          </div>
          <div class="modal-body">
            <div class="error-details">
              <div class="error-detail-row">
                <span class="error-detail-label">Date:</span>
                <span class="error-detail-value" v-if="wordWithError.ProcessedAt" :title="formatDateAbsolute(wordWithError.ProcessedAt)">{{ formatDateRelative(wordWithError.ProcessedAt) }}</span>
                <span class="error-detail-value" v-else>—</span>
              </div>
              <div class="error-detail-row">
                <span class="error-detail-label">Error:</span>
                <span class="error-detail-value error-text-full">{{ wordWithError.ProcessingError }}</span>
              </div>
            </div>
          </div>
          <div class="modal-actions">
            <button @click="closeErrorDetailsModal" class="btn btn-secondary">Close</button>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import Icon from '../components/Icon.vue'


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
  pos?: string
  display_word?: string
  created_at?: string
}

interface WordCard {
  ID: number
  Word: string
  Definition: string
  POS?: string | null
  Transcription?: string | null
  DefinitionRU?: string | null
  ExamplesJSON?: string | null
  VerbFormsJSON?: string | null
  DisplayEN?: string | null
  ProcessedAt?: string | null
  ProcessingError?: string | null
  HasTrainingCards: boolean
  RequestingUsers?: number[]
  showingCards?: boolean
  cards?: TrainingCard[]
  cardsLoading?: boolean
}

const loading = ref(false)

// Words management
const words = ref<WordCard[]>([])
const wordsLoading = ref(false)
const wordsError = ref<string | null>(null)
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

// Sorting state
const sortColumn = ref<string | null>(null)
const sortDirection = ref<'asc' | 'desc'>('asc')

// Training card editing
const showEditWordModal = ref(false)
const wordToEdit = ref<WordCard | null>(null)
const editWordForm = ref({
  word: '',
  pos: '',
  transcription: '',
  definition_ru: '',
  display_en: '',
  examples_json: '',
  verb_forms_json: '',
  definition: ''
})

const showEditCardModal = ref(false)
const showCreateCardModal = ref(false)
const showGenerateCardModal = ref(false)
const showDeleteCardConfirm = ref(false)
const showResetErrorConfirm = ref(false)
const showErrorDetailsModal = ref(false)
const wordToResetError = ref<WordCard | null>(null)
const wordWithError = ref<WordCard | null>(null)
const cardToEdit = ref<TrainingCard | null>(null)
const cardToDelete = ref<{ card: TrainingCard; word: WordCard } | null>(null)
const wordToCreateCard = ref<WordCard | null>(null)
const wordToGenerateCard = ref<WordCard | null>(null)
const generateCardConstraints = ref('')
const generatingCard = ref(false)
const generateCardConstraintsTextarea = ref<HTMLTextAreaElement | null>(null)
const generatingWordData = ref(false)
const editCardForm = ref({
  word_en: '',
  pos: '',
  display_word: '',
  transcription: '',
  word_ru: '',
  meaning_en: '',
  example_en: '',
  example_ru: '',
  distractors_ru: ['', '', ''],
  distractors_en: ['', '', ''],
  hint: ''
})
const createCardForm = ref({
  word_en: '',
  pos: '',
  display_word: '',
  transcription: '',
  word_ru: '',
  meaning_en: '',
  example_en: '',
  example_ru: '',
  distractors_ru: ['', '', ''],
  distractors_en: ['', '', ''],
  hint: ''
})

onMounted(async () => {
  await loadUsers()
  await loadWords()
})

const loadUsers = async () => {
  try {
    const data: { users: User[] } = await apiClient.request('/api/admin/users')
    users.value = data.users
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

const loadWords = async () => {
  wordsLoading.value = true
  wordsError.value = null
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
    
    // Add sorting parameters
    if (sortColumn.value) {
      // Map frontend column names to backend sort_by values
      const sortByMap: Record<string, string> = {
        'ID': 'id',
        'Word': 'word',
        'POS': 'pos',
        'HasCards': 'has_cards'
      }
      const backendSortBy = sortByMap[sortColumn.value]
      if (backendSortBy) {
        params.append('sort_by', backendSortBy)
        params.append('sort_order', sortDirection.value)
      }
    }

    const data: { words: WordCard[]; pagination: { page: number; limit: number; total: number; total_pages: number } } = await apiClient.request(`/api/admin/words?${params.toString()}`)
    words.value = (data.words || []).map(w => ({ 
      ...w, 
      editing: false, 
      showingCards: false,
      cardsLoading: false,
      cards: undefined
    }))
    if (data.pagination) {
      wordsPagination.value = data.pagination
    } else {
      wordsPagination.value = {
        page: 1,
        limit: 50,
        total: 0,
        total_pages: 0
      }
    }
  } catch (error: any) {
    console.error('Failed to load words:', error)
    wordsError.value = error.message || 'Failed to load words'
    words.value = []
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
  }, 200)
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

const toggleOnlyErrors = () => {
  wordsOnlyErrors.value = !wordsOnlyErrors.value
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
      const data: { word_en: string; cards: TrainingCard[] } = await apiClient.request(`/api/admin/training/${word.Word}`)
      word.cards = data.cards || []
    } catch (error) {
      console.error('Failed to load training cards:', error)
      await showAlert('Failed to load training cards')
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
    word_en: card.word_en || '',
    pos: card.pos || '',
    display_word: card.display_word || '',
    transcription: card.transcription || '',
    word_ru: card.word_ru || '',
    meaning_en: card.meaning_en || '',
    example_en: card.example_en || '',
    example_ru: card.example_ru || '',
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
    const data: { word_en: string; cards: TrainingCard[] } = await apiClient.request(`/api/admin/training/${word.Word}`)
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
    
    await apiClient.request(`/api/admin/training/card/${cardToEdit.value.id}`, { 
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
    
    await showAlert('Card updated successfully')
  } catch (error) {
    console.error('Failed to save card:', error)
    await showAlert('Failed to save card')
  }
}

const deleteTrainingCard = async () => {
  if (!cardToDelete.value) return
  
  try {
    await apiClient.request(`/api/admin/training/card/${cardToDelete.value.card.id}`, { method: 'DELETE' })
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
    
    await showAlert('Card deleted successfully')
  } catch (error) {
    console.error('Failed to delete card:', error)
    await showAlert('Failed to delete card')
  }
}

const closeEditCardModal = () => {
  showEditCardModal.value = false
  cardToEdit.value = null
}

const createTrainingCard = (word: WordCard) => {
  wordToCreateCard.value = word
  createCardForm.value = {
    word_en: word.Word || '',
    pos: word.POS || '',
    display_word: word.DisplayEN || '',
    transcription: word.Transcription || '',
    word_ru: '',
    meaning_en: '',
    example_en: '',
    example_ru: '',
    distractors_ru: ['', '', ''],
    distractors_en: ['', '', ''],
    hint: ''
  }
  showCreateCardModal.value = true
}

const generateAdditionalCard = (word: WordCard) => {
  wordToGenerateCard.value = word
  generateCardConstraints.value = ''
  showGenerateCardModal.value = true
  // Focus textarea after modal opens
  nextTick(() => {
    if (generateCardConstraintsTextarea.value) {
      generateCardConstraintsTextarea.value.focus()
    }
  })
}

const closeGenerateCardModal = () => {
  showGenerateCardModal.value = false
  wordToGenerateCard.value = null
  generateCardConstraints.value = ''
  generatingCard.value = false
}

const doGenerateCard = async () => {
  if (!wordToGenerateCard.value || generatingCard.value) return
  
  generatingCard.value = true
  try {
    const response: any = await apiClient.request(`/api/admin/training/${wordToGenerateCard.value.Word}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        constraints: generateCardConstraints.value
      })
    })
    
    if (response.success && response.card) {
      const card = response.card
      
      // Parse distractors from JSON strings
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
      
      // Fill create card form with generated data
      wordToCreateCard.value = wordToGenerateCard.value
      createCardForm.value = {
        word_en: card.word_en || wordToGenerateCard.value.Word || '',
        pos: card.pos || '',
        display_word: card.display_word || '',
        transcription: card.transcription || '',
        word_ru: card.word_ru || '',
        meaning_en: card.meaning_en || '',
        example_en: card.example_en || '',
        example_ru: card.example_ru || '',
        distractors_ru: distractorsRU,
        distractors_en: distractorsEN,
        hint: card.hint || ''
      }
      
      // Close generate modal and open create modal
      closeGenerateCardModal()
      showCreateCardModal.value = true
    } else {
      await showAlert('Failed to generate card')
    }
  } catch (error: any) {
    console.error('Failed to generate card:', error)
    await showAlert(error.message || 'Failed to generate card')
  } finally {
    generatingCard.value = false
  }
}

const closeCreateCardModal = () => {
  showCreateCardModal.value = false
  wordToCreateCard.value = null
}

const saveNewTrainingCard = async () => {
  if (!wordToCreateCard.value) return
  
  try {
    // Convert distractors arrays to JSON
    const distractorsRU = JSON.stringify(createCardForm.value.distractors_ru.filter(v => v.trim() !== ''))
    const distractorsEN = JSON.stringify(createCardForm.value.distractors_en.filter(v => v.trim() !== ''))
    
    const params = new URLSearchParams()
    params.append('word_ru', createCardForm.value.word_ru || '')
    params.append('meaning_en', createCardForm.value.meaning_en || '')
    params.append('example_en', createCardForm.value.example_en || '')
    params.append('example_ru', createCardForm.value.example_ru || '')
    params.append('transcription', createCardForm.value.transcription || '')
    params.append('distractors_ru', distractorsRU)
    params.append('distractors_en', distractorsEN)
    params.append('hint', createCardForm.value.hint || '')
    if (createCardForm.value.pos) {
      params.append('pos', createCardForm.value.pos)
    }
    if (createCardForm.value.display_word) {
      params.append('display_word', createCardForm.value.display_word)
    }
    
    await apiClient.request(`/api/admin/training/${wordToCreateCard.value.Word}`, { 
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString()
    })
    
    showCreateCardModal.value = false
    const word = wordToCreateCard.value
    wordToCreateCard.value = null
    
    // Update HasTrainingCards status
    word.HasTrainingCards = true
    
    // If cards are not shown, reload them
    if (word.showingCards) {
      await reloadWordCards(word)
    } else {
      // Just update status in list
      const wordInList = words.value.find(w => w.ID === word.ID)
      if (wordInList) {
        wordInList.HasTrainingCards = true
      }
    }
    
    await showAlert('Card created successfully')
  } catch (error) {
    console.error('Failed to create card:', error)
    await showAlert('Failed to create card')
  }
}

const closeDeleteCardConfirm = () => {
  showDeleteCardConfirm.value = false
  cardToDelete.value = null
}

const startEditWord = (word: WordCard) => {
  wordToEdit.value = word
  editWordForm.value = {
    word: word.Word || '',
    pos: word.POS || '',
    transcription: word.Transcription || '',
    definition_ru: word.DefinitionRU || '',
    display_en: word.DisplayEN || '',
    examples_json: word.ExamplesJSON || '',
    verb_forms_json: word.VerbFormsJSON || '',
    definition: word.Definition || ''
  }
  showEditWordModal.value = true
}

const closeEditWordModal = () => {
  showEditWordModal.value = false
  wordToEdit.value = null
}

const generateWordCardData = async () => {
  if (!wordToEdit.value || generatingWordData.value) return
  
  generatingWordData.value = true
  try {
    const response: any = await apiClient.request(`/api/admin/words/${wordToEdit.value.ID}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (response.success && response.word_card) {
      const wordCard = response.word_card
      
      // Fill edit form with generated data
      if (wordCard.word) {
        editWordForm.value.word = wordCard.word
      }
      if (wordCard.pos) {
        editWordForm.value.pos = wordCard.pos
      }
      if (wordCard.transcription) {
        editWordForm.value.transcription = wordCard.transcription
      }
      if (wordCard.definition_ru) {
        editWordForm.value.definition_ru = wordCard.definition_ru
      }
      if (wordCard.display_en) {
        editWordForm.value.display_en = wordCard.display_en
      }
      if (wordCard.examples_json) {
        editWordForm.value.examples_json = wordCard.examples_json
      }
      if (wordCard.verb_forms_json) {
        editWordForm.value.verb_forms_json = wordCard.verb_forms_json
      }
      
      await showAlert('Word card data generated successfully')
    } else {
      await showAlert('Failed to generate word card data')
    }
  } catch (error: any) {
    console.error('Failed to generate word card data:', error)
    await showAlert(error.message || 'Failed to generate word card data')
  } finally {
    generatingWordData.value = false
  }
}

const saveWord = async () => {
  if (!wordToEdit.value) return

  try {
    await apiClient.request(`/api/admin/words/${wordToEdit.value.ID}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(editWordForm.value)
    })
    
    // Update word in list
    const word = words.value.find(w => w.ID === wordToEdit.value!.ID)
    if (word) {
      word.Word = editWordForm.value.word
      word.POS = editWordForm.value.pos || null
      word.Transcription = editWordForm.value.transcription || null
      word.DefinitionRU = editWordForm.value.definition_ru || null
      word.DisplayEN = editWordForm.value.display_en || null
      word.ExamplesJSON = editWordForm.value.examples_json || null
      word.VerbFormsJSON = editWordForm.value.verb_forms_json || null
      word.Definition = editWordForm.value.definition
    }
    
    closeEditWordModal()
  } catch (error: any) {
    console.error('Failed to update word:', error)
    const errorMessage = error.message || 'Failed to update word'
    await showAlert(errorMessage)
  }
}

const resetWordError = (word: WordCard) => {
  wordToResetError.value = word
  showResetErrorConfirm.value = true
}

const closeResetErrorConfirm = () => {
  showResetErrorConfirm.value = false
  wordToResetError.value = null
}

const confirmResetError = async () => {
  const targetWord = wordToResetError.value
  if (!targetWord) return

  try {
    await apiClient.request(`/api/admin/words/${targetWord.ID}/reset`, { method: 'POST' })
    targetWord.ProcessingError = null
    targetWord.ProcessedAt = null
    closeResetErrorConfirm()
  } catch (error) {
    console.error('Failed to reset word error:', error)
    // Show error in modal or use a toast notification
    // For now, just close and let user see the error in console
    closeResetErrorConfirm()
  }
}

const deleteWord = async (word: WordCard) => {
  const confirmed = await showConfirm(`Are you sure you want to delete word "${word.Word}"?\n\nThis will delete:\n- The word itself\n- All training cards\n- All user cards\n- All request history\n\nThis action cannot be undone!`)
  if (!confirmed) {
    return
  }

  try {
    await apiClient.request(`/api/admin/words/${word.ID}`, { method: 'DELETE' })
    // Remove word from list
    const index = words.value.findIndex(w => w.ID === word.ID)
    if (index !== -1) {
      words.value.splice(index, 1)
    }
  } catch (error) {
    console.error('Failed to delete word:', error)
    await showAlert('Failed to delete word')
  }
}

const showErrorDetails = (word: WordCard) => {
  wordWithError.value = word
  showErrorDetailsModal.value = true
}

const closeErrorDetailsModal = () => {
  showErrorDetailsModal.value = false
  wordWithError.value = null
}


const formatDateRelative = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '—'
  
  let date: Date
  if (dateStr.includes(' ')) {
    date = new Date(dateStr.replace(' ', 'T'))
  } else {
    date = new Date(dateStr)
  }
  
  if (isNaN(date.getTime())) return '—'
  
  const now = new Date()
  const diffTime = now.getTime() - date.getTime()
  const diffDays = Math.floor(Math.abs(diffTime) / (1000 * 60 * 60 * 24))
  const diffHours = Math.floor(Math.abs(diffTime) / (1000 * 60 * 60))
  const diffMinutes = Math.floor(Math.abs(diffTime) / (1000 * 60))
  const isFuture = diffTime < 0
  
  // Today
  if (diffDays === 0) {
    if (diffHours === 0) {
      if (diffMinutes < 1) {
        return 'just now'
      }
      if (isFuture) {
        return `in ${diffMinutes} ${diffMinutes === 1 ? 'minute' : 'minutes'}`
      }
      return `${diffMinutes} ${diffMinutes === 1 ? 'minute' : 'minutes'} ago`
    }
    if (isFuture) {
      return `in ${diffHours} ${diffHours === 1 ? 'hour' : 'hours'}`
    }
    return `${diffHours} ${diffHours === 1 ? 'hour' : 'hours'} ago`
  }
  
  // Tomorrow / Yesterday
  if (diffDays === 1) {
    return isFuture ? 'tomorrow' : 'yesterday'
  }
  
  // Days
  if (diffDays < 7) {
    return isFuture ? `in ${diffDays} days` : `${diffDays} days ago`
  }
  
  // Weeks
  const diffWeeks = Math.floor(diffDays / 7)
  if (diffWeeks < 4) {
    if (isFuture) {
      return `in ${diffWeeks} ${diffWeeks === 1 ? 'week' : 'weeks'}`
    }
    return `${diffWeeks} ${diffWeeks === 1 ? 'week' : 'weeks'} ago`
  }
  
  // Months
  const diffMonths = Math.floor(diffDays / 30)
  if (diffMonths < 12) {
    if (isFuture) {
      return `in ${diffMonths} ${diffMonths === 1 ? 'month' : 'months'}`
    }
    return `${diffMonths} ${diffMonths === 1 ? 'month' : 'months'} ago`
  }
  
  // Years
  const diffYears = Math.floor(diffDays / 365)
  if (isFuture) {
    return `in ${diffYears} ${diffYears === 1 ? 'year' : 'years'}`
  }
  return `${diffYears} ${diffYears === 1 ? 'year' : 'years'} ago`
}

const formatDateAbsolute = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '—'
  
  let date: Date
  if (dateStr.includes(' ')) {
    date = new Date(dateStr.replace(' ', 'T'))
  } else {
    date = new Date(dateStr)
  }
  
  if (isNaN(date.getTime())) return '—'
  
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  
  return `${day}.${month}.${year}`
}

// Sorting logic
const handleSort = (column: string) => {
  if (sortColumn.value === column) {
    // Toggle direction if same column
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    // New column, default to ascending
    sortColumn.value = column
    sortDirection.value = 'asc'
  }
  // Reload words with new sorting
  wordsPagination.value.page = 1
  loadWords()
}

</script>

<style scoped>
.admin-content {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  font-size: 16px;
}

.admin h1 {
  margin-bottom: 24px;
}

.admin .card h2 {
  margin-bottom: 20px;
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
  padding: 8px 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  box-sizing: border-box;
  height: 40px;
  line-height: 1.5;
  margin-bottom: 0;
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
  height: 40px;
  box-sizing: border-box;
  font-size: 14px;
  line-height: 1.5;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary);
  cursor: pointer;
  height: 40px;
}

.checkbox-label input[type="checkbox"] {
  cursor: pointer;
  margin: 0;
}

.words-filters .btn {
  height: 40px;
  padding: 8px 16px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn-toggle {
  background-color: var(--input-bg);
  color: var(--text-primary);
  border: 1px solid var(--input-border);
  transition: all 0.2s ease;
}

.btn-toggle:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.1));
  border-color: var(--color-primary);
}

.btn-toggle-active {
  background-color: var(--color-primary);
  color: white;
  border-color: var(--color-primary);
}

.btn-toggle-active:hover {
  background-color: var(--color-primary-hover, var(--color-primary));
  opacity: 0.9;
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
  padding: 10px 12px;
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

.words-table th.sortable {
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s;
  padding-right: 24px;
  position: relative;
}

.words-table th.sortable:hover {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.1));
}

.words-table th.sortable.sort-asc,
.words-table th.sortable.sort-desc {
  background-color: var(--table-row-hover, rgba(0, 0, 0, 0.15));
}

.sort-indicator {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 12px;
  color: var(--color-primary);
  font-weight: bold;
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

.action-buttons .btn,
.card-actions .btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  padding: 6px 10px;
}

.action-buttons .btn .icon,
.card-actions .btn .icon {
  width: 16px;
  height: 16px;
  margin: 0;
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

.cards-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 10px;
}

.cards-header h4 {
  margin: 0;
}

.cards-header-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
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

.form-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
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

.words-content {
  position: relative;
}

.words-table-container {
  position: relative;
}

.words-table-container.loading-overlay {
  min-height: 200px;
}

.loading-overlay-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--card-bg);
  opacity: 0.9;
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  border-radius: 8px;
}

.loading-overlay-content .loading {
  font-size: 16px;
  color: var(--text-primary);
  padding: 20px;
}

.card-link.card-link-error {
  color: var(--color-danger) !important;
  border-bottom-color: var(--color-danger) !important;
}

.card-link.card-link-error:hover {
  color: var(--color-danger) !important;
  border-bottom-color: var(--color-danger) !important;
  opacity: 0.8;
}

.error-details {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.error-detail-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.error-detail-label {
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 14px;
}

.error-detail-value {
  color: var(--text-primary);
  word-break: break-word;
}

.error-text-full {
  color: var(--color-danger);
  font-size: 14px;
  line-height: 1.5;
}

.desktop-only {
  display: table-cell;
}

@media (max-width: 767px) {
  .admin-content {
    margin-top: 0 !important;
  }

  .desktop-only {
    display: none;
  }

  .words-filters {
    flex-direction: column;
    align-items: stretch;
  }

  .words-filters .search-box {
    max-width: 100%;
    width: 100%;
  }

  .words-filters .admin-select {
    width: 100%;
    min-width: auto;
  }

  .words-filters .btn {
    width: 100%;
  }

  .words-table th,
  .words-table td {
    padding: 6px;
    font-size: 13px;
  }

  .action-buttons {
    flex-direction: column;
    gap: 4px;
  }

  .btn-sm {
    padding: 4px 8px;
    font-size: 11px;
  }

  .admin-tab {
    padding: 10px 16px;
  }
}

@media (max-width: 480px) {
  .admin-tab {
    padding: 8px 12px;
  }
}
</style>

