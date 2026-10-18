import { describe, expect, it } from 'vitest'
import type { LocationQuery } from 'vue-router'
import { createJourneyFixture } from '../dev/dailyJourneyFixture'
import { parseDailyJourneyContext, dailyJourneyExerciseRoute, dailyJourneyReturnRoute, verifyDailyJourneyContext } from './useDailyJourneyContext'
describe('dailyJourney context for existing screens', () => {
  it('keeps standalone navigation separate and rejects partial or repeated query parameters', () => {
    expect(parseDailyJourneyContext({ course_code: 'en_ru' })).toBeNull()
    const invalidQueries: LocationQuery[] = [
      { journey_id: 'id' }, { journey_step_id: 'step', course_code: 'en_ru' },
      { journey_id: ['a', 'b'], journey_step_id: 'step', course_code: 'en_ru' },
    ]
    for (const query of invalidQueries) expect(() => parseDailyJourneyContext(query)).toThrow('Некорректный контекст')
  })
  it('opens the existing screen with explicit course and attempt context', () => {
    const j = createJourneyFixture('es_ru', 'short')
    const step = j.steps[0]
    step.status = 'active'
    step.launch = { routeName: 'Training' }
    const context = { journeyId: j.id, stepId: step.id, courseCode: 'es_ru' }
    expect(dailyJourneyExerciseRoute(j, step)).toEqual({ name: 'Training', params: {}, query: { course_code: 'es_ru', journey_id: j.id, journey_step_id: step.id } })
    expect(verifyDailyJourneyContext(context, j, 'Training', {})).toBe(step)
    expect(dailyJourneyReturnRoute(context)).toEqual({ name: 'CityDailyRoute', query: { course_code: 'es_ru', journey_id: j.id } })
    expect(() => verifyDailyJourneyContext({ ...context, courseCode: 'en_ru' }, j, 'Training', {})).toThrow()
  })
  it('rejects future steps, other content and arbitrary routes', () => {
    const j = createJourneyFixture('en_ru', 'short')
    j.steps[0].status = 'completed'
    const step = j.steps[1]
    step.status = 'active'
    step.launch = { routeName: 'ReadingText', params: { textId: 'text-a' } }
    const context = { journeyId: j.id, stepId: step.id, courseCode: j.courseCode }
    expect(() => verifyDailyJourneyContext(context, j, 'ReadingText', { textId: 'text-b' })).toThrow('не соответствует')
    expect(() => dailyJourneyExerciseRoute(j, j.steps[0])).toThrow()
    step.launch = { routeName: 'Chat' }
    expect(() => dailyJourneyExerciseRoute(j, step)).toThrow('Некорректный экран')
  })
})
