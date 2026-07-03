<template>
  <div class="sc lg-page">
    <LgPageHeader :title="t('sentence.title')" :show-back="true" @back="goBack" />

    <div v-if="loading" class="sc-state">{{ t('common.loading') }}</div>

    <!-- No set available -->
    <div v-else-if="!available && !finished" class="sc-card sc-state-card">
      <div class="sc-empty-icon"><LgIcon name="sprout" :s="40" c="var(--salvia, currentColor)" /></div>
      <p class="sc-empty-text">{{ t('sentence.noSet') }}</p>
      <button class="sc-next" type="button" @click="goBack">{{ t('common.back') }}</button>
    </div>

    <!-- Completed -->
    <div v-else-if="finished" class="sc-card sc-summary">
      <div class="sc-empty-icon"><LgIcon name="party-popper" :s="40" c="var(--salvia, currentColor)" /></div>
      <h2 class="sc-summary-title">{{ t('sentence.doneTitle') }}</h2>
      <div class="sc-summary-stats">
        <div class="sc-stat sc-stat--star"><span class="sc-stat-num">{{ setState.stars }}</span><LgIcon name="star-filled" :s="16" /></div>
        <div class="sc-stat sc-stat--passed"><span class="sc-stat-num">{{ setState.passed }}</span><LgIcon name="check" :s="16" /></div>
        <div class="sc-stat sc-stat--failed"><span class="sc-stat-num">{{ setState.failed }}</span><LgIcon name="x" :s="16" /></div>
      </div>
      <button class="sc-next" type="button" @click="goBack">{{ t('common.back') }}</button>
    </div>

    <!-- Exercise -->
    <main v-else-if="current" class="sc-card">
      <!-- progress -->
      <div class="sc-progress-row">
        <div class="sc-progress-count">{{ attempted }} / {{ total }}</div>
        <div class="sc-progress-track">
          <div class="sc-progress-fill" :style="{ width: progressPct + '%' }" />
        </div>
      </div>

      <!-- task -->
      <section class="sc-task">
        <div class="sc-task-icon"><LgIcon name="lightbulb" :s="22" /></div>
        <div>
          <h1 class="sc-task-title">{{ current.prompt_ru }}</h1>
          <p class="sc-task-sub">{{ t('sentence.taskSubtitle', { lang: targetLangDisplay }) }}</p>
        </div>
      </section>

      <!-- answer field -->
      <section class="sc-answer">
        <textarea
          v-model="input"
          class="sc-answer-input"
          :placeholder="t('sentence.inputPlaceholder')"
          :disabled="grading || !!result"
          rows="2"
          @keydown.ctrl.enter="submit"
          @input="markInput"
        />
        <VoiceMicButton
          :lang="learning?.target_lang ?? 'en'"
          :disabled="grading || !!result"
          :label="t('sentence.voiceInput')"
          @transcript="onVoiceTranscript"
        />
      </section>

      <!-- result -->
      <template v-if="result">
        <!-- verdict + handwriting markup -->
        <section
          class="sc-result-card"
          :class="`sc-result-card--${result.outcome}`"
        >
          <div class="sc-result-title">
            <template v-if="result.outcome === 'star'"><span class="sc-star"><LgIcon name="star-filled" :s="16" /></span> {{ t('sentence.outcomeStar') }}</template>
            <template v-else-if="result.outcome === 'passed'"><LgIcon name="check" :s="16" /> {{ t('sentence.outcomePassed') }}</template>
            <template v-else><LgIcon name="x" :s="16" /> {{ t('sentence.outcomeFailed') }}</template>
          </div>
          <SentenceGrading
            v-if="result.error_count > 0"
            :user-input="resultInput"
            :corrected="result.corrected_es"
          />
        </section>

        <!-- correct answer -->
        <section class="sc-success-card">
          <div class="sc-success-icon"><LgIcon name="check" :s="18" /></div>
          <div>
            <div class="sc-success-label">{{ t('sentence.correctAnswer') }}:</div>
            <div class="sc-success-answer">{{ result.corrected_es }}</div>
          </div>
        </section>

        <!-- explanation -->
        <section v-if="result.explanation" class="sc-explanation-card">
          <div class="sc-explanation-icon"><LgIcon name="lightbulb" :s="22" /></div>
          <div class="sc-explanation-text">{{ result.explanation }}</div>
        </section>

        <button class="sc-next" type="button" @click="next">{{ t('sentence.nextBtn') }}</button>
      </template>

      <button
        v-else
        class="sc-next"
        type="button"
        :disabled="grading || !input.trim()"
        @click="submit"
      >
        {{ grading ? t('sentence.checking') : t('sentence.checkBtn') }}
      </button>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import SentenceGrading from '../components/SentenceGrading.vue'
import { sentenceClient, type SentenceGrade, type SentenceItem, type SentenceSetState } from '../api/sentenceClient'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { useLearningConfig } from '../composables/useLearningConfig'
import VoiceMicButton from '../components/VoiceMicButton.vue'

