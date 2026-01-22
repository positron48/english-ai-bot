<template>
  <div class="grammar-test">
    <div v-if="loading" class="loading">
      <p>Loading test...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadTest" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else-if="testSubmitted" class="test-results">
      <div class="results-header">
        <h1>Test Results</h1>
        <div class="score-display-wrapper">
          <div class="score-display" :class="{ 'passed': result.passed, 'failed': !result.passed }">
            <div class="score-value">{{ animatedScore }}%</div>
            <div class="score-label">{{ result.passed ? 'Passed' : 'Failed' }}</div>
            
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
        </div>
      </div>
      
      <div class="results-summary">
        <p>You answered <strong>{{ result.correct }}</strong> out of <strong>{{ result.total }}</strong> questions correctly.</p>
        <p v-if="!result.passed" class="retry-hint">
          You need at least 50% to pass. Try again to improve your score!
        </p>
      </div>
      
      <div class="results-details">
        <h2>Detailed Results</h2>
        <div
          v-for="(item, index) in result.results"
          :key="index"
          class="result-item"
          :class="{ 'correct': item.correct, 'incorrect': !item.correct }"
        >
          <div class="result-header">
            <span class="result-number">Question {{ index + 1 }}</span>
            <span class="result-status" :class="{ 'correct': item.correct, 'incorrect': !item.correct }">
              {{ item.correct ? '✓ Correct' : '✗ Incorrect' }}
            </span>
          </div>
          
          <!-- Question -->
          <div class="result-question-prompt">
            <strong>Question:</strong>
            <div v-html="renderMarkdown(item.prompt || getQuestionPrompt(item.question_id))"></div>
          </div>
          
          <!-- User Answer -->
          <div class="result-user-answer">
            <strong>Your Answer:</strong>
            <div class="answer-display">{{ formatAnswer(item.question_id, item.user_answer) || '(not answered)' }}</div>
          </div>
          
          <!-- Correct Answer (only if incorrect) -->
          <div v-if="!item.correct" class="result-correct-answer">
            <strong>Correct Answer:</strong>
            <div class="answer-display">{{ formatAnswer(item.question_id, item.correct_answer) }}</div>
          </div>
          
          <!-- Hint/Feedback for incorrect answers -->
          <div v-if="!item.correct" class="result-hint">
            <div v-if="getChoiceFeedback(item.question_id, item.user_answer)" class="choice-feedback">
              <strong>Hint:</strong>
              <div v-html="renderMarkdown(getChoiceFeedback(item.question_id, item.user_answer))"></div>
            </div>
          </div>
          
          <!-- Explanation -->
          <div v-if="item.explanation" class="result-explanation">
            <strong>Explanation:</strong>
            <div v-html="renderMarkdown(item.explanation)"></div>
          </div>
        </div>
      </div>
      
      <div class="results-actions">
        <button @click="goBack" class="btn btn-secondary">Back to Chapter</button>
        <button v-if="!result.passed" @click="retryTest" class="btn btn-primary">Retry Test</button>
      </div>
    </div>
    
    <div v-else class="test-content">
      <div class="test-header">
        <h1>{{ testTitle }}</h1>
        <div class="test-progress">
          Question {{ currentQuestionIndex + 1 }} of {{ questions.length }}
        </div>
        <button @click="exitTest" class="btn btn-secondary btn-exit">Exit Test</button>
      </div>
      
      <div class="test-questions">
        <GrammarQuestion
          v-if="currentQuestion"
          :key="currentQuestion.id || currentQuestionIndex"
          :ref="el => setQuestionRef(currentQuestionIndex, el)"
          :question="currentQuestion"
          :show-answers="false"
          :initial-answer="answers.get(currentQuestionIndex)"
          @answer="handleAnswerWithAutoNext(currentQuestionIndex, $event)"
        />
      </div>
      
      <div class="test-navigation">
        <button 
          v-if="currentQuestionIndex > 0"
          @click="previousQuestion"
          class="btn btn-secondary"
        >
          Previous
        </button>
        <button 
          v-if="currentQuestionIndex < questions.length - 1"
          @click="nextQuestion"
          :disabled="!hasAnswer(currentQuestionIndex)"
          class="btn btn-primary"
        >
          Next
        </button>
        <button 
          v-else
          @click="submitTest"
          :disabled="submitting || !hasAnswer(currentQuestionIndex)"
          class="btn btn-primary"
        >
          {{ submitting ? 'Submitting...' : 'Submit Test' }}
        </button>
      </div>
    </div>
    
    <!-- Exit confirmation modal -->
    <ConfirmModal
      :visible="showExitConfirm"
      message="Are you sure you want to exit the test? Your progress will be lost."
      @confirm="handleExitConfirm"
      @cancel="showExitConfirm = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import { apiClient } from '../api/client'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'

