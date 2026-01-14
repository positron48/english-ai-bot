<template>
  <div class="training">
    <h1 v-if="!(sessionActive && currentCard)" class="training-title">Training</h1>
    
    <div v-if="sessionComplete && !sessionActive" class="card completion-screen">
      <!-- Animated percentage display -->
      <div class="completion-percentage">
        <div class="percentage-circle-wrapper">
          <svg class="percentage-circle" viewBox="0 0 120 120">
            <circle
              class="percentage-circle-bg"
              cx="60"
              cy="60"
              r="54"
              fill="none"
              stroke="var(--bg-secondary, rgba(0, 0, 0, 0.1))"
              stroke-width="8"
            />
            <circle
              class="percentage-circle-outline"
              cx="60"
              cy="60"
              r="54"
              fill="none"
              :stroke="percentageColor"
              stroke-width="8"
              stroke-opacity="0.2"
            />
            <circle
              class="percentage-circle-fill"
              cx="60"
              cy="60"
              r="54"
              fill="none"
              :stroke="percentageColor"
              stroke-width="8"
              stroke-linecap="round"
              :style="{
                strokeDasharray: circumference,
                strokeDashoffset: percentageOffset
              }"
            />
          </svg>
          <div class="percentage-text">
            <span class="percentage-number">{{ animatedPercentage }}%</span>
            <span class="percentage-ratio">{{ trainingStats.correctCards }}/{{ trainingStats.totalCards }}</span>
          </div>
          
          <!-- Fireworks/Confetti for >90% -->
          <div v-if="accuracyPercentage > 90 && percentageAnimationComplete" class="celebration-container">
            <div class="fireworks">
              <div v-for="i in 20" :key="i" class="firework" :style="getFireworkStyle(i)">
                <div class="firework-core"></div>
                <div class="firework-particle" v-for="j in 12" :key="j" :data-particle-index="j" :style="{ '--angle': (j * 30) + 'deg' }"></div>
              </div>
            </div>
            <div class="confetti">
              <div v-for="i in 50" :key="i" class="confetti-piece" :style="getConfettiStyle(i)"></div>
            </div>
          </div>
          
          <!-- Failure animation for <10% -->
          <div v-if="accuracyPercentage < 10 && percentageAnimationComplete" class="failure-container">
            <div class="failure-rain">
              <div v-for="i in 12" :key="i" class="failure-item" :style="getFailureItemStyle(i)">
                <span class="failure-emoji">💩</span>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Motivational message -->
        <div class="motivational-message" :class="messageClass">
          <p class="message-text">{{ motivationalMessage }}</p>
        </div>
        
        <!-- Compact info and continue button -->
        <div class="completion-actions" v-if="statsLoaded">
          <div class="remaining-cards-info">
            <span class="remaining-text">
              <span class="remaining-label">Available:</span>
              {{ stats.availableForTraining }} cards
              <span v-if="estimatedTimeForRemaining">({{ estimatedTimeForRemaining }})</span>
            </span>
          </div>
          <button 
            v-if="stats.availableForTraining > 0" 
            @click="startTraining" 
            class="btn btn-primary btn-continue"
          >
            Continue Training
          </button>
        </div>
      </div>
    </div>

    <div v-if="!sessionActive && !loading && !sessionComplete" class="card start-screen">
      <div class="start-screen-content">
        <div class="start-screen-stats" v-if="statsLoaded">
          <div class="start-stat-item">
            <span class="start-stat-label">Available for Training</span>
            <span class="start-stat-value">
              {{ stats.availableForTraining }} cards
              <span v-if="estimatedTime">({{ estimatedTime }})</span>
            </span>
          </div>
        </div>
        
        <!-- Upcoming cards chart -->
        <div v-if="statsLoaded && upcomingCardsLoaded" class="upcoming-cards-chart">
          <div class="chart-header">
            <h3 class="chart-title">Upcoming Cards This Week</h3>
            <div class="chart-subtitle">Cards scheduled to appear in the next 7 days</div>
          </div>
          <div class="chart-container">
            <canvas ref="upcomingChartCanvas"></canvas>
          </div>
        </div>
        
        <button v-if="statsLoaded && stats.availableForTraining > 0" @click="startTraining" class="btn btn-primary btn-start">
          Start Training
        </button>
        <p v-if="statsLoaded && stats.availableForTraining === 0" class="no-cards-message">
          No cards available for training. Add some words to your vocabulary first!
        </p>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading...</div>

    <!-- Network error notification -->
    <div v-if="networkError" class="network-error-notification">
      <div class="network-error-content">
        <Icon name="warning" class="network-error-icon" />
        <div class="network-error-text">
          <div class="network-error-title">Проблемы с сетью</div>
          <div class="network-error-message">
            {{ networkErrorRetrying ? `Попытка восстановить связь (${networkErrorAttempt}/${networkErrorMaxAttempts})...` : 'Не удалось подключиться к серверу' }}
          </div>
        </div>
      </div>
    </div>

    <div v-if="sessionActive && currentCard" class="card">
      <div class="training-progress" v-if="cardIndex > 0 && totalCards > 0">
        <p>Card {{ cardIndex }} of {{ totalCards }}</p>
      </div>

      <div class="question" v-html="processedQuestion"></div>

      <div v-if="optionsShown" class="options">
        <button
          v-for="(option, index) in options"
          :key="index"
          @click="!feedback && submitAnswer(index)"
          :class="[
            'btn',
            'option-btn',
            {
              'option-correct': feedback && option === feedback.correct_answer,
              'option-incorrect': feedback && !feedback.is_correct && option === feedback.chosen_option,
              'option-disabled': !!feedback
            }
          ]"
          :disabled="answering || !!feedback"
        >
          <span class="option-number">{{ index + 1 }}</span>
          <span class="option-text">{{ option }}</span>
        </button>
      </div>

      <div v-if="feedback" class="feedback-section">
        <div v-if="feedback.is_correct" class="feedback-badge feedback-success">
          <!-- Success particles -->
          <div class="success-particles">
            <div v-for="i in 12" :key="i" class="success-particle" :style="getSuccessParticleStyle(i)"></div>
          </div>
          <span class="feedback-icon">✓</span>
          <span class="feedback-text">{{ currentEncouragingPhrase }}</span>
        </div>
        
        <!-- For incorrect answers: show hint, example, then notification with circular progress -->
        <template v-if="!feedback.is_correct">
          <div v-if="feedback.hint" class="hint">{{ feedback.hint }}</div>
          <div v-if="feedback.example" class="example">{{ feedback.example }}</div>
          <div class="feedback-badge feedback-error">
            <div 
              v-if="waitingDelay" 
              class="error-progress-wrapper"
              @mousedown="handleTimerMouseDown"
              @mouseup="handleTimerMouseUp"
              @mouseleave="handleTimerMouseLeave"
              @touchstart="handleTimerMouseDown"
              @touchend="handleTimerMouseUp"
              @touchcancel="handleTimerMouseLeave"
            >
              <svg class="error-progress-ring" width="40" height="40">
                <circle
                  class="error-progress-circle-bg"
                  stroke="rgba(255, 255, 255, 0.2)"
                  stroke-width="2.5"
                  fill="transparent"
                  r="16"
                  cx="20"
                  cy="20"
                />
                <circle
                  class="error-progress-circle"
                  stroke="white"
                  stroke-width="2.5"
                  fill="transparent"
                  r="16"
                  cx="20"
                  cy="20"
                  :style="{ 
                    strokeDasharray: errorCircumference, 
                    strokeDashoffset: errorProgressOffset 
                  }"
                />
              </svg>
              <svg class="error-icon-svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </div>
            <svg v-else class="error-icon-svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
            <span class="feedback-text">{{ currentDisappointingPhrase }}</span>
          </div>
        </template>
        
        <!-- For correct answers: show example after notification -->
        <div v-if="feedback.is_correct && feedback.example" class="example">{{ feedback.example }}</div>
        
        <!-- Circular progress for correct answers delay (if any) -->
        <div 
          v-if="waitingDelay && feedback.is_correct" 
          class="waiting-progress"
          @mousedown="handleTimerMouseDown"
          @mouseup="handleTimerMouseUp"
          @mouseleave="handleTimerMouseLeave"
          @touchstart="handleTimerMouseDown"
          @touchend="handleTimerMouseUp"
          @touchcancel="handleTimerMouseLeave"
        >
          <div class="circular-progress">
            <svg class="progress-ring" width="80" height="80">
              <circle
                class="progress-ring-circle-bg"
                stroke="var(--bg-secondary, rgba(0, 0, 0, 0.1))"
                stroke-width="6"
                fill="transparent"
                r="34"
                cx="40"
                cy="40"
              />
              <circle
                class="progress-ring-circle"
                stroke="var(--color-primary)"
                stroke-width="6"
                fill="transparent"
                r="34"
                cx="40"
                cy="40"
                :style="{ strokeDasharray: delayCircumference, strokeDashoffset: strokeDashoffset }"
              />
            </svg>
            <div class="progress-text">{{ delaySeconds }}</div>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { apiClient } from '../api/client'
