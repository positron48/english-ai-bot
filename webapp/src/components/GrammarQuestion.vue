<template>
  <div class="grammar-question" :class="{ 'answered': answered, 'correct': isCorrect, 'incorrect': !isCorrect && answered }">
    <button
      v-if="showTheoryHelpButton && hasTheoryBlock"
      type="button"
      class="question-block-indicator"
      :title="theoryHelpText"
      :aria-label="theoryHelpText"
      @click="toggleTheoryHelp"
    >
      <span class="question-block-icon">i</span>
    </button>
    <teleport to="body">
      <div
        v-if="showTheoryHelp && hasTheoryBlock"
        class="theory-modal-overlay"
        @click.self="closeTheoryHelp"
      >
        <div class="theory-modal">
          <button
            type="button"
            class="theory-modal-close"
            :aria-label="t('common.close')"
            :title="t('common.close')"
            @click="closeTheoryHelp"
          >
            ×
          </button>
          <template v-if="theoryBlock">
            <div
              v-if="theoryBlock.theory?.content_md"
              class="theory-tooltip-content markdown-content"
              v-html="renderMarkdown(theoryBlock.theory.content_md)"
            ></div>
            <ul v-if="Array.isArray(theoryBlock.theory?.key_points) && theoryBlock.theory.key_points.length" class="theory-tooltip-points">
              <li v-for="(point, idx) in theoryBlock.theory.key_points" :key="idx">{{ point }}</li>
            </ul>
          </template>
          <template v-else>
            {{ t('common.loading') }}
          </template>
          <div
            v-if="theoryMetaLines.length"
            class="theory-modal-source"
          >
            <div
              v-for="(line, idx) in theoryMetaLines"
              :key="idx"
              class="theory-modal-source-line"
            >
              <span class="theory-modal-source-label">{{ line.label }}</span>
              <span class="theory-modal-source-value">{{ line.value }}</span>
            </div>
          </div>
        </div>
      </div>
    </teleport>
    <div v-if="chapterTitle" class="question-chapter">{{ chapterTitle }}</div>
    <div class="question-prompt" v-html="renderMarkdown(question.prompt)"></div>
    
    <!-- MCQ Single -->
    <div v-if="question.type === 'mcq_single'" class="question-choices">
      <button
        v-for="choice in shuffledChoices"
        :key="choice.id"
        @click="selectAnswer(choice.id)"
        :disabled="answered"
        :class="['choice-btn', { 
          'selected': userAnswer === choice.id,
          'correct': answered && choice.id === correctAnswer,
          'incorrect': answered && userAnswer === choice.id && choice.id !== correctAnswer
        }]"
      >
        <div class="choice-content">
          <span v-html="renderMarkdown(choice.text)"></span>
          <span v-if="answered && choice.id === correctAnswer" class="check-icon">✓</span>
        </div>
        <!-- Show feedback inside incorrect selected choice -->
        <div 
          v-if="answered && userAnswer === choice.id && choice.id !== correctAnswer && choice.feedback"
          class="choice-feedback-inline"
        >
          <div v-html="renderMarkdown(choice.feedback)"></div>
        </div>
      </button>
    </div>
    
    <!-- MCQ Multi -->
    <div v-if="question.type === 'mcq_multi'" class="question-choices">
      <button
        v-for="choice in shuffledChoices"
        :key="choice.id"
        @click="toggleAnswer(choice.id)"
        :disabled="answered"
        :class="['choice-btn', 'multi', { 
          'selected': Array.isArray(userAnswer) && userAnswer.includes(choice.id),
          'correct': answered && Array.isArray(correctAnswer) && correctAnswer.includes(choice.id),
          'incorrect': answered && Array.isArray(userAnswer) && userAnswer.includes(choice.id) && Array.isArray(correctAnswer) && !correctAnswer.includes(choice.id)
        }]"
      >
        <div class="choice-content">
          <span class="checkbox" :class="{ 'checked': Array.isArray(userAnswer) && userAnswer.includes(choice.id) }"></span>
          <span v-html="renderMarkdown(choice.text)"></span>
        </div>
        <!-- Show feedback inside incorrect selected choice -->
        <div 
          v-if="answered && Array.isArray(userAnswer) && userAnswer.includes(choice.id) && Array.isArray(correctAnswer) && !correctAnswer.includes(choice.id) && choice.feedback"
          class="choice-feedback-inline"
        >
          <div v-html="renderMarkdown(choice.feedback)"></div>
        </div>
      </button>
    </div>
    
    <!-- Fill Blank -->
    <div v-if="question.type === 'fill_blank'" class="question-input">
      <div class="fill-input-wrapper">
        <input
          v-model="userAnswer"
          @input="onAnswerChange"
          @keydown.enter.prevent="handleFillBlankEnter"
          :disabled="answered"
          :class="['fill-input', { 'correct': answered && isCorrect, 'incorrect': answered && !isCorrect }]"
          type="text"
          :placeholder="question.prompt.includes('___') ? 'Fill in the blank' : 'Your answer'"
          ref="fillInputRef"
        />
        <button
          v-if="!answered && userAnswer && userAnswer.trim()"
          @click="handleFillBlankCheck"
          class="check-btn"
          :disabled="answered"
        >
          {{ showAnswers ? 'Check' : 'Confirm' }}
        </button>
      </div>
      <!-- Show correct/incorrect indicator for fill_blank after check -->
      <div v-if="answered && showAnswers" class="fill-blank-feedback" :class="{ 'correct': isCorrect, 'incorrect': !isCorrect }">
        <span v-if="isCorrect" class="feedback-icon">✓</span>
        <span v-else class="feedback-icon">✗</span>
        <span v-if="isCorrect" class="feedback-text">{{ t('grammar.correct') }}</span>
        <span v-else class="feedback-text">{{ t('grammar.wrong') }}</span>
      </div>
      <!-- Show correct answer if incorrect -->
      <div v-if="answered && !isCorrect && showAnswers && correctAnswer" class="correct-answer-display">
        <strong>{{ t('grammar.correctAnswer') }}:</strong> {{ correctAnswer }}
      </div>
    </div>
    
    <!-- True/False -->
    <div v-if="question.type === 'true_false'" class="question-choices">
      <button
        @click="selectAnswer('true')"
        :disabled="answered"
        :class="['choice-btn', { 
          'selected': userAnswer === 'true',
          'correct': answered && correctAnswer === 'true',
          'incorrect': answered && userAnswer === 'true' && correctAnswer !== 'true'
        }]"
      >
        Да
      </button>
      <button
        @click="selectAnswer('false')"
        :disabled="answered"
        :class="['choice-btn', { 
          'selected': userAnswer === 'false',
          'correct': answered && correctAnswer === 'false',
          'incorrect': answered && userAnswer === 'false' && correctAnswer !== 'false'
        }]"
      >
        Нет
      </button>
    </div>
    
    <!-- Error Spotting -->
    <div v-if="question.type === 'error_spotting'" class="question-choices">
      <button
        v-for="choice in shuffledChoices"
        :key="choice.id"
        @click="selectAnswer(choice.id)"
        :disabled="answered"
        :class="['choice-btn', { 
          'selected': userAnswer === choice.id,
          'correct': answered && choice.id === correctAnswer,
          'incorrect': answered && userAnswer === choice.id && choice.id !== correctAnswer
        }]"
      >
        <div class="choice-content">
          <span v-html="renderMarkdown(choice.text)"></span>
          <span v-if="answered && choice.id === correctAnswer" class="check-icon">✓</span>
        </div>
        <!-- Show feedback inside incorrect selected choice -->
        <div 
          v-if="answered && userAnswer === choice.id && choice.id !== correctAnswer && choice.feedback"
          class="choice-feedback-inline"
        >
          <div v-html="renderMarkdown(choice.feedback)"></div>
        </div>
      </button>
    </div>
    
    <!-- Reorder -->
    <div v-if="question.type === 'reorder'" class="reorder-container">
      <!-- Sentence container (where selected words go) -->
      <div class="reorder-sentence">
        <button
          v-for="(word, index) in selectedWords"
          :key="`selected-${index}`"
          @click="moveWordToAvailable(index)"
          :disabled="answered"
          :class="['reorder-word-btn', 'sentence-word', {
            'correct': answered && isCorrect,
            'incorrect': answered && !isCorrect
          }]"
        >
          {{ word }}
        </button>
        <span v-if="lastPunctuation" class="reorder-punctuation">{{ lastPunctuation }}</span>
      </div>
      
      <!-- Available words container -->
      <div class="reorder-words">
        <button
          v-for="(word, index) in availableWords"
          :key="`available-${index}`"
          @click="moveWordToSentence(index)"
          :disabled="answered"
          class="reorder-word-btn available-word"
        >
          {{ word }}
        </button>
      </div>
      
      <!-- Show correct answer if incorrect -->
      <div v-if="answered && !isCorrect" class="correct-answer-display">
        <strong>{{ t('grammar.correctAnswer') }}:</strong> {{ correctAnswer }}
      </div>
    </div>
    
    <div v-if="showExplanation && answered && question.explanation" class="question-explanation">
      <strong>{{ t('grammar.explanation') }}:</strong>
      <div v-html="renderMarkdown(question.explanation)"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { marked } from 'marked'
