<template>
  <div class="admin">
    <AdminMenu />
    
    <!-- Orphaned Training Cards Block -->
    <div class="card">
      <h2>Orphaned Training Cards</h2>
      <p class="description">Training cards that reference non-existent word cards (likely left after word deletion)</p>
      
      <div class="cards-content">
        <div v-if="trainingCardsError && !trainingCardsLoading" class="empty-message">
          <p>{{ trainingCardsError }}</p>
        </div>
        <div v-else-if="trainingCards.length === 0 && !trainingCardsLoading" class="empty-message">
          <p>No orphaned training cards found</p>
        </div>
        <div v-else class="cards-table-container" :class="{ 'loading-overlay': trainingCardsLoading }">
          <div v-if="trainingCardsLoading" class="loading-overlay-content">
            <div class="loading">Loading cards...</div>
          </div>
          <table class="cards-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Word EN</th>
                <th>Word RU</th>
                <th>Meaning EN</th>
                <th>Sense</th>
                <th>Word Card ID</th>
                <th>User Cards</th>
                <th>Created At</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="card in trainingCards" :key="card.id">
                <td>{{ card.id }}</td>
                <td>
                  <strong>{{ card.word_en }}</strong>
                  <span v-if="card.transcription" class="transcription">{{ card.transcription }}</span>
                  <span v-if="card.display_word" class="display-word">({{ card.display_word }})</span>
                </td>
                <td>{{ card.word_ru }}</td>
                <td class="meaning-cell">{{ card.meaning_en }}</td>
                <td>
                  <span class="badge badge-sense">{{ card.sense_index + 1 }}</span>
                </td>
                <td>
                  <span class="invalid-id">{{ card.word_card_id }}</span>
                </td>
                <td>{{ card.user_cards_count }}</td>
                <td :title="formatDateAbsolute(card.created_at)">{{ formatDateRelative(card.created_at) }}</td>
                <td>
                  <button 
                    @click="deleteTrainingCard(card)" 
                    class="btn btn-sm btn-danger"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <div class="pagination" v-if="trainingCardsPagination.total_pages > 1 && !trainingCardsLoading">
        <button 
          @click="goToTrainingCardsPage(trainingCardsPagination.page - 1)" 
          :disabled="trainingCardsPagination.page <= 1"
          class="btn btn-secondary"
        >
          Previous
        </button>
        <span class="page-info">
          Page {{ trainingCardsPagination.page }} of {{ trainingCardsPagination.total_pages }} 
          ({{ trainingCardsPagination.total }} total)
        </span>
        <button 
          @click="goToTrainingCardsPage(trainingCardsPagination.page + 1)" 
          :disabled="trainingCardsPagination.page >= trainingCardsPagination.total_pages"
          class="btn btn-secondary"
        >
          Next
        </button>
      </div>
    </div>

    <!-- Orphaned User Cards Block -->
    <div class="card">
      <h2>Orphaned User Cards</h2>
      <p class="description">User cards that reference non-existent training cards. These cards appear in dashboard but cannot be used in training.</p>
      
      <div class="cards-content">
        <div v-if="userCardsError && !userCardsLoading" class="empty-message">
          <p>{{ userCardsError }}</p>
        </div>
        <div v-else-if="userCards.length === 0 && !userCardsLoading" class="empty-message">
          <p>No orphaned user cards found</p>
        </div>
        <div v-else class="cards-table-container" :class="{ 'loading-overlay': userCardsLoading }">
          <div v-if="userCardsLoading" class="loading-overlay-content">
            <div class="loading">Loading cards...</div>
          </div>
          <table class="cards-table">
            <thead>
              <tr>
                <th>User Card ID</th>
                <th>User ID</th>
                <th>Telegram ID</th>
                <th>Username</th>
                <th>Training Card ID</th>
                <th>Direction</th>
                <th>State</th>
                <th>Reps</th>
                <th>Review Events</th>
                <th>Created At</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="card in userCards" :key="card.user_card_id">
                <td>{{ card.user_card_id }}</td>
                <td>{{ card.user_id }}</td>
                <td>{{ card.telegram_id }}</td>
                <td>{{ card.telegram_username || '—' }}</td>
                <td>
                  <span class="invalid-id">{{ card.training_card_id }}</span>
                </td>
                <td>{{ card.direction }}</td>
                <td>
                  <span class="badge badge-state" :class="`badge-${card.state}`">{{ card.state }}</span>
                </td>
                <td>{{ card.reps }}</td>
                <td>{{ card.review_events_count }}</td>
                <td :title="formatDateAbsolute(card.created_at)">{{ formatDateRelative(card.created_at) }}</td>
                <td>
                  <button 
                    @click="deleteUserCard(card)" 
                    class="btn btn-sm btn-danger"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <div class="pagination" v-if="userCardsPagination.total_pages > 1 && !userCardsLoading">
        <button 
          @click="goToUserCardsPage(userCardsPagination.page - 1)" 
          :disabled="userCardsPagination.page <= 1"
          class="btn btn-secondary"
        >
          Previous
        </button>
        <span class="page-info">
          Page {{ userCardsPagination.page }} of {{ userCardsPagination.total_pages }} 
          ({{ userCardsPagination.total }} total)
        </span>
        <button 
          @click="goToUserCardsPage(userCardsPagination.page + 1)" 
          :disabled="userCardsPagination.page >= userCardsPagination.total_pages"
          class="btn btn-secondary"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import AdminMenu from '../components/AdminMenu.vue'

