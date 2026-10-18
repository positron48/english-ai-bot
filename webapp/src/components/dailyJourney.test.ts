import { describe, it, expect } from 'vitest'
import { createI18n } from 'vue-i18n'
import ru from '../locales/ru.json'
import { config, mount } from '@vue/test-utils'
import DailyJourneyShell from './DailyJourneyShell.vue'
import DailyJourneyCard from './DailyJourneyCard.vue'
import { createFixtureTransport, createJourneyFixture } from '../dev/dailyJourneyFixture'
import type { JourneyActionType } from './dailyJourney/types'
config.global.plugins = [createI18n({ legacy: false, locale: 'ru', messages: { ru } })]
function storage() {
  const data = new Map<string, string>()
  return { getItem: (key: string) => data.get(key) ?? null, setItem: (key: string, value: string) => { data.set(key, value) } }
}
describe('dailyJourney presentation', () => {
  it('shows resume position and a completed summary without a new start', () => {
    const journey = createJourneyFixture('en_ru', 'full')
    journey.steps[0].status = 'completed'
    journey.steps[1].status = 'active'
    const card = mount(DailyJourneyCard, { props: { journey }, global: { plugins: [createI18n({ legacy: false, locale: 'ru', messages: { ru } })] } })
    expect(card.text()).toContain('Продолжить — остановка 2 из 5')
    journey.status = 'completed'
    const summary = mount(DailyJourneyShell, { props: { journey, busy: false } })
    expect(summary.text()).toContain('Маршрут пройден')
    expect(summary.findAll('button')).toHaveLength(0)
  })
  it('opens the usual exercise screen without rendering questions or feedback in the route', async () => {
    const journey = createJourneyFixture('en_ru', 'full')
    journey.steps[0].status = 'active'
    journey.steps[0].feedback = 'Stored subject feedback'
    const view = mount(DailyJourneyShell, { props: { journey, busy: false } })
    expect(view.text()).not.toContain('Stored subject feedback')
    expect(view.findAll('input, form, progress')).toHaveLength(0)
    expect(view.text()).not.toContain('Пропустить этап')
    const open = view.findAll('button').find(button => button.text().includes('Продолжить задание'))!
    await open.trigger('click')
    expect(view.emitted('open-step')?.[0]).toEqual([journey.steps[0]])
  })
  it('offers exactly one optional skip and blocks actions while retry is pending', () => {
    const journey = createJourneyFixture('en_ru', 'full')
    journey.steps[0].status = journey.steps[1].status = 'completed'
    const view = mount(DailyJourneyShell, { props: { journey, busy: false, hasPendingAction: true } })
    expect(view.text()).toContain('Остался 1 пропуск')
    expect(view.findAll('button').every(button => button.attributes('disabled') !== undefined)).toBe(true)
    journey.steps[2].status = 'skipped'
    const after = mount(DailyJourneyShell, { props: { journey, busy: false } })
    expect(after.text()).not.toContain('Пропустить этап')
    expect(after.text()).toContain('Пропуск использован')
  })
})
describe('dailyJourney fixture state transitions', () => {
  it('enforces linear steps, mandatory stops and a single skip', async () => {
    const transport = createFixtureTransport('linear', 'en_ru', 'full', storage())
    let journey = await transport.start()
    let serial = 0
    const act = async (action: JourneyActionType, stepId: string, text = '') => {
      journey = await transport.act(journey.id, stepId, { client_action_id: `action-${serial++}`, expected_revision: journey.revision, action, payload: { text } })
    }
    await expect(act('start', 'step-2')).rejects.toThrow('недоступна')
    await expect(act('skip', 'step-0')).rejects.toThrow('нельзя')
    for (const id of ['step-0', 'step-1']) {
      await act('start', id)
      for (let n = 0; n < 2; n++) {
        await act('answer', id, 'answer')
        await expect(act('answer', id, 'duplicate')).rejects.toThrow('разбор')
        await act('ack_feedback', id)
      }
    }
    await act('skip', 'step-2')
    await expect(act('skip', 'step-3')).rejects.toThrow('нельзя')
    expect(journey.steps[2].result).toBeNull()
    expect((await transport.get())?.revision).toBe(journey.revision)
  })
  it('rejects reused keys with different payloads and stale revisions', async () => {
    const transport = createFixtureTransport('receipts', 'es_ru', 'short', storage())
    const journey = await transport.start()
    const request = { client_action_id: 'same', expected_revision: 0, action: 'start' as const, payload: {} }
    const first = await transport.act(journey.id, 'step-0', request)
    expect(await transport.act(journey.id, 'step-0', request)).toEqual(first)
    await expect(transport.act(journey.id, 'step-0', { ...request, payload: { text: 'different' } })).rejects.toThrow('другим ответом')
    await expect(transport.act(journey.id, 'step-0', { ...request, client_action_id: 'stale' })).rejects.toThrow('другой вкладке')
  })
})
