<template>
  <div class="dst-page">

    <!-- HEADER -->
    <LgPageHeader
      :title="district?.title || t('city.district')"
      :show-back="true"
      @back="router.push('/city')"
    />

    <!-- SUBTITLE -->
    <div class="dst-sub-row">
      <p class="dst-desc">{{ districtDesc }}</p>
    </div>

    <div v-if="loading" class="dst-loading">{{ t('common.loading') }}</div>
    <template v-else>

      <!-- 4 ACTIVITY AREAS -->
      <div class="dst-areas-card">
        <div v-for="(a, i) in areas" :key="i" class="dst-area-row" :class="{ 'dst-area-row--bordered': i > 0 }">
          <div class="dst-area-icon">
            <LgActivityIcon :type="a.type" :status="a.status" :size="22" />
          </div>
          <div class="dst-area-body">
            <div class="dst-area-label">{{ a.label }}</div>
            <div class="dst-area-meta">{{ a.meta }}</div>
            <div v-if="a.pct !== null" class="dst-bar-track">
              <div class="dst-bar-fill" :style="{ width: a.pct + '%', background: a.color }" />
            </div>
          </div>
          <button class="dst-area-cta" type="button" @click="a.action">
            {{ a.cta }}
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        </div>
      </div>

      <!-- LUMI FACT -->
      <div class="dst-lumi-wrap">
        <LgLumiFact context="district" />
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { courseClient, type CourseMap, type CourseProgress } from '../api/courseClient'
import { useLocale } from '../composables/useLocale'
import { grammarClient } from '../api/grammarClient'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const courseMap = ref<CourseMap | null>(null)
const progress = ref<CourseProgress | null>(null)
const grammarCategories = ref<any[]>([])
const loading = ref(true)
const courseCode = computed(() => (typeof route.query.course_code === 'string' ? route.query.course_code : undefined))
const districtCode = computed(() => String(route.params.districtCode || ''))

const district = computed(() => {
  const dists = Array.isArray(courseMap.value?.districts) ? courseMap.value!.districts : []
  return dists.find(d => d.code === districtCode.value) || null
})

const districtSignal = computed(() => {
  return (progress.value?.by_district || []).find(d => d.district_code === districtCode.value) || {
    foundation: 0, confidence: 0, stability: 0, weakness: 0,
    progress_percent: 0, attempted_items: 0, total_items: 0, mastered_items: 0,
  }
})

// Per-type progress for this district's locations
const typeProgress = computed(() => {
  const locs = (progress.value?.by_location || []).filter(l => l.district_code === districtCode.value)
  const out: Record<string, { attempted: number; total: number; pct: number }> = {}
  for (const loc of locs) {
    const key = loc.location_type || 'other'
    if (!out[key]) out[key] = { attempted: 0, total: 0, pct: 0 }
    out[key].attempted += loc.attempted_items
    out[key].total += loc.total_items
  }
  for (const key of Object.keys(out)) {
    const e = out[key]
    e.pct = e.total > 0 ? Math.round((e.attempted / e.total) * 100) : 0
  }
  return out
})

const wordsPct = computed(() => Math.round(districtSignal.value.confidence))
const grammarPct = computed(() => {
  const lv = districtLevelCode.value
  if (!lv || !grammarCategories.value.length) return 0
  const cats = grammarCategories.value.filter(c => String(c.level || '').toUpperCase() === lv)
  if (!cats.length) return 0
  const total = cats.reduce((s: number, c: any) => s + (c.total_chapters || 0), 0)
  const passed = cats.reduce((s: number, c: any) => s + (c.passed_chapters || 0), 0)
  return total > 0 ? Math.round((passed / total) * 100) : 0
})
const readingPct = computed(() => {
  const r = typeProgress.value['reading_text'] || typeProgress.value['reading']
  return r ? r.pct : 0
})
// Level code for the district (e.g. "A0", "A1") — used to filter grammar categories
const districtLevelCode = computed(() => district.value?.level_code?.toUpperCase() || '')
const districtLevelQuery = computed(() => districtLevelCode.value ? { level: districtLevelCode.value } : {})

const { currentLocale } = useLocale()
const districtDesc = computed(() => {
  const d = district.value
  if (!d) return ''
  const i18nDesc = d.metadata?.desc_i18n?.[currentLocale.value]
  return i18nDesc || d.description || d.title || ''
})

const areas = computed(() => {
  function pctToStatus(pct: number): 'gray' | 'orange' | 'yellow' | 'green' {
    if (pct <= 0) return 'gray'
    if (pct < 34) return 'orange'
    if (pct < 67) return 'yellow'
    return 'green'
  }
  return [
    {
      type: 'grammar' as const,
      status: pctToStatus(grammarPct.value),
      label: t('city.areaGrammar'), color: '#2d6b3a',
      meta: t('city.areaMetaGrammar', { pct: grammarPct.value }),
      pct: grammarPct.value, cta: t('city.ctaContinue'),
      action: () => router.push({ name: 'LearningGrammar', query: districtLevelQuery.value }),
    },
    {
      type: 'words' as const,
      status: pctToStatus(wordsPct.value),
      label: t('city.areaWords'), color: '#2d6b3a',
      meta: t('city.areaMetaWords', { pct: wordsPct.value }),
      pct: wordsPct.value, cta: t('city.ctaContinue'),
      action: () => router.push({ name: 'WordSets', query: districtLevelQuery.value }),
    },
    {
      type: 'reading' as const,
      status: pctToStatus(readingPct.value),
      label: t('city.areaReading'), color: '#c8a84b',
      meta: t('city.areaMetaReading', { pct: readingPct.value }),
      pct: readingPct.value, cta: t('city.ctaRead'),
      action: () => router.push({ name: 'ReadingCategories', query: districtLevelQuery.value }),
    },
  ]
})