import { showAlert } from '../composables/useDialog'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'
import { Chart, registerables } from 'chart.js'
import Icon from '../components/Icon.vue'

Chart.register(...registerables)

interface Card {
  question: string
  card_index: number
  total_cards: number
  session_id: number
  user_card_id: number
  delay_ms: number
}

interface OptionsResponse {
  options: string[]
  user_card_id: number
}

interface Feedback {
  is_correct: boolean
  chosen_option: string
  correct_answer: string
  hint?: string
  example?: string
  delay_seconds?: number
}

const sessionActive = ref(false)
const loading = ref(false)
const currentCard = ref<Card | null>(null)
const optionsShown = ref(false)
const options = ref<string[]>([])
const feedback = ref<Feedback | null>(null)
const currentEncouragingPhrase = ref('Great job!')
const currentDisappointingPhrase = ref('Not quite')
const answering = ref(false)
const waitingDelay = ref(false)
const delaySeconds = ref(0)
const initialDelaySeconds = ref(0)
const remainingMs = ref(0)
const initialDelayMs = ref(0)
const sessionComplete = ref(false)
const cardsCompleted = ref(0)
const trainingStats = ref({
  totalCards: 0,
  correctCards: 0
})
const stats = ref({
  dueCount: 0,
  totalCards: 0,
  availableForTraining: 0
})
const statsLoaded = ref(false)
const networkError = ref(false)
const networkErrorRetrying = ref(false)
const networkErrorAttempt = ref(0)
const networkErrorMaxAttempts = ref(3)
const animatedPercentage = ref(0)
const percentageAnimationComplete = ref(false)
const upcomingChartCanvas = ref<HTMLCanvasElement | null>(null)
const upcomingCardsLoaded = ref(false)
const upcomingCardsData = ref<Record<string, { date: string; label: string; count: number }>>({})
let upcomingChartInstance: Chart | null = null

// Settings
const { settings } = useSettings()
const { playSuccess, playFail, playVictory, playDefeat } = useAudio()

const estimatedTime = computed(() => {
  const cards = stats.value.availableForTraining
  if (cards === 0) return null
  
  // Average 15 seconds per card (same as notification service)
  const avgSecondsPerCard = 15
  const totalSeconds = cards * avgSecondsPerCard
  const minutes = Math.floor(totalSeconds / 60)
  
  if (minutes < 1) {
    return 'less than 1 minute'
  } else if (minutes === 1) {
    return '~1 minute'
  } else {
    return `~${minutes} minutes`
  }
})

const estimatedTimeForRemaining = computed(() => {
  const cards = stats.value.availableForTraining
  if (cards === 0) return null
  
  // Average 15 seconds per card (same as notification service)
  const avgSecondsPerCard = 15
  const totalSeconds = cards * avgSecondsPerCard
  const minutes = Math.floor(totalSeconds / 60)
  
  if (minutes < 1) {
    return '~1 min'
  } else if (minutes === 1) {
    return '~1 min'
  } else {
    return `~${minutes} min`
  }
})

// Calculate accuracy percentage
const accuracyPercentage = computed(() => {
  if (trainingStats.value.totalCards === 0) return 0
  return Math.round((trainingStats.value.correctCards / trainingStats.value.totalCards) * 100)
})

// Percentage circle calculations
const circumference = computed(() => 2 * Math.PI * 54)
const animatedPercentageOffset = ref(0)
const percentageOffset = computed(() => {
  // Use animated offset during animation, computed offset after
  if (!percentageAnimationComplete.value) {
    return animatedPercentageOffset.value
  }
  const progress = animatedPercentage.value / 100
  return circumference.value * (1 - progress)
})

// Color based on percentage
const percentageColor = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return '#10b981' // green
  if (percent >= 70) return '#3b82f6' // blue
  if (percent >= 50) return '#f59e0b' // orange
  return '#ef4444' // red
})

// Motivational messages with multiple variants
const motivationalMessages = {
  excellent: [
    'Outstanding work! You\'re a true master!',
    'Perfect score! You\'re absolutely brilliant!',
    'Incredible performance! You\'ve mastered this!',
    'Flawless execution! You\'re a champion!',
    'Exceptional results! You\'re unstoppable!',
    'Perfect mastery! You\'re a true expert!',
    'Outstanding achievement! You\'re incredible!',
    'Brilliant work! You\'re a natural!',
    'Perfect performance! You\'re amazing!',
    'Exceptional skill! You\'re a pro!'
  ],
  great: [
    'Excellent! You\'re doing amazing! Keep it up!',
    'Fantastic work! You\'re on fire!',
    'Wonderful progress! You\'re crushing it!',
    'Superb results! You\'re doing great!',
    'Impressive performance! Keep going!',
    'Terrific work! You\'re making great strides!',
    'Awesome job! You\'re on the right path!',
    'Brilliant effort! You\'re improving fast!',
    'Splendid work! You\'re doing fantastic!',
    'Marvelous results! You\'re getting better!'
  ],
  good: [
    'Great job! You\'re making excellent progress!',
    'Well done! You\'re improving steadily!',
    'Nice work! You\'re on the right track!',
    'Good progress! Keep up the momentum!',
    'Solid effort! You\'re doing well!',
    'Decent results! You\'re moving forward!',
    'Not bad at all! You\'re getting there!',
    'Good going! You\'re making progress!',
    'Keep it up! You\'re doing fine!',
    'Steady progress! You\'re on track!'
  ],
  okay: [
    'Good work! You\'re on the right track!',
    'Not bad! Keep practicing and you\'ll improve!',
    'You\'re getting there! Don\'t give up!',
    'Keep going! Practice makes perfect!',
    'You\'re improving! Stay focused!',
    'Good effort! Every attempt counts!',
    'Keep trying! You\'re making progress!',
    'Don\'t stop! You\'re learning!',
    'Stay motivated! You can do better!',
    'Keep pushing! Improvement is coming!'
  ],
  needsWork: [
    'Keep practicing! Every mistake is a learning opportunity!',
    'Don\'t give up! Every attempt makes you stronger!',
    'Stay focused! Practice will improve your results!',
    'Keep trying! You\'re building your skills!',
    'Don\'t lose heart! Learning takes time!',
    'Stay persistent! You\'ll get better!',
    'Keep going! Every session helps!',
    'Don\'t quit! Progress comes with practice!',
    'Stay determined! You can improve!',
    'Keep learning! Mistakes teach us!'
  ],
  poor: [
    'Time to review! Let\'s go over the basics again.',
    'Back to basics! Review will help you improve.',
    'Let\'s practice more! Focus on the fundamentals.',
    'Review time! Understanding comes with repetition.',
    'Study harder! The basics are important.',
    'More practice needed! Review the material.',
    'Focus on learning! Review helps retention.',
    'Time to study! Practice the fundamentals.',
    'Review and practice! You\'ll improve.',
    'Back to studying! Master the basics first.'
  ]
}

// Generate weights for motivational messages (first 30%, last 1%)
const generateMessageWeights = (count: number) => {
  const weights: number[] = []
  const maxWeight = 30 // 30%
  const minWeight = 1 // 1%
  
  // Exponential decay: weight = maxWeight * (minWeight/maxWeight)^((i)/(n-1))
  for (let i = 0; i < count; i++) {
    const ratio = i / (count - 1)
    const weight = maxWeight * Math.pow(minWeight / maxWeight, ratio)
    weights.push(weight)
  }
  
  // Normalize to sum to 100
  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map(w => w * 100 / sum)
}

const motivationalMessage = computed(() => {
  const percent = accuracyPercentage.value
  let messages: string[]
  
  if (percent >= 95) {
    messages = motivationalMessages.excellent
  } else if (percent >= 90) {
    messages = motivationalMessages.great
  } else if (percent >= 80) {
    messages = motivationalMessages.good
  } else if (percent >= 70) {
    messages = motivationalMessages.okay
  } else if (percent >= 50) {
    messages = motivationalMessages.needsWork
  } else if (percent >= 10) {
    messages = motivationalMessages.poor
  } else {
    messages = motivationalMessages.poor
  }
  
  // Weighted random selection (first 30%, last 1%)
  return getWeightedMessage(messages)
})

