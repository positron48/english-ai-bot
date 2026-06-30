// Deterministic teacher-style correction diff between the learner's answer and the corrected
// sentence. Produces a minimal set of marks:
//   - equal   : an unchanged word (rendered plainly)
//   - del     : an extra word the learner added (struck whole)
//   - edit    : a single near-miss word, edited in place at character level — only the wrong
//               runs of letters are struck (red), missing letters are inserted (green) at their
//               exact position; correct letters are NEVER struck
//   - replace : one or more words swapped for different word(s) — struck, correct written above
//   - insert  : one or more missing words — written above with an arrow pointing to the gap
//
// The same logic is mirrored verbatim in webapp/public/grading-preview.html for visual
// validation. Keep the two in sync.

export type SegKind = 'same' | 'del' | 'ins'
export interface CharSeg { t: string; kind: SegKind }

export type RenderToken =
  | { kind: 'equal'; text: string }
  | { kind: 'del'; text: string }
  | { kind: 'edit'; segs: CharSeg[] }
  | { kind: 'replace'; from: string; to: string }
  | { kind: 'insert'; text: string }

// A single word is edited in place (not struck whole) only when it still shares at least this
// fraction of characters with its correction — so typos/inflections stay inline while genuine
// word swaps ("Este"→"Es") are struck whole.
const SIMILAR_THRESHOLD = 0.5

function splitWords(s: string): string[] {
  return s.trim().split(/\s+/).filter(Boolean)
}

function stripPunct(w: string): string {
  return w.replace(/^[^\p{L}\p{N}]+|[^\p{L}\p{N}]+$/gu, '')
}

// Alignment key: lowercased, accents stripped, surrounding punctuation removed. Used only to
// PAIR words across the two sentences, so "esta"/"está" and "El"/"Él" align as the same slot
// (their difference is then shown as an in-word edit rather than a whole-word swap).
function alignKey(w: string): string {
  return stripPunct(w.toLowerCase().normalize('NFD').replace(/\p{Diacritic}/gu, ''))
}

// Case key: lowercased + de-punctuated but ACCENTS KEPT. When two aligned words match here,
// they differ only by capitalization/punctuation — never an error — so they render plainly.
function caseKey(w: string): string {
  return stripPunct(w.toLowerCase())
}

interface WordOp { type: 'equal' | 'del' | 'ins'; u?: string; c?: string }

// Longest-common-subsequence word alignment between user words (u) and corrected words (c).
function wordDiff(u: string[], c: string[]): WordOp[] {
  const n = u.length
  const m = c.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      dp[i][j] = alignKey(u[i]) === alignKey(c[j])
        ? dp[i + 1][j + 1] + 1
        : Math.max(dp[i + 1][j], dp[i][j + 1])
    }
  }
  const ops: WordOp[] = []
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (alignKey(u[i]) === alignKey(c[j])) {
      ops.push({ type: 'equal', u: u[i], c: c[j] }); i++; j++
    } else if (dp[i + 1][j] >= dp[i][j + 1]) {
      ops.push({ type: 'del', u: u[i] }); i++
    } else {
      ops.push({ type: 'ins', c: c[j] }); j++
    }
  }
  while (i < n) ops.push({ type: 'del', u: u[i++] })
  while (j < m) ops.push({ type: 'ins', c: c[j++] })
  return ops
}

// Character-level LCS diff, with consecutive same-kind characters merged into one run (so a
// run of wrong letters is a single strike, not one strike per letter).
export function charDiff(a: string, b: string): CharSeg[] {
  const n = a.length
  const m = b.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      dp[i][j] = a[i] === b[j]
        ? dp[i + 1][j + 1] + 1
        : Math.max(dp[i + 1][j], dp[i][j + 1])
    }
  }
  const raw: CharSeg[] = []
  let i = 0
  let j = 0
  while (i < n && j < m) {
    if (a[i] === b[j]) { raw.push({ t: a[i], kind: 'same' }); i++; j++ }
    else if (dp[i + 1][j] >= dp[i][j + 1]) { raw.push({ t: a[i], kind: 'del' }); i++ }
    else { raw.push({ t: b[j], kind: 'ins' }); j++ }
  }
  while (i < n) raw.push({ t: a[i++], kind: 'del' })
  while (j < m) raw.push({ t: b[j++], kind: 'ins' })
  const segs: CharSeg[] = []
  for (const s of raw) {
    const last = segs[segs.length - 1]
    if (last && last.kind === s.kind) last.t += s.t
    else segs.push({ ...s })
  }
  return segs
}

function commonChars(a: string, b: string): number {
  let c = 0
  for (const s of charDiff(a, b)) if (s.kind === 'same') c += s.t.length
  return c
}

export function renderCorrection(userInput: string, corrected: string): RenderToken[] {
  const ops = wordDiff(splitWords(userInput), splitWords(corrected))
  const out: RenderToken[] = []
  let dels: string[] = []
  let inss: string[] = []

  const flush = () => {
    if (dels.length && inss.length) {
      const single = dels.length === 1 && inss.length === 1
      const ratio = single ? commonChars(dels[0], inss[0]) / Math.max(dels[0].length, inss[0].length, 1) : 0
      if (single && ratio >= SIMILAR_THRESHOLD) {
        out.push({ kind: 'edit', segs: charDiff(dels[0], inss[0]) })
      } else {
        out.push({ kind: 'replace', from: dels.join(' '), to: inss.join(' ') })
      }
    } else if (dels.length) {
      for (const d of dels) out.push({ kind: 'del', text: d })
    } else if (inss.length) {
      // Adjacent insertions are coalesced into one group (one arrow), never two in a row.
      out.push({ kind: 'insert', text: inss.join(' ') })
    }
    dels = []
    inss = []
  }

  for (const op of ops) {
    if (op.type === 'equal') {
      flush()
      const u = op.u || ''
      const c = op.c || ''
      // Same word slot: plain when identical or only case/punctuation differs (not an error);
      // otherwise an in-word edit (e.g. a missing accent "esta"→"está").
      if (u === c || caseKey(u) === caseKey(c)) {
        out.push({ kind: 'equal', text: c })
      } else {
        out.push({ kind: 'edit', segs: charDiff(u, c) })
      }
    } else if (op.type === 'del') {
      dels.push(op.u as string)
    } else {
      inss.push(op.c as string)
    }
  }
  flush()
  return out
}
