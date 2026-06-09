<template>
  <div class="city-view">
    <header class="city-header">
      <div>
        <p class="city-kicker">{{ t('city.kicker') }}</p>
        <h1>{{ courseMap?.course.city_name || courseMap?.course.title || t('city.title') }}</h1>
      </div>
      <div class="city-actions">
        <label class="course-select">
          <span>{{ t('city.course') }}</span>
          <select v-model="selectedCourseCode" :disabled="loading || selectingCourse" @change="selectCourse">
            <option v-for="course in courses" :key="course.code" :value="course.code">
              {{ course.title }}
            </option>
          </select>
        </label>
        <button type="button" class="secondary-button" :disabled="loading" @click="loadCity">
          {{ t('common.retry') }}
        </button>
      </div>
    </header>

    <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="error-card">
      <strong>{{ t('common.error') }}</strong>
      <p>{{ error }}</p>
    </div>

    <template v-else-if="safeCourseMap">
      <section class="city-overview">
        <div class="overview-metric">
          <span>{{ formatPercent(progress?.summary.progress_percent || 0) }}</span>
          <label>{{ t('city.progress') }}</label>
        </div>
        <div class="overview-metric">
          <span>{{ progress?.summary.due_review_count || reviewQueue?.summary.due_count || 0 }}</span>
          <label>{{ t('city.reviewPressure') }}</label>
        </div>
        <div class="overview-metric">
          <span>{{ dailyRoute?.summary.new_item_count || 0 }}</span>
          <label>{{ t('city.nextOpenings') }}</label>
        </div>
        <div class="overview-metric">
          <span>{{ formatPercent(progress?.summary.accuracy_percent || 0) }}</span>
          <label>{{ t('city.accuracy') }}</label>
        </div>
      </section>

      <section class="city-work-grid">
        <article class="work-panel">
          <div class="panel-head">
            <h2>{{ t('city.dailyRoute') }}</h2>
            <RouterLink class="panel-action" :to="dailyRouteLink">
              {{ t('city.openDailyRoute') }}
            </RouterLink>
          </div>
          <div class="route-list">
            <RouterLink v-for="item in dailyItems" :key="`${item.mode}:${item.learning_item_id}`" class="route-item route-link" :to="routeForLinglowItem(item)">
              <span class="item-mode">{{ formatType(item.mode) }}</span>
              <strong>{{ item.title || formatType(item.type) }}</strong>
              <small>{{ item.location_title || item.module_title || item.cefr_level }}</small>
            </RouterLink>
            <div v-if="dailyItems.length === 0" class="empty-line">{{ t('city.noDailyItems') }}</div>
          </div>
        </article>

        <article class="work-panel">
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

      <section class="simple-mode" aria-label="Simple mode">
        <RouterLink class="simple-link primary" to="/training">
          <strong>{{ t('city.simple.review') }}</strong>
          <span>{{ t('city.simple.reviewHint') }}</span>
        </RouterLink>
        <RouterLink class="simple-link" :to="dailyRouteLink">
          <strong>{{ t('city.simple.dailyRoute') }}</strong>
          <span>{{ t('city.simple.dailyRouteHint') }}</span>
        </RouterLink>
        <RouterLink class="simple-link" to="/learning/grammar">
          <strong>{{ t('city.simple.grammar') }}</strong>
          <span>{{ t('city.simple.grammarHint') }}</span>
        </RouterLink>
        <RouterLink class="simple-link" to="/learning/reading">
          <strong>{{ t('city.simple.reading') }}</strong>
          <span>{{ t('city.simple.readingHint') }}</span>
        </RouterLink>
        <RouterLink class="simple-link" to="/learning/words">
          <strong>{{ t('city.simple.words') }}</strong>
          <span>{{ t('city.simple.wordsHint') }}</span>
        </RouterLink>
        <RouterLink class="simple-link" to="/settings">
          <strong>{{ t('city.simple.settings') }}</strong>
          <span>{{ t('city.simple.settingsHint') }}</span>
        </RouterLink>
      </section>

      <section class="city-stats" aria-label="Course totals">
        <div class="city-stat">
          <span>{{ safeCourseMap.totals.districts }}</span>
          <label>{{ t('city.districts') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ safeCourseMap.totals.locations }}</span>
          <label>{{ t('city.locations') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ safeCourseMap.totals.modules }}</span>
          <label>{{ t('city.modules') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ safeCourseMap.totals.items }}</span>
          <label>{{ t('city.items') }}</label>
        </div>
      </section>

      <section class="type-strip" aria-label="Content item types">
        <div v-for="[type, count] in itemTypes" :key="type" class="type-pill">
          <span>{{ formatType(type) }}</span>
          <strong>{{ count }}</strong>
        </div>
      </section>

      <section class="district-list">
        <article v-for="district in safeCourseMap.districts" :key="district.id" class="district-row">
          <aside class="district-meta">
            <span class="level-badge">{{ district.level_code }}</span>
            <h2>{{ district.title }}</h2>
            <p>{{ safeLocations(district).length }} {{ t('city.locationsShort') }}</p>
            <div class="signal-stack">
              <span>{{ t('city.foundation') }} {{ formatPercent(districtSignal(district.code).foundation) }}</span>
              <span>{{ t('city.confidence') }} {{ formatPercent(districtSignal(district.code).confidence) }}</span>
              <span>{{ t('city.stability') }} {{ formatPercent(districtSignal(district.code).stability) }}</span>
            </div>
            <RouterLink class="district-link" :to="{ name: 'CityDistrict', params: { districtCode: district.code }, query: selectedCourseCode ? { course_code: selectedCourseCode } : undefined }">
              {{ t('city.openDistrict') }}
            </RouterLink>
          </aside>

          <div class="location-grid">
            <RouterLink v-for="location in safeLocations(district)" :key="location.id" class="location-cell location-link" :to="{ name: 'CityDistrict', params: { districtCode: district.code }, query: selectedCourseCode ? { course_code: selectedCourseCode, location: location.code } : { location: location.code } }">
              <div class="location-head">
                <span>{{ locationTitle(location.location_type, location.title) }}</span>
                <strong>{{ formatPercent(locationSignal(location.code).foundation) }}</strong>
              </div>
              <div class="location-signals">
                <span>{{ t('city.confidenceShort') }} {{ formatPercent(locationSignal(location.code).confidence) }}</span>
                <span>{{ t('city.weaknessShort') }} {{ locationSignal(location.code).due_review_count }}</span>
              </div>
              <div class="module-list">
                <div v-for="module in visibleModules(location)" :key="module.id" class="module-line">
                  <span>{{ module.title }}</span>
                  <small>{{ safeItems(module).length }}</small>
                </div>
                <div v-if="safeModules(location).length > visibleModules(location).length" class="module-more">
                  {{ t('city.moreModules', { count: safeModules(location).length - visibleModules(location).length }) }}
                </div>
              </div>
            </RouterLink>
          </div>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { courseClient, CourseMap, CourseMapLocation, CourseProgress, DailyRoute, ReviewQueue } from '../api/courseClient'
