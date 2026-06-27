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
      <router-link v-if="lastGrammarChapter" :to="lastGrammarChapter.url" class="lg-list-row">
        <div class="lg-icon-box"><LgActivityIcon type="grammar" status="green" :size="28" /></div>
        <div class="quick-text">
          <div class="lg-list-row-title">{{ t('lg.quickLastGrammarChapter') }}</div>
          <div class="lg-list-row-sub">{{ lastGrammarChapter.title }}</div>
        </div>
        <LgIcon name="chevron-right" :s="14" c="var(--subtext)" />
      </router-link>
    </div>

    <!-- Mode grid 2×2 -->
    <div class="practice-modes">
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
        <div class="practice-dict-sub">{{ vocabSummaryText || t('lg.myDictionarySub') }}</div>
      </div>
      <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
    </router-link>

    <!-- AI conversations (Pro only) -->
    <router-link v-if="!isOffline && isPro" to="/city" class="lg-card practice-dict">
      <div class="practice-dict-left">
        <div class="practice-dict-emoji"><LgActivityIcon type="words" status="green" :size="28" /></div>
        <div class="practice-dict-title">{{ t('lg.quickChat') }}</div>
        <div class="practice-dict-sub">{{ t('lg.quickChatSub') }}</div>
      </div>
      <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
    </router-link>

    <!-- Lumi fact -->
    <LgLumiFact :lumi-size="46" context="practice" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LgLumi from '../components/linglow/LgLumi.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import { useMe } from '../composables/useMe'
import { apiClient } from '../api/client'
import artWords from '../assets/linglow/art/bg-word-cards-440.jpg'
import artGrammar from '../assets/linglow/art/bg-grammar-440.jpg'
import artReading from '../assets/linglow/art/bg-read-440.jpg'

const { t } = useI18n()
const { ensureMe, hasFeature } = useMe()
const isOffline = ref(typeof navigator !== 'undefined' && navigator.onLine === false)
const isPro = ref(false)
const vocabSummaryText = ref('')

const lastGrammarChapter = ref<{ id: string; title: string; url: string } | null>(null)

const modes = computed(() => [
  {
    type: 'words' as const,
    bg: 'rgba(45,107,58,0.10)',
    title: t('learning.words'),
    desc: t('learning.wordsDescription'),
    art: artWords,
    to: '/learning/words',
    disabled: isOffline.value,
  },
  {
    type: 'grammar' as const,
    bg: 'rgba(45,107,58,0.10)',
    title: t('learning.grammar'),
    desc: t('learning.grammarDescription'),
    art: artGrammar,
    to: '/learning/grammar',
    disabled: false,
  },
  {
    type: 'reading' as const,
    bg: 'rgba(45,107,58,0.10)',
    title: t('learning.reading'),
    desc: t('learning.readingDescription'),
    art: artReading,
    to: '/learning/reading',
    disabled: isOffline.value,
  },
])

const handleNetworkChange = () => {
  isOffline.value = typeof navigator !== 'undefined' && navigator.onLine === false
}

onMounted(async () => {
  window.addEventListener('online', handleNetworkChange)
  window.addEventListener('offline', handleNetworkChange)
  try {
    const stored = localStorage.getItem('linglow:last-grammar-chapter')
    if (stored) lastGrammarChapter.value = JSON.parse(stored)
  } catch { /* ignore */ }
  ensureMe().then(() => { isPro.value = hasFeature('conversation') })
  if (!isOffline.value) {
    try {
      const summary = await apiClient.request<{ total: number; known: number; learning: number; new: number }>('/api/vocab/summary')
      if (summary && summary.total > 0) {
        vocabSummaryText.value = `${summary.total} слов · ${summary.known} изучено`
      }
    } catch { /* ignore */ }
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
  padding: 16px 4px 0;
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
.practice-dict-sub {
  font-size: 11px;
  color: var(--subtext);
  margin-top: 2px;
}
</style>
