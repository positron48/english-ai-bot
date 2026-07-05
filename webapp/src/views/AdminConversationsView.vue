<template>
  <div class="admin-content">
    <h2>Обсуждения (Conversation Quests)</h2>

    <div class="course-selector">
      <label for="conv-course">Курс:</label>
      <select id="conv-course" v-model="selectedCourseCode" class="level-select">
        <option disabled value="">Выберите курс</option>
        <option v-for="course in availableCourses" :key="course.code" :value="course.code">
          {{ course.title || course.code }}
        </option>
      </select>
      <button class="btn btn-primary" :disabled="!selectedCourseCode" @click="newScenario">+ Новый сценарий</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="openImport">Импорт JSON</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="openPrompt">Промпт для генерации</button>
      <button class="btn" :disabled="!selectedCourseCode" @click="load">Обновить</button>
    </div>

    <div v-if="coursesError" class="error">{{ coursesError }}</div>
    <div v-if="loading" class="loading">Загрузка…</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="selectedCourseCode && !scenarios.length" class="empty-message">Сценариев нет.</div>

    <!-- NPC LIST -->
    <div v-if="!loading && npcGroups.length" class="npc-list">
      <div v-for="npc in npcGroups" :key="npc.code" class="npc-card">
        <div class="npc-head">
          <button class="npc-toggle" type="button" @click="toggleNpc(npc.code)" :aria-expanded="isNpcOpen(npc.code)" title="Раскрыть список квестов">
            <span class="npc-chevron">{{ isNpcOpen(npc.code) ? '⌄' : '›' }}</span>
            <span class="npc-summary">
              <span class="npc-name">{{ npc.name }}</span>
              <span class="npc-code mono">{{ npc.code }}</span>
            </span>
          </button>
          <label class="npc-photo" title="Загрузить или заменить фото">
            <img v-if="npc.imageUrl" :src="mediaUrl(npc.imageUrl)" class="npc-avatar-img" alt="" />
            <span v-else class="npc-avatar-placeholder">+</span>
            <input type="file" accept="image/png,image/jpeg,image/webp" class="file-input" @change="uploadNpcImage(npc.code, $event)" />
          </label>
          <div class="npc-stats">
            <span>{{ npc.scenarios.length }} квестов</span>
            <span>{{ npc.taskCount }} задач</span>
          </div>
        </div>

        <div v-if="isNpcOpen(npc.code)" class="scenario-list">
          <div v-for="s in npc.scenarios" :key="s.id" class="scenario-card">
            <div class="scenario-head">
              <div class="scenario-main">
                <label class="quest-photo" title="Загрузить или заменить картинку квеста">
                  <img v-if="s.image_url" :src="mediaUrl(s.image_url)" class="quest-img" alt="" />
                  <span v-else class="quest-placeholder">+</span>
                  <input type="file" accept="image/png,image/jpeg,image/webp" class="file-input" @change="uploadScenarioImage(s, $event)" />
                </label>
                <div class="scenario-text">
                  <div>
                    <span class="scenario-title">{{ s.title }}</span>
                    <span class="badge" :class="'badge--' + s.status">{{ s.status }}</span>
                    <span v-if="s.is_quest" class="badge badge--quest">quest</span>
                    <span v-else class="badge badge--free">free</span>
                  </div>
                  <div class="scenario-meta mono">
                    {{ s.code }} · {{ s.cefr_level }} · {{ s.place_type }} · NPC: {{ s.npc_name }}
                    · {{ s.tasks.length }} задач · {{ s.max_turns }} ходов · order {{ s.sort_order }}
                  </div>
                  <div v-if="s.npc_code || s.prerequisite_code" class="scenario-meta mono chain">
                    <span v-if="s.npc_code">npc: {{ s.npc_code }}</span>
                    <span v-if="s.prerequisite_code">после: {{ s.prerequisite_code }}</span>
                  </div>
                </div>
              </div>
              <div class="scenario-actions">
                <button class="btn btn-sm" @click="editScenario(s)">Изменить</button>
                <button class="btn btn-sm" @click="newTask(s)">+ Задача</button>
                <button class="btn btn-sm btn-danger" @click="removeScenario(s)">Удалить</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- SCENARIO EDIT MODAL -->
    <div v-if="scenarioForm" class="modal-overlay" @click.self="scenarioForm = null">
      <div class="modal">
        <h3>{{ scenarioForm.id ? 'Изменить сценарий' : 'Новый сценарий' }}</h3>
        <div class="form-grid">
          <label>code <input v-model="scenarioForm.code" class="inp mono" /></label>
          <label>CEFR уровень
            <select v-model="scenarioForm.cefr_level" class="inp">
              <option v-for="l in levels" :key="l.level_code" :value="l.level_code">{{ l.level_code }} — {{ l.title }}</option>
            </select>
          </label>
          <label>Название <input v-model="scenarioForm.title" class="inp" /></label>
          <label>place_type <input v-model="scenarioForm.place_type" class="inp mono" /></label>
          <label>NPC имя <input v-model="scenarioForm.npc_name" class="inp" /></label>
          <label>Статус
            <select v-model="scenarioForm.status" class="inp">
              <option value="draft">draft</option>
              <option value="active">active</option>
              <option value="locked">locked</option>
              <option value="archived">archived</option>
            </select>
          </label>
          <label class="check"><input type="checkbox" v-model="scenarioForm.is_quest" /> Квест (есть задачи)</label>
          <label>Макс. ходов <input v-model.number="scenarioForm.max_turns" type="number" class="inp" /></label>
          <label>Бюджет токенов <input v-model.number="scenarioForm.token_budget" type="number" class="inp" /></label>
          <label>Порядок <input v-model.number="scenarioForm.sort_order" type="number" class="inp" /></label>
          <label>NPC код (цепочка)
            <input v-model="scenarioForm.npc_code" class="inp mono" placeholder="например mara_barista" />
          </label>
          <label>Открывается после (code сценария)
            <input v-model="scenarioForm.prerequisite_code" class="inp mono" placeholder="пусто = доступен сразу" />
          </label>
        </div>
        <label class="full">Картинка квеста (URL или загрузите файл)
          <div class="image-field">
            <input v-model="scenarioForm.image_url" class="inp mono" placeholder="https://..." />
            <label class="btn btn-sm file-btn">
              Загрузить
              <input type="file" accept="image/png,image/jpeg,image/webp" style="display:none" @change="uploadQuestImage($event)" />
            </label>
          </div>
          <img v-if="scenarioForm.image_url" :src="mediaUrl(scenarioForm.image_url)" class="image-preview" alt="" />
        </label>
        <label class="full">NPC персона (инструкция для ИИ)
          <textarea v-model="scenarioForm.npc_persona" class="inp" rows="2"></textarea>
        </label>
        <label class="full">Сцена / завязка (инструкция для ИИ)
          <textarea v-model="scenarioForm.scene_setup" class="inp" rows="2"></textarea>
        </label>
        <div class="modal-actions">
          <button class="btn" @click="scenarioForm = null">Отмена</button>
          <button class="btn btn-primary" :disabled="saving" @click="saveScenario">{{ saving ? 'Сохранение…' : 'Сохранить' }}</button>
        </div>
      </div>
    </div>

    <!-- TASK EDIT MODAL -->
    <div v-if="taskForm" class="modal-overlay" @click.self="taskForm = null">
      <div class="modal">
        <h3>{{ taskForm.id ? 'Изменить задачу' : 'Новая задача' }}</h3>
        <div class="form-grid">
          <label>code <input v-model="taskForm.code" class="inp mono" /></label>
          <label>Порядок <input v-model.number="taskForm.sort_order" type="number" class="inp" /></label>
          <label>Название <input v-model="taskForm.title" class="inp" /></label>
          <label class="check"><input type="checkbox" v-model="taskForm.is_required" /> Обязательная</label>
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
        <h3>Импорт сценария из JSON</h3>
        <p class="hint">
          Вставьте JSON одного сценария с задачами (структура — как из кнопки «Промпт для генерации»).
          Сценарий с таким же <span class="mono">code</span> в курсе <span class="mono">{{ selectedCourseCode }}</span>
          будет перезаписан вместе с задачами.
        </p>
        <textarea v-model="importText" class="inp code-area" rows="16" placeholder='{ "code": "...", "title": "...", "tasks": [ ... ] }'></textarea>
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
        <h3>Промпт для генерации сценария</h3>
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

