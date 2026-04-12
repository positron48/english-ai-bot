/**
 * Subject pronoun labels for Spanish verb paradigm tables (Peninsular: 2pl = vosotros).
 * person: "1" | "2" | "3" from API; number: "singular" | "plural".
 */
export function spanishVerbSubjectPronoun(person: string, number: string): string {
  const p = String(person ?? '')
    .trim()
    .toLowerCase()
  const n = String(number ?? '')
    .trim()
    .toLowerCase()

  const isSing = n === 'singular' || n === 'sing' || n === 's'
  const isPl = n === 'plural' || n === 'pl' || n === 'p'

  const is1 = p === '1' || p === '1st' || p === 'first'
  const is2 = p === '2' || p === '2nd' || p === 'second'
  const is3 = p === '3' || p === '3rd' || p === 'third'

  if (is1 && isSing) return 'yo'
  if (is2 && isSing) return 'tú'
  if (is3 && isSing) return 'él / ella / usted'
  if (is1 && isPl) return 'nosotros / nosotras'
  if (is2 && isPl) return 'vosotros / vosotras'
  if (is3 && isPl) return 'ellos / ellas / ustedes'

  const raw = [String(person).trim(), String(number).trim()].filter(Boolean)
  if (raw.length === 0) return '—'
  return raw.join(' · ')
}
