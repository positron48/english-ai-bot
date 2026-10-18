export type JourneyMode = 'words' | 'reading' | 'grammar' | 'grammar_srs' | 'verb_forms' | 'picture' | 'conversation'
export type StepStatus = 'pending' | 'active' | 'completed' | 'skipped' | 'unavailable'
export interface JourneyStep {
  id: string
  mode: JourneyMode
  title: string
  goal: string
  required: boolean
  status: StepStatus
  completedCount: number
  targetCount: number
  contentRef?: Record<string, unknown>
  contentVersion?: string
  launch?: { routeName: string; params?: Record<string, string> }
  feedback: string | null
  result: string | null
  resultDetails?: Record<string, unknown>
}
export interface DailyJourney {
  id: string
  courseCode: string
  date: string
  revision: number
  status: 'active' | 'completed' | 'unavailable' | 'archived'
  steps: JourneyStep[]
}
export type JourneyActionType = 'start' | 'add_word' | 'answer' | 'message' | 'ack_feedback' | 'skip'
export interface JourneyAction {
  client_action_id: string
  expected_revision: number
  action: JourneyActionType
  payload: Record<string, unknown>
}
export interface JourneyTransport {
  get(): Promise<DailyJourney | null>
  start(): Promise<DailyJourney>
  act(journeyId: string, stepId: string, action: JourneyAction): Promise<DailyJourney>
}
export function currentJourneyStep(journey: DailyJourney | null): JourneyStep | undefined {
  return journey?.steps.find(step => step.status === 'pending' || step.status === 'active')
}
export function journeySkipAvailable(journey: DailyJourney | null): boolean {
  return journey?.status === 'active' && !journey.steps.some(step => step.status === 'skipped')
}

export type JourneyCardData = Pick<DailyJourney, 'id' | 'courseCode' | 'date' | 'status' | 'revision'> & { steps: (Pick<JourneyStep, 'id' | 'title' | 'status'> & { mode?: JourneyMode })[] }
