<template>
  <div class="district-view">
    <header class="district-header">
      <RouterLink class="back-link" to="/city">{{ t('city.backToCity') }}</RouterLink>
      <div>
        <p class="city-kicker">{{ courseMap?.course.city_name || t('city.title') }}</p>
        <h1>{{ district?.title || t('city.district') }}</h1>
      </div>
      <span v-if="district" class="level-badge">{{ district.level_code }}</span>
    </header>

    <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="error-card">
      <strong>{{ t('common.error') }}</strong>
      <p>{{ error }}</p>
    </div>

    <template v-else-if="district">
      <section class="district-stats">
        <div class="stat-box">
          <span>{{ locations.length }}</span>
          <label>{{ t('city.locations') }}</label>
        </div>
        <div class="stat-box">
          <span>{{ moduleCount }}</span>
          <label>{{ t('city.modules') }}</label>
        </div>
        <div class="stat-box">
          <span>{{ itemCount }}</span>
          <label>{{ t('city.items') }}</label>
        </div>
        <div class="stat-box">
          <span>{{ formatPercent(districtSignal.foundation) }}</span>
          <label>{{ t('city.foundation') }}</label>
        </div>
      </section>

      <section class="location-sections">
        <article v-for="location in locations" :id="location.code" :key="location.id" class="location-section">
          <div class="location-title">
            <div>
              <p>{{ locationTitle(location.location_type, location.title) }}</p>
              <h2>{{ location.title }}</h2>
            </div>
            <strong>{{ formatPercent(locationSignal(location.code).foundation) }}</strong>
          </div>
          <div class="location-signals">
            <span>{{ t('city.confidence') }} {{ formatPercent(locationSignal(location.code).confidence) }}</span>
            <span>{{ t('city.stability') }} {{ formatPercent(locationSignal(location.code).stability) }}</span>
            <span>{{ t('city.reviewPressure') }} {{ locationSignal(location.code).due_review_count }}</span>
          </div>

          <div class="module-grid">
            <article v-for="module in safeModules(location)" :key="module.id" class="module-card">
              <div class="module-head">
                <div>
                  <span>{{ formatType(module.type) }}</span>
                  <h3>{{ module.title }}</h3>
                </div>
                <RouterLink class="module-action" :to="routeForLinglowItem(module)">
                  {{ t('city.openModule') }}
                </RouterLink>
              </div>

              <div class="item-list">
                <RouterLink v-for="item in visibleItems(module)" :key="item.id" class="item-row" :to="routeForLinglowItem(item)">
                  <span>{{ item.title || formatType(item.type) }}</span>
                  <small>{{ item.cefr_level || formatType(item.type) }}</small>
                </RouterLink>
                <div v-if="safeItems(module).length > visibleItems(module).length" class="item-more">
                  {{ t('city.moreItems', { count: safeItems(module).length - visibleItems(module).length }) }}
                </div>
                <div v-if="safeItems(module).length === 0" class="item-more">{{ t('city.noItems') }}</div>
              </div>
            </article>
          </div>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { courseClient, CourseMap, CourseMapLocation, CourseProgress } from '../api/courseClient'
import { routeForLinglowItem } from '../utils/linglowNavigation'

const { t } = useI18n()
const route = useRoute()

const courseMap = ref<CourseMap | null>(null)
const progress = ref<CourseProgress | null>(null)
const loading = ref(false)
const error = ref('')

const district = computed(() => {
  const districts = Array.isArray(courseMap.value?.districts) ? courseMap.value?.districts : []
  return districts.find((item) => item.code === route.params.districtCode) || null
})

