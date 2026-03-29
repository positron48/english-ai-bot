<template>
  <div class="admin-users-view">
    <div class="admin-users-header">
      <h1>Пользователи</h1>
      <p class="admin-users-description">Список зарегистрированных пользователей и уровень грамматики после placement (доступ к разделам курса)</p>
    </div>

    <div v-if="loading" class="admin-users-loading">
      <p>Загрузка…</p>
    </div>

    <div v-else-if="error" class="admin-users-error">
      <p>{{ error }}</p>
    </div>

    <div v-else-if="users.length === 0" class="admin-users-empty">
      <p>Пользователей нет</p>
    </div>

    <div v-else class="admin-users-table-container">
      <table class="admin-users-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Telegram ID</th>
            <th>Username</th>
            <th>Регистрация</th>
            <th>Грамматика (placement)</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.telegram_id }}</td>
            <td>{{ user.telegram_username || '—' }}</td>
            <td>{{ formatDate(user.created_at) }}</td>
            <td class="grammar-cell">
              <div v-if="user.grammar_placement" class="grammar-meta">
                <span class="grammar-level-label">{{ user.grammar_placement.level || '—' }}</span>
                <span v-if="user.grammar_placement.admin_set" class="admin-badge" title="Установлено вручную в админке">админ</span>
                <span class="grammar-score" v-if="user.grammar_placement.total_questions > 0">
                  тест: {{ user.grammar_placement.score }}%
                </span>
              </div>
              <div v-else class="grammar-meta muted">нет данных placement</div>
              <select
                class="grammar-select"
                v-model="levelDraft[user.id]"
                :disabled="saving[user.id] === true"
              >
                <option v-for="opt in levelOptions" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </select>
            </td>
            <td>
              <button
                type="button"
                class="save-btn"
                :disabled="saving[user.id] === true || !levelDirty(user)"
                @click="saveGrammarLevel(user.id)"
              >
                {{ saving[user.id] ? '…' : 'Сохранить' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface GrammarPlacement {
  level: string
  score: number
  total_questions: number
  completed_at: string
  admin_set: boolean
}

interface User {
  id: number
  telegram_id: number
  telegram_username: string | null
  created_at: string
  grammar_placement: GrammarPlacement | null
}

const levelOptions = [
  { value: '', label: '— сброс (нет placement)' },
  { value: 'below_a1', label: 'Ниже A1' },
  { value: 'A0', label: 'A0' },
  { value: 'A1', label: 'A1' },
  { value: 'A2', label: 'A2' },
  { value: 'B1', label: 'B1' },
  { value: 'B2', label: 'B2' },
  { value: 'C1', label: 'C1' },
  { value: 'C2', label: 'C2' },
] as const

const users = ref<User[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
/** Select value per user (API values: '', below_a1, A0…C2) */
const levelDraft = ref<Record<number, string>>({})
/** Original select value after last load (for dirty check) */
const levelBaseline = ref<Record<number, string>>({})
const saving = ref<Record<number, boolean>>({})

function placementToSelectValue(p: GrammarPlacement | null): string {
  if (!p || !p.level) {
    return ''
  }
  if (p.level === 'Below A1') {
    return 'below_a1'
  }
  return p.level
}

function syncDraftsFromUsers(list: User[]) {
  const d: Record<number, string> = {}
  const b: Record<number, string> = {}
  for (const u of list) {
    const v = placementToSelectValue(u.grammar_placement)
    d[u.id] = v
    b[u.id] = v
  }
  levelDraft.value = d
  levelBaseline.value = { ...b }
}

function levelDirty(user: User): boolean {
  return levelDraft.value[user.id] !== levelBaseline.value[user.id]
}

const loadUsers = async () => {
  loading.value = true
  error.value = null

  try {
    const data: { users: User[] } = await apiClient.request('/api/admin/users')
    users.value = data.users
    syncDraftsFromUsers(data.users)
  } catch (err: any) {
    error.value = err.message || 'Не удалось загрузить пользователей'
    console.error('Failed to load users:', err)
  } finally {
    loading.value = false
  }
}

const saveGrammarLevel = async (userId: number) => {
  saving.value = { ...saving.value, [userId]: true }
  error.value = null
  try {
    const level = levelDraft.value[userId] ?? ''
    await apiClient.request(`/api/admin/users/${userId}/grammar-placement`, {
      method: 'PUT',
      body: { level },
    })
    levelBaseline.value = { ...levelBaseline.value, [userId]: level }
    await loadUsers()
  } catch (err: any) {
    error.value = err.message || 'Ошибка сохранения уровня'
    console.error('saveGrammarLevel:', err)
  } finally {
    saving.value = { ...saving.value, [userId]: false }
  }
}

const formatDate = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '—'

  let date: Date
  try {
    const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})$/)
    if (match) {
      const [, year, month, day, hour, minute, second] = match.map(Number)
      date = new Date(year, month - 1, day, hour, minute, second || 0)
    } else {
      date = new Date(dateStr)
    }

    if (isNaN(date.getTime()) || date.getTime() === 0) {
      return '—'
    }

    return (
      date.toLocaleDateString() +
      ' ' +
      date.toLocaleTimeString([], {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      })
    )
  } catch (e) {
    console.error('Failed to parse date:', dateStr, e)
    return '—'
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.admin-users-view {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.admin-users-header {
  margin-bottom: 24px;
}

.admin-users-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
}

.admin-users-description {
  margin: 0;
  color: var(--text-secondary);
  font-size: 14px;
}

.admin-users-loading,
.admin-users-error,
.admin-users-empty {
  padding: 40px;
  text-align: center;
  color: var(--text-secondary);
}

.admin-users-error {
  color: var(--color-error, #d32f2f);
}

.admin-users-table-container {
  overflow-x: auto;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--bg-primary);
}

.admin-users-table {
  width: 100%;
  border-collapse: collapse;
}

.admin-users-table thead {
  background: var(--bg-secondary);
}

.admin-users-table th {
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
  border-bottom: 2px solid var(--border-primary);
}

.admin-users-table td {
  padding: 12px 16px;
  font-size: 14px;
  color: var(--text-primary);
  border-bottom: 1px solid var(--border-primary);
  vertical-align: top;
}

.admin-users-table tbody tr:hover {
  background: var(--bg-hover);
}

.admin-users-table tbody tr:last-child td {
  border-bottom: none;
}

.grammar-cell {
  min-width: 220px;
}

.grammar-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 13px;
}

.grammar-meta.muted {
  color: var(--text-secondary);
}

.grammar-level-label {
  font-weight: 600;
}

.admin-badge {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.grammar-score {
  color: var(--text-secondary);
  font-size: 12px;
}

.grammar-select {
  width: 100%;
  max-width: 260px;
  padding: 6px 8px;
  border-radius: 6px;
  border: 1px solid var(--border-primary);
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 13px;
}

.save-btn {
  padding: 6px 12px;
  font-size: 13px;
  border-radius: 6px;
  border: 1px solid var(--border-primary);
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
}

.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-btn:not(:disabled):hover {
  filter: brightness(1.05);
}
</style>
