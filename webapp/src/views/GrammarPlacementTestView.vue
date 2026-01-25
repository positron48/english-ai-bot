<template>
  <div class="grammar-placement-test">
    <div v-if="loading" class="loading">
      <p>Loading placement test...</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadTest" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else-if="testSubmitted" class="test-results">
      <div class="results-header">
        <h1>Placement Test Results</h1>
        <div class="score-display-wrapper">
          <div class="score-display" :class="{ 'passed': result && result.opened_sections && result.opened_sections.length > 0 }">
            <div class="score-value level-value">{{ result?.level || '—' }}</div>
            <div class="score-label">Your Level</div>
            <div class="test-score-secondary">Test score: {{ result && result.score !== undefined ? animatedScore : 0 }}%</div>
          </div>
        </div>
      </div>
      
      <div class="results-summary">
        <p>You answered <strong>{{ result?.correct || 0 }}</strong> out of <strong>{{ result?.total_questions || 0 }}</strong> questions correctly.</p>
        <div v-if="result && result.opened_sections && result.opened_sections.length > 0" class="opened-sections">
          <h3>Opened for you:</h3>
          <ul>
            <li v-for="sectionId in result.opened_sections" :key="sectionId">
              {{ getSectionName(sectionId) }}
            </li>
          </ul>
          <p class="info-text">
            All chapters in these sections are now available for you to study!
          </p>
        </div>
        <div v-else class="no-sections-opened">
          <p>Keep practicing to unlock more sections!</p>
        </div>
      </div>
      
      <div v-if="result?.results && result.results.length > 0" class="results-details">
        <div class="results-scale" aria-label="Question overview">
          <div class="scale-level-labels">
            <span
              v-for="(group, gIdx) in scaleLevelGroups"
              :key="'l-' + gIdx"
              class="scale-level-label"
              :style="{ left: group.labelLeft + '%', width: group.labelWidth + '%' }"
            >{{ group.level }}</span>
          </div>
          <div class="scale-line-wrap">
            <div class="scale-line"></div>
            <div
              v-for="(pos, i) in scaleSeparators"
              :key="'sep-' + i"
              class="scale-sep"
              :style="{ left: pos + '%' }"
            />
            <button
              v-for="(item, index) in result.results"
              :key="'dot-' + (item.question_id || index)"
              type="button"
              class="scale-dot"
              :class="{ 'correct': item.correct, 'incorrect': !item.correct }"
              :style="{ left: scaleDotPosition(index) + '%' }"
              :title="`Question ${index + 1}${item.level ? ' · ' + item.level : ''}`"
              @click="scrollToResult(index)"
            />
          </div>
        </div>
        <h2>Detailed Results</h2>
        <div
          v-for="(item, index) in result.results"
          :key="item.question_id || index"
          :id="`result-${index}`"
          class="result-item"
          :class="{ 'correct': item.correct, 'incorrect': !item.correct, 'clickable': item.correct }"
          role="button"
          :tabindex="item.correct ? 0 : -1"
          :aria-expanded="item.correct ? String(isResultExpanded(item, index)) : null"
          @click="toggleResult(item, index)"
          @keydown.enter.prevent="toggleResult(item, index)"
          @keydown.space.prevent="toggleResult(item, index)"
        >
          <div class="result-header">
            <span class="result-number">Question {{ index + 1 }}</span>
            <span v-if="item.placement_chapter_title" class="result-chapter">From: {{ item.placement_chapter_title }}</span>
            <span class="result-status" :class="{ 'correct': item.correct, 'incorrect': !item.correct }">
              {{ item.correct ? '✓ Correct' : '✗ Incorrect' }}
            </span>
            <span v-if="item.correct" class="result-toggle" aria-hidden="true">
              <span class="result-chevron" :class="{ 'expanded': isResultExpanded(item, index) }"></span>
            </span>
          </div>
          
          <div v-if="isResultExpanded(item, index)" class="result-body">
            <div class="result-question-prompt">
              <strong>Question:</strong>
              <div v-html="renderMarkdown(getQuestionPrompt(item.question_id))"></div>
            </div>
            
            <div class="result-user-answer" :class="{ 'result-user-answer--correct': item.correct }">
              <strong>Your Answer:</strong>
              <div class="answer-display">{{ formatAnswer(item.question_id, item.user_answer) || '(not answered)' }}</div>
            </div>
            
            <div v-if="!item.correct && item.correct_answer != null" class="result-correct-answer">
              <strong>Correct Answer:</strong>
              <div class="answer-display">{{ formatAnswer(item.question_id, item.correct_answer) }}</div>
            </div>
            
            <div v-if="!item.correct && getChoiceFeedback(item.question_id, item.user_answer)" class="result-hint">
              <div class="choice-feedback">
                <strong>Hint:</strong>
                <div v-html="renderMarkdown(getChoiceFeedback(item.question_id, item.user_answer))"></div>
              </div>
            </div>
            
            <div v-if="item.explanation" class="result-explanation">
              <strong>Explanation:</strong>
              <div v-html="renderMarkdown(item.explanation)"></div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="results-actions">
        <button @click="goToGrammar" class="btn btn-primary">Go to Grammar Course</button>
        <button @click="retryTest" class="btn btn-secondary">Retry Test</button>
      </div>
    </div>
    
    <div v-else-if="questions.length === 0" class="error">
      <p>No questions available for the placement test.</p>
      <button @click="loadTest" class="btn btn-primary">Retry</button>
    </div>
    
    <div v-else class="test-content">
      <div class="test-header">
        <h1>Placement Test</h1>
        <p class="test-description">
          This test will help determine your current grammar level. 
          Answer 25 questions from different topics, and we'll unlock the appropriate sections for you.
        </p>
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
          :chapter-title="currentQuestion.placement_chapter_title"
          @answer="handleAnswerWithAutoNext(currentQuestionIndex, $event)"
        />
        <div v-else class="loading">
          <p>Loading question...</p>
        </div>
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
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { marked } from 'marked'
import { apiClient } from '../api/client'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import ConfirmModal from '../components/ConfirmModal.vue'