onMounted(async () => {
  try {
    const [map, prog, grammarData] = await Promise.all([
      courseClient.getCourseMap(courseCode.value),
      courseClient.getProgress(courseCode.value),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    courseMap.value = map
    progress.value = prog
    grammarCategories.value = grammarData.categories || []
  } catch { /* ignore */ } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dst-page { padding-bottom: 32px; }

/* SUB ROW */
.dst-sub-row {
  padding: 0 16px 14px;
  text-align: center;
}
.dst-desc {
  margin: 0 0 10px;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  color: var(--subtext);
  line-height: 1.5;
}
.dst-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 14px;
  border-radius: 20px;
  background: rgba(45, 107, 58, 0.1);
  border: 1px solid rgba(45, 107, 58, 0.25);
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 600;
  color: #2d6b3a;
}
:root[data-theme="dark"] .dst-chip {
  background: rgba(45, 107, 58, 0.22);
  color: #7fd896;
  border-color: rgba(45, 107, 58, 0.4);
}

.dst-loading {
  padding: 32px 16px;
  text-align: center;
  color: var(--subtext);
  font-size: 14px;
}

/* ILLUSTRATION */
.dst-illustration {
  margin: 0 16px;
  border-radius: 18px;
  overflow: hidden;
  position: relative;
  height: 220px;
  background: var(--chip-bg);
  border: 1px solid var(--border);
}
.dst-illus-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.dst-bldg-label {
  position: absolute;
  transform: translate(-50%, -50%);
  background: rgba(255, 251, 240, 0.92);
  border: 1px solid rgba(200, 168, 75, 0.4);
  border-radius: 10px;
  padding: 5px 9px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  max-width: 110px;
  pointer-events: none;
}
:root[data-theme="dark"] .dst-bldg-label {
  background: rgba(20, 32, 24, 0.92);
  border-color: rgba(200, 168, 75, 0.3);
}
.dst-bldg-name {
  font-family: 'Inter', sans-serif;
  font-weight: 700;
  font-size: 9.5px;
  color: var(--text);
  line-height: 1.2;
}
.dst-bldg-sub {
  font-family: 'Inter', sans-serif;
  font-size: 9px;
  color: var(--subtext);
  margin-top: 1px;
}

/* AREAS CARD */
.dst-areas-card {
  margin: 14px 16px 0;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  overflow: hidden;
  box-shadow: var(--shadow-card);
}
.dst-area-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.dst-area-row--bordered { border-top: 1px solid var(--border); }
.dst-area-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  flex-shrink: 0;
  background: var(--chip-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}
.dst-area-body { flex: 1; min-width: 0; }
.dst-area-label {
  font-family: 'Inter', sans-serif;
  font-weight: 700;
  font-size: 14px;
  color: var(--text);
  margin-bottom: 1px;
}
.dst-area-meta {
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  color: var(--subtext);
}
.dst-bar-track {
  margin-top: 5px;
  height: 3px;
  border-radius: 999px;
  background: var(--progress-track, rgba(0,0,0,0.08));
  overflow: hidden;
}
.dst-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease;
}
.dst-area-cta {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 7px 14px;
  border-radius: 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
  font-family: 'Inter', sans-serif;
  font-weight: 600;
  font-size: 13px;
  color: var(--salvia);
}
.dst-area-cta:hover { opacity: 0.8; }

/* DAILY TASKS */
.dst-tasks-card {
  margin: 0 16px;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  overflow: hidden;
  box-shadow: var(--shadow-card);
}
.dst-task-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
}
.dst-task-row--bordered { border-top: 1px solid var(--border); }
.dst-task-check {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
  border: 1.5px solid var(--border);
  background: var(--surface-2);
  display: flex;
  align-items: center;
  justify-content: center;
}
.dst-task-check--done { background: #3F6F3F; border-color: #3F6F3F; }
.dst-task-text {
  flex: 1;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  color: var(--text);
}
.dst-task-text--done { color: var(--subtext); text-decoration: line-through; }
.dst-task-count {
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 600;
  color: var(--subtext);
}

/* SECTION TITLE */
.dst-section-title {
  margin: 18px 16px 10px;
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 18px;
  color: var(--text);
}

/* DISCOVERY CARD */
.dst-discovery-card {
  width: 100%;
  margin: 0 0 0;
  padding: 0 16px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  display: flex;
  align-items: center;
  gap: 12px;
}
.dst-discovery-thumb {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--chip-bg);
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid var(--border);
}
.dst-discovery-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.dst-discovery-body { flex: 1; min-width: 0; }
.dst-discovery-kicker {
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  color: var(--subtext);
  margin-bottom: 2px;
}
.dst-discovery-title {
  font-family: 'Lora', serif;
  font-weight: 600;
  font-size: 15px;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dst-discovery-desc {
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  color: var(--subtext);
  margin-top: 1px;
}

/* LUMI */
.dst-lumi-wrap { margin: 14px 16px 0; }
</style>
