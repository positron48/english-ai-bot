<template>
  <main class="placement-page" :aria-busy="busy">
    <header class="placement-header">
      <router-link to="/learning/grammar" class="placement-back">
        <LgIcon name="chevron-left" :s="18" aria-hidden="true" /> {{ t('placement.backToCourse') }}
      </router-link>
      <span class="placement-language">{{ t(courseCode === 'es_ru' ? 'placement.spanish' : 'placement.english') }}</span>
    </header>

    <div v-if="errorKey" class="placement-error" role="alert">
      <p>{{ t(errorKey) }}</p>
      <button v-if="expired" type="button" :disabled="busy" @click="start(true)">{{ t('placement.restart') }}</button>
      <button v-else-if="!session" type="button" :disabled="busy" @click="ready ? start() : initialize()">{{ t('placement.retry') }}</button>
      <button v-else-if="conflict" type="button" :disabled="busy" @click="start()">{{ t('placement.resume') }}</button>
    </div>

    <section v-if="!session" class="placement-welcome placement-card">
      <div class="placement-symbol"><LgIcon name="book-open" :s="32" aria-hidden="true" /></div>
      <p class="placement-eyebrow">{{ t('placement.eyebrow') }}</p>
      <h1>{{ t('placement.title') }}</h1>
      <p class="placement-lead">{{ t('placement.description') }}</p>
      <ul class="placement-facts">
        <li><LgIcon name="check" :s="18" aria-hidden="true" /> {{ t('placement.independent') }}</li>
        <li><LgIcon name="check" :s="18" aria-hidden="true" /> {{ t('placement.length') }}</li>
        <li><LgIcon name="refresh" :s="18" aria-hidden="true" /> {{ t('placement.variety') }}</li>
      </ul>
      <p class="placement-note">{{ t('placement.honestAnswers') }}</p>
      <button class="placement-primary" data-test="start" type="button" :disabled="!ready || busy || !supported" @click="start()">
        {{ busy ? t('placement.loading') : (draft?.sessionId ? t('placement.resume') : t('placement.start')) }}
        <LgIcon name="chevron-right" :s="18" aria-hidden="true" />
      </button>
      <p class="placement-fine">{{ t('placement.online') }}</p>
    </section>

    <section v-else-if="session.result" class="placement-results">
      <div class="placement-card placement-result-intro">
        <p class="placement-eyebrow">{{ t('placement.resultTitle') }}</p>
        <h1 ref="questionHeading" tabindex="-1">{{ resultLevel }}</h1>
        <p class="placement-lead">{{ t('placement.resultScope') }}</p>
        <p>{{ t('placement.score', { correct: session.result.correct, total: session.result.total }) }}</p>
        <p class="placement-note">{{ t('placement.accessOnly') }}</p>
      </div>

      <section class="placement-card" aria-labelledby="profile-title">
        <h2 id="profile-title">{{ t('placement.profile') }}</h2>
        <p class="placement-fine">{{ t('placement.profileNote') }}</p>
        <div v-for="level in session.result.profile" :key="level.level" class="placement-level">
          <strong>{{ level.level }}</strong>
          <span>{{ level.correct }} / {{ level.total }}</span>
          <span class="placement-level-status" :class="level.status">{{ t(`placement.${level.status}`) }}</span>
        </div>
      </section>

      <section class="placement-card" aria-labelledby="recommendations-title">
        <h2 id="recommendations-title">{{ t('placement.recommendations') }}</h2>
        <p v-if="!session.result.recommended_skills?.length">{{ t('placement.noRecommendations') }}</p>
        <ul v-else class="placement-recommendations">
          <li v-for="skill in session.result.recommended_skills" :key="skill.id">
            <strong>{{ skill.title }}</strong>
            <router-link v-if="availableChapter(skill.chapter_ids)" :to="chapterLink(availableChapter(skill.chapter_ids)!)">
              {{ t('placement.reviewInCourse') }} <span aria-hidden="true">→</span>
            </router-link>
            <p v-else class="placement-fine">{{ t('placement.topicLater') }}</p>
          </li>
        </ul>
        <router-link class="placement-primary" to="/learning/grammar">{{ t('placement.goToCourse') }}</router-link>
      </section>

      <section class="placement-card" aria-labelledby="review-title">
        <h2 id="review-title">{{ t('placement.review') }}</h2>
        <details v-for="(review, index) in session.result.review" :key="review.id" class="placement-review">
          <summary>
            <LgIcon :name="review.correct ? 'check' : 'x'" :s="18" aria-hidden="true" />
            <span>{{ t('placement.reviewQuestion', { number: index + 1 }) }} · {{ review.level }}</span>
            <span>{{ t(review.correct ? 'placement.correct' : 'placement.incorrect') }}</span>
          </summary>
          <p class="placement-context">{{ review.context }}</p>
          <p>{{ review.instruction }}</p>
          <p class="placement-prompt" :lang="targetLanguage">{{ review.prompt }}</p>
          <p><strong>{{ t('placement.yourAnswer') }}:</strong> {{ choiceText(review, review.answer) }}</p>
          <p><strong>{{ t('placement.correctAnswer') }}:</strong> {{ choiceText(review, review.correct_answer) }}</p>
          <p>{{ review.explanation }}</p>
          <router-link v-if="availableChapter(review.chapter_ids)" :to="chapterLink(availableChapter(review.chapter_ids)!)" class="placement-review-link">
            {{ t('placement.reviewInCourse') }} <span aria-hidden="true">→</span>
          </router-link>
        </details>
      </section>
      <button data-test="retake" class="placement-secondary" type="button" :disabled="busy" @click="start(true)">
        <LgIcon name="refresh" :s="18" aria-hidden="true" /> {{ t('placement.retake') }}
      </button>
    </section>

    <section v-else class="placement-test">
      <div class="placement-progress-header" aria-live="polite">
        <span>{{ t(session.clarifying && questionIndex >= 30 ? 'placement.clarification' : 'placement.baseBlock') }}</span>
        <span>{{ t('placement.savedCount', { count: savedCount, total: session.questions.length }) }}</span>
      </div>
      <progress :value="savedCount" :max="session.questions.length" :aria-label="t('placement.progress')" />
      <p v-if="session.clarifying && questionIndex >= 30" class="placement-note">{{ t('placement.clarificationNote') }}</p>

      <form v-if="question" :key="question.id" class="placement-card placement-question" @submit.prevent="saveAnswer">
        <p class="placement-eyebrow">{{ t('placement.questionNumber', { number: questionIndex + 1, total: session.questions.length }) }}</p>
        <h1 ref="questionHeading" tabindex="-1">{{ question.instruction }}</h1>
        <p v-if="question.context" class="placement-context">{{ question.context }}</p>
        <p class="placement-prompt" :lang="targetLanguage">{{ question.prompt }}</p>
        <fieldset :disabled="busy || conflict || expired">
          <legend class="placement-sr-only">{{ t('placement.chooseAnswer') }}</legend>
          <label v-for="choice in question.choices" :key="choice.id" class="placement-choice" :class="{ selected: selection === choice.id }">
            <input v-model="selection" type="radio" name="answer" :value="choice.id" />
            <span :lang="targetLanguage">{{ choice.text }}</span>
          </label>
          <label class="placement-choice placement-unknown" :class="{ selected: selection === '' }">
            <input v-model="selection" data-test="unknown" type="radio" name="answer" value="" />
            <span>{{ t('placement.dontKnow') }}</span>
          </label>
        </fieldset>
        <div class="placement-question-actions">
          <button v-if="questionIndex > minQuestionIndex" class="placement-secondary" type="button" :disabled="busy || conflict || expired" @click="previousQuestion">
            <LgIcon name="chevron-left" :s="18" aria-hidden="true" /> {{ t('placement.previous') }}
          </button>
          <button data-test="save" class="placement-primary" type="submit" :disabled="selection === null || busy || conflict || expired">
            {{ busy ? t('placement.saving') : t('placement.saveNext') }}
            <LgIcon name="chevron-right" :s="18" aria-hidden="true" />
          </button>
        </div>
        <p class="placement-fine" aria-live="polite">{{ t(errorKey && draft?.pending ? 'placement.unsentAnswer' : 'placement.savedAutomatically') }}</p>
      </form>

      <div v-else class="placement-card placement-finish">
        <LgIcon name="check" :s="32" aria-hidden="true" />
        <h1 ref="questionHeading" tabindex="-1">{{ t('placement.allSaved') }}</h1>
        <p>{{ t('placement.finishDescription') }}</p>
        <button data-test="finish" class="placement-primary" type="button" :disabled="busy || expired || conflict" @click="finish">
          {{ busy ? t('placement.loading') : t('placement.showResult') }}
        </button>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LgIcon from '../components/linglow/LgIcon.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { placementClient, type PlacementQuestion, type PlacementSession } from '../api/placementClient'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { clearCategoriesCache } from '../api/grammarClient'

