<template>
  <div class="training">
    <h1>Training</h1>
    
    <div v-if="!sessionActive && !loading" class="card">
      <button @click="startTraining" class="btn btn-primary">Start Training</button>
    </div>

    <div v-if="loading" class="loading">Loading...</div>

    <div v-if="sessionActive && currentCard" class="card">
      <div class="training-progress">
        <p>Card {{ cardIndex }} of {{ totalCards }}</p>
      </div>

      <div class="question" v-html="currentCard.question"></div>

      <div v-if="optionsShown && !feedback" class="options">
        <button
          v-for="(option, index) in options"
          :key="index"
          @click="submitAnswer(index)"
          class="btn option-btn"
          :disabled="answering"
        >
          {{ option }}
        </button>
      </div>

      <div v-if="feedback" class="feedback">
        <p :class="feedback.is_correct ? 'success' : 'error'">
          {{ feedback.is_correct ? 'Correct!' : 'Wrong!' }}
        </p>
        <p v-if="!feedback.is_correct">
          Correct answer: {{ feedback.correct_answer }}
        </p>
        <p v-if="feedback.example" class="example">{{ feedback.example }}</p>
        <p v-if="waitingDelay" class="waiting-message">Wait {{ delaySeconds }}s...</p>
      </div>

      <div v-if="sessionComplete" class="card">
        <h2>Training Complete!</h2>
        <p>You completed {{ cardsCompleted }} cards.</p>
        <button @click="resetSession" class="btn btn-primary">Start New Training</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
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
const answering = ref(false)
const waitingDelay = ref(false)
const delaySeconds = ref(0)
const sessionComplete = ref(false)
const cardsCompleted = ref(0)

const cardIndex = ref(0)
const totalCards = ref(0)
const userCardId = ref(0)

// Timer for automatic options reveal
let autoRevealTimer: ReturnType<typeof setTimeout> | null = null
// Timer for automatic next card transition
let autoNextCardTimer: ReturnType<typeof setTimeout> | null = null
const cardShownAt = ref<Date | null>(null)

onMounted(async () => {
  await checkCurrentSession()
})

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
  } catch (error) {
    console.error('Failed to reveal options:', error)
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
    
    // Schedule automatic transition to next card
    const delayMs = data.delay_seconds ? data.delay_seconds * 1000 : 0
    
    if (delayMs > 0) {
      delaySeconds.value = data.delay_seconds!
      waitingDelay.value = true
      
      // Update countdown
      const interval = setInterval(() => {
        delaySeconds.value--
        if (delaySeconds.value <= 0) {
          clearInterval(interval)
          waitingDelay.value = false
        }
      }, 1000)
      
      // Schedule automatic next card
      autoNextCardTimer = setTimeout(() => {
        nextCard()
      }, delayMs)
    } else {
      // No delay, go to next card immediately
      autoNextCardTimer = setTimeout(() => {
        nextCard()
      }, 1000) // Small delay to show feedback
    }
  } catch (error) {
    console.error('Failed to submit answer:', error)
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
  cardShownAt.value = null

  try {
    const card: Card = await apiClient.request('/app/training/current')
    setupCard(card)
    
    if (card.card_index > card.total_cards) {
      sessionComplete.value = true
      cardsCompleted.value = card.card_index - 1
    }
  } catch (error: any) {
    if (error.message?.includes('404')) {
      sessionComplete.value = true
    } else {
      console.error('Failed to get next card:', error)
    }
  }
}

const resetSession = () => {
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
  sessionComplete.value = false
  cardsCompleted.value = 0
  cardShownAt.value = null
}
</script>

<style scoped>
.training-progress {
  margin-bottom: 20px;
  text-align: center;
}

.question {
  font-size: 24px;
  margin: 30px 0;
  text-align: center;
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
}

.feedback {
  margin-top: 20px;
  text-align: center;
}

.example {
  font-style: italic;
  margin: 10px 0;
  padding: 10px;
  background: var(--example-bg);
  border-radius: 4px;
  color: var(--text-primary);
}

.waiting-message {
  margin-top: 20px;
  font-size: 18px;
  color: var(--training-progress-text);
}
</style>

