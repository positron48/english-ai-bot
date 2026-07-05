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

    <LgLoader v-if="loadingInitial" />
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
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { courseClient, type CourseMap, type CourseProgress } from '../api/courseClient'
import { apiClient } from '../api/client'
import { getCachedScreen } from '../api/appDataCache'
import { useCachedOverviewScreen } from '../composables/useCachedOverviewScreen'
import { useLocale } from '../composables/useLocale'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { useAuth } from '../composables/useAuth'
import { metricForLevel, metricPercentToStatus } from '../utils/masteryDisplay'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()
const { isAuthenticated } = useAuth()
const { currentLocale } = useLocale()
const conversationPro = ref(false)
const picturePro = ref(false)

const courseMap = ref<CourseMap | null>(null)
const progress = ref<CourseProgress | null>(null)

function applyCitySnapshot(ov: any) {
  if (ov?.course_map) courseMap.value = ov.course_map
  if (ov?.progress) progress.value = ov.progress
}

const { loadingInitial, load: loadDistrictCache } = useCachedOverviewScreen<any>({
  screenKey: 'city',
  courseCode: resolvedCourseCode,
  locale: currentLocale,
  fetcher: async () => {
    const code = resolvedCourseCode.value
    return apiClient.request(
      code ? `/api/overview/city?course_code=${encodeURIComponent(code)}` : '/api/overview/city',
    )
  },
  applyPayload: (ov) => applyCitySnapshot(ov),
})
const courseCode = computed(() => (typeof route.query.course_code === 'string' ? route.query.course_code : undefined))
const resolvedCourseCode = computed(() => courseCode.value || currentCourseCode.value)
const districtCode = computed(() => String(route.params.districtCode || ''))

const district = computed(() => {
  const dists = Array.isArray(courseMap.value?.districts) ? courseMap.value!.districts : []
  return dists.find(d => d.code === districtCode.value) || null
})

// Level code for the district (e.g. "A0", "A1")
const districtLevelCode = computed(() => district.value?.level_code?.toUpperCase() || '')
const districtLevelQuery = computed(() => districtLevelCode.value ? { level: districtLevelCode.value } : {})

const metricPct = (key: string) => Math.round(metricForLevel(progress.value?.mastery, districtLevelCode.value, key)?.percent || 0)

const districtDesc = computed(() => {
  const d = district.value
  if (!d) return ''
  const i18nDesc = d.metadata?.desc_i18n?.[currentLocale.value]
  return i18nDesc || d.description || d.title || ''
})

interface AreaItem {
  type: 'grammar' | 'words' | 'reading' | 'conversation'
  status: 'gray' | 'orange' | 'yellow' | 'green'
  label: string
  color: string
  meta: string
  pct: number | null
  cta: string
  action: () => void
}

const areas = computed(() => {
  const list: AreaItem[] = [
    {
      type: 'grammar' as const,
      status: metricPercentToStatus(metricPct('grammar')),
      label: t('city.areaGrammar'), color: '#2d6b3a',
      meta: t('city.areaMetaGrammar', { pct: metricPct('grammar') }),
      pct: metricPct('grammar'), cta: t('city.ctaContinue'),
      action: () => router.push({ name: 'LearningGrammar', query: districtLevelQuery.value }),
    },
    {
      type: 'words' as const,
      status: metricPercentToStatus(metricPct('words')),
      label: t('city.areaWords'), color: '#2d6b3a',
      meta: t('city.areaMetaWords', { pct: metricPct('words') }),
      pct: metricPct('words'), cta: t('city.ctaContinue'),
      action: () => router.push({ name: 'WordSets', query: districtLevelQuery.value }),
    },
    {
      type: 'reading' as const,
      status: metricPercentToStatus(metricPct('reading')),
      label: t('city.areaReading'), color: '#c8a84b',
      meta: t('city.areaMetaReading', { pct: metricPct('reading') }),
      pct: metricPct('reading'), cta: t('city.ctaRead'),
      action: () => router.push({ name: 'ReadingCategories', query: districtLevelQuery.value }),
    },
  ]
  const convMetric = metricForLevel(progress.value?.mastery, districtLevelCode.value, 'conversation')
  if (conversationPro.value && convMetric?.included) {
    list.push({
      type: 'conversation' as const,
      status: metricPercentToStatus(metricPct('conversation')),
      label: t('city.areaChat'), color: '#2d6b3a',
      meta: t('city.areaMetaChat', { pct: metricPct('conversation') }),
      pct: metricPct('conversation'), cta: t('city.ctaPractice'),
      action: () => router.push({ name: 'PlaceChatList', params: { districtCode: districtCode.value } }),
    })
  }
  const picMetric = metricForLevel(progress.value?.mastery, districtLevelCode.value, 'picture')
  if (picturePro.value && picMetric?.included) {
    list.push({
      type: 'conversation' as const,
      status: metricPercentToStatus(metricPct('picture')),
      label: t('city.areaPicture'), color: '#2d6b3a',
      meta: t('city.areaMetaPicturePct', { pct: metricPct('picture') }),
      pct: metricPct('picture'), cta: t('city.ctaOpen'),
      action: () => router.push({ name: 'PictureQuestDistrict', params: { districtCode: districtCode.value } }),
    })
  }
  return list
})

onMounted(async () => {
  try {
    await ensureCourseLoaded()
    ensureMe().then(() => {
      conversationPro.value = hasFeature('conversation')
      picturePro.value = hasFeature('picture_description')
    })
    const code = courseCode.value || currentCourseCode.value
    if (code) {
      const cached = await getCachedScreen('city', code, undefined, currentLocale.value)
      if (cached?.payload) applyCitySnapshot(cached.payload)
    }
    await loadDistrictCache()
  } catch { /* ignore */ }
})

watch(currentCourseCode, () => {
  if (isAuthenticated.value) void loadDistrictCache(true)
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
