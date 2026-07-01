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
          <LgLumi :size="66" pose="proud" />
        </div>
        <div class="prg-summary-body">
          <div class="prg-summary-title">{{ t('progress.monthTitle', { lang: targetLangDisplay }) }} <span class="prg-star-s">✦</span></div>
          <div class="prg-summary-sub">{{ monthMotivation }}</div>
          <div class="prg-metrics-grid">
            <div v-for="m in metrics" :key="m.label" class="prg-metric-cell">
              <LgActivityIcon v-if="m.type" :type="m.type" status="green" :size="40" />
              <span v-else class="prg-metric-icon">{{ m.icon }}</span>
              <div class="prg-metric-val">{{ m.value }}</div>
              <div class="prg-metric-label">{{ m.label }}</div>
              <div class="prg-metric-sub">{{ m.sub }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ROW: RHYTHM + DISTRICTS -->
      <div class="prg-row2" :class="{ 'prg-row2--single': districtItems.length === 0 }">
        <!-- Rhythm card -->
        <div class="prg-card">
          <div class="prg-card-title">{{ t('progress.rhythmTitle') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-week-row">
            <div v-for="(cell, i) in weekCells" :key="i" class="prg-week-col">
              <div class="prg-week-dot" :class="`prg-week-dot--${cell.status}`">
                <svg v-if="cell.status === 'done'" width="9" height="9" viewBox="0 0 24 24" fill="none">
                  <path d="M5 12L10 17L19 7" stroke="white" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                <span v-else class="prg-week-empty">○</span>
              </div>
              <span class="prg-week-label">{{ cell.label }}</span>
            </div>
          </div>
          <div class="prg-streak-row">
            <span class="prg-streak-num">{{ streakDays }}</span>
            <span class="prg-streak-unit">{{ (t as any)('progress.daysRow', streakDays) }}</span>
            <span class="prg-star-s" style="margin-left:4px">✦</span>
          </div>
          <div class="prg-card-sub" style="margin-top:2px">{{ weekMotivation }}</div>
        </div>

        <!-- Districts card -->
        <div v-if="districtItems.length > 0" class="prg-card">
          <div class="prg-card-title">{{ t('progress.districtsTitle') }}</div>
          <div class="prg-card-sub" style="margin-bottom:12px">{{ t('progress.districtsSub') }}</div>
          <div class="prg-districts-list">
            <div v-for="d in districtItems" :key="d.key" class="prg-district-row" :style="{ opacity: d.locked ? 0.5 : 1 }">
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

      <!-- ROW: STRONGEST SKILL -->
      <div v-if="strongSkill" class="prg-row2 prg-row2--single">
        <div v-if="strongSkill" class="prg-card prg-card--rel">
          <div class="prg-card-title-spk">{{ t('progress.strongSkill') }} <span class="prg-star-s">✦</span></div>
          <div class="prg-skill-row">
            <div class="prg-skill-icon-box">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                <path d="M12 2v4M8 6h8M9 10h6M10.5 14h3" stroke="var(--salvia)" stroke-width="1.8" stroke-linecap="round"/>
                <rect x="6" y="17" width="12" height="3.5" rx="1.8" fill="var(--salvia)" opacity="0.85"/>
              </svg>
            </div>
            <div>
              <div class="prg-skill-name">{{ strongSkill.name }}</div>
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
      <div v-if="skills.length > 0 || improvements.length > 0" class="prg-row2">
        <div v-if="skills.length > 0" class="prg-card">
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

        <div v-if="improvements.length > 0" class="prg-card">
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
      <div v-if="achievements.length > 0" class="prg-achievements-card">
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
        <LgLumiFact context="progress" />
      </div>

      <!-- Legacy detailed progress + activity charts -->
      <div class="prg-charts-wrap">
        <LgProgressCharts />
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { type LinglowStats } from '../api/statsClient'
import { type CourseProgress } from '../api/courseClient'
import { apiClient } from '../api/client'
import { useCourse } from '../composables/useCourse'
import { useLearningConfig } from '../composables/useLearningConfig'
import { useLocale } from '../composables/useLocale'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgProgressCharts from '../components/linglow/LgProgressCharts.vue'
import mapCityImg from '../assets/linglow/art/city-map-826x664.jpg'

const { t } = useI18n()
const { currentCourseCode } = useCourse()
const { targetLangDisplay, ensureLearningLoaded } = useLearningConfig()
const { currentLocale } = useLocale()

const loading = ref(true)
const stats = ref<LinglowStats | null>(null)
const progress = ref<CourseProgress | null>(null)

onMounted(async () => {
  try {
    const code = currentCourseCode.value || undefined
    // Single aggregated round trip instead of separate stats + progress calls.
    const [ov] = await Promise.all([
      apiClient.request<any>(code ? `/api/overview/progress?course_code=${encodeURIComponent(code)}` : '/api/overview/progress'),
      ensureLearningLoaded().catch(() => {}),
    ])
    stats.value = (ov?.stats as LinglowStats) ?? null
    progress.value = (ov?.progress as CourseProgress) ?? null
  } catch { /* ignore */ } finally {
    loading.value = false
  }
})

const streakDays = computed(() => stats.value?.streak.current_days ?? 0)

const monthMotivation = computed(() => {
  const m = stats.value?.month
  if (!m) return ''
  const words = m.words_learned ?? 0
  const texts = m.texts_read ?? 0
  const mins = m.active_minutes ?? 0
  if (words === 0 && texts === 0 && mins === 0) return t('progress.monthStart')
  if (words >= 100 || texts >= 20 || mins >= 300) return t('progress.monthGreat')
  if (words >= 30 || texts >= 5 || mins >= 60) return t('progress.monthGood')
  return t('progress.monthLittle')
})

const weekMotivation = computed(() => {
  const done = weekCells.value.filter(c => c.status === 'done').length
  const s = streakDays.value
  if (s === 0 && done === 0)
    return t('progress.weekNone')
  if (s === 0 && done > 0)
    return t('progress.weekBrokenStreak', done)
  if (s === 1)
    return t('progress.weekStreak1')
  if (s <= 3)
    return t('progress.weekStreakFew', s)
  if (s <= 6)
    return t('progress.weekStreakWeek', s)
  if (s === 7)
    return t('progress.weekStreak7')
  if (s <= 13)
    return t('progress.weekStreakMid', s)
  if (s <= 29)
    return t('progress.weekStreakLong', s)
  return t('progress.weekStreakEpic', s)
})

// Show the current calendar week (Mon–Sun). Future days are always empty,
// today is highlighted, past days reflect real activity from the backend.
const weekCells = computed(() => {
  const locale = currentLocale.value === 'ru' ? 'ru-RU' : 'en-US'
  const toKey = (d: Date) =>
    `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  const labelFor = (d: Date) => {
    const wd = d.toLocaleDateString(locale, { weekday: 'short' })
    return wd.charAt(0).toUpperCase() + wd.slice(1)
  }

  // Map real activity (only "done" matters; today/empty are recomputed locally).
  const statusByDate = new Map<string, string>()
  for (const d of stats.value?.week || []) statusByDate.set(d.date, d.status)

  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const todayKey = toKey(today)
  // Monday of the current week (getDay(): 0=Sun..6=Sat → 0=Mon..6=Sun).
  const mondayOffset = (today.getDay() + 6) % 7
  const monday = new Date(today)
  monday.setDate(today.getDate() - mondayOffset)

  const cells = []
  for (let i = 0; i < 7; i++) {
    const date = new Date(monday)
    date.setDate(monday.getDate() + i)
    const key = toKey(date)
    let status: string
    if (key > todayKey) status = 'empty'                       // future day — never done
    else if (statusByDate.get(key) === 'done') status = 'done' // past/today with activity
    else if (key === todayKey) status = 'today'                // today, not yet active
    else status = 'empty'                                      // past day, no activity
    cells.push({ label: labelFor(date), status })
  }
  return cells
})

const metrics = computed(() => {
  const m = stats.value?.month
  return [
    { icon: '⏱', value: String(m?.active_minutes ?? 0), label: (t as any)('progress.metricMinutes', m?.active_minutes ?? 0), sub: t('progress.metricMinutesSub') },
    { icon: '',  type: 'words' as const,       value: String(m?.words_learned ?? 0),  label: (t as any)('progress.metricWords', m?.words_learned ?? 0),   sub: t('progress.metricWordsSub') },
    { icon: '',  type: 'reading' as const,     value: String(m?.texts_read ?? 0),     label: (t as any)('progress.metricTexts', m?.texts_read ?? 0),   sub: t('progress.metricTextsSub') },
  ]
})

const DISTRICT_FILLS = ['#3F6F3F', '#7FAE6A', '#D9A83F', '#E3D8C6']
// Order districts by the CEFR level at which they unlock (A0 → C2).
const CEFR_ORDER = ['A0', 'A1', 'A2', 'B1', 'B2', 'C1', 'C2']
const cefrRank = (code: string) => {
  const i = CEFR_ORDER.indexOf((code || '').toUpperCase())
  return i === -1 ? CEFR_ORDER.length : i
}
const districtItems = computed(() => {
  const rows = [...(progress.value?.by_district || [])].sort(
    (a, b) => cefrRank(a.level_code) - cefrRank(b.level_code),
  )
  return rows.map((d) => {
    const pct = Math.round(d.progress_percent)
    let status = ''
    let fill = DISTRICT_FILLS[3]
    if (pct >= 75) { status = t('progress.distExcellent'); fill = DISTRICT_FILLS[0] }
    else if (pct >= 40) { status = t('progress.distGood'); fill = DISTRICT_FILLS[1] }
    else if (pct >= 10) { status = t('progress.distInProgress'); fill = DISTRICT_FILLS[2] }
    else if (d.attempted_items > 0) { status = t('progress.distJustStarted') }
    return { key: d.district_code || String(d.district_id), name: d.title, status, pct, fill, locked: d.attempted_items === 0 && pct === 0 }
  })
})


const MODE_LABELS: Record<string, string> = {
  word_training: 'progress.wordsLabel',
  grammar_training: 'progress.grammarLabel',
  grammar_test: 'progress.grammarLabel',
  reading_completion: 'progress.readingLabel',
  chat: 'progress.chatLabel',
  speaking: 'progress.speakingLabel',
}
const SKILL_FILLS = ['#3F6F3F', '#7FAE6A', '#D9A83F', '#5B9ED4']

// Merge backend modes sharing one label (grammar test + training) into one skill row
const skills = computed(() => {
  const byLabel = new Map<string, { attempts: number; correct: number }>()
  for (const s of stats.value?.skills || []) {
    const key = MODE_LABELS[s.mode]
    if (!key) continue
    const cur = byLabel.get(key) || { attempts: 0, correct: 0 }
    cur.attempts += s.attempt_count
    cur.correct += s.correct_count
    byLabel.set(key, cur)
  }
  return [...byLabel.entries()].map(([key, v], i) => ({
    name: t(key),
    pct: v.attempts > 0 ? Math.round((v.correct / v.attempts) * 100) : 0,
    attempts: v.attempts,
    fill: SKILL_FILLS[i % SKILL_FILLS.length],
  })).sort((a, b) => b.pct - a.pct)
})

const strongSkill = computed(() => {
  const candidates = skills.value.filter(s => s.attempts >= 20)
  return candidates.length > 0 ? candidates[0] : null
})
const skillPct = computed(() => strongSkill.value?.pct ?? 0)

const ACH_META: Record<string, { icon: string; title: string; sub: string }> = {
  streak:    { icon: '🔥', title: 'progress.achStreak',    sub: 'progress.achStreakSub' },
  reader:    { icon: '📚', title: 'progress.achReader',    sub: 'progress.achReaderSub' },
  collector: { icon: '🌿', title: 'progress.achCollector', sub: 'progress.achCollectorSub' },
  explorer:  { icon: '🏙', title: 'progress.achExplorer',  sub: 'progress.achExplorerSub' },
  expert:    { icon: '👑', title: 'progress.achExpert',    sub: 'progress.achExpertSub' },
}
const achievements = computed(() => {
  return (stats.value?.achievements || []).map((a) => {
    const meta = ACH_META[a.code]
    if (!meta) return null
    return {
      icon: meta.icon,
      val: a.unlocked && a.value > 0 ? String(a.value) : '',
      title: t(meta.title),
      sub: (t as any)(meta.sub, a.value, { n: a.value }),
      locked: !a.unlocked,
    }
  }).filter((a): a is NonNullable<typeof a> => a !== null)
})

const IMP_TIPS: Record<string, { skill: string; tip: string }> = {
  mode_accuracy: { skill: '', tip: 'progress.impModeAccuracy' },
  due_backlog:   { skill: 'progress.wordsLabel', tip: 'progress.impDueBacklog' },
  weak_district: { skill: '', tip: 'progress.impWeakDistrict' },
  no_reading:    { skill: 'progress.readingLabel', tip: 'progress.impNoReading' },
  no_chat:       { skill: 'progress.chatLabel', tip: 'progress.impNoChat' },
}
const improvements = computed(() => {
  return (stats.value?.improvements || []).map((imp) => {
    const meta = IMP_TIPS[imp.kind]
    if (!meta) return null
    let skill = meta.skill ? t(meta.skill) : ''
    if (imp.kind === 'mode_accuracy' && imp.mode) skill = t(MODE_LABELS[imp.mode] || 'progress.wordsLabel')
    if (imp.kind === 'weak_district') skill = imp.title || ''
    return {
      skill,
      tip: (t as any)(meta.tip, imp.count ?? 0, { n: imp.count ?? 0, pct: Math.round(imp.accuracy ?? 0), name: imp.title ?? '' }),
    }
  }).filter((i): i is NonNullable<typeof i> => i !== null)
})
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
  border: 1px solid var(--border);
  box-shadow: var(--shadow-card);
  /* no overflow:hidden so Lumi can peek out above */
}
.prg-summary-overlay1 {
  position: absolute; inset: 0;
  border-radius: 18px;
  background: linear-gradient(90deg, rgba(255,249,237,0.98) 0%, rgba(255,244,226,0.90) 100%);
}
:root[data-theme="dark"] .prg-summary-overlay1 {
  background: linear-gradient(90deg, rgba(11,23,17,0.98) 0%, rgba(16,28,21,0.88) 100%);
}
.prg-summary-overlay2 {
  position: absolute; inset: 0;
  border-radius: 18px;
  background: radial-gradient(circle at 84% 14%, rgba(217,168,63,0.18), transparent 32%);
}
.prg-summary-map {
  position: absolute; right: 0; top: 0;
  width: 54%; height: 100%;
  object-fit: cover; opacity: 0.70;
  border-radius: 0 18px 18px 0;
  mask-image: linear-gradient(90deg, transparent 0%, black 36%);
  -webkit-mask-image: linear-gradient(90deg, transparent 0%, black 36%);
  pointer-events: none;
}
:root[data-theme="dark"] .prg-summary-map { opacity: 0.38; }
.prg-summary-lumi {
  position: absolute; right: 6px; top: -18px; z-index: 3; pointer-events: none;
}
.prg-summary-body {
  position: relative; z-index: 2;
  padding: 16px 16px 14px;
}
.prg-summary-title {
  font-family: 'Lora', serif;
  font-size: 18px; font-weight: 600; color: var(--text);
  line-height: 1.15; margin-bottom: 2px;
  /* Keep clear of the Lumi firefly that peeks in at the top-right corner. */
  padding-right: 64px;
}
.prg-summary-sub {
  font-family: 'Inter', sans-serif;
  font-size: 12px; color: var(--subtext); margin-bottom: 14px;
  padding-right: 64px;
}
@media (max-width: 360px) {
  .prg-summary-title { font-size: 16px; }
  .prg-summary-sub { font-size: 11px; }
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
.prg-row2--single { grid-template-columns: 1fr; }
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
.prg-week-row { display: flex; justify-content: space-between; gap: 2px; margin-top: 12px; }
.prg-week-col { display: flex; flex-direction: column; align-items: center; gap: 3px; min-width: 0; }
.prg-week-dot {
  width: 22px; height: 22px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
  background: var(--surface-3);
}
/* Shrink the week dots on narrow screens so the 7 days don't overflow the card. */
@media (max-width: 400px) {
  .prg-week-dot { width: 18px; height: 18px; }
  .prg-week-label { font-size: 7.5px; }
}
@media (max-width: 340px) {
  .prg-week-dot { width: 15px; height: 15px; }
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
.prg-charts-wrap { margin: 16px 16px 0; }

.prg-achievements-card {
  margin: 10px 16px 0; padding: 16px; border-radius: 16px;
  background: var(--card-bg); border: 1px solid var(--border);
  box-shadow: var(--shadow-card);
}
.prg-ach-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.prg-ach-grid { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 6px; }
.prg-ach-item { text-align: center; min-width: 0; }
.prg-ach-item--locked { opacity: 0.52; }
.prg-ach-bubble {
  width: 48px; height: 48px; margin: 0 auto 5px; border-radius: 50%;
  background: var(--surface-2); border: 1px solid var(--border);
  display: flex; align-items: center; justify-content: center; flex-direction: column;
}
.prg-ach-emoji { font-size: 13px; line-height: 1; }
.prg-ach-val { font-family: 'Lora', serif; font-size: 10px; font-weight: 700; color: var(--text); line-height: 1; margin-top: 1px; }
.prg-ach-title { font-family: 'Inter', sans-serif; font-size: 9px; font-weight: 600; color: var(--text); margin-top: 2px; line-height: 1.25; overflow-wrap: anywhere; }
.prg-ach-sub { font-family: 'Inter', sans-serif; font-size: 8.5px; color: var(--subtext); line-height: 1.25; margin-top: 1px; overflow-wrap: anywhere; }
@media (max-width: 380px) {
  .prg-achievements-card { padding: 14px 12px; }
  .prg-ach-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px 6px; }
  .prg-ach-bubble { width: 42px; height: 42px; }
}
</style>
