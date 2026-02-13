<template>
  <div class="vocab">
    <div class="vocab-header">
      <div class="header-controls">
        <div class="search-box">
          <input
            type="text"
            v-model="searchQuery"
            @input="onSearchInput"
            :placeholder="t('vocab.searchWords')"
            class="search-input"
          />
        </div>
        <div class="filter-controls">
          <label for="status-filter" class="filter-label">{{ t('vocab.status') }}</label>
          <select id="status-filter" v-model="statusFilter" @change="onFilterChange" class="filter-select">
            <option value="">{{ t('vocab.all') }}</option>
            <option value="new">{{ t('vocab.new') }}</option>
            <option value="learning">{{ t('vocab.learning') }}</option>
            <option value="mastered">{{ t('vocab.mastered') }}</option>
            <option value="known">{{ t('vocab.known') }}</option>
          </select>
        </div>
        <div class="sort-controls">
          <label for="sort-select" class="sort-label">{{ t('vocab.sortBy') }}</label>
          <select id="sort-select" v-model="sortField" @change="onSortChange" class="sort-select">
            <option value="display_word">A→Z</option>
            <option value="display_word_desc">Z→A</option>
            <option value="added_at">{{ t('vocab.recentlyAdded') }}</option>
            <option value="mastery_level">{{ t('vocab.mastery') }}</option>
            <option value="mastery_level_desc">{{ t('vocab.masteryReversed') }}</option>
            <option value="mastering_score">{{ t('vocab.masteringScore') }}</option>
            <option value="mastering_score_desc">{{ t('vocab.masteringScoreDesc') }}</option>
          </select>
        </div>
      </div>
    </div>

    <div class="vocab-content">
      <div v-if="words.length === 0 && !loading" class="empty-state">
        <p v-if="searchQuery">
          {{ t('vocab.noWordsFound', { query: searchQuery }) }}
        </p>
        <p v-else>
          {{ t('vocab.noWordsInVocabulary') }}
        </p>
      </div>
      
      <div v-else>
        <div class="words-list" :class="{ 'loading-overlay': loading }">
          <div v-if="loading" class="loading-overlay-content">
            <div class="loading">{{ t('common.loading') }}</div>
          </div>
          
          <template v-if="!loading">
            <!-- Alphabetical sections when sorting by display_word -->
            <template v-if="sortField === 'display_word' || sortField === 'display_word_desc'">
              <div v-for="section in alphabetSections" :key="section.letter" class="alphabet-section">
                <h2 class="section-header">{{ section.letter }}</h2>
                <div class="words-grid">
                  <div 
                    v-for="word in section.words" 
                    :key="word.word_card_id"
                    class="word-card"
                    @click="showCards(word.lemma)"
                  >
                    <div class="word-main">
                      <div class="word-text">
                        <span class="word-display">{{ cleanLemma(word.lemma) }}</span>
                      </div>
                      <span
                        class="mastery-marker"
                        :style="{ backgroundColor: masteryColor(word.mastering_score) }"
                        :title="t(`vocab.${word.mastery_level}`)"
                      />
                      <span :class="['mastery-badge', `mastery-${word.mastery_level}`]">
                        {{ t(`vocab.${word.mastery_level}`) }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </template>
            
            <!-- Regular list for other sortings -->
            <template v-else>
              <div class="words-grid">
                <div 
                  v-for="word in words" 
                  :key="word.word_card_id"
                  class="word-card"
                  @click="showCards(word.lemma)"
                >
                  <div class="word-main">
                    <div class="word-text">
                      <span class="word-display">{{ cleanLemma(word.lemma) }}</span>
                    </div>
                    <span
                      class="mastery-marker"
                      :style="{ backgroundColor: masteryColor(word.mastering_score) }"
                      :title="t(`vocab.${word.mastery_level}`)"
                    />
                    <span :class="['mastery-badge', `mastery-${word.mastery_level}`]">
                      {{ t(`vocab.${word.mastery_level}`) }}
                    </span>
                  </div>
                </div>
              </div>
            </template>
          </template>
        </div>

        <div class="pagination" v-if="pagination.total_pages > 1 && !loading">
          <button 
            @click="goToPage(pagination.page - 1)" 
            :disabled="pagination.page <= 1"
            class="btn btn-secondary"
          >
            {{ t('vocab.previous') }}
          </button>
          <span class="page-info">
            {{ t('vocab.page', { page: pagination.page, totalPages: pagination.total_pages }) }} 
            {{ t('vocab.total', { total: pagination.total }) }}
          </span>
          <button 
            @click="goToPage(pagination.page + 1)" 
            :disabled="pagination.page >= pagination.total_pages"
            class="btn btn-secondary"
          >
            {{ t('vocab.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Word Details Modal (Simplified) -->
    <div v-if="showCardsModal" class="modal" @click.self="closeCardsModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <div class="word-header-info">
            <div class="word-title-row">
              <h3>{{ selectedWordDisplay }}</h3>
              <div v-if="selectedTranscription" class="transcription">{{ selectedTranscription }}</div>
            </div>
            <div class="word-summary">
              <span>{{ t('vocab.cards', totalCards, { n: totalCards }) }}</span>
              <span v-if="totalDue > 0">{{ t('vocab.due', totalDue, { n: totalDue }) }}</span>
              <span v-if="lastReview" :title="formatDateAbsolute(lastReview)">{{ t('vocab.last') }} {{ formatDateRelative(lastReview) }}</span>
            </div>
          </div>
          <button @click="closeCardsModal" class="btn-close">&times;</button>
        </div>
        <div v-if="cardsLoading" class="loading">{{ t('vocab.loadingCards') }}</div>
        <div v-else-if="cards.length === 0" class="no-cards">{{ t('vocab.noCardsFound') }}</div>
        <div v-else>
          <!-- Verb Forms Section -->
          <div v-if="wordPOS === 'verb' && verbForms" class="verb-forms-section">
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
          
          <div class="cards-list-simple">
          <div v-for="senseGroup in groupedCards" :key="senseGroup.sense_index" class="sense-group-simple">
            <div class="sense-header-simple">
              <h4>
                <span class="word-ru">{{ senseGroup.word_ru }}</span>
                <span v-if="senseGroup.pos" class="pos-badge">{{ senseGroup.pos }}</span>
              </h4>
            </div>
            
            <div class="sense-info-simple">
              <div v-if="senseGroup.meaning_en" class="meaning">
                {{ senseGroup.meaning_en }}
              </div>
              <div v-if="senseGroup.example_en" class="example">
                <strong>{{ t('vocab.example') }}:</strong> {{ senseGroup.example_en }}
              </div>
            </div>
            
            <div class="directions-simple">
              <div v-for="directionCard in senseGroup.directions" :key="directionCard.direction" class="direction-item-simple">
                <div class="direction-header-simple">
                  <span class="direction-badge" :class="`direction-${directionCard.direction}`">
                    {{ directionCard.direction === 'ru_en' ? 'RU→EN' : 'EN→RU' }}
                  </span>
                  <span :class="['state-badge', `state-${directionCard.state}`]">{{ directionCard.state }}</span>
                </div>
                <div class="direction-stats-simple">
                  <span v-if="directionCard.reps > 0" :title="t('vocab.reps')">{{ t('vocab.reps') }} {{ directionCard.reps }}</span>
                  <span v-else-if="directionCard.review_count > 0" :title="t('vocab.reviews')">{{ t('vocab.reviews') }} {{ directionCard.review_count }}</span>
                  <span v-if="directionCard.next_due_at" :title="`${t('vocab.due')}: ${formatDateAbsolute(directionCard.next_due_at)}`">{{ t('vocab.due') }}: {{ formatDateRelative(directionCard.next_due_at) }}</span>
                </div>
              </div>
            </div>
          </div>
          </div>
        </div>
        <div class="modal-footer">
          <div class="footer-actions">
            <button 
              v-if="hasUserCards" 
              @click="markKnown" 
              class="btn btn-primary"
              :disabled="processingAction"
            >
              {{ t('vocab.moveToKnown') }}
            </button>
            <button 
              v-if="!hasUserCards && isKnown" 
              @click="moveToTraining" 
              class="btn btn-primary"
              :disabled="processingAction"
            >
              {{ t('vocab.moveToTraining') }}
            </button>
            <button @click="confirmDelete" class="btn btn-danger" :disabled="processingAction">
              {{ t('vocab.removeFromVocabulary') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirm" class="modal" @click.self="showDeleteConfirm = false">
      <div class="modal-content">
        <h3>{{ t('vocab.removeFromVocabularyTitle') }}</h3>
        <p>{{ t('vocab.removeConfirm', { word: wordToDelete }) }}</p>
        <p class="warning-text">{{ t('vocab.removeWarning') }}</p>
        <div class="modal-actions">
          <button @click="deleteWord" class="btn btn-danger">{{ t('vocab.remove') }}</button>
          <button @click="showDeleteConfirm = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import { useAuth } from '../composables/useAuth'
import { showAlert } from '../composables/useDialog'

const { t } = useI18n()
const { isAdmin } = useAuth()

interface VocabWord {
  word_card_id: number
  lemma: string
  display_word: string
  total_cards: number
  due_count: number
  last_review: string | null
  total_reps: number
  added_at: string | null
  mastery_level: string
  mastering_score: number
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
  reps: number
  next_due_at: string | null
  last_review_at: string | null
  word_ru: string
  meaning_en: string
  example_en: string
  example_ru: string
  transcription: string
  sense_index: number
  pos?: string
  review_count: number
}

const words = ref<VocabWord[]>([])
const loading = ref(true)
const searchQuery = ref('')
const searchTimeout = ref<number | null>(null)
const statusFilter = ref<string>('')
const sortField = ref<string>('display_word')
const pagination = ref<Pagination>({
  page: 1,
  limit: 100,
  total: 0,
  total_pages: 0
})

const showDeleteConfirm = ref(false)
const wordToDelete = ref('')
const lemmaToDelete = ref('')

const showCardsModal = ref(false)
const selectedWord = ref('')
const selectedWordDisplay = ref('')
const selectedTranscription = ref('')
const cards = ref<CardDetail[]>([])
const cardsLoading = ref(false)
const verbForms = ref<any>(null)
const wordPOS = ref<string | null>(null)
const hasUserCards = ref(false)
const isKnown = ref(false)
const processingAction = ref(false)

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

const onSortChange = () => {
  pagination.value.page = 1
  loadVocab()
}

const onFilterChange = () => {
  pagination.value.page = 1
  loadVocab()
}

const loadVocab = async () => {
  loading.value = true
  try {
    // Map frontend sort field to backend sort_by
    let sortBy = 'display_word'
    let sortOrder = 'asc'
    
    if (sortField.value === 'display_word') {
      sortBy = 'display_word'
      sortOrder = 'asc'
    } else if (sortField.value === 'display_word_desc') {
      sortBy = 'display_word'
      sortOrder = 'desc'
    } else if (sortField.value === 'due_count') {
      sortBy = 'due_count'
      sortOrder = 'desc'
    } else if (sortField.value === 'added_at') {
      sortBy = 'added_at'
      sortOrder = 'desc'
    } else if (sortField.value === 'mastery_level') {
      sortBy = 'mastery_level'
      sortOrder = 'asc' // known -> mastered -> learning -> new
    } else if (sortField.value === 'mastery_level_desc') {
      sortBy = 'mastery_level_desc'
      sortOrder = 'asc' // new -> learning -> mastered -> known
    } else if (sortField.value === 'mastering_score') {
      sortBy = 'mastering_score'
      sortOrder = 'asc'
    } else if (sortField.value === 'mastering_score_desc') {
      sortBy = 'mastering_score_desc'
      sortOrder = 'asc'
    }
    
    const params = new URLSearchParams({
      page: pagination.value.page.toString(),
      limit: pagination.value.limit.toString(),
      sort_by: sortBy,
      sort_order: sortOrder
    })
    if (searchQuery.value) {
      params.append('search', searchQuery.value)
    }
    if (statusFilter.value) {
      params.append('mastery_level', statusFilter.value)
    }
    
    const data: { words: VocabWord[], pagination: Pagination } = await apiClient.request(`/api/vocab?${params.toString()}`)
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

// Red (0) → green (100) color for mastering score
const masteryColor = (score: number): string => {
  const s = Math.max(0, Math.min(100, score ?? 0))
  const hue = (120 * s) / 100 // 0 = red, 120 = green
  return `hsl(${hue}, 72%, 42%)`
}

// Alphabet sections for A→Z sorting
const cleanLemma = (lemma: string): string => {
  // Remove "to " prefix from verbs
  return lemma.replace(/^to\s+/i, '')
}

const alphabetSections = computed(() => {
  if (sortField.value !== 'display_word' && sortField.value !== 'display_word_desc') {
    return []
  }
  
  const sections: { letter: string; words: VocabWord[] }[] = []
  const letterMap = new Map<string, VocabWord[]>()
  
  // Group words by first letter of cleaned lemma
  words.value.forEach(word => {
    const cleaned = cleanLemma(word.lemma)
    const firstLetter = cleaned.charAt(0).toUpperCase()
    if (!letterMap.has(firstLetter)) {
      letterMap.set(firstLetter, [])
    }
    letterMap.get(firstLetter)!.push(word)
  })
  
  // Convert to array and sort based on sortField
  const letters = Array.from(letterMap.keys())
  if (sortField.value === 'display_word_desc') {
    letters.sort((a, b) => b.localeCompare(a)) // Z→A
  } else {
    letters.sort((a, b) => a.localeCompare(b)) // A→Z
  }
  
  letters.forEach(letter => {
    const sectionWords = letterMap.get(letter)!
    // Also sort words within each section by cleaned lemma
    if (sortField.value === 'display_word_desc') {
      sectionWords.sort((a, b) => cleanLemma(b.lemma).localeCompare(cleanLemma(a.lemma)))
    } else {
      sectionWords.sort((a, b) => cleanLemma(a.lemma).localeCompare(cleanLemma(b.lemma)))
    }
    
    sections.push({
      letter,
      words: sectionWords
    })
  })
  
  return sections
})

const confirmDelete = () => {
  wordToDelete.value = selectedWordDisplay.value
  lemmaToDelete.value = selectedWord.value
  showDeleteConfirm.value = true
}

const deleteWord = async () => {
  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/api/vocab/${lemmaToDelete.value}/delete`, formData)
    showDeleteConfirm.value = false
    showCardsModal.value = false
    await loadVocab()
  } catch (error) {
    console.error('Failed to delete word:', error)
    await showAlert('Failed to remove word from training')
  }
}

const showCards = async (lemma: string) => {
  selectedWord.value = lemma
  showCardsModal.value = true
  cardsLoading.value = true
  cards.value = []
  selectedTranscription.value = ''
  
  try {
    const data: { lemma: string; word_card_id: number; cards: CardDetail[]; verb_forms?: any; pos?: string; has_user_cards?: boolean; is_known?: boolean } = await apiClient.request(`/api/vocab/${lemma}/cards`)
    cards.value = data.cards || []
    verbForms.value = data.verb_forms || null
    wordPOS.value = data.pos || null
    hasUserCards.value = data.has_user_cards || false
    isKnown.value = data.is_known || false
    
    // Find display word and transcription from first card
    if (cards.value.length > 0) {
      const word = words.value.find(w => w.lemma === lemma)
      if (word) {
        selectedWordDisplay.value = cleanLemma(word.lemma)
      } else {
        selectedWordDisplay.value = cleanLemma(lemma)
      }
      selectedTranscription.value = cards.value[0].transcription || ''
    } else {
      selectedWordDisplay.value = cleanLemma(lemma)
    }
  } catch (error) {
    console.error('Failed to load cards:', error)
    await showAlert('Failed to load cards')
  } finally {
    cardsLoading.value = false
  }
}

const closeCardsModal = () => {
  showCardsModal.value = false
  selectedWord.value = ''
  selectedWordDisplay.value = ''
  selectedTranscription.value = ''
  cards.value = []
  verbForms.value = null
  wordPOS.value = null
  hasUserCards.value = false
  isKnown.value = false
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    if (showCardsModal.value) {
      event.preventDefault()
      closeCardsModal()
    } else if (showDeleteConfirm.value) {
      event.preventDefault()
      showDeleteConfirm.value = false
    }
  }
}

watch([showCardsModal, showDeleteConfirm], ([cardsOpen, deleteOpen]) => {
  if (cardsOpen || deleteOpen) {
    window.addEventListener('keydown', handleKeydown)
  } else {
    window.removeEventListener('keydown', handleKeydown)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

const markKnown = async () => {
  if (processingAction.value || !selectedWord.value) return
  
  processingAction.value = true
  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/api/vocab/${selectedWord.value}/mark_known`, formData)
    showCardsModal.value = false
    await loadVocab()
  } catch (error) {
    console.error('Failed to mark as known:', error)
    await showAlert('Failed to mark word as known')
  } finally {
    processingAction.value = false
  }
}

const moveToTraining = async () => {
  if (processingAction.value || !selectedWord.value) return
  
  processingAction.value = true
  try {
    const formData = new FormData()
    await apiClient.requestFormData(`/api/vocab/${selectedWord.value}/move_to_training`, formData)
    showCardsModal.value = false
    await loadVocab()
  } catch (error) {
    console.error('Failed to move to training:', error)
    await showAlert('Failed to move word to training')
  } finally {
    processingAction.value = false
  }
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

const totalCards = computed(() => cards.value.length)
const totalDue = computed(() => cards.value.filter(c => c.next_due_at && new Date(c.next_due_at) <= new Date()).length)
const lastReview = computed(() => {
  const reviews = cards.value
    .map(c => c.last_review_at)
    .filter((d): d is string => d !== null)
    .sort()
    .reverse()
  return reviews.length > 0 ? reviews[0] : null
})

const formatDate = (dateStr: string | null) => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const formatDateShort = (dateStr: string | null) => {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleDateString()
}

const formatDateRelative = (dateStr: string | null): string => {
  if (!dateStr) return '—'
  
  const date = new Date(dateStr)
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
        return t('vocab.justNow')
      }
      if (isFuture) {
        return t('vocab.inMinutes', diffMinutes, { n: diffMinutes })
      }
      return t('vocab.minutesAgo', diffMinutes, { n: diffMinutes })
    }
    if (isFuture) {
      return t('vocab.inHours', diffHours, { n: diffHours })
    }
    return t('vocab.hoursAgo', diffHours, { n: diffHours })
  }
  
  // Tomorrow / Yesterday
  if (diffDays === 1) {
    return isFuture ? t('vocab.tomorrow') : t('vocab.yesterday')
  }
  
  // Days
  if (diffDays < 7) {
    return isFuture ? t('vocab.inDays', diffDays, { n: diffDays }) : t('vocab.daysAgo', diffDays, { n: diffDays })
  }
  
  // Weeks
  const diffWeeks = Math.floor(diffDays / 7)
  if (diffWeeks < 4) {
    if (isFuture) {
      return t('vocab.inWeeks', diffWeeks, { n: diffWeeks })
    }
    return t('vocab.weeksAgo', diffWeeks, { n: diffWeeks })
  }
  
  // Months
  const diffMonths = Math.floor(diffDays / 30)
  if (diffMonths < 12) {
    if (isFuture) {
      return t('vocab.inMonths', diffMonths, { n: diffMonths })
    }
    return t('vocab.monthsAgo', diffMonths, { n: diffMonths })
  }
  
  // Years
  const diffYears = Math.floor(diffDays / 365)
  if (isFuture) {
    return t('vocab.inYears', diffYears, { n: diffYears })
  }
  return t('vocab.yearsAgo', diffYears, { n: diffYears })
}

const formatDateAbsolute = (dateStr: string | null): string => {
  if (!dateStr) return '—'
  
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '—'
  
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  
  return `${day}.${month}.${year}`
}

</script>

<style scoped>
.vocab {
  padding: 10px;
  max-width: 1200px;
  margin: 0 auto;
}

.vocab-header {
  margin-bottom: 24px;
}


.header-controls {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
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
  margin-bottom: 0;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-label {
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

.filter-select {
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  cursor: pointer;
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.sort-label {
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

.sort-select {
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  cursor: pointer;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.words-list {
  position: relative;
  min-height: 200px;
}

.loading-overlay {
  min-height: 400px;
}

.loading-overlay-content {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-primary);
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

.alphabet-section {
  margin-bottom: 32px;
}

.section-header {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 12px 0;
  padding-bottom: 8px;
  border-bottom: 2px solid var(--table-border, rgba(0, 0, 0, 0.1));
  color: var(--text-primary);
  position: sticky;
  top: 0;
  background: transparent;
  z-index: 5;
  padding-top: 8px;
}

.words-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 8px;
}

.word-card {
  background: var(--card-bg);
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  border-radius: 6px;
  padding: 10px 12px;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 0;
  overflow: hidden;
}

.word-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.word-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.word-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.word-display {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  word-break: break-word;
  overflow-wrap: break-word;
  hyphens: auto;
}

.mastery-marker {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-right: 6px;
}

.mastery-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 600;
  text-transform: capitalize;
  flex-shrink: 0;
}

.mastery-known {
  background-color: var(--color-success, #10b981);
  color: white;
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
  margin-top: 24px;
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
  max-width: 800px;
  width: 95%;
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

.word-summary {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
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

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
  justify-content: flex-end;
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

.cards-list-simple {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.directions-simple {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.direction-item-simple {
  padding: 12px;
  background: var(--card-bg);
  border-radius: 6px;
  border-left: 3px solid var(--color-primary);
}

.direction-header-simple {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
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

.direction-stats-simple {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--text-secondary);
  flex-wrap: wrap;
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

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background-color: var(--color-danger, #ef4444);
  color: white;
}

.btn-danger:hover {
  background-color: var(--color-danger-hover, #dc2626);
}

@media (max-width: 768px) {
  .vocab {
    padding: 10px;
  }
  
  .words-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px;
  }
  
  .header-controls {
    flex-direction: column;
    align-items: stretch;
  }
  
  .search-box {
    max-width: 100%;
  }
  
  .section-header {
    position: static;
  }
  
  .modal-content {
    padding: 20px;
    width: 95%;
  }
  
  .word-summary {
    flex-direction: column;
    gap: 4px;
  }
  
  .directions-simple {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }
}

@media (max-width: 400px) {
  .words-grid {
    grid-template-columns: 1fr;
  }
  
  .directions-simple {
    grid-template-columns: 1fr;
  }
}
</style>