import { useI18n } from 'vue-i18n'

export interface TheoryChapterContext {
  categoryTitle: string
  categoryTitleTranslations?: Record<string, string>
  chapterTitle: string
  chapterTitleTranslations?: Record<string, string>
  level: string
}

interface Question {
  id: string
  type: 'mcq_single' | 'mcq_multi' | 'fill_blank' | 'reorder' | 'error_spotting' | 'true_false'
  prompt: string
  choices?: Array<{ id: string; text: string; feedback?: string }>
  correct_answer?: any
  explanation?: string
  theory_block_id?: string
  difficulty?: number
}

const props = withDefaults(defineProps<{
  question: Question
  theoryBlock?: any
  showAnswers?: boolean
  initialAnswer?: any
  chapterTitle?: string
  showExplanation?: boolean
  /** When false, parent renders the help control (e.g. next to Continue) and calls toggleTheoryHelp via ref */
  showTheoryHelpButton?: boolean
  theoryChapterContext?: TheoryChapterContext | null
}>(), {
  showExplanation: true,
  showTheoryHelpButton: true
})

const { t, locale } = useI18n()

const emit = defineEmits<{
  answer: [answer: any]
}>()

const userAnswer = ref<any>(null)
const answered = ref(false)
const correctAnswer = ref<any>(props.question.correct_answer)
const fillInputRef = ref<HTMLInputElement | null>(null)

