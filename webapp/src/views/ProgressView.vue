<template>
  <div class="lg-page">
    <LgPageHeader :title="t('progress.title')" />
    <div class="lg-page-body">
      <div v-if="loading" class="lg-loading">{{ t('common.loading') }}</div>
      <template v-else>
        <!-- Words learned + activity summary -->
        <div v-if="stats" class="lg-stat-grid">
          <div class="lg-card lg-stat">
            <div class="lg-stat-num">{{ stats.total_words ?? 0 }}</div>
            <div class="lg-stat-label">{{ t('progress.totalWords') }}</div>
          </div>
          <div class="lg-card lg-stat">
            <div class="lg-stat-num">{{ stats.learned_words ?? 0 }}</div>
            <div class="lg-stat-label">{{ t('progress.learnedWords') }}</div>
          </div>
          <div class="lg-card lg-stat">
            <div class="lg-stat-num">{{ stats.due_words ?? 0 }}</div>
            <div class="lg-stat-label">{{ t('progress.dueWords') }}</div>
          </div>
        </div>

        <!-- Learning progress ring -->
        <div v-if="stats && (stats.total_words ?? 0) > 0" class="lg-card lg-ring-card">
          <LgCircleRing :val="stats.learned_words ?? 0" :max="stats.total_words ?? 1" :size="84" :stroke="7">
            <span class="lg-ring-pct">{{ learnedPct }}%</span>
          </LgCircleRing>
          <div>
            <div class="lg-ring-title">{{ t('progress.vocabProgress') }}</div>
            <div class="lg-ring-sub">{{ t('progress.vocabProgressHint') }}</div>
          </div>
        </div>

        <LgLumiFact />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgCircleRing from '../components/linglow/LgCircleRing.vue'
import LgLumiFact from '../components/linglow/LgLumiFact.vue'

const { t } = useI18n()

interface DashboardStats {
  total_words?: number
  learned_words?: number
  due_words?: number
}

const loading = ref(true)
const stats = ref<DashboardStats | null>(null)

const learnedPct = computed(() => {
  const total = stats.value?.total_words ?? 0
  if (!total) return 0
  return Math.round(((stats.value?.learned_words ?? 0) / total) * 100)
})

onMounted(async () => {
  try {
    const data: any = await apiClient.request('/api/dashboard')
    stats.value = data?.stats ?? data ?? null
  } catch {
    stats.value = null
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.lg-page { padding-bottom: 24px; }
.lg-page-body {
  padding: 8px 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.lg-loading {
  padding: 40px 0;
  text-align: center;
  color: var(--subtext);
}
.lg-card {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px;
  box-shadow: var(--shadow-soft);
}
.lg-stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}
.lg-stat { text-align: center; }
.lg-stat-num {
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 24px;
  color: var(--text);
}
.lg-stat-label {
  font-size: 12px;
  color: var(--subtext);
  margin-top: 2px;
}
.lg-ring-card {
  display: flex;
  align-items: center;
  gap: 16px;
}
.lg-ring-pct { font-weight: 700; font-size: 16px; color: var(--text); }
.lg-ring-title { font-weight: 600; font-size: 15px; color: var(--text); }
.lg-ring-sub { font-size: 13px; color: var(--subtext); margin-top: 2px; }
</style>
