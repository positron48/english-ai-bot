<template>
  <button
    v-if="supported"
    type="button"
    class="voice-mic"
    :class="{ 'voice-mic--active': listening }"
    :disabled="disabled"
    :aria-label="label"
    :title="label"
    @click="start"
  >
    <LgIcon name="mic" :s="20" />
  </button>
</template>

<script setup lang="ts">
import LgIcon from './linglow/LgIcon.vue'
import { useSpeechRecognition } from '../composables/useSpeechRecognition'

const props = withDefaults(defineProps<{
  // Target recognition language code (e.g. "es", "en"); mapped to a locale in the composable.
  lang?: string
  disabled?: boolean
  label?: string
}>(), { lang: 'en', disabled: false, label: '' })

const emit = defineEmits<{ (e: 'transcript', text: string): void }>()

const { supported, listening, start } = useSpeechRecognition({
  lang: () => props.lang,
  onFinalTranscript: (text) => emit('transcript', text),
})
</script>

<style scoped>
.voice-mic {
  flex: 0 0 auto;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 12px;
  background: var(--surface-2, rgba(0, 0, 0, 0.05));
  color: var(--text);
  cursor: pointer;
  transition: background 0.15s, transform 0.15s;
}
.voice-mic:disabled { opacity: 0.4; cursor: default; }
.voice-mic--active {
  background: var(--salvia, #4caf50);
  color: #fff;
  animation: voiceMicPulse 1.2s ease-in-out infinite;
}
@keyframes voiceMicPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.12); }
}
</style>
