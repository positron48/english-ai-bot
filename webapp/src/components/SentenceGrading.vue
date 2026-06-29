<template>
  <div class="sg">
    <div class="sg-marked">
      <template v-for="(tok, i) in grade.tokens" :key="i">
        <span v-if="tok.status === 'ok'" class="sg-tok sg-ok">{{ tok.text }}</span>
        <span v-else-if="tok.status === 'wrong'" class="sg-tok sg-wrong">
          <span class="sg-correction">{{ tok.correction }}</span>
          <span class="sg-struck">{{ tok.text }}</span>
        </span>
        <span v-else class="sg-tok sg-insert">
          <span class="sg-correction">{{ tok.correction }}</span>
          <span class="sg-caret">^</span>
        </span>
        {{ ' ' }}
      </template>
    </div>

    <div class="sg-verdict" :class="`sg-verdict--${grade.outcome}`">
      <span v-if="grade.outcome === 'star'" class="sg-badge">★ {{ t('sentence.outcomeStar') }}</span>
      <span v-else-if="grade.outcome === 'passed'" class="sg-badge">✓ {{ t('sentence.outcomePassed') }}</span>
      <span v-else class="sg-badge">✗ {{ t('sentence.outcomeFailed') }}</span>
    </div>

    <div class="sg-correct">
      <span class="sg-correct-label">{{ t('sentence.correctAnswer') }}:</span>
      <span class="sg-correct-text">{{ grade.corrected_es }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SentenceGrade } from '../api/sentenceClient'

defineProps<{ grade: SentenceGrade }>()
const { t } = useI18n()
</script>

<style scoped>
.sg {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.sg-marked {
  font-size: 1.25rem;
  line-height: 2.2;
  font-family: 'Caveat', 'Comic Sans MS', cursive, sans-serif;
}
.sg-tok {
  position: relative;
  display: inline-block;
  white-space: nowrap;
}
.sg-ok {
  color: var(--lg-text, #1c1c1e);
}
.sg-wrong .sg-struck {
  color: #d23b3b;
  text-decoration: line-through;
  text-decoration-thickness: 2px;
}
.sg-correction {
  position: absolute;
  top: -1.15em;
  left: 0;
  font-size: 0.8em;
  color: #1f9d57;
  white-space: nowrap;
}
.sg-insert .sg-caret {
  color: #1f9d57;
  font-weight: 700;
}
.sg-verdict {
  font-weight: 600;
}
.sg-verdict--star { color: #d9a400; }
.sg-verdict--passed { color: #1f9d57; }
.sg-verdict--failed { color: #d23b3b; }
.sg-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 1.05rem;
}
.sg-correct {
  font-size: 0.95rem;
  color: var(--lg-text-secondary, #6b6b70);
}
.sg-correct-label {
  font-weight: 600;
  margin-right: 6px;
}
.sg-correct-text {
  color: #1f9d57;
}
</style>