const router = useRouter()

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
const sectionNames = ref<Record<string, string>>({})
const expandedCorrectResults = ref<Record<string, boolean>>({})

const currentQuestion = computed(() => {
  const question = questions.value[currentQuestionIndex.value] || null
  if (question) {
    console.log('Current question:', question.id || currentQuestionIndex.value, question.type)
  }
  return question
})

const scaleLevelGroups = computed(() => {
  const res = result.value?.results
  if (!res || !Array.isArray(res) || res.length === 0) return []
  const total = res.length
  const groups: { level: string; labelLeft: number; labelWidth: number }[] = []
  let i = 0
  while (i < total) {
    const raw = (res[i] as any)?.level || '—'
    const levelKey = raw.toString().toLowerCase()
    let j = i
    while (j < total && ((res[j] as any)?.level || '—').toString().toLowerCase() === levelKey) j++
    const count = j - i
    const levelDisplay = levelKey === '—' ? '—' : levelKey.toUpperCase()
    groups.push({
      level: levelDisplay,
      labelLeft: (i / total) * 100,
      labelWidth: (count / total) * 100
    })
    i = j
  }
  return groups
})

const scaleSeparators = computed(() => {
  const res = result.value?.results
  if (!res || res.length < 2) return []
  const out: number[] = []
  const total = res.length
  for (let i = 1; i < total; i++) {
    const prev = ((res[i - 1] as any)?.level || '').toString().toLowerCase()
    const curr = ((res[i] as any)?.level || '').toString().toLowerCase()
    if (prev !== curr) out.push((i / total) * 100)
  }
  return out
})

const scaleDotPosition = (index: number): number => {
  const res = result.value?.results
  if (!res || res.length === 0) return 0
  const total = res.length
  return total > 1 ? ((index + 0.5) / total * 100) : 50
}

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
  if (Array.isArray(answer)) {
    return answer.length > 0
  }
  if (typeof answer === 'string') {
    return answer.trim().length > 0
  }
  return true
}

