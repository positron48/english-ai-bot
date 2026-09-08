<template>
  <section id="verb-forms-training" class="verb-practice" :aria-busy="busy">
    <p v-if="error" class="practice-error" role="alert">{{ error }} <button type="button" @click="recover">{{ t('verbPractice.reload') }}</button></p>
    <p v-if="loading" role="status">{{ t('common.loading') }}</p>
    <template v-else-if="session?.completed">
      <h2>{{ t('verbPractice.finished') }}</h2>
      <p v-if="session.retry">{{ t('verbPractice.retryNotice') }}</p>
      <dl class="result-grid">
        <div><dt>{{ t('verbPractice.independent') }}</dt><dd>{{ count('correct', false) }}</dd></div>
        <div><dt>{{ t('verbPractice.withHelp') }}</dt><dd>{{ count('correct', true) }}</dd></div>
        <div><dt>{{ t('verbPractice.mistakes') }}</dt><dd>{{ count('incorrect') }}</dd></div>
        <div><dt>{{ t('verbPractice.unknownCount') }}</dt><dd>{{ count('unknown') }}</dd></div>
      </dl>
      <div v-if="mistakes.length" class="mistake-list">
        <h3>{{ t('verbPractice.review') }}</h3>
        <p v-for="result in mistakes" :key="result.card_id"><strong>{{ result.lemma }} → {{ result.correct_answer }}</strong><br>{{ tenseLabel(result.tense) }} · {{ pronoun(result.person, result.number) }}</p>
      </div>
      <div class="practice-actions">
        <button v-if="mistakes.length" class="btn btn-primary" :disabled="busy" @click="act('repeat')">{{ t('verbPractice.repeat') }}</button>
        <button class="btn btn-secondary" :disabled="busy" @click="act('start')">{{ t('verbPractice.continue') }}</button>
      </div>
    </template>
    <template v-else-if="session?.card_id && session.prompt">
      <div class="practice-progress">
        <span>{{ t('verbTraining.cardProgress', { current: session.card_index, total: session.total_cards }) }}</span>
        <span v-if="session.retry">{{ t('verbPractice.repeat') }}</span>
        <progress :value="session.card_index - 1 + (session.feedback ? 1 : 0)" :max="session.total_cards" :aria-label="t('verbPractice.progress')" />
      </div>
      <p class="practice-lemma"><strong>{{ session.prompt.lemma }}</strong><span v-if="session.prompt.ru_gloss"> — {{ session.prompt.ru_gloss }}</span></p>
      <h2 ref="questionHeading" tabindex="-1" class="practice-question">{{ session.prompt.question }}</h2>
      <p v-if="session.prompt.example_translation" class="practice-translation">{{ session.prompt.example_translation }}</p>
      <p v-if="!session.feedback" class="practice-instruction">{{ t(session.input_mode === 'typed' ? 'verbPractice.type' : 'verbPractice.choose') }}</p>
      <div v-if="session.input_mode === 'choice'" class="practice-options">
        <button v-for="option in session.options" :key="option" type="button" class="practice-option"
          :class="{ correct: session.feedback?.correct_answer === option, incorrect: session.feedback?.outcome === 'incorrect' && session.feedback.chosen_option === option }"
          :disabled="busy || !!session.feedback" @click="answer(option)">{{ option }}</button>
      </div>
      <form v-else class="practice-input" @submit.prevent="answer(typedAnswer)">
        <label for="verb-answer">{{ t('verbTraining.typeFormPlaceholder') }}</label>
        <input id="verb-answer" v-model="typedAnswer" autocomplete="off" autocapitalize="none" :spellcheck="false" :disabled="busy || !!session.feedback" />
        <button v-if="!session.feedback" class="btn btn-primary" :disabled="busy || !typedAnswer.trim()">{{ t('verbTraining.submitAnswer') }}</button>
      </form>
      <template v-if="!session.feedback">
        <div class="practice-help-actions">
          <button class="practice-tool practice-tool--hint" type="button" :disabled="busy" :aria-expanded="showHint" aria-controls="verb-practice-hint" @click="help">
            <span class="practice-tool-icon"><LgIcon name="lightbulb" :s="20" aria-hidden="true" /></span>
            <span>{{ t('verbPractice.hint') }}</span>
          </button>
          <button class="practice-tool" type="button" :disabled="busy" aria-haspopup="dialog" @click="openForms">
            <span class="practice-tool-icon"><LgIcon name="book-open" :s="20" aria-hidden="true" /></span>
            <span>{{ t('verbPractice.allForms') }}</span>
          </button>
          <button class="practice-skip" type="button" :disabled="busy" @click="act('answer', { skip: true })">
            <span>{{ t('verbPractice.dontKnow') }}</span><LgIcon name="chevron-right" :s="18" aria-hidden="true" />
          </button>
        </div>
        <p v-if="session.assisted" class="practice-assisted">{{ t('verbPractice.helpUsed') }}</p>
        <div v-if="showHint && session.prompt.rule" id="verb-practice-hint" class="practice-hint"><strong>{{ tenseLabel(session.prompt.tense) }} · {{ moodLabel(session.prompt.mood) }}</strong><p>{{ ruleText(session.prompt.rule) }}</p></div>
      </template>
      <div v-else class="practice-feedback" role="status" aria-live="polite">
        <p class="feedback-title"><LgIcon :name="session.feedback.outcome === 'correct' ? 'check' : 'book-open'" />{{ t('verbPractice.' + session.feedback.outcome) }}</p>
        <p v-if="session.feedback.chosen_option && session.feedback.outcome === 'incorrect'">{{ t('verbPractice.yourAnswer') }}: {{ session.feedback.chosen_option }}</p>
        <p class="practice-answer">{{ session.feedback.correct_answer }}</p>
        <p v-if="session.feedback.accent_only">{{ t('verbPractice.accent') }}</p>
        <p class="practice-tense">{{ tenseLabel(session.prompt.tense) }} · {{ moodLabel(session.prompt.mood) }}</p>
        <p>{{ ruleText(session.feedback.rule) }}</p>
        <p><strong>{{ session.feedback.sentence }}</strong><br><span v-if="session.feedback.translation">{{ session.feedback.translation }}</span></p>
        <p v-if="session.feedback.assisted">{{ t('verbPractice.helpUsed') }}</p>
        <button class="practice-tool feedback-forms" type="button" :disabled="busy" aria-haspopup="dialog" @click="openForms"><LgIcon name="book-open" :s="20" aria-hidden="true" /><span>{{ t('verbPractice.allForms') }}</span></button>
        <div class="practice-next"><button ref="nextButton" class="btn btn-primary" :disabled="busy" @click="advance">{{ t('verbPractice.next') }}</button></div>
      </div>
    </template>
    <template v-else>
      <h2 v-if="!hidePageTitle">{{ t('verbTraining.title') }}</h2>
      <p>{{ t(session?.empty ? 'verbPractice.empty' : 'verbPractice.intro') }}</p>
      <button type="button" class="btn btn-primary" :disabled="busy" @click="act('start')">{{ t('verbTraining.start') }}</button>
    </template>
    <div v-if="showForms" class="forms-overlay" @click.self="closeForms" @keydown.esc="closeForms">
      <section ref="formsDialog" role="dialog" aria-modal="true" :aria-label="t('verbPractice.allForms')" class="forms-dialog" tabindex="-1" @keydown.tab="trapFocus">
        <header><h3>{{ session?.prompt?.lemma }}</h3><button type="button" :aria-label="t('verbPractice.close')" @click="closeForms">×</button></header>
        <p v-if="formsLoading">{{ t('common.loading') }}</p>
        <p v-if="formsError" role="alert">{{ t('verbPractice.networkError') }}</p>
        <label for="verb-tense">{{ t('verbPractice.tense') }}</label>
        <select id="verb-tense" v-model="selectedScope"><option v-for="group in formGroups" :key="group.key" :value="group.key">{{ tenseLabel(group.tense) }} · {{ moodLabel(group.mood) }}</option></select>
        <table><thead><tr><th>{{ t('verbPractice.person') }}</th><th>{{ t('verbPractice.form') }}</th></tr></thead><tbody><tr v-for="(row,index) in selectedForms" :key="index"><td>{{ pronoun(row.person,row.number) }}</td><td>{{ row.surface_form }}</td></tr></tbody></table>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import LgIcon from './linglow/LgIcon.vue'

