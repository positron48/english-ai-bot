<template>
  <div class="card completion-screen">
    <div class="completion-percentage">
      <div class="percentage-circle-wrapper">
        <svg class="percentage-circle" viewBox="0 0 120 120">
          <circle class="percentage-circle-bg" cx="60" cy="60" r="54" fill="none" stroke="var(--bg-secondary, rgba(0, 0, 0, 0.1))" stroke-width="8" />
          <circle class="percentage-circle-outline" cx="60" cy="60" r="54" fill="none" :stroke="percentageColor" stroke-width="8" stroke-opacity="0.2" />
          <circle
            class="percentage-circle-fill"
            cx="60"
            cy="60"
            r="54"
            fill="none"
            :stroke="percentageColor"
            stroke-width="8"
            stroke-linecap="round"
            :style="{ strokeDasharray: circumference, strokeDashoffset: percentageOffset }"
          />
        </svg>
        <div class="percentage-text">
          <span class="percentage-number">{{ animatedPercentage }}%</span>
          <span class="percentage-ratio">{{ correctCards }}/{{ totalCards }}</span>
        </div>

        <div v-if="accuracyPercentage > 90 && percentageAnimationComplete" class="celebration-container">
          <div class="fireworks">
            <div v-for="i in 20" :key="i" class="firework" :style="getFireworkStyle(i)">
              <div class="firework-core"></div>
              <div v-for="j in 12" :key="j" class="firework-particle" :data-particle-index="j" :style="{ '--angle': `${j * 30}deg` }"></div>
            </div>
          </div>
          <div class="confetti">
            <div v-for="i in 50" :key="i" class="confetti-piece" :style="getConfettiStyle(i)"></div>
          </div>
        </div>

        <div v-if="accuracyPercentage < 10 && percentageAnimationComplete" class="failure-container">
          <div class="failure-rain">
            <div v-for="i in 12" :key="i" class="failure-item" :style="getFailureItemStyle()">
              <span class="failure-emoji">💩</span>
            </div>
          </div>
        </div>
      </div>

      <div class="motivational-message" :class="messageClass">
        <p class="message-text">{{ motivationalMessage }}</p>
      </div>

      <div v-if="statsLoaded && availableForTraining !== null" class="completion-actions">
        <div class="remaining-cards-info">
          <span class="remaining-text">
            <span class="remaining-label">{{ t('training.available') }}</span>
            {{ availableForTraining }} {{ t('common.cards') || 'cards' }}
            <span v-if="estimatedTimeForRemaining">({{ estimatedTimeForRemaining }})</span>
          </span>
        </div>
        <button
          v-if="showContinueButton && availableForTraining > 0"
          type="button"
          class="btn btn-primary btn-continue"
          @click="$emit('continue')"
        >
          {{ continueLabel || t('training.continueTraining') || 'Continue Training' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAudio } from '../composables/useAudio'

const props = withDefaults(defineProps<{
  totalCards: number
  correctCards: number
  statsLoaded?: boolean
  availableForTraining?: number | null
  estimatedTimeForRemaining?: string | null
  continueLabel?: string
  showContinueButton?: boolean
  soundsEnabled?: boolean
  soundTheme?: string
}>(), {
  statsLoaded: false,
  availableForTraining: null,
  estimatedTimeForRemaining: null,
  continueLabel: '',
  showContinueButton: true,
  soundsEnabled: true,
  soundTheme: 'tick',
})

defineEmits<{ (e: 'continue'): void }>()

const { t, tm } = useI18n()
const { playVictory, playDefeat } = useAudio()

const animatedPercentage = ref(0)
const percentageAnimationComplete = ref(false)
const animatedPercentageOffset = ref(0)

const accuracyPercentage = computed(() => {
  if (!props.totalCards) return 0
  return Math.round((props.correctCards / props.totalCards) * 100)
})
const circumference = computed(() => 2 * Math.PI * 54)
const percentageOffset = computed(() => {
  if (!percentageAnimationComplete.value) return animatedPercentageOffset.value
  return circumference.value * (1 - animatedPercentage.value / 100)
})
const percentageColor = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return '#10b981'
  if (percent >= 70) return '#3b82f6'
  if (percent >= 50) return '#f59e0b'
  return '#ef4444'
})

function phraseList(key: string): string[] {
  const raw = tm(key) as unknown
  if (!Array.isArray(raw)) return []
  return raw.filter((x): x is string => typeof x === 'string' && x.length > 0)
}

const motivationalBuckets = computed(() => ({
  excellent: phraseList('trainingFeedback.motivational.excellent'),
  great: phraseList('trainingFeedback.motivational.great'),
  good: phraseList('trainingFeedback.motivational.good'),
  okay: phraseList('trainingFeedback.motivational.okay'),
  needsWork: phraseList('trainingFeedback.motivational.needsWork'),
  poor: phraseList('trainingFeedback.motivational.poor'),
}))

