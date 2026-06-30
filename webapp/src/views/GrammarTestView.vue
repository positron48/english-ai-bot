<template>
  <div class="grammar-test">
    <div v-if="loading" class="loading">
      <p>{{ t('grammar.loadingTest') }}</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadTest" class="btn btn-primary">{{ t('common.retry') }}</button>
    </div>
    
    <div v-else-if="testSubmitted" class="test-results">
      <div class="results-header">
        <h1>{{ t('grammar.testResults') }}</h1>
        <div class="score-display-wrapper">
          <div class="score-display" :class="{ 'passed': result.passed, 'failed': !result.passed }">
            <div class="score-value">{{ animatedScore }}%</div>
            <div class="score-label">{{ result.passed ? t('grammar.passed') : t('grammar.failed') }}</div>
            
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
        <p>{{ t('grammar.youAnswered', { correct: result.correct, total: result.total }) }}</p>
        <p v-if="!result.passed" class="retry-hint">
          {{ t('grammar.needAtLeast50') }}
        </p>
      </div>
      
      <div class="results-details">
        <h2>{{ t('grammar.detailedResults') }}</h2>
        <div
          v-for="(item, index) in result.results"
          :key="index"
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
            <span class="result-number">{{ t('grammar.question') }} {{ index + 1 }}</span>
            <span class="result-status" :class="{ 'correct': item.correct, 'incorrect': !item.correct }">
              {{ item.correct ? '✓ ' + t('grammar.correct') : '✗ ' + t('grammar.failed') }}
            </span>
            <span v-if="item.correct" class="result-toggle" aria-hidden="true">
              <span class="result-chevron" :class="{ 'expanded': isResultExpanded(item, index) }"></span>
            </span>
          </div>
          
          <div v-if="isResultExpanded(item, index)" class="result-body">
            <!-- Question -->
            <div class="result-question-prompt">
              <strong>{{ t('grammar.question') }}:</strong>
              <div v-html="renderMarkdown(item.prompt || getQuestionPrompt(item.question_id, item.chapter_id))"></div>
            </div>
            
            <!-- User Answer -->
            <div class="result-user-answer">
              <strong>{{ t('grammar.yourAnswer') }}:</strong>
              <div class="answer-display">{{ formatAnswer(item.question_id, item.user_answer, item.chapter_id) || t('grammar.notAnswered') }}</div>
            </div>
            
            <!-- Correct Answer (only if incorrect) -->
            <div v-if="!item.correct" class="result-correct-answer">
              <strong>{{ t('grammar.correctAnswer') }}:</strong>
              <div class="answer-display">{{ formatAnswer(item.question_id, item.correct_answer, item.chapter_id) }}</div>
            </div>
            
            <!-- Hint/Feedback for incorrect answers -->
            <div v-if="!item.correct" class="result-hint">
              <div v-if="getChoiceFeedback(item.question_id, item.user_answer, item.chapter_id)" class="choice-feedback">
                <strong>{{ t('grammar.hint') }}:</strong>
                <div v-html="renderMarkdown(getChoiceFeedback(item.question_id, item.user_answer, item.chapter_id))"></div>
              </div>
            </div>
            
            <!-- Explanation -->
            <div v-if="item.explanation" class="result-explanation">
              <strong>{{ t('grammar.explanation') }}:</strong>
              <div v-html="renderMarkdown(item.explanation)"></div>
            </div>

            <div class="result-report-row">
              <button
                type="button"
                class="report-text-link"
                :disabled="reportSubmitting || isTestReportSent(item, index)"
                @click.stop="openTestReportDialog(item, index)"
              >
                {{ isTestReportSent(item, index) ? t('training.reportSent') : t('training.reportIssue') }}
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <div class="results-actions">
        <button @click="goBack" class="btn btn-secondary">{{ t('grammar.backToChapter') }}</button>
        <button
          v-if="showNextActionButton"
          @click.stop.prevent="handleNextActionClick"
          @mousedown.stop
          @mouseup.stop
          class="btn btn-primary"
          :disabled="nextActionLoading || !nextActionKind"
          type="button"
          ref="nextActionButtonRef"
          style="z-index: 9999; position: relative;"
        >
          {{ nextActionLoading ? t('common.loading') : nextActionLabel }}
        </button>
        <button v-if="!result.passed" @click="retryTest" class="btn btn-primary">{{ t('grammar.retryTest') }}</button>
      </div>
    </div>
    
    <div v-else class="test-content">
      <div class="test-header">
        <h1>{{ testTitle }}</h1>
        <div class="test-progress">
          {{ t('grammar.questionOf', { current: currentQuestionIndex + 1, total: questions.length }) }}
        </div>
        <button @click="exitTest" class="btn btn-secondary btn-exit">{{ t('grammar.exitTest') }}</button>
      </div>
      
      <div class="test-questions">
        <GrammarQuestion
          v-if="currentQuestion"
          :key="currentQuestion.id || currentQuestionIndex"
          :ref="el => setQuestionRef(currentQuestionIndex, el)"
          :question="currentQuestion"
          :show-answers="false"
          :show-theory-help-button="false"
          :initial-answer="currentQuestion ? answers.get(getQuestionKey(currentQuestion)) : undefined"
          @answer="handleAnswerWithAutoNext(currentQuestionIndex, $event)"
        />
      </div>
      
      <div class="test-navigation">
        <button 
          v-if="currentQuestionIndex > 0"
          @click="previousQuestion"
          class="btn btn-secondary"
        >
          {{ t('grammar.previous') }}
        </button>
        <button 
          v-if="currentQuestionIndex < questions.length - 1"
          @click="nextQuestion"
          :disabled="!hasAnswer(currentQuestionIndex)"
          class="btn btn-primary"
        >
          {{ t('grammar.next') }}
        </button>
        <button 
          v-else
          @click="submitTest"
          :disabled="submitting || !hasAnswer(currentQuestionIndex)"
          class="btn btn-primary"
        >
          {{ submitting ? t('grammar.submitting') : t('grammar.submitTest') }}
        </button>
      </div>
    </div>
    
    <!-- Exit confirmation modal -->
    <ConfirmModal
      :visible="showExitConfirm"
      :message="t('grammar.exitTestConfirm')"
      @confirm="handleExitConfirm"
      @cancel="showExitConfirm = false"
    />

    <ContentReportDialog
      :open="reportDialogOpen"
      :submitting="reportSubmitting"
      :categories="testReportCategories"
      :category="reportCategory"
      :details="reportDetails"
      @update:category="reportCategory = $event"
      @update:details="reportDetails = $event"
      @close="closeTestReportDialog"
      @submit="submitTestReport"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter, onBeforeRouteUpdate } from 'vue-router'