interface AdminTask {
  id: number
  code: string
  title: string
  completion_criteria: string
  is_required: boolean
  sort_order: number
}
interface AdminScenario {
  id: number
  code: string
  title: string
  cefr_level: string
  place_type: string
  npc_name: string
  npc_persona: string
  scene_setup: string
  is_quest: boolean
  max_turns: number
  token_budget: number
  npc_code: string
  prerequisite_code: string
  image_url: string
  sort_order: number
  status: string
  tasks: AdminTask[]
}
interface LevelOption { level_code: string; title: string }

const availableCourses = ref<CourseSummary[]>([])
const selectedCourseCode = ref('')
const coursesError = ref<string | null>(null)
const scenarios = ref<AdminScenario[]>([])
const levels = ref<LevelOption[]>([])
const npcImages = ref<Record<string, string>>({})
const openNpcCodes = ref<Set<string>>(new Set())
const loading = ref(false)
const error = ref<string | null>(null)
const saving = ref(false)

const scenarioForm = ref<Partial<AdminScenario> & { id?: number } | null>(null)
const taskForm = ref<(Partial<AdminTask> & { id?: number, scenarioId: number }) | null>(null)

const importOpen = ref(false)
const importText = ref('')
const importError = ref<string | null>(null)
const importing = ref(false)

