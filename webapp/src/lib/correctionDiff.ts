// Deterministic teacher-style correction diff between the learner's answer and the corrected
// sentence. Produces a minimal set of marks:
//   - equal   : an unchanged word (rendered plainly)
//   - del     : an extra word the learner added (struck whole)
//   - edit    : a word kept but fixed at character level — only the wrong runs of letters are
//               struck (red), missing letters are inserted (green) at their exact position;
//               correct letters are NEVER struck
//   - replace : a word swapped for a different word — struck, correct written above
//   - insert  : one or more missing words — written above with an arrow pointing to the gap
//
// Words are aligned with Needleman-Wunsch using a character-similarity substitution cost, so
// close words ("cibolla"→"cebolla", "du"→"dulce") pair up and become in-word edits instead of
// a whole-block strike. The same logic is mirrored in webapp/public/grading-preview.html.

export type SegKind = 'same' | 'del' | 'ins'
export interface CharSeg { t: string; kind: SegKind }

export type RenderToken =
  | { kind: 'equal'; text: string }
  | { kind: 'del'; text: string }
  | { kind: 'edit'; segs: CharSeg[] }
  | { kind: 'replace'; from: string; to: string }
  | { kind: 'insert'; text: string }

// An aligned word pair renders as an in-word edit (keep + fix letters) when at least this
// fraction of the learner's characters are correct; otherwise the word is struck whole and the
// correction written above.
const EDIT_KEEP_RATIO = 0.5

function splitWords(s: string): string[] {
  return s.trim().split(/\s+/).filter(Boolean)
}

function stripPunct(w: string): string {
  return w.replace(/^[^\p{L}\p{N}]+|[^\p{L}\p{N}]+$/gu, '')
}

// Alignment key: lowercased, accents stripped, surrounding punctuation removed.
function alignKey(w: string): string {
  return stripPunct(w.toLowerCase().normalize('NFD').replace(/\p{Diacritic}/gu, ''))
}

// Case key: lowercased + de-punctuated but ACCENTS KEPT. Two aligned words equal here differ
// only by capitalization/punctuation — never an error — so they render plainly.
function caseKey(w: string): string {
  return stripPunct(w.toLowerCase())
}

// Character-level LCS diff, with consecutive same-kind characters merged into one run (so a
// run of wrong letters is a single strike, not one strike per letter).
export function charDiff(a: string, b: string): CharSeg[] {
  const n = a.length
  const m = b.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      dp[i][j] = a[i] === b[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1])
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

// Pair (and later edit) two different words only when most of the learner's letters are kept;
// otherwise the word is too different to be a typo of the other and should not be paired.
const PAIR_KEEP_RATIO = 0.6

// Substitution cost for aligning u[i] with c[j]. 0 for the same word; a small value when they
// are a near-miss (so they pair up and become an in-word edit); a large value (> delete+insert)
// when too different, so the alignment uses delete+insert instead and frees each word to match
// its true counterpart elsewhere (e.g. keeps "es" free to match "Es" instead of pairing
// "Este"→"Es").
function subCost(a: string, b: string): number {
  if (alignKey(a) === alignKey(b)) return 0
  const kept = commonChars(a, b)
  if (kept / Math.max(a.length, 1) < PAIR_KEEP_RATIO) return 3
  return 1 - kept / Math.max(a.length, b.length, 1)
}

interface AlignOp { type: 'pair' | 'del' | 'ins'; u?: string; c?: string }

function wordAlign(u: string[], c: string[]): AlignOp[] {
  const n = u.length
  const m = c.length
  const GAP = 1
  const dp: number[][] = Array.from({ length: n + 1 }, () => new Array(m + 1).fill(0))
  for (let i = 1; i <= n; i++) dp[i][0] = i * GAP
  for (let j = 1; j <= m; j++) dp[0][j] = j * GAP
  for (let i = 1; i <= n; i++) {
    for (let j = 1; j <= m; j++) {
      dp[i][j] = Math.min(
        dp[i - 1][j - 1] + subCost(u[i - 1], c[j - 1]),
        dp[i - 1][j] + GAP,
        dp[i][j - 1] + GAP,
      )
    }
  }
  const ops: AlignOp[] = []
  let i = n
  let j = m
  while (i > 0 && j > 0) {
    const diag = dp[i - 1][j - 1] + subCost(u[i - 1], c[j - 1])
    const up = dp[i - 1][j] + GAP
    const left = dp[i][j - 1] + GAP
    if (diag <= up && diag <= left) {
      ops.push({ type: 'pair', u: u[i - 1], c: c[j - 1] }); i--; j--
    } else if (up <= left) {
      ops.push({ type: 'del', u: u[i - 1] }); i--
    } else {
      ops.push({ type: 'ins', c: c[j - 1] }); j--
    }
  }
  while (i > 0) { ops.push({ type: 'del', u: u[i - 1] }); i-- }
  while (j > 0) { ops.push({ type: 'ins', c: c[j - 1] }); j-- }
  ops.reverse()
  return ops
}

export function renderCorrection(userInput: string, corrected: string): RenderToken[] {
  const ops = wordAlign(splitWords(userInput), splitWords(corrected))
  const out: RenderToken[] = []
  let inss: string[] = []

  const flushIns = () => {
    if (inss.length) {
      // Adjacent insertions are coalesced into one group (one arrow), never two in a row.
      out.push({ kind: 'insert', text: inss.join(' ') })
      inss = []
    }
  }

  for (const op of ops) {
    if (op.type === 'ins') {
      inss.push(op.c as string)
      continue
    }
    flushIns()
    if (op.type === 'del') {
      out.push({ kind: 'del', text: op.u as string })
      continue
    }
    // paired word: plain when identical or only case/punctuation differs; otherwise an in-word
    // edit (keep correct letters) or, when most of the word is wrong, a whole-word replacement.
    const u = op.u || ''
    const c = op.c || ''
    if (u === c || caseKey(u) === caseKey(c)) {
      out.push({ kind: 'equal', text: c })
      continue
    }
    const segs = charDiff(u, c)
    let kept = 0
    for (const s of segs) if (s.kind === 'same') kept += s.t.length
    if (kept / Math.max(u.length, 1) >= EDIT_KEEP_RATIO) {
      out.push({ kind: 'edit', segs })
    } else {
      out.push({ kind: 'replace', from: u, to: c })
    }
  }
  flushIns()
  return out
}
