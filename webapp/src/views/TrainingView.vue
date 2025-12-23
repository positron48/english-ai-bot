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

      <div v-if="!optionsShown && !feedback" class="card-actions">
        <button @click="revealOptions" class="btn btn-primary">Show Options</button>
      </div>

      <div v-if="optionsShown && !feedback" class="options">
        <button
          v-for="(option, index) in currentCard.options"
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
        <button
          @click="nextCard"
          class="btn btn-primary"
          :disabled="waitingDelay"
        >
          {{ waitingDelay ? `Wait ${delaySeconds}s...` : 'Next Card' }}
        </button>
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
import { ref, onMounted } from 'vue'
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

onMounted(async () => {
  await checkCurrentSession()
})

const checkCurrentSession = async () => {
  try {
    const card: Card = await apiClient.request('/app/training/current')
    sessionActive.value = true
    currentCard.value = card
    cardIndex.value = card.card_index
    totalCards.value = card.total_cards
    userCardId.value = card.user_card_id
  } catch (error: any) {
    if (!error.message?.includes('404')) {
      console.error('Failed to check session:', error)
    }
  }
}

const startTraining = async () => {
  loading.value = true
  try {
    const card: Card = await apiClient.request('/app/training/start', { method: 'POST' })
    sessionActive.value = true
    currentCard.value = card
    cardIndex.value = card.card_index
    totalCards.value = card.total_cards
    userCardId.value = card.user_card_id
    optionsShown.value = false
    feedback.value = null
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

const revealOptions = async () => {
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
    
    if (data.delay_seconds) {
      delaySeconds.value = data.delay_seconds
      waitingDelay.value = true
      const interval = setInterval(() => {
        delaySeconds.value--
        if (delaySeconds.value <= 0) {
          clearInterval(interval)
          waitingDelay.value = false
        }
      }, 1000)
    }
  } catch (error) {
    console.error('Failed to submit answer:', error)
  } finally {
    answering.value = false
  }
}

const nextCard = async () => {
  if (waitingDelay.value) return

  feedback.value = null
  optionsShown.value = false
  options.value = []

  try {
    const card: Card = await apiClient.request('/app/training/current')
    currentCard.value = card
    cardIndex.value = card.card_index
    totalCards.value = card.total_cards
    userCardId.value = card.user_card_id
    
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
  sessionActive.value = false
  currentCard.value = null
  optionsShown.value = false
  feedback.value = null
  sessionComplete.value = false
  cardsCompleted.value = 0
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

.card-actions {
  text-align: center;
  margin: 20px 0;
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
  background: #f0f0f0;
  border-radius: 4px;
}
</style>