// Reorder-specific state
const selectedWords = ref<string[]>([])
const availableWords = ref<string[]>([])
const lastPunctuation = ref<string>('')

// Shuffled choices for MCQ questions
const shuffledChoices = ref<Array<{ id: string; text: string; feedback?: string }>>([])

const isCorrect = computed(() => {
  if (!answered.value || userAnswer.value === null) return false
  return compareAnswers(userAnswer.value, correctAnswer.value)
})

const hasTheoryBlock = computed(() => typeof props.question?.theory_block_id === 'string' && props.question.theory_block_id.length > 0)
const showTheoryHelp = ref(false)
const theoryHelpText = t('grammar.theoryBlock')

const localizedTitle = (title: string, titleTranslations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && titleTranslations?.[currentLocale]) {
    return titleTranslations[currentLocale]
  }
  return title
}

const theoryMetaLines = computed(() => {
  const ctx = props.theoryChapterContext
  if (!ctx) return [] as { label: string; value: string }[]
  const out: { label: string; value: string }[] = []
  const cat = localizedTitle(ctx.categoryTitle, ctx.categoryTitleTranslations)
  if (cat) {
    out.push({ label: t('grammar.theoryFooterCategory'), value: cat })
  }
  const ch = localizedTitle(ctx.chapterTitle, ctx.chapterTitleTranslations)
  if (ch) {
    out.push({ label: t('grammar.theoryFooterChapter'), value: ch })
  }
  if (ctx.level) {
    out.push({ label: t('grammar.theoryFooterLevel'), value: ctx.level })
  }
  return out
})

const toggleTheoryHelp = () => {
  showTheoryHelp.value = !showTheoryHelp.value
}

const closeTheoryHelp = () => {
  showTheoryHelp.value = false
}

// Shuffle function (Fisher-Yates algorithm)
const shuffleArray = <T>(array: T[]): T[] => {
  const shuffled = [...array]
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
  }
  return shuffled
}