interface Draft {
  key: string
  sessionId?: string
  newAttempt?: boolean
  pending?: { questionID: string; answer: string }
}

const { t } = useI18n()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { me, ensureMe } = useMe()
const courseCode = computed(() => currentCourseCode.value)
const targetLanguage = computed(() => courseCode.value.split('_')[0])
const supported = computed(() => ['en_ru', 'es_ru'].includes(courseCode.value))
const storageKey = computed(() => me.value?.id && courseCode.value ? `linglow.placement.v1:${me.value.id}:${courseCode.value}` : '')
const ready = ref(false)
const busy = ref(false)
const errorKey = ref('')
const expired = ref(false)
const conflict = ref(false)
const session = ref<PlacementSession | null>(null)
const draft = ref<Draft | null>(null)
const questionIndex = ref(0)
const selection = ref<string | null>(null)
const questionHeading = ref<HTMLElement | null>(null)
let generation = 0
const savedCount = computed(() => Object.keys(session.value?.answers || {}).length)
const minQuestionIndex = computed(() => session.value?.base_closed ? 30 : 0)
const question = computed(() => session.value?.questions[questionIndex.value])
const resultLevel = computed(() => {
  const result = session.value?.result
  if (!result) return ''
  const lower = result.level === 'below_a1' ? t('placement.belowA1') : result.level
  return result.upper_level && result.upper_level !== result.level ? `${lower}–${result.upper_level}` : lower
})