const promptOpen = ref(false)
const promptText = ref('')
const promptArea = ref<HTMLTextAreaElement | null>(null)
const copied = ref(false)

const npcGroups = computed(() => {
  const byCode = new Map<string, { code: string; name: string; imageUrl: string; scenarios: AdminScenario[]; taskCount: number }>()
  for (const scenario of scenarios.value) {
    const code = (scenario.npc_code || scenario.npc_name || scenario.code).trim()
    const group = byCode.get(code) || {
      code,
      name: scenario.npc_name || code,
      imageUrl: npcImages.value[code] || '',
      scenarios: [],
      taskCount: 0,
    }
    if (!group.name && scenario.npc_name) group.name = scenario.npc_name
    group.imageUrl = npcImages.value[code] || group.imageUrl
    group.scenarios.push(scenario)
    group.taskCount += scenario.tasks.length
    byCode.set(code, group)
  }
  return [...byCode.values()]
    .map(group => ({
      ...group,
      scenarios: [...group.scenarios].sort((a, b) => a.sort_order - b.sort_order || a.title.localeCompare(b.title)),
    }))
    .sort((a, b) => {
      const firstA = a.scenarios[0]?.sort_order ?? 0
      const firstB = b.scenarios[0]?.sort_order ?? 0
      return firstA - firstB || a.name.localeCompare(b.name)
    })
})

function isNpcOpen(npcCode: string) {
  return openNpcCodes.value.has(npcCode)
}

function toggleNpc(npcCode: string) {
  const next = new Set(openNpcCodes.value)
  if (next.has(npcCode)) next.delete(npcCode)
  else next.add(npcCode)
  openNpcCodes.value = next
}

async function load() {
  if (!selectedCourseCode.value) return
  loading.value = true
  error.value = null
  try {
    const data: { scenarios?: AdminScenario[], levels?: LevelOption[], npc_images?: Record<string, string> } =
      await apiClient.request(`/api/admin/conversations/scenarios?course_code=${encodeURIComponent(selectedCourseCode.value)}`)
    scenarios.value = data.scenarios || []
    levels.value = data.levels || []
    npcImages.value = data.npc_images || {}
    if (!openNpcCodes.value.size && scenarios.value.length) {
      const firstNpcCode = scenarios.value[0].npc_code || scenarios.value[0].npc_name || scenarios.value[0].code
      openNpcCodes.value = new Set([firstNpcCode])
    }
  } catch (e: any) {
    error.value = e?.message || 'Не удалось загрузить сценарии'
  } finally {
    loading.value = false
  }
}

function newScenario() {
  scenarioForm.value = {
    code: '', title: '', cefr_level: levels.value[0]?.level_code || 'A0', place_type: 'cafe',
    npc_name: '', npc_persona: '', scene_setup: '', is_quest: true,
    max_turns: 30, token_budget: 40000, npc_code: '', prerequisite_code: '', image_url: '',
    sort_order: scenarios.value.length, status: 'draft',
  }
}

async function uploadQuestImage(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file || !scenarioForm.value) return
  try {
    const res = await courseClient.uploadAdminMedia(file, 'quest')
    scenarioForm.value.image_url = res.url
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось загрузить изображение')
  } finally {
    input.value = ''
  }
}

