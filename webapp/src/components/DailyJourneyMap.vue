<template>
  <nav class="journey-map" :aria-label="t('dailyJourney.stops')">
    <div class="map-heading"><span>{{ t('dailyJourney.today') }}</span><span>{{ t('dailyJourney.progress', { done: completed, total: steps.length }) }}</span></div>
    <div class="landscape">
      <svg class="roads" viewBox="0 0 1000 260" preserveAspectRatio="none" aria-hidden="true">
        <path class="hill" d="M0 235 Q170 120 350 230 T700 230 T1000 200 V260 H0Z" />
        <path class="road-shadow" :d="road" /><path class="road" :d="road" />
        <path class="road-progress" :d="travelledRoad" />
      </svg>
      <div class="traveller" :style="{ left: `clamp(56px, ${point(currentIndex).x / 10}%, calc(100% - 56px))`, top: `${point(currentIndex).y / 2.6}%` }" aria-hidden="true">
        <LgLumi :pose="finished ? 'cup' : 'map'" :size="124" contained />
      </div>
      <ol>
        <li v-for="(step, index) in steps" :key="step.id" :class="[step.status, { current: step.id === currentId }]" :style="{ left: `${point(index).x / 10}%`, top: `${point(index).y / 2.6}%` }" :aria-current="step.id === currentId ? 'step' : undefined">
          <span class="stop" aria-hidden="true"><LgIcon :name="step.status === 'completed' ? 'check' : icons[step.mode]" :s="21" c="currentColor" /></span>
          <span class="stop-copy"><span class="title">{{ journeyStepTitle(step, t) }}</span><span class="status">{{ step.id === currentId ? t('dailyJourney.here') : statusLabel(step.status) }}</span></span>
        </li>
      </ol>
    </div>
  </nav>
</template>
<script setup lang="ts">
import { journeyStepTitle } from './dailyJourney/labels'
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import LgLumi from './linglow/LgLumi.vue'
import LgIcon from './linglow/LgIcon.vue'
import type { JourneyStep, StepStatus, JourneyMode } from './dailyJourney/types'
const { t } = useI18n()
const props = defineProps<{ steps: JourneyStep[]; currentId?: string; finished?: boolean }>()
const icons: Record<JourneyMode, string> = { words: 'book', reading: 'book-open', grammar: 'lightbulb', grammar_srs: 'refresh', verb_forms: 'ruler', picture: 'palette', conversation: 'message-circle' }
const completed = computed(() => props.steps.filter(s => s.status === 'completed').length)
const currentIndex = computed(() => props.finished ? Math.max(0, props.steps.length - 1) : Math.max(0, props.steps.findIndex(s => s.id === props.currentId)))
function point(index: number) { return { x: (index + .5) / Math.max(1, props.steps.length) * 1000, y: index % 2 ? 142 : 165 } }
function path(end: number) {
  const first = point(0)
  let value = `M${first.x} ${first.y}`
  for (let i = 1; i <= end; i++) {
    const a = point(i - 1), b = point(i), mid = (a.x + b.x) / 2
    value += ` C${mid} ${a.y} ${mid} ${b.y} ${b.x} ${b.y}`
  }
  return value
}
const road = computed(() => path(props.steps.length - 1))
const travelledRoad = computed(() => path(currentIndex.value))
function statusLabel(status: StepStatus) {
  return t('dailyJourney.statuses.' + status)
}
</script>
<style scoped>
.journey-map { position: relative; background: radial-gradient(ellipse at 15% 40%, color-mix(in srgb, var(--hoja) 11%, transparent), transparent 70%), var(--surface-2); border-bottom: 1px solid var(--border); border-radius: 24px 24px 0 0; }
.map-heading { display: flex; justify-content: space-between; gap: 12px; padding: 22px 28px 0; font-size: .68rem; letter-spacing: .12em; font-weight: 700; color: var(--subtext); }
.map-heading span:last-child { letter-spacing: 0; font-weight: 500; }
.landscape { position: relative; height: 260px; }
.roads { position: absolute; inset: 0; width: 100%; height: 100%; }
.hill { fill: var(--hoja); opacity: .06; }.road-shadow, .road, .road-progress { fill: none; stroke-linecap: round; }
.road-shadow { stroke: var(--border); stroke-width: 16; }.road { stroke: var(--card-bg); stroke-width: 12; }.road-progress { stroke: var(--hoja); stroke-width: 5; }
ol { list-style: none; padding: 0; margin: 0; }
li { position: absolute; width: 18%; transform: translate(-50%, -23px); display: flex; align-items: center; flex-direction: column; text-align: center; gap: 12px; }
.stop { display: grid; place-items: center; width: 46px; height: 46px; border: 1px solid var(--border); border-radius: 16px; background: var(--card-bg); color: var(--subtext); box-shadow: 0 4px 0 var(--border); }
.current .stop { color: var(--text); border-color: var(--hoja); background: var(--surface-3); box-shadow: 0 0 0 5px color-mix(in srgb, var(--hoja) 15%, transparent), 0 4px 0 var(--border-green); }
.completed .stop { color: var(--crema); background: var(--salvia); border-color: var(--salvia); }.skipped .stop { opacity: .45; }
.stop-copy { display: grid; gap: 4px; }.title { font-size: .8rem; font-weight: 650; line-height: 1.3; overflow-wrap: anywhere; }.status { font-size: .67rem; color: var(--subtext); }.current .status { color: var(--hoja); }
.traveller { position: absolute; z-index: 2; transform: translate(-50%, -108%); pointer-events: none; transition: left 550ms ease, top 550ms ease; }
@media (max-width: 540px) { .map-heading { padding: 18px 16px 0; font-size: .6rem; }.landscape { height: 232px; }.traveller :deep(img) { width: 100px !important; }.stop { width: 36px; height: 36px; border-radius: 12px; }li { transform: translate(-50%, -18px); gap: 10px; }.title { font-size: .64rem; }.status { font-size: .58rem; } }
@media (prefers-reduced-motion: reduce) { .traveller { transition: none; } }
</style>