import { marked } from 'marked'
import { grammarClient } from '../api/grammarClient'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import ContentReportDialog from '../components/ContentReportDialog.vue'
import { contentReportClient } from '../api/contentReportClient'
import {
  GRAMMAR_TEST_REPORT_CATEGORIES,
  buildReportComment,
} from '../constants/contentReportCategories'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'

const { t } = useI18n()
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
    return t('grammar.chapterTest')
  }
  return t('grammar.categoryTest')
})

const questions = ref<any[]>([])
// Store answers by unique question identifier (question_id for chapter tests, chapter_id:question_id for category tests)
// This ensures answers are always correctly matched to questions, even if user navigates back and forth
const answers = ref<Map<string, any>>(new Map())
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
const expandedCorrectResults = ref<Record<string, boolean>>({})
const nextChapterId = ref<string | null>(null)
const nextSectionId = ref<string | null>(null)
const isLastChapterInCategory = ref(false)
const nextActionLoading = ref(false)
const nextActionButtonRef = ref<HTMLButtonElement | null>(null)

const testReportCategories = GRAMMAR_TEST_REPORT_CATEGORIES
const reportDialogOpen = ref(false)
const reportSubmitting = ref(false)
const reportCategory = ref('')
const reportDetails = ref('')
const reportTarget = ref<{ item: any; index: number } | null>(null)
const reportSentKeys = ref<Set<string>>(new Set())

const testReportKey = (item: any, index: number) => `${item.question_id || index}:${item.chapter_id || scopeId.value}:${index}`
const isTestReportSent = (item: any, index: number) => reportSentKeys.value.has(testReportKey(item, index))

