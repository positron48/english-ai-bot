import type { RouteLocationRaw } from 'vue-router'
import type { CourseMapItem, CourseMapModule, DailyRouteItem } from '../api/courseClient'

type NavigableLearningItem = Pick<CourseMapItem, 'source_kind' | 'source_id' | 'type'>
type NavigableModule = Pick<CourseMapModule, 'source_kind' | 'source_id' | 'type'>
type NavigableRouteItem = Pick<DailyRouteItem, 'source_kind' | 'source_id' | 'type'>

export type LinglowNavigable = NavigableLearningItem | NavigableModule | NavigableRouteItem

export function routeForLinglowItem(item: LinglowNavigable): RouteLocationRaw {
  const sourceKind = item.source_kind || item.type
  const sourceID = item.source_id || ''

  switch (sourceKind) {
    case 'grammar_chapter':
      return sourceID ? { name: 'GrammarChapter', params: { chapterId: sourceID } } : { name: 'LearningGrammar' }
    case 'grammar_theory_block':
      return sourceID ? { name: 'GrammarChapter', params: { chapterId: sourceID.split(':')[0] } } : { name: 'GrammarTraining' }
    case 'reading_text':
      return sourceID ? { name: 'ReadingText', params: { textId: sourceID } } : { name: 'ReadingCategories' }
    case 'speaking_task':
      return { name: 'SpeakingCategories' }
    case 'word_set':
      return sourceID ? { name: 'WordSetDetail', params: { setId: sourceID } } : { name: 'WordSets' }
    case 'word_card':
    case 'word':
      return { name: 'Training' }
    default:
      if (item.type === 'grammar_chapter') return { name: 'LearningGrammar' }
      if (item.type === 'reading_text') return { name: 'ReadingCategories' }
      if (item.type === 'speaking_task') return { name: 'SpeakingCategories' }
      if (item.type === 'word' || item.type === 'word_card') return { name: 'Training' }
      return { name: 'City' }
  }
}
