<template>
  <main class="preview">
    <header><div><span class="page-kicker">LINGLOW · ЕЖЕДНЕВНАЯ ПРАКТИКА</span><h1>Маршрут дня<span class="title-dot">.</span></h1></div><button v-if="opened" class="back-button" @click="opened = false">← К обзору</button></header>
    <p class="preview-note">Предпросмотр · ответы не влияют на учебный прогресс</p>
    <DailyJourneyCard v-if="!opened" :journey="journey" :busy="busy" @open="open" />
    <DailyJourneyShell v-if="opened" :journey="journey" :busy="busy" :error="error" :has-pending-action="hasPendingAction" @start="start" @retry="retry" @reload="reload" @action="act" @open-step="openStep" />
    <p v-if="launchNotice" role="status" class="launch-notice">{{ launchNotice }}</p>
    <details class="preview-tools"><summary>Настройки макета</summary>
    <div class="controls">
      <label>Сценарий <select v-model="scenario"><option value="full">Пять остановок</option><option value="add_words">Добор слов</option><option value="short">Короткий маршрут</option><option value="verbs">Формы глаголов</option><option value="srs">SRS-грамматика</option><option value="unavailable">Нет контента</option></select></label>
      <label>Курс <select v-model="course"><option value="en_ru">Английский</option><option value="es_ru">Испанский</option></select></label>
      <button @click="reset">Сбросить макет</button>
      <button v-if="isAuthenticated" @click="logout">Выйти из тестовой сессии</button>
    </div>
    </details>
  </main>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import DailyJourneyCard from '../components/DailyJourneyCard.vue'
import DailyJourneyShell from '../components/DailyJourneyShell.vue'
import { useAuth } from '../composables/useAuth'
import { useDailyJourney } from '../composables/useDailyJourney'
import { createFixtureTransport, type FixtureScenario } from './dailyJourneyFixture'
import type { JourneyStep } from '../components/dailyJourney/types'
const { isAuthenticated, logout } = useAuth()
const scenario = ref<FixtureScenario>('full')
const course = ref('en_ru')
const scope = computed(() => `${course.value}:${scenario.value}`)
const opened = ref(false)
const launchNotice = ref('')
const { journey, busy, error, hasPendingAction, start, retry, reload, act } = useDailyJourney(scope, key => createFixtureTransport(key, course.value, scenario.value, localStorage))
watch(scope, () => { launchNotice.value = '' })
async function open() { opened.value = true; if (!journey.value) await start() }
async function openStep(step: JourneyStep) {
  if (step.status === 'pending') await act('start')
  if (error.value) return
  launchNotice.value = `Здесь откроется обычный экран «${step.title}». В макете нет предметной сессии: переход с фиктивным идентификатором отключён.`
}
function reset() {
  localStorage.removeItem(`linglow:daily-journey-preview:v1:${scope.value}`)
  launchNotice.value = ''
  void reload()
}
</script>
<style scoped>
.preview { max-width: 900px; margin: 0 auto; padding: 44px 24px 60px; color: var(--text-primary, #223830); display: grid; gap: 24px; }
header { display: flex; justify-content: space-between; align-items: center; gap: 16px; } h1 { font-size: clamp(2rem, 5vw, 2.9rem); font-weight: 500; letter-spacing: -.035em; margin-top: 10px; } .title-dot { color: var(--hoja); } .page-kicker { font-size: .62rem; font-weight: 650; letter-spacing: .15em; color: var(--subtext); } .preview-note { color: var(--text-muted); font-size: .72rem; margin-top: -14px; } .preview-tools { border-top: 1px solid var(--border); padding-top: 20px; color: var(--subtext); font-size: .8rem; } summary { cursor: pointer; } .preview-tools .controls { padding-top: 20px; } .back-button { border: 0; background: transparent; font-size: .8rem; }.controls { display: flex; flex-wrap: wrap; gap: 16px; align-items: end; }
label { display: grid; gap: 8px; } select, input, button { font: inherit; padding: 10px 12px; border-radius: 10px; border: 1px solid var(--border, #d4dce4); background: var(--card-bg, white); color: inherit; }
button { cursor: pointer; } button:disabled { opacity: .5; cursor: default; } .launch-notice { padding: 16px; border: 1px solid var(--border); border-radius: 12px; font-size: .8rem; line-height: 1.6; color: var(--subtext); } @media (max-width: 540px) { .preview { padding: 28px 14px 40px; gap: 20px; } .page-kicker { font-size: .52rem; } .back-button { padding: 8px 0; } }
</style>
