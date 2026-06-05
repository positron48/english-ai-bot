<template>
  <div class="admin-linglow-srs">
    <header class="page-header">
      <div>
        <h1>{{ t('adminLinglowSRS.title') }}</h1>
        <p>{{ t('adminLinglowSRS.subtitle') }}</p>
      </div>
      <div class="header-actions">
        <label class="course-select">
          <span>{{ t('city.course') }}</span>
          <select v-model="selectedCourseCode" :disabled="loading" @change="loadReport">
            <option v-for="course in courses" :key="course.code" :value="course.code">
              {{ course.title }}
            </option>
          </select>
        </label>
        <button type="button" class="btn btn-primary" :disabled="loading" @click="loadReport">
          {{ t('common.retry') }}
        </button>
      </div>
    </header>

    <div v-if="loading" class="state-card">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="state-card error-card">
      <strong>{{ t('common.error') }}</strong>
      <p>{{ error }}</p>
    </div>

    <template v-else-if="report">
      <section class="readiness-panel" :class="{ ready: aggregate?.ready_for_canonical_read }">
        <div>
          <span class="readiness-label">{{ t('adminLinglowSRS.readiness') }}</span>
          <strong>{{ aggregate?.ready_for_canonical_read ? t('adminLinglowSRS.ready') : t('adminLinglowSRS.notReady') }}</strong>
        </div>
        <small>{{ report.course.title }} · {{ formatDate(aggregate?.generated_at || report.generated_at) }}</small>
      </section>

      <section class="metric-grid">
        <article class="metric-card">
          <span>{{ aggregate?.user_courses_total || 0 }}</span>
          <label>{{ t('adminLinglowSRS.userCourses') }}</label>
        </article>
        <article class="metric-card">
          <span>{{ aggregate?.ready_count || 0 }} / {{ aggregate?.not_ready_count || 0 }}</span>
          <label>{{ t('adminLinglowSRS.readyUsers') }}</label>
        </article>
        <article class="metric-card">
          <span>{{ aggregate?.legacy_due_total || 0 }} / {{ aggregate?.canonical_due_total || 0 }}</span>
          <label>{{ t('adminLinglowSRS.queueTotals') }}</label>
        </article>
        <article class="metric-card danger">
          <span>{{ aggregate?.legacy_only_total || 0 }} / {{ aggregate?.canonical_only_total || 0 }}</span>
          <label>{{ t('adminLinglowSRS.onlyMismatch') }}</label>
        </article>
      </section>

      <section class="report-grid">
        <article class="report-card">
          <h2>{{ t('adminLinglowSRS.aggregateReviewQueue') }}</h2>
          <div class="type-list">
            <div v-for="[type, count] in aggregateTypes" :key="type" class="type-row">
              <span>{{ formatType(type) }}</span>
              <strong>{{ count }}</strong>
            </div>
            <div v-if="aggregateTypes.length === 0" class="empty-line">{{ t('common.noItemsFound') }}</div>
          </div>
        </article>

        <article class="report-card">
          <h2>{{ t('adminLinglowSRS.currentUserQueue') }}</h2>
          <dl class="detail-list">
            <div><dt>{{ t('adminLinglowSRS.legacy') }}</dt><dd>{{ report.review_queue.legacy_due_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.canonical') }}</dt><dd>{{ report.review_queue.canonical_due_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.overlap') }}</dt><dd>{{ report.review_queue.overlap_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.legacyOnly') }}</dt><dd>{{ report.review_queue.legacy_only_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.canonicalOnly') }}</dt><dd>{{ report.review_queue.canonical_only_count }}</dd></div>
          </dl>
        </article>

        <article class="report-card">
          <h2>{{ t('adminLinglowSRS.wordDue') }}</h2>
          <dl class="detail-list">
            <div><dt>{{ t('adminLinglowSRS.legacy') }}</dt><dd>{{ report.due.legacy_due_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.canonical') }}</dt><dd>{{ report.due.linglow_due_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.overlap') }}</dt><dd>{{ report.due.overlap_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.legacyOnly') }}</dt><dd>{{ report.due.legacy_only_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.canonicalOnly') }}</dt><dd>{{ report.due.linglow_only_count }}</dd></div>
          </dl>
        </article>

        <article class="report-card">
          <h2>{{ t('adminLinglowSRS.mastery') }}</h2>
          <dl class="detail-list">
            <div><dt>{{ t('adminLinglowSRS.compared') }}</dt><dd>{{ report.mastery.compared_count }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.avgLegacy') }}</dt><dd>{{ formatNumber(report.mastery.average_legacy) }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.avgCanonical') }}</dt><dd>{{ formatNumber(report.mastery.average_linglow) }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.avgDiff') }}</dt><dd>{{ formatNumber(report.mastery.average_difference) }}</dd></div>
            <div><dt>{{ t('adminLinglowSRS.maxDiff') }}</dt><dd>{{ formatNumber(report.mastery.max_difference) }}</dd></div>
          </dl>
        </article>
      </section>

      <section v-if="aggregate?.not_ready_users?.length" class="report-card">
        <h2>{{ t('adminLinglowSRS.notReadyUsers') }}</h2>
        <div class="user-list">
          <div v-for="user in aggregate.not_ready_users" :key="user.user_course_id" class="user-row">
            <span>#{{ user.user_id }} / {{ user.user_course_id }}</span>
            <strong>{{ user.legacy_only_count }} / {{ user.canonical_only_count }}</strong>
          </div>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { courseClient, CourseSummary, SRSReadinessAggregateReport, SRSShadowReport } from '../api/courseClient'

const { t } = useI18n()

const courses = ref<CourseSummary[]>([])
const selectedCourseCode = ref('')
const report = ref<SRSShadowReport | null>(null)
const aggregate = ref<SRSReadinessAggregateReport | null>(null)
const loading = ref(false)
const error = ref('')

const aggregateTypes = computed(() => {
  const byType = aggregate.value?.by_type || {}
  return Object.entries(byType).sort((left, right) => right[1] - left[1])
})

function formatType(type: string): string {
  return type.replace(/_/g, ' ')
}

function formatNumber(value: number): string {
  return Number.isFinite(value) ? value.toFixed(1) : '0.0'
}

function formatDate(value: string): string {
  try {
    return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
  } catch {
    return value
  }
}

async function loadReport() {
  loading.value = true
  error.value = ''
  try {
    if (courses.value.length === 0) {
      const courseList = await courseClient.getCourses()
      courses.value = courseList.courses || []
      selectedCourseCode.value = courses.value.find((course) => course.is_current)?.code || courses.value[0]?.code || ''
    }
    const courseCode = selectedCourseCode.value || undefined
    const [shadowReport, aggregateReport] = await Promise.all([
      courseClient.getSRSShadowReport(courseCode),
      courseClient.getSRSReadinessAggregate(courseCode, 20),
    ])
    report.value = shadowReport
    aggregate.value = aggregateReport
  } catch (err: any) {
    error.value = err?.message || t('common.networkError')
  } finally {
    loading.value = false
  }
}

onMounted(loadReport)
</script>

<style scoped>
.admin-linglow-srs {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
}

.page-header {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 16px;
}

.page-header h1 {
  margin: 0 0 6px;
  color: var(--text-primary);
  font-size: 1.75rem;
}

.page-header p {
  margin: 0;
  color: var(--text-secondary);
}

.header-actions {
  display: flex;
  align-items: end;
  gap: 10px;
}

.course-select {
  display: grid;
  gap: 4px;
  min-width: 240px;
}

.course-select span {
  color: var(--text-secondary);
  font-size: 0.78rem;
  font-weight: 700;
}

.course-select select {
  width: 100%;
  min-height: 40px;
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--card-bg);
  color: var(--text-primary);
  padding: 8px 10px;
}

