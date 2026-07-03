// Cheat-sheet vocabulary for the "describe a picture" quest. Content is tied to the COURSE
// (its target + native language), NOT to the UI locale: a Spanish course always shows the
// RU→ES sheet, an English course the RU→EN one. `term` is in the target language, `gloss`
// in the course's native language. Extend `PHRASEBOOK[target][native]` for new course pairs.

export interface HintEntry {
  term: string
  gloss: string
}
export interface HintSection {
  key: string
  title: string // in the native language
  icon: string  // emoji, locale-neutral
  entries: HintEntry[]
}

type NativeMap = Record<string, HintSection[]>

const ES_RU: HintSection[] = [
  {
    key: 'location', title: 'Расположение', icon: '📍',
    entries: [
      { term: 'arriba', gloss: 'сверху / вверху' },
      { term: 'abajo', gloss: 'снизу / внизу' },
      { term: 'a la izquierda', gloss: 'слева' },
      { term: 'a la derecha', gloss: 'справа' },
      { term: 'en el centro', gloss: 'в центре' },
      { term: 'encima de', gloss: 'на / над' },
      { term: 'debajo de', gloss: 'под' },
      { term: 'delante de', gloss: 'перед' },
      { term: 'detrás de', gloss: 'позади / сзади' },
      { term: 'al lado de', gloss: 'рядом с' },
      { term: 'entre', gloss: 'между' },
      { term: 'dentro de', gloss: 'внутри' },
      { term: 'fuera de', gloss: 'снаружи' },
      { term: 'cerca', gloss: 'близко' },
      { term: 'lejos', gloss: 'далеко' },
      { term: 'en la esquina', gloss: 'в углу' },
    ],
  },
  {
    key: 'colors', title: 'Цвета', icon: '🎨',
    entries: [
      { term: 'rojo', gloss: 'красный' },
      { term: 'azul', gloss: 'синий' },
      { term: 'verde', gloss: 'зелёный' },
      { term: 'amarillo', gloss: 'жёлтый' },
      { term: 'naranja', gloss: 'оранжевый' },
      { term: 'morado', gloss: 'фиолетовый' },
      { term: 'rosa', gloss: 'розовый' },
      { term: 'marrón', gloss: 'коричневый' },
      { term: 'negro', gloss: 'чёрный' },
      { term: 'blanco', gloss: 'белый' },
      { term: 'gris', gloss: 'серый' },
      { term: 'claro', gloss: 'светлый' },
      { term: 'oscuro', gloss: 'тёмный' },
    ],
  },
  {
    key: 'sizes', title: 'Размеры и форма', icon: '📐',
    entries: [
      { term: 'grande', gloss: 'большой' },
      { term: 'pequeño', gloss: 'маленький' },
      { term: 'largo', gloss: 'длинный' },
      { term: 'corto', gloss: 'короткий' },
      { term: 'alto', gloss: 'высокий' },
      { term: 'bajo', gloss: 'низкий' },
      { term: 'ancho', gloss: 'широкий' },
      { term: 'estrecho', gloss: 'узкий' },
      { term: 'redondo', gloss: 'круглый' },
      { term: 'cuadrado', gloss: 'квадратный' },
    ],
  },
  {
    key: 'quantity', title: 'Количество', icon: '🔢',
    entries: [
      { term: 'uno', gloss: 'один' },
      { term: 'dos', gloss: 'два' },
      { term: 'tres', gloss: 'три' },
      { term: 'muchos', gloss: 'много' },
      { term: 'pocos', gloss: 'мало' },
      { term: 'algunos', gloss: 'несколько' },
      { term: 'varios', gloss: 'несколько / разные' },
    ],
  },
  {
    key: 'phrases', title: 'Полезные фразы', icon: '💬',
    entries: [
      { term: 'hay', gloss: 'есть / имеется' },
      { term: 'veo…', gloss: 'я вижу…' },
      { term: 'está', gloss: 'находится' },
      { term: 'en el fondo', gloss: 'на заднем плане' },
      { term: 'en primer plano', gloss: 'на переднем плане' },
      { term: 'a la izquierda de', gloss: 'слева от' },
      { term: 'a la derecha de', gloss: 'справа от' },
    ],
  },
]

