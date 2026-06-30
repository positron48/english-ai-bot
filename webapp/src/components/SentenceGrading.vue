<template>
  <div class="correction-line">
    <span
      v-for="(tok, i) in rendered"
      :key="i"
      class="sg-tok"
      :class="`sg-tok--${tok.kind}`"
      :style="{ '--o': tok.order }"
    >
      <span v-if="tok.top" class="sg-top">{{ tok.top }}</span>
      <span class="sg-body">
        <template v-if="tok.kind === 'insert'">
          <span class="sg-caret">∧</span>
        </template>
        <template v-else>
          <span
            v-for="(s, j) in tok.segs"
            :key="j"
            :class="{ 'sg-del': s.kind === 'del', 'sg-ins': s.kind === 'ins' }"
          >{{ s.t }}</span>
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

type SegKind = 'same' | 'del' | 'ins'
interface Seg { t: string; kind: SegKind }
interface RenderTok {
  // ok: untouched word · wrong: in-place char edits · replace: whole word swapped
  // (struck, new word above) · insert: a missing word (caret + word above)
  kind: 'ok' | 'wrong' | 'replace' | 'insert'
  top: string
  segs: Seg[]
  order: number // sequential slot among marked tokens, drives the pen-stroke timing
}

// Only treat a wrong token as a near-miss (edit the letters in place) when the words still
// share most of their characters; otherwise strike the whole word and write the correction
// above it. 0.55 keeps real typos/inflections inline ("pero"→"perro") while short function-word
// swaps ("Este"→"Es", ratio 0.5) are struck whole instead of leaving a confusing fragment.
const SIMILAR_THRESHOLD = 0.55

const rendered = computed<RenderTok[]>(() => {
  const out: RenderTok[] = []
  let mark = 0
  for (const tok of props.grade.tokens || []) {
    if (tok.status === 'ok') {
      out.push({ kind: 'ok', top: '', segs: [{ t: tok.text, kind: 'same' }], order: -1 })
      continue
    }
    if (tok.status === 'insert') {
      out.push({ kind: 'insert', top: (tok.correction || '').trim(), segs: [], order: mark++ })
      continue
    }
    // wrong
    const original = tok.text || ''
    const correction = (tok.correction || '').trim()
    if (!correction) {
      out.push({ kind: 'replace', top: '', segs: [{ t: original, kind: 'del' }], order: mark++ })
      continue
    }
    const diff = diffChars(original, correction)
    let common = 0
    for (const d of diff) if (!d.added && !d.removed) common += d.value.length
    const similar = common / Math.max(original.length, correction.length, 1) >= SIMILAR_THRESHOLD
    if (similar) {
      // Edit in place: keep shared letters, strike removed ones (red), insert added ones
      // (green) exactly where they belong — so a missing letter shows as an insertion at the
      // right spot instead of striking a correct neighbour.
      const segs: Seg[] = diff.map((d) => ({
        t: d.value,
        kind: d.added ? 'ins' : d.removed ? 'del' : 'same',
      }))
      out.push({ kind: 'wrong', top: '', segs, order: mark++ })
    } else {
      out.push({ kind: 'replace', top: correction, segs: [{ t: original, kind: 'del' }], order: mark++ })
    }
  }
  return out
})
</script>

<style scoped>
.correction-line {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 10px 8px;
  font-family: 'Lora', Georgia, serif;
  font-size: 22px;
  line-height: 1.25;
  color: var(--text);
  /* pen-stroke timing (teacher marks words left-to-right, one after another) */
  --base: 220ms;   /* wait for the card to settle */
  --step: 380ms;   /* gap between consecutive marked words */
  --strike: 240ms; /* time to draw one strikethrough */
  --write: 320ms;  /* time to write one correction */
}

.sg-tok {
  display: inline-flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}
/* reserve room for the word written above so rows never collide */
.sg-tok--replace,
.sg-tok--insert {
  padding-top: 1.05em;
}

.sg-top {
  position: absolute;
  top: -0.1em;
  left: 50%;
  transform: translateX(-50%) rotate(-2deg);
  font-size: 0.74em;
  line-height: 1;
  white-space: nowrap;
  color: var(--correct-ink, #2f9d57);
}

.sg-body { white-space: nowrap; }

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

/* inserted letters: green, written in place with a small caret marking the spot */
.sg-ins {
  position: relative;
  color: var(--correct-ink, #2f9d57);
  border-bottom: 1.5px solid var(--correct-ink, #2f9d57);
  animation: sg-write var(--write) ease-out both;
  animation-delay: calc(var(--base) + var(--o, 0) * var(--step) + var(--strike) * 0.6);
}

.sg-tok--insert .sg-caret { color: var(--correct-ink, #2f9d57); font-weight: 700; }

/* write each correction (green, above) just after striking the same word */
.sg-tok--replace .sg-top,
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
  .sg-tok--replace .sg-top,
  .sg-tok--insert .sg-top,
  .sg-tok--insert .sg-caret,
  .sg-ins {
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