function readDraft(): Draft | null {
  try {
    const value = JSON.parse(localStorage.getItem(storageKey.value) || 'null')
    if (!value || typeof value.key !== 'string') return null
    if (value.sessionId !== undefined && typeof value.sessionId !== 'string') return null
    if (value.pending && (typeof value.pending.questionID !== 'string' || typeof value.pending.answer !== 'string')) delete value.pending
    return value
  } catch { return null }
}
function persist() {
  if (!storageKey.value || !draft.value) return
  try { localStorage.setItem(storageKey.value, JSON.stringify(draft.value)) } catch { /* server resume remains available */ }
}
function clearError() { errorKey.value = ''; expired.value = false; conflict.value = false }
function showError(error: unknown) {
  const code = (error as { code?: string })?.code
  expired.value = code === 'placement_expired' || code === 'placement_not_found'
  conflict.value = code === 'placement_conflict'
  errorKey.value = expired.value ? 'placement.expired' : conflict.value ? 'placement.conflict' : 'placement.networkError'
  if (code === 'placement_unavailable' || code === 'placement_course_not_found') errorKey.value = 'placement.unavailable'
}
function resetView() {
  generation++
  busy.value = false
  session.value = null
  questionIndex.value = 0
  selection.value = null
  clearError()
  draft.value = readDraft()
  if (!supported.value) errorKey.value = 'placement.unavailable'
}
async function initialize() {
  busy.value = true
  clearError()
  await Promise.all([ensureCourseLoaded(), ensureMe()])
  ready.value = !!storageKey.value
  resetView()
  if (!ready.value) errorKey.value = 'placement.networkError'
}
function hasAnswer(answers: Record<string, string>, id: string) {
  return Object.prototype.hasOwnProperty.call(answers, id)
}
function choiceText(q: PlacementQuestion, answer: string) {
  return answer === '' ? t('placement.dontKnow') : q.choices.find(choice => choice.id === answer)?.text || t('placement.notAnswered')
}
function availableChapter(ids: string[] = []) {
  return ids.find(id => session.value?.available_chapter_ids?.includes(id))
}
function chapterLink(chapterId: string) {
  return { name: 'GrammarChapter', params: { chapterId }, query: { course_code: courseCode.value } }
}
function setQuestion(index: number) {
  questionIndex.value = index
  const id = session.value?.questions[index]?.id
  selection.value = id && hasAnswer(session.value?.answers || {}, id) ? session.value!.answers[id] : null
  if (id && draft.value?.pending?.questionID === id) selection.value = draft.value.pending.answer
  void nextTick(() => questionHeading.value?.focus())
}
function accept(v: PlacementSession) {
  if (v.course_code !== courseCode.value) throw new Error('Course changed')
  if (v.status === 'abandoned') throw Object.assign(new Error('Attempt ended'), { code: 'placement_expired' })
  session.value = v
  if (draft.value) {
    draft.value.sessionId = v.id
    draft.value.newAttempt = false
    const pending = draft.value.pending
    if (pending) {
      const pendingIndex = v.questions.findIndex(q => q.id === pending.questionID)
      const alreadySaved = hasAnswer(v.answers || {}, pending.questionID) && v.answers[pending.questionID] === pending.answer
      if (alreadySaved || v.status === 'completed' || (v.base_closed && pendingIndex < 30)) delete draft.value.pending
    }
    persist()
  }
  let index = v.questions.findIndex(q => !hasAnswer(v.answers || {}, q.id))
  if (index < 0) index = v.questions.length
  const pendingIndex = v.questions.findIndex(q => q.id === draft.value?.pending?.questionID)
  if (pendingIndex >= (v.base_closed ? 30 : 0)) index = pendingIndex
  setQuestion(index)
}
async function start(newAttempt = false) {
  if (busy.value || !ready.value || !supported.value) return
  const run = generation
  const course = courseCode.value
  busy.value = true
  clearError()
  try {
    if (!draft.value || (newAttempt && !draft.value.newAttempt)) {
      draft.value = { key: crypto.randomUUID(), newAttempt }
      persist()
    }
    const d = draft.value
    const v = d.sessionId && !newAttempt
      ? await placementClient.get(d.sessionId, course)
      : await placementClient.start(course, d.key, !!d.newAttempt)
    if (generation === run) accept(v)
  } catch (error) { if (generation === run) showError(error) }
  finally { if (generation === run) busy.value = false }
}
async function saveAnswer() {
  if (!session.value || !question.value || selection.value === null || busy.value) return
  const run = generation
  const course = courseCode.value
  const questionID = question.value.id
  const answer = selection.value
  busy.value = true
  clearError()
  if (draft.value) { draft.value.pending = { questionID, answer }; persist() }
  try {
    const v = await placementClient.answer(session.value.id, course, questionID, answer)
    if (generation === run) accept(v)
  } catch (error) { if (generation === run) showError(error) }
  finally { if (generation === run) busy.value = false }
}
function previousQuestion() {
  if (questionIndex.value > minQuestionIndex.value) setQuestion(questionIndex.value - 1)
}
async function finish() {
  if (!session.value || busy.value) return
  const run = generation
  const course = courseCode.value
  busy.value = true
  clearError()
  try {
    const v = await placementClient.finish(session.value.id, course)
    if (generation === run) {
      accept(v)
      clearCategoriesCache()
      emitAppDataEvent('grammar-test-submitted', course)
    }
  } catch (error) { if (generation === run) showError(error) }
  finally { if (generation === run) busy.value = false }
}
watch(storageKey, () => {
  if (ready.value) { resetView(); ready.value = !!storageKey.value }
})
onMounted(initialize)
onUnmounted(() => { generation++ })
</script>