const route = useRoute()
const router = useRouter()

const { settings } = useSettings()
const { playVictory, playDefeat } = useAudio()

const scope = computed(() => {
  if (route.path.includes('/chapter/')) {
    return 'chapter'
  }
  return 'category'
})

const scopeId = computed(() => {
  if (scope.value === 'chapter') {
    return route.params.chapterId as string
  }
  return route.params.sectionId as string
})

const testTitle = computed(() => {
  if (scope.value === 'chapter') {
    return 'Chapter Test'
  }
  return 'Category Test'
})

const questions = ref<any[]>([])
const answers = ref<Map<number, any>>(new Map())
const currentQuestionIndex = ref(0)
const loading = ref(true)
const error = ref<string | null>(null)
const submitting = ref(false)
const testSubmitted = ref(false)
const result = ref<any>(null)
const questionRefs = ref<any[]>([])
const showExitConfirm = ref(false)
const animatedScore = ref(0)
const percentageAnimationComplete = ref(false)

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value] || null
})

// Calculate accuracy percentage
const accuracyPercentage = computed(() => {
  if (!result.value || result.value.total === 0) return 0
  return Math.round((result.value.correct / result.value.total) * 100)
})

const setQuestionRef = (index: number, el: any) => {
  if (el) {
    questionRefs.value[index] = el
  }
}

const hasAnswer = (index: number): boolean => {
  const answer = answers.value.get(index)
  if (answer === undefined || answer === null) {
    return false
  }
  // For arrays (mcq_multi), check if at least one option is selected
  if (Array.isArray(answer)) {
    return answer.length > 0
  }
  // For strings, check if not empty
  if (typeof answer === 'string') {
    return answer.trim().length > 0
  }
  // For other types (numbers, booleans), consider them valid
  return true
}

