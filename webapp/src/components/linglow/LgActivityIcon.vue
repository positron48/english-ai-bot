<template>
  <img
    :src="src"
    :width="size"
    :height="size"
    class="lg-act-icon"
    :class="{ 'lg-act-icon--inactive': status === 'gray' }"
    alt=""
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import grammarRaw from '../../assets/linglow/icon-grammar.svg?raw'
import wordsRaw from '../../assets/linglow/icon-words.svg?raw'
import readRaw from '../../assets/linglow/icon-read.svg?raw'
import convRaw from '../../assets/linglow/icon-conversation.svg?raw'

const ICON_RAW: Record<string, string> = {
  grammar: grammarRaw,
  words: wordsRaw,
  reading: readRaw,
  conversation: convRaw,
}

const STATUS_COLORS: Record<string, string> = {
  gray:   '#888888',
  orange: '#d97706',
  yellow: '#ca8a04',
  green:  '#2d6b3a',
}

const props = withDefaults(defineProps<{
  type: 'grammar' | 'words' | 'reading' | 'conversation'
  status?: 'gray' | 'orange' | 'yellow' | 'green'
  size?: number
}>(), {
  status: 'gray',
  size: 22,
})

const src = computed(() => {
  const raw = ICON_RAW[props.type] ?? ICON_RAW.grammar
  const color = STATUS_COLORS[props.status] ?? STATUS_COLORS.gray
  const colored = raw
    .replace(/<!DOCTYPE[^>]*>/g, '')
    .replace(/fill="#000000"/g, `fill="${color}"`)
  return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(colored)}`
})
</script>

<style scoped>
.lg-act-icon {
  display: inline-block;
  flex-shrink: 0;
  object-fit: contain;
}
.lg-act-icon--inactive {
  opacity: 0.4;
}
</style>
