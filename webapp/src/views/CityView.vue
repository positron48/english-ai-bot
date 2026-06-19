<template>
  <div class="city-view lg-page">
    <!-- Map hero with floating title (prototype MapScreen; static map fallback) -->
    <div class="city-hero">
      <LgCityMap
        class="city-hero-map"
        :map-src="mapCityImg"
        :districts="safeCourseMap?.districts || []"
        :progress-by-code="districtProgressByCode"
        @select="openDistrict"
      >
        <div class="city-hero-overlay">
          <div class="city-hero-title">{{ courseMap?.course.city_name || courseMap?.course.title || t('city.title') }}</div>
          <div class="city-hero-sub">{{ t('city.kicker') }}</div>
        </div>
      </LgCityMap>
    </div>

    <div v-if="loading" class="lg-loading">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="lg-error">
      <strong>{{ t('common.error') }}</strong>
      <p style="margin: 6px 0 0">{{ error }}</p>
      <button type="button" class="city-retry" :disabled="loading" @click="loadCity">{{ t('common.retry') }}</button>
    </div>

    <template v-else-if="safeCourseMap">
      <!-- Overview metrics -->
      <section class="city-overview">
        <div class="lg-card overview-metric">
          <span>{{ formatPercent(progress?.summary.progress_percent || 0) }}</span>
          <label>{{ t('city.progress') }}</label>
        </div>
        <div class="lg-card overview-metric">
          <span>{{ progress?.summary.due_review_count || reviewQueue?.summary.due_count || 0 }}</span>
          <label>{{ t('city.reviewPressure') }}</label>
        </div>
        <div class="lg-card overview-metric">
          <span>{{ formatPercent(progress?.summary.accuracy_percent || 0) }}</span>
          <label>{{ t('city.accuracy') }}</label>
        </div>
      </section>

      <!-- Review station -->
      <section class="city-work-grid">
        <article class="lg-card work-panel">
          <div class="panel-head">
            <h2>{{ t('city.reviewStation') }}</h2>
            <span>{{ reviewQueue?.summary.due_count || 0 }}</span>
          </div>
          <div class="route-list">
            <RouterLink v-for="item in reviewItems" :key="`review:${item.learning_item_id}`" class="route-item route-link" :to="routeForLinglowItem(item)">
              <span class="item-mode">{{ formatType(item.state || item.mode) }}</span>
              <strong>{{ item.title || formatType(item.type) }}</strong>
              <small>{{ item.due_at ? formatDate(item.due_at) : item.location_title }}</small>
            </RouterLink>
            <div v-if="reviewItems.length === 0" class="empty-line">{{ t('city.noReviewItems') }}</div>
          </div>
        </article>
      </section>

      <!-- Districts -->
      <section class="district-list">
        <RouterLink
          v-for="(district, di) in safeCourseMap.districts"
          :key="district.id"
          class="lg-card district-card"
          :to="{ name: 'CityDistrict', params: { districtCode: district.code }, query: selectedCourseCode ? { course_code: selectedCourseCode } : undefined }"
        >
          <img class="district-card-art" :src="districtArt(district, di)" alt="" />
          <div class="district-card-body">
            <div class="district-card-head">
              <LgChip :label="district.level_code" active />
              <h2>{{ district.title }}</h2>
            </div>
            <p class="district-card-locations">{{ safeLocations(district).length }} {{ t('city.locationsShort') }}</p>
            <div class="district-card-signals">
              <div class="district-signal">
                <span>{{ t('city.foundation') }}</span>
                <LgProgressBar :pct="districtSignal(district.code).foundation" :h="4" />
              </div>
              <div class="district-signal">
                <span>{{ t('city.confidence') }}</span>
                <LgProgressBar :pct="districtSignal(district.code).confidence" :h="4" color="var(--hoja)" />
              </div>
              <div class="district-signal">
                <span>{{ t('city.stability') }}</span>
                <LgProgressBar :pct="districtSignal(district.code).stability" :h="4" color="var(--dorado)" />
              </div>
            </div>
          </div>
          <LgIcon name="chevron-right" :s="16" c="var(--text-muted)" />
        </RouterLink>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { courseClient, CourseMap, CourseMapLocation, CourseProgress, ReviewQueue } from '../api/courseClient'
