<template>
  <div class="daily-route-view">
    <header class="route-header">
      <RouterLink class="back-link" :to="cityLink">{{ t('city.backToCity') }}</RouterLink>
      <div>
        <p class="city-kicker">{{ courseTitle }}</p>
        <h1>{{ t('city.dailyRoute') }}</h1>
      </div>
      <button type="button" class="secondary-button" :disabled="refreshing" @click="loadRoute(true)">
        {{ t('common.retry') }}
      </button>
    </header>

    <div v-if="loadingInitial" class="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="error-card">
      <strong>{{ t('common.error') }}</strong>
      <p>{{ error }}</p>
    </div>

    <template v-else>
      <section class="route-summary">
        <div class="summary-box">
          <span>{{ route?.summary.due_review_count || review?.summary.due_count || 0 }}</span>
          <label>{{ t('city.reviewPressure') }}</label>
        </div>
        <div class="summary-box">
          <span>{{ route?.summary.new_item_count || 0 }}</span>
          <label>{{ t('city.nextOpenings') }}</label>
        </div>
        <div class="summary-box">
          <span>{{ formatPercent(progress?.summary.accuracy_percent || 0) }}</span>
          <label>{{ t('city.accuracy') }}</label>
        </div>
        <div class="summary-box">
          <span>{{ formatPercent(progress?.summary.progress_percent || 0) }}</span>
          <label>{{ t('city.foundation') }}</label>
        </div>
      </section>

      <section class="route-grid">
        <article class="route-panel">
          <div class="panel-head">
            <div>
              <p>{{ t('city.routeSteps.reviewBlock') }}</p>
              <h2>{{ t('city.reviewStation') }}</h2>
            </div>
            <span>{{ reviewItems.length }}</span>
          </div>
          <div class="item-list">
            <RouterLink v-for="item in reviewItems" :key="`review:${item.learning_item_id}`" class="route-item" :to="routeForLinglowItem(item)">
              <span class="item-mode">{{ formatType(item.state || item.mode) }}</span>
              <strong>{{ item.title || formatType(item.type) }}</strong>
              <small>{{ item.due_at ? formatDate(item.due_at) : item.location_title || item.module_title }}</small>
            </RouterLink>
            <div v-if="reviewItems.length === 0" class="empty-line">{{ t('city.noReviewItems') }}</div>
          </div>
        </article>

        <article class="route-panel">
          <div class="panel-head">
            <div>
              <p>{{ t('city.routeSteps.newBlock') }}</p>
              <h2>{{ t('city.nextOpenings') }}</h2>
            </div>
            <span>{{ newItems.length }}</span>
          </div>
          <div class="item-list">
            <RouterLink v-for="item in newItems" :key="`new:${item.learning_item_id}`" class="route-item" :to="routeForLinglowItem(item)">
              <span class="item-mode">{{ formatType(item.type) }}</span>
              <strong>{{ item.title || formatType(item.type) }}</strong>
              <small>{{ item.location_title || item.module_title || item.cefr_level }}</small>
            </RouterLink>
            <div v-if="newItems.length === 0" class="empty-line">{{ t('city.noDailyItems') }}</div>
          </div>
        </article>

        <article class="route-panel route-panel-wide">
          <div class="panel-head">
            <div>
              <p>{{ t('city.routeSteps.mistakeBlock') }}</p>
              <h2>{{ t('city.mistakeWorkshop') }}</h2>
            </div>
            <span>{{ weakLocations.length }}</span>
          </div>
          <div class="weak-grid">
            <RouterLink v-for="location in weakLocations" :key="location.location_code" class="weak-card" :to="districtLink(location.district_code, location.location_code)">
              <div>
                <span>{{ location.district_code }}</span>
                <strong>{{ location.title }}</strong>
              </div>
              <dl>
                <div>
                  <dt>{{ t('city.confidenceShort') }}</dt>
                  <dd>{{ formatPercent(location.confidence) }}</dd>
                </div>
                <div>
                  <dt>{{ t('city.weaknessShort') }}</dt>
                  <dd>{{ location.due_review_count }}</dd>
                </div>
                <div>
                  <dt>{{ t('city.stability') }}</dt>
                  <dd>{{ formatPercent(location.stability) }}</dd>
                </div>
              </dl>
            </RouterLink>
            <div v-if="weakLocations.length === 0" class="empty-line">{{ t('city.noWeakItems') }}</div>
          </div>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { courseClient, CourseProgress, DailyRoute, ReviewQueue } from '../api/courseClient'
import { routeForLinglowItem } from '../utils/linglowNavigation'
import { useCachedOverviewScreen } from '../composables/useCachedOverviewScreen'
import { useCourse } from '../composables/useCourse'
import { useLocale } from '../composables/useLocale'
import { useAuth } from '../composables/useAuth'

const { t } = useI18n()
const currentRoute = useRoute()
const { currentCourseCode } = useCourse()
const { currentLocale } = useLocale()
const { isAuthenticated } = useAuth()

const route = ref<DailyRoute | null>(null)
const review = ref<ReviewQueue | null>(null)
const progress = ref<CourseProgress | null>(null)
const error = ref('')

