<template>
  <div class="admin-content">
    <h2>Картинки (Picture Quests)</h2>

    <div class="course-selector">
      <label for="pq-course">Курс:</label>
      <select id="pq-course" v-model="selectedCourseCode" class="level-select">
        <option disabled value="">Выберите курс</option>
        <option v-for="course in availableCourses" :key="course.code" :value="course.code">
          {{ course.title || course.code }}
        </option>
      </select>
      <button class="btn btn-primary" :disabled="!selectedCourseCode" @click="newQuest">+ Новый квест</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="openImport">Импорт JSON</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="openPrompt">Промпт для генерации</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="load()">Обновить</button>
    </div>

    <div v-if="coursesError" class="error">{{ coursesError }}</div>
    <div v-if="loading" class="loading">Загрузка…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="selectedCourseCode && !quests.length" class="empty-message">Квестов нет.</div>

    <!-- TOOLBAR: filter + sort -->
    <div v-if="!loading && quests.length" class="pq-toolbar">
      <label>Статус:
        <select v-model="statusFilter" class="level-select">
          <option value="">все</option>
          <option value="draft">draft</option>
          <option value="active">active</option>
          <option value="locked">locked</option>
          <option value="archived">archived</option>
        </select>
      </label>
      <label>Сортировка по id:
        <select v-model="sortDir" class="level-select">
          <option value="asc">по возрастанию ↑</option>
          <option value="desc">по убыванию ↓</option>
        </select>
      </label>
      <span class="pq-count">Найдено: {{ filteredQuests.length }}</span>
    </div>

    <!-- QUEST LIST (compact) -->
    <div v-if="!loading && quests.length" class="pq-list">
      <div v-if="!filteredQuests.length" class="empty-message">Нет квестов с выбранным статусом.</div>
      <template v-else>
        <div v-for="q in pagedQuests" :key="q.id" class="pq-card">
          <div class="pq-row">
            <button class="pq-expand" :class="{ open: expanded.has(q.id) }" @click="toggleExpand(q.id)" title="Задания">▶</button>
            <label
              class="pq-thumb-upload"
              :class="{ 'pq-thumb-upload--empty': !q.image_url, busy: uploadingId === q.id }"
              :title="q.image_url ? 'Заменить картинку' : 'Загрузить картинку'"
            >
              <img v-if="q.image_url" :src="mediaUrl(q.image_url)" class="pq-thumb" alt="" />
              <span v-if="uploadingId === q.id" class="pq-thumb-state">…</span>
              <span v-else-if="!q.image_url" class="pq-thumb-state">＋</span>
              <span v-else class="pq-thumb-replace">↻</span>
              <input type="file" accept="image/png,image/jpeg,image/webp" style="display:none"
                :disabled="uploadingId === q.id" @change="uploadPictureForQuest(q, $event)" />
            </label>
            <div class="pq-main">
              <span class="pq-title">{{ q.title }}</span>
              <span class="badge" :class="'badge--' + q.status">{{ q.status }}</span>
              <span class="pq-id mono">#{{ q.id }}</span>
            </div>
            <span class="pq-tasks-count">{{ q.tasks.length }} заданий</span>
            <div class="pq-actions">
              <button
                v-if="q.status === 'draft' && q.image_url"
                class="btn btn-sm btn-primary"
                :disabled="publishingId === q.id"
                @click="publishQuest(q)"
              >{{ publishingId === q.id ? '…' : 'Опубликовать' }}</button>
              <button class="btn btn-sm" @click="editQuest(q)">Изменить</button>
              <button class="btn btn-sm btn-danger" @click="removeQuest(q)">Удалить</button>
            </div>
          </div>

          <!-- expanded: meta + tasks -->
          <div v-if="expanded.has(q.id)" class="pq-details">
            <div class="scenario-meta mono">
              {{ q.code }} · {{ q.cefr_level }}
              · {{ q.max_turns }} ходов · {{ q.token_budget }} токенов · order {{ q.sort_order }}
            </div>
            <div v-if="q.image_description" class="scenario-meta description">{{ q.image_description }}</div>
            <div class="tasks-block">
              <div class="tasks-head">
                <span>Задания квеста ({{ q.tasks.length }})</span>
                <button class="btn btn-sm" @click="newTask(q)">+ Задание</button>
              </div>
              <table v-if="q.tasks.length" class="tasks-table">
                <thead>
                  <tr><th>code</th><th>Название</th><th>Критерий выполнения</th><th>req</th><th>ord</th><th></th></tr>
                </thead>
                <tbody>
                  <tr v-for="t in q.tasks" :key="t.id">
                    <td class="mono">{{ t.code }}</td>
                    <td>{{ t.title }}</td>
                    <td class="criteria">{{ t.completion_criteria }}</td>
                    <td>{{ t.is_required ? '✓' : '—' }}</td>
                    <td>{{ t.sort_order }}</td>
                    <td class="nowrap">
                      <button class="btn btn-xs" @click="editTask(q, t)">ред.</button>
                      <button class="btn btn-xs btn-danger" @click="removeTask(q, t)">×</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- PAGINATION -->
        <div v-if="totalPages > 1" class="pq-pagination">
          <button class="btn btn-sm" :disabled="page <= 1" @click="page--">← Назад</button>
          <span class="pq-page-info">Стр. {{ page }} из {{ totalPages }}</span>
          <button class="btn btn-sm" :disabled="page >= totalPages" @click="page++">Вперёд →</button>
        </div>
      </template>
    </div>

    <!-- QUEST EDIT MODAL -->
    <div v-if="questForm" class="modal-overlay" @click.self="questForm = null">
      <div class="modal">
        <h3>{{ questForm.id ? 'Изменить квест' : 'Новый квест' }}</h3>
        <div class="form-grid">
          <label>code <input v-model="questForm.code" class="inp mono" /></label>
          <label>CEFR уровень
            <select v-model="questForm.cefr_level" class="inp">
              <option v-for="l in levels" :key="l.level_code" :value="l.level_code">{{ l.level_code }} — {{ l.title }}</option>
            </select>
          </label>
          <label>Название <input v-model="questForm.title" class="inp" /></label>
          <label>Статус
            <select v-model="questForm.status" class="inp">
              <option value="draft">draft</option>
              <option value="active">active</option>
              <option value="locked">locked</option>
              <option value="archived">archived</option>
            </select>
          </label>
          <label>Макс. ходов <input v-model.number="questForm.max_turns" type="number" class="inp" /></label>
          <label>Бюджет токенов <input v-model.number="questForm.token_budget" type="number" class="inp" /></label>
          <label>Порядок <input v-model.number="questForm.sort_order" type="number" class="inp" /></label>
        </div>
        <label class="full">Картинка (URL или загрузите файл)
          <div class="image-field">
            <input v-model="questForm.image_url" class="inp mono" placeholder="https://..." />
            <label class="btn btn-sm file-btn">
              Загрузить
              <input type="file" accept="image/png,image/jpeg,image/webp" style="display:none" @change="uploadPicture($event)" />
            </label>
          </div>
          <img v-if="questForm.image_url" :src="mediaUrl(questForm.image_url)" class="image-preview" alt="" />
        </label>
        <label class="full">Описание картинки (инструкция для ИИ, на англ. — ВСЕ факты: объекты, цвета, количество, расположение)
          <textarea v-model="questForm.image_description" class="inp" rows="5"></textarea>
        </label>
        <div class="modal-actions">
          <button class="btn" @click="questForm = null">Отмена</button>
          <button class="btn btn-primary" :disabled="saving" @click="saveQuest">{{ saving ? 'Сохранение…' : 'Сохранить' }}</button>
        </div>
      </div>
    </div>

    <!-- TASK EDIT MODAL -->
    <div v-if="taskForm" class="modal-overlay" @click.self="taskForm = null">
      <div class="modal">
        <h3>{{ taskForm.id ? 'Изменить задание' : 'Новое задание' }}</h3>
        <div class="form-grid">
          <label>code <input v-model="taskForm.code" class="inp mono" /></label>
          <label>Порядок <input v-model.number="taskForm.sort_order" type="number" class="inp" /></label>
          <label>Название <input v-model="taskForm.title" class="inp" /></label>
          <label class="check"><input type="checkbox" v-model="taskForm.is_required" /> Обязательное</label>
        </div>
        <label class="full">Критерий выполнения (инструкция для ИИ, на англ.)
          <textarea v-model="taskForm.completion_criteria" class="inp" rows="3"></textarea>
        </label>
        <div class="modal-actions">
          <button class="btn" @click="taskForm = null">Отмена</button>
          <button class="btn btn-primary" :disabled="saving" @click="saveTask">{{ saving ? 'Сохранение…' : 'Сохранить' }}</button>
        </div>
      </div>
    </div>

    <!-- IMPORT MODAL -->
    <div v-if="importOpen" class="modal-overlay" @click.self="importOpen = false">
      <div class="modal">
        <h3>Импорт квеста из JSON</h3>
        <p class="hint">
          Вставьте JSON одного квеста или массива квестов с заданиями (структура — как из кнопки «Промпт для генерации»).
          Квесты с тем же <span class="mono">code</span> в курсе <span class="mono">{{ selectedCourseCode }}</span>
          будут перезаписаны вместе с заданиями.
          Пустой <span class="mono">image_url</span> в JSON не затирает уже загруженную картинку.
        </p>
        <textarea ref="importArea" v-model="importText" class="inp code-area" rows="16" placeholder='{ "code": "...", "title": "...", "tasks": [ ... ] } или [ { ... }, { ... } ]'></textarea>
        <div v-if="importError" class="error">{{ importError }}</div>
        <div class="modal-actions">
          <button class="btn" @click="importOpen = false">Отмена</button>
          <button class="btn btn-primary" :disabled="importing || !importText.trim()" @click="doImport">
            {{ importing ? 'Импорт…' : 'Импортировать' }}
          </button>
        </div>
      </div>
    </div>

    <!-- PROMPT MODAL -->
    <div v-if="promptOpen" class="modal-overlay" @click.self="promptOpen = false">
      <div class="modal">
        <h3>Промпт для генерации квеста</h3>
        <p class="hint">Скопируйте промпт в любую LLM, получите JSON и вставьте его через «Импорт JSON».</p>
        <textarea ref="promptArea" :value="promptText" readonly class="inp code-area" rows="18"></textarea>
        <div class="modal-actions">
          <button class="btn" @click="promptOpen = false">Закрыть</button>
          <button class="btn btn-primary" @click="copyPrompt">{{ copied ? 'Скопировано ✓' : 'Скопировать' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import { courseClient, type CourseSummary } from '../api/courseClient'
import { mediaUrl } from '../utils/mediaUrl'

interface AdminPictureTask {
  id: number
  code: string
  title: string
  completion_criteria: string
  is_required: boolean
  sort_order: number
}
interface AdminPictureQuest {
  id: number
  code: string
  title: string
  cefr_level: string
  image_url: string
  image_description: string
  max_turns: number
  token_budget: number
  sort_order: number
  status: string
  tasks: AdminPictureTask[]
}
interface LevelOption { level_code: string; title: string }

const availableCourses = ref<CourseSummary[]>([])
const selectedCourseCode = ref('')
const coursesError = ref<string | null>(null)
const quests = ref<AdminPictureQuest[]>([])
const levels = ref<LevelOption[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const saving = ref(false)
const publishingId = ref<number | null>(null)

// list controls
const statusFilter = ref('')
const sortDir = ref<'asc' | 'desc'>('asc')
const page = ref(1)
const pageSize = 20
const expanded = ref<Set<number>>(new Set())

const filteredQuests = computed(() => {
  const arr = quests.value.filter(q => !statusFilter.value || q.status === statusFilter.value)
  arr.sort((a, b) => sortDir.value === 'asc' ? a.id - b.id : b.id - a.id)
  return arr
})
const totalPages = computed(() => Math.max(1, Math.ceil(filteredQuests.value.length / pageSize)))
const pagedQuests = computed(() => {
  const start = (page.value - 1) * pageSize
  return filteredQuests.value.slice(start, start + pageSize)
})

// keep page in range when the filtered set shrinks
watch([filteredQuests, totalPages], () => {
  if (page.value > totalPages.value) page.value = totalPages.value
})
// reset to first page when the user changes filter/sort
watch([statusFilter, sortDir], () => { page.value = 1 })

function toggleExpand(id: number) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}

const questForm = ref<Partial<AdminPictureQuest> & { id?: number } | null>(null)
const taskForm = ref<(Partial<AdminPictureTask> & { id?: number, questId: number }) | null>(null)

const importOpen = ref(false)
const importText = ref('')
const importError = ref<string | null>(null)
const importing = ref(false)
const importArea = ref<HTMLTextAreaElement | null>(null)
const uploadingId = ref<number | null>(null)

const promptOpen = ref(false)
const promptText = ref('')
const promptArea = ref<HTMLTextAreaElement | null>(null)
const copied = ref(false)

async function load(silent = false) {
  if (!selectedCourseCode.value) return
  // Silent refresh (after save/delete) keeps the list rendered so it doesn't
  // flash/jump — only the initial/course-change load shows the spinner.
  if (!silent) loading.value = true
  error.value = null
  try {
    const data: { quests?: AdminPictureQuest[], levels?: LevelOption[] } =
      await apiClient.request(`/api/admin/picture-quests?course_code=${encodeURIComponent(selectedCourseCode.value)}`)
    quests.value = data.quests || []
    levels.value = data.levels || []
  } catch (e: any) {
    error.value = e?.message || 'Не удалось загрузить квесты'
  } finally {
    loading.value = false
  }
}

function newQuest() {
  questForm.value = {
    code: '', title: '', cefr_level: levels.value[0]?.level_code || 'A0',
    image_url: '', image_description: '',
    max_turns: 30, token_budget: 40000,
    sort_order: quests.value.length, status: 'draft',
  }
}

async function uploadPicture(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file || !questForm.value) return
  try {
    const res = await courseClient.uploadAdminMedia(file, 'picture')
    questForm.value.image_url = res.url
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось загрузить изображение')
  }
}

function editQuest(q: AdminPictureQuest) {
  questForm.value = { ...q }
}

async function saveQuest() {
  const f = questForm.value
  if (!f) return
  if (!f.code?.trim() || !f.title?.trim() || !f.cefr_level) {
    await showAlert('Заполните code, название и уровень')
    return
  }
  if (!f.image_url?.trim()) {
    await showAlert('Загрузите картинку или укажите её URL')
    return
  }
  if (!f.image_description?.trim()) {
    await showAlert('Заполните описание картинки — по нему модель проверяет ответы ученика')
    return
  }
  saving.value = true
  try {
    const url = `/api/admin/picture-quests${f.id ? '/' + f.id : ''}?course_code=${encodeURIComponent(selectedCourseCode.value)}`
    await apiClient.request(url, { method: f.id ? 'PUT' : 'POST', body: JSON.stringify(f) })
    questForm.value = null
    await load(true)
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось сохранить')
  } finally {
    saving.value = false
  }
}

async function publishQuest(q: AdminPictureQuest) {
  if (!q.image_url) {
    await showAlert('Нельзя опубликовать квест без картинки')
    return
  }
  publishingId.value = q.id
  try {
    await apiClient.request(
      `/api/admin/picture-quests/${q.id}?course_code=${encodeURIComponent(selectedCourseCode.value)}`,
      { method: 'PUT', body: JSON.stringify({ ...q, status: 'active' }) },
    )
    q.status = 'active'
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось опубликовать')
  } finally {
    publishingId.value = null
  }
}

async function removeQuest(q: AdminPictureQuest) {
  if (!await showConfirm(`Удалить квест «${q.title}» и все его задания/сессии?`)) return
  try {
    await apiClient.request(`/api/admin/picture-quests/${q.id}`, { method: 'DELETE' })
    await load(true)
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось удалить')
  }
}

function newTask(q: AdminPictureQuest) {
  taskForm.value = { questId: q.id, code: '', title: '', completion_criteria: '', is_required: true, sort_order: q.tasks.length }
}
function editTask(q: AdminPictureQuest, t: AdminPictureTask) {
  taskForm.value = { ...t, questId: q.id }
}

async function saveTask() {
  const f = taskForm.value
  if (!f) return
  if (!f.code?.trim() || !f.title?.trim() || !f.completion_criteria?.trim()) {
    await showAlert('Заполните code, название и критерий')
    return
  }
  saving.value = true
  try {
    const body = JSON.stringify({
      code: f.code, title: f.title, completion_criteria: f.completion_criteria,
      is_required: f.is_required, sort_order: f.sort_order,
    })
    if (f.id) {
      await apiClient.request(`/api/admin/picture-quests/tasks/${f.id}`, { method: 'PUT', body })
    } else {
      await apiClient.request(`/api/admin/picture-quests/${f.questId}/tasks`, { method: 'POST', body })
    }
    taskForm.value = null
    await load(true)
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось сохранить')
  } finally {
    saving.value = false
  }
}

async function removeTask(_q: AdminPictureQuest, t: AdminPictureTask) {
  if (!await showConfirm(`Удалить задание «${t.title}»?`)) return
  try {
    await apiClient.request(`/api/admin/picture-quests/tasks/${t.id}`, { method: 'DELETE' })
    await load(true)
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось удалить')
  }
}

function openImport() {
  importText.value = ''
  importError.value = null
  importOpen.value = true
  nextTick(() => importArea.value?.focus())
}

// One-click image upload straight from the list row (no edit modal). Uploads the file,
// then PUTs the quest with the new image_url and patches the row in place so the list
// doesn't reload or jump.
async function uploadPictureForQuest(q: AdminPictureQuest, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadingId.value = q.id
  try {
    const res = await courseClient.uploadAdminMedia(file, 'picture')
    await apiClient.request(
      `/api/admin/picture-quests/${q.id}?course_code=${encodeURIComponent(selectedCourseCode.value)}`,
      { method: 'PUT', body: JSON.stringify({ ...q, image_url: res.url }) },
    )
    q.image_url = res.url
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось загрузить изображение')
  } finally {
    uploadingId.value = null
    input.value = ''
  }
}

async function doImport() {
  importError.value = null
  let parsed: unknown
  try {
    parsed = JSON.parse(importText.value)
  } catch (e: any) {
    importError.value = 'Невалидный JSON: ' + (e?.message || '')
    return
  }
  if (Array.isArray(parsed) && parsed.length === 0) {
    importError.value = 'Массив квестов пуст'
    return
  }
  importing.value = true
  try {
    const res: {
      created?: boolean
      task_count?: number
      imported?: number
      created_count?: number
      updated_count?: number
    } = await apiClient.request(
      `/api/admin/picture-quests/import?course_code=${encodeURIComponent(selectedCourseCode.value)}`,
      { method: 'POST', body: JSON.stringify(parsed) },
    )
    importOpen.value = false
    if (typeof res.imported === 'number' && res.imported > 1) {
      await showAlert(`Импортировано квестов: ${res.imported} (создано: ${res.created_count ?? 0}, обновлено: ${res.updated_count ?? 0})`)
    } else if (typeof res.imported === 'number') {
      await showAlert(`${res.created_count ? 'Создан' : 'Обновлён'} 1 квест из массива`)
    }
    // Silent refresh: keep current filter/sort/page, no success popup for single quest, no jump.
    await load(true)
  } catch (e: any) {
    importError.value = e?.message || 'Не удалось импортировать'
  } finally {
    importing.value = false
  }
}

async function openPrompt() {
  copied.value = false
  promptText.value = 'Загрузка…'
  promptOpen.value = true
  try {
    const data: { prompt?: string } = await apiClient.request(
      `/api/admin/picture-quests/prompt-template?course_code=${encodeURIComponent(selectedCourseCode.value)}`)
    promptText.value = data.prompt || ''
  } catch (e: any) {
    promptText.value = ''
    await showAlert(e?.message || 'Не удалось получить промпт')
    promptOpen.value = false
  }
}

async function copyPrompt() {
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

watch(selectedCourseCode, () => { quests.value = []; page.value = 1; expanded.value = new Set(); load() })

onMounted(async () => {
  try {
    const data = await courseClient.getCourses()
    availableCourses.value = data.courses || []
    selectedCourseCode.value = availableCourses.value.find(c => c.is_current)?.code
      || availableCourses.value[0]?.code || ''
  } catch (e: any) {
    coursesError.value = e?.message || 'Не удалось загрузить курсы'
  }
})
</script>

<style scoped>
.course-selector { display: flex; align-items: center; gap: 10px; margin-bottom: 16px; flex-wrap: wrap; }
.level-select, .inp {
  border: 1px solid var(--border-primary); border-radius: 6px; padding: 6px 10px;
  background: var(--bg-primary); color: var(--text-primary); font-size: 13px;
}
.inp { width: 100%; box-sizing: border-box; }
textarea.inp { resize: vertical; font-family: inherit; }
.mono { font-family: ui-monospace, monospace; }
.loading, .empty-message { padding: 24px; color: var(--text-secondary); }
.error { padding: 12px; color: #c0392b; }

/* toolbar */
.pq-toolbar { display: flex; align-items: center; gap: 16px; margin-bottom: 12px; flex-wrap: wrap; }
.pq-toolbar label { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text-secondary); }
.pq-count { font-size: 12px; color: var(--text-secondary); margin-left: auto; }

/* compact list */
.pq-list { display: flex; flex-direction: column; gap: 6px; }
.pq-card { border: 1px solid var(--border-primary); border-radius: 8px; background: var(--bg-secondary); }
.pq-row { display: flex; align-items: center; gap: 10px; padding: 8px 10px; }
.pq-expand { border: none; background: none; cursor: pointer; color: var(--text-secondary); font-size: 11px; padding: 2px 4px; transition: transform 0.15s; flex-shrink: 0; }
.pq-expand.open { transform: rotate(90deg); }
.pq-thumb-upload { position: relative; width: 40px; height: 40px; border-radius: 6px; flex-shrink: 0; cursor: pointer; overflow: hidden; border: 1px solid transparent; box-sizing: border-box; }
.pq-thumb-upload:hover { border-color: #2d6b3a; }
.pq-thumb-upload.busy { cursor: default; opacity: 0.6; }
.pq-thumb-upload--empty { display: flex; align-items: center; justify-content: center; background: var(--bg-primary); color: var(--text-secondary); font-size: 18px; border: 1px dashed var(--border-primary); }
.pq-thumb-upload--empty:hover { border-color: #2d6b3a; color: #2d6b3a; }
.pq-thumb { width: 100%; height: 100%; object-fit: cover; display: block; }
.pq-thumb-state { line-height: 1; }
.pq-thumb-replace { position: absolute; right: 2px; bottom: 2px; width: 16px; height: 16px; border-radius: 999px; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.68); color: #fff; font-size: 12px; line-height: 1; opacity: 0; transition: opacity 0.15s; }
.pq-thumb-upload:hover .pq-thumb-replace { opacity: 1; }
.pq-main { display: flex; align-items: center; gap: 8px; min-width: 0; flex: 1; }
.pq-title { font-weight: 600; font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pq-id { font-size: 11px; color: var(--text-secondary); }
.pq-tasks-count { font-size: 12px; color: var(--text-secondary); white-space: nowrap; flex-shrink: 0; }
.pq-actions { display: flex; gap: 6px; flex-shrink: 0; }
.pq-details { border-top: 1px solid var(--border-primary); padding: 10px 12px; }

/* pagination */
.pq-pagination { display: flex; align-items: center; justify-content: center; gap: 12px; margin-top: 12px; }
.pq-page-info { font-size: 12px; color: var(--text-secondary); }

.scenario-meta { font-size: 12px; color: var(--text-secondary); margin: 6px 0 10px; }
.scenario-meta.description { white-space: pre-wrap; }

.badge { display: inline-block; padding: 1px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; margin-right: 5px; }
.badge--active { background: rgba(45,107,58,0.15); color: #2d6b3a; }
.badge--draft { background: rgba(150,150,150,0.18); color: #777; }
.badge--locked { background: rgba(200,150,40,0.18); color: #9a7b1e; }
.badge--archived { background: rgba(150,150,150,0.12); color: #999; }

.tasks-block { border-top: 1px solid var(--border-primary); padding-top: 8px; }
.tasks-head { display: flex; justify-content: space-between; align-items: center; font-size: 12px; font-weight: 600; color: var(--text-secondary); margin-bottom: 6px; }
.tasks-table { width: 100%; border-collapse: collapse; font-size: 12px; }
.tasks-table th, .tasks-table td { text-align: left; padding: 4px 6px; border-bottom: 1px solid var(--border-primary); vertical-align: top; }
.tasks-table .criteria { color: var(--text-secondary); max-width: 360px; }
.nowrap { white-space: nowrap; }

.btn { border: 1px solid var(--border-primary); border-radius: 6px; padding: 6px 12px; background: var(--bg-primary); color: var(--text-primary); cursor: pointer; font-size: 13px; }
.btn-primary { background: #2d6b3a; color: #fff; border-color: #2d6b3a; }
.btn-danger { color: #c0392b; }
.btn-sm { padding: 4px 9px; font-size: 12px; }
.btn-xs { padding: 2px 7px; font-size: 11px; margin-left: 4px; }
.btn:disabled { opacity: 0.5; cursor: default; }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.45); display: flex; align-items: flex-start; justify-content: center; padding: 40px 16px; z-index: 1000; overflow-y: auto; }
.modal { background: var(--bg-primary); border-radius: 12px; padding: 20px; width: 100%; max-width: 640px; }
.modal h3 { margin: 0 0 14px; }
.form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px 14px; }
.form-grid label, .full { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--text-secondary); }
.full { margin-top: 10px; }
.check { flex-direction: row !important; align-items: center; gap: 6px; }
.modal-actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 16px; }
.hint { font-size: 12px; color: var(--text-secondary); margin: 0 0 10px; line-height: 1.5; }
.code-area { width: 100%; box-sizing: border-box; font-family: ui-monospace, monospace; font-size: 12px; resize: vertical; }
.file-btn { cursor: pointer; }

.image-field { display: flex; gap: 8px; align-items: center; }
.image-field .inp { flex: 1; }
.image-preview { margin-top: 6px; max-height: 120px; border-radius: 8px; object-fit: cover; }
</style>