const messageClass = computed(() => {
  const percent = accuracyPercentage.value
  if (percent >= 90) return 'message-excellent'
  if (percent >= 70) return 'message-good'
  if (percent >= 50) return 'message-okay'
  return 'message-needs-improvement'
})

// Confetti styles - start from random points on circle edge
const getConfettiStyle = (index: number) => {
  const colors = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
  const color = colors[index % colors.length]
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: viewBox 120x120, wrapper 200x200px, so scale = 200/120 = 1.67
  // Radius in viewBox = 54, so real radius = 54 * 1.67 ≈ 90px
  const circleRadius = 90
  // Start position on circle edge
  const startX = Math.cos(startAngleRad) * circleRadius
  const startY = Math.sin(startAngleRad) * circleRadius
  // Random direction outward
  const directionAngle = startAngle + (Math.random() - 0.5) * 40 // ±20 degrees variation
  const directionAngleRad = (directionAngle * Math.PI) / 180
  // Distance to travel
  const distance = 150 + Math.random() * 100 // 150-250px
  const endX = Math.cos(directionAngleRad) * distance
  const endY = Math.sin(directionAngleRad) * distance
  // Random delay and duration
  const delay = Math.random() * 0.5
  const duration = 1.5 + Math.random() * 1
  return {
    '--confetti-color': color,
    '--confetti-start-x': `${startX}px`,
    '--confetti-start-y': `${startY}px`,
    '--confetti-end-x': `${endX}px`,
    '--confetti-end-y': `${endY}px`,
    '--confetti-delay': `${delay}s`,
    '--confetti-duration': `${duration}s`
  }
}

// Failure item styles - falling poop from top
const getFailureItemStyle = (index: number) => {
  // Start from random position at top (well above visible area)
  const startX = (Math.random() - 0.5) * 400 // -200 to 200px from center
  const startY = -200 - Math.random() * 150 // Start well above screen
  
  // End position - fall down and slightly outward
  const endX = startX + (Math.random() - 0.5) * 100 // Some horizontal drift
  const endY = 250 + Math.random() * 150 // Fall well below center
  
  // Rotation and scale
  const rotation = (Math.random() - 0.5) * 720 // Random rotation
  const scale = 0.6 + Math.random() * 0.4 // 0.6-1.0
  
  // Timing - staggered fall, ensure they start invisible and well above
  const delay = 0.2 + Math.random() * 0.6 // Start after 0.2s to avoid center flash
  const duration = 1.8 + Math.random() * 0.7 // 1.8-2.5s
  
  return {
    '--start-x': `${startX}px`,
    '--start-y': `${startY}px`,
    '--end-x': `${endX}px`,
    '--end-y': `${endY}px`,
    '--rotation': `${rotation}deg`,
    '--scale': scale,
    '--delay': `${delay}s`,
    '--duration': `${duration}s`
  }
}

// Success particle styles - explode outward from center
const getSuccessParticleStyle = (index: number) => {
  // Random angle (0-360 degrees)
  const angle = (index * 30) + Math.random() * 15 // Spread evenly with some randomness
  const angleRad = (angle * Math.PI) / 180
  // Distance to travel
  const distance = 60 + Math.random() * 40 // 60-100px
  const endX = Math.cos(angleRad) * distance
  const endY = Math.sin(angleRad) * distance
  // Random size
  const size = 4 + Math.random() * 4 // 4-8px
  // Random delay
  const delay = Math.random() * 0.2
  return {
    '--particle-end-x': `${endX}px`,
    '--particle-end-y': `${endY}px`,
    '--particle-size': `${size}px`,
    '--particle-delay': `${delay}s`
  }
}

// Firework styles - start from random points on circle edge
const getFireworkStyle = (index: number) => {
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: viewBox 120x120, wrapper 200x200px, so scale = 200/120 = 1.67
  // Radius in viewBox = 54, so real radius = 54 * 1.67 ≈ 90px
  const circleRadius = 90
  // Start position on circle edge
  const startX = Math.cos(startAngleRad) * circleRadius
  const startY = Math.sin(startAngleRad) * circleRadius
  // Random delay for staggered explosion
  const delay = Math.random() * 1.2
  // Random size variation
  const size = 0.8 + Math.random() * 0.4
  return {
    '--firework-x': `${startX}px`,
    '--firework-y': `${startY}px`,
    '--delay': `${delay}s`,
    '--firework-size': size
  }
}

// Animate percentage when training completes
watch(() => sessionComplete.value, (complete) => {
  if (complete) {
    animatedPercentage.value = 0
    animatedPercentageOffset.value = circumference.value // Start from full (empty circle)
    percentageAnimationComplete.value = false
    const target = accuracyPercentage.value
    const duration = 1500 // 1.5 seconds
    const startTime = Date.now()
    
    const animate = () => {
      const elapsed = Date.now() - startTime
      const progress = Math.min(elapsed / duration, 1)
      // Easing function (ease-out cubic) - matches CSS cubic-bezier(0.4, 0, 0.2, 1)
      // CSS cubic-bezier(0.4, 0, 0.2, 1) approximates to ease-out cubic
      const eased = 1 - Math.pow(1 - progress, 3)
      animatedPercentage.value = Math.round(target * eased)
      // Animate circle offset simultaneously
      animatedPercentageOffset.value = circumference.value * (1 - eased * target / 100)
      
      if (progress < 1) {
        requestAnimationFrame(animate)
      } else {
        animatedPercentage.value = target
        animatedPercentageOffset.value = circumference.value * (1 - target / 100)
        percentageAnimationComplete.value = true
      }
    }
    
    requestAnimationFrame(animate)
  }
})

// Play victory/defeat melodies when animations start
watch([() => percentageAnimationComplete.value, () => accuracyPercentage.value], ([complete, percentage]) => {
  if (!complete || !settings.value.soundsEnabled) return
  
  if (percentage > 90) {
    // Victory - play when fireworks animation starts
    playVictory(settings.value.soundTheme)
  } else if (percentage < 10) {
    // Defeat - play when failure animation starts
    playDefeat(settings.value.soundTheme)
  }
})

// Calculate progress for circular progress bar (delay timer)
const delayCircumference = computed(() => {
  const radius = 34
  return 2 * Math.PI * radius
})

const strokeDashoffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0) {
    return delayCircumference.value
  }
  // Calculate progress based on remaining milliseconds for precision
  const progress = remainingMs.value / initialDelayMs.value
  return delayCircumference.value * (1 - progress)
})

// Calculate progress for error circular progress bar (5 seconds countdown)
const errorCircumference = computed(() => {
  const radius = 16
  return 2 * Math.PI * radius
})

const errorProgressOffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0 || feedback.value?.is_correct) {
    return 0
  }
  // For incorrect answers: progress goes from 0 to full (reverse countdown)
  const progress = remainingMs.value / initialDelayMs.value
  return errorCircumference.value * (1 - progress)
})

const cardIndex = ref(0)
const totalCards = ref(0)
const userCardId = ref(0)

// Encouraging phrases with weighted distribution
// First phrase has 30% probability, last has 0.01%
const encouragingPhrases = [
  'Perfect!', 'Excellent!', 'Well done!', 'Great job!', 'Awesome!',
  'Fantastic!', 'Brilliant!', 'Outstanding!', 'Superb!', 'Terrific!',
  'Amazing!', 'Wonderful!', 'Impressive!', 'Splendid!', 'Marvelous!',
  'Bravo!', 'Nice work!', 'Keep it up!', 'You got it!', 'Spot on!',
  'Right on!', 'Exactly!', 'Precisely!', 'Correct!', 'That\'s it!',
  'You nailed it!', 'Way to go!', 'Good thinking!', 'Smart!', 'Clever!',
  'Genius!', 'Incredible!', 'Phenomenal!', 'Exceptional!', 'Remarkable!',
  'Stellar!', 'Magnificent!', 'Fabulous!', 'Sensational!', 'Triumphant!',
  'Victorious!', 'Champion!', 'Masterful!', 'Skillful!', 'Proficient!',
  'Admirable!', 'Commendable!', 'Praiseworthy!', 'Laudable!', 'Notable!'
]

// Disappointing phrases with weighted distribution
// First phrase has 30% probability, last has 0.01%
const disappointingPhrases = [
  'Not quite', 'Almost there', 'Close!', 'Try again', 'Not quite right',
  'Almost!', 'Close, but...', 'Not exactly', 'Almost correct', 'Nearly there',
  'So close!', 'Almost got it', 'Close one!', 'Almost right', 'Nearly correct',
  'Just missed', 'Almost perfect', 'Close call', 'Nearly there!', 'Almost!',
  'Not quite yet', 'Keep trying', 'Almost had it', 'Close attempt', 'Nearly!',
  'Almost correct!', 'Close, try again', 'Not quite right', 'Almost there!', 'Nearly got it',
  'Close but no', 'Almost perfect!', 'Not exactly right', 'Close one!', 'Almost!',
  'Nearly correct!', 'Close attempt!', 'Almost there!', 'Not quite yet!', 'Keep going!',
  'Almost had it!', 'Close call!', 'Nearly there!', 'Almost correct!', 'Keep trying!',
  'Not quite right!', 'Almost!', 'Close!', 'Nearly!', 'Try again!'
]

