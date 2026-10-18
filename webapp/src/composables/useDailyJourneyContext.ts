import type { LocationQuery, RouteLocationRaw } from 'vue-router'
import { currentJourneyStep, type DailyJourney, type JourneyStep } from '../components/dailyJourney/types'

export interface DailyJourneyContext { journeyId: string; stepId: string; courseCode: string }

// An invalid/partial context is an error, never a signal to start a normal session.
export function parseDailyJourneyContext(query: LocationQuery): DailyJourneyContext | null {
  if (query.journey_id === undefined && query.journey_step_id === undefined) return null
  const { journey_id: journeyId, journey_step_id: stepId, course_code: courseCode } = query
  if (typeof journeyId !== 'string' || !journeyId || typeof stepId !== 'string' || !stepId || typeof courseCode !== 'string' || !courseCode || journeyId.length > 100 || stepId.length > 100) {
    throw new Error('Некорректный контекст маршрута. Вернитесь на маршрут дня.')
  }
  return { journeyId, stepId, courseCode }
}

const launchRoutes: Record<JourneyStep['mode'], Record<string, string[]>> = {
  words: { Training: [], WordSets: [] }, reading: { ReadingText: ['textId'] },
  grammar: { GrammarChapter: ['chapterId'], GrammarChapterTest: ['chapterId'], GrammarCategoryTest: ['sectionId'] },
  grammar_srs: { GrammarTraining: [] }, verb_forms: { VerbTraining: [] },
  picture: { PictureQuestChat: ['questCode'] }, conversation: { PlaceChat: ['districtCode', 'scenarioCode'] },
}

// Navigation is derived from a confirmed server response. A fixture without a
// bound attempt cannot manufacture a launch. Exercise APIs must recheck context.
export function dailyJourneyExerciseRoute(journey: DailyJourney, step: JourneyStep): RouteLocationRaw {
  if (journey.status !== 'active' || currentJourneyStep(journey)?.id !== step.id || step.status !== 'active' || !step.launch) throw new Error('Эта остановка сейчас недоступна.')
  const keys = launchRoutes[step.mode]?.[step.launch.routeName]
  const params = step.launch.params || {}
  if (!keys || Object.keys(params).length !== keys.length || keys.some(key => !params[key])) throw new Error('Некорректный экран задания.')
  return {
    name: step.launch.routeName, params,
    query: { course_code: journey.courseCode, journey_id: journey.id, journey_step_id: step.id },
  }
}

export function dailyJourneyReturnRoute(context: DailyJourneyContext): RouteLocationRaw {
  return { name: 'CityDailyRoute', query: { course_code: context.courseCode, journey_id: context.journeyId } }
}

export function verifyDailyJourneyContext(context: DailyJourneyContext, journey: DailyJourney, routeName: string, params: Record<string, unknown>): JourneyStep {
  const step = currentJourneyStep(journey)
  if (journey.id !== context.journeyId || journey.courseCode !== context.courseCode || !step || step.id !== context.stepId) throw new Error('Контекст маршрута изменился. Вернитесь на маршрут дня.')
  // Also enforces active journey, current step and allowed routes/parameter keys.
  dailyJourneyExerciseRoute(journey, step)
  if (step.mode === 'words' && step.launch!.routeName === 'WordSets' && ['WordSets', 'WordSetDetail', 'WordSetStudy'].includes(routeName)) {
    if (routeName !== 'WordSets' && (typeof params.setId !== 'string' || !params.setId)) throw new Error('Неверный словарный набор.')
    return step
  }
  if (step.launch!.routeName !== routeName || Object.entries(step.launch!.params || {}).some(([key, value]) => params[key] !== value)) throw new Error('Задание не соответствует текущей остановке.')
  return step
}
