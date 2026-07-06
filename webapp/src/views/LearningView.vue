<template>
  <div class="practice lg-page">
    <!-- Brand + screen title -->
    <div class="practice-header">
      <div class="practice-title-wrap">
        <span class="practice-brand">Linglow.</span>
        <span class="practice-title">{{ t('learning.title') }}</span>
      </div>
      <LgLumi :size="52" pose="teacher" />
    </div>
    <p class="practice-sub">{{ t('lg.practiceSub') }}</p>

    <p v-if="isOffline" class="lg-card lg-section-gap lg-muted">
      {{ t('offline.learningNote') }}
    </p>

    <!-- Quick launches -->
    <div class="practice-quick">
      <router-link v-if="!isOffline" to="/training" class="lg-list-row">
        <div class="lg-icon-box"><LgActivityIcon type="words" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('navigation.training') }}</div>
          <div class="lg-list-row-sub">{{ t('lg.quickTrainingSub') }}</div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
      <router-link to="/learning/grammar/training" class="lg-list-row">
        <div class="lg-icon-box"><LgActivityIcon type="grammar" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('lg.quickGrammarTraining') }}</div>
          <div class="lg-list-row-sub">{{ t('lg.quickGrammarTrainingSub') }}</div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
      <router-link
        v-if="!isOffline && showSpanishVerbFormsTraining"
        to="/training/verbs?start=1"
        class="lg-list-row"
      >
        <div class="lg-icon-box"><LgActivityIcon type="grammar" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('lg.quickVerbFormsTraining') }}</div>
          <div class="lg-list-row-sub">
            <template v-if="verbFormsTotalCardsPool !== null">
              {{ t('verbTraining.totalCardsAvailable', { count: verbFormsTotalCardsPool }) }}
            </template>
            <template v-else>{{ t('lg.quickVerbFormsTrainingSub') }}</template>
          </div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
      <router-link v-if="lastGrammarChapter" :to="lastGrammarChapter.url" class="lg-list-row">
        <div class="lg-icon-box"><LgActivityIcon type="grammar" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('lg.quickStudyGrammar') }}</div>
          <div class="lg-list-row-sub">{{ lastGrammarChapter.title }}</div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
      <router-link v-if="!isOffline && sentenceAvailable" to="/training/sentences" class="lg-list-row">
        <div class="lg-icon-box"><LgActivityIcon type="conversation" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('sentence.shortTitle') }}</div>
          <div class="lg-list-row-sub">{{ t('sentence.dashboardSub') }}</div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
    </div>

    <!-- Mode grid 2×2 (+ optional Pro modes) -->
    <div
      class="practice-modes"
      :class="{ 'practice-modes--extended': conversationPro || picturePro }"
    >
      <router-link
        v-for="mode in modes"
        :key="mode.title"
        :to="mode.to"
        class="practice-mode"
        :class="{ 'practice-mode--disabled': mode.disabled }"
        @click.capture="(e: Event) => { if (mode.disabled) e.preventDefault() }"
      >
        <div class="practice-mode-icon" :style="{ background: mode.bg }">
          <LgActivityIcon :type="mode.type" status="green" :size="34" />
        </div>
        <div class="practice-mode-title">{{ mode.title }}</div>
        <div class="practice-mode-desc">{{ mode.desc }}</div>
        <img class="practice-mode-art" :src="mode.art" alt="" />
      </router-link>
    </div>

    <!-- My dictionary -->
    <router-link v-if="!isOffline" to="/vocab" class="lg-card practice-dict">
      <div class="practice-dict-left">
        <div class="practice-dict-emoji"><LgActivityIcon type="reading" status="green" :size="28" /></div>
        <div class="practice-dict-title">{{ t('lg.myDictionary') }}</div>
        <div class="practice-dict-body" :class="{ 'practice-dict-body--stats': showDictionaryStats }">
          <div v-show="showDictionaryStats" class="practice-dict-stats">
          <div class="practice-dict-stats__head">
            <span class="practice-dict-stats__total">
              {{ vocabStats.total }} {{ (t as any)('common.words', vocabStats.total) }}
            </span>
          </div>
          <div class="practice-dict-stats__grid">
            <div
              v-for="item in vocabBreakdownItems"
              :key="item.key"
              class="practice-dict-stats__item"
              :class="`practice-dict-stats__item--${item.key}`"
            >
              <span class="practice-dict-stats__count">{{ item.count }}</span>
              <span class="practice-dict-stats__label">{{ item.label }}</span>
            </div>
          </div>
          </div>
          <div v-show="!showDictionaryStats" class="practice-dict-sub">{{ t('lg.myDictionarySub') }}</div>
        </div>
      </div>
      <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
    </router-link>

    <!-- Lumi fact -->
    <LgLumiFact :lumi-size="46" context="practice" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import LgLumi from '../components/linglow/LgLumi.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import { useMe } from '../composables/useMe'