// Generate weights helper function
const generateWeights = (phrases: string[]) => {
  const weights: number[] = []
  const n = phrases.length
  const maxWeight = 30 // 30%
  const minWeight = 0.01 // 0.01%
  
  // Exponential decay: weight = maxWeight * (minWeight/maxWeight)^((i)/(n-1))
  for (let i = 0; i < n; i++) {
    const ratio = i / (n - 1)
    const weight = maxWeight * Math.pow(minWeight / maxWeight, ratio)
    weights.push(weight)
  }
  
  // Normalize to sum to 100
  const sum = weights.reduce((a, b) => a + b, 0)
  return weights.map(w => w * 100 / sum)
}

// Cumulative distribution helper function
const generateCumulativeWeights = (weights: number[]) => {
  const cumulative: number[] = []
  let sum = 0
  for (const weight of weights) {
    sum += weight
    cumulative.push(sum)
  }
  return cumulative
}

// Get weighted random message for motivational messages
const getWeightedMessage = (messages: string[]): string => {
  const weights = generateMessageWeights(messages.length)
  const cumulative = generateCumulativeWeights(weights)
  
  const random = Math.random() * 100
  for (let i = 0; i < cumulative.length; i++) {
    if (random <= cumulative[i]) {
      return messages[i]
    }
  }
  return messages[0] // Fallback
}

// Generate weights and cumulative distributions
const encouragingWeights = generateWeights(encouragingPhrases)
const encouragingCumulative = generateCumulativeWeights(encouragingWeights)

const disappointingWeights = generateWeights(disappointingPhrases)
const disappointingCumulative = generateCumulativeWeights(disappointingWeights)

// Get random encouraging phrase based on weighted distribution
const getRandomEncouragingPhrase = (): string => {
  const random = Math.random() * 100
  for (let i = 0; i < encouragingCumulative.length; i++) {
    if (random <= encouragingCumulative[i]) {
      return encouragingPhrases[i]
    }
  }
  return encouragingPhrases[0] // Fallback
}

// Get random disappointing phrase based on weighted distribution
const getRandomDisappointingPhrase = (): string => {
  const random = Math.random() * 100
  for (let i = 0; i < disappointingCumulative.length; i++) {
    if (random <= disappointingCumulative[i]) {
      return disappointingPhrases[i]
    }
  }
  return disappointingPhrases[0] // Fallback
}

// Timer for automatic options reveal
let autoRevealTimer: ReturnType<typeof setTimeout> | null = null
// Timer for automatic next card transition
let autoNextCardTimer: ReturnType<typeof setTimeout> | null = null
const cardShownAt = ref<Date | null>(null)

// Timer pause state
const timerPaused = ref(false)
let timerPauseStartTime: number | null = null
let timerPausedRemainingMs: number | null = null
let countdownAnimationFrameId: number | null = null
let timerEndTime: number | null = null

// Process question to wrap transcription in span if not already wrapped
const processedQuestion = computed(() => {
  if (!currentCard.value?.question) return ''
  
  let question = currentCard.value.question
  
  // Check if transcription is already wrapped
  if (question.includes('<span class="transcription">')) {
    return question
  }
  
  // Pattern to match transcription: /.../ after </strong>
  // Match: </strong> /.../
  const transcriptionPattern = /(<\/strong>)\s*(\/[^\/]+\/)/g
  question = question.replace(transcriptionPattern, '$1 <span class="transcription">$2</span>')
  
  return question
})

const handleKeyPress = (event: KeyboardEvent) => {
  // Only handle if training is active, options are shown, no feedback, and not answering
  if (!sessionActive.value || !optionsShown.value || feedback.value || answering.value) {
    return
  }
  
  // Handle number keys 1-4
  const key = event.key
  if (key >= '1' && key <= '4') {
    const optionIndex = parseInt(key) - 1 // Convert 1-4 to 0-3
    // Check if option index is valid
    if (optionIndex >= 0 && optionIndex < options.value.length) {
      event.preventDefault()
      submitAnswer(optionIndex)
    }
  }
}

onMounted(async () => {
  // Set up network error callback
  apiClient.setNetworkErrorCallback((isRetrying: boolean, attempt: number, maxAttempts: number) => {
    networkError.value = true
    networkErrorRetrying.value = isRetrying
    networkErrorAttempt.value = attempt
    networkErrorMaxAttempts.value = maxAttempts
  })
  
  // Set up network success callback to hide error notification
  apiClient.setNetworkSuccessCallback(() => {
    networkError.value = false
    networkErrorRetrying.value = false
  })
  
  // Add keyboard event listener
  window.addEventListener('keydown', handleKeyPress)
  
  await loadStats()
  await loadUpcomingCards()
  await checkCurrentSession()
})

const loadStats = async () => {
  try {
    const data: { due_count: number; total_cards?: number; available_for_training?: number } = await apiClient.request('/app/dashboard')
    stats.value.dueCount = data.due_count || 0
    stats.value.totalCards = data.total_cards || 0
    stats.value.availableForTraining = data.available_for_training || data.due_count || 0
    statsLoaded.value = true
  } catch (error) {
    console.error('Failed to load stats:', error)
    statsLoaded.value = true // Mark as loaded even on error to avoid infinite loading state
  }
}

const loadUpcomingCards = async () => {
  try {
    const data = await apiClient.request('/app/training/upcoming')
    console.log('Upcoming cards data:', data)
    
    // Ensure data is in correct format
    if (data && typeof data === 'object') {
      upcomingCardsData.value = data
    } else {
      console.warn('Invalid data format:', data)
      upcomingCardsData.value = {}
    }
    
    upcomingCardsLoaded.value = true
    await nextTick()
    setTimeout(() => {
      updateUpcomingChart()
    }, 150)
  } catch (error) {
    console.error('Failed to load upcoming cards:', error)
    upcomingCardsLoaded.value = true // Mark as loaded even on error
  }
}

const updateUpcomingChart = () => {
  if (!upcomingChartCanvas.value) {
    setTimeout(() => {
      if (upcomingChartCanvas.value) {
        updateUpcomingChart()
      }
    }, 100)
    return
  }
  
  if (!upcomingCardsData.value || Object.keys(upcomingCardsData.value).length === 0) {
    return
  }
  
  // Destroy existing chart if it exists
  if (upcomingChartInstance) {
    upcomingChartInstance.destroy()
    upcomingChartInstance = null
  }
  
  // Prepare data - ensure we process dates in order
  const dates = Object.keys(upcomingCardsData.value).sort()
  const labels: string[] = []
  const counts: number[] = []
  
  console.log('Processing upcoming cards data:', upcomingCardsData.value)
  console.log('Sorted dates:', dates)
  
  dates.forEach(date => {
    const item = upcomingCardsData.value[date]
    console.log(`Processing date ${date}:`, item)
    if (item && typeof item === 'object' && 'label' in item && 'count' in item) {
      labels.push(item.label)
      counts.push(item.count)
      console.log(`Added: label="${item.label}", count=${item.count}`)
    } else {
      // Fallback if data structure is different
      console.warn('Unexpected data structure for date:', date, item)
      // Try to extract date part for display
      const datePart = date.split('T')[0] || date
      labels.push(datePart)
      counts.push(0)
    }
  })
  
  console.log('Final chart data - labels:', labels, 'counts:', counts)
  
  // Get theme colors
  const root = getComputedStyle(document.documentElement)
  const isDark = document.documentElement.getAttribute('data-theme') === 'dark'
  const primaryColor = root.getPropertyValue('--color-primary').trim() || '#007bff'
  const textPrimary = root.getPropertyValue('--text-primary').trim() || '#333333'
  const textSecondary = root.getPropertyValue('--text-secondary').trim() || '#666666'
  const borderColor = root.getPropertyValue('--border-primary').trim() || '#dddddd'
  
  // Convert hex to rgba
  const hexToRgba = (hex: string, alpha: number) => {
    const r = parseInt(hex.slice(1, 3), 16)
    const g = parseInt(hex.slice(3, 5), 16)
    const b = parseInt(hex.slice(5, 7), 16)
    return `rgba(${r}, ${g}, ${b}, ${alpha})`
  }
  
  // Create bar chart
  upcomingChartInstance = new Chart(upcomingChartCanvas.value, {
    type: 'bar',
    data: {
      labels: labels,
      datasets: [{
        label: 'Cards',
        data: counts,
        backgroundColor: hexToRgba(primaryColor, isDark ? 0.7 : 0.6),
        borderColor: primaryColor,
        borderWidth: 1
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          titleColor: '#fff',
          bodyColor: '#fff',
          borderColor: borderColor,
          borderWidth: 1,
          padding: 12,
          callbacks: {
            label: function(context) {
              const value = context.parsed.y || 0
              return `${value} ${value === 1 ? 'card' : 'cards'}`
            }
          }
        }
      },
      scales: {
        x: {
          ticks: {
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            }
          },
          grid: {
            color: borderColor,
            display: false
          }
        },
        y: {
          type: 'linear',
          display: true,
          beginAtZero: true,
          ticks: {
            stepSize: 1,
            color: isDark ? textSecondary : '#555555',
            font: {
              size: 11
            },
            callback: function(value) {
              return Number.isInteger(value) ? value : ''
            }
          },
          grid: {
            color: isDark ? borderColor : '#e0e0e0'
          }
        }
      }
    }
  })
}

