<template>
  <div class="lumi-facts-admin">
    <h1>Lumi Facts</h1>
    <div class="top-actions">
      <button @click="openPrompt">{{ copied ? 'Промпт скопирован ✓' : 'Промпт для генерации' }}</button>
    </div>

    <!-- Bulk add -->
    <div class="card add-card">
      <h2>Добавить факты</h2>
      <p class="hint">Один факт = один абзац (факты разделяются пустой строкой).</p>
      <div class="form-row">
        <label>Курс
          <select v-model="addForm.course_code">
            <option value="">(все курсы)</option>
            <option v-for="course in courseOptions" :key="course.code" :value="course.code">
              {{ course.title || course.code }}
            </option>
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

    <!-- JSON import -->
    <div class="card import-card">
      <h2>Импорт JSON</h2>
      <p class="hint">
        Вставьте массив фактов или объект <span class="mono">{ "facts": [...] }</span>.
        Поля: <span class="mono">course_code</span>, <span class="mono">context</span>,
        <span class="mono">locale</span>, <span class="mono">body</span>.
      </p>
      <textarea
        v-model="jsonImportText"
        rows="10"
        class="code-area"
        placeholder='{ "facts": [{ "course_code": "es_ru", "context": "grammar", "locale": "ru", "body": "..." }] }'
      ></textarea>
      <div class="actions">
        <button :disabled="importing || !jsonImportText.trim()" @click="importJson">
          {{ importing ? 'Импорт…' : 'Импортировать JSON' }}
        </button>
        <span v-if="importResult" :class="importOk ? 'ok' : 'error'">{{ importResult }}</span>
      </div>
    </div>

    <!-- Filters -->
    <div class="card">
      <div class="form-row">
        <label>Курс
          <select v-model="filter.course_code" @change="load">
            <option value="">все</option>
            <option v-for="course in courseOptions" :key="course.code" :value="course.code">
              {{ course.title || course.code }}
            </option>
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

    <!-- PROMPT MODAL -->
    <div v-if="promptOpen" class="modal-overlay" @click.self="promptOpen = false">
      <div class="modal">
        <h3>Промпт для генерации Lumi Facts</h3>
        <p class="hint">Скопируйте промпт в любую LLM, получите JSON и вставьте его в поле «Импорт JSON».</p>
        <textarea ref="promptArea" :value="promptText" readonly class="code-area" rows="18"></textarea>
        <div class="actions right">
          <button @click="promptOpen = false">Закрыть</button>
          <button @click="copyPrompt">{{ copied ? 'Скопировано ✓' : 'Скопировать' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { apiClient } from '../api/client'
import { courseClient, type CourseSummary } from '../api/courseClient'

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
const fallbackCourses: CourseSummary[] = [
  { id: 0, code: 'en_ru', title: 'en_ru', city_name: '', target_language: 'en', native_language: 'ru', ui_locale: 'ru', status: 'active', is_current: false },
  { id: 0, code: 'es_ru', title: 'es_ru', city_name: '', target_language: 'es', native_language: 'ru', ui_locale: 'ru', status: 'active', is_current: false },
]

const facts = ref<FactRow[]>([])
const availableCourses = ref<CourseSummary[]>([])
const courseOptions = computed(() => availableCourses.value.length ? availableCourses.value : fallbackCourses)
const total = ref(0)
const loading = ref(false)
const adding = ref(false)
const addResult = ref('')
const jsonImportText = ref('')
const importing = ref(false)
const importResult = ref('')
const importOk = ref(false)
const promptOpen = ref(false)
const promptText = ref('')
const copied = ref(false)
const promptArea = ref<HTMLTextAreaElement | null>(null)
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

const importJson = async () => {
  importResult.value = ''
  importOk.value = false
  let parsed: unknown
  try {
    parsed = JSON.parse(jsonImportText.value)
  } catch (e: any) {
    importResult.value = 'Невалидный JSON: ' + (e?.message || '')
    return
  }
  importing.value = true
  try {
    const res: any = await apiClient.request('/api/admin/lumi-facts', {
      method: 'POST',
      body: JSON.stringify(parsed),
    })
    importOk.value = true
    importResult.value = `Добавлено: ${res.inserted}`
    jsonImportText.value = ''
    await load()
  } catch (e: any) {
    importResult.value = e?.message || 'Ошибка импорта'
  } finally {
    importing.value = false
  }
}

const openPrompt = async () => {
  copied.value = false
  promptText.value = 'Загрузка…'
  promptOpen.value = true
  try {
    const data: { prompt?: string } = await apiClient.request('/api/admin/lumi-facts/prompt-template')
    promptText.value = data.prompt || ''
  } catch (e: any) {
    promptText.value = e?.message || 'Не удалось получить промпт'
  }
}

const copyPrompt = async () => {
  try {
    await navigator.clipboard.writeText(promptText.value)
  } catch {
    await nextTick()
    promptArea.value?.select()
    document.execCommand('copy')
  }
  copied.value = true
  setTimeout(() => { copied.value = false }, 1500)
}

const loadCourses = async () => {
  try {
    const data = await courseClient.getCourses()
    availableCourses.value = data.courses || []
    const current = availableCourses.value.find(c => c.is_current)
    if (current && addForm.value.course_code === 'es_ru') {
      addForm.value.course_code = current.code
    }
  } catch {
    availableCourses.value = []
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

onMounted(async () => {
  await loadCourses()
  await load()
})
</script>

<style scoped>
.lumi-facts-admin { max-width: 1100px; margin: 0 auto; padding: 16px; }
.top-actions { display: flex; justify-content: flex-end; margin: -4px 0 12px; }
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
.code-area { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 0.85rem; line-height: 1.45; }
.actions { margin-top: 10px; display: flex; gap: 12px; align-items: center; }
.actions.right { justify-content: flex-end; }
.ok { color: var(--color-success, #1e7e34); }
.error { color: var(--color-danger, #b91c1c); }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }
table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
th, td { text-align: left; padding: 6px 8px; border-bottom: 1px solid var(--border-primary, #eee); vertical-align: top; }
.body-cell { max-width: 480px; }
.archived { opacity: 0.5; }
.row-actions { white-space: nowrap; }
.row-actions button { margin-right: 4px; cursor: pointer; }
.total { margin-left: auto; color: var(--text-secondary); }
.empty { padding: 16px; color: var(--text-secondary); }
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px;
  z-index: 1000;
}
.modal {
  width: min(900px, 100%);
  max-height: 90vh;
  overflow: auto;
  background: var(--card-bg);
  border: 1px solid var(--border-primary, #ddd);
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 18px 60px rgba(0, 0, 0, 0.18);
}
</style>