import { useCourse } from '../composables/useCourse'
import { routeForLinglowItem } from '../utils/linglowNavigation'
import LgChip from '../components/linglow/LgChip.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgProgressBar from '../components/linglow/LgProgressBar.vue'
import LgCityMap from '../components/linglow/LgCityMap.vue'
import mapCityImg from '../assets/linglow/art/city-map-826x664.jpg'
import distViajes from '../assets/linglow/art/district-travel-124x90.jpg'
import distCafeterias from '../assets/linglow/art/district-cafes-124x90.jpg'
import distMercados from '../assets/linglow/art/district-markets-124x90.jpg'
import distParques from '../assets/linglow/art/district-parks-124x90.jpg'
import distVida from '../assets/linglow/art/district-life-124x90.jpg'

// District art comes from district metadata; rotation is the legacy fallback
const DISTRICT_ART = [distViajes, distCafeterias, distMercados, distParques, distVida]
const DISTRICT_ART_BY_KEY: Record<string, string> = {
  dist_viajes: distViajes,
  dist_cafeterias: distCafeterias,
  dist_mercados: distMercados,
  dist_parques: distParques,
  dist_vida: distVida,
}
const districtArt = (district: { metadata?: { image?: string } }, index: number) => {
  const key = district.metadata?.image
  if (key && DISTRICT_ART_BY_KEY[key]) return DISTRICT_ART_BY_KEY[key]
  return DISTRICT_ART[index % DISTRICT_ART.length]
}

const { t } = useI18n()
const router = useRouter()
const { courses, currentCourseCode, ensureCourseLoaded, selectCourse: setCurrentCourse } = useCourse()

const openDistrict = (districtCode: string) => {
  router.push({
    name: 'CityDistrict',
    params: { districtCode },
    query: selectedCourseCode.value ? { course_code: selectedCourseCode.value } : undefined,
  })
}

const courseMap = ref<CourseMap | null>(null)
const reviewQueue = ref<ReviewQueue | null>(null)
const progress = ref<CourseProgress | null>(null)
const selectedCourseCode = ref('')
const loading = ref(false)
const selectingCourse = ref(false)
const error = ref('')

// keep local select in sync with global course state
watch(currentCourseCode, (code) => {
  if (code && selectedCourseCode.value !== code) {
    selectedCourseCode.value = code
    loadCity()
  }
})

const itemTypes = computed(() => {
  const totals = safeCourseMap.value?.totals.by_type || {}
  return Object.entries(totals).sort((left, right) => right[1] - left[1])
})

const safeCourseMap = computed(() => {
  if (!courseMap.value) return null
  return {
    ...courseMap.value,
    districts: Array.isArray(courseMap.value.districts) ? courseMap.value.districts : [],
    totals: {
      districts: courseMap.value.totals?.districts || 0,
      locations: courseMap.value.totals?.locations || 0,
      modules: courseMap.value.totals?.modules || 0,
      items: courseMap.value.totals?.items || 0,
      by_type: courseMap.value.totals?.by_type || {},
    },
  }
})

const reviewItems = computed(() => (reviewQueue.value?.items || []).slice(0, 6))

const districtProgressByCode = computed(() => {
  const out: Record<string, NonNullable<CourseProgress['by_district']>[number]> = {}
  for (const row of progress.value?.by_district || []) {
    out[row.district_code] = row
  }
  return out
})

const locationProgressByCode = computed(() => {
  const out: Record<string, NonNullable<CourseProgress['by_location']>[number]> = {}
  for (const row of progress.value?.by_location || []) {
    out[row.location_code] = row
  }
  return out
})

function formatType(type: string): string {
  return type.replace(/_/g, ' ')
}

function formatPercent(value: number): string {
  return `${Math.round(value)}%`
}

function emptySignal() {
  return {
    foundation: 0,
    confidence: 0,
    stability: 0,
    weakness: 0,
    due_review_count: 0,
  }
}

function districtSignal(code: string) {
  return districtProgressByCode.value[code] || emptySignal()
}

function locationSignal(code: string) {
  return locationProgressByCode.value[code] || emptySignal()
}

function formatDate(value: string): string {
  try {
    return new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' }).format(new Date(value))
  } catch {
    return value
  }
}