onUnmounted(() => {
  // Remove keyboard event listener
  window.removeEventListener('keydown', handleKeyPress)
  
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  
  // Destroy chart
  if (upcomingChartInstance) {
    upcomingChartInstance.destroy()
    upcomingChartInstance = null
  }
})

const checkCurrentSession = async () => {
  try {
    const card: Card = await apiClient.request('/app/training/current')
    sessionActive.value = true
    setupCard(card)
  } catch (error: any) {
    if (!error.message?.includes('404')) {
      console.error('Failed to check session:', error)
    }
  }
}

const setupCard = (card: Card) => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }

  currentCard.value = card
  cardIndex.value = card.card_index
  totalCards.value = card.total_cards
  userCardId.value = card.user_card_id
  optionsShown.value = false
  options.value = []
  feedback.value = null
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  timerEndTime = null
  cardShownAt.value = new Date()

  // Schedule automatic options reveal
  if (card.delay_ms > 0) {
    autoRevealTimer = setTimeout(() => {
      if (!optionsShown.value) {
        revealOptions(false) // false = not early reveal
      }
    }, card.delay_ms)
  }
}

const startTraining = async () => {
  loading.value = true
  try {
    const card: Card = await apiClient.request('/app/training/start', { method: 'POST' })
    sessionActive.value = true
    setupCard(card)
    sessionComplete.value = false
    // Update stats after starting
    await loadStats()
  } catch (error: any) {
    if (error.message?.includes('No cards available')) {
      await showAlert('No cards available for training. Request some words first!')
    } else {
      console.error('Failed to start training:', error)
      await showAlert('Failed to start training')
    }
  } finally {
    loading.value = false
  }
}

const revealOptions = async (isEarly: boolean = false) => {
  // Clear timer if it exists
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }

  // If already shown, don't do anything
  if (optionsShown.value) {
    return
  }

  try {
    const data: OptionsResponse = await apiClient.request('/app/training/reveal', { 
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
    options.value = data.options
    optionsShown.value = true
  } catch (error: any) {
    console.error('Failed to reveal options:', error)
    // Network error is already handled by callback, but we should handle other errors
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert('Не удалось загрузить варианты ответов. Попробуйте обновить страницу.')
    }
  }
}

// Sound functions using new audio system
const playCorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playSuccess(settings.value.soundTheme)
}

const playIncorrectSound = () => {
  if (!settings.value.soundsEnabled) return
  playFail(settings.value.soundTheme)
}

// Haptic feedback helper function
const triggerHapticFeedback = (isCorrect: boolean) => {
  if (!settings.value.vibrationEnabled) return
  
  const tg = (window as any).Telegram?.WebApp
  
  // Try Telegram Web App API first
  if (tg?.HapticFeedback) {
    try {
      const haptic = tg.HapticFeedback
      if (isCorrect) {
        // Success feedback - try notificationOccurred first, then impactOccurred
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('success')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('medium')
        }
      } else {
        // Error feedback - try notificationOccurred first, then impactOccurred
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('error')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('heavy')
        }
      }
      return
    } catch (error) {
      console.warn('Telegram haptic feedback failed:', error)
    }
  }
  
  // Fallback to native Vibration API
  if ('vibrate' in navigator && typeof navigator.vibrate === 'function') {
    try {
      if (isCorrect) {
        // Short, pleasant vibration for correct answer
        navigator.vibrate(50)
      } else {
        // Longer, more noticeable vibration for incorrect answer
        navigator.vibrate([100, 50, 100])
      }
    } catch (error) {
      console.warn('Native vibration failed:', error)
    }
  }
}

const submitAnswer = async (optionIndex: number) => {
  answering.value = true
  try {
    const formData = new FormData()
    formData.append('option_index', optionIndex.toString())
    formData.append('user_card_id', userCardId.value.toString())
    
    const data: Feedback = await apiClient.requestFormData('/app/training/answer', formData)
    feedback.value = data
    
    // Trigger haptic feedback based on answer correctness
    triggerHapticFeedback(data.is_correct)
    
    // Play sound based on answer correctness
    if (data.is_correct) {
      playCorrectSound()
    } else {
      playIncorrectSound()
    }
    
    // Generate random phrase based on answer correctness
    if (data.is_correct) {
      currentEncouragingPhrase.value = getRandomEncouragingPhrase()
    } else {
      currentDisappointingPhrase.value = getRandomDisappointingPhrase()
    }
    
    // Schedule automatic transition to next card
    const delayMs = data.delay_seconds ? data.delay_seconds * 1000 : 0
    
    if (delayMs > 0) {
      const totalSeconds = data.delay_seconds!
      initialDelaySeconds.value = totalSeconds
      initialDelayMs.value = delayMs
      delaySeconds.value = totalSeconds
      remainingMs.value = delayMs
      waitingDelay.value = true
      
      // Reset pause state
      timerPaused.value = false
      timerPauseStartTime = null
      timerPausedRemainingMs = null
      
      const startTime = Date.now()
      timerEndTime = startTime + delayMs
      
      // Update countdown with precise timing using requestAnimationFrame
      const updateCountdown = () => {
        if (!timerEndTime) {
          return
        }
        
        if (timerPaused.value) {
          // Timer is paused, don't update but keep the loop running
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
          return
        }
        
        const now = Date.now()
        const currentRemainingMs = Math.max(0, timerEndTime! - now)
        const currentRemainingSeconds = Math.ceil(currentRemainingMs / 1000)
        
        remainingMs.value = currentRemainingMs
        delaySeconds.value = currentRemainingSeconds
        
        if (currentRemainingMs > 0) {
          countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
        } else {
          delaySeconds.value = 0
          remainingMs.value = 0
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
          if (countdownAnimationFrameId) {
            cancelAnimationFrame(countdownAnimationFrameId)
            countdownAnimationFrameId = null
          }
          nextCard()
        }
      }
      
      // Start updating immediately
      countdownAnimationFrameId = requestAnimationFrame(updateCountdown)
      
      // Schedule automatic next card as backup
      autoNextCardTimer = setTimeout(() => {
        if (countdownAnimationFrameId) {
          cancelAnimationFrame(countdownAnimationFrameId)
          countdownAnimationFrameId = null
        }
        if (waitingDelay.value) {
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          delaySeconds.value = 0
          remainingMs.value = 0
          timerPaused.value = false
          timerPauseStartTime = null
          timerPausedRemainingMs = null
          timerEndTime = null
        }
        nextCard()
      }, delayMs)
    } else {
      // No delay, go to next card immediately
      autoNextCardTimer = setTimeout(() => {
        nextCard()
      }, 1000) // Small delay to show feedback
    }
  } catch (error: any) {
    console.error('Failed to submit answer:', error)
    // Network error is already handled by callback
    if (!error.isNetworkError) {
      // For non-network errors, show a simple message
      await showAlert('Не удалось отправить ответ. Попробуйте еще раз.')
    }
  } finally {
    answering.value = false
  }
}