const loadTest = async () => {
  loading.value = true
  error.value = null
  try {
    if (scope.value === 'chapter') {
      const data: { questions: any[]; total: number } = await apiClient.request(
        `/api/learning/grammar/chapters/${scopeId.value}/test`
      )
      questions.value = data.questions || []
    } else {
      // Category test
      const data: { questions: any[]; total: number } = await apiClient.request(
        `/api/learning/grammar/categories/${scopeId.value}/test`
      )
      questions.value = data.questions || []
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load test'
    console.error('Failed to load grammar test:', err)
  } finally {
    loading.value = false
  }
}

const handleAnswer = (index: number, answer: any) => {
  answers.value.set(index, answer)
}

const handleAnswerWithAutoNext = (index: number, answer: any) => {
  answers.value.set(index, answer)
  
  if (hasAnswer(index)) {
    const isLast = index === questions.value.length - 1
    const isFillBlank = questions.value[index]?.type === 'fill_blank'
    if (isLast && isFillBlank) {
      submitTest()
    } else {
      setTimeout(() => {
        if (index < questions.value.length - 1) nextQuestion()
      }, 300)
    }
  }
}

const nextQuestion = () => {
  if (currentQuestionIndex.value < questions.value.length - 1 && hasAnswer(currentQuestionIndex.value)) {
    currentQuestionIndex.value++
  }
}

const previousQuestion = () => {
  if (currentQuestionIndex.value > 0) {
    currentQuestionIndex.value--
  }
}

const exitTest = () => {
  showExitConfirm.value = true
}

const handleExitConfirm = () => {
  showExitConfirm.value = false
  goBack()
}

const animateScore = (targetScore: number) => {
  animatedScore.value = 0
  percentageAnimationComplete.value = false
  const duration = 1500 // 1.5 seconds
  const startTime = Date.now()
  
  const animate = () => {
    const elapsed = Date.now() - startTime
    const progress = Math.min(elapsed / duration, 1)
    // Easing function (ease-out cubic) - matches CSS cubic-bezier(0.4, 0, 0.2, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    animatedScore.value = Math.round(targetScore * eased)
    
    if (progress < 1) {
      requestAnimationFrame(animate)
    } else {
      animatedScore.value = targetScore
      percentageAnimationComplete.value = true
    }
  }
  
  requestAnimationFrame(animate)
}

const submitTest = async () => {
  // Build answers map and preserve question order
  const answersMap: Record<string, any> = {}
  const questionIds: string[] = []
  questions.value.forEach((q, index) => {
    const answer = answers.value.get(index)
    if (answer !== undefined && q.id) {
      answersMap[q.id] = answer
      questionIds.push(q.id)
    }
  })
  
  submitting.value = true
  try {
    const data: { score: number; passed: boolean; correct: number; total: number; results: any[] } = 
      await apiClient.request('/api/learning/grammar/tests/submit', {
        method: 'POST',
        body: {
          scope: scope.value,
          scope_id: scopeId.value,
          answers: answersMap,
          question_ids: questionIds
        }
      })
    
    result.value = data
    testSubmitted.value = true
    
    // Trigger haptic feedback
    triggerHapticFeedback(data.passed)
    
    // Start score animation after DOM update
    // Sounds will be played by watch when animation completes
    nextTick(() => {
      animateScore(data.score)
    })
  } catch (err: any) {
    error.value = err.message || 'Failed to submit test'
    console.error('Failed to submit grammar test:', err)
  } finally {
    submitting.value = false
  }
}

// Confetti styles - start from random points on circle edge
const getConfettiStyle = (index: number) => {
  const colors = ['#10b981', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
  const color = colors[index % colors.length]
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: approximate 90px from center
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

// Firework styles - explode from random points on circle edge
const getFireworkStyle = (index: number) => {
  // Random angle on circle edge
  const startAngle = Math.random() * 360
  const startAngleRad = (startAngle * Math.PI) / 180
  // Circle radius: approximate 90px from center
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

const goBack = () => {
  if (scope.value === 'chapter') {
    router.push(`/learning/grammar/chapter/${scopeId.value}`)
  } else {
    router.push(`/learning/grammar/${scopeId.value}`)
  }
}

const retryTest = () => {
  testSubmitted.value = false
  answers.value.clear()
  currentQuestionIndex.value = 0
  animatedScore.value = 0
  percentageAnimationComplete.value = false
  loadTest()
}

const renderMarkdown = (text: string): string => {
  if (!text) return ''
  try {
    marked.setOptions({
      breaks: true,
      gfm: true,
    })
    return marked.parse(text) as string
  } catch (error) {
    return text
  }
}

// Helper function to get question by ID
const getQuestionById = (questionId: string): any => {
  return questions.value.find(q => q.id === questionId) || null
}

// Helper function to get question prompt
const getQuestionPrompt = (questionId: string): string => {
  const question = getQuestionById(questionId)
  return question?.prompt || ''
}

// Helper function to format answer for display
const formatAnswer = (questionId: string, answer: any): string => {
  if (answer === undefined || answer === null) {
    return ''
  }
  
  const question = getQuestionById(questionId)
  if (!question) {
    return String(answer)
  }
  
  const questionType = question.type
  
  switch (questionType) {
    case 'mcq_single':
    case 'true_false':
    case 'error_spotting':
      // Find the choice text for the answer ID
      if (question.choices && Array.isArray(question.choices)) {
        const choice = question.choices.find((c: any) => c.id === answer)
        if (choice) {
          return choice.text
        }
      }
      // Fallback for true/false
      if (questionType === 'true_false') {
        return (answer === 'true' || answer === true) ? 'Да' : 'Нет'
      }
      return String(answer)
    
    case 'mcq_multi':
      // Multiple choice - format as comma-separated list
      if (Array.isArray(answer)) {
        if (question.choices && Array.isArray(question.choices)) {
          const choiceTexts = answer
            .map((id: string) => {
              const choice = question.choices.find((c: any) => c.id === id)
              return choice ? choice.text : id
            })
            .filter(Boolean)
          return choiceTexts.join(', ')
        }
        return answer.join(', ')
      }
      return String(answer)
    
    case 'fill_blank':
    case 'reorder':
      return String(answer)
    
    default:
      return String(answer)
  }
}

// Helper function to get choice feedback/hint
const getChoiceFeedback = (questionId: string, userAnswer: any): string => {
  const question = getQuestionById(questionId)
  if (!question || !question.choices) {
    return ''
  }
  
  // For single choice questions, find the selected choice's feedback
  if (question.type === 'mcq_single' || question.type === 'error_spotting') {
    if (question.choices && Array.isArray(question.choices)) {
      const choice = question.choices.find((c: any) => c.id === userAnswer)
      return choice?.feedback || ''
    }
  }
  
  // For multi-choice, we could show feedback for each selected incorrect choice
  // For now, return empty string
  return ''
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

onMounted(() => {
  loadTest()
})
</script>

<style scoped>
.grammar-test {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error {
  text-align: center;
  padding: 40px 20px;
}

.error {
  color: var(--color-danger);
}

.test-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--border-primary);
  gap: 16px;
}

.btn-exit {
  flex-shrink: 0;
}

.test-header h1 {
  margin: 0;
}

.test-progress {
  color: var(--text-secondary);
  font-size: 14px;
}

.test-questions {
  margin-bottom: 32px;
}

.test-navigation {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding-top: 24px;
  border-top: 2px solid var(--border-primary);
}

.test-results {
  max-width: 800px;
  margin: 0 auto;
}

.results-header {
  text-align: center;
  margin-bottom: 32px;
}

.score-display-wrapper {
  position: relative;
  display: inline-block;
  width: 100%;
  max-width: 400px;
}

.score-display {
  margin: 24px 0;
  padding: 32px;
  border-radius: 12px;
  background: var(--card-bg);
  border: 3px solid;
  position: relative;
  overflow: visible;
}

.score-display.passed {
  border-color: var(--color-success);
  background: rgba(40, 167, 69, 0.1);
}

.score-display.failed {
  border-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.1);
}

.score-value {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 8px;
  transition: transform 0.1s ease-out;
}

.score-display.passed .score-value {
  color: var(--color-success);
}

.score-display.failed .score-value {
  color: var(--color-danger);
}

.score-label {
  font-size: 18px;
  font-weight: 600;
}

.results-summary {
  text-align: center;
  margin-bottom: 32px;
  padding: 20px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.retry-hint {
  margin-top: 12px;
  color: var(--text-secondary);
  font-size: 14px;
}

.results-details {
  margin-bottom: 32px;
}

.results-details h2 {
  margin-bottom: 20px;
}

.result-item {
  padding: 16px;
  margin-bottom: 16px;
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
}

.result-item.correct {
  border-color: var(--color-success);
  background: rgba(40, 167, 69, 0.05);
}

.result-item.incorrect {
  border-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.05);
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.result-number {
  font-weight: 600;
  color: var(--text-primary);
}

.result-status {
  font-weight: 600;
}

.result-status.correct {
  color: var(--color-success);
}

.result-status.incorrect {
  color: var(--color-danger);
}

.result-question-prompt {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
}

.result-question-prompt strong {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.result-question-prompt :deep(p) {
  margin: 0;
  line-height: 1.6;
}

.result-user-answer {
  margin-bottom: 12px;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border-left: 3px solid var(--color-primary);
}

.result-user-answer strong {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.result-correct-answer {
  margin-bottom: 12px;
  padding: 12px;
  background: rgba(40, 167, 69, 0.1);
  border-radius: 6px;
  border-left: 3px solid var(--color-success);
}

.result-correct-answer strong {
  display: block;
  margin-bottom: 8px;
  color: var(--color-success);
}

.result-hint {
  margin-bottom: 12px;
}

.choice-feedback {
  padding: 12px;
  background: rgba(255, 193, 7, 0.1);
  border-radius: 6px;
  border-left: 3px solid rgba(255, 193, 7, 0.5);
}

.choice-feedback strong {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.choice-feedback :deep(p) {
  margin: 0;
  line-height: 1.6;
}

.answer-display {
  color: var(--text-primary);
  font-weight: 500;
  line-height: 1.6;
}

.result-explanation {
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
  color: var(--text-secondary);
  line-height: 1.6;
}

.result-explanation strong {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.results-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  padding-top: 24px;
  border-top: 2px solid var(--border-primary);
}

.btn {
  padding: 12px 24px;
  border-radius: 8px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 2px solid var(--border-primary);
}

.btn-secondary:hover {
  border-color: var(--color-primary);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Celebration animations for >90% */
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
</style>
