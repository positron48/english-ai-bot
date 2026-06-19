<template>
  <div v-if="fact" class="lg-lumi-fact">
    <LgLumi :size="lumiSize" />
    <div class="lg-lumi-fact-body">
      <div class="lg-lumi-fact-head">
        <span class="lg-lumi-fact-star">✦</span>
        <span class="lg-lumi-fact-title">{{ t('lg.lumiKnows') }}</span>
      </div>
      <p class="lg-lumi-fact-text">{{ fact }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { factClient } from '../../api/factClient'
import { useCourse } from '../../composables/useCourse'
import { useLocale } from '../../composables/useLocale'
import LgLumi from './LgLumi.vue'

const props = withDefaults(defineProps<{ lumiSize?: number; context?: string }>(), {
  lumiSize: 52,
  context: 'general',
})

const { t } = useI18n()
const { currentCourseCode } = useCourse()
const { currentLocale } = useLocale()

const fact = ref('')

const cacheKey = () => {
  const day = new Date().toISOString().slice(0, 10)
  return `lumi-fact:${currentCourseCode.value || ''}:${props.context}:${currentLocale.value}:${day}`
}

onMounted(async () => {
  // Same-day cache: instant render and offline support
  try {
    const cached = localStorage.getItem(cacheKey())
    if (cached) {
      fact.value = cached
      return
    }
  } catch { /* ignore */ }
  const dto = await factClient.getDailyFact(props.context, currentCourseCode.value || undefined, currentLocale.value)
  if (dto) {
    fact.value = dto.body
    try { localStorage.setItem(cacheKey(), dto.body) } catch { /* ignore */ }
  }
})
</script>

<style scoped>
.lg-lumi-fact {
  display: flex;
  gap: 12px;
  align-items: center;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px;
}
.lg-lumi-fact-body { flex: 1; }
.lg-lumi-fact-head {
  display: flex;
  gap: 4px;
  align-items: center;
  margin-bottom: 4px;
}
.lg-lumi-fact-star { font-size: 11px; color: var(--dorado); }
.lg-lumi-fact-title { font-size: 12px; font-weight: 600; color: var(--dorado); }
.lg-lumi-fact-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text);
}
</style>
