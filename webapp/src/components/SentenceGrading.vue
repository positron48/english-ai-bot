<template>
  <div class="correction-line">
    <span
      v-for="(tok, i) in rendered"
      :key="i"
      class="sg-tok"
      :class="`sg-tok--${tok.kind}`"
      :style="{ '--o': tok.order }"
    >
      <span class="sg-top">{{ tok.top }}</span>
      <span class="sg-bottom">
        <template v-if="tok.kind === 'insert'">
          <span class="sg-caret">⌃</span>
        </template>
        <template v-else>
          <span
            v-for="(p, j) in tok.parts"
            :key="j"
            :class="{ 'sg-del': p.struck }"
          >{{ p.t }}</span>
        </template>
      </span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { diffChars } from 'diff'
import type { SentenceGrade } from '../api/sentenceClient'

const props = defineProps<{ grade: SentenceGrade }>()

interface RenderTok {
  kind: 'ok' | 'wrong' | 'insert'
  top: string
  parts: { t: string; struck: boolean }[]
  order: number // sequential slot among marked tokens, drives the pen-stroke timing
}

// Build a teacher-markup model: for "wrong" tokens we diff at the character level so only
// the differing letters are struck (red) and only the corrected letters appear above (green).
// "insert" tokens render a caret with the missing word written above.
const rendered = computed<RenderTok[]>(() => {
  const out: RenderTok[] = []
  let mark = 0
  for (const tok of props.grade.tokens || []) {
    if (tok.status === 'ok') {
      out.push({ kind: 'ok', top: '', parts: [{ t: tok.text, struck: false }], order: -1 })
      continue
    }
    if (tok.status === 'insert') {
      out.push({ kind: 'insert', top: (tok.correction || '').trim(), parts: [], order: mark++ })
      continue
    }
    // wrong: char-level diff when it's a near-miss (misspelling/inflection) so only the
    // differing letters are struck; whole-word strike when the words are too different
    // (a real word swap), to avoid noisy per-letter coincidences.
    const original = tok.text || ''
    const correction = (tok.correction || '').trim()
    const parts: { t: string; struck: boolean }[] = []
    let top = ''
    if (!correction) {
      parts.push({ t: original, struck: true })
    } else {
      const diff = diffChars(original, correction)
      let common = 0
      for (const d of diff) if (!d.added && !d.removed) common += d.value.length
      const similar = common / Math.max(original.length, correction.length, 1) >= 0.4
      if (similar) {
        for (const d of diff) {
          if (d.added) top += d.value
          else parts.push({ t: d.value, struck: !!d.removed })
        }
      } else {
        parts.push({ t: original, struck: true })
        top = correction
      }
    }
    out.push({ kind: 'wrong', top, parts, order: mark++ })
  }
  return out
})
</script>

<style scoped>
.correction-line {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 4px 10px;
  font-family: 'Caveat', 'Marck Script', 'Comic Sans MS', cursive;
  font-size: 30px;
  line-height: 1.2;
  color: var(--text);
  /* pen-stroke timing (teacher marks words left-to-right, one after another) */
  --base: 240ms;   /* wait for the card to settle */
  --step: 420ms;   /* gap between consecutive marked words */
  --strike: 260ms; /* time to draw one strikethrough */
  --write: 340ms;  /* time to write one correction */
}

.sg-tok {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  /* reserve room for the line written above so rows never collide */
  padding-top: 1.1em;
  position: relative;
}

.sg-top {
  position: absolute;
  top: -0.15em;
  left: 50%;
  transform: translateX(-50%) rotate(-3deg);
  font-size: 0.72em;
  line-height: 1;
  white-space: nowrap;
  color: var(--correct-ink, #2f9d57);
}

.sg-bottom { white-space: nowrap; }

/* struck letters: keep the glyph, draw the red line as a pen stroke on top of it */
.sg-del {
  position: relative;
  color: var(--wrong-ink, #c4443c);
}
.sg-del::after {
  content: '';
  position: absolute;
  left: -1px;
  right: -1px;
  top: 56%;
  height: 2px;
  background: var(--wrong-ink, #c4443c);
  transform-origin: left center;
  /* both = hold the hidden start-state during the delay, keep the final state after */
  animation: sg-strike var(--strike) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step));
}

.sg-tok--insert .sg-caret { color: var(--correct-ink, #2f9d57); font-weight: 700; }

/* write each correction (green, above) just after striking the same word */
.sg-tok--wrong .sg-top,
.sg-tok--insert .sg-top {
  animation: sg-write var(--write) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step) + var(--strike) * 0.6);
}
.sg-tok--insert .sg-caret {
  animation: sg-pop 200ms ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step));
}

@keyframes sg-strike { from { transform: scaleX(0); } to { transform: scaleX(1); } }
@keyframes sg-write {
  from { clip-path: inset(0 100% 0 0); }
  to { clip-path: inset(0 -4px 0 0); }
}
@keyframes sg-pop {
  from { opacity: 0; transform: translateY(3px); }
  to { opacity: 1; transform: translateY(0); }
}

/* accessibility: no pen animation, show the final marked-up state immediately */
@media (prefers-reduced-motion: reduce) {
  .sg-tok--wrong .sg-top,
  .sg-tok--insert .sg-top,
  .sg-tok--insert .sg-caret {
    animation: none;
    clip-path: none;
    opacity: 1;
  }
  .sg-del::after {
    animation: none;
    transform: scaleX(1);
  }
}
</style>