// Initialize shuffled choices for MCQ and error_spotting questions
const initializeChoices = () => {
  if (props.question.type === 'mcq_single' || props.question.type === 'mcq_multi' || props.question.type === 'error_spotting') {
    if (props.question.choices && props.question.choices.length > 0) {
      shuffledChoices.value = shuffleArray(props.question.choices)
    } else {
      shuffledChoices.value = []
    }
  }
}

// Initialize on mount
initializeChoices()

// Initialize reorder question
const initializeReorder = () => {
  if (!props.question || props.question.type !== 'reorder') {
    availableWords.value = []
    selectedWords.value = []
    lastPunctuation.value = ''
    return
  }
  
  if (typeof props.question.correct_answer === 'string') {
    const correctAnswerStr = props.question.correct_answer.trim()
    
    // Extract punctuation from the end (.!?)
    const punctuationMatch = correctAnswerStr.match(/[.!?]$/)
    lastPunctuation.value = punctuationMatch ? punctuationMatch[0] : ''
    
    // Remove punctuation for word parsing
    let textForWords = lastPunctuation.value 
      ? correctAnswerStr.slice(0, -1).trim() 
      : correctAnswerStr
    
    // Normalize spaces around commas
    textForWords = textForWords.replace(/\s*,\s*/g, ' , ')
    
    // Split into words, keeping commas as separate elements
    const words = textForWords
      .split(/\s+/)
      .filter(w => w.trim().length > 0)
    
    // Shuffle words using Fisher-Yates algorithm
    const shuffled = [...words]
    for (let i = shuffled.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]]
    }
    
    availableWords.value = shuffled
    selectedWords.value = []
    userAnswer.value = null
    answered.value = false
    correctAnswer.value = props.question.correct_answer
  } else {
    // Reset if correct_answer is not a string
    availableWords.value = []
    selectedWords.value = []
    lastPunctuation.value = ''
  }
}

// Initialize on mount
initializeReorder()
// Restore answer if provided
if (props.initialAnswer !== undefined && props.initialAnswer !== null) {
  setTimeout(() => restoreAnswer(), 0)
}

const selectAnswer = (answer: any) => {
  if (answered.value) return
  userAnswer.value = answer
  answered.value = props.showAnswers || false
  emit('answer', answer)
  // If not showing answers (test mode), emit immediately for auto-advance
  if (!props.showAnswers) {
    // Answer is already emitted above
  }
}

const toggleAnswer = (choiceId: string) => {
  if (answered.value) return
  if (!Array.isArray(userAnswer.value)) {
    userAnswer.value = []
  }
  const index = userAnswer.value.indexOf(choiceId)
  if (index > -1) {
    userAnswer.value.splice(index, 1)
  } else {
    userAnswer.value.push(choiceId)
  }
  emit('answer', [...userAnswer.value])
  // For multi-choice, auto-advance happens when at least one option is selected
  // The parent component will handle the auto-advance logic
}

const onAnswerChange = () => {
  // Don't emit answer on every input change for fill_blank type
  // Answer will be emitted only when user submits (Enter or Check button)
  // This prevents sounds from playing on every keystroke
  if (props.question.type !== 'fill_blank') {
    emit('answer', userAnswer.value)
  }
}

const handleFillBlankCheck = () => {
  if (answered.value || !userAnswer.value || !userAnswer.value.trim()) {
    return
  }
  // Emit answer when user submits (Check button or Enter)
  emit('answer', userAnswer.value)
  // If showing answers (quiz mode), mark as answered to show feedback
  if (props.showAnswers) {
    answered.value = true
  }
  // In test mode (showAnswers=false), this will trigger auto-advance in parent
}

const handleFillBlankEnter = () => {
  if (answered.value || !userAnswer.value || !userAnswer.value.trim()) {
    return
  }
  // Emit answer when user submits (Enter key)
  emit('answer', userAnswer.value)
  // In test mode (showAnswers=false), parent will handle auto-advance
  // In quiz mode (showAnswers=true), check the answer
  if (props.showAnswers) {
    handleFillBlankCheck()
  }
}

// Reorder functions
const moveWordToSentence = (index: number) => {
  if (answered.value) return
  
  const word = availableWords.value[index]
  availableWords.value.splice(index, 1)
  selectedWords.value.push(word)
  
  // Check answer only when all words are selected
  if (availableWords.value.length === 0) {
    checkReorderAnswer()
  }
}

