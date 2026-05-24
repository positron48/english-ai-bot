<template>
  <div class="speaking-session">
    <div v-if="loading" class="state">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="state error">{{ error }}</div>
    <div v-else-if="session?.status === 'completed'" class="summary">
      <h1>{{ t('speaking.sessionComplete') }}</h1>
      <router-link to="/learning/speaking" class="btn">{{ t('speaking.backToLevels') }}</router-link>
    </div>
    <template v-else-if="session && currentTask">
      <header class="session-header">
        <span>{{ t('speaking.progress', { current: session.current_task_index + 1, total: session.total_tasks }) }}</span>
      </header>

      <section v-if="phase === 'task'" class="task-card">
        <h2>{{ currentTask.title }}</h2>
        <p class="prompt">{{ currentTask.prompt_ru }}</p>
        <p class="display-text">{{ currentTask.display_text }}</p>
        <div class="record-controls">
          <button v-if="recorder.state.value !== 'recording'" type="button" class="btn primary" @click="recorder.startRecording()">
            {{ t('speaking.tapToRecord') }}
          </button>
          <button v-else type="button" class="btn danger" @click="recorder.stopRecording()">
            {{ t('speaking.stopRecording') }}
          </button>
          <button
            v-if="recorder.state.value === 'recorded'"
            type="button"
            class="btn primary"
            :disabled="submitting"
            @click="submit('initial')"
          >
            {{ submitting ? t('speaking.evaluating') : t('speaking.submit') }}
          </button>
          <button v-if="recorder.state.value === 'recorded'" type="button" class="btn ghost" @click="recorder.resetRecording()">
            {{ t('speaking.recordAgain') }}
          </button>
        </div>
        <p v-if="recorder.errorMessage.value" class="error">{{ recorder.errorMessage.value }}</p>
      </section>

      <section v-else-if="phase === 'feedback' && evaluation" class="feedback-card">
        <h2>{{ t('speaking.feedbackTitle') }}</h2>
        <p class="understood"><strong>{{ t('speaking.understood') }}</strong> {{ evaluation.understood_answer }}</p>
        <div class="scores">
          <span>{{ t('speaking.scoreMeaning') }}: {{ evaluation.meaning_score }}/5</span>
          <span>{{ t('speaking.scoreGrammar') }}: {{ evaluation.grammar_score }}/5</span>
          <span>{{ t('speaking.scorePronunciation') }}: {{ evaluation.pronunciation_score }}/5</span>
          <span>{{ t('speaking.scoreFluency') }}: {{ evaluation.fluency_score }}/5</span>
        </div>
        <p class="feedback-text">{{ evaluation.short_feedback_ru }}</p>
        <p v-if="evaluation.better_version" class="better">
          <strong>{{ t('speaking.betterVersion') }}</strong> {{ evaluation.better_version }}
        </p>
        <p class="attempts">{{ t('speaking.attempts', { n: evaluation.attempt_no, max: evaluation.max_attempts }) }}</p>
        <div class="actions">
          <button
            v-if="!evaluation.can_advance"
            type="button"
            class="btn primary"
            @click="retry"
          >
            {{ t('speaking.tryAgain') }}
          </button>
          <button
            v-if="evaluation.better_version && !evaluation.is_acceptable"
            type="button"
            class="btn"
            @click="startRepair"
          >
            {{ t('speaking.repeatImproved') }}
          </button>
          <button
            v-if="evaluation.can_advance"
            type="button"
            class="btn primary"
            @click="advance"
          >
            {{ t('speaking.nextTask') }}
          </button>
        </div>
      </section>

      <section v-else-if="phase === 'repair' && evaluation" class="task-card">
        <h2>{{ t('speaking.repeatImproved') }}</h2>
        <p class="display-text">{{ evaluation.repeat_task || evaluation.better_version }}</p>
        <div class="record-controls">
          <button v-if="recorder.state.value !== 'recording'" type="button" class="btn primary" @click="recorder.startRecording()">
            {{ t('speaking.tapToRecord') }}
          </button>
          <button v-else type="button" class="btn danger" @click="recorder.stopRecording()">
            {{ t('speaking.stopRecording') }}
          </button>
          <button
            v-if="recorder.state.value === 'recorded'"
            type="button"
            class="btn primary"
            :disabled="submitting"
            @click="submit('repair')"
          >
            {{ submitting ? t('speaking.evaluating') : t('speaking.submit') }}
          </button>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAudioRecorder } from '../composables/useAudioRecorder'
import { useSpeaking, type SpeakingEvaluation, type SpeakingSession, type SpeakingTask } from '../composables/useSpeaking'

const { t } = useI18n()
const route = useRoute()
const { getSession, submitAudio, nextTask } = useSpeaking()
const recorder = useAudioRecorder()

const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const session = ref<SpeakingSession | null>(null)
const evaluation = ref<SpeakingEvaluation | null>(null)
const phase = ref<'task' | 'feedback' | 'repair'>('task')

const sessionId = computed(() => Number(route.params.sessionId))
const currentTask = computed<SpeakingTask | null>(() => session.value?.current_task ?? null)

onMounted(async () => {
  try {
    session.value = await getSession(sessionId.value)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('speaking.loadFailed')
  } finally {
    loading.value = false
  }
})

async function submit(mode: 'initial' | 'repair') {
  if (!session.value || !currentTask.value || !recorder.blob.value) return
  submitting.value = true
  error.value = ''
  try {
    evaluation.value = await submitAudio(sessionId.value, currentTask.value.id, recorder.blob.value, mode)
    phase.value = 'feedback'
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('speaking.submitFailed')
  } finally {
    submitting.value = false
  }
}

function retry() {
  recorder.resetRecording()
  evaluation.value = null
  phase.value = 'task'
}

function startRepair() {
  recorder.resetRecording()
  phase.value = 'repair'
}

async function advance() {
  if (!session.value) return
  try {
    session.value = await nextTask(sessionId.value)
    evaluation.value = null
    recorder.resetRecording()
    if (session.value.status === 'completed') return
    phase.value = 'task'
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('speaking.advanceFailed')
  }
}
</script>

<style scoped>
.speaking-session {
  max-width: 720px;
  margin: 0 auto;
  padding: 20px;
}
.session-header {
  margin-bottom: 16px;
  color: var(--text-secondary);
}
.task-card, .feedback-card, .summary {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  padding: 24px;
}
.prompt {
  color: var(--text-secondary);
}
.display-text {
  font-size: 22px;
  font-weight: 600;
  margin: 16px 0 24px;
}
.record-controls {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.btn {
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--card-bg);
  cursor: pointer;
}
.btn.primary {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.btn.danger {
  background: #c0392b;
  color: #fff;
  border-color: #c0392b;
}
.btn.ghost {
  background: transparent;
}
.scores {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin: 12px 0;
  font-size: 14px;
}
.feedback-text {
  margin: 12px 0;
  line-height: 1.5;
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
}
.error, .state.error {
  color: #c0392b;
}
</style>
