import type { JourneyStep } from './types'
type Translate = (key: string, values?: Record<string, unknown>) => string
const defaultTitles: Partial<Record<JourneyStep['mode'], string>> = { words: 'Слова', reading: 'Чтение', grammar: 'Грамматика', grammar_srs: 'Тренировка грамматики', verb_forms: 'Формы глаголов' }
export function journeyStepTitle(step: Pick<JourneyStep, 'title'> & { mode?: JourneyStep['mode'] }, t: Translate): string {
  return step.mode && (step.title === defaultTitles[step.mode] || step.title === step.mode) ? t('dailyJourney.modes.' + step.mode) : step.title
}
export function journeyStepGoal(step: JourneyStep, t: Translate): string {
  const goals: Record<string, string> = {
    'Добавь новые слова и повтори знакомые': 'words',
    'Изучить следующий блок и ответить на вопросы': 'grammar',
    'Изучить главу и пройти её тест': 'grammar',
    'Пройти полный тест категории': 'categoryTest', 'Пройти полный тест главы': 'chapterTest',
    'Повторить доступные правила': 'grammar_srs', 'Пройти короткую тренировку форм': 'verb_forms',
  }
  if (goals[step.goal]) return t('dailyJourney.goals.' + goals[step.goal])
  if ((step.mode === 'picture' || step.mode === 'conversation') && step.goal.startsWith('Закончить попытку:')) return t('dailyJourney.goals.chat', { count: step.targetCount })
  return step.goal
}