const courseCode = computed(() => (typeof currentRoute.query.course_code === 'string' ? currentRoute.query.course_code : undefined))
const resolvedCourseCode = computed(() => courseCode.value || currentCourseCode.value)

function applyDailyRouteBundle(bundle: any) {
  route.value = bundle?.route ?? null
  review.value = bundle?.review ?? null
  progress.value = bundle?.progress ?? null
}

const { loadingInitial, refreshing, load } = useCachedOverviewScreen<any>({
  screenKey: 'daily-route',
  courseCode: resolvedCourseCode,
  locale: currentLocale,
  fetcher: async () => {
    const code = resolvedCourseCode.value
    const [routeData, reviewData, progressData] = await Promise.all([
      courseClient.getDailyRoute(16, code),
      courseClient.getReviewQueue(16, code),
      courseClient.getProgress(code),
    ])
    return { route: routeData, review: reviewData, progress: progressData }
  },
  applyPayload: (bundle) => applyDailyRouteBundle(bundle),
})
const courseTitle = computed(() => route.value?.course.city_name || route.value?.course.title || t('city.title'))
const reviewItems = computed(() => (review.value?.items || route.value?.review || []).slice(0, 12))
const newItems = computed(() => (route.value?.new_items || []).slice(0, 12))
const cityLink = computed(() => ({
  name: 'City',
  query: courseCode.value ? { course_code: courseCode.value } : undefined,
}))

const weakLocations = computed(() => {
  return [...(progress.value?.by_location || [])]
    .filter((location) => location.due_review_count > 0 || location.weakness > 0)
    .sort((left, right) => {
      if (right.due_review_count !== left.due_review_count) return right.due_review_count - left.due_review_count
      return right.weakness - left.weakness
    })
    .slice(0, 6)
})

function formatType(type: string): string {
  return type.replace(/_/g, ' ')
}

function formatPercent(value: number): string {
  return `${Math.round(value)}%`
}

function formatDate(value: string): string {
  try {
    return new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' }).format(new Date(value))
  } catch {
    return value
  }
}

function districtLink(districtCode: string, locationCode: string) {
  return {
    name: 'CityDistrict',
    params: { districtCode },
    query: courseCode.value ? { course_code: courseCode.value, location: locationCode } : { location: locationCode },
  }
}

async function loadRoute(force = false) {
  error.value = ''
  try {
    await load(force)
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  }
}

watch(isAuthenticated, (authenticated) => {
  if (authenticated) void loadRoute()
}, { immediate: true })

onMounted(() => {
  if (isAuthenticated.value) void loadRoute()
})
</script>

<style scoped>
.daily-route-view {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding-bottom: 28px;
}

.route-header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: end;
  gap: 14px;
}

.back-link,
.route-item,
.weak-card {
  color: inherit;
  text-decoration: none;
}

.back-link {
  color: var(--primary-color);
  font-weight: 750;
}

.city-kicker {
  margin: 0 0 4px;
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 650;
  text-transform: uppercase;
}

.route-header h1 {
  margin: 0;
  font-size: 1.9rem;
  line-height: 1.1;
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
.error-card,
.summary-box,
.route-panel,
.weak-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
}

.loading,
.error-card {
  padding: 18px;
}

.error-card p {
  margin: 6px 0 0;
  color: var(--text-secondary);
}

.route-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.summary-box {
  padding: 14px;
}

.summary-box span {
  display: block;
  font-size: 1.45rem;
  font-weight: 800;
}

.summary-box label {
  color: var(--text-secondary);
  font-size: 0.86rem;
}

.route-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.route-panel {
  min-height: 260px;
  padding: 14px;
}

.route-panel-wide {
  grid-column: 1 / -1;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.panel-head p {
  margin: 0 0 3px;
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 750;
  text-transform: uppercase;
}

.panel-head h2 {
  margin: 0;
  font-size: 1rem;
}

.panel-head span {
  color: var(--text-secondary);
  font-weight: 800;
}

.item-list {
  display: grid;
  gap: 8px;
}

.route-item {
  display: grid;
  grid-template-columns: 112px minmax(0, 1fr);
  gap: 4px 10px;
  padding: 10px 0;
  border-top: 1px solid var(--border-color);
}

.route-item:hover strong,
.weak-card:hover strong {
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

.weak-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.weak-card {
  display: grid;
  gap: 12px;
  padding: 12px;
}

.weak-card span {
  display: block;
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 750;
}

.weak-card strong {
  display: block;
  margin-top: 4px;
}

.weak-card dl {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin: 0;
}

.weak-card dt {
  color: var(--text-secondary);
  font-size: 0.75rem;
}

.weak-card dd {
  margin: 2px 0 0;
  font-weight: 800;
}

@media (max-width: 900px) {
  .route-grid,
  .route-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .weak-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .route-header {
    grid-template-columns: 1fr;
    align-items: stretch;
  }

  .route-summary,
  .route-grid,
  .weak-grid {
    grid-template-columns: 1fr;
  }

  .route-item {
    grid-template-columns: 1fr;
  }

  .route-item small {
    grid-column: auto;
  }
}
</style>
