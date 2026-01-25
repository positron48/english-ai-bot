<template>
  <div class="admin-users-view">
    <div class="admin-users-header">
      <h1>Users</h1>
      <p class="admin-users-description">List of all registered users</p>
    </div>

    <div v-if="loading" class="admin-users-loading">
      <p>Loading users...</p>
    </div>

    <div v-else-if="error" class="admin-users-error">
      <p>{{ error }}</p>
    </div>

    <div v-else-if="users.length === 0" class="admin-users-empty">
      <p>No users found</p>
    </div>

    <div v-else class="admin-users-table-container">
      <table class="admin-users-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Telegram ID</th>
            <th>Username</th>
            <th>Registration Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.telegram_id }}</td>
            <td>{{ user.telegram_username || '-' }}</td>
            <td>{{ formatDate(user.created_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface User {
  id: number
  telegram_id: number
  telegram_username: string | null
  created_at: string
}

const users = ref<User[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const loadUsers = async () => {
  loading.value = true
  error.value = null
  
  try {
    const data: { users: User[] } = await apiClient.request('/api/admin/users')
    users.value = data.users
  } catch (err: any) {
    error.value = err.message || 'Failed to load users'
    console.error('Failed to load users:', err)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr: string | null | undefined): string => {
  if (!dateStr) return '—'
  
  // Handle SQL datetime format "2006-01-02 15:04:05"
  let date: Date
  try {
    // SQLite datetime format: "YYYY-MM-DD HH:MM:SS"
    // Parse manually to ensure correct local time interpretation
    const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})$/)
    if (match) {
      const [, year, month, day, hour, minute, second] = match.map(Number)
      date = new Date(year, month - 1, day, hour, minute, second || 0)
    } else {
      // Fallback to standard parsing
      date = new Date(dateStr)
    }
    
    // Check if date is valid (not NaN and not epoch 0)
    if (isNaN(date.getTime()) || date.getTime() === 0) {
      return '—'
    }
    
    // Format same way as CircuitBreakerView
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit',
      second: '2-digit'
    })
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
}

.admin-users-table tbody tr:hover {
  background: var(--bg-hover);
}

.admin-users-table tbody tr:last-child td {
  border-bottom: none;
}
</style>