async function uploadScenarioImage(scenario: AdminScenario, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const res = await courseClient.uploadAdminMedia(file, 'quest')
    const updated = { ...scenario, image_url: res.url }
    await apiClient.request(
      `/api/admin/conversations/scenarios/${scenario.id}?course_code=${encodeURIComponent(selectedCourseCode.value)}`,
      { method: 'PUT', body: JSON.stringify(updated) },
    )
    scenarios.value = scenarios.value.map(s => s.id === scenario.id ? updated : s)
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось загрузить картинку квеста')
  } finally {
    input.value = ''
  }
}

async function uploadNpcImage(npcCode: string, event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const res = await courseClient.uploadAdminMedia(file, 'npc')
    await courseClient.setAdminNpcImage(npcCode, res.url, selectedCourseCode.value)
    npcImages.value = { ...npcImages.value, [npcCode]: res.url }
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось загрузить изображение NPC')
  } finally {
    input.value = ''
  }
}
function editScenario(s: AdminScenario) {
  scenarioForm.value = { ...s }
}

async function saveScenario() {
  const f = scenarioForm.value
  if (!f) return
  if (!f.code?.trim() || !f.title?.trim() || !f.cefr_level) {
    await showAlert('Заполните code, название и уровень')
    return
  }
  saving.value = true
  try {
    const url = `/api/admin/conversations/scenarios${f.id ? '/' + f.id : ''}?course_code=${encodeURIComponent(selectedCourseCode.value)}`
    await apiClient.request(url, { method: f.id ? 'PUT' : 'POST', body: JSON.stringify(f) })
    scenarioForm.value = null
    await load()
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось сохранить')
  } finally {
    saving.value = false
  }
}

async function removeScenario(s: AdminScenario) {
  if (!await showConfirm(`Удалить сценарий «${s.title}» и все его задачи/сессии?`)) return
  try {
    await apiClient.request(`/api/admin/conversations/scenarios/${s.id}`, { method: 'DELETE' })
    await load()
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось удалить')
  }
}

function newTask(s: AdminScenario) {
  taskForm.value = { scenarioId: s.id, code: '', title: '', completion_criteria: '', is_required: true, sort_order: s.tasks.length }
}
function editTask(s: AdminScenario, t: AdminTask) {
  taskForm.value = { ...t, scenarioId: s.id }
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
      await apiClient.request(`/api/admin/conversations/tasks/${f.id}`, { method: 'PUT', body })
    } else {
      await apiClient.request(`/api/admin/conversations/scenarios/${f.scenarioId}/tasks`, { method: 'POST', body })
    }
    taskForm.value = null
    await load()
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось сохранить')
  } finally {
    saving.value = false
  }
}

async function removeTask(_s: AdminScenario, t: AdminTask) {
  if (!await showConfirm(`Удалить задачу «${t.title}»?`)) return
  try {
    await apiClient.request(`/api/admin/conversations/tasks/${t.id}`, { method: 'DELETE' })
    await load()
  } catch (e: any) {
    await showAlert(e?.message || 'Не удалось удалить')
  }
}

