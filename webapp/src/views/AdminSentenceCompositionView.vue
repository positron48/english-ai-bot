<template>
  <div class="sc-admin">
    <h1>Составление предложений</h1>
    <p class="hint">Результаты режима по пользователям и ручной запуск генерации набора.</p>

    <div v-if="!selectedUser">
      <div class="card">
        <div class="toolbar">
          <button :disabled="loadingUsers" @click="loadUsers">
            {{ loadingUsers ? 'Загрузка…' : 'Обновить' }}
          </button>
          <span v-if="!enabled" class="warn">Режим выключен (SENTENCE_COMPOSITION_ENABLED=false) — генерация недоступна.</span>
        </div>
        <table v-if="users.length" class="grid">
          <thead>
            <tr>
              <th>Пользователь</th>
              <th>Тариф</th>
              <th>Наборов</th>
              <th>Последний</th>
              <th>★ / ✓ / ✗</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.user_id">
              <td>
                <span class="mono">#{{ u.user_id }}</span>
                <span v-if="u.telegram_username"> @{{ u.telegram_username }}</span>
              </td>
              <td>{{ u.subscription_tier }}</td>
              <td>{{ u.set_count }}</td>
              <td>{{ u.last_generation_on }}</td>
              <td class="mono">{{ u.total_stars }} / {{ u.total_passed }} / {{ u.total_failed }}</td>
              <td><button @click="openUser(u)">Открыть</button></td>
            </tr>
          </tbody>
        </table>
        <p v-else-if="!loadingUsers" class="muted">Нет пользователей с наборами предложений.</p>
      </div>
    </div>

    <div v-else>
      <div class="card">
        <div class="toolbar">
          <button @click="closeUser">← К списку</button>
          <strong>
            Пользователь #{{ selectedUser.user_id }}
            <span v-if="selectedUser.telegram_username">@{{ selectedUser.telegram_username }}</span>
          </strong>
          <button
            v-if="enabled"
            class="primary"
            :disabled="generating"
            @click="generate"
          >
            {{ generating ? 'Генерирую…' : 'Сгенерировать новый набор' }}
          </button>
          <span v-if="genMessage" :class="genError ? 'warn' : 'ok'">{{ genMessage }}</span>
        </div>
      </div>

      <div v-if="loadingDetail" class="muted">Загрузка…</div>

      <div v-for="set in sets" :key="set.id" class="card set-card">
        <div class="set-head">
          <div>
            <strong>{{ set.generation_date }}</strong>
            <span class="badge">{{ set.status }}</span>
            <span class="mono">{{ set.course_code }}</span>
          </div>
          <div class="mono set-stats">
            {{ set.attempted }}/{{ set.total }} · ★{{ set.stars }} ✓{{ set.passed }} ✗{{ set.failed }}
          </div>
        </div>
        <table class="grid items">
          <thead>
            <tr>
              <th>#</th>
              <th>Русский</th>
              <th>Эталон (ES)</th>
              <th>Ответ пользователя</th>
              <th>Итог</th>
              <th>Объяснение</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="it in set.items" :key="it.id" :class="outcomeClass(it)">
              <td>{{ it.position + 1 }}</td>
              <td>{{ it.prompt_ru }}</td>
              <td class="mono">{{ it.reference_es }}</td>
              <td class="mono">{{ it.user_input ?? '—' }}</td>
              <td>
                <span v-if="it.attempted">{{ it.outcome }} <span v-if="it.error_count != null">({{ it.error_count }})</span></span>
                <span v-else class="muted">не пройдено</span>
              </td>
              <td class="explain">{{ it.grading?.explanation || '' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <p v-if="!loadingDetail && !sets.length" class="muted">У пользователя ещё нет наборов.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiClient } from '../api/client'

interface UserOverview {
  user_id: number
  telegram_id: number
  telegram_username: string
  subscription_tier: string
  set_count: number
  last_generation_on: string
  total_stars: number
  total_passed: number
  total_failed: number
}

interface Item {
  id: number
  position: number
  prompt_ru: string
  reference_es: string
  user_input?: string
  attempted: boolean
  outcome?: string
  error_count?: number
  grading?: { explanation?: string; corrected_es?: string }
}

interface SetRow {
  id: number
  course_code: string
  generation_date: string
  status: string
  total: number
  attempted: number
  stars: number
  passed: number
  failed: number
  items: Item[]
}

const enabled = ref(true)
const users = ref<UserOverview[]>([])
const loadingUsers = ref(false)

const selectedUser = ref<UserOverview | null>(null)
const sets = ref<SetRow[]>([])
const loadingDetail = ref(false)

const generating = ref(false)
const genMessage = ref('')
const genError = ref(false)

async function loadUsers() {
  loadingUsers.value = true
  try {
    const res: any = await apiClient.request('/api/admin/sentence-composition/users')
    users.value = res.users || []
    enabled.value = !!res.enabled
  } catch (e: any) {
    users.value = []
  } finally {
    loadingUsers.value = false
  }
}

async function loadDetail(userID: number) {
  loadingDetail.value = true
  try {
    const res: any = await apiClient.request(`/api/admin/sentence-composition/users/${userID}`)
    sets.value = res.sets || []
    enabled.value = !!res.enabled
  } catch (e: any) {
    sets.value = []
  } finally {
    loadingDetail.value = false
  }
}

function openUser(u: UserOverview) {
  selectedUser.value = u
  genMessage.value = ''
  genError.value = false
  loadDetail(u.user_id)
}

function closeUser() {
  selectedUser.value = null
  sets.value = []
}

async function generate() {
  if (!selectedUser.value) return
  generating.value = true
  genMessage.value = ''
  genError.value = false
  try {
    const res: any = await apiClient.request(
      `/api/admin/sentence-composition/users/${selectedUser.value.user_id}/generate`,
      { method: 'POST' }
    )
    genError.value = false
    genMessage.value = `Готово: набор #${res.set_id}`
    await loadDetail(selectedUser.value.user_id)
  } catch (e: any) {
    genError.value = true
    genMessage.value = e?.message || 'Ошибка генерации'
  } finally {
    generating.value = false
  }
}

function outcomeClass(it: Item) {
  if (!it.attempted) return ''
  if (it.outcome === 'star') return 'row-star'
  if (it.outcome === 'passed') return 'row-passed'
  if (it.outcome === 'failed') return 'row-failed'
  return ''
}

onMounted(loadUsers)
</script>

<style scoped>
.sc-admin { max-width: 100%; }
.hint { color: var(--text-secondary); margin-top: -8px; }
.card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}
.toolbar { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
button {
  padding: 6px 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
}
button.primary { background: var(--accent-color, #3b82f6); color: #fff; border-color: transparent; }
button:disabled { opacity: 0.6; cursor: default; }
.grid { width: 100%; border-collapse: collapse; margin-top: 12px; }
.grid th, .grid td {
  text-align: left;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: top;
  font-size: 14px;
}
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }
.muted, .warn, .ok { font-size: 14px; }
.muted { color: var(--text-secondary); }
.warn { color: #dc2626; }
.ok { color: #16a34a; }
.badge {
  display: inline-block;
  padding: 1px 8px;
  margin: 0 8px;
  border-radius: 10px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  font-size: 12px;
}
.set-head { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; }
.set-stats { color: var(--text-secondary); }
.items td { font-size: 13px; }
.explain { color: var(--text-secondary); max-width: 280px; }
.row-star { background: rgba(22, 163, 74, 0.08); }
.row-passed { background: rgba(234, 179, 8, 0.08); }
.row-failed { background: rgba(220, 38, 38, 0.08); }
</style>