const moveWordToAvailable = (index: number) => {
  if (answered.value) return
  
  const word = selectedWords.value[index]
  selectedWords.value.splice(index, 1)
  availableWords.value.push(word)
  
  // Don't check answer when words are moved back - answer is incomplete
  // Clear answer if it was set
  if (userAnswer.value !== null) {
    userAnswer.value = null
    emit('answer', null)
  }
}

const checkReorderAnswer = () => {
  // Only check if all words are selected
  if (availableWords.value.length > 0) {
    return
  }
  
  // Build sentence from selected words
  let userSentence = ''
  for (let i = 0; i < selectedWords.value.length; i++) {
    const word = selectedWords.value[i]
    if (word === ',') {
      // Comma without space before it
      userSentence += ','
      // Space after comma (if not last element)
      if (i < selectedWords.value.length - 1) {
        userSentence += ' '
      }
    } else {
      // Regular word: space before it (if not first element and previous is not comma)
      if (i > 0 && selectedWords.value[i - 1] !== ',') {
        userSentence += ' '
      }
      userSentence += word
    }
  }
  userSentence = userSentence.trim()
  
  // Add punctuation
  if (lastPunctuation.value) {
    userSentence += lastPunctuation.value
  }
  
  userAnswer.value = userSentence
  // Mark as answered when all words are selected (for quizzes)
  // This ensures the answer is considered given in quiz context
  answered.value = props.showAnswers || false
  emit('answer', userSentence)
}

const compareAnswers = (user: any, correct: any): boolean => {
  if (typeof correct === 'string') {
    // For reorder, compare normalized strings
    return user?.trim().toLowerCase() === correct.trim().toLowerCase()
  }
  if (Array.isArray(correct)) {
    if (!Array.isArray(user)) return false
    if (user.length !== correct.length) return false
    return user.every((val, idx) => val === correct[idx])
  }
  return user === correct
}

const renderMarkdown = (text: string): string => {
  if (!text) return ''
  try {
    return marked.parse(text) as string
  } catch (error) {
    return text
  }
}

// Restore answer from initialAnswer prop
const restoreAnswer = () => {
  if (props.initialAnswer === undefined || props.initialAnswer === null) {
    return
  }
  
  if (props.question.type === 'mcq_single' || props.question.type === 'true_false' || props.question.type === 'error_spotting') {
    userAnswer.value = props.initialAnswer
  } else if (props.question.type === 'mcq_multi') {
    userAnswer.value = Array.isArray(props.initialAnswer) ? [...props.initialAnswer] : [props.initialAnswer]
  } else if (props.question.type === 'fill_blank') {
    userAnswer.value = props.initialAnswer
  } else if (props.question.type === 'reorder') {
    // For reorder, we need to reconstruct the word selection from the answer string
    if (typeof props.initialAnswer === 'string') {
      restoreReorderAnswer(props.initialAnswer)
    }
  }
}