const openTestReportDialog = (item: any, index: number) => {
  if (reportSubmitting.value || isTestReportSent(item, index)) return
  reportTarget.value = { item, index }
  reportCategory.value = ''
  reportDetails.value = ''
  reportDialogOpen.value = true
}

const closeTestReportDialog = () => {
  if (reportSubmitting.value) return
  reportDialogOpen.value = false
}

const submitTestReport = async () => {
  if (!reportTarget.value || reportSubmitting.value) return
  const { item, index } = reportTarget.value
  const comment = buildReportComment(
    reportCategory.value,
    reportDetails.value,
    t(`training.reportCategories.${reportCategory.value}`)
  )
  if (!comment) return
  reportSubmitting.value = true
  try {
    const question = getQuestionById(item.question_id, item.chapter_id)
    await contentReportClient.submit({
      sourceType: 'grammar_test',
      reportCategory: reportCategory.value,
      comment,
      grammarQuestionID: item.question_id,
      grammarChapterID: item.chapter_id || scopeId.value,
      payload: {
        scope: scope.value,
        scope_id: scopeId.value,
        question_snapshot: question || item,
        user_answer: item.user_answer,
        correct_answer: item.correct_answer,
      },
    })
    reportSentKeys.value.add(testReportKey(item, index))
    reportDialogOpen.value = false
  } catch (err) {
    console.error('Failed to submit grammar test report:', err)
  } finally {
    reportSubmitting.value = false
  }
}

const nextActionKind = computed<null | 'nextChapter' | 'categoryTest'>(() => {
  if (scope.value !== 'chapter' || !result.value?.passed) return null
  if (nextActionLoading.value) return 'nextChapter' // placeholder while loading to keep button stable
  if (nextChapterId.value) return 'nextChapter'
  if (isLastChapterInCategory.value && nextSectionId.value) return 'categoryTest'
  return null
})

const nextActionLabel = computed(() => {
  return nextActionKind.value === 'categoryTest' ? t('grammar.categoryTest') : t('grammar.nextChapter')
})

const showNextActionButton = computed(() => {
  return scope.value === 'chapter' && !!result.value?.passed && (nextActionLoading.value || !!nextActionKind.value)
})

const currentQuestion = computed(() => {
  return questions.value[currentQuestionIndex.value] || null
})

// Get unique key for a question (for storing/retrieving answers)
// For category tests: chapter_id:question_id (since question IDs are only unique within a chapter)
// For chapter tests: question_id (question IDs are unique within a chapter)
const getQuestionKey = (question: any): string => {
  if (!question || !question.id) {
    return ''
  }
  if (scope.value === 'category' && (question as any)._category_test_chapter_id) {
    return `${(question as any)._category_test_chapter_id}:${question.id}`
  }
  return question.id
}

const resetTestStateForNewRoute = () => {
  // Reset everything that is route-specific so the same component can be reused between:
  // - chapter test -> category test
  // - category test -> chapter test
  // CRITICAL: Set loading FIRST and testSubmitted to false to force Vue to switch to loading block
  testSubmitted.value = false
  result.value = null
  loading.value = true
  error.value = null
  questions.value = []
  answers.value = new Map<string, any>()
  currentQuestionIndex.value = 0
  submitting.value = false
  questionRefs.value = []
  animatedScore.value = 0
  percentageAnimationComplete.value = false
  expandedCorrectResults.value = {}
  nextChapterId.value = null
  nextSectionId.value = null
  isLastChapterInCategory.value = false
  nextActionLoading.value = false
}

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
  const question = questions.value[index]
  if (!question) {
    return false
  }
  const questionKey = getQuestionKey(question)
  const answer = answers.value.get(questionKey)
  if (answer === undefined || answer === null) {
    return false
  }
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
  // CRITICAL: Reset testSubmitted FIRST to ensure Vue switches to loading block
  testSubmitted.value = false
  result.value = null
  loading.value = true
  error.value = null
  try {
    if (scope.value === 'chapter') {
      const data: { questions: any[]; total: number } = await grammarClient.getChapterTest(scopeId.value)
      questions.value = data.questions || []
    } else {
      // Category test
      const data: { questions: any[]; total: number } = await grammarClient.getCategoryTest(scopeId.value)
      questions.value = data.questions || []
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load test'
  } finally {
    loading.value = false
  }
}