const nextCard = async () => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }

  feedback.value = null
  optionsShown.value = false
  options.value = []
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  initialDelaySeconds.value = 0
  cardShownAt.value = null
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null

  try {
    const response = await apiClient.request('/app/training/current')
    
    // Check if training is complete (response has complete field)
    if (response && typeof response === 'object' && 'complete' in response) {
      const completionData = response as { complete: boolean; cards_completed: number; total_cards?: number; correct_cards?: number }
      sessionComplete.value = true
      cardsCompleted.value = completionData.cards_completed || 0
      trainingStats.value = {
        totalCards: completionData.total_cards || completionData.cards_completed || 0,
        correctCards: completionData.correct_cards || 0
      }
      sessionActive.value = false
      currentCard.value = null
      await loadStats() // Refresh stats after completion
      return
    }
    
    // Normal card response
    const card = response as Card
    setupCard(card)
    
    // Check if training is complete based on card index
    if (card.card_index > card.total_cards) {
      // Training completed - get stats from last request
      sessionComplete.value = true
      cardsCompleted.value = card.card_index - 1
      // Try to get stats by making another request
      try {
        const statsResponse = await apiClient.request('/app/training/current')
        if (statsResponse && typeof statsResponse === 'object' && 'complete' in statsResponse) {
          const statsData = statsResponse as { total_cards?: number; correct_cards?: number }
          trainingStats.value = {
            totalCards: statsData.total_cards || card.card_index - 1,
            correctCards: statsData.correct_cards || 0
          }
        } else {
          trainingStats.value = {
            totalCards: card.card_index - 1,
            correctCards: 0
          }
        }
      } catch {
        trainingStats.value = {
          totalCards: card.card_index - 1,
          correctCards: 0
        }
      }
      sessionActive.value = false
      currentCard.value = null
      await loadStats() // Refresh stats after completion
    }
  } catch (error: any) {
    if (error.message?.includes('404')) {
      // No more cards - training completed
      sessionComplete.value = true
      sessionActive.value = false
      currentCard.value = null
      // Stats will be 0 if we can't get them
      trainingStats.value = {
        totalCards: cardsCompleted.value || 0,
        correctCards: 0
      }
      await loadStats() // Refresh stats after completion
    } else {
      console.error('Failed to get next card:', error)
      // Network error is already handled by callback
      if (!error.isNetworkError) {
        // For non-network errors, show a simple message
        await showAlert('Не удалось загрузить следующую карточку. Попробуйте обновить страницу.')
      }
    }
  }
}

const resetSession = async () => {
  // Clear any existing timers
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
  }
  if (countdownAnimationFrameId) {
    cancelAnimationFrame(countdownAnimationFrameId)
    countdownAnimationFrameId = null
  }

  sessionActive.value = false
  currentCard.value = null
  optionsShown.value = false
  options.value = []
  feedback.value = null
  waitingDelay.value = false
  delaySeconds.value = 0
  initialDelaySeconds.value = 0
  remainingMs.value = 0
  initialDelayMs.value = 0
  sessionComplete.value = false
  cardsCompleted.value = 0
  trainingStats.value = {
    totalCards: 0,
    correctCards: 0
  }
  cardShownAt.value = null
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
  timerEndTime = null
  
  // Refresh stats and upcoming cards
  await loadStats()
  await loadUpcomingCards()
}

// Timer pause/resume handlers
const pauseTimer = () => {
  if (!waitingDelay.value || timerPaused.value || !timerEndTime) return
  
  timerPaused.value = true
  timerPauseStartTime = Date.now()
  timerPausedRemainingMs = remainingMs.value
}

const resumeTimer = () => {
  if (!waitingDelay.value || !timerPaused.value || timerPauseStartTime === null || timerPausedRemainingMs === null || !timerEndTime) return
  
  // Calculate how long the timer was paused
  const pauseDuration = Date.now() - timerPauseStartTime
  
  // Adjust the end time by adding the pause duration
  timerEndTime = timerEndTime + pauseDuration
  
  timerPaused.value = false
  timerPauseStartTime = null
  timerPausedRemainingMs = null
}

// Handle mouse/touch events for timer pause
const handleTimerMouseDown = (event: MouseEvent | TouchEvent) => {
  event.preventDefault()
  pauseTimer()
}

const handleTimerMouseUp = (event: MouseEvent | TouchEvent) => {
  event.preventDefault()
  resumeTimer()
}

const handleTimerMouseLeave = () => {
  // Resume if mouse leaves while button might be pressed
  if (timerPaused.value) {
    resumeTimer()
  }
}
</script>

<style scoped>
.training {
  max-width: 1200px;
  margin: 0 auto;
  padding: 10px;
}

.training h1 {
  margin-bottom: 24px;
}
.training-progress {
  margin-bottom: 20px;
  text-align: center;
}

.question {
  font-size: 24px;
  margin: 30px 0;
  text-align: center;
  overflow-x: auto;
  overflow-y: visible;
}

/* Prevent word and transcription from breaking into multiple lines */
.question :deep(strong),
.question :deep(.transcription),
.question :deep(span.transcription) {
  white-space: nowrap;
}

/* Style transcription in question HTML content */
.question :deep(.transcription),
.question :deep(span.transcription) {
  font-family: 'Arial Unicode MS', 'Lucida Sans Unicode', 'Charis SIL', 'Doulos SIL', 'Gentium Plus', 'DejaVu Sans', Arial, sans-serif !important;
  font-style: italic !important;
  letter-spacing: 0.5px !important;
  font-size: 0.9em !important;
}

/* Style any text that looks like transcription (starts and ends with /) */
.question {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

.question :deep(strong) {
  font-family: inherit;
}

/* Target text nodes after strong that contain transcription pattern */
.question :deep(*) {
  font-family: inherit;
}

.options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 20px 0;
}

.option-btn {
  min-height: 60px;
  font-size: 16px;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}

.option-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
}

.option-btn.option-correct .option-number,
.option-btn.option-incorrect .option-number {
  background: rgba(255, 255, 255, 0.3);
}

.option-text {
  flex: 1;
  text-align: left;
}

.option-btn.option-disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.option-btn.option-correct {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: 2px solid #10b981;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  animation: correct-success 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
  position: relative;
  overflow: hidden;
}

.option-btn.option-correct::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: linear-gradient(
    45deg,
    transparent 30%,
    rgba(255, 255, 255, 0.3) 50%,
    transparent 70%
  );
  animation: correct-shine 0.8s ease-out;
  pointer-events: none;
}

.option-btn.option-incorrect {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: 2px solid #ef4444;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  animation: incorrect-fail 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55);
  position: relative;
}

.option-btn.option-incorrect::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  border-radius: 8px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  transform: translate(-50%, -50%) scale(0);
  animation: incorrect-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes incorrect-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 0.6;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.2);
    opacity: 0.3;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

@keyframes correct-success {
  0% {
    transform: scale(1);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  30% {
    transform: scale(1.15) rotate(2deg);
    box-shadow: 0 8px 24px rgba(16, 185, 129, 0.5);
  }
  60% {
    transform: scale(1.08) rotate(-1deg);
    box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
  }
  100% {
    transform: scale(1);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
}

@keyframes correct-shine {
  0% {
    transform: translateX(-100%) translateY(-100%) rotate(45deg);
  }
  100% {
    transform: translateX(100%) translateY(100%) rotate(45deg);
  }
}

@keyframes incorrect-fail {
  0%, 100% {
    transform: translateX(0) scale(1);
  }
  10% {
    transform: translateX(-12px) scale(0.95) rotate(-3deg);
  }
  20% {
    transform: translateX(12px) scale(0.95) rotate(3deg);
  }
  30% {
    transform: translateX(-10px) scale(0.97) rotate(-2deg);
  }
  40% {
    transform: translateX(10px) scale(0.97) rotate(2deg);
  }
  50% {
    transform: translateX(-8px) scale(0.98) rotate(-1deg);
  }
  60% {
    transform: translateX(8px) scale(0.98) rotate(1deg);
  }
  70% {
    transform: translateX(-4px) scale(0.99);
  }
  80% {
    transform: translateX(4px) scale(0.99);
  }
  90% {
    transform: translateX(-2px) scale(1);
  }
}

.feedback-section {
  margin-top: 30px;
  text-align: center;
}

.feedback-badge {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 16px 32px;
  border-radius: 12px;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  position: relative;
  overflow: hidden;
}

.feedback-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  animation: feedback-success-appear 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feedback-success::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.3) 0%, transparent 70%);
  animation: feedback-success-glow 1.5s ease-out;
  pointer-events: none;
}

.feedback-error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  animation: feedback-error-appear 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.feedback-error::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 100%;
  height: 100%;
  border-radius: 12px;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.2) 0%, transparent 70%);
  transform: translate(-50%, -50%) scale(0);
  animation: feedback-error-pulse 0.6s ease-out;
  pointer-events: none;
}

@keyframes feedback-error-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 0.8;
  }
  50% {
    transform: translate(-50%, -50%) scale(1.3);
    opacity: 0.4;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.6);
    opacity: 0;
  }
}