// Restore reorder answer by reconstructing word positions
const restoreReorderAnswer = (answerStr: string) => {
  if (!props.question.correct_answer || typeof props.question.correct_answer !== 'string') {
    return
  }
  
  // Parse the answer string back into words
  const answerTrimmed = answerStr.trim()
  const correctAnswerStr = props.question.correct_answer.trim()
  
  // Extract punctuation
  const punctuationMatch = answerTrimmed.match(/[.!?]$/)
  const answerPunctuation = punctuationMatch ? punctuationMatch[0] : ''
  
  // Remove punctuation for comparison
  let answerWithoutPunct = answerPunctuation ? answerTrimmed.slice(0, -1).trim() : answerTrimmed
  let correctWithoutPunct = correctAnswerStr.match(/[.!?]$/) 
    ? correctAnswerStr.slice(0, -1).trim() 
    : correctAnswerStr
  
  // Normalize spaces around commas (same as in initializeReorder)
  answerWithoutPunct = answerWithoutPunct.replace(/\s*,\s*/g, ' , ')
  correctWithoutPunct = correctWithoutPunct.replace(/\s*,\s*/g, ' , ')
  
  // Split into words
  const answerWords = answerWithoutPunct.split(/\s+/).filter(w => w.trim().length > 0)
  const allCorrectWords = correctWithoutPunct.split(/\s+/).filter(w => w.trim().length > 0)
  
  // Reconstruct selectedWords from answer
  selectedWords.value = [...answerWords]
  
  // Reconstruct availableWords - words that are in correct answer but not in selected
  // Count occurrences to handle duplicate words
  const answerWordCounts = new Map<string, number>()
  answerWords.forEach(word => {
    answerWordCounts.set(word, (answerWordCounts.get(word) || 0) + 1)
  })
  
  const correctWordCounts = new Map<string, number>()
  allCorrectWords.forEach(word => {
    correctWordCounts.set(word, (correctWordCounts.get(word) || 0) + 1)
  })
  
  // Build available words list - all words from correct answer minus those used in answer
  const available: string[] = []
  correctWordCounts.forEach((count, word) => {
    const usedCount = answerWordCounts.get(word) || 0
    const remaining = count - usedCount
    for (let i = 0; i < remaining; i++) {
      available.push(word)
    }
  })
  
  // Shuffle available words (they should be in random order)
  for (let i = available.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [available[i], available[j]] = [available[j], available[i]]
  }
  
  availableWords.value = available
  userAnswer.value = answerStr
}

// Watch for question changes to reinitialize
watch(() => props.question, (newQuestion) => {
  if (!newQuestion) return
  
  initializeReorder()
  initializeChoices()
  // Reset answer state
  userAnswer.value = null
  answered.value = false
  correctAnswer.value = newQuestion.correct_answer
  
  // Focus input when switching to a fill_blank question — only on test pages (chapter test, placement test), not on chapter quiz
  if (newQuestion.type === 'fill_blank' && props.showAnswers === false) {
    nextTick(() => {
      fillInputRef.value?.focus()
    })
  }
  
  // Restore answer after initialization (use nextTick to ensure initialization is complete)
  setTimeout(() => {
    restoreAnswer()
  }, 0)
}, { immediate: true, deep: true })

// Watch for initialAnswer changes
watch(() => props.initialAnswer, () => {
  // Only restore if question is already initialized
  if (props.question) {
    restoreAnswer()
  }
}, { immediate: true })

watch(() => props.showAnswers, (newVal) => {
  if (newVal && userAnswer.value !== null) {
    answered.value = true
  }
})

watch(() => props.question?.id, () => {
  showTheoryHelp.value = false
})

defineExpose({
  toggleTheoryHelp,
  closeTheoryHelp
})
</script>

<style scoped>
.grammar-question {
  position: relative;
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}

.question-block-indicator {
  position: absolute;
  top: 12px;
  right: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  width: 24px;
  height: 24px;
  border: 1px solid var(--border-primary);
  border-radius: 50%;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.question-block-indicator:hover {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.question-block-icon {
  width: 100%;
  height: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  color: var(--text-secondary);
}

.theory-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 16px;
}

.theory-modal {
  position: relative;
  width: min(900px, 100%);
  max-height: 80vh;
  overflow: auto;
  padding: 42px 44px 16px 18px;
  border: 1px solid var(--border-primary);
  border-radius: 10px;
  background: var(--card-bg);
  color: var(--text-secondary);
  font-size: 15px;
  line-height: 1.65;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.3);
}

.theory-modal-close {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 2;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  margin: 0;
  padding: 0;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 26px;
  font-weight: 300;
  line-height: 1;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}

.theory-modal-close:hover {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.theory-tooltip-content {
  color: var(--text-primary);
  margin-bottom: 16px;
}

.theory-tooltip-points {
  margin: 0 0 0;
  padding-left: 18px;
}

.theory-tooltip-points li {
  margin-bottom: 6px;
}

.theory-modal-source {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-secondary);
}

.theory-modal-source-line {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  margin-bottom: 6px;
}

.theory-modal-source-line:last-child {
  margin-bottom: 0;
}

.theory-modal-source-label {
  font-weight: 600;
  color: var(--text-primary);
  min-width: 5.5em;
}

.theory-modal-source-value {
  flex: 1;
  min-width: 0;
}

.question-chapter {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  font-weight: 500;
}

.question-prompt {
  margin-bottom: 16px;
  font-size: 16px;
  line-height: 1.6;
  color: var(--text-primary);
}