const props = withDefaults(defineProps<{ embedded?: boolean; autoStart?: boolean; hidePageTitle?: boolean }>(), { embedded: true, autoStart: false, hidePageTitle: false })
const { t, te } = useI18n()
interface Rule { id: string; regular: boolean; ending?: string; stem?: string; pattern?: string }
interface Feedback { card_id: number; outcome: string; assisted: boolean; chosen_option: string; correct_answer: string; sentence: string; translation: string; lemma: string; tense: string; mood: string; person: string; number: string; rule: Rule | null; accent_only: boolean }
interface Prompt { question: string; example_translation?: string; lemma: string; ru_gloss?: string; mood: string; tense: string; person: string; number: string; rule?: Rule }
interface Session { session_id: number; card_id?: number; card_index: number; total_cards: number; input_mode: string; options?: string[]; prompt?: Prompt; completed?: boolean; idle?: boolean; empty?: boolean; retry?: boolean; assisted?: boolean; feedback?: Feedback; results?: Feedback[] }
interface Form { mood: string; tense: string; person: string; number: string; surface_form: string }
const session = ref<Session | null>(null)
const loading = ref(true), busy = ref(false), error = ref(''), typedAnswer = ref(''), showHint = ref(false)
const showForms = ref(false), formsLoading = ref(false), formsError = ref(false), forms = ref<Form[]>([]), selectedScope = ref('')
const nextButton = ref<HTMLButtonElement>(), questionHeading = ref<HTMLElement>(), formsDialog = ref<HTMLElement>()
let previousFocus: HTMLElement | null = null
const mistakes = computed(() => session.value?.results?.filter(r => r.outcome !== 'correct') || [])
const formGroups = computed(() => Array.from(new Map(forms.value.map(row => [scopeKey(row.mood, row.tense), { key: scopeKey(row.mood, row.tense), mood: row.mood, tense: row.tense }])).values()))
const selectedForms = computed(() => {
  const slots = new Map<string, Form>()
  for (const row of forms.value.filter(row => scopeKey(row.mood, row.tense) === selectedScope.value)) {
    const key = row.person + row.number, existing = slots.get(key)
    if (!existing) slots.set(key, { ...row })
    else if (!existing.surface_form.split(' / ').includes(row.surface_form)) existing.surface_form += ' / ' + row.surface_form
  }
  return [...slots.values()]
})
function count(outcome: string, assisted?: boolean) { return session.value?.results?.filter(r => r.outcome === outcome && (assisted === undefined || r.assisted === assisted)).length || 0 }
function scopeKey(mood: string, tense: string) {
  const aliases: Record<string, string> = { preterito_indefinido: 'pretérito', preterito_imperfecto: 'imperfecto', futuro_simple: 'futuro', condicional_simple: 'condicional', preterito_perfecto_compuesto: 'pretérito perfecto', preterito_perfecto: 'pretérito perfecto', preterito_pluscuamperfecto: 'pluscuamperfecto', preterito_anterior: 'pretérito anterior', futuro_perfecto: 'futuro perfecto', condicional_perfecto: 'condicional perfecto' }
  return mood + '|' + (aliases[tense] || tense)
}
function tenseLabel(tense: string) { const key = 'verbPractice.tenses.' + tense; return te(key) ? t(key) : tense }
function moodLabel(mood: string) { const key = 'verbPractice.moods.' + mood; return te(key) ? t(key) : mood }
function pronoun(person: string, number: string) { return ({ '1singular': 'yo', '2singular': 'tú', '3singular': 'él / ella / usted', '1plural': 'nosotros / nosotras', '2plural': 'vosotros / vosotras', '3plural': 'ellos / ellas / ustedes' } as Record<string,string>)[person + number] || person }
function ruleText(rule?: Rule | null) {
  const p = session.value?.prompt
  if (rule?.pattern && te('verbPractice.patterns.' + rule.pattern)) return t('verbPractice.patterns.' + rule.pattern)
  if (rule?.pattern) return t('verbPractice.stemChange', { change: rule.pattern.replace('_', ' → '), ending: rule.ending })
  if (rule?.regular) return t('verbPractice.regularRule', { pronoun: pronoun(p?.person || '', p?.number || ''), ending: rule.ending })
  return t('verbPractice.exceptionRule', { pronoun: pronoun(p?.person || '', p?.number || '') })
}
async function act(action: string, extra: Record<string,unknown> = {}) {
  if (busy.value) return
  busy.value = true; error.value = ''
  try {
    session.value = await apiClient.request<Session>('/api/verb-training/v2/' + action, { method: 'POST', body: { session_id: session.value?.session_id, card_id: session.value?.card_id, ...extra } })
    if (action === 'start' || action === 'repeat' || action === 'advance') { typedAnswer.value = ''; showHint.value = false }

  } catch { error.value = t('verbPractice.networkError') }
  finally { busy.value = false; await nextTick(); if (session.value?.feedback) nextButton.value?.focus() }
}
async function answer(value: string) { if (value.trim()) await act('answer', { answer: value.trim() }) }
async function advance() { await act('advance'); await nextTick(); questionHeading.value?.focus() }
async function help() { await act('help'); if (!error.value) showHint.value = true }
async function recover() {
  error.value = ''
  try { session.value = await apiClient.request<Session>('/api/verb-training/v2/current') }
  catch { error.value = t('verbPractice.networkError') }
}
async function openForms() {
  previousFocus = document.activeElement as HTMLElement
  await act('help'); if (error.value) return
  showForms.value = true; formsLoading.value = true; formsError.value = false
  await nextTick(); formsDialog.value?.focus()
  try {
    const response = await apiClient.request<{ forms: Form[] }>('/api/verb-training/forms-by-lemma?lemma=' + encodeURIComponent(session.value?.prompt?.lemma || ''))
    forms.value = response.forms || []
    const current = scopeKey(session.value?.prompt?.mood || '', session.value?.prompt?.tense || '')
    selectedScope.value = formGroups.value.some(g => g.key === current) ? current : (formGroups.value[0]?.key || '')
  } catch { formsError.value = true } finally { formsLoading.value = false }
}
function closeForms() { showForms.value = false; previousFocus?.focus() }
function trapFocus(event: KeyboardEvent) {
  const nodes = formsDialog.value?.querySelectorAll<HTMLElement>('button, select, [tabindex="0"]')
  if (!nodes?.length) return
  const first = nodes[0], last = nodes[nodes.length - 1]
  if (event.shiftKey && (document.activeElement === first || document.activeElement === formsDialog.value)) { event.preventDefault(); last.focus() }
  else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus() }
}
onMounted(async () => { await recover(); loading.value = false; if (!error.value && props.autoStart && session.value?.idle) await act('start') })
</script>