<style scoped>
.placement-page { max-width: 780px; margin: 0 auto; padding: 24px 20px 64px; color: var(--text-primary, #25313a); }
.placement-header { display: flex; justify-content: space-between; align-items: center; gap: 16px; margin-bottom: 24px; }
.placement-back { display: inline-flex; align-items: center; gap: 4px; color: inherit; font-size: 14px; text-decoration: none; }
.placement-language { font-size: 13px; font-weight: 700; padding: 6px 12px; border-radius: 20px; background: var(--bg-tertiary, #edf3f0); }
.placement-card { padding: 32px; border: 1px solid var(--border-color, #dce5df); border-radius: 24px; background: var(--bg-secondary, #fff); box-shadow: 0 8px 30px #172d2010; }
.placement-card + .placement-card { margin-top: 20px; }
.placement-welcome h1 { font-size: clamp(28px, 6vw, 38px); line-height: 1.18; margin: 10px 0 20px; }
.placement-symbol { display: inline-flex; padding: 18px; border-radius: 20px; background: #eaf5e9; color: #3c7148; margin-bottom: 20px; }
.placement-eyebrow { font-size: 12px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; color: var(--text-secondary, #65706a); }
.placement-lead { font-size: 18px; line-height: 1.65; }
.placement-facts { list-style: none; padding: 16px 0; display: grid; gap: 16px; }
.placement-facts li { display: flex; align-items: start; gap: 12px; line-height: 1.5; }
.placement-facts svg { flex-shrink: 0; margin-top: 2px; }
.placement-note { padding: 16px 18px; background: var(--bg-tertiary, #f0f5f1); border-radius: 14px; line-height: 1.6; }
.placement-fine { color: var(--text-secondary, #65706a); font-size: 13px; line-height: 1.6; }
.placement-primary, .placement-secondary { display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: 48px; padding: 12px 20px; border-radius: 14px; border: 1px solid transparent; cursor: pointer; font: inherit; font-weight: 650; text-decoration: none; }
.placement-primary { background: #376d45; color: #fff; }
.placement-secondary { background: var(--bg-secondary, #fff); color: inherit; border-color: var(--border-color, #ccd9cf); }
button:disabled { cursor: default; opacity: .55; }
button:focus-visible, a:focus-visible, summary:focus-visible { outline: 3px solid #91b849; outline-offset: 4px; }
.placement-progress-header { display: flex; justify-content: space-between; gap: 12px; font-size: 14px; margin-bottom: 10px; }
progress { width: 100%; height: 8px; border: 0; border-radius: 8px; overflow: hidden; margin-bottom: 24px; accent-color: #376d45; background: var(--bg-tertiary, #e4ebe6); }
progress::-webkit-progress-bar { background: var(--bg-tertiary, #e4ebe6); }
progress::-webkit-progress-value { background: #376d45; }
.placement-question h1 { font-size: 23px; line-height: 1.4; margin: 12px 0 22px; }
.placement-context { color: var(--text-secondary, #65706a); white-space: pre-line; line-height: 1.65; }
.placement-prompt { font-size: 21px; font-weight: 550; line-height: 1.65; white-space: pre-line; margin: 22px 0; }
fieldset { border: 0; padding: 0; margin: 0; display: grid; gap: 10px; min-width: 0; }
.placement-choice { display: flex; align-items: center; gap: 12px; padding: 16px; min-height: 54px; border: 1px solid var(--border-color, #dce5df); border-radius: 14px; cursor: pointer; line-height: 1.5; overflow-wrap: anywhere; }
.placement-choice.selected { border-color: #376d45; box-shadow: inset 0 0 0 1px #376d45; background: var(--bg-tertiary, #f1f7f1); }
.placement-choice:focus-within { outline: 3px solid #91b849; outline-offset: 2px; }
.placement-choice input { flex-shrink: 0; margin: 0; width: 18px; height: 18px; accent-color: #376d45; }
.placement-unknown { margin-top: 6px; color: var(--text-secondary, #65706a); }
.placement-question-actions { display: flex; gap: 10px; justify-content: space-between; margin-top: 26px; flex-wrap: wrap; }
.placement-question-actions .placement-primary { margin-left: auto; }
.placement-error { border: 1px solid #c86b57; border-radius: 14px; padding: 12px 20px; margin-bottom: 20px; line-height: 1.5; background: var(--bg-secondary, #fff); }
.placement-error button { color: inherit; background: none; border: 0; text-decoration: underline; padding: 8px 0; font: inherit; cursor: pointer; }
.placement-finish { text-align: center; line-height: 1.6; }
.placement-result-intro h1 { font-size: 52px; margin: 12px 0; color: #4c8559; }
.placement-results h2 { font-size: 22px; margin: 0 0 16px; }
.placement-results > button { margin-top: 24px; }
.placement-level { display: grid; grid-template-columns: 45px 50px 1fr; gap: 14px; align-items: center; padding: 16px 0; border-bottom: 1px solid var(--border-color, #dce5df); }
.placement-level-status { font-size: 14px; justify-self: end; text-align: right; }
.placement-recommendations { padding-left: 22px; line-height: 1.6; }
.placement-recommendations li { padding-bottom: 22px; }
.placement-recommendations p { margin: 4px 0 10px; }
.placement-recommendations a { display: block; margin-right: 18px; color: var(--text-primary, #315f3b); text-decoration: underline; }
.placement-review-link { color: inherit; text-decoration: underline; }
.placement-review { padding: 16px 0; border-bottom: 1px solid var(--border-color, #dce5df); line-height: 1.6; }
.placement-review summary { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.placement-review summary span:last-child { font-size: 13px; margin-left: auto; }
.placement-review summary svg { flex-shrink: 0; }
.placement-sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0,0,0,0); white-space: nowrap; border: 0; }
@media (max-width: 520px) {
  .placement-page { padding: 16px 12px 40px; }
  .placement-card { padding: 22px 18px; border-radius: 20px; }
  .placement-language { max-width: 42%; text-align: center; }
  .placement-lead { font-size: 16px; }
  .placement-primary { width: 100%; box-sizing: border-box; }
  .placement-question h1 { font-size: 21px; }
  .placement-prompt { font-size: 19px; }
  .placement-question-actions { flex-direction: column; }
  .placement-level { grid-template-columns: 34px 42px 1fr; gap: 8px; }
  .placement-review summary { flex-wrap: wrap; }
}
</style>