import { useCourse } from '../composables/useCourse'
import { routeForLinglowItem } from '../utils/linglowNavigation'

const { t } = useI18n()
const { courses, currentCourseCode, ensureCourseLoaded, selectCourse: setCurrentCourse } = useCourse()

const courseMap = ref<CourseMap | null>(null)
const dailyRoute = ref<DailyRoute | null>(null)
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

const dailyItems = computed(() => {
  const route = dailyRoute.value
  if (!route) return []
  return [...(route.review || []), ...(route.new_items || [])].slice(0, 8)
})

const reviewItems = computed(() => (reviewQueue.value?.items || []).slice(0, 6))

const dailyRouteLink = computed(() => ({
  name: 'CityDailyRoute',
  query: selectedCourseCode.value ? { course_code: selectedCourseCode.value } : undefined,
}))

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
    const [map, route, review, progressData] = await Promise.all([
      courseClient.getCourseMap(courseCode),
      courseClient.getDailyRoute(8, courseCode),
      courseClient.getReviewQueue(8, courseCode),
      courseClient.getProgress(courseCode),
    ])
    courseMap.value = map
    dailyRoute.value = route
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
.city-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-bottom: 28px;
}

.city-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.city-kicker {
  margin: 0 0 4px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
}

.city-header h1 {
  margin: 0;
  font-size: 2rem;
  line-height: 1.1;
}

.city-actions {
  display: flex;
  align-items: flex-end;
  gap: 10px;
}

.course-select {
  display: grid;
  gap: 4px;
  min-width: 220px;
}

.course-select span {
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 700;
}

.course-select select {
  width: 100%;
  min-height: 40px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
  color: var(--text-primary);
  padding: 8px 10px;
  font-weight: 650;
}