function openImport() {
  importText.value = ''
  importError.value = null
  importOpen.value = true
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
  importing.value = true
  try {
    const res: { created?: boolean, task_count?: number } = await apiClient.request(
      `/api/admin/conversations/import?course_code=${encodeURIComponent(selectedCourseCode.value)}`,
      { method: 'POST', body: JSON.stringify(parsed) },
    )
    importOpen.value = false
    await load()
    await showAlert(`${res.created ? 'Создан' : 'Обновлён'} сценарий, задач: ${res.task_count ?? 0}`)
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
      `/api/admin/conversations/prompt-template?course_code=${encodeURIComponent(selectedCourseCode.value)}`)
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

watch(selectedCourseCode, () => { scenarios.value = []; openNpcCodes.value = new Set(); load() })

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

.npc-list { display: flex; flex-direction: column; gap: 12px; }
.npc-card { border: 1px solid var(--border-primary); border-radius: 8px; background: var(--bg-secondary); overflow: hidden; }
.npc-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 10px 12px; }
.npc-toggle { flex: 1; min-width: 0; display: flex; align-items: center; gap: 10px; border: 0; background: transparent; color: var(--text-primary); padding: 0; cursor: pointer; text-align: left; }
.npc-chevron { width: 18px; flex: 0 0 18px; font-size: 22px; line-height: 1; color: var(--text-secondary); text-align: center; }
.npc-photo { position: relative; flex: 0 0 auto; display: block; width: 52px; height: 52px; cursor: pointer; }
.npc-photo::after {
  content: 'Заменить';
  position: absolute;
  inset: auto 3px 3px;
  border-radius: 5px;
  padding: 2px 3px;
  background: rgba(0,0,0,0.64);
  color: #fff;
  font-size: 9px;
  line-height: 1;
  text-align: center;
  opacity: 0;
  transition: opacity .15s ease;
}
.npc-photo:hover::after { opacity: 1; }
.npc-avatar-img { width: 52px; height: 52px; border-radius: 50%; object-fit: cover; display: block; }
.npc-avatar-placeholder { width: 52px; height: 52px; border-radius: 50%; background: rgba(0,0,0,0.08); display: flex; align-items: center; justify-content: center; font-size: 24px; color: var(--text-secondary); }
.file-input { display: none; }
.npc-summary { min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.npc-name { font-weight: 700; font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.npc-code { font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.npc-stats { flex: 0 0 auto; display: flex; gap: 8px; color: var(--text-secondary); font-size: 12px; white-space: nowrap; }
.npc-stats span { border: 1px solid var(--border-primary); border-radius: 999px; padding: 3px 8px; background: var(--bg-primary); }
.scenario-list { display: flex; flex-direction: column; gap: 10px; padding: 0 12px 12px 40px; }
.scenario-card { border: 1px solid var(--border-primary); border-radius: 8px; padding: 10px 12px; background: var(--bg-primary); }
.scenario-head { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
.scenario-main { min-width: 0; display: flex; align-items: flex-start; gap: 12px; }
.scenario-text { min-width: 0; }
.scenario-title { font-weight: 600; font-size: 15px; margin-right: 8px; }
.scenario-meta { font-size: 12px; color: var(--text-secondary); margin: 6px 0 10px; }
.scenario-actions { display: flex; gap: 6px; flex-shrink: 0; }
.quest-photo { position: relative; flex: 0 0 auto; display: block; width: 72px; height: 52px; cursor: pointer; }
.quest-photo::after {
  content: 'Заменить';
  position: absolute;
  inset: auto 4px 4px;
  border-radius: 5px;
  padding: 2px 3px;
  background: rgba(0,0,0,0.64);
  color: #fff;
  font-size: 9px;
  line-height: 1;
  text-align: center;
  opacity: 0;
  transition: opacity .15s ease;
}
.quest-photo:hover::after { opacity: 1; }
.quest-img,
.quest-placeholder { width: 72px; height: 52px; border-radius: 7px; display: block; }
.quest-img { object-fit: cover; }
.quest-placeholder { background: rgba(0,0,0,0.08); color: var(--text-secondary); display: flex; align-items: center; justify-content: center; font-size: 24px; }

.badge { display: inline-block; padding: 1px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; margin-right: 5px; }
.badge--active { background: rgba(45,107,58,0.15); color: #2d6b3a; }
.badge--draft { background: rgba(150,150,150,0.18); color: #777; }
.badge--locked { background: rgba(200,150,40,0.18); color: #9a7b1e; }
.badge--archived { background: rgba(150,150,150,0.12); color: #999; }
.badge--quest { background: rgba(45,107,58,0.12); color: #2d6b3a; }
.badge--free { background: rgba(200,168,75,0.18); color: #9a7b1e; }

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
.scenario-meta.chain { display: flex; gap: 14px; margin-top: -4px; color: var(--text-secondary); }

.file-btn { cursor: pointer; }

/* Quest image upload in form */
.image-field { display: flex; gap: 8px; align-items: center; }
.image-field .inp { flex: 1; }
.image-preview { margin-top: 6px; max-height: 80px; border-radius: 8px; object-fit: cover; }

@media (max-width: 720px) {
  .npc-head { align-items: flex-start; flex-direction: column; }
  .npc-toggle { width: 100%; }
  .npc-stats { padding-left: 28px; flex-wrap: wrap; }
  .scenario-list { padding-left: 12px; }
  .scenario-head { flex-direction: column; }
  .scenario-main { width: 100%; }
  .scenario-text { flex: 1; }
  .form-grid { grid-template-columns: 1fr; }
}
</style>