.btn {
  min-height: 40px;
  border: none;
  border-radius: 8px;
  padding: 0 14px;
  font-weight: 700;
  cursor: pointer;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn:disabled {
  opacity: 0.6;
  cursor: default;
}

.state-card,
.readiness-panel,
.metric-card,
.report-card {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  background: var(--card-bg);
}

.state-card {
  padding: 18px;
}

.error-card p {
  margin: 6px 0 0;
  color: var(--text-secondary);
}

.readiness-panel {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border-color: var(--color-danger);
}

.readiness-panel.ready {
  border-color: var(--color-success);
}

.readiness-label {
  display: block;
  color: var(--text-secondary);
  font-size: 0.82rem;
  font-weight: 700;
}

.readiness-panel strong {
  display: block;
  margin-top: 4px;
  font-size: 1.35rem;
}

.readiness-panel small {
  color: var(--text-secondary);
}

.metric-grid,
.report-grid {
  display: grid;
  gap: 12px;
}

.metric-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.report-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.metric-card,
.report-card {
  padding: 14px;
}

.metric-card span {
  display: block;
  font-size: 1.5rem;
  font-weight: 800;
}

.metric-card label,
.empty-line {
  color: var(--text-secondary);
}

.metric-card.danger span {
  color: var(--color-danger);
}

.report-card h2 {
  margin: 0 0 12px;
  font-size: 1rem;
}

.type-list,
.detail-list,
.user-list {
  display: grid;
  gap: 8px;
}

.type-row,
.detail-list div,
.user-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 12px;
  padding: 8px 0;
  border-top: 1px solid var(--border-primary);
}

.type-row span,
.detail-list dt {
  color: var(--text-secondary);
  text-transform: capitalize;
}

.detail-list {
  margin: 0;
}

.detail-list dd {
  margin: 0;
  font-weight: 800;
}

.user-row span {
  color: var(--text-secondary);
}

@media (max-width: 900px) {
  .page-header,
  .header-actions,
  .readiness-panel {
    align-items: stretch;
    flex-direction: column;
  }

  .course-select {
    min-width: 0;
  }

  .metric-grid,
  .report-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .metric-grid,
  .report-grid {
    grid-template-columns: 1fr;
  }
}
</style>
