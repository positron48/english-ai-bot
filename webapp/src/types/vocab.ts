export interface CardDetail {
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
  mastery_level?: string
  mastering_score?: number
}