const EN_RU: HintSection[] = [
  {
    key: 'location', title: 'Расположение', icon: '📍',
    entries: [
      { term: 'above', gloss: 'сверху / над' },
      { term: 'below', gloss: 'снизу / под' },
      { term: 'on the left', gloss: 'слева' },
      { term: 'on the right', gloss: 'справа' },
      { term: 'in the middle', gloss: 'в центре / посередине' },
      { term: 'on', gloss: 'на' },
      { term: 'under', gloss: 'под' },
      { term: 'in front of', gloss: 'перед' },
      { term: 'behind', gloss: 'позади / сзади' },
      { term: 'next to', gloss: 'рядом с' },
      { term: 'between', gloss: 'между' },
      { term: 'inside', gloss: 'внутри' },
      { term: 'outside', gloss: 'снаружи' },
      { term: 'near', gloss: 'близко' },
      { term: 'far', gloss: 'далеко' },
      { term: 'in the corner', gloss: 'в углу' },
    ],
  },
  {
    key: 'colors', title: 'Цвета', icon: '🎨',
    entries: [
      { term: 'red', gloss: 'красный' },
      { term: 'blue', gloss: 'синий' },
      { term: 'green', gloss: 'зелёный' },
      { term: 'yellow', gloss: 'жёлтый' },
      { term: 'orange', gloss: 'оранжевый' },
      { term: 'purple', gloss: 'фиолетовый' },
      { term: 'pink', gloss: 'розовый' },
      { term: 'brown', gloss: 'коричневый' },
      { term: 'black', gloss: 'чёрный' },
      { term: 'white', gloss: 'белый' },
      { term: 'grey', gloss: 'серый' },
      { term: 'light', gloss: 'светлый' },
      { term: 'dark', gloss: 'тёмный' },
    ],
  },
  {
    key: 'sizes', title: 'Размеры и форма', icon: '📐',
    entries: [
      { term: 'big', gloss: 'большой' },
      { term: 'small', gloss: 'маленький' },
      { term: 'long', gloss: 'длинный' },
      { term: 'short', gloss: 'короткий' },
      { term: 'tall', gloss: 'высокий' },
      { term: 'wide', gloss: 'широкий' },
      { term: 'narrow', gloss: 'узкий' },
      { term: 'round', gloss: 'круглый' },
      { term: 'square', gloss: 'квадратный' },
    ],
  },
  {
    key: 'quantity', title: 'Количество', icon: '🔢',
    entries: [
      { term: 'one', gloss: 'один' },
      { term: 'two', gloss: 'два' },
      { term: 'three', gloss: 'три' },
      { term: 'many', gloss: 'много' },
      { term: 'few', gloss: 'мало' },
      { term: 'some', gloss: 'несколько' },
      { term: 'several', gloss: 'несколько' },
    ],
  },
  {
    key: 'phrases', title: 'Полезные фразы', icon: '💬',
    entries: [
      { term: 'there is / there are', gloss: 'есть / имеется' },
      { term: 'I can see…', gloss: 'я вижу…' },
      { term: 'it is', gloss: 'это / находится' },
      { term: 'in the background', gloss: 'на заднем плане' },
      { term: 'in the foreground', gloss: 'на переднем плане' },
      { term: 'to the left of', gloss: 'слева от' },
      { term: 'to the right of', gloss: 'справа от' },
    ],
  },
]

const PHRASEBOOK: Record<string, NativeMap> = {
  es: { ru: ES_RU },
  en: { ru: EN_RU },
}

// Title of the sheet per native language.
const SHEET_TITLE: Record<string, string> = {
  ru: 'Слова-подсказки',
}

export function getHintSections(targetLang: string, nativeLang: string): HintSection[] {
  const byNative = PHRASEBOOK[targetLang?.toLowerCase()]
  if (!byNative) return []
  return byNative[nativeLang?.toLowerCase()] || byNative.ru || []
}

export function getHintSheetTitle(nativeLang: string): string {
  return SHEET_TITLE[nativeLang?.toLowerCase()] || SHEET_TITLE.ru
}
