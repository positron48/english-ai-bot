<template>
  <div class="correction-line">
    <span v-for="(tok, i) in rendered" :key="i" class="sg-tok" :class="`sg-tok--${tok.kind}`">
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
}

// Build a teacher-markup model: for "wrong" tokens we diff at the character level so only
// the differing letters are struck (red) and only the corrected letters appear above (green).
// "insert" tokens render a caret with the missing word written above.
const rendered = computed<RenderTok[]>(() => {
  const out: RenderTok[] = []
  for (const tok of props.grade.tokens || []) {
    if (tok.status === 'ok') {
      out.push({ kind: 'ok', top: '', parts: [{ t: tok.text, struck: false }] })
      continue
    }
    if (tok.status === 'insert') {
      out.push({ kind: 'insert', top: (tok.correction || '').trim(), parts: [] })
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
    out.push({ kind: 'wrong', top, parts })
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
  transform: translateX(-50%);
  font-size: 0.72em;
  line-height: 1;
  white-space: nowrap;
  color: var(--correct-ink, #2f9d57);
}

.sg-bottom {
  white-space: nowrap;
}

.sg-del {
  color: var(--wrong-ink, #c4443c);
  text-decoration: line-through;
  text-decoration-thickness: 2px;
}

.sg-tok--insert .sg-caret {
  color: var(--correct-ink, #2f9d57);
  font-weight: 700;
}
</style>
