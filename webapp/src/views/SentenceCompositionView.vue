<template>
  <div class="sc lg-page">
    <LgPageHeader :title="t('sentence.title')" :show-back="true" @back="goBack" />

    <div v-if="loading" class="sc-state">{{ t('common.loading') }}</div>

    <div v-else-if="!available && !finished" class="sc-state">
      <p>{{ t('sentence.noSet') }}</p>
      <button class="lg-btn" type="button" @click="goBack">{{ t('common.back') }}</button>
    </div>

    <div v-else-if="finished" class="sc-summary">
      <h2 class="sc-summary-title">{{ t('sentence.doneTitle') }}</h2>
      <div class="sc-summary-stats">
        <div class="sc-stat sc-stat--star"><span class="sc-stat-num">{{ setState.stars }}</span>★</div>
        <div class="sc-stat sc-stat--passed"><span class="sc-stat-num">{{ setState.passed }}</span>✓</div>
        <div class="sc-stat sc-stat--failed"><span class="sc-stat-num">{{ setState.failed }}</span>✗</div>
      </div>
      <button class="lg-btn" type="button" @click="goBack">{{ t('common.back') }}</button>
    </div>

    <div v-else-if="current" class="sc-card">
      <div class="sc-progress">{{ t('sentence.progress', { done: attempted, total }) }}</div>

      <div class="sc-prompt">{{ current.prompt_ru }}</div>

      <textarea
        v-model="input"
        class="sc-input"
        :placeholder="t('sentence.inputPlaceholder')"
        :disabled="grading || !!result"
        rows="2"
        @keydown.ctrl.enter="submit"
        @input="markInput"
      />

      <div v-if="result" class="sc-result">
        <SentenceGrading :grade="result" />
        <button class="lg-btn" type="button" @click="next">{{ t('sentence.nextBtn') }}</button>
      </div>
      <div v-else class="sc-actions">
        <button class="lg-btn" type="button" :disabled="grading || !input.trim()" @click="submit">
          {{ grading ? t('sentence.checking') : t('sentence.checkBtn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import SentenceGrading from '../components/SentenceGrading.vue'
import { sentenceClient, type SentenceGrade, type SentenceItem, type SentenceSetState } from '../api/sentenceClient'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'

const router = useRouter()
const { t } = useI18n()
const { currentCourseCode } = useCourse()
const { ensureMe, hasFeature } = useMe()

const loading = ref(true)
const available = ref(false)
const finished = ref(false)
const grading = ref(false)
const current = ref<SentenceItem | null>(null)
const result = ref<SentenceGrade | null>(null)
const input = ref('')
const total = ref(0)
const attempted = ref(0)
const setState = ref<SentenceSetState>({ status: '', stars: 0, passed: 0, failed: 0, total: 0, attempted: 0 })

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
}
.sc-state {
  text-align: center;
  padding: 48px 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  align-items: center;
}
.sc-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}
.sc-progress {
  font-size: 0.85rem;
  color: var(--lg-text-secondary, #6b6b70);
}
.sc-prompt {
  font-size: 1.4rem;
  font-weight: 600;
  line-height: 1.4;
}
.sc-input {
  width: 100%;
  font-size: 1.15rem;
  padding: 12px;
  border: 1px solid var(--lg-border, #d7d7db);
  border-radius: 10px;
  resize: vertical;
}
.sc-result,
.sc-actions {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.sc-summary {
  text-align: center;
  padding: 40px 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  align-items: center;
}
.sc-summary-stats {
  display: flex;
  gap: 24px;
  font-size: 1.5rem;
}
.sc-stat-num {
  font-weight: 700;
  margin-right: 4px;
}
.sc-stat--star { color: #d9a400; }
.sc-stat--passed { color: #1f9d57; }
.sc-stat--failed { color: #d23b3b; }
</style>
