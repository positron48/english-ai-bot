<template>
  <div class="cl">
    <span
      v-for="(tk, i) in tokens"
      :key="i"
      class="tk"
      :class="`tk--${tk.kind}`"
      :style="{ '--o': tk.order }"
    >
      <!-- layer written ABOVE the line (corrections / inserted words) -->
      <span v-if="tk.kind === 'replace' || tk.kind === 'insert'" class="above">
        <span class="above-word">{{ tk.kind === 'replace' ? tk.to : tk.text }}</span>
        <svg v-if="tk.kind === 'insert'" class="arrow" viewBox="0 0 14 16" aria-hidden="true">
          <path d="M7 1 V11" />
          <path d="M3 8 L7 12 L11 8" fill="none" />
        </svg>
      </span>

      <!-- the learner's text on the baseline -->
      <span class="base">
        <template v-if="tk.kind === 'equal'">{{ tk.text }}</template>
        <span v-else-if="tk.kind === 'del'" class="del">{{ tk.text }}</span>
        <span v-else-if="tk.kind === 'replace'" class="del">{{ tk.from }}</span>
        <span v-else-if="tk.kind === 'insert'" class="ins-gap">·</span>
        <template v-else>
          <span
            v-for="(s, j) in tk.segs"
            :key="j"
            :class="{ del: s.kind === 'del', ins: s.kind === 'ins' }"
          >{{ s.t }}</span>
        </template>
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { renderCorrection, type RenderToken } from '../lib/correctionDiff'

const props = defineProps<{ userInput: string; corrected: string }>()

type Marked = RenderToken & { order: number }

// Assign a sequential slot to every marked token (not the untouched words) so the pen-stroke
// animation plays left-to-right, one correction after another.
const tokens = computed<Marked[]>(() => {
  let mark = 0
  return renderCorrection(props.userInput || '', props.corrected || '').map((tk) => ({
    ...tk,
    order: tk.kind === 'equal' ? -1 : mark++,
  }))
})
</script>

<style scoped>
.cl {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 8px 7px;
  font-family: 'Lora', Georgia, serif;
  font-size: 19px;
  line-height: 1.3;
  color: var(--text);
  /* pen-stroke timing (teacher marks corrections left-to-right) */
  --base: 200ms;
  --step: 360ms;
  --strike: 240ms;
  --write: 300ms;
}

/* each token is an in-flow column: anything written above sits in its own width, so marks
   never overlap their neighbours horizontally or vertically */
.tk {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
}
.base { white-space: nowrap; }

/* corrections / inserted words written above the line */
.above {
  display: flex;
  flex-direction: column;
  align-items: center;
  line-height: 1.05;
  margin-bottom: 1px;
}
.above-word {
  font-size: 0.78em;
  color: var(--correct-ink, #2f9d57);
  white-space: nowrap;
  animation: write var(--write) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step) + var(--strike) * 0.5);
}
.arrow {
  width: 13px;
  height: 15px;
  stroke: var(--correct-ink, #2f9d57);
  stroke-width: 1.6;
  stroke-linecap: round;
  stroke-linejoin: round;
  animation: fade 200ms ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step) + var(--strike) * 0.5);
}
.ins-gap { color: transparent; }

/* struck letters/words: keep the glyph, draw a red pen line across it */
.del {
  position: relative;
  color: var(--wrong-ink, #c4443c);
}
.del::after {
  content: '';
  position: absolute;
  left: -1px;
  right: -1px;
  top: 56%;
  height: 2px;
  background: var(--wrong-ink, #c4443c);
  transform-origin: left center;
  animation: strike var(--strike) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step));
}

/* letters inserted inside a word: green, written in at their exact position */
.ins {
  color: var(--correct-ink, #2f9d57);
  animation: write var(--write) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step) + var(--strike) * 0.5);
}

@keyframes strike { from { transform: scaleX(0); } to { transform: scaleX(1); } }
@keyframes write { from { clip-path: inset(0 100% 0 0); } to { clip-path: inset(0 -4px 0 0); } }
@keyframes fade { from { opacity: 0; } to { opacity: 1; } }

@media (prefers-reduced-motion: reduce) {
  .above-word, .arrow, .ins { animation: none; clip-path: none; opacity: 1; }
  .del::after { animation: none; transform: scaleX(1); }
}
</style>
