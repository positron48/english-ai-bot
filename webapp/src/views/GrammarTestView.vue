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
        <div class="score-display" :class="{ 'passed': result.passed, 'failed': !result.passed }">
          <div class="score-value">{{ result.score }}%</div>
          <div class="score-label">{{ result.passed ? 'Passed' : 'Failed' }}</div>
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
          <div v-if="item.explanation" class="result-explanation" v-html="renderMarkdown(item.explanation)"></div>
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
          @answer="handleAnswer(currentQuestionIndex, $event)"
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import { apiClient } from '../api/client'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()

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

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value] || null
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
      // Category test - to be implemented
      error.value = 'Category tests not yet implemented'
      return
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

const submitTest = async () => {
  // Build answers map
  const answersMap: Record<string, any> = {}
  questions.value.forEach((q, index) => {
    const answer = answers.value.get(index)
    if (answer !== undefined && q.id) {
      answersMap[q.id] = answer
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
          answers: answersMap
        }
      })
    
    result.value = data
    testSubmitted.value = true
  } catch (err: any) {
    error.value = err.message || 'Failed to submit test'
    console.error('Failed to submit grammar test:', err)
  } finally {
    submitting.value = false
  }
}

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

.score-display {
  margin: 24px 0;
  padding: 32px;
  border-radius: 12px;
  background: var(--card-bg);
  border: 3px solid;
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

.result-explanation {
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
  color: var(--text-secondary);
  line-height: 1.6;
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
</style>