function locationTitle(type: string, fallback: string): string {
  const key = `city.locationTypes.${type}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

function visibleModules(location: CourseMapLocation) {
  return safeModules(location).slice(0, 4)
}

function safeLocations(district: CourseMap['districts'][number]) {
  return Array.isArray(district.locations) ? district.locations : []
}

function safeModules(location: CourseMapLocation) {
  return Array.isArray(location.modules) ? location.modules : []
}

function safeItems(module: CourseMapLocation['modules'][number]) {
  return Array.isArray(module.items) ? module.items : []
}

async function loadCity() {
  loading.value = true
  error.value = ''
  try {
    const courseCode = selectedCourseCode.value || undefined
    const [map, review, progressData] = await Promise.all([
      courseClient.getCourseMap(courseCode),
      courseClient.getReviewQueue(8, courseCode),
      courseClient.getProgress(courseCode),
    ])
    courseMap.value = map
    reviewQueue.value = review
    progress.value = progressData
    selectedCourseCode.value = map.course.code
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  } finally {
    loading.value = false
  }
}

async function selectCourse() {
  if (!selectedCourseCode.value) return
  selectingCourse.value = true
  error.value = ''
  try {
    await setCurrentCourse(selectedCourseCode.value)
    await loadCity()
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  } finally {
    selectingCourse.value = false
  }
}

onMounted(async () => {
  await ensureCourseLoaded()
  if (!selectedCourseCode.value && currentCourseCode.value) {
    selectedCourseCode.value = currentCourseCode.value
  }
  await loadCity()
})
</script>

<style scoped>
.city-view { padding-top: 16px; display: flex; flex-direction: column; gap: 14px; }

.city-hero {
  position: relative;
  border-radius: 22px;
  overflow: hidden;
  border: 1px solid var(--border);
  box-shadow: var(--shadow-card);
  min-height: 200px;
}
.city-hero-map {
  position: absolute;
  inset: 0;
}
.city-hero-overlay {
  position: relative;
  z-index: 2;
  padding: 18px 20px 110px;
  background: linear-gradient(to bottom, rgba(0,0,0,0.30) 0%, transparent 100%);
  text-align: center;
}
.city-hero-title {
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 22px;
  color: #fff;
  text-shadow: 0 1px 8px rgba(0,0,0,0.5);
  line-height: 1;
}
.city-hero-sub {
  font-size: 11px;
  color: rgba(255,255,255,0.78);
  text-shadow: 0 1px 4px rgba(0,0,0,0.4);
  margin-top: 4px;
  letter-spacing: 0.1px;
}

.city-retry {
  margin-top: 10px;
  padding: 8px 16px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--card-bg);
  color: var(--text);
  font-weight: 600;
  cursor: pointer;
}

.city-overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}
.overview-metric { text-align: center; padding: 14px 8px; }
.overview-metric span {
  display: block;
  font-family: 'Lora', serif;
  font-size: 22px;
  font-weight: 600;
  color: var(--text);
}
.overview-metric label { color: var(--subtext); font-size: 12px; }

.city-work-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;
}
.work-panel { min-height: 180px; }
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}
.panel-head h2 {
  margin: 0;
  font-family: 'Lora', serif;
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}
.panel-head span { color: var(--subtext); font-weight: 700; }
.panel-action {
  color: var(--salvia);
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
}
.route-list { display: grid; }
.route-item {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 2px 10px;
  padding: 9px 0;
  border-top: 1px solid var(--border);
  text-decoration: none;
  color: var(--text);
}
.route-item strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 14px;
}
.route-item small {
  grid-column: 2;
  color: var(--subtext);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.item-mode {
  align-self: start;
  color: var(--salvia);
  font-size: 12px;
  font-weight: 700;
  text-transform: capitalize;
}
.empty-line { padding: 18px 0; color: var(--subtext); font-size: 13px; }

.district-list { display: flex; flex-direction: column; gap: 10px; }
.district-card {
  display: flex;
  align-items: center;
  gap: 14px;
  text-decoration: none;
  overflow: hidden;
  position: relative;
}
.district-card-art {
  width: 84px;
  height: 84px;
  border-radius: 14px;
  object-fit: cover;
  flex-shrink: 0;
  border: 1px solid var(--border);
}
.district-card-body { flex: 1; min-width: 0; }
.district-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.district-card-head h2 {
  margin: 0;
  font-family: 'Lora', serif;
  font-size: 17px;
  font-weight: 600;
  color: var(--text);
}
.district-card-locations { margin: 3px 0 8px; color: var(--subtext); font-size: 12px; }
.district-card-signals { display: grid; gap: 5px; }
.district-signal {
  display: grid;
  grid-template-columns: 90px 1fr;
  align-items: center;
  gap: 8px;
}
.district-signal span { font-size: 11px; color: var(--subtext); }

@media (max-width: 620px) {
  .city-overview { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .city-work-grid { grid-template-columns: 1fr; }
  .route-item { grid-template-columns: 1fr; }
  .route-item small { grid-column: auto; }
}
</style>