import { useCourse } from '../composables/useCourse'
import { useGrammarContinueChapter } from '../composables/useGrammarContinueChapter'
import { ensureLearningLoaded } from '../composables/useLearningConfig'
import { useSpanishVerbFormsPractice } from '../composables/useSpanishVerbFormsPractice'
import { apiClient } from '../api/client'
import { useCachedOverviewScreen } from '../composables/useCachedOverviewScreen'
import { useLocale } from '../composables/useLocale'
import {
  syncPracticeProGatesFromMeCache,
  usePracticeScreenState,
  type VocabSummaryPayload,
} from '../composables/usePracticeScreenState'
import { sentenceClient } from '../api/sentenceClient'
import artWords from '../assets/linglow/art/bg-word-cards-440.jpg'
import artGrammar from '../assets/linglow/art/bg-grammar-440.jpg'
import artReading from '../assets/linglow/art/bg-read-440.jpg'
import artConversation from '../assets/linglow/art/bg-conversation-440.jpg'
import artPictureQuest from '../assets/linglow/art/bg-picture-quest-440.jpg'

const { t } = useI18n()
const { ensureMe, hasFeature } = useMe()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { currentLocale } = useLocale()
syncPracticeProGatesFromMeCache()
const {
  vocabStats,
  vocabStatsLoaded,
  sentenceAvailable,
  conversationPro,
  picturePro,
  applyVocabSummary,
  applySentenceToday,
  resetPracticeScreenState,
} = usePracticeScreenState()
const isOffline = ref(typeof navigator !== 'undefined' && navigator.onLine === false)
const isOnline = computed(() => !isOffline.value)
const {
  verbFormsTotalCardsPool,
  showSpanishVerbFormsTraining,
  refreshVerbFormsPoolCount,
  applyVerbFormsPool,
  resetVerbFormsPool,
} = useSpanishVerbFormsPractice(isOnline)
const showDictionaryStats = computed(() => vocabStatsLoaded.value && vocabStats.value.total > 0)

const vocabBreakdownItems = computed(() => [
  { key: 'new', label: t('training.vocabStatusNew'), count: vocabStats.value.newCount },
  { key: 'learning', label: t('training.vocabStatusLearning'), count: vocabStats.value.learningCount },
  { key: 'review', label: t('training.vocabStatusReview'), count: vocabStats.value.reviewCount },
  { key: 'known', label: t('training.vocabStatusKnown'), count: vocabStats.value.masteredCount },
])

const { continueChapter: lastGrammarChapter, loadContinueChapter, applyContinueChapter } = useGrammarContinueChapter()

function applyLearningOverview(ov: any) {
  applyContinueChapter(ov?.continue_chapter)
  applyVerbFormsPool(ov?.verb_upcoming)
  applyVocabSummary(ov?.vocab_summary)
  applySentenceToday(ov?.sentence_today)
}

const { load: loadLearningCache, hydrateFromCache } = useCachedOverviewScreen<any>({
  screenKey: 'learning',
  courseCode: currentCourseCode,
  locale: currentLocale,
  fetcher: async () => apiClient.request(
    currentCourseCode.value
      ? `/api/overview/learning?course_code=${encodeURIComponent(currentCourseCode.value)}`
      : '/api/overview/learning',
  ),
  applyPayload: (ov) => applyLearningOverview(ov),
})

async function loadLearningData(force = false) {
  await loadLearningCache(force)
  const meProfile = await ensureMe()
  if (meProfile?.features) {
    conversationPro.value = !!meProfile.features.conversation
    picturePro.value = !!meProfile.features.picture_description
  }
}