const loadTest = async () => {
  loading.value = true
  error.value = null
  try {
    // Load section names for display (don't fail if this fails)
    loadSectionNames().catch(err => {
      console.warn('Failed to load section names:', err)
    })
    
    const data: { questions: any[]; total: number } = await apiClient.request(
      '/api/learning/grammar/placement-test'
    )
    
    console.log('Placement test data:', data)
    
    questions.value = data.questions || []
    
    if (questions.value.length === 0) {
      error.value = 'No questions available for the placement test. Please try again later.'
    } else {
      console.log(`Loaded ${questions.value.length} questions for placement test`)
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load placement test'
    console.error('Failed to load placement test:', err)
  } finally {
    loading.value = false
  }
}

const loadSectionNames = async () => {
  try {
    const response: { categories: Array<{ section_id: string; title: string }> } = await apiClient.request(
      '/api/learning/grammar/categories'
    )
    const names: Record<string, string> = {}
    if (response.categories && Array.isArray(response.categories)) {
      response.categories.forEach((item: any) => {
        if (item.section_id) {
          names[item.section_id] = item.title || item.section_id
        }
      })
    }
    sectionNames.value = names
  } catch (err) {
    console.warn('Failed to load section names:', err)
  }
}

const getSectionName = (sectionId: string): string => {
  return sectionNames.value[sectionId] || sectionId
}

const scrollToResult = (index: number) => {
  document.getElementById(`result-${index}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const getQuestionById = (questionId: string): any => {
  return questions.value.find((q: any) => q.id === questionId) || null
}

const getQuestionPrompt = (questionId: string): string => {
  const q = getQuestionById(questionId)
  return q?.prompt || ''
}

const formatAnswer = (questionId: string, answer: any): string => {
  if (answer === undefined || answer === null) return ''
  const q = getQuestionById(questionId)
  if (!q) return String(answer)
  const t = q.type
  if (t === 'mcq_single' || t === 'true_false' || t === 'error_spotting') {
    if (q.choices?.length) {
      const c = q.choices.find((x: any) => x.id === answer)
      if (c) return c.text
    }
    if (t === 'true_false') return (answer === 'true' || answer === true) ? 'Да' : 'Нет'
    return String(answer)
  }
  if (t === 'mcq_multi' && Array.isArray(answer)) {
    if (q.choices?.length) {
      return answer.map((id: string) => q.choices.find((x: any) => x.id === id)?.text || id).filter(Boolean).join(', ')
    }
    return answer.join(', ')
  }
  return String(answer)
}

const getChoiceFeedback = (questionId: string, userAnswer: any): string => {
  const q = getQuestionById(questionId)
  if (!q?.choices || (q.type !== 'mcq_single' && q.type !== 'error_spotting')) return ''
  const c = q.choices.find((x: any) => x.id === userAnswer)
  return c?.feedback || ''
}

const renderMarkdown = (text: string): string => {
  if (!text) return ''
  try {
    return marked.parse(text) as string
  } catch {
    return text
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
  goToGrammar()
}

const animateScore = (targetScore: number) => {
  animatedScore.value = 0
  const duration = 1500
  const startTime = Date.now()
  
  const animate = () => {
    const elapsed = Date.now() - startTime
    const progress = Math.min(elapsed / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    animatedScore.value = Math.round(targetScore * eased)
    
    if (progress < 1) {
      requestAnimationFrame(animate)
    } else {
      animatedScore.value = targetScore
    }
  }
  
  requestAnimationFrame(animate)
}

const submitTest = async () => {
  const answersMap: Record<string, any> = {}
  questions.value.forEach((q, index) => {
    const answer = answers.value.get(index)
    if (answer !== undefined && q.id) {
      answersMap[q.id] = answer
    }
  })
  
  submitting.value = true
  try {
    const data: { 
      score: number
      total_questions: number
      correct: number
      opened_sections: string[]
    } = await apiClient.request('/api/learning/grammar/placement-test/submit', {
      method: 'POST',
      body: answersMap
    })
    
    console.log('Placement test result:', data)
    
    result.value = data
    testSubmitted.value = true
    expandedCorrectResults.value = {}
    
    nextTick(() => {
      if (data.score !== undefined) {
        animateScore(data.score)
      } else {
        console.error('Score is undefined in result:', data)
        animatedScore.value = 0
      }
    })
  } catch (err: any) {
    error.value = err.message || 'Failed to submit test'
    console.error('Failed to submit placement test:', err)
  } finally {
    submitting.value = false
  }
}

const goToGrammar = () => {
  router.push('/learning/grammar')
}

const retryTest = () => {
  testSubmitted.value = false
  answers.value.clear()
  currentQuestionIndex.value = 0
  animatedScore.value = 0
  expandedCorrectResults.value = {}
  loadTest()
}

const getResultKey = (item: any, index: number): string => {
  return String(item?.question_id ?? index)
}

const isResultExpanded = (item: any, index: number): boolean => {
  if (!item?.correct) return true
  const key = getResultKey(item, index)
  return !!expandedCorrectResults.value[key]
}

const toggleResult = (item: any, index: number) => {
  if (!item?.correct) return
  const key = getResultKey(item, index)
  expandedCorrectResults.value = {
    ...expandedCorrectResults.value,
    [key]: !expandedCorrectResults.value[key]
  }
}

onMounted(() => {
  loadTest()
})
</script>

<style scoped>
.grammar-placement-test {
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
  margin-bottom: 32px;
  padding-bottom: 16px;
  border-bottom: 2px solid var(--border-primary);
}

.test-header h1 {
  margin: 0 0 16px 0;
}

.test-description {
  color: var(--text-secondary);
  margin-bottom: 16px;
  line-height: 1.6;
}

.test-progress {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 16px;
}

.btn-exit {
  margin-top: 16px;
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
  border: 3px solid var(--color-primary);
}

.score-value {
  font-size: 48px;
  font-weight: bold;
  margin-bottom: 8px;
  color: var(--color-primary);
}

.score-label {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-secondary);
}

.level-value {
  font-size: 42px;
  letter-spacing: 0.02em;
}

.test-score-secondary {
  margin-top: 8px;
  font-size: 14px;
  color: var(--text-secondary);
}

.results-summary {
  text-align: center;
  margin-bottom: 32px;
  padding: 20px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.opened-sections {
  margin-top: 20px;
  text-align: left;
}

.opened-sections h3 {
  margin-bottom: 12px;
  color: var(--text-primary);
}

.opened-sections ul {
  list-style: none;
  padding: 0;
  margin-bottom: 16px;
}

.opened-sections li {
  padding: 8px 12px;
  margin-bottom: 8px;
  background: var(--card-bg);
  border-left: 3px solid var(--color-success);
  border-radius: 4px;
}

.info-text {
  margin-top: 16px;
  color: var(--text-secondary);
  font-size: 14px;
}

.no-sections-opened {
  margin-top: 20px;
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  color: var(--text-secondary);
}

.results-details {
  margin-bottom: 32px;
}

.results-scale {
  width: 100%;
  margin-bottom: 28px;
}

.scale-level-labels {
  position: relative;
  height: 20px;
  margin-bottom: 6px;
  width: 100%;
}

.scale-level-label {
  position: absolute;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  text-align: center;
  pointer-events: none;
}

.scale-line-wrap {
  position: relative;
  width: 100%;
  height: 20px;
}

.scale-line {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  margin-top: -1px;
  height: 2px;
  background: var(--border-primary);
}

.scale-sep {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--text-secondary);
  transform: translateX(-50%);
  pointer-events: none;
}

.scale-dot {
  position: absolute;
  top: 50%;
  width: 16px;
  height: 16px;
  min-width: 16px;
  min-height: 16px;
  border-radius: 50%;
  border: 2px solid transparent;
  padding: 0;
  cursor: pointer;
  transform: translate(-50%, -50%);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.scale-dot:hover {
  transform: translate(-50%, -50%) scale(1.25);
  box-shadow: 0 2px 8px rgba(0,0,0,0.25);
}

.scale-dot.correct {
  background: var(--color-success);
  border-color: var(--color-success);
}

.scale-dot.incorrect {
  background: var(--color-danger);
  border-color: var(--color-danger);
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

.result-item.clickable {
  cursor: pointer;
  user-select: none;
}

.result-item.clickable:hover {
  border-color: var(--color-success);
}

.result-item.clickable:focus-visible {
  outline: 3px solid rgba(59, 130, 246, 0.45);
  outline-offset: 2px;
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
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.result-toggle {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.04);
}

.result-chevron {
  width: 9px;
  height: 9px;
  border-right: 2px solid var(--text-secondary);
  border-bottom: 2px solid var(--text-secondary);
  transform: rotate(-45deg);
  transition: transform 0.18s ease;
}

.result-chevron.expanded {
  transform: rotate(45deg);
}

.result-number {
  font-weight: 600;
  color: var(--text-primary);
}

.result-chapter {
  font-size: 13px;
  color: var(--text-secondary);
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

.result-user-answer {
  margin-bottom: 12px;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border-left: 3px solid var(--color-primary);
}

.result-user-answer.result-user-answer--correct {
  background: rgba(40, 167, 69, 0.1);
  border-left-color: var(--color-success);
}

.result-user-answer strong {
  display: block;
  margin-bottom: 8px;
}

.result-user-answer.result-user-answer--correct strong {
  color: var(--color-success);
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

.result-hint .choice-feedback {
  padding: 12px;
  background: rgba(255, 193, 7, 0.1);
  border-radius: 6px;
  border-left: 3px solid rgba(255, 193, 7, 0.5);
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

.answer-display {
  color: var(--text-primary);
  font-weight: 500;
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
