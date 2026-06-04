<template>
  <div class="admin-content">
    <div class="card">
      <h2>Content Reports</h2>
      <div class="toolbar">
        <select v-model="statusFilter" class="admin-select" @change="loadReports">
          <option value="active">Active only</option>
          <option value="">All</option>
          <option value="resolved">Resolved only</option>
        </select>
        <select v-model="categoryFilter" class="admin-select" @change="loadReports">
          <option value="">All categories</option>
          <option v-for="c in categoryOptions" :key="c" :value="c">{{ c }}</option>
        </select>
        <button class="btn btn-primary" @click="loadReports">Refresh</button>
      </div>

      <div v-if="loading">Loading...</div>
      <div v-else-if="reports.length === 0" class="empty-message">No reports</div>
      <table v-else class="reports-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Status</th>
            <th>Source</th>
            <th>Category</th>
            <th>Word/Question</th>
            <th>Comment</th>
            <th>User</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="report in reports"
            :key="report.id"
            class="clickable-row"
            @click="openReport(report.id)"
          >
            <td>{{ report.id }}</td>
            <td><span :class="['status-badge', report.status]">{{ report.status }}</span></td>
            <td>{{ report.source_type }}</td>
            <td>{{ report.report_category || '—' }}</td>
            <td>{{ report.word || report.grammar_question_id || '—' }}</td>
            <td class="comment-cell">{{ truncateComment(report.comment_text) }}</td>
            <td>{{ report.user_id }}</td>
            <td>{{ formatDate(report.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showModal && selectedReport" class="modal" @click.self="closeModal">
      <div class="modal-content modal-large">
        <div class="modal-header">
          <h3>Report #{{ selectedReport.id }}</h3>
          <button class="btn-close" @click="closeModal">&times;</button>
        </div>
        <div class="modal-body">
          <div class="kv">
            <div><b>Status:</b> {{ selectedReport.status }}</div>
            <div><b>Source:</b> {{ selectedReport.source_type }}</div>
            <div v-if="selectedReport.report_category"><b>Category:</b> {{ selectedReport.report_category }}</div>
            <div><b>User:</b> {{ selectedReport.user_id }}</div>
            <div><b>Created:</b> {{ formatDate(selectedReport.created_at) }}</div>
            <div v-if="selectedReport.word"><b>Word:</b> {{ selectedReport.word }}</div>
            <div v-if="selectedReport.translation_direction"><b>Direction:</b> {{ selectedReport.translation_direction }}</div>
            <div v-if="selectedReport.word_category"><b>Category:</b> {{ selectedReport.word_category }}</div>
            <div v-if="selectedReport.grammar_chapter_id"><b>Chapter:</b> {{ selectedReport.grammar_chapter_id }}</div>
            <div v-if="selectedReport.theory_block_id"><b>Theory block:</b> {{ selectedReport.theory_block_id }}</div>
            <div v-if="selectedReport.grammar_question_id"><b>Question:</b> {{ selectedReport.grammar_question_id }}</div>
            <div v-if="selectedReport.comment_text" class="comment-full"><b>Comment:</b> {{ selectedReport.comment_text }}</div>
          </div>

          <div v-if="selectedReport.training_card" class="edit-card-section">
            <h4>Edit Training Card</h4>
            <div class="form-grid">
              <label>Word RU <input v-model="cardForm.word_ru" class="form-input" /></label>
              <label>Meaning EN <input v-model="cardForm.meaning_en" class="form-input" /></label>
              <label>Example EN <input v-model="cardForm.example_en" class="form-input" /></label>
              <label>Example RU <input v-model="cardForm.example_ru" class="form-input" /></label>
              <label>Hint <input v-model="cardForm.hint" class="form-input" /></label>
              <label>POS <input v-model="cardForm.pos" class="form-input" /></label>
              <label>Display Word <input v-model="cardForm.display_word" class="form-input" /></label>
            </div>
            <div class="row-actions">
              <button class="btn btn-primary" @click="saveTrainingCard">Save card</button>
            </div>
          </div>

          <div class="readonly-block">
            <h4>Report Payload</h4>
            <pre>{{ pretty(selectedReport.payload || {}) }}</pre>
          </div>
        </div>

        <div class="modal-actions">
          <button
            v-if="selectedReport.status === 'active'"
            class="btn btn-secondary"
            @click="resolveSelected"
          >
            Resolve report
          </button>
          <button class="btn btn-secondary" @click="closeModal">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { apiClient } from '../api/client'
import { showAlert } from '../composables/useDialog'

const statusFilter = ref('active')
const categoryFilter = ref('')
const categoryOptions = [
  'wrong_translation', 'wrong_example', 'wrong_distractors', 'typo', 'bad_audio', 'unclear_question',
  'wrong_answer', 'ambiguous', 'wrong_explanation', 'theory_mismatch', 'too_hard', 'other'
]
const loading = ref(false)
const reports = ref<any[]>([])
const showModal = ref(false)
const selectedReport = ref<any | null>(null)
const cardForm = ref({
  word_en: '',
  word_ru: '',
  meaning_en: '',
  example_en: '',
  example_ru: '',
  transcription: '',
  distractors_ru: '',
  distractors_en: '',
  hint: '',
  pos: '',
  display_word: ''
})

const loadReports = async () => {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (statusFilter.value) params.set('status', statusFilter.value)
    const qs = params.toString() ? `?${params.toString()}` : ''
    const data: any = await apiClient.request(`/api/admin/content-reports${qs}`)
    let rows = data.reports || []
    if (categoryFilter.value) {
      rows = rows.filter((r: any) => r.report_category === categoryFilter.value)
    }
    reports.value = rows
  } finally {
    loading.value = false
  }
}

const openReport = async (id: number) => {
  const data: any = await apiClient.request(`/api/admin/content-reports/${id}`)
  selectedReport.value = data
  if (data.training_card) {
    cardForm.value = {
      word_en: data.training_card.word_en || '',
      word_ru: data.training_card.word_ru || '',
      meaning_en: data.training_card.meaning_en || '',
      example_en: data.training_card.example_en || '',
      example_ru: data.training_card.example_ru || '',
      transcription: data.training_card.transcription || '',
      distractors_ru: data.training_card.distractors_ru || '',
      distractors_en: data.training_card.distractors_en || '',
      hint: data.training_card.hint || '',
      pos: data.training_card.pos || '',
      display_word: data.training_card.display_word || ''
    }
  }
  showModal.value = true
}

const saveTrainingCard = async () => {
  if (!selectedReport.value?.training_card?.id) return
  const params = new URLSearchParams()
  Object.entries(cardForm.value).forEach(([k, v]) => params.append(k, String(v || '')))
  await apiClient.request(`/api/admin/training/card/${selectedReport.value.training_card.id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: params.toString()
  })
  await showAlert('Card updated')
}

const resolveSelected = async () => {
  if (!selectedReport.value) return
  await apiClient.request(`/api/admin/content-reports/${selectedReport.value.id}/resolve`, { method: 'POST' })
  showModal.value = false
  selectedReport.value = null
  await loadReports()
}

const closeModal = () => {
  showModal.value = false
  selectedReport.value = null
}

const pretty = (v: any) => JSON.stringify(v, null, 2)
const formatDate = (v: string) => (v ? new Date(v).toLocaleString() : '—')
const truncateComment = (v: string) => {
  const s = String(v || '').trim()
  if (!s) return '—'
  return s.length > 120 ? `${s.slice(0, 117)}...` : s
}

void loadReports()
</script>

<style scoped>
.toolbar { display: flex; gap: 10px; margin-bottom: 16px; }
.reports-table { width: 100%; border-collapse: collapse; }
.reports-table th, .reports-table td { padding: 8px; border-bottom: 1px solid var(--border-primary); text-align: left; }
.comment-cell { max-width: 360px; white-space: normal; word-break: break-word; color: var(--text-secondary); }
.clickable-row { cursor: pointer; }
.clickable-row:hover { background: var(--bg-secondary); }
.status-badge { padding: 2px 8px; border-radius: 10px; font-size: 12px; }
.status-badge.active { background: rgba(220, 53, 69, 0.15); color: #d9534f; }
.status-badge.resolved { background: rgba(40, 167, 69, 0.15); color: #28a745; }
.modal { position: fixed; inset: 0; background: rgba(0,0,0,.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-content { background: var(--card-bg); border-radius: 10px; width: 92%; max-width: 980px; max-height: 90vh; overflow: auto; padding: 20px; }
.modal-large { max-width: 1100px; }
.modal-header { display: flex; justify-content: space-between; align-items: center; }
.btn-close { border: 0; background: transparent; font-size: 26px; cursor: pointer; color: var(--text-primary); }
.kv { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 8px 14px; margin-bottom: 14px; }
.comment-full { grid-column: 1 / -1; white-space: pre-wrap; }
.edit-card-section, .readonly-block { margin-top: 16px; }
.form-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 10px; }
.form-input { width: 100%; margin-top: 4px; }
.row-actions { margin-top: 10px; }
pre { white-space: pre-wrap; background: var(--bg-secondary); border: 1px solid var(--border-primary); padding: 10px; border-radius: 8px; }
.modal-actions { margin-top: 16px; display: flex; gap: 10px; justify-content: flex-end; }
</style>