const generateMessageWeights = (count: number) => {
  if (count <= 0) return []
  if (count === 1) return [100]
  const weights: number[] = []
  for (let i = 0; i < count; i++) {
    const ratio = i / (count - 1)
    weights.push(30 * Math.pow(1 / 30, ratio))
  }
  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map((w) => (w * 100) / sum)
}
const getWeightedMessage = (messages: string[]) => {
  if (!messages.length) return ''
  if (messages.length === 1) return messages[0]
  const cumulative: number[] = []
  let sum = 0
  for (const weight of generateMessageWeights(messages.length)) {
    sum += weight
    cumulative.push(sum)
  }
  const random = Math.random() * 100
  for (let i = 0; i < cumulative.length; i++) {
    if (random <= cumulative[i]) return messages[i]
  }
  return messages[0]
}
const pickNonEmpty = (preferred: string[], fallback: string[]) => (preferred.length > 0 ? preferred : fallback)
const motivationalMessage = computed(() => {
  const percent = accuracyPercentage.value
  const b = motivationalBuckets.value
  const messages = percent >= 95 ? pickNonEmpty(b.excellent, b.great)
    : percent >= 90 ? pickNonEmpty(b.great, b.good)
      : percent >= 80 ? pickNonEmpty(b.good, b.okay)
        : percent >= 70 ? pickNonEmpty(b.okay, b.needsWork)
          : percent >= 50 ? pickNonEmpty(b.needsWork, b.poor)
            : pickNonEmpty(b.poor, b.needsWork)
  return getWeightedMessage(messages)
})
const messageClass = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return 'message-excellent'
  if (percent >= 70) return 'message-good'
  if (percent >= 50) return 'message-okay'
  return 'message-needs-improvement'
})

const getConfettiStyle = (index: number) => {
  const colors = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
  const angle = Math.random() * Math.PI * 2
  const distance = 140 + Math.random() * 120
  return {
    '--confetti-color': colors[index % colors.length],
    '--confetti-start-x': `${Math.cos(angle) * 90}px`,
    '--confetti-start-y': `${Math.sin(angle) * 90}px`,
    '--confetti-end-x': `${Math.cos(angle) * distance}px`,
    '--confetti-end-y': `${Math.sin(angle) * distance}px`,
    '--confetti-delay': `${Math.random() * 0.5}s`,
    '--confetti-duration': `${1.5 + Math.random()}s`,
  } as Record<string, string>
}
const getFailureItemStyle = () => ({
  '--start-x': `${(Math.random() - 0.5) * 400}px`,
  '--end-x': `${(Math.random() - 0.5) * 400}px`,
  '--delay': `${0.2 + Math.random() * 0.6}s`,
  '--duration': `${1.8 + Math.random() * 0.7}s`,
}) as Record<string, string>
const getFireworkStyle = () => ({
  '--firework-x': `${Math.cos(Math.random() * Math.PI * 2) * 90}px`,
  '--firework-y': `${Math.sin(Math.random() * Math.PI * 2) * 90}px`,
  '--delay': `${Math.random() * 1.2}s`,
  '--firework-size': `${0.8 + Math.random() * 0.4}`,
}) as Record<string, string>

