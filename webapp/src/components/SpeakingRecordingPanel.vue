<template>
  <div class="record-controls">
    <template v-if="recorder.state.value === 'recording'">
      <div class="status-block">
        <div class="progress-track">
          <div class="progress-fill recording" :style="{ width: `${recordingProgress}%` }" />
        </div>
        <p class="status-text">
          {{ t('speaking.recordingCountdown', { seconds: remainingSeconds }) }}
        </p>
        <button type="button" class="btn danger" @click="$emit('stop')">
          {{ t('speaking.stopRecording') }}
        </button>
      </div>
    </template>

    <template v-else-if="recorder.state.value === 'processing'">
      <div class="status-block">
        <div class="progress-track">
          <div class="progress-fill indeterminate" />
        </div>
        <p class="status-text">{{ t('speaking.processing') }}</p>
      </div>
    </template>

    <template v-else-if="submitting">
      <div class="status-block">
        <div class="progress-track">
          <div class="progress-fill indeterminate" />
        </div>
        <p class="status-text">{{ t('speaking.evaluating') }}</p>
      </div>
    </template>

    <template v-else-if="recorder.state.value === 'no_speech'">
      <div class="status-block no-speech">
        <p class="status-text">{{ t('speaking.noSpeechDetected') }}</p>
        <button type="button" class="btn primary" @click="$emit('retry')">
          {{ t('speaking.tryAgain') }}
        </button>
      </div>
    </template>

    <template v-else-if="showIdle">
      <button type="button" class="btn primary" @click="$emit('start')">
        {{ t('speaking.tapToRecord') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAudioRecorder } from '../composables/useAudioRecorder'

const props = defineProps<{
  recorder: ReturnType<typeof useAudioRecorder>
  submitting: boolean
}>()

defineEmits<{
  start: []
  stop: []
  retry: []
}>()

const { t } = useI18n()

const recordingProgress = computed(() => {
  const max = props.recorder.maxDurationMs.value
  const remaining = props.recorder.remainingMs.value
  if (max <= 0) return 0
  return Math.min(100, Math.max(0, ((max - remaining) / max) * 100))
})

const remainingSeconds = computed(() =>
  Math.max(1, Math.ceil(props.recorder.remainingMs.value / 1000))
)

const showIdle = computed(() =>
  ['idle', 'no_speech', 'error'].includes(props.recorder.state.value)
)
</script>

<style scoped>
.record-controls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.status-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.status-text {
  margin: 0;
  color: var(--text-secondary);
}
.progress-track {
  width: 100%;
  height: 8px;
  border-radius: 999px;
  background: var(--bg-secondary);
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  border-radius: 999px;
  background: var(--color-primary);
}
.progress-fill.recording {
  transition: width 0.1s linear;
}
.progress-fill.indeterminate {
  width: 35%;
  animation: speaking-indeterminate 1.2s ease-in-out infinite;
}
@keyframes speaking-indeterminate {
  0% { transform: translateX(-120%); }
  100% { transform: translateX(320%); }
}
.btn {
  padding: 10px 16px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--card-bg);
  cursor: pointer;
  align-self: flex-start;
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
.no-speech .status-text {
  color: #c0392b;
}
</style>
