<template>
  <div class="admin-content">
    <h2>Reading Texts</h2>

    <div class="toolbar">
      <input
        v-model="search"
        type="text"
        class="search-input"
        placeholder="Search by title..."
        @input="loadTexts"
      />
      <select v-model="level" class="level-select" @change="loadTexts">
        <option value="">All levels</option>
        <option v-for="value in levels" :key="value" :value="value">{{ value }}</option>
      </select>
      <button class="btn btn-primary" @click="loadTexts">Refresh</button>
    </div>

    <div v-if="loading" class="loading">Loading reading texts...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="texts.length === 0" class="empty-message">No texts found.</div>

    <div v-else class="table-wrap">
      <table class="texts-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Level</th>
            <th>Lang</th>
            <th>Category</th>
            <th>Segments</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="text in texts" :key="text.text_id">
            <td>{{ text.title }}</td>
            <td>{{ text.level || '-' }}</td>
            <td>{{ text.target_language || '-' }}</td>
            <td class="mono">{{ text.category_id || '-' }}</td>
            <td>{{ text.segments_count }}</td>
            <td>
              <button class="btn btn-danger" :disabled="deletingId === text.text_id" @click="deleteText(text)">
                {{ deletingId === text.text_id ? 'Deleting...' : 'Delete' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'

interface ReadingTextAdminItem {
  text_id: string
  category_id: string
  title: string
  level: string
  target_language: string
  segments_count: number
}

const levels = ['A0', 'A1', 'A2', 'B1', 'B2', 'C1']
const texts = ref<ReadingTextAdminItem[]>([])
const search = ref('')
const level = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const deletingId = ref<string | null>(null)

const loadTexts = async () => {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    if (search.value.trim()) {
      params.set('search', search.value.trim())
    }
    if (level.value.trim()) {
      params.set('level', level.value.trim())
    }
    const query = params.toString()
    const url = query ? `/api/admin/reading/texts?${query}` : '/api/admin/reading/texts'
    const data: { texts?: ReadingTextAdminItem[] } = await apiClient.request(url)
    texts.value = data.texts || []
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading texts'
  } finally {
    loading.value = false
  }
}

const deleteText = async (text: ReadingTextAdminItem) => {
  const ok = await showConfirm(`Delete "${text.title}" and all related files?`)
  if (!ok) return

  deletingId.value = text.text_id
  try {
    await apiClient.request(`/api/admin/reading/texts/${encodeURIComponent(text.text_id)}`, {
      method: 'DELETE',
    })
    await loadTexts()
    await showAlert('Reading text deleted.')
  } catch (e: any) {
    await showAlert(e?.message || 'Failed to delete reading text')
  } finally {
    deletingId.value = null
  }
}

onMounted(loadTexts)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.search-input,
.level-select {
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 10px;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.search-input {
  min-width: 260px;
}

.table-wrap {
  overflow-x: auto;
}

.texts-table {
  width: 100%;
  border-collapse: collapse;
}

.texts-table th,
.texts-table td {
  text-align: left;
  border-bottom: 1px solid var(--border-primary);
  padding: 10px 8px;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
}
</style>