interface OrphanedTrainingCard {
  id: number
  word_card_id: number
  word_en: string
  transcription?: string
  sense_index: number
  word_ru: string
  meaning_en: string
  example_en?: string
  example_ru?: string
  pos?: string | null
  display_word?: string | null
  created_at: string
  user_cards_count: number
}

interface OrphanedUserCard {
  user_card_id: number
  user_id: number
  telegram_id: number
  telegram_username?: string | null
  training_card_id: number
  direction: string
  state: string
  reps: number
  created_at: string
  review_events_count: number
}

// Training cards state
const trainingCardsLoading = ref(false)
const trainingCardsError = ref<string | null>(null)
const trainingCards = ref<OrphanedTrainingCard[]>([])
const trainingCardsPagination = ref({
  page: 1,
  limit: 50,
  total: 0,
  total_pages: 0
})

// User cards state
const userCardsLoading = ref(false)
const userCardsError = ref<string | null>(null)
const userCards = ref<OrphanedUserCard[]>([])
const userCardsPagination = ref({
  page: 1,
  limit: 50,
  total: 0,
  total_pages: 0
})

onMounted(async () => {
  await Promise.all([loadTrainingCards(), loadUserCards()])
})

const loadTrainingCards = async () => {
  trainingCardsLoading.value = true
  trainingCardsError.value = null
  try {
    const offset = (trainingCardsPagination.value.page - 1) * trainingCardsPagination.value.limit
    const params = new URLSearchParams()
    params.append('limit', trainingCardsPagination.value.limit.toString())
    params.append('offset', offset.toString())

    const data: { 
      cards: OrphanedTrainingCard[]
      pagination: { page: number; limit: number; total: number; total_pages: number }
    } = await apiClient.request(`/app/admin/orphaned-cards?${params.toString()}`)
    
    trainingCards.value = data.cards || []
    if (data.pagination) {
      trainingCardsPagination.value = data.pagination
    } else {
      trainingCardsPagination.value = {
        page: 1,
        limit: 50,
        total: 0,
        total_pages: 0
      }
    }
  } catch (err: any) {
    console.error('Failed to load orphaned training cards:', err)
    trainingCardsError.value = err.message || 'Failed to load orphaned training cards'
    trainingCards.value = []
  } finally {
    trainingCardsLoading.value = false
  }
}

const loadUserCards = async () => {
  userCardsLoading.value = true
  userCardsError.value = null
  try {
    const offset = (userCardsPagination.value.page - 1) * userCardsPagination.value.limit
    const params = new URLSearchParams()
    params.append('limit', userCardsPagination.value.limit.toString())
    params.append('offset', offset.toString())

    const data: { 
      cards: OrphanedUserCard[]
      pagination: { page: number; limit: number; total: number; total_pages: number }
    } = await apiClient.request(`/app/admin/orphaned-user-cards?${params.toString()}`)
    
    userCards.value = data.cards || []
    if (data.pagination) {
      userCardsPagination.value = data.pagination
    } else {
      userCardsPagination.value = {
        page: 1,
        limit: 50,
        total: 0,
        total_pages: 0
      }
    }
  } catch (err: any) {
    console.error('Failed to load orphaned user cards:', err)
    userCardsError.value = err.message || 'Failed to load orphaned user cards'
    userCards.value = []
  } finally {
    userCardsLoading.value = false
  }
}