<style scoped>
.practice-translation { color: var(--text-secondary); font-size: 1rem; line-height: 1.5; margin: 0 0 1rem; }
.practice-hint p { margin-bottom: 0; }
.verb-practice { max-width: 620px; margin: 0 auto; padding: 20px; border-radius: 24px; background: var(--lg-surface, var(--card-bg, #fff9ed)); color: var(--lg-text, inherit); }
.practice-progress { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 8px; font-size: .9rem; }
progress { width: 100%; height: 6px; border: 0; border-radius: 8px; overflow: hidden; accent-color: var(--salvia, #52754a); background: var(--surface-3); }
progress::-webkit-progress-bar { background: var(--surface-3); }
progress::-webkit-progress-value { background: var(--salvia, #52754a); border-radius: 8px; }
.practice-tense { margin-top: 12px; }
.practice-instruction { margin-bottom: 12px; }
.practice-feedback p + p { margin-top: 10px; }
.practice-tense, .practice-instruction, .practice-assisted { font-size: .9rem; opacity: .75; }
.practice-lemma { margin: 20px 0 12px; }
.practice-question { font: inherit; font-size: clamp(1.3rem, 5vw, 1.75rem); font-weight: 650; line-height: 1.4; margin: 0 0 12px; overflow-wrap: anywhere; }
.practice-question:focus { outline: none; }
.practice-options { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; }
.practice-option { min-height: 54px; padding: 14px 10px; border: 1px solid var(--border-green, currentColor); border-radius: 16px; background: transparent; color: inherit; font: inherit; font-weight: 600; overflow-wrap: anywhere; cursor: pointer; }
.practice-option:disabled { cursor: default; opacity: .8; }
.practice-option.correct { background: #deedd8; color: #24431c; border-color: #52754a; }
.practice-option.incorrect { background: #f9e0dc; color: #792f27; border-color: #b75145; }
.practice-help-actions { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 10px; margin-top: 24px; padding-top: 18px; border-top: 1px solid var(--border, #8883); }
.practice-tool, .practice-skip { display: flex; align-items: center; justify-content: center; gap: 9px; min-height: 50px; padding: 10px 12px; border: 1px solid var(--border-green, #52754a40); border-radius: 16px; background: var(--surface-2, #fff4e2); color: var(--text, inherit); font: inherit; font-size: .9rem; font-weight: 650; line-height: 1.25; text-align: center; cursor: pointer; transition: background .15s, border-color .15s, transform .15s; }
.practice-tool-icon { display: grid; place-items: center; flex-shrink: 0; width: 30px; height: 30px; border-radius: 10px; background: var(--card-bg, #fff9ed); color: var(--salvia, #52754a); }
:global([data-theme="dark"]) .practice-tool-icon { color: var(--hoja, #7fae6a); }
.practice-tool--hint { border-color: color-mix(in srgb, var(--dorado, #d9a83f) 45%, transparent); }
.practice-tool--hint .practice-tool-icon { color: var(--dorado, #a37a29); }
.practice-tool[aria-expanded="true"] { background: var(--surface-3, #f5e9d4); }
.practice-skip { grid-column: 1 / -1; min-height: 44px; background: transparent; border-color: var(--border, #8883); color: var(--subtext, inherit); font-weight: 550; }
.practice-tool svg, .practice-skip svg { flex-shrink: 0; }
.feedback-forms { width: 100%; margin-top: 14px; }
.practice-tool:disabled, .practice-skip:disabled { opacity: .55; cursor: default; }
@media (hover: hover) { .practice-tool:not(:disabled):hover, .practice-skip:not(:disabled):hover { background: var(--surface-3, #f5e9d4); border-color: var(--salvia, #52754a); } }
.practice-tool:not(:disabled):active, .practice-skip:not(:disabled):active { transform: translateY(1px); }
@media (prefers-reduced-motion: reduce) { .practice-tool, .practice-skip { transition: none; } }
.practice-error button { padding: 10px 0; background: transparent; border: 0; color: inherit; text-decoration: underline; font: inherit; cursor: pointer; min-height: 44px; }
.practice-hint, .practice-feedback { margin-top: 16px; padding: 16px; border: 1px solid currentColor; border-radius: 16px; line-height: 1.5; }
.feedback-title { display: flex; align-items: center; gap: 8px; font-weight: 600; margin-top: 0; }
.feedback-title svg { width: 20px; height: 20px; }
.practice-answer { font-size: 1.5rem; font-weight: 700; margin: 8px 0; }
.practice-next { position: sticky; bottom: calc(88px + env(safe-area-inset-bottom)); padding-top: 12px; background: var(--card-bg); }
.practice-next button { width: 100%; min-height: 48px; }
.practice-input { display: grid; gap: 10px; }
.practice-input input { min-height: 48px; padding: 10px; border: 1px solid currentColor; border-radius: 12px; font: inherit; color: inherit; background: transparent; width: 100%; box-sizing: border-box; }
.result-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.result-grid div { padding: 12px; border: 1px solid currentColor; border-radius: 12px; }
.result-grid dd { margin: 8px 0 0; font-size: 1.5rem; font-weight: 650; }
.practice-actions { display: flex; flex-wrap: wrap; gap: 12px; }
.practice-error { padding: 12px; border: 1px solid #b75145; border-radius: 12px; }
.forms-overlay { position: fixed; inset: 0; z-index: 1100; background: #0008; display: grid; place-items: center; padding: 16px; }
.forms-dialog { width: min(100%, 520px); max-height: 80dvh; overflow-y: auto; padding: 20px; box-sizing: border-box; border-radius: 20px; background: var(--lg-surface, var(--card-bg, #fff9ed)); }
.forms-dialog header { display: flex; justify-content: space-between; align-items: center; }
.forms-dialog header button { font-size: 1.6rem; border: 0; background: transparent; color: inherit; width: 44px; height: 44px; }
.forms-dialog label { display: block; }
.forms-dialog select { width: 100%; padding: 12px; margin: 8px 0 16px; font: inherit; color: inherit; background: transparent; }
table { width: 100%; border-collapse: collapse; } th, td { padding: 12px 4px; text-align: left; border-bottom: 1px solid #8885; }
button:focus-visible, input:focus-visible, select:focus-visible { outline: 2px solid var(--lg-primary, #52754a); outline-offset: 3px; }
@media (max-width: 390px) { .verb-practice { padding: 16px; } }
</style>