interface PracticeMode {
  type: 'words' | 'grammar' | 'reading' | 'conversation'
  bg: string
  title: string
  desc: string
  art: string
  to: string
  disabled: boolean
}

const modes = computed(() => {
  const items: PracticeMode[] = [
    {
      type: 'words',
      bg: 'rgba(45,107,58,0.10)',
      title: t('learning.words'),
      desc: t('learning.wordsDescription'),
      art: artWords,
      to: '/learning/words',
      disabled: isOffline.value,
    },
    {
      type: 'grammar',
      bg: 'rgba(45,107,58,0.10)',
      title: t('learning.grammar'),
      desc: t('learning.grammarDescription'),
      art: artGrammar,
      to: '/learning/grammar',
      disabled: false,
    },
    {
      type: 'reading',
      bg: 'rgba(45,107,58,0.10)',
      title: t('learning.reading'),
      desc: t('learning.readingDescription'),
      art: artReading,
      to: '/learning/reading',
      disabled: isOffline.value,
    },
  ]
  if (conversationPro.value) {
    items.push({
      type: 'conversation',
      bg: 'rgba(45,107,58,0.10)',
      title: t('learning.conversation'),
      desc: t('learning.conversationDescription'),
      art: artConversation,
      to: '/learning/conversations',
      disabled: isOffline.value,
    })
  }
  if (picturePro.value) {
    items.push({
      type: 'conversation',
      bg: 'rgba(45,107,58,0.10)',
      title: t('learning.pictureQuest'),
      desc: t('learning.pictureQuestDescription'),
      art: artPictureQuest,
      to: '/learning/picture-quests',
      disabled: isOffline.value,
    })
  }
  return items
})

const handleNetworkChange = () => {
  isOffline.value = typeof navigator !== 'undefined' && navigator.onLine === false
}

// applySentenceToday sets availability from the aggregate's sentence_today part (available=true
// already implies the feature is enabled), so the initial load needs no separate /me + /today.
async function loadSentenceAvailability() {
  if (isOffline.value) {
    sentenceAvailable.value = false
    return
  }
  if (!hasFeature('sentence_composition')) {
    sentenceAvailable.value = false
    return
  }
  try {
    const today = await sentenceClient.today(currentCourseCode.value)
    applySentenceToday(today)
  } catch {
    sentenceAvailable.value = false
  }
}

async function loadVocabSummary(raw?: VocabSummaryPayload) {
  if (isOffline.value) return
  try {
    const summary = raw ?? await apiClient.request<VocabSummaryPayload>(
      `/api/vocab/summary${currentCourseCode.value ? `?course_code=${encodeURIComponent(currentCourseCode.value)}` : ''}`,
    )
    applyVocabSummary(summary)
  } catch { /* ignore */ }
}

void ensureCourseLoaded().then(() => {
  if (currentCourseCode.value) void hydrateFromCache()
})

watch(currentCourseCode, (code, prev) => {
  if (!code) return
  if (prev && prev !== code) {
    resetPracticeScreenState()
    resetVerbFormsPool()
  }
  void hydrateFromCache()
})

onMounted(() => {
  window.addEventListener('online', handleNetworkChange)
  window.addEventListener('offline', handleNetworkChange)
  void (async () => {
    await Promise.all([ensureCourseLoaded(), ensureLearningLoaded()])
    try {
      await loadLearningData()
    } catch {
      await Promise.all([loadContinueChapter(), refreshVerbFormsPoolCount(), loadVocabSummary(), loadSentenceAvailability()])
    }
  })()
})

watch(currentCourseCode, async () => {
  await ensureLearningLoaded()
  await loadLearningData(true)
})

watch(isOffline, async (offline) => {
  if (!offline) {
    await refreshVerbFormsPoolCount()
  }
})

onUnmounted(() => {
  window.removeEventListener('online', handleNetworkChange)
  window.removeEventListener('offline', handleNetworkChange)
})
</script>