.secondary-button {
  border: 1px solid var(--border-color);
  background: var(--surface-color);
  color: var(--text-primary);
  border-radius: 8px;
  padding: 10px 14px;
  font-weight: 600;
  cursor: pointer;
}

.secondary-button:disabled {
  opacity: 0.55;
  cursor: default;
}

.loading,
.error-card {
  padding: 18px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
}

.error-card p {
  margin: 6px 0 0;
  color: var(--text-secondary);
}

.city-overview,
.city-work-grid,
.city-stats {
  display: grid;
  gap: 12px;
}

.city-overview,
.city-stats {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.city-work-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.simple-mode {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 10px;
}

.city-stat,
.overview-metric,
.work-panel,
.type-pill,
.location-cell,
.simple-link {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
}

.city-stat {
  padding: 14px;
}

.overview-metric {
  padding: 14px;
}

.city-stat span {
  display: block;
  font-size: 1.6rem;
  font-weight: 750;
}

.overview-metric span {
  display: block;
  font-size: 1.45rem;
  font-weight: 800;
}

.city-stat label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.overview-metric label {
  color: var(--text-secondary);
  font-size: 0.86rem;
}

.work-panel {
  min-height: 220px;
  padding: 14px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-head h2 {
  margin: 0;
  font-size: 1rem;
}

.panel-head span {
  color: var(--text-secondary);
  font-weight: 750;
}

.panel-action {
  color: var(--primary-color);
  font-size: 0.86rem;
  font-weight: 750;
  text-decoration: none;
}

.route-list {
  display: grid;
  gap: 8px;
}

.route-item {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 4px 10px;
  padding: 9px 0;
  border-top: 1px solid var(--border-color);
}

.route-link,
.location-link,
.district-link,
.simple-link {
  color: inherit;
  text-decoration: none;
}

.route-link:hover strong,
.location-link:hover .location-head span,
.district-link:hover,
.simple-link:hover strong {
  color: var(--primary-color);
}

.route-item strong {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-item small {
  grid-column: 2;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-mode {
  align-self: start;
  color: var(--primary-color);
  font-size: 0.78rem;
  font-weight: 800;
  text-transform: capitalize;
}

.empty-line {
  padding: 18px 0;
  color: var(--text-secondary);
}

.type-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.simple-link {
  display: grid;
  gap: 4px;
  min-height: 72px;
  padding: 12px;
}

.simple-link.primary {
  border-color: var(--primary-color);
}

.simple-link strong {
  font-size: 0.95rem;
}

.simple-link span {
  color: var(--text-secondary);
  font-size: 0.82rem;
  line-height: 1.3;
}

.type-pill {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  text-transform: capitalize;
}

.district-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.district-row {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 14px;
  padding: 14px 0;
  border-top: 1px solid var(--border-color);
}

.district-meta h2 {
  margin: 8px 0 4px;
  font-size: 1.05rem;
}

.district-meta p {
  margin: 0;
  color: var(--text-secondary);
}

.signal-stack {
  display: grid;
  gap: 4px;
  margin-top: 10px;
  color: var(--text-secondary);
  font-size: 0.82rem;
}

.district-link {
  display: inline-flex;
  margin-top: 10px;
  color: var(--primary-color);
  font-weight: 700;
}

.level-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 28px;
  border-radius: 6px;
  background: var(--primary-color);
  color: white;
  font-weight: 750;
}

.location-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.location-cell {
  min-height: 128px;
  padding: 12px;
}

.location-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
  font-weight: 700;
}

.location-head strong {
  color: var(--primary-color);
}

.location-signals {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 8px;
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 650;
}

.module-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.module-line {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.module-line span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.module-line small,
.module-more {
  color: var(--text-muted);
}

@media (max-width: 900px) {
  .district-row {
    grid-template-columns: 1fr;
  }

  .location-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .city-work-grid {
    grid-template-columns: 1fr;
  }

  .simple-mode {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .city-header {
    align-items: stretch;
    flex-direction: column;
  }

  .city-actions {
    align-items: stretch;
    flex-direction: column;
  }

  .course-select {
    min-width: 0;
  }

  .city-overview,
  .city-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .route-item {
    grid-template-columns: 1fr;
  }

  .route-item small {
    grid-column: auto;
  }

  .location-grid {
    grid-template-columns: 1fr;
  }

  .simple-mode {
    grid-template-columns: 1fr;
  }
}
</style>