const goToTrainingCardsPage = (page: number) => {
  if (page >= 1 && page <= trainingCardsPagination.value.total_pages) {
    trainingCardsPagination.value.page = page
    loadTrainingCards()
  }
}

const goToUserCardsPage = (page: number) => {
  if (page >= 1 && page <= userCardsPagination.value.total_pages) {
    userCardsPagination.value.page = page
    loadUserCards()
  }
}

const deleteTrainingCard = async (card: OrphanedTrainingCard) => {
  const confirmed = await showConfirm(
    `Are you sure you want to delete orphaned training card #${card.id}?\n\n` +
    `Word: ${card.word_en}\n` +
    `This will delete:\n` +
    `- The training card\n` +
    `- ${card.user_cards_count} user card(s) and their review events\n\n` +
    `This action cannot be undone!`
  )
  if (!confirmed) {
    return
  }

  try {
    await apiClient.request(`/app/admin/orphaned-cards/${card.id}`, { method: 'DELETE' })
    // Reload both lists to update pagination
    await Promise.all([loadTrainingCards(), loadUserCards()])
  } catch (err: any) {
    console.error('Failed to delete training card:', err)
    await showAlert(err.message || 'Failed to delete training card')
  }
}

const deleteUserCard = async (card: OrphanedUserCard) => {
  const confirmed = await showConfirm(
    `Are you sure you want to delete orphaned user card #${card.user_card_id}?\n\n` +
    `User ID: ${card.user_id}\n` +
    `Training Card ID: ${card.training_card_id} (invalid)\n` +
    `State: ${card.state}\n` +
    `This will delete:\n` +
    `- The user card\n` +
    `- ${card.review_events_count} review event(s)\n\n` +
    `This action cannot be undone!`
  )
  if (!confirmed) {
    return
  }

  try {
    await apiClient.request(`/app/admin/orphaned-user-cards/${card.user_card_id}`, { method: 'DELETE' })
    // Reload to update pagination
    await loadUserCards()
  } catch (err: any) {
    console.error('Failed to delete user card:', err)
    await showAlert(err.message || 'Failed to delete user card')
  }
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
</script>

<style scoped>
.admin {
  max-width: 1200px;
  margin: 0 auto;
  padding: 10px;
}

.admin h1 {
  margin-bottom: 24px;
}

.admin .card h2 {
  margin-bottom: 20px;
}

.description {
  color: var(--text-secondary);
  margin-bottom: 20px;
  font-size: 14px;
}

.cards-content {
  position: relative;
}

.cards-table-container {
  overflow-x: auto;
  margin-top: 20px;
  position: relative;
}

.cards-table-container.loading-overlay {
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

.cards-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.cards-table th,
.cards-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid var(--border-primary);
}

.cards-table th {
  background: var(--bg-secondary);
  font-weight: 600;
  color: var(--text-primary);
  position: sticky;
  top: 0;
}

.cards-table tbody tr:hover {
  background: var(--bg-secondary);
}

.transcription {
  color: var(--text-secondary);
  font-style: italic;
  font-size: 12px;
  margin-left: 8px;
}

.display-word {
  color: var(--text-secondary);
  font-size: 12px;
  margin-left: 4px;
}

.meaning-cell {
  max-width: 300px;
  word-break: break-word;
}

.badge {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.badge-sense {
  background: rgba(108, 117, 125, 0.1);
  color: var(--text-secondary);
}

.invalid-id {
  color: var(--color-danger);
  font-weight: 500;
}

.badge-state {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.badge-new {
  background: rgba(0, 123, 255, 0.1);
  color: #007bff;
}

.badge-learning {
  background: rgba(255, 193, 7, 0.1);
  color: #ffc107;
}

.badge-review {
  background: rgba(40, 167, 69, 0.1);
  color: #28a745;
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

.empty-message {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.btn-danger {
  background: var(--color-danger);
  color: white;
  border: none;
}

.btn-danger:hover {
  background: var(--color-danger-hover, #c82333);
  opacity: 0.9;
}

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
}

.btn-secondary:hover {
  background: var(--bg-hover);
}

.btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
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