watch(() => [props.totalCards, props.correctCards], () => {
  animatedPercentage.value = 0
  animatedPercentageOffset.value = circumference.value
  percentageAnimationComplete.value = false
  const target = accuracyPercentage.value
  const duration = 1500
  const startTime = Date.now()
  const animate = () => {
    const progress = Math.min((Date.now() - startTime) / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    animatedPercentage.value = Math.round(target * eased)
    animatedPercentageOffset.value = circumference.value * (1 - (eased * target) / 100)
    if (progress < 1) requestAnimationFrame(animate)
    else {
      animatedPercentage.value = target
      animatedPercentageOffset.value = circumference.value * (1 - target / 100)
      percentageAnimationComplete.value = true
    }
  }
  requestAnimationFrame(animate)
}, { immediate: true })

watch([() => percentageAnimationComplete.value, () => accuracyPercentage.value], ([complete, percentage]) => {
  if (!complete || !props.soundsEnabled) return
  if (percentage > 90) playVictory(props.soundTheme)
  else if (percentage < 10) playDefeat(props.soundTheme)
})
</script>

<style scoped>
.completion-screen { text-align: center; position: relative; overflow: hidden; padding: 40px 20px; }
.completion-percentage { margin: 20px 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 24px; }
.percentage-circle-wrapper { position: relative; width: 200px; height: 200px; flex-shrink: 0; }
.percentage-circle { width: 100%; height: 100%; transform: rotate(-90deg); }
.percentage-circle-bg { opacity: 0.2; }
.percentage-circle-fill { transition: stroke 0.3s ease; }
.percentage-text { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center; display: flex; flex-direction: column; align-items: center; gap: 4px; }
.percentage-number { font-size: 48px; font-weight: 700; color: var(--text-primary); line-height: 1; }
.percentage-ratio { font-size: 20px; font-weight: 500; color: var(--text-secondary); opacity: 0.85; line-height: 1; }
.motivational-message { padding: 18px 28px; border-radius: 12px; max-width: 600px; width: 100%; text-align: center; }
.message-text { margin: 0; font-size: 18px; font-weight: 500; line-height: 1.5; }
.message-excellent { background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(5, 150, 105, 0.15) 100%); color: var(--color-success, #10b981); border: 2px solid rgba(16, 185, 129, 0.3); }
.message-good { background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(37, 99, 235, 0.15) 100%); color: #3b82f6; border: 2px solid rgba(59, 130, 246, 0.3); }
.message-okay { background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(217, 119, 6, 0.15) 100%); color: #f59e0b; border: 2px solid rgba(245, 158, 11, 0.3); }
.message-needs-improvement { background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(220, 38, 38, 0.15) 100%); color: #ef4444; border: 2px solid rgba(239, 68, 68, 0.3); }
.completion-actions { width: 100%; max-width: 420px; display: flex; flex-direction: column; gap: 12px; }
.remaining-cards-info { padding: 10px 14px; border-radius: 10px; background: var(--bg-secondary); color: var(--text-secondary); }
.remaining-label { font-weight: 600; margin-right: 6px; }
.btn-continue { width: 100%; }
.celebration-container, .failure-container { position: absolute; inset: 0; pointer-events: none; }
.fireworks, .confetti, .failure-rain { position: absolute; inset: 0; }
.firework, .confetti-piece, .failure-item { position: absolute; top: 50%; left: 50%; }
.firework { width: 6px; height: 6px; border-radius: 50%; transform: translate(var(--firework-x), var(--firework-y)) scale(var(--firework-size)); animation: firework-explode 2s ease-out var(--delay) forwards; }
.firework-core { width: 6px; height: 6px; border-radius: 50%; background: radial-gradient(circle, #fff 0%, #ffd700 50%, transparent 100%); animation: firework-core-pulse 0.35s ease-out var(--delay) forwards; }
.firework-particle { width: 8px; height: 8px; border-radius: 50%; background: #10b981; box-shadow: 0 0 6px #10b981; transform: translate(-50%, -50%); animation: particle-fly 2s ease-out var(--delay) forwards; }
.confetti-piece { width: 8px; height: 12px; background: var(--confetti-color); transform: translate(var(--confetti-start-x), var(--confetti-start-y)); animation: confetti-fly var(--confetti-duration) ease-out var(--confetti-delay) forwards; }
.failure-item { font-size: 28px; transform: translate(var(--start-x), -200px); animation: failure-fall var(--duration) ease-in var(--delay) forwards; }
.failure-emoji { display: inline-block; }
@keyframes firework-explode { from { opacity: 1; } to { opacity: 0; transform: translate(var(--firework-x), var(--firework-y)) scale(0); } }
@keyframes firework-core-pulse { from { opacity: 1; transform: scale(0); } to { opacity: 0; transform: scale(3); } }
@keyframes particle-fly { from { opacity: 1; } to { opacity: 0; transform: translate(-50%, -50%) translateX(var(--offset-x, 0)) translateY(var(--offset-y, 0)); } }
@keyframes confetti-fly { from { opacity: 1; } to { opacity: 0; transform: translate(var(--confetti-end-x), var(--confetti-end-y)) rotate(480deg); } }
@keyframes failure-fall { from { opacity: 0; } 20% { opacity: 1; } to { opacity: 0; transform: translate(var(--end-x), 260px) rotate(360deg); } }
.firework-particle[data-particle-index="1"] { --offset-x: 100px; --offset-y: 0px; }
.firework-particle[data-particle-index="2"] { --offset-x: 86px; --offset-y: -50px; }
.firework-particle[data-particle-index="3"] { --offset-x: 50px; --offset-y: -86px; }
.firework-particle[data-particle-index="4"] { --offset-x: 0px; --offset-y: -100px; }
.firework-particle[data-particle-index="5"] { --offset-x: -50px; --offset-y: -86px; }
.firework-particle[data-particle-index="6"] { --offset-x: -86px; --offset-y: -50px; }
.firework-particle[data-particle-index="7"] { --offset-x: -100px; --offset-y: 0px; }
.firework-particle[data-particle-index="8"] { --offset-x: -86px; --offset-y: 50px; }
.firework-particle[data-particle-index="9"] { --offset-x: -50px; --offset-y: 86px; }
.firework-particle[data-particle-index="10"] { --offset-x: 0px; --offset-y: 100px; }
.firework-particle[data-particle-index="11"] { --offset-x: 50px; --offset-y: 86px; }
.firework-particle[data-particle-index="12"] { --offset-x: 86px; --offset-y: 50px; }
</style>