const router = useRouter()
const { t } = useI18n()
const { currentCourseCode } = useCourse()
const { ensureMe, hasFeature } = useMe()
const { learning, targetLangDisplay } = useLearningConfig()

// Voice input: dictate the answer in the target language when the device/browser
// supports speech recognition.
const onVoiceTranscript = (text: string) => {
  input.value = input.value ? `${input.value.trimEnd()} ${text}` : text
  markInput()
}

const loading = ref(true)
const available = ref(false)
const finished = ref(false)
const grading = ref(false)
const current = ref<SentenceItem | null>(null)
const result = ref<SentenceGrade | null>(null)
const resultInput = ref('') // the exact text the learner submitted, for the correction diff
const input = ref('')
const total = ref(0)
const attempted = ref(0)
const setState = ref<SentenceSetState>({ status: '', stars: 0, passed: 0, failed: 0, total: 0, attempted: 0 })

const progressPct = computed(() => (total.value > 0 ? Math.round((attempted.value / total.value) * 100) : 0))

function goBack() {
  if (typeof window !== 'undefined' && window.history.length > 1) router.back()
  else router.push('/dashboard')
}

// Bump the activity tracker's idle timer so time on this screen counts as active study.
function markInput() {
  document.dispatchEvent(new Event('mousemove'))
}

async function loadCurrent() {
  const cur = await sentenceClient.current(currentCourseCode.value)
  if (cur.done) {
    finished.value = true
    current.value = null
    return
  }
  current.value = cur
  total.value = cur.total ?? total.value
  attempted.value = cur.attempted_count ?? (cur as any).attempted ?? attempted.value
  result.value = null
  input.value = ''
}

async function submit() {
  if (!current.value || grading.value || !input.value.trim()) return
  grading.value = true
  try {
    resultInput.value = input.value.trim()
    const res = await sentenceClient.answer(current.value.id, input.value)
    result.value = res.grading
    setState.value = res.set
    attempted.value = res.set.attempted
    total.value = res.set.total
  } finally {
    grading.value = false
  }
}

async function next() {
  if (setState.value.status === 'completed') {
    finished.value = true
    current.value = null
    return
  }
  await loadCurrent()
}