const locations = computed(() => (district.value && Array.isArray(district.value.locations) ? district.value.locations : []))
const moduleCount = computed(() => locations.value.reduce((sum, location) => sum + safeModules(location).length, 0))
const itemCount = computed(() => locations.value.reduce((sum, location) => sum + countLocationItems(location), 0))
const locationProgressByCode = computed(() => {
  const out: Record<string, NonNullable<CourseProgress['by_location']>[number]> = {}
  for (const row of progress.value?.by_location || []) {
    out[row.location_code] = row
  }
  return out
})
const districtSignal = computed(() => {
  return (progress.value?.by_district || []).find((item) => item.district_code === route.params.districtCode) || emptySignal()
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

function locationSignal(code: string) {
  return locationProgressByCode.value[code] || emptySignal()
}

function locationTitle(type: string, fallback: string): string {
  const key = `city.locationTypes.${type}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

function safeModules(location: CourseMapLocation) {
  return Array.isArray(location.modules) ? location.modules : []
}

function safeItems(module: CourseMapLocation['modules'][number]) {
  return Array.isArray(module.items) ? module.items : []
}

function visibleItems(module: CourseMapLocation['modules'][number]) {
  return safeItems(module).slice(0, 8)
}

function countLocationItems(location: CourseMapLocation): number {
  return safeModules(location).reduce((sum, module) => sum + safeItems(module).length, 0)
}

async function loadDistrict() {
  loading.value = true
  error.value = ''
  try {
    const courseCode = typeof route.query.course_code === 'string' ? route.query.course_code : undefined
    const [map, progressData] = await Promise.all([
      courseClient.getCourseMap(courseCode),
      courseClient.getProgress(courseCode),
    ])
    courseMap.value = map
    progress.value = progressData
    await nextTick()
    if (typeof route.query.location === 'string') {
      document.getElementById(route.query.location)?.scrollIntoView({ block: 'start' })
    }
    if (!district.value) {
      error.value = t('city.districtNotFound')
    }
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  } finally {
    loading.value = false
  }
}

onMounted(loadDistrict)
</script>

<style scoped>
.district-view {
  display: flex;
  flex-direction: column;
  gap: 18px;
  padding-bottom: 28px;
}

.district-header {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: end;
  gap: 14px;
}

.back-link,
.module-action,
.item-row {
  color: inherit;
  text-decoration: none;
}

.back-link,
.module-action {
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

.district-header h1 {
  margin: 0;
  font-size: 1.8rem;
  line-height: 1.1;
}

.level-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  height: 30px;
  border-radius: 6px;
  background: var(--primary-color);
  color: white;
  font-weight: 800;
}

.loading,
.error-card,
.stat-box,
.module-card {
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

.district-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.stat-box {
  padding: 14px;
}

.stat-box span {
  display: block;
  font-size: 1.45rem;
  font-weight: 800;
}

.stat-box label,
.location-title p,
.module-head span,
.item-row small,
.item-more {
  color: var(--text-secondary);
}

.location-sections {
  display: grid;
  gap: 22px;
}

.location-section {
  scroll-margin-top: 18px;
}

.location-title {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  padding-top: 14px;
  border-top: 1px solid var(--border-color);
}

.location-title p,
.location-title h2,
.module-head h3 {
  margin: 0;
}

.location-title h2 {
  font-size: 1.2rem;
}

.location-title strong {
  color: var(--primary-color);
}

.location-signals {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin: -2px 0 10px;
  color: var(--text-secondary);
  font-size: 0.82rem;
  font-weight: 650;
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.module-card {
  padding: 13px;
}

.module-head {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.module-head h3 {
  font-size: 1rem;
  line-height: 1.25;
}

.item-list {
  display: grid;
  gap: 6px;
}

.item-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
  padding: 8px 0;
  border-top: 1px solid var(--border-color);
}

.item-row span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-row:hover span,
.module-action:hover,
.back-link:hover {
  color: var(--primary-color);
}

.item-more {
  padding-top: 8px;
}

@media (max-width: 760px) {
  .district-header {
    grid-template-columns: 1fr auto;
  }

  .back-link {
    grid-column: 1 / -1;
  }

  .module-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .district-stats {
    grid-template-columns: 1fr;
  }

  .item-row {
    grid-template-columns: 1fr;
  }
}
</style>
