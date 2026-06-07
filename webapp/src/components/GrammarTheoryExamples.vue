<template>
  <div v-if="examples.length > 0" class="grammar-theory-examples" :class="variantClass">
    <h3 v-if="showHeading" class="grammar-theory-examples-title">{{ t('grammar.examples') }}</h3>
    <h4 v-else-if="variant === 'modal'" class="grammar-theory-examples-subtitle">{{ t('grammar.examples') }}</h4>
    <div
      v-for="example in examples"
      :key="exampleKey(example)"
      class="grammar-theory-example-item"
    >
      <div class="grammar-theory-example-text">{{ example.text }}</div>
      <button
        v-if="example.translation && !isRevealed(example)"
        type="button"
        class="grammar-theory-show-translation-btn"
        @click="reveal(example)"
      >
        {{ t('grammar.showTranslation') }}
      </button>
      <div v-if="example.translation && isRevealed(example)" class="grammar-theory-example-translation">
        {{ example.translation }}
      </div>
      <div v-if="example.notes" class="grammar-theory-example-notes">{{ example.notes }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

export interface GrammarTheoryExample {
  id?: string
  text: string
  translation?: string
  notes?: string
}

const props = withDefaults(defineProps<{
  examples: GrammarTheoryExample[]
  variant?: 'chapter' | 'modal'
  showHeading?: boolean
}>(), {
  variant: 'chapter',
  showHeading: true,
})

const { t } = useI18n()
const revealedKeys = ref<Record<string, boolean>>({})

const variantClass = `grammar-theory-examples--${props.variant}`

function exampleKey(example: GrammarTheoryExample): string {
  return example.id || example.text
}

function isRevealed(example: GrammarTheoryExample): boolean {
  return !!revealedKeys.value[exampleKey(example)]
}

function reveal(example: GrammarTheoryExample) {
  revealedKeys.value = { ...revealedKeys.value, [exampleKey(example)]: true }
}
</script>

<style scoped>
.grammar-theory-examples-title {
  margin: 0 0 12px;
  font-size: 1.1rem;
  color: var(--text-primary);
}

.grammar-theory-examples-subtitle {
  margin: 0 0 10px;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.grammar-theory-example-item {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--example-bg, var(--bg-tertiary));
  border-radius: 6px;
  border-left: 4px solid var(--color-primary);
}

.grammar-theory-examples--modal .grammar-theory-example-item {
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 8px;
}

.grammar-theory-example-item:last-child {
  margin-bottom: 0;
}

.grammar-theory-example-text {
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text-primary);
  line-height: 1.45;
}

.grammar-theory-examples--modal .grammar-theory-example-text {
  font-size: 14px;
}

.grammar-theory-show-translation-btn {
  margin: 4px 0 0;
  padding: 4px 10px;
  border: 1px solid var(--border-primary);
  border-radius: 999px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
}

.grammar-theory-show-translation-btn:hover {
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.grammar-theory-example-translation {
  color: var(--text-secondary);
  font-size: 14px;
  margin-top: 6px;
  line-height: 1.45;
}

.grammar-theory-examples--modal .grammar-theory-example-translation {
  font-size: 13px;
}

.grammar-theory-example-notes {
  color: var(--text-tertiary);
  font-size: 12px;
  font-style: italic;
  margin-top: 6px;
  line-height: 1.4;
}
</style>
