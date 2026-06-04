<template>
  <div class="city-view">
    <header class="city-header">
      <div>
        <p class="city-kicker">{{ t('city.kicker') }}</p>
        <h1>{{ courseMap?.course.city_name || courseMap?.course.title || t('city.title') }}</h1>
      </div>
      <button type="button" class="secondary-button" :disabled="loading" @click="loadCourseMap">
        {{ t('common.retry') }}
      </button>
    </header>

    <div v-if="loading" class="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="error-card">
      <strong>{{ t('common.error') }}</strong>
      <p>{{ error }}</p>
    </div>

    <template v-else-if="courseMap">
      <section class="city-stats" aria-label="Course totals">
        <div class="city-stat">
          <span>{{ courseMap.totals.districts }}</span>
          <label>{{ t('city.districts') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ courseMap.totals.locations }}</span>
          <label>{{ t('city.locations') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ courseMap.totals.modules }}</span>
          <label>{{ t('city.modules') }}</label>
        </div>
        <div class="city-stat">
          <span>{{ courseMap.totals.items }}</span>
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
        <article v-for="district in courseMap.districts" :key="district.id" class="district-row">
          <aside class="district-meta">
            <span class="level-badge">{{ district.level_code }}</span>
            <h2>{{ district.title }}</h2>
            <p>{{ district.locations.length }} {{ t('city.locationsShort') }}</p>
          </aside>

          <div class="location-grid">
            <div v-for="location in district.locations" :key="location.id" class="location-cell">
              <div class="location-head">
                <span>{{ locationTitle(location.location_type, location.title) }}</span>
                <strong>{{ countLocationItems(location) }}</strong>
              </div>
              <div class="module-list">
                <div v-for="module in visibleModules(location)" :key="module.id" class="module-line">
                  <span>{{ module.title }}</span>
                  <small>{{ module.items.length }}</small>
                </div>
                <div v-if="location.modules.length > visibleModules(location).length" class="module-more">
                  {{ t('city.moreModules', { count: location.modules.length - visibleModules(location).length }) }}
                </div>
              </div>
            </div>
          </div>
        </article>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { courseClient, CourseMap, CourseMapLocation } from '../api/courseClient'

const { t } = useI18n()

const courseMap = ref<CourseMap | null>(null)
const loading = ref(false)
const error = ref('')

const itemTypes = computed(() => {
  const totals = courseMap.value?.totals.by_type || {}
  return Object.entries(totals).sort((left, right) => right[1] - left[1])
})

function formatType(type: string): string {
  return type.replace(/_/g, ' ')
}

function locationTitle(type: string, fallback: string): string {
  const key = `city.locationTypes.${type}`
  const translated = t(key)
  return translated === key ? fallback : translated
}

function visibleModules(location: CourseMapLocation) {
  return location.modules.slice(0, 4)
}

function countLocationItems(location: CourseMapLocation): number {
  return location.modules.reduce((sum, module) => sum + module.items.length, 0)
}

async function loadCourseMap() {
  loading.value = true
  error.value = ''
  try {
    courseMap.value = await courseClient.getCourseMap()
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  } finally {
    loading.value = false
  }
}

onMounted(loadCourseMap)
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

.city-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.city-stat,
.type-pill,
.location-cell {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--surface-color);
}

.city-stat {
  padding: 14px;
}

.city-stat span {
  display: block;
  font-size: 1.6rem;
  font-weight: 750;
}

.city-stat label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.type-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
}

@media (max-width: 620px) {
  .city-header {
    align-items: stretch;
    flex-direction: column;
  }

  .city-stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .location-grid {
    grid-template-columns: 1fr;
  }
}
</style>