@keyframes feedback-success-appear {
  0% {
    opacity: 0;
    transform: scale(0.3) translateY(-30px) rotate(-10deg);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
  50% {
    transform: scale(1.1) translateY(0) rotate(5deg);
    box-shadow: 0 12px 32px rgba(16, 185, 129, 0.5);
  }
  70% {
    transform: scale(0.95) translateY(0) rotate(-2deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0) rotate(0deg);
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
  }
}

@keyframes feedback-success-glow {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 1;
  }
  100% {
    transform: translate(-50%, -50%) scale(1.5);
    opacity: 0;
  }
}

@keyframes feedback-error-appear {
  0% {
    opacity: 0;
    transform: scale(0.5) translateY(20px) rotate(10deg);
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  }
  30% {
    transform: scale(1.15) translateY(-5px) rotate(-5deg);
    box-shadow: 0 8px 24px rgba(239, 68, 68, 0.5);
  }
  60% {
    transform: scale(0.9) translateY(2px) rotate(2deg);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0) rotate(0deg);
    box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  }
}

.feedback-icon {
  font-size: 28px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  flex-shrink: 0;
  animation: feedback-icon-spin 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.feedback-error .feedback-icon {
  animation: feedback-icon-shake 0.5s ease-out;
}

@keyframes feedback-icon-spin {
  0% {
    transform: scale(0) rotate(-180deg);
  }
  60% {
    transform: scale(1.3) rotate(10deg);
  }
  100% {
    transform: scale(1) rotate(0deg);
  }
}

@keyframes feedback-icon-shake {
  0%, 100% {
    transform: translateX(0) rotate(0deg);
  }
  25% {
    transform: translateX(-8px) rotate(-10deg);
  }
  75% {
    transform: translateX(8px) rotate(10deg);
  }
}

/* Success particles animation */
.success-particles {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 0;
  height: 0;
  pointer-events: none;
  z-index: 1;
}

.success-particle {
  position: absolute;
  top: 0;
  left: 0;
  width: var(--particle-size, 6px);
  height: var(--particle-size, 6px);
  background: radial-gradient(circle, rgba(255, 255, 255, 0.9) 0%, rgba(16, 185, 129, 0.8) 100%);
  border-radius: 50%;
  box-shadow: 0 0 6px rgba(16, 185, 129, 0.6);
  animation: success-particle-fly 1s ease-out var(--particle-delay, 0s) forwards;
}

@keyframes success-particle-fly {
  0% {
    opacity: 1;
    transform: translate(0, 0) scale(1);
  }
  100% {
    opacity: 0;
    transform: translate(var(--particle-end-x, 0), var(--particle-end-y, 0)) scale(0);
  }
}

.feedback-text {
  letter-spacing: 0.5px;
}

.error-progress-wrapper {
  position: relative;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}

.error-progress-ring {
  position: absolute;
  width: 40px;
  height: 40px;
  transform: rotate(-90deg);
}

.error-progress-circle-bg {
  opacity: 0.3;
}

.error-progress-circle {
  transition: stroke-dashoffset 0.1s linear;
  stroke-linecap: round;
}

.error-icon-svg {
  position: relative;
  z-index: 1;
  color: white;
  flex-shrink: 0;
}

.hint {
  margin: 20px 0 10px 0;
  padding: 12px 18px;
  background: var(--hint-bg, rgba(245, 158, 11, 0.1));
  border-radius: 8px;
  color: var(--text-primary);
  border-left: 4px solid #f59e0b;
  text-align: left;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
  font-size: 14px;
  line-height: 1.5;
}

.example {
  font-style: italic;
  margin: 10px 0 20px 0;
  padding: 15px 20px;
  background: var(--example-bg, rgba(59, 130, 246, 0.1));
  border-radius: 8px;
  color: var(--text-primary);
  border-left: 4px solid var(--color-primary);
  text-align: left;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.waiting-progress {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}

.circular-progress {
  position: relative;
  width: 80px;
  height: 80px;
}

.progress-ring {
  transform: rotate(-90deg);
  transition: none;
}

.progress-ring-circle {
  transition: none;
  stroke-linecap: round;
}

.progress-ring-circle-bg {
  opacity: 0.3;
}

.progress-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 24px;
  font-weight: bold;
  color: var(--color-primary);
  user-select: none;
}

.start-screen {
  text-align: center;
  padding: 40px 20px;
}

.start-screen-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 28px;
  max-width: 600px;
  margin: 0 auto;
}

.start-screen-stats {
  display: flex;
  flex-direction: column;
  gap: 20px;
  width: 100%;
}

.start-stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 24px;
  background: linear-gradient(135deg, var(--bg-secondary, rgba(0, 0, 0, 0.05)) 0%, var(--bg-secondary, rgba(0, 0, 0, 0.08)) 100%);
  border-radius: 16px;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.start-stat-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

.start-stat-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 1px;
  opacity: 0.8;
}

.start-stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--color-primary);
  display: inline-block;
  line-height: 1.2;
}

.start-stat-value span:last-child {
  font-size: 18px;
  font-weight: 500;
  color: var(--text-secondary);
  opacity: 0.7;
  margin-left: 6px;
}

.upcoming-cards-chart {
  width: 100%;
  margin-top: 28px;
  padding: 24px;
  background: linear-gradient(135deg, var(--bg-secondary, rgba(0, 0, 0, 0.05)) 0%, var(--bg-secondary, rgba(0, 0, 0, 0.08)) 100%);
  border-radius: 16px;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.upcoming-cards-chart:hover {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
}

.chart-header {
  margin-bottom: 20px;
  text-align: center;
}

.chart-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 6px 0;
  letter-spacing: -0.3px;
}

.chart-subtitle {
  font-size: 13px;
  font-weight: 400;
  color: var(--text-secondary);
  opacity: 0.7;
  margin: 0;
}

.chart-container {
  position: relative;
  height: 220px;
  width: 100%;
  margin-top: 8px;
}

.btn-start {
  width: 100%;
  max-width: 320px;
  padding: 16px 32px;
  font-size: 17px;
  font-weight: 600;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  letter-spacing: 0.3px;
}

.btn-start:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}

.btn-start:active {
  transform: translateY(0);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
}

.stat-item:last-child {
  border-bottom: none;
}

.stat-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--color-primary);
}

.no-cards-message {
  margin-top: 15px;
  color: var(--text-secondary);
  font-style: italic;
}

.completion-screen {
  text-align: center;
  position: relative;
  overflow: hidden;
  padding: 40px 20px;
}

.completion-percentage {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
}

.percentage-circle-wrapper {
  position: relative;
  width: 200px;
  height: 200px;
  flex-shrink: 0;
}

.percentage-circle {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.percentage-circle-bg {
  opacity: 0.2;
}

.percentage-circle-fill {
  transition: stroke 0.3s ease;
  /* No transition for stroke-dashoffset - animated via JavaScript for perfect sync */
}

.percentage-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.percentage-number {
  font-size: 48px;
  font-weight: 700;
  color: var(--text-primary);
  transition: color 0.3s ease;
  line-height: 1;
}

.percentage-ratio {
  font-size: 20px;
  font-weight: 500;
  color: var(--text-secondary);
  opacity: 0.85;
  line-height: 1;
}

.motivational-message {
  padding: 18px 28px;
  border-radius: 12px;
  max-width: 600px;
  width: 100%;
  animation: messageAppear 0.6s ease-out 0.3s both;
  text-align: center;
}

.message-text {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
  line-height: 1.5;
}

.message-excellent {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(5, 150, 105, 0.15) 100%);
  color: var(--color-success, #10b981);
  border: 2px solid rgba(16, 185, 129, 0.3);
}

.message-good {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(37, 99, 235, 0.15) 100%);
  color: #3b82f6;
  border: 2px solid rgba(59, 130, 246, 0.3);
}

.message-okay {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(217, 119, 6, 0.15) 100%);
  color: #f59e0b;
  border: 2px solid rgba(245, 158, 11, 0.3);
}

.message-needs-improvement {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(220, 38, 38, 0.15) 100%);
  color: #ef4444;
  border: 2px solid rgba(239, 68, 68, 0.3);
}

@keyframes messageAppear {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Celebration animations */
.celebration-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  overflow: visible;
  z-index: 2;
  width: 0;
  height: 0;
}

.fireworks {
  position: absolute;
  width: 100%;
  height: 100%;
}

.firework {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  top: 0;
  left: 0;
  transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(var(--firework-size, 1));
  animation: firework-explode 2s ease-out var(--delay, 0s) forwards;
}

