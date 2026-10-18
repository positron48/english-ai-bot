// Development-only transport. No learning APIs, SRS writes or provider calls.
import { currentJourneyStep, journeySkipAvailable, type DailyJourney, type JourneyAction, type JourneyMode, type JourneyStep, type JourneyTransport } from '../components/dailyJourney/types'

export type FixtureScenario = 'full' | 'short' | 'add_words' | 'unavailable' | 'verbs' | 'srs'
const modes: Record<JourneyMode, [string, string]> = {
  words: ['Слова', 'Повтори слова и посмотри разбор каждого ответа.'],
  reading: ['Чтение', 'Прочитай текст и ответь на вопросы.'],
  grammar: ['Грамматика', 'Изучи следующий блок и проверь понимание.'],
  picture: ['Картинка', 'Опиши изображение и получи обратную связь.'],
  conversation: ['Диалог', 'Выполни задачи разговора с NPC.'],
  verb_forms: ['Формы глаголов', 'Вспомни формы доступных испанских глаголов.'],
  grammar_srs: ['Повторение грамматики', 'Повтори вопросы из изученных глав.'],
}
export function createJourneyFixture(courseCode: string, scenario: FixtureScenario): DailyJourney {
  const extras: JourneyMode[] = scenario === 'short' || scenario === 'unavailable' ? [] : scenario === 'verbs' ? ['verb_forms', 'grammar', 'picture'] : scenario === 'srs' ? ['grammar_srs', 'picture', 'conversation'] : ['grammar', 'picture', 'conversation']
  const steps: JourneyStep[] = (['words', 'reading', ...extras] as JourneyMode[]).map((mode, index) => ({
    id: `step-${index}`, mode, title: modes[mode][0], goal: modes[mode][1], required: index < 2,
    status: scenario === 'unavailable' ? 'unavailable' : 'pending',
    targetCount: mode === 'words' && scenario === 'add_words' ? 10 : mode === 'picture' ? 1 : 2,
    completedCount: 0, feedback: null, result: null,
  }))
  return { id: `preview-${courseCode}-${scenario}`, courseCode, date: new Date().toISOString().slice(0, 10), revision: 0, status: scenario === 'unavailable' ? 'unavailable' : 'active', steps }
}

interface SavedFixture {
  journey: DailyJourney
  receipts: Record<string, { fingerprint: string; result: DailyJourney }>
}
function copy<T>(value: T): T { return JSON.parse(JSON.stringify(value)) }
export function createFixtureTransport(scope: string, courseCode: string, scenario: FixtureScenario, storage: Pick<Storage, 'getItem' | 'setItem'>): JourneyTransport {
  const key = `linglow:daily-journey-preview:v1:${scope}`
  function load(): SavedFixture | null {
    const raw = storage.getItem(key)
    if (!raw) return null
    try {
      const saved = JSON.parse(raw) as SavedFixture
      if (saved.journey.id !== `preview-${courseCode}-${scenario}` || !Array.isArray(saved.journey.steps) || !saved.receipts) return null
      return saved
    } catch { return null }
  }
  function save(saved: SavedFixture) { storage.setItem(key, JSON.stringify(saved)) }
  return {
    async get() { return load()?.journey ?? null },
    async start() {
      const existing = load()
      if (existing) return existing.journey
      const journey = createJourneyFixture(courseCode, scenario)
      save({ journey, receipts: {} })
      return journey
    },
    async act(id: string, stepId: string, request: JourneyAction) {
      const saved = load()
      if (!saved || saved.journey.id !== id) throw new Error('Маршрут не найден.')
      const fingerprint = JSON.stringify({ stepId, request })
      const receipt = saved.receipts[request.client_action_id]
      if (receipt) {
        if (receipt.fingerprint !== fingerprint) throw new Error('Ключ действия уже использован с другим ответом.')
        return copy(receipt.result)
      }
      const journey = saved.journey
      if (journey.revision !== request.expected_revision) throw new Error('Маршрут изменился в другой вкладке. Обновите состояние.')
      const step = currentJourneyStep(journey)
      if (journey.status !== 'active' || !step || step.id !== stepId) throw new Error('Эта остановка сейчас недоступна.')
      switch (request.action) {
        case 'start':
          if (step.status !== 'pending') throw new Error('Остановка уже начата.')
          step.status = 'active'
          break
        case 'skip':
          if (step.required || !journeySkipAvailable(journey)) throw new Error('Пропустить эту остановку нельзя.')
          step.status = 'skipped'
          step.feedback = null
          break
        case 'answer':
        case 'add_word':
        case 'message':
          if (step.status !== 'active' || step.feedback) throw new Error('Сначала прочитайте разбор.')
          if (typeof request.payload.text !== 'string' || !request.payload.text.trim()) throw new Error('Введите ответ.')
          step.completedCount++
          step.feedback = request.action === 'add_word' ? 'Слово добавлено в демонстрационный счётчик.' : 'Демонстрационный разбор: здесь появятся проверка ответа и объяснение из учебного режима.'
          break
        case 'ack_feedback':
          if (!step.feedback) throw new Error('Нет разбора для подтверждения.')
          step.feedback = null
          if (step.completedCount >= step.targetCount) {
            if (scenario === 'add_words' && step.mode === 'words' && !step.result) {
              step.result = 'Добавлено 10 слов'
              step.completedCount = 0
              step.targetCount = 2
            } else {
              step.status = 'completed'
              step.result = `${step.result ? `${step.result}; ` : ''}выполнено заданий: ${step.completedCount}`
            }
          }
          break
        default: throw new Error('Неизвестное действие.')
      }
      if (!currentJourneyStep(journey)) journey.status = 'completed'
      journey.revision++
      saved.receipts[request.client_action_id] = { fingerprint, result: copy(journey) }
      save(saved)
      return copy(journey)
    },
  }
}