onMounted(async () => {
  await ensureMe()
  if (!hasFeature('sentence_composition')) {
    router.replace('/dashboard')
    return
  }
  try {
    const today = await sentenceClient.today(currentCourseCode.value)
    if (!today.available) {
      available.value = false
      loading.value = false
      return
    }
    available.value = true
    await sentenceClient.start(currentCourseCode.value)
    await loadCurrent()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.sc {
  max-width: 720px;
  margin: 0 auto;
  /* component-scoped ink colors for the handwriting markup */
  --correct-ink: var(--salvia, #2f7d46);
  --wrong-ink: #c4443c;
}

.sc-state {
  text-align: center;
  padding: 48px 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  align-items: center;
}

/* Main exercise card holds everything */
.sc-card {
  margin: 8px 16px 112px;
  padding: 22px;
  border-radius: 28px;
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sc-state-card,
.sc-summary {
  align-items: center;
  text-align: center;
}
.sc-empty-icon { font-size: 44px; }
.sc-empty-text { color: var(--subtext); font-size: 17px; margin: 0; }

/* progress */
.sc-progress-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.sc-progress-count {
  font-size: 16px;
  color: var(--subtext);
  white-space: nowrap;
}
.sc-progress-track {
  flex: 1;
  height: 8px;
  border-radius: 999px;
  background: var(--progress-track, rgba(32,53,42,0.12));
  overflow: hidden;
}
.sc-progress-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--hoja, #7fae6a), var(--salvia, #3f6f3f));
  transition: width 240ms ease;
}

/* task */
.sc-task {
  display: flex;
  align-items: flex-start;
  gap: 14px;
}
.sc-task-icon {
  width: 52px;
  height: 52px;
  flex: 0 0 52px;
  border-radius: 50%;
  background: var(--chip-bg, rgba(63,111,63,0.10));
  color: var(--salvia, #3f6f3f);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}
.sc-task-title {
  margin: 0;
  font-family: 'Lora', Georgia, serif;
  font-size: 21px;
  line-height: 1.25;
  font-weight: 700;
  color: var(--text);
}
.sc-task-sub {
  margin: 8px 0 0;
  font-size: 15px;
  line-height: 1.35;
  color: var(--subtext);
}

/* answer field */
.sc-answer {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  border-radius: 16px;
  background: var(--input-bg, var(--card-bg));
  border: 1px solid var(--input-border, var(--border));
  box-shadow: var(--shadow-soft);
  padding: 16px;
}
.sc-answer-input {
  flex: 1 1 auto;
  width: 100%;
  min-height: 48px;
  border: 0;
  outline: 0;
  resize: vertical;
  background: transparent;
  font-size: 17px;
  line-height: 1.35;
  color: var(--text);
}
.sc-answer-input::placeholder { color: var(--text-muted); }

/* result cards — cascade in one after another */
.sc-result-card,
.sc-success-card,
.sc-explanation-card {
  border-radius: 16px;
  padding: 18px;
  animation: cardIn 220ms ease-out both;
}
.sc-success-card { animation-delay: 90ms; }
.sc-explanation-card { animation-delay: 180ms; }
@keyframes cardIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.sc-result-card--failed {
  background: var(--wrong-soft, rgba(196,68,60,0.08));
  border: 1px solid rgba(196,68,60,0.22);
}
.sc-result-card--passed {
  background: var(--surface-2);
  border: 1px solid var(--border-green);
}
.sc-result-card--star {
  background: var(--chip-bg, rgba(63,111,63,0.10));
  border: 1px solid var(--border-green);
}
.sc-result-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 12px;
}
.sc-result-card--failed .sc-result-title { color: var(--wrong-ink); }
.sc-result-card--passed .sc-result-title { color: var(--salvia, #3f6f3f); }
.sc-result-card--star .sc-result-title { color: var(--dorado, #d9a83f); }
.sc-result-card--star { margin-bottom: 0; }
.sc-result-card--star .sc-result-title { margin-bottom: 0; }

/* gold star for a flawless answer: twinkle in, then a soft idle shimmer */
.sc-star {
  display: inline-block;
  color: var(--dorado, #d9a83f);
  text-shadow: 0 0 10px rgba(217, 168, 63, 0.55);
  transform-origin: center;
  animation: sc-star-pop 520ms cubic-bezier(0.34, 1.56, 0.64, 1) both,
             sc-star-shine 2.4s ease-in-out 620ms infinite;
}
@keyframes sc-star-pop {
  0% { opacity: 0; transform: scale(0.2) rotate(-40deg); }
  60% { opacity: 1; transform: scale(1.25) rotate(8deg); }
  100% { opacity: 1; transform: scale(1) rotate(0); }
}
@keyframes sc-star-shine {
  0%, 100% { text-shadow: 0 0 8px rgba(217, 168, 63, 0.45); }
  50% { text-shadow: 0 0 16px rgba(217, 168, 63, 0.85); }
}
.sc-result-card--star {
  background: var(--luz-soft, rgba(217, 168, 63, 0.12));
  border-color: rgba(217, 168, 63, 0.35);
}

@media (prefers-reduced-motion: reduce) {
  .sc-star,
  .sc-result-card,
  .sc-success-card,
  .sc-explanation-card,
  .sc-success-icon {
    animation: none;
    opacity: 1;
    transform: none;
  }
}

/* success (correct answer) */
.sc-success-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  background: var(--surface-2);
  border: 1px solid var(--border-green);
}
.sc-success-icon {
  width: 40px;
  height: 40px;
  flex: 0 0 40px;
  border-radius: 50%;
  background: var(--salvia, #3f6f3f);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  animation: checkPop 380ms cubic-bezier(0.34, 1.56, 0.64, 1) 140ms both;
}
@keyframes checkPop {
  0% { opacity: 0; transform: scale(0.3); }
  100% { opacity: 1; transform: scale(1); }
}
.sc-success-label {
  font-size: 15px;
  font-weight: 700;
  color: var(--salvia, #3f6f3f);
  margin-bottom: 6px;
}
.sc-success-answer {
  font-size: 16px;
  line-height: 1.4;
  color: var(--text);
}

/* explanation */
.sc-explanation-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  background: var(--luz-soft, rgba(217,168,63,0.12));
  border: 1px solid rgba(217,168,63,0.30);
}
.sc-explanation-icon {
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  border-radius: 50%;
  background: rgba(217,168,63,0.20);
  color: var(--dorado, #d9a83f);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}
.sc-explanation-text {
  font-size: 16px;
  line-height: 1.45;
  color: var(--text);
}

/* primary button */
.sc-next {
  width: 100%;
  height: 60px;
  border: 1px solid var(--btn-border, transparent);
  border-radius: 16px;
  background: var(--btn-gradient);
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  box-shadow: var(--shadow-soft);
  cursor: pointer;
}
.sc-next:active { transform: translateY(1px); }
.sc-next:disabled { opacity: 0.5; cursor: default; }

/* summary */
.sc-summary-title { margin: 0; font-size: 22px; color: var(--text); }
.sc-summary-stats { display: flex; gap: 24px; font-size: 1.5rem; }
.sc-stat-num { font-weight: 700; margin-right: 4px; }
.sc-stat--star { color: var(--dorado, #d9a83f); }
.sc-stat--passed { color: var(--salvia, #3f6f3f); }
.sc-stat--failed { color: var(--wrong-ink, #c4443c); }

@media (max-width: 360px) {
  .sc-card { margin-left: 12px; margin-right: 12px; padding: 18px; border-radius: 22px; }
  .sc-task-title { font-size: 19px; }
  .sc-answer-input { font-size: 16px; }
  .sc-success-answer { font-size: 15px; }
}
</style>