.firework-core {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: radial-gradient(circle, #fff 0%, #ffd700 50%, transparent 100%);
  box-shadow: 0 0 10px rgba(255, 215, 0, 0.8), 0 0 20px rgba(255, 215, 0, 0.5);
  animation: firework-core-pulse 0.3s ease-out var(--delay, 0s) forwards;
  transform: translate(-50%, -50%);
  top: 50%;
  left: 50%;
}

.firework-particle {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--particle-color, #10b981);
  top: 0;
  left: 0;
  transform: translate(-50%, -50%);
  box-shadow: 0 0 6px var(--particle-color, #10b981);
  animation: particle-fly 2s ease-out var(--delay, 0s) forwards;
}

.firework:nth-child(3n+1) .firework-particle {
  --particle-color: #3b82f6;
}

.firework:nth-child(3n+2) .firework-particle {
  --particle-color: #f59e0b;
}

.firework:nth-child(3n+3) .firework-particle {
  --particle-color: #ec4899;
}

.firework:nth-child(4n+1) .firework-particle {
  --particle-color: #10b981;
}

.firework:nth-child(4n+2) .firework-particle {
  --particle-color: #8b5cf6;
}

@keyframes firework-explode {
  0% {
    opacity: 1;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(var(--firework-size, 1));
  }
  15% {
    opacity: 1;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(calc(var(--firework-size, 1) * 1.5));
  }
  100% {
    opacity: 0;
    transform: translate(var(--firework-x, 0), var(--firework-y, 0)) scale(0);
  }
}

@keyframes firework-core-pulse {
  0% {
    transform: translate(-50%, -50%) scale(0);
    opacity: 1;
  }
  50% {
    transform: translate(-50%, -50%) scale(2);
    opacity: 0.8;
  }
  100% {
    transform: translate(-50%, -50%) scale(3);
    opacity: 0;
  }
}

@keyframes particle-fly {
  0% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1) rotate(0deg);
  }
  50% {
    opacity: 1;
    transform: translate(-50%, -50%) translateX(calc(var(--offset-x, 0) * 0.5)) translateY(calc(var(--offset-y, 0) * 0.5)) scale(1.2) rotate(180deg);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) translateX(var(--offset-x, 0)) translateY(var(--offset-y, 0)) scale(0) rotate(360deg);
  }
}

/* Calculate offsets for 12 particles in a circle (30 degrees each) */
.firework-particle[data-particle-index="1"] { --offset-x: 100px; --offset-y: 0px; }
.firework-particle[data-particle-index="2"] { --offset-x: 86.6px; --offset-y: -50px; }
.firework-particle[data-particle-index="3"] { --offset-x: 50px; --offset-y: -86.6px; }
.firework-particle[data-particle-index="4"] { --offset-x: 0px; --offset-y: -100px; }
.firework-particle[data-particle-index="5"] { --offset-x: -50px; --offset-y: -86.6px; }
.firework-particle[data-particle-index="6"] { --offset-x: -86.6px; --offset-y: -50px; }
.firework-particle[data-particle-index="7"] { --offset-x: -100px; --offset-y: 0px; }
.firework-particle[data-particle-index="8"] { --offset-x: -86.6px; --offset-y: 50px; }
.firework-particle[data-particle-index="9"] { --offset-x: -50px; --offset-y: 86.6px; }
.firework-particle[data-particle-index="10"] { --offset-x: 0px; --offset-y: 100px; }
.firework-particle[data-particle-index="11"] { --offset-x: 50px; --offset-y: 86.6px; }
.firework-particle[data-particle-index="12"] { --offset-x: 86.6px; --offset-y: 50px; }

.confetti {
  position: absolute;
  width: 0;
  height: 0;
  top: 0;
  left: 0;
}

.confetti-piece {
  position: absolute;
  width: 12px;
  height: 12px;
  background: var(--confetti-color, #10b981);
  top: 0;
  left: 0;
  animation: confetti-fly var(--confetti-duration, 3s) 0s ease-out forwards;
  border-radius: 2px;
  box-shadow: 0 0 4px var(--confetti-color, #10b981);
}

.confetti-piece:nth-child(3n+1) {
  --confetti-color: #10b981;
  clip-path: polygon(50% 0%, 0% 100%, 100% 100%);
}

.confetti-piece:nth-child(3n+2) {
  --confetti-color: #3b82f6;
  border-radius: 50%;
}

.confetti-piece:nth-child(3n+3) {
  --confetti-color: #f59e0b;
  clip-path: polygon(25% 0%, 100% 0%, 75% 100%, 0% 100%);
}

.confetti-piece:nth-child(4n+1) {
  --confetti-color: #ec4899;
}

.confetti-piece:nth-child(4n+2) {
  --confetti-color: #8b5cf6;
}

.confetti-piece:nth-child(4n+3) {
  --confetti-color: #06b6d4;
}

@keyframes confetti-fly {
  0% {
    transform: translate(var(--confetti-start-x, 0), var(--confetti-start-y, 0)) rotate(0deg) scale(1);
    opacity: 1;
  }
  50% {
    opacity: 1;
  }
  100% {
    transform: translate(var(--confetti-end-x, 0), var(--confetti-end-y, 0)) rotate(1080deg) scale(0);
    opacity: 0;
  }
}

/* Failure animation for <10% */
.failure-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  overflow: visible;
  z-index: 2;
  width: 0;
  height: 0;
}

.failure-rain {
  position: absolute;
  width: 0;
  height: 0;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.failure-item {
  position: absolute;
  top: 50%;
  left: 50%;
  font-size: 48px;
  line-height: 1;
  animation: failure-fall var(--duration, 2s) var(--delay, 0s) ease-in forwards;
  transform-origin: center;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
  opacity: 0; /* Start completely invisible */
  visibility: hidden; /* Hidden until animation starts */
}

.failure-emoji {
  display: block;
  filter: grayscale(0.4) brightness(0.8);
  user-select: none;
}

@keyframes failure-fall {
  0% {
    opacity: 0;
    visibility: hidden;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0);
  }
  3% {
    opacity: 0;
    visibility: visible;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0);
  }
  6% {
    opacity: 1;
    transform: translate(-50%, -50%) translate(var(--start-x, 0), var(--start-y, 0)) rotate(0deg) scale(0.5);
  }
  90% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) translate(var(--end-x, 0), var(--end-y, 0)) rotate(var(--rotation, 0deg)) scale(var(--scale, 1));
  }
}


.completion-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  width: 100%;
  max-width: 400px;
}

.remaining-cards-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 24px;
  background: var(--bg-secondary, rgba(0, 0, 0, 0.05));
  border-radius: 10px;
  width: 100%;
  border: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
}

.remaining-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-primary);
  text-align: center;
  line-height: 1.4;
}

.remaining-text .remaining-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  margin-right: 6px;
}

.btn-continue {
  width: 100%;
  max-width: 300px;
  padding: 14px 28px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 10px;
}


.network-error-notification {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: white;
  padding: 16px 24px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
  animation: slideDown 0.3s ease-out;
  max-width: 90%;
  width: auto;
  min-width: 300px;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

.network-error-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.network-error-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.network-error-text {
  flex: 1;
}

.network-error-title {
  font-weight: 600;
  font-size: 16px;
  margin-bottom: 4px;
}

.network-error-message {
  font-size: 14px;
  opacity: 0.95;
}

/* Hide option number on mobile */
@media (max-width: 768px) {
  .option-number {
    display: none;
  }
  
  /* Compact completion screen on mobile */
  .completion-screen {
    padding: 30px 8px;
  }
  
  .completion-percentage {
    gap: 20px;
  }
  
  .percentage-circle-wrapper {
    width: 180px;
    height: 180px;
  }
  
  .percentage-number {
    font-size: 42px;
  }
  
  .percentage-ratio {
    font-size: 18px;
  }
  
  .motivational-message {
    padding: 14px 8px;
    font-size: 16px;
  }
  
  .completion-actions {
    max-width: 100%;
  }
  
  .remaining-cards-info {
    padding: 14px 8px;
  }
  
  .remaining-text {
    font-size: 16px;
  }
  
  .btn-continue {
    max-width: 100%;
  }
  
  .start-screen {
    padding: 30px 12px;
  }
  
  .start-screen-content {
    gap: 24px;
  }
  
  .start-stat-item {
    padding: 20px;
    border-radius: 14px;
  }
  
  .start-stat-value {
    font-size: 28px;
  }
  
  .start-stat-value span:last-child {
    font-size: 16px;
  }
  
  .upcoming-cards-chart {
    padding: 20px;
    border-radius: 14px;
    margin-top: 24px;
  }
  
  .chart-title {
    font-size: 16px;
  }
  
  .chart-subtitle {
    font-size: 12px;
  }
  
  .chart-container {
    height: 200px;
  }
  
  .btn-start {
    max-width: 100%;
    padding: 14px 28px;
    font-size: 16px;
  }
}
</style>

