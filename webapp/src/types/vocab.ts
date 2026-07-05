export interface CardDetail {
  id: number
  training_card_id: number
  direction: string
  state: string
  ef: number
  reps: number
  interval_days: number
  learning_step: number
  lapse_count: number
  next_due_at: string | null
  last_review_at: string | null
  last_quality: number | null
  created_at?: string
  updated_at?: string
  word_ru: string
  word_native?: string
  meaning_en: string
  meaning_target?: string
  example_en: string
  example_target?: string
  example_ru: string
  example_native?: string
  transcription: string
  sense_index: number
  pos?: string
  review_count: number
}

export interface MorphInfo {
  pos?: string
  noun_gender?: string
  article?: string
  opposite_gender_word?: string
  verb_forms?: Record<string, unknown>
}

export interface VocabCardsAPIResponse {
  lemma: string
  word_card_id: number
  cards: CardDetail[]
  verb_forms?: Record<string, unknown>
  pos?: string
  noun_gender?: string
  opposite_gender_word?: string
  morph?: MorphInfo
  has_user_cards?: boolean
  is_known?: boolean
}