<style scoped>
.practice {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.practice-header {
  padding: 16px 0 0;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}
.practice-title-wrap {
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: wrap;
}
.practice-brand {
  font-family: 'Lora', serif;
  font-size: 34px;
  color: var(--text);
  letter-spacing: -0.02em;
  line-height: 1;
}
.practice-title {
  font-family: 'Lora', serif;
  font-size: 34px;
  font-weight: 600;
  color: var(--text);
  letter-spacing: -0.03em;
  line-height: 1;
}
.practice-sub {
  margin: 0 4px;
  font-size: 13px;
  line-height: 1.3;
  color: var(--subtext);
}

.practice-quick {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 4px;
}
.quick-text { flex: 1; }

.practice-modes {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 7px;
  min-height: 231px;
}
.practice-modes--extended {
  min-height: 350px;
}
.practice-mode {
  position: relative;
  min-height: 112px;
  padding: 12px;
  border-radius: 18px;
  background: var(--card-bg);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-soft);
  overflow: hidden;
  cursor: pointer;
  text-align: left;
  text-decoration: none;
}
.practice-mode--disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.practice-mode-icon {
  position: relative;
  z-index: 1;
  width: 36px;
  height: 36px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}
.practice-mode-title {
  position: relative;
  z-index: 1;
  margin-top: 6px;
  font-family: 'Lora', serif;
  font-size: 15px;
  line-height: 1.1;
  font-weight: 600;
  color: var(--text);
}
.practice-mode-desc {
  position: relative;
  z-index: 1;
  margin-top: 2px;
  max-width: 60%;
  font-size: 10px;
  line-height: 1.3;
  color: var(--subtext);
}
.practice-mode-art {
  position: absolute;
  right: 0;
  top: 0;
  width: 55%;
  height: 100%;
  object-fit: cover;
  object-position: top;
  opacity: 0.40;
  border-radius: 0 18px 18px 0;
  -webkit-mask-image: linear-gradient(to right, transparent 0%, black 28%);
  mask-image: linear-gradient(to right, transparent 0%, black 28%);
}
:root[data-theme="dark"] .practice-mode-art { opacity: 0.40; }

.practice-dict {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  position: relative;
  overflow: hidden;
  background: var(--card-bg);
}
.practice-dict::after {
  content: "";
  position: absolute;
  inset: 0 0 0 auto;
  width: 62%;
  background: url('../assets/linglow/art/bg-dictionary-440.jpg') right center / cover no-repeat;
  opacity: 0.42;
  -webkit-mask-image: linear-gradient(to right, transparent 0%, transparent 12%, black 62%);
  mask-image: linear-gradient(to right, transparent 0%, transparent 12%, black 62%);
  pointer-events: none;
}
.practice-dict > * {
  position: relative;
  z-index: 1;
}
.practice-dict-left { flex: 1; }
.practice-dict-emoji { font-size: 24px; line-height: 1; }
.practice-dict-title {
  font-family: 'Lora', serif;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
  margin-top: 4px;
}
.practice-dict-body {
  margin-top: 2px;
}
.practice-dict-body--stats {
  min-height: 88px;
}
.practice-dict-sub {
  font-size: 11px;
  color: var(--subtext);
  margin-top: 2px;
}

.practice-dict-stats {
  margin-top: 8px;
}

.practice-dict-stats__head {
  margin-bottom: 8px;
}

.practice-dict-stats__total {
  font-size: 12px;
  color: var(--subtext);
}

.practice-dict-stats__grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 6px;
}

.practice-dict-stats__item {
  min-width: 0;
  border-radius: 10px;
  padding: 8px 6px;
  background: var(--bg-secondary, rgba(0, 0, 0, 0.04));
  text-align: center;
}

.practice-dict-stats__item--new {
  background: color-mix(in srgb, #3b82f6 12%, var(--card-bg));
}

.practice-dict-stats__item--learning {
  background: color-mix(in srgb, #f59e0b 12%, var(--card-bg));
}

.practice-dict-stats__item--review {
  background: color-mix(in srgb, #8b5cf6 12%, var(--card-bg));
}

.practice-dict-stats__item--known {
  background: color-mix(in srgb, var(--color-primary, #2d6b3a) 12%, var(--card-bg));
}

.practice-dict-stats__count {
  display: block;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.practice-dict-stats__label {
  display: block;
  margin-top: 2px;
  font-size: 9px;
  line-height: 1.2;
  color: var(--subtext);
}
</style>