const handleAnswer = (index: number, answer: any) => {
  const question = questions.value[index]
  if (!question) {
    return
  }
  const questionKey = getQuestionKey(question)
  answers.value.set(questionKey, answer)
}

const handleAnswerWithAutoNext = (index: number, answer: any) => {
  const question = questions.value[index]
  if (!question) {
    return
  }
  const questionKey = getQuestionKey(question)
  answers.value.set(questionKey, answer)
  
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
  // Build array of answer objects with explicit question identification
  // Each answer is explicitly linked to its question via question_id and chapter_id (for category tests)
  // This ensures correct matching even if questions are reordered or some are skipped
  const answerItems: Array<{
    question_id: string
    chapter_id?: string
    answer: any
  }> = []
  
  // Process questions in the exact order they appear in the test
  for (let index = 0; index < questions.value.length; index++) {
    const q = questions.value[index]
    if (!q || !q.id) {
      continue // Skip invalid questions
    }
    
    // Get answer for this specific question by unique key
    const questionKey = getQuestionKey(q)
    const answer = answers.value.get(questionKey)
    
    // Create answer item with explicit question identification
    const answerItem: {
      question_id: string
      chapter_id?: string
      answer: any
    } = {
      question_id: q.id,
      answer: answer !== undefined && answer !== null ? answer : null
    }
    
    // For category tests, include chapter_id to uniquely identify questions
    // (since question IDs are only unique within a chapter)
    if (scope.value === 'category' && (q as any)._category_test_chapter_id) {
      answerItem.chapter_id = (q as any)._category_test_chapter_id
    }
    
    answerItems.push(answerItem)
  }
  
  submitting.value = true
  try {
    const data: { score: number; passed: boolean; correct: number; total: number; results: any[] } = 
      await grammarClient.submitTest(scope.value as 'chapter' | 'category', scopeId.value, answerItems)
    
    result.value = data
    testSubmitted.value = true
    expandedCorrectResults.value = {}
    nextChapterId.value = null
    nextSectionId.value = null
    isLastChapterInCategory.value = false
    if (scope.value === 'chapter' && data.passed) {
      loadNextChapterId().catch(() => {
        // Failed to load next chapter id
      })
    }
    
    // Trigger haptic feedback
    triggerHapticFeedback(data.passed)
    
    // Start score animation after DOM update
    // Sounds will be played by watch when animation completes
    nextTick(() => {
      animateScore(data.score)
    })
  } catch (err: any) {
    error.value = err.message || 'Failed to submit test'
    // Failed to submit grammar test
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
  expandedCorrectResults.value = {}
  nextChapterId.value = null
  nextSectionId.value = null
  isLastChapterInCategory.value = false
  nextActionLoading.value = false
  loadTest()
}

const loadNextChapterId = async () => {
  if (scope.value !== 'chapter') return
  const chapterId = scopeId.value
  nextActionLoading.value = true
  try {
    const data: { section_id: string; is_last: boolean; next_chapter_id: string } =
      await grammarClient.getNextChapter(chapterId)

    nextSectionId.value = data.section_id || null
    isLastChapterInCategory.value = !!data.is_last
    nextChapterId.value = data.next_chapter_id || null
  } finally {
    nextActionLoading.value = false
  }
}

const handleNextActionClick = (event?: Event) => {
  if (event) {
    event.preventDefault()
    event.stopPropagation()
  }
  goToNextAction()
}

const goToNextAction = async () => {
  if (nextActionKind.value === 'nextChapter') {
    if (!nextChapterId.value) return
    await router.push({ name: 'GrammarChapter', params: { chapterId: nextChapterId.value } })
    return
  }
  if (nextActionKind.value === 'categoryTest') {
    if (!nextSectionId.value) return
    const targetSectionId = nextSectionId.value
    
    // Use router.push with replace to navigate
    // The key on router-view in App.vue should force component recreation
    await router.push({ 
      name: 'GrammarCategoryTest', 
      params: { sectionId: targetSectionId },
      // Add query param to force route change detection
      query: { _t: Date.now().toString() }
    })
    
    // Wait for navigation to complete
    await nextTick()
    await new Promise(resolve => setTimeout(resolve, 50))
    
    // Force state reset
    resetTestStateForNewRoute()
    await nextTick()
    
    // Load test
    loadTest()
  }
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

// Helper function to get question by ID and optionally chapter_id
// For category tests, chapter_id is required to uniquely identify questions
const getQuestionById = (questionId: string, chapterId?: string): any => {
  if (scope.value === 'category' && chapterId) {
    // For category tests, find question by both id and chapter_id
    return questions.value.find(q => 
      q.id === questionId && 
      (q as any)._category_test_chapter_id === chapterId
    ) || null
  }
  // For chapter tests or if chapter_id not provided, find by id only
  return questions.value.find(q => q.id === questionId) || null
}

// Helper function to get question prompt
const getQuestionPrompt = (questionId: string, chapterId?: string): string => {
  const question = getQuestionById(questionId, chapterId)
  return question?.prompt || ''
}

// Helper function to format answer for display
const formatAnswer = (questionId: string, answer: any, chapterId?: string): string => {
  if (answer === undefined || answer === null) {
    return ''
  }
  
  const question = getQuestionById(questionId, chapterId)
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
    
    case 'fill_blank':
    case 'reorder':
      return String(answer)
    
    default:
      return String(answer)
  }
}

// Helper function to get choice feedback/hint
const getChoiceFeedback = (questionId: string, userAnswer: any, chapterId?: string): string => {
  const question = getQuestionById(questionId, chapterId)
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
      // Telegram haptic feedback failed
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
      // Native vibration failed
    }
  }
}

