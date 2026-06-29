// Compares dotted numeric version strings such as "0.12.49".
// Returns -1 if a < b, 1 if a > b, 0 if equal. Missing trailing
// segments are treated as 0, so "0.12" == "0.12.0".
export const compareVersions = (a: string, b: string): -1 | 0 | 1 => {
  const pa = normalize(a)
  const pb = normalize(b)
  const len = Math.max(pa.length, pb.length)
  for (let i = 0; i < len; i++) {
    const na = pa[i] ?? 0
    const nb = pb[i] ?? 0
    if (na < nb) return -1
    if (na > nb) return 1
  }
  return 0
}

const normalize = (v: string): number[] => {
  return String(v ?? '')
    .trim()
    .replace(/^v/i, '')
    .split('.')
    .map((seg) => {
      const n = parseInt(seg, 10)
      return Number.isFinite(n) ? n : 0
    })
}
