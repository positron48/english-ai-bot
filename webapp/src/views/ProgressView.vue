<template>
  <div class="prg-page">
    <!-- HEADER -->
    <div class="prg-header">
      <span class="prg-logo">Linglow</span>
      <span class="prg-star">✦</span>
    </div>

    <div v-if="loading" class="lg-loading">{{ t('common.loading') }}</div>
    <template v-else>

      <!-- MONTH SUMMARY CARD -->
      <div class="prg-summary-card">
        <div class="prg-summary-overlay1" />
        <div class="prg-summary-overlay2" />
        <img :src="mapCityImg" alt="" class="prg-summary-map" aria-hidden="true" />
        <div class="prg-summary-lumi">
          <LgLumi :size="66" pose="default" />
        </div>
        <div class="prg-summary-body">
          <div class="prg-summary-title">{{ t('progress.monthTitle') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-summary-sub">{{ t('progress.monthSub') }}</div>
          <div class="prg-metrics-grid">
            <div v-for="m in metrics" :key="m.label" class="prg-metric-cell">
              <span class="prg-metric-icon">{{ m.icon }}</span>
              <div class="prg-metric-val">{{ m.value }}</div>
              <div class="prg-metric-label">{{ m.label }}</div>
              <div class="prg-metric-sub">{{ m.sub }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ROW: RHYTHM + DISTRICTS -->
      <div class="prg-row2">
        <!-- Rhythm card -->
        <div class="prg-card">
          <div class="prg-card-title">{{ t('progress.rhythmTitle') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-card-sub">{{ t('progress.rhythmSub') }}</div>
          <div class="prg-week-row">
            <div v-for="(d, i) in weekDays" :key="i" class="prg-week-col">
              <div class="prg-week-dot" :class="`prg-week-dot--${weekAct[i]}`">
                <svg v-if="weekAct[i] === 'done'" width="9" height="9" viewBox="0 0 24 24" fill="none">
                  <path d="M5 12L10 17L19 7" stroke="white" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <span v-else class="prg-week-empty">○</span>
              </div>
              <span class="prg-week-label">{{ d }}</span>
            </div>
          </div>
          <div class="prg-streak-row">
            <span class="prg-streak-num">{{ streakDays }}</span>
            <span class="prg-streak-unit">{{ t('progress.daysRow') }}</span>
            <span class="prg-star-s" style="margin-left:4px">✦</span>
          </div>
          <div class="prg-card-sub" style="margin-top:2px">{{ t('progress.keepGoing') }}</div>
        </div>

        <!-- Districts card -->
        <div class="prg-card">
          <div class="prg-card-title">{{ t('progress.districtsTitle') }}</div>
          <div class="prg-card-sub" style="margin-bottom:12px">{{ t('progress.districtsSub') }}</div>
          <div class="prg-districts-list">
            <div v-for="(d, i) in districtItems" :key="i" class="prg-district-row" :style="{ opacity: d.locked ? 0.5 : 1 }">
              <div class="prg-district-head">
                <span class="prg-district-name">{{ d.name }}</span>
                <span class="prg-district-status">{{ d.locked ? '🔒' : d.status }}</span>
              </div>
              <div class="prg-bar-track">
                <div class="prg-bar-fill" :style="{ width: d.pct + '%', background: d.fill }" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ROW: FAVORITE ZONE + STRONGEST SKILL -->
      <div class="prg-row2">
        <div class="prg-card prg-card--rel">
          <div class="prg-card-title-spk">{{ t('progress.favZone') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-fav-name">{{ t('progress.favZoneName') }}</div>
          <div class="prg-card-sub">{{ t('progress.favZoneSub') }}</div>
          <div class="prg-donut-wrap">
            <LgCircleRing :val="learnedPct" :max="100" :size="66" :stroke="7">
              <div class="prg-donut-inner">
                <span class="prg-donut-pct">{{ learnedPct }}%</span>
                <span class="prg-donut-sub">{{ t('progress.pctLabel') }}</span>
              </div>
            </LgCircleRing>
          </div>
          <img :src="mapCityImg" alt="" class="prg-card-bg-img" aria-hidden="true" />
        </div>

        <div class="prg-card prg-card--rel">
          <div class="prg-card-title-spk">{{ t('progress.strongSkill') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-skill-row">
            <div class="prg-skill-icon-box">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <path d="M12 2v4M8 6h8M9 10h6M10.5 14h3" stroke="var(--salvia)" stroke-width="1.8" stroke-linecap="round"/>
                <rect x="6" y="17" width="12" height="3.5" rx="1.8" fill="var(--salvia)" opacity="0.85"/>
              </svg>
            </div>
            <div>
              <div class="prg-skill-name">{{ t('progress.grammarLabel') }}</div>
              <div class="prg-card-sub">{{ t('progress.skillSuperpower') }}</div>
            </div>
          </div>
          <div class="prg-big-num-row">
            <span class="prg-big-num">{{ skillPct }}</span>
            <span class="prg-big-pct">%</span>
            <span class="prg-star-s" style="margin-left:4px">✦</span>
          </div>
          <div class="prg-card-sub">{{ t('progress.levelLabel') }}</div>
        </div>
      </div>

      <!-- ROW: SKILLS + IMPROVEMENTS -->
      <div class="prg-row2">
        <div class="prg-card">
          <div class="prg-card-title">{{ t('progress.strengths') }}</div>
          <div class="prg-skills-list">
            <div v-for="sk in skills" :key="sk.name" class="prg-skill-item">
              <div class="prg-skill-item-head">
                <span class="prg-skill-item-name">{{ sk.name }}</span>
                <span class="prg-skill-item-pct">{{ sk.pct }}%</span>
              </div>
              <div class="prg-bar-track">
                <div class="prg-bar-fill" :style="{ width: sk.pct + '%', background: sk.fill }" />
              </div>
            </div>
          </div>
        </div>

        <div class="prg-card">
          <div class="prg-card-title">{{ t('progress.improvements') }}</div>
          <div class="prg-improve-list">
            <div v-for="(imp, i) in improvements" :key="i" class="prg-improve-row" :class="{ 'prg-improve-row--bordered': i < improvements.length - 1 }">
              <div class="prg-improve-name">{{ imp.skill }}</div>
              <div class="prg-improve-tip">{{ imp.tip }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ACHIEVEMENTS -->
      <div class="prg-achievements-card">
        <div class="prg-ach-head">
          <div class="prg-card-title">{{ t('progress.achievements') }} <span class="prg-star-s">✦</span></div>
        </div>
        <div class="prg-ach-grid">
          <div v-for="(a, i) in achievements" :key="i" class="prg-ach-item" :class="{ 'prg-ach-item--locked': a.locked }">
            <div class="prg-ach-bubble">
              <span v-if="a.locked" class="prg-ach-emoji">🔒</span>
              <template v-else>
                <span class="prg-ach-emoji">{{ a.icon }}</span>
                <span v-if="a.val" class="prg-ach-val">{{ a.val }}</span>
              </template>
            </div>
            <div class="prg-ach-title">{{ a.title }}</div>
            <div class="prg-ach-sub">{{ a.sub }}</div>
          </div>
        </div>
      </div>

      <div class="prg-lumi-wrap">
        <LgLumiFact />
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import LgCircleRing from '../components/linglow/LgCircleRing.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import mapCityImg from '../assets/linglow/map_city.jpg'

const { t } = useI18n()

const loading = ref(true)
const totalWords = ref(0)
const learnedWords = ref(0)
const streakDays = ref(6)

const learnedPct = computed(() => {
  if (!totalWords.value) return 0
  return Math.round((learnedWords.value / totalWords.value) * 100)
})

const skillPct = computed(() => Math.min(learnedPct.value + 15, 99))

onMounted(async () => {
  try {
    const data: any = await apiClient.request('/api/dashboard')
    const s = data?.stats ?? data ?? {}
    totalWords.value = s.total_words ?? 0
    learnedWords.value = s.learned_words ?? 0
  } catch { /* ignore */ } finally {
    loading.value = false
  }
})

const weekDays = computed(() => t('progress.weekDays').split(','))
const weekAct = ['done', 'done', 'done', 'done', 'done', 'today', 'empty']

const metrics = computed(() => [
  { icon: '⏱', value: String(Math.round(learnedWords.value * 3.2)), label: t('progress.metricMinutes'), sub: t('progress.metricMinutesSub') },
  { icon: '🌿', value: String(learnedWords.value),  label: t('progress.metricWords'),   sub: t('progress.metricWordsSub') },
  { icon: '📖', value: String(Math.round(learnedWords.value * 0.06)), label: t('progress.metricTexts'),   sub: t('progress.metricTextsSub') },
  { icon: '💬', value: String(Math.round(learnedWords.value * 0.05)), label: t('progress.metricChats'),   sub: t('progress.metricChatsSub') },
])

const districtItems = [
  { name: 'Plaza Clara',    status: t('progress.distExcellent'),  pct: 88, fill: '#3F6F3F' },
  { name: 'Distrito Alto',  status: t('progress.distGood'),       pct: 62, fill: '#7FAE6A' },
  { name: 'Barrio del Mar', status: t('progress.distInProgress'), pct: 38, fill: '#D9A83F' },
  { name: 'El Mercado',     status: t('progress.distJustStarted'),pct: 14, fill: '#E3D8C6' },
  { name: 'Colina Verde',   status: '',                           pct: 0,  fill: '#E3D8C6', locked: true },
]

const skills = [
  { name: t('progress.grammarLabel'), pct: computed(() => skillPct.value).value, fill: '#3F6F3F' },
  { name: t('progress.wordsLabel'),   pct: computed(() => Math.max(learnedPct.value - 3, 0)).value, fill: '#3F6F3F' },
  { name: t('progress.readingLabel'), pct: computed(() => Math.round(learnedPct.value * 0.73)).value, fill: '#D9A83F' },
  { name: t('progress.chatLabel'),    pct: computed(() => Math.round(learnedPct.value * 0.80)).value, fill: '#7FAE6A' },
]

const improvements = [
  { skill: t('progress.grammarLabel'), tip: t('progress.impGrammar') },
  { skill: t('progress.readingLabel'), tip: t('progress.impReading') },
  { skill: t('progress.wordsLabel'),   tip: t('progress.impWords') },
]

const achievements = computed(() => [
  { icon: '🔥', val: String(streakDays.value), title: t('progress.achStreak'),  sub: t('progress.achStreakSub', { n: streakDays.value }) },
  { icon: '📚', val: String(Math.round(learnedWords.value * 0.06)), title: t('progress.achReader'), sub: t('progress.achReaderSub') },
  { icon: '🌿', val: String(learnedWords.value), title: t('progress.achCollector'), sub: t('progress.achCollectorSub') },
  { icon: '🏙', val: '4',   title: t('progress.achExplorer'), sub: t('progress.achExplorerSub') },
  { icon: '👑', val: '',    title: t('progress.achExpert'),   sub: t('progress.achExpertSub'), locked: true },
])
</script>

<style scoped>
.prg-page {
  padding-bottom: 32px;
}
.prg-header {
  padding: 16px 16px 0;
  display: flex;
  align-items: baseline;
  gap: 4px;
}
.prg-logo {
  font-family: 'Lora', serif;
  font-size: 34px;
  color: var(--text);
  letter-spacing: -0.02em;
  line-height: 1;
}
.prg-star { color: var(--dorado); font-size: 14px; margin-left: 2px; }
.prg-star-s { color: var(--dorado); font-size: 10px; }

/* SUMMARY CARD */
.prg-summary-card {
  margin: 12px 16px;
  position: relative;
  border-radius: 18px;
  overflow: hidden;
  border: 1px solid var(--border);
  box-shadow: var(--shadow-card);
}
.prg-summary-overlay1 {
  position: absolute; inset: 0;
  background: linear-gradient(90deg, rgba(255,249,237,0.98) 0%, rgba(255,244,226,0.90) 100%);
}
:root[data-theme="dark"] .prg-summary-overlay1 {
  background: linear-gradient(90deg, rgba(11,23,17,0.98) 0%, rgba(16,28,21,0.88) 100%);
}
.prg-summary-overlay2 {
  position: absolute; inset: 0;
  background: radial-gradient(circle at 84% 14%, rgba(217,168,63,0.18), transparent 32%);
}
.prg-summary-map {
  position: absolute; right: 0; top: 0;
  width: 54%; height: 52%;
  object-fit: cover; opacity: 0.70;
  mask-image: linear-gradient(90deg, transparent 0%, black 36%);
  -webkit-mask-image: linear-gradient(90deg, transparent 0%, black 36%);
  pointer-events: none;
}
:root[data-theme="dark"] .prg-summary-map { opacity: 0.38; }
.prg-summary-lumi {
  position: absolute; right: 10px; top: 6px; z-index: 3; pointer-events: none;
}
.prg-summary-body {
  position: relative; z-index: 2;
  padding: 16px 16px 14px;
}
.prg-summary-title {
  font-family: 'Lora', serif;
  font-size: 18px; font-weight: 600; color: var(--text);
  line-height: 1.15; margin-bottom: 2px;
}
.prg-summary-sub {
  font-family: 'Inter', sans-serif;
  font-size: 12px; color: var(--subtext); margin-bottom: 14px;
}
.prg-metrics-grid {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 7px;
}
.prg-metric-cell {
  padding: 9px 4px; border-radius: 10px; text-align: center;
  display: flex; flex-direction: column; align-items: center;
  background: rgba(255,249,237,0.74);
  border: 1px solid rgba(129,93,42,0.11);
}
:root[data-theme="dark"] .prg-metric-cell {
  background: rgba(32,53,42,0.42);
  border-color: rgba(255,255,255,0.06);
}
.prg-metric-icon { font-size: 15px; line-height: 1; }
.prg-metric-val { font-family: 'Lora', serif; font-size: 19px; font-weight: 600; color: var(--text); margin-top: 4px; line-height: 1; }
.prg-metric-label { font-family: 'Inter', sans-serif; font-size: 9px; color: var(--subtext); line-height: 1.3; margin-top: 2px; }
.prg-metric-sub { font-family: 'Inter', sans-serif; font-size: 9px; color: var(--subtext); line-height: 1.3; }

/* 2-column rows */
.prg-row2 {
  display: grid; grid-template-columns: 1fr 1fr; gap: 10px;
  padding: 10px 16px 0;
}
.prg-card {
  padding: 14px; border-radius: 16px;
  background: var(--card-bg); border: 1px solid var(--border);
  box-shadow: var(--shadow-card); overflow: hidden;
}
.prg-card--rel { position: relative; min-height: 180px; }
.prg-card-title {
  font-family: 'Lora', serif; font-size: 15px; font-weight: 600;
  color: var(--text); line-height: 1.2;
}
.prg-card-title-spk {
  font-family: 'Lora', serif; font-size: 13px; font-weight: 600;
  color: var(--text); line-height: 1.2;
}
.prg-card-sub {
  font-family: 'Inter', sans-serif; font-size: 11px; color: var(--subtext);
  margin-top: 3px; line-height: 1.35;
}

/* WEEK */
.prg-week-row { display: flex; justify-content: space-between; margin-top: 12px; }
.prg-week-col { display: flex; flex-direction: column; align-items: center; gap: 3px; }
.prg-week-dot {
  width: 22px; height: 22px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  background: var(--surface-3);
}
.prg-week-dot--done { background: #3F6F3F; }
.prg-week-dot--today { background: #FFF0C7; border: 1.5px dashed var(--dorado); }
.prg-week-empty { font-size: 7px; color: var(--subtext); }
.prg-week-label { font-family: 'Inter', sans-serif; font-size: 8px; color: var(--subtext); }
.prg-streak-row { display: flex; align-items: baseline; gap: 3px; margin-top: 12px; }
.prg-streak-num { font-family: 'Lora', serif; font-size: 30px; font-weight: 600; color: var(--text); }
.prg-streak-unit { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); }

/* DISTRICTS */
.prg-districts-list { display: flex; flex-direction: column; gap: 8px; margin-top: 8px; }
.prg-district-row {}
.prg-district-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 3px; }
.prg-district-name { font-family: 'Inter', sans-serif; font-size: 11px; color: var(--text); font-weight: 500; }
.prg-district-status { font-family: 'Inter', sans-serif; font-size: 9.5px; color: var(--subtext); flex-shrink: 0; margin-left: 4px; }
.prg-bar-track { height: 5px; border-radius: 999px; background: var(--progress-track); overflow: hidden; }
.prg-bar-fill { height: 100%; border-radius: 999px; }

/* FAV ZONE */
.prg-fav-name { font-family: 'Lora', serif; font-size: 13px; font-weight: 600; color: var(--text); margin-top: 4px; }
.prg-donut-wrap { margin-top: 12px; }
.prg-donut-inner { display: flex; flex-direction: column; align-items: center; }
.prg-donut-pct { font-family: 'Lora', serif; font-size: 14px; font-weight: 600; color: var(--text); line-height: 1; }
.prg-donut-sub { font-family: 'Inter', sans-serif; font-size: 8px; color: var(--subtext); }
.prg-card-bg-img {
  position: absolute; right: 0; bottom: 0;
  width: 68%; height: 50%; object-fit: cover; opacity: 0.48;
  mask-image: linear-gradient(135deg, transparent 0%, black 42%);
  -webkit-mask-image: linear-gradient(135deg, transparent 0%, black 42%);
  pointer-events: none;
}
:root[data-theme="dark"] .prg-card-bg-img { opacity: 0.24; }

/* STRONGEST SKILL */
.prg-skill-row { display: flex; align-items: center; gap: 9px; margin-top: 12px; }
.prg-skill-icon-box {
  width: 40px; height: 40px; flex-shrink: 0; border-radius: 10px;
  background: rgba(63,111,63,0.08); border: 1px solid rgba(63,111,63,0.14);
  display: flex; align-items: center; justify-content: center;
}
:root[data-theme="dark"] .prg-skill-icon-box { background: rgba(63,111,63,0.18); border-color: rgba(63,111,63,0.3); }
.prg-skill-name { font-family: 'Lora', serif; font-size: 15px; font-weight: 600; color: var(--text); line-height: 1.15; }
.prg-big-num-row { display: flex; align-items: baseline; gap: 1px; margin-top: 14px; }
.prg-big-num { font-family: 'Lora', serif; font-size: 32px; font-weight: 600; color: var(--text); }
.prg-big-pct { font-family: 'Lora', serif; font-size: 20px; font-weight: 600; color: var(--text); }

/* SKILLS LIST */
.prg-skills-list { display: flex; flex-direction: column; gap: 10px; margin-top: 12px; }
.prg-skill-item {}
.prg-skill-item-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 3px; }
.prg-skill-item-name { font-family: 'Inter', sans-serif; font-size: 11px; color: var(--text); }
.prg-skill-item-pct { font-family: 'Inter', sans-serif; font-size: 11px; color: var(--text); font-weight: 600; }

/* IMPROVEMENTS */
.prg-improve-list { margin-top: 9px; }
.prg-improve-row { padding: 9px 0; }
.prg-improve-row:first-child { padding-top: 0; }
.prg-improve-row--bordered { border-bottom: 1px solid var(--border); }
.prg-improve-name { font-family: 'Inter', sans-serif; font-size: 11px; color: var(--text); font-weight: 500; }
.prg-improve-tip { font-family: 'Inter', sans-serif; font-size: 10px; color: var(--subtext); line-height: 1.35; margin-top: 2px; }

/* ACHIEVEMENTS */
.prg-lumi-wrap { margin: 10px 16px 0; }

.prg-achievements-card {
  margin: 10px 16px 0; padding: 16px; border-radius: 16px;
  background: var(--card-bg); border: 1px solid var(--border);
  box-shadow: var(--shadow-card);
}
.prg-ach-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.prg-ach-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 6px; }
.prg-ach-item { text-align: center; }
.prg-ach-item--locked { opacity: 0.52; }
.prg-ach-bubble {
  width: 48px; height: 48px; margin: 0 auto 5px; border-radius: 50%;
  background: var(--surface-2); border: 1px solid var(--border);
  display: flex; align-items: center; justify-content: center; flex-direction: column;
}
.prg-ach-emoji { font-size: 13px; line-height: 1; }
.prg-ach-val { font-family: 'Lora', serif; font-size: 10px; font-weight: 700; color: var(--text); line-height: 1; margin-top: 1px; }
.prg-ach-title { font-family: 'Inter', sans-serif; font-size: 9px; font-weight: 600; color: var(--text); margin-top: 2px; line-height: 1.25; }
.prg-ach-sub { font-family: 'Inter', sans-serif; font-size: 8.5px; color: var(--subtext); line-height: 1.25; margin-top: 1px; }
</style>
