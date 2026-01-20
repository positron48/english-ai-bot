<template>
  <div class="grammar-question" :class="{ 'answered': answered, 'correct': isCorrect, 'incorrect': !isCorrect && answered }">
    <div class="question-prompt" v-html="renderMarkdown(question.prompt)"></div>
    
    <!-- MCQ Single -->
    <div v-if="question.type === 'mcq_single'" class="question-choices">
      <button
        v-for="choice in question.choices"
        :key="choice.id"
        @click="selectAnswer(choice.id)"
        :disabled="answered"
        :class="['choice-btn', { 
          'selected': userAnswer === choice.id,
          'correct': answered && choice.id === correctAnswer,
          'incorrect': answered && userAnswer === choice.id && choice.id !== correctAnswer
        }]"
      >
        <span v-html="renderMarkdown(choice.text)"></span>
        <span v-if="answered && choice.id === correctAnswer" class="check-icon">✓</span>
      </button>
    </div>
    
    <!-- MCQ Multi -->
    <div v-if="question.type === 'mcq_multi'" class="question-choices">
      <button
        v-for="choice in question.choices"
        :key="choice.id"
        @click="toggleAnswer(choice.id)"
        :disabled="answered"
        :class="['choice-btn', 'multi', { 
          'selected': Array.isArray(userAnswer) && userAnswer.includes(choice.id),
          'correct': answered && Array.isArray(correctAnswer) && correctAnswer.includes(choice.id),
          'incorrect': answered && Array.isArray(userAnswer) && userAnswer.includes(choice.id) && Array.isArray(correctAnswer) && !correctAnswer.includes(choice.id)
        }]"
      >
        <span class="checkbox" :class="{ 'checked': Array.isArray(userAnswer) && userAnswer.includes(choice.id) }"></span>
        <span v-html="renderMarkdown(choice.text)"></span>
      </button>
    </div>
    
    <!-- Fill Blank -->
    <div v-if="question.type === 'fill_blank'" class="question-input">
      <input
        v-model="userAnswer"
        @input="onAnswerChange"
        :disabled="answered"
        type="text"
        :placeholder="question.prompt.includes('___') ? 'Fill in the blank' : 'Your answer'"
        class="fill-input"
      />
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
        True
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
        False
      </button>
    </div>
    
    <!-- Error Spotting -->
    <div v-if="question.type === 'error_spotting'" class="question-input">
      <textarea
        v-model="userAnswer"
        @input="onAnswerChange"
        :disabled="answered"
        :placeholder="'Correct the error'"
        class="error-input"
        rows="3"
      ></textarea>
    </div>
    
    <!-- Reorder -->
    <div v-if="question.type === 'reorder'" class="reorder-container">
      <div class="reorder-items">
        <div
          v-for="(item, index) in reorderItems"
          :key="index"
          class="reorder-item"
          :class="{ 'dragging': draggingIndex === index }"
          @mousedown="startDrag(index, $event)"
          @touchstart="startDrag(index, $event)"
        >
          {{ item }}
        </div>
      </div>
    </div>
    
    <div v-if="answered && question.explanation" class="question-explanation">
      <strong>Explanation:</strong>
      <div v-html="renderMarkdown(question.explanation)"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { marked } from 'marked'

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

const props = defineProps<{
  question: Question
  showAnswers?: boolean
}>()

const emit = defineEmits<{
  answer: [answer: any]
}>()

const userAnswer = ref<any>(null)
const answered = ref(false)
const correctAnswer = ref<any>(props.question.correct_answer)
const draggingIndex = ref<number | null>(null)
const reorderItems = ref<string[]>([])

const isCorrect = computed(() => {
  if (!answered.value || userAnswer.value === null) return false
  return compareAnswers(userAnswer.value, correctAnswer.value)
})

// Initialize reorder items if needed
if (props.question.type === 'reorder' && Array.isArray(props.question.correct_answer)) {
  reorderItems.value = [...props.question.correct_answer]
  // Shuffle for display
  for (let i = reorderItems.value.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [reorderItems.value[i], reorderItems.value[j]] = [reorderItems.value[j], reorderItems.value[i]]
  }
}

const selectAnswer = (answer: any) => {
  if (answered.value) return
  userAnswer.value = answer
  answered.value = props.showAnswers || false
  emit('answer', answer)
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
}

const onAnswerChange = () => {
  emit('answer', userAnswer.value)
}

const startDrag = (index: number, event: MouseEvent | TouchEvent) => {
  if (answered.value) return
  draggingIndex.value = index
  // Simple drag implementation - could be enhanced
}

const compareAnswers = (user: any, correct: any): boolean => {
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

watch(() => props.showAnswers, (newVal) => {
  if (newVal && userAnswer.value !== null) {
    answered.value = true
  }
})
</script>

<style scoped>
.grammar-question {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
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
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  color: var(--text-primary);
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

.fill-input, .error-input {
  width: 100%;
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

.reorder-container {
  margin-bottom: 16px;
}

.reorder-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.reorder-item {
  padding: 12px 16px;
  background: var(--bg-secondary);
  border: 2px solid var(--border-primary);
  border-radius: 6px;
  cursor: move;
  user-select: none;
}

.reorder-item.dragging {
  opacity: 0.5;
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
</style>
