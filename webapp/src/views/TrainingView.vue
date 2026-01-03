<template>
  <div class="training">
    <h1>Training</h1>
    
    <div v-if="sessionComplete && !sessionActive" class="card completion-screen">
      <h2>Training Complete!</h2>
      <div class="completion-stats">
        <div class="stat-item">
          <span class="stat-label">Total Cards:</span>
          <span class="stat-value">{{ trainingStats.totalCards }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Correct Answers:</span>
          <span class="stat-value correct-stat">{{ trainingStats.correctCards }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Accuracy:</span>
          <span class="stat-value">
            {{ trainingStats.totalCards > 0 ? Math.round((trainingStats.correctCards / trainingStats.totalCards) * 100) : 0 }}%
          </span>
        </div>
      </div>
    </div>

    <div v-if="!sessionActive && !loading" class="card start-screen">
      <h2>Ready to Train?</h2>
      <div class="training-stats">
        <div class="stat-item">
          <span class="stat-label">Available for Training:</span>
          <span class="stat-value">{{ stats.availableForTraining }} cards</span>
        </div>
        <div v-if="estimatedTime" class="stat-item">
          <span class="stat-label">Estimated Time:</span>
          <span class="stat-value">{{ estimatedTime }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Total Cards:</span>
          <span class="stat-value">{{ stats.totalCards }}</span>
        </div>
      </div>
      <button v-if="stats.availableForTraining > 0" @click="startTraining" class="btn btn-primary">
        Start Training
      </button>
      <p v-if="stats.availableForTraining === 0" class="no-cards-message">
        No cards available for training. Add some words to your vocabulary first!
      </p>
    </div>

    <div v-if="loading" class="loading">Loading...</div>

    <!-- Network error notification -->
    <div v-if="networkError" class="network-error-notification">
      <div class="network-error-content">
        <span class="network-error-icon">⚠️</span>
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
          {{ option }}
        </button>
      </div>

      <div v-if="feedback" class="feedback-section">
        <div v-if="feedback.is_correct" class="feedback-badge feedback-success">
          <span class="feedback-icon">✓</span>
          <span class="feedback-text">{{ currentEncouragingPhrase }}</span>
        </div>
        
        <!-- For incorrect answers: show example first, then notification -->
        <template v-if="!feedback.is_correct">
          <div v-if="feedback.example" class="example">{{ feedback.example }}</div>
          <div class="feedback-badge feedback-error">
            <span class="feedback-icon">✗</span>
            <span class="feedback-text">{{ currentDisappointingPhrase }}</span>
          </div>
        </template>
        
        <!-- For correct answers: show example after notification -->
        <div v-if="feedback.is_correct && feedback.example" class="example">{{ feedback.example }}</div>
        
        <div v-if="waitingDelay" class="waiting-progress">
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
                :style="{ strokeDasharray: circumference, strokeDashoffset: strokeDashoffset }"
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { apiClient } from '../api/client'

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
const networkError = ref(false)
const networkErrorRetrying = ref(false)
const networkErrorAttempt = ref(0)
const networkErrorMaxAttempts = ref(3)

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

// Calculate progress for circular progress bar
const circumference = computed(() => {
  const radius = 34
  return 2 * Math.PI * radius
})

const strokeDashoffset = computed(() => {
  if (initialDelayMs.value === 0 || remainingMs.value <= 0) {
    return circumference.value
  }
  // Calculate progress based on remaining milliseconds for precision
  const progress = remainingMs.value / initialDelayMs.value
  return circumference.value * (1 - progress)
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
  
  await loadStats()
  await checkCurrentSession()
})

const loadStats = async () => {
  try {
    const data: { due_count: number; total_cards?: number; available_for_training?: number } = await apiClient.request('/app/dashboard')
    stats.value.dueCount = data.due_count || 0
    stats.value.totalCards = data.total_cards || 0
    stats.value.availableForTraining = data.available_for_training || data.due_count || 0
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

onUnmounted(() => {
  if (autoRevealTimer) {
    clearTimeout(autoRevealTimer)
    autoRevealTimer = null
  }
  if (autoNextCardTimer) {
    clearTimeout(autoNextCardTimer)
    autoNextCardTimer = null
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
      alert('No cards available for training. Request some words first!')
    } else {
      console.error('Failed to start training:', error)
      alert('Failed to start training')
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
      alert('Не удалось загрузить варианты ответов. Попробуйте обновить страницу.')
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
      
      const startTime = Date.now()
      const endTime = startTime + delayMs
      
      // Update countdown with precise timing using requestAnimationFrame
      let animationFrameId: number
      const updateCountdown = () => {
        const now = Date.now()
        const currentRemainingMs = Math.max(0, endTime - now)
        const currentRemainingSeconds = Math.ceil(currentRemainingMs / 1000)
        
        remainingMs.value = currentRemainingMs
        delaySeconds.value = currentRemainingSeconds
        
        if (currentRemainingMs > 0) {
          animationFrameId = requestAnimationFrame(updateCountdown)
        } else {
          delaySeconds.value = 0
          remainingMs.value = 0
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          if (animationFrameId) {
            cancelAnimationFrame(animationFrameId)
          }
          nextCard()
        }
      }
      
      // Start updating immediately
      animationFrameId = requestAnimationFrame(updateCountdown)
      
      // Schedule automatic next card as backup
      autoNextCardTimer = setTimeout(() => {
        if (animationFrameId) {
          cancelAnimationFrame(animationFrameId)
        }
        if (waitingDelay.value) {
          waitingDelay.value = false
          initialDelaySeconds.value = 0
          initialDelayMs.value = 0
          delaySeconds.value = 0
          remainingMs.value = 0
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
      alert('Не удалось отправить ответ. Попробуйте еще раз.')
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
        alert('Не удалось загрузить следующую карточку. Попробуйте обновить страницу.')
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
  
  // Refresh stats
  await loadStats()
}
</script>

<style scoped>
.training {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
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
  animation: correct-pulse 0.5s ease-out;
}

.option-btn.option-incorrect {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: 2px solid #ef4444;
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
  animation: incorrect-shake 0.5s ease-out;
}

@keyframes correct-pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
  100% {
    transform: scale(1);
  }
}

@keyframes incorrect-shake {
  0%, 100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-5px);
  }
  75% {
    transform: translateX(5px);
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
  animation: feedback-appear 0.3s ease-out;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

@keyframes feedback-appear {
  from {
    opacity: 0;
    transform: scale(0.8) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.feedback-success {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.feedback-error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
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
}

.feedback-text {
  letter-spacing: 0.5px;
}

.example {
  font-style: italic;
  margin: 20px 0;
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
}

.start-screen h2 {
  margin-bottom: 20px;
}

.training-stats {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin: 20px 0;
  padding: 20px;
  background: var(--bg-secondary, rgba(0, 0, 0, 0.05));
  border-radius: 8px;
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
}

.completion-screen h2 {
  margin-bottom: 15px;
  color: var(--color-success, #10b981);
}

.completion-stats {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin: 20px 0;
  padding: 20px;
  background: var(--bg-secondary, rgba(0, 0, 0, 0.05));
  border-radius: 8px;
}

.completion-stats .stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid var(--border-primary, rgba(0, 0, 0, 0.1));
}

.completion-stats .stat-item:last-child {
  border-bottom: none;
}

.completion-stats .stat-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.completion-stats .stat-value {
  font-size: 24px;
  font-weight: bold;
  color: var(--color-primary);
}

.completion-stats .stat-value.correct-stat {
  color: var(--color-success, #10b981);
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
</style>

