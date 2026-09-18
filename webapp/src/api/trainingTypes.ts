export interface Card {
  question: string
  card_index: number
  total_cards: number
  session_id: number
  user_card_id: number
  delay_ms: number
  direction: string
  word_en?: string
  word_target?: string
  transcription?: string
  display_word?: string
  display_target?: string
  example_en?: string
  example_target?: string
  /** Spell challenge: compose word from letters; type: type-the-word (no letters) */
  type?: 'card' | 'spell' | 'type'
  word_ru?: string
  /** Non-editable prefix for spell (e.g. "to " for verbs); user composes the rest */
  prefix?: string
  letters?: string[]
  correct_answer?: string
  /** Type challenge: first letter for hint (on demand) */
  hint_first_letter?: string
  /** Type challenge: word length for hint */
  hint_length?: number
  morph?: MorphInfo
  word_card_id?: number
  training_card_id?: number
  word_category?: string
  offline?: boolean
}

export interface MorphVerbForms {
  v1?: string
  v2?: string
  v3?: string
}

export interface MorphInfo {
  pos?: string
  noun_gender?: string
  article?: string
  opposite_gender_word?: string
  verb_forms?: MorphVerbForms
}

export interface OptionsResponse {
  options: string[]
  user_card_id: number
  session_id?: number
  card_index?: number
}

export interface Feedback {
  is_correct: boolean
  chosen_option: string
  correct_answer: string
  hint?: string
  example?: string
  example_target?: string
  delay_seconds?: number
}