onMounted(() => {
  // Always reset state on mount to ensure clean state
  resetTestStateForNewRoute()
  loadTest()
})

// IMPORTANT: Vue Router may reuse the same component instance when only params change,
// OR create a new instance when switching between different routes (chapter test -> category test).
// We need to handle both cases.

// Primary: Watch route.path directly - most reliable way to detect route changes
watch(
  () => route.path,
  async (newPath, oldPath) => {
    // Skip on initial mount
    if (!oldPath) return
    // Only reload if path actually changed
    if (newPath === oldPath) return
    // Check if this is a test route (chapter test or category test)
    const isTestRoute = newPath.includes('/test')
    const wasTestRoute = oldPath.includes('/test')
    // Only reload if we're switching between test routes
    if (!isTestRoute || !wasTestRoute) return
    resetTestStateForNewRoute()
    await nextTick()
    loadTest()
  },
  { immediate: false }
)

// Secondary: Watch scope and scopeId - these change when switching between chapter and category tests
watch(
  [() => scope.value, () => scopeId.value],
  async ([newScope, newScopeId], [oldScope, oldScopeId]) => {
    // Skip on initial mount
    if (oldScope === undefined) return
    // Only reload if scope or scopeId actually changed
    if (newScope === oldScope && newScopeId === oldScopeId) return
    resetTestStateForNewRoute()
    await nextTick()
    loadTest()
  },
  { immediate: false }
)

// Tertiary: Use onBeforeRouteUpdate for cases where component is reused
onBeforeRouteUpdate(async (to, from) => {
  // Check if this is a test route (chapter test or category test)
  const isTestRoute = to.path.includes('/test')
  const wasTestRoute = from.path.includes('/test')
  // Only reload if we're switching between test routes
  if (!isTestRoute || !wasTestRoute) return
  // Only reload if path actually changed
  if (to.path === from.path) return
  resetTestStateForNewRoute()
  await nextTick()
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

.result-report-row {
  margin-top: 12px;
  display: flex;
  justify-content: center;
}

.report-text-link {
  border: 0;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.2;
  padding: 0;
  margin: 0;
  cursor: pointer;
  text-decoration: none;
}

.report-text-link:hover:not(:disabled) {
  color: var(--text-primary);
}

.report-text-link:disabled {
  cursor: default;
  opacity: 0.8;
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
  justify-content: space-between;
  align-items: center;
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
