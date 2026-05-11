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
        <label class="filter-controls" for="status-filter">
          <span class="filter-label">{{ t('vocab.status') }}</span>
          <select id="status-filter" v-model="statusFilter" @change="onFilterChange" class="filter-select">
            <option value="">{{ t('vocab.all') }}</option>
            <option value="new">{{ t('vocab.new') }}</option>
            <option value="learning">{{ t('vocab.learning') }}</option>
            <option value="mastered">{{ t('vocab.mastered') }}</option>
            <option value="known">{{ t('vocab.known') }}</option>
          </select>
        </label>
        <label class="sort-controls" for="sort-select">
          <span class="sort-label">{{ t('vocab.sortBy') }}</span>
          <select id="sort-select" v-model="sortField" @change="onSortChange" class="sort-select">
            <option value="display_word">A→Z</option>
            <option value="display_word_desc">Z→A</option>
            <option value="added_at">{{ t('vocab.recentlyAdded') }}</option>
            <option value="mastery_level">{{ t('vocab.mastery') }}</option>
            <option value="mastery_level_desc">{{ t('vocab.masteryReversed') }}</option>
            <option value="mastering_score">{{ t('vocab.masteringScore') }}</option>
            <option value="mastering_score_desc">{{ t('vocab.masteringScoreDesc') }}</option>
          </select>
        </label>
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
                        <span class="word-display">{{ word.display_target || word.display_word || cleanLemma(word.lemma) }}</span>
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
                      <span class="word-display">{{ word.display_target || word.display_word || cleanLemma(word.lemma) }}</span>
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

    <!-- Word detail (shared with Reading) -->
    <div v-if="showCardsModal" class="modal" @click.self="closeCardsModal">
      <div class="modal-content modal-large">
        <VocabWordCardsDetail
          :lemma="selectedWord"
          :list-mastering-score="selectedWordMasteringScore"
          @close="closeCardsModal"
          @vocab-changed="loadVocab"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import VocabWordCardsDetail from '../components/VocabWordCardsDetail.vue'

interface VocabWord {
  word_card_id: number
  lemma: string
  display_word: string
  display_target?: string
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

const { t } = useI18n()

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

const showCardsModal = ref(false)
const selectedWord = ref('')
const selectedWordMasteringScore = ref<number | null>(null)

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

const showCards = (lemma: string) => {
  const wordFromList = words.value.find((w) => w.lemma === lemma)
  selectedWordMasteringScore.value = wordFromList != null ? wordFromList.mastering_score : null
  selectedWord.value = lemma
  showCardsModal.value = true
}

const closeCardsModal = () => {
  showCardsModal.value = false
  selectedWord.value = ''
  selectedWordMasteringScore.value = null
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && showCardsModal.value) {
    event.preventDefault()
    closeCardsModal()
  }
}

watch(showCardsModal, (open) => {
  if (open) {
    window.addEventListener('keydown', handleKeydown)
  } else {
    window.removeEventListener('keydown', handleKeydown)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})

</script>

<style scoped>
.vocab {
  padding: 10px;
  max-width: 1200px;
  margin: 0 auto;
}

.vocab-header {
  margin-bottom: 24px;
  position: relative;
  z-index: 20;
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

.filter-controls,
.sort-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-weight: normal;
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
}

.filter-label,
.sort-label {
  font-size: 14px;
  color: var(--text-primary);
  white-space: nowrap;
}

.filter-select {
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 16px;
  line-height: 1.25;
  background-color: var(--input-bg);
  color: var(--text-primary);
  cursor: pointer;
  touch-action: manipulation;
  -webkit-appearance: menulist;
  appearance: menulist;
}

.sort-select {
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
  padding: 10px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 16px;
  line-height: 1.25;
  background-color: var(--input-bg);
  color: var(--text-primary);
  cursor: pointer;
  touch-action: manipulation;
  -webkit-appearance: menulist;
  appearance: menulist;
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

.transcription-with-audio {
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

.word-summary {
  display: flex;
  gap: 16px;
  font-size: 14px;
  color: var(--text-secondary);
  flex-wrap: wrap;
  align-items: center;
}

.mastering-score-inline {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: var(--text-primary);
}

.mastery-dot-inline {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
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

.verb-forms-table-wrap {
  overflow-x: auto;
  margin-top: 4px;
  border: 1px solid var(--border-primary, var(--table-border, rgba(0, 0, 0, 0.12)));
  border-radius: 8px;
  background: var(--card-bg, var(--input-bg, rgba(0, 0, 0, 0.02)));
}

.verb-forms-table {
  width: 100%;
  min-width: 520px;
  border-collapse: collapse;
  font-size: 13px;
  line-height: 1.35;
}

.verb-forms-table thead {
  background: var(--bg-tertiary, rgba(0, 0, 0, 0.06));
}

.verb-forms-table th,
.verb-forms-table td {
  border: 1px solid var(--border-primary, var(--table-border, rgba(0, 0, 0, 0.1)));
  padding: 8px 10px;
  text-align: left;
  vertical-align: middle;
}

.verb-forms-table th {
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
}

.verb-forms-table td {
  color: var(--text-primary);
}

.verb-forms-table tbody tr:nth-child(even) {
  background: var(--bg-secondary, rgba(0, 0, 0, 0.03));
}

.verb-forms-table tbody tr:hover {
  background: var(--hover-bg, rgba(0, 0, 0, 0.06));
}

.verb-forms-table td.verb-forms-pronoun {
  white-space: normal;
  max-width: 14rem;
  line-height: 1.35;
}

.verb-forms-table td:last-child {
  font-weight: 600;
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

.srs-info-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  margin-left: auto;
  cursor: help;
}

.srs-info-icon {
  width: 16px;
  height: 16px;
  opacity: 0.7;
  color: var(--text-secondary);
}

.srs-info-wrap:hover .srs-info-icon {
  opacity: 1;
  color: var(--color-primary);
}

/* Fixed tooltip (teleported to body) — not clipped by modal */
.srs-tooltip-fixed {
  position: fixed;
  transform: translate(-50%, -100%);
  min-width: 240px;
  max-width: 320px;
  padding: 10px 12px;
  background: var(--card-bg);
  border: 1px solid var(--table-border, rgba(0, 0, 0, 0.15));
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-primary);
  white-space: nowrap;
  z-index: 1100;
  pointer-events: auto;
}

.srs-tooltip-title {
  font-weight: 600;
  margin-bottom: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
}

.srs-tooltip-row {
  margin-bottom: 2px;
}

.srs-tooltip-row:last-child {
  margin-bottom: 0;
}

.srs-tooltip-row span {
  color: var(--text-secondary);
  margin-right: 6px;
}

.srs-tooltip-hint {
  cursor: help;
  opacity: 0.8;
}

.srs-tooltip-step {
  color: var(--color-primary);
  margin-right: 0;
}

.srs-tooltip-reason {
  margin-top: 8px;
  padding-top: 6px;
  border-top: 1px solid var(--table-border, rgba(0, 0, 0, 0.1));
  font-size: 11px;
  line-height: 1.4;
  color: var(--text-secondary);
  white-space: normal;
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
  
  .filter-controls,
  .sort-controls {
    flex-direction: column;
    align-items: stretch;
    gap: 6px;
    width: 100%;
  }

  .filter-label,
  .sort-label {
    white-space: normal;
  }

  .filter-select,
  .sort-select {
    width: 100%;
    min-height: 44px;
    padding: 12px 10px;
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
