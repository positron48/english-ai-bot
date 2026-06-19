<template>
  <div class="lumi-facts-admin">
    <h1>Lumi Facts</h1>

    <!-- Bulk add -->
    <div class="card add-card">
      <h2>Добавить факты</h2>
      <p class="hint">Один факт = один абзац (факты разделяются пустой строкой).</p>
      <div class="form-row">
        <label>Курс
          <select v-model="addForm.course_code">
            <option value="">(все курсы)</option>
            <option value="es_ru">es_ru</option>
            <option value="en_ru">en_ru</option>
          </select>
        </label>
        <label>Контекст
          <select v-model="addForm.context">
            <option v-for="c in contexts" :key="c" :value="c">{{ c }}</option>
          </select>
        </label>
        <label>Локаль
          <select v-model="addForm.locale">
            <option value="ru">ru</option>
            <option value="en">en</option>
            <option value="es">es</option>
          </select>
        </label>
      </div>
      <textarea v-model="addForm.text" rows="8" placeholder="Факт 1...&#10;&#10;Факт 2..."></textarea>
      <div class="actions">
        <button :disabled="adding || !addForm.text.trim()" @click="bulkAdd">
          {{ adding ? 'Сохраняю…' : 'Добавить' }}
        </button>
        <span v-if="addResult" class="ok">{{ addResult }}</span>
      </div>
    </div>

    <!-- Filters -->
    <div class="card">
      <div class="form-row">
        <label>Курс
          <select v-model="filter.course_code" @change="load">
            <option value="">все</option>
            <option value="es_ru">es_ru</option>
            <option value="en_ru">en_ru</option>
          </select>
        </label>
        <label>Контекст
          <select v-model="filter.context" @change="load">
            <option value="">все</option>
            <option v-for="c in contexts" :key="c" :value="c">{{ c }}</option>
          </select>
        </label>
        <label>Статус
          <select v-model="filter.status" @change="load">
            <option value="">все</option>
            <option value="active">active</option>
            <option value="archived">archived</option>
          </select>
        </label>
        <span class="total">Всего: {{ total }}</span>
      </div>

      <table>
        <thead>
          <tr>
            <th>ID</th><th>Курс</th><th>Контекст</th><th>Текст</th>
            <th>Показан</th><th>Раз</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="f in facts" :key="f.id" :class="{ archived: f.status === 'archived' }">
            <td>{{ f.id }}</td>
            <td>{{ f.course_code || '—' }}</td>
            <td>{{ f.context }}</td>
            <td class="body-cell">
              <textarea v-if="editingId === f.id" v-model="editBody" rows="3"></textarea>
              <span v-else>{{ f.body }}</span>
            </td>
            <td>{{ f.last_shown_on || '—' }}</td>
            <td>{{ f.shown_count }}</td>
            <td class="row-actions">
              <template v-if="editingId === f.id">
                <button @click="saveEdit(f)">💾</button>
                <button @click="editingId = null">✕</button>
              </template>
              <template v-else>
                <button @click="startEdit(f)">✏️</button>
                <button v-if="f.status === 'active'" title="Архивировать" @click="setStatus(f, 'archived')">🗑</button>
                <button v-else title="Активировать" @click="setStatus(f, 'active')">↩️</button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && facts.length === 0" class="empty">Нет фактов</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiClient } from '../api/client'

interface FactRow {
  id: number
  course_code: string
  context: string
  locale: string
  body: string
  status: string
  last_shown_on?: string
  shown_count: number
}

const contexts = ['general', 'grammar', 'reading', 'practice', 'progress', 'city']

const facts = ref<FactRow[]>([])
const total = ref(0)
const loading = ref(false)
const adding = ref(false)
const addResult = ref('')
const editingId = ref<number | null>(null)
const editBody = ref('')

const filter = ref({ course_code: '', context: '', status: '' })
const addForm = ref({ course_code: 'es_ru', context: 'general', locale: 'ru', text: '' })

const load = async () => {
  loading.value = true
  try {
    const p = new URLSearchParams()
    for (const [k, v] of Object.entries(filter.value)) if (v) p.set(k, v)
    const res: any = await apiClient.request(`/api/admin/lumi-facts?${p.toString()}`)
    facts.value = res.facts || []
    total.value = res.total || 0
  } catch { /* shown by global toast */ } finally {
    loading.value = false
  }
}

const bulkAdd = async () => {
  adding.value = true
  addResult.value = ''
  try {
    const res: any = await apiClient.request('/api/admin/lumi-facts', {
      method: 'POST',
      body: JSON.stringify(addForm.value),
    })
    addResult.value = `Добавлено: ${res.inserted}`
    addForm.value.text = ''
    await load()
  } catch (e: any) {
    addResult.value = e?.message || 'Ошибка'
  } finally {
    adding.value = false
  }
}

const startEdit = (f: FactRow) => {
  editingId.value = f.id
  editBody.value = f.body
}

const update = async (f: FactRow) => {
  await apiClient.request('/api/admin/lumi-facts', {
    method: 'PUT',
    body: JSON.stringify(f),
  })
  await load()
}

const saveEdit = async (f: FactRow) => {
  await update({ ...f, body: editBody.value })
  editingId.value = null
}

const setStatus = (f: FactRow, status: string) => update({ ...f, status })

onMounted(load)
</script>

<style scoped>
.lumi-facts-admin { max-width: 1100px; margin: 0 auto; padding: 16px; }
.card {
  background: var(--card-bg);
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}
.hint { color: var(--text-secondary); font-size: 0.85rem; margin: 4px 0 10px; }
.form-row { display: flex; gap: 16px; align-items: center; flex-wrap: wrap; margin-bottom: 10px; }
.form-row label { display: flex; flex-direction: column; gap: 4px; font-size: 0.85rem; }
textarea { width: 100%; box-sizing: border-box; font: inherit; padding: 8px; }
.actions { margin-top: 10px; display: flex; gap: 12px; align-items: center; }
.ok { color: var(--color-success, #1e7e34); }
table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid var(--border-primary, #eee); vertical-align: top; }
.body-cell { max-width: 480px; }
.archived { opacity: 0.5; }
.row-actions { white-space: nowrap; }
.row-actions button { margin-right: 4px; cursor: pointer; }
.total { margin-left: auto; color: var(--text-secondary); }
.empty { padding: 16px; color: var(--text-secondary); }
</style>