.question-choices {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.choice-btn {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  color: var(--text-primary);
}

.choice-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.choice-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.choice-btn.selected {
  border-color: var(--color-primary);
  background: var(--color-primary-light);
}

.choice-btn.correct {
  border-color: var(--color-success);
  background: rgba(40, 167, 69, 0.1);
}

.choice-btn.incorrect {
  border-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.1);
}

.choice-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.choice-btn.multi {
  gap: 12px;
}

.checkbox {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-primary);
  border-radius: 4px;
  flex-shrink: 0;
  position: relative;
}

.checkbox.checked {
  background: var(--color-primary);
  border-color: var(--color-primary);
}

.checkbox.checked::after {
  content: '✓';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: white;
  font-size: 14px;
  font-weight: bold;
}

.check-icon {
  color: var(--color-success);
  font-weight: bold;
}

.question-input {
  margin-bottom: 16px;
}

.fill-input-wrapper {
  display: flex;
  gap: 8px;
  align-items: center;
}

.fill-input, .error-input {
  flex: 1;
  padding: 12px;
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  background: var(--input-bg);
  color: var(--text-primary);
  font-size: 16px;
  font-family: inherit;
}

.fill-input:focus, .error-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.fill-input:disabled, .error-input:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.fill-input.correct {
  border-color: var(--color-success);
  background: rgba(40, 167, 69, 0.1);
}

.fill-input.incorrect {
  border-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.1);
}

.fill-blank-feedback {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 12px;
  border-radius: 6px;
  font-weight: 600;
}

.fill-blank-feedback.correct {
  background: rgba(40, 167, 69, 0.1);
  color: var(--color-success);
}

.fill-blank-feedback.incorrect {
  background: rgba(220, 53, 69, 0.1);
  color: var(--color-danger);
}

.feedback-icon {
  font-size: 18px;
  font-weight: bold;
}

.feedback-text {
  font-size: 14px;
}

.check-btn {
  padding: 12px 20px;
  height: calc(12px * 2 + 1.5em); /* Match input height: padding top + bottom + line height */
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  margin-bottom: 10px;
  box-sizing: border-box;
}

.check-btn:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.check-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.reorder-container {
  margin-bottom: 16px;
}

.reorder-sentence {
  min-height: 50px;
  padding: 15px;
  border: 2px dashed var(--border-primary);
  border-radius: 8px;
  margin-bottom: 15px;
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  align-items: center;
}

.reorder-words {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 15px;
}

.reorder-word-btn {
  padding: 8px 12px;
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  background: var(--bg-secondary);
  cursor: pointer;
  font-size: 1em;
  transition: all 0.2s ease;
  color: var(--text-primary);
}

.reorder-word-btn:hover:not(:disabled) {
  border-color: var(--color-primary);
  transform: translateY(-2px);
}

.reorder-word-btn:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

.reorder-word-btn.sentence-word {
  background: var(--card-bg);
}

.reorder-punctuation {
  font-size: 1.2em;
  font-weight: bold;
  color: var(--text-primary);
  margin-left: 2px;
  display: inline-block;
  vertical-align: middle;
}

.reorder-word-btn.correct {
  border-color: var(--color-success);
  background: rgba(40, 167, 69, 0.1);
  color: var(--color-success);
}

.reorder-word-btn.incorrect {
  border-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.1);
  color: var(--color-danger);
}

.correct-answer-display {
  margin-top: 12px;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  color: var(--text-secondary);
}

.correct-answer-display strong {
  color: var(--text-primary);
  margin-right: 8px;
}

.question-explanation {
  margin-top: 16px;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  border-left: 4px solid var(--color-primary);
}

.question-explanation strong {
  display: block;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.choice-feedback {
  margin-top: 12px;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  border-left: 4px solid var(--color-primary);
  color: var(--text-secondary);
  line-height: 1.6;
}

.choice-feedback.incorrect-feedback {
  border-left-color: var(--color-danger);
  background: rgba(220, 53, 69, 0.05);
}

.choice-feedback-inline {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid rgba(220, 53, 69, 0.3);
  width: 100%;
  color: var(--text-secondary);
  font-size: 1em;
  line-height: 1.5;
}
</style>
