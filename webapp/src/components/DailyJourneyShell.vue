<template>
  <section class="journey-shell" :aria-label="t('dailyJourney.title')" :aria-busy="busy">
    <DailyJourneyMap v-if="journey" :steps="journey.steps" :current-id="current?.id" :finished="journey.status === 'completed'" />
    <div v-if="error" class="error" role="alert"><p>{{ error }}</p><button :disabled="busy" @click="$emit('retry')">{{ t('common.retry') }}</button><button :disabled="busy" @click="$emit('reload')">{{ t('dailyJourney.reload') }}</button></div>
    <div v-if="!journey" class="intro"><LgLumi :size="160" contained pose="map" /><h2 ref="heading" tabindex="-1">{{ t('dailyJourney.introTitle') }}</h2><p>{{ t('dailyJourney.intro') }}</p><button :disabled="busy || !!error" @click="$emit('start')">{{ busy ? t('common.loading') : t('dailyJourney.start') }}</button></div>
    <DailyJourneySummary v-else-if="journey.status !== 'active'" :journey="journey" />
    <Transition v-else name="step" mode="out-in" @after-enter="focusHeading">
      <article v-if="current" :key="current.id" class="stop-detail">
        <div class="stop-copy">
          <p class="eyebrow">{{ t('dailyJourney.stop', { position: journey.steps.indexOf(current) + 1, total: journey.steps.length }) }}</p>
          <h2 ref="heading" tabindex="-1">{{ journeyStepTitle(current, t) }}</h2>
          <p>{{ journeyStepGoal(current, t) }}</p>
          <p v-if="current.status === 'active'" class="resume">{{ t('dailyJourney.resume') }}</p>
          <button :disabled="blocked" @click="$emit('open-step', current)">{{ current.status === 'active' ? t('dailyJourney.resumeTask') : t('dailyJourney.open') }} <span aria-hidden="true">→</span></button>
        </div>
        <div class="stop-art" aria-hidden="true"><LgLumi :size="180" contained :pose="pose" /></div>
        <footer><span>{{ current.required ? t('dailyJourney.required') : canSkip ? t('dailyJourney.skipAvailable') : t('dailyJourney.skipUsed') }}</span><button v-if="canSkip" class="secondary" :disabled="blocked" @click="$emit('action', 'skip', {})">{{ t('dailyJourney.skip') }}</button></footer>
      </article>
    </Transition>
    <span class="sr-only" aria-live="polite">{{ announcement }}</span>
  </section>
</template>
<script setup lang="ts">
import { journeyStepTitle, journeyStepGoal } from './dailyJourney/labels'
import { useI18n } from 'vue-i18n'
import { computed, nextTick, ref, watch } from 'vue'
import DailyJourneyMap from './DailyJourneyMap.vue'
import DailyJourneySummary from './DailyJourneySummary.vue'
import LgLumi from './linglow/LgLumi.vue'
import { currentJourneyStep, journeySkipAvailable, type DailyJourney, type JourneyActionType, type JourneyStep } from './dailyJourney/types'
const { t } = useI18n()
const props = defineProps<{ journey: DailyJourney | null; busy: boolean; error?: string; hasPendingAction?: boolean }>()
defineEmits<{ start: []; retry: []; reload: []; action: [action: JourneyActionType, payload: Record<string, unknown>]; 'open-step': [step: JourneyStep] }>()
const heading = ref<HTMLElement | null>(null)
const current = computed(() => currentJourneyStep(props.journey))
const canSkip = computed(() => current.value && !current.value.required && journeySkipAvailable(props.journey))
const blocked = computed(() => props.busy || props.hasPendingAction)
const pose = computed(() => current.value?.mode === 'reading' ? 'book' : current.value?.mode === 'words' ? 'word-card' : 'teacher')
const announcement = computed(() => props.journey?.status === 'completed' ? t('dailyJourney.finished') : current.value ? t('dailyJourney.current', { title: journeyStepTitle(current.value, t) }) : '')
function focusHeading() { heading.value?.focus() }
watch(() => current.value?.id, async () => { await nextTick(); focusHeading() })
</script>
<style scoped>
.journey-shell { border: 1px solid var(--border); border-radius: 24px; background: var(--card-bg); box-shadow: var(--shadow-soft); }
.intro { display: grid; justify-items: start; gap: 20px; padding: 32px 40px; }
.stop-detail { display: grid; grid-template-columns: 1fr 180px; gap: 28px; align-items: center; padding: 32px 40px; }
.stop-copy { display: grid; gap: 16px; justify-items: start; }.stop-art { display: grid; place-items: center; }
h2, p { margin: 0; }h2 { font-size: 1.85rem; font-weight: 500; }h2:focus { outline: none; }p { line-height: 1.6; font-size: .9rem; color: var(--subtext); }.eyebrow { font-size: .75rem; }.resume { color: var(--hoja); font-size: .8rem; }
button { display: inline-flex; align-items: center; gap: 20px; font: inherit; border: 0; border-radius: 12px; padding: 13px 18px; color: var(--crema); background: var(--btn-gradient); box-shadow: var(--btn-shadow); cursor: pointer; font-size: .85rem; font-weight: 600; }button:disabled { opacity: .55; cursor: default; }button:focus-visible { outline: 3px solid var(--text); outline-offset: 3px; }.secondary { color: var(--text); border: 1px solid var(--border); background: transparent; box-shadow: none; }
footer { grid-column: 1 / -1; display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding-top: 20px; border-top: 1px solid var(--border); font-size: .75rem; color: var(--subtext); }
.error { padding: 20px; border: 1px solid var(--error); margin: 16px; border-radius: 12px; }.error button { margin: 12px 12px 0 0; }
.step-enter-active, .step-leave-active { transition: opacity 180ms ease, transform 180ms ease; }.step-enter-from, .step-leave-to { opacity: 0; transform: translateY(8px); }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }
@media (max-width: 540px) { .intro, .stop-detail { padding: 24px 20px; }.stop-detail { grid-template-columns: 1fr 90px; gap: 20px 0; }.stop-art :deep(img) { width: 110px !important; }h2 { font-size: 1.6rem; }button { gap: 8px; font-size: .78rem; padding: 12px; } }
@media (prefers-reduced-motion: reduce) { .step-enter-active, .step-leave-active { transition: none; } }
</style>
