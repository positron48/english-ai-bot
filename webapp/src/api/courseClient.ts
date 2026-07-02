import { apiClient } from './client'

export interface CourseMapItem {
  id: number
  type: string
  source_kind: string
  source_id: string
  title?: string
  cefr_level?: string
  status: string
}

export interface CourseMapModule {
  id: number
  code: string
  type: string
  title: string
  source_kind?: string
  source_id?: string
  order: number
  status: string
  items: CourseMapItem[]
}

export interface CourseMapLocation {
  id: number
  code: string
  location_type: string
  title: string
  order: number
  status: string
  modules: CourseMapModule[]
}

export interface DistrictBuilding {
  name: string
  type: string
  x: number
  y: number
}

export interface DistrictMetadata {
  image?: string
  desc_i18n?: Record<string, string>
  lumi_tips?: { low?: string; mid?: string; high?: string }
  buildings?: DistrictBuilding[]
  polygon?: Array<[number, number]>
}

export interface CourseMapDistrict {
  id: number
  code: string
  level_code: string
  title: string
  order: number
  status: string
  description?: string
  metadata?: DistrictMetadata
  locations: CourseMapLocation[]
}

export interface DistrictExtras {
  discovery: {
    kind: string
    text_id: string
    title: string
    category_id: string
  } | null
  tasks: Array<{ kind: string; target: number; done: number }>
}

export interface CourseMap {
  course: {
    id: number
    code: string
    slug: string
    title: string
    target_language: string
    native_language: string
    ui_locale: string
    status: string
    city_name: string
  }
  user_course?: {
    id: number
    status: string
  }
  districts: CourseMapDistrict[]
  totals: {
    districts: number
    locations: number
    modules: number
    items: number
    by_type: Record<string, number>
  }
}

export interface CourseSummary {
  id: number
  code: string
  title: string
  city_name: string
  target_language: string
  native_language: string
  ui_locale: string
  status: string
  is_current: boolean
  user_course_id?: number
  user_status?: string
}

export interface CurrentCourse {
  course: CourseSummary
  user_course: {
    id: number
    status: string
  }
}

export interface DailyRouteItem {
  learning_item_id: number
  srs_item_id?: number
  type: string
  source_kind: string
  source_id: string
  title?: string
  cefr_level?: string
  mode: string
  state?: string
  due_at?: string
  district_code?: string
  district_title?: string
  location_code?: string
  location_type?: string
  location_title?: string
  module_code?: string
  module_title?: string
}

export interface DailyRoute {
  course: CourseMap['course']
  user_course: {
    id: number
    status: string
  }
  summary: {
    due_review_count: number
    new_item_count: number
    by_type: Record<string, number>
    read_source?: string
  }
  review: DailyRouteItem[]
  new_items: DailyRouteItem[]
  today?: {
    words_due: number
    words_done: number
    reading_done: boolean
    reading_suggestion?: { text_id: string; title: string }
    chat_done: boolean
  }
  generated_at: string
}

export interface ReviewQueue {
  course: CourseMap['course']
  user_course: {
    id: number
    status: string
  }
  summary: {
    due_count: number
    learning_count: number
    review_count: number
    relearning_count: number
    upcoming_count: number
    by_type: Record<string, number>
    read_source?: string
  }
  items: DailyRouteItem[]
  generated_at: string
}

export interface CourseProgress {
  course: CourseMap['course']
  user_course: {
    id: number
    status: string
  }
  summary: {
    total_items: number
    attempted_items: number
    mastered_items: number
    due_review_count: number
    attempt_count: number
    correct_count: number
    progress_percent: number
    accuracy_percent: number
  }
  by_type: Array<{
    type: string
    total_items: number
    attempted_items: number
    mastered_items: number
    progress_percent: number
  }>
  by_district?: CourseProgressDistrict[]
  by_location?: CourseProgressLocation[]
  generated_at: string
}

export interface CourseProgressDistrict {
  district_id: number
  district_code: string
  level_code: string
  title: string
  total_items: number
  attempted_items: number
  mastered_items: number
  due_review_count: number
  attempt_count: number
  correct_count: number
  foundation: number
  confidence: number
  stability: number
  weakness: number
  progress_percent: number
}

export interface CourseProgressLocation {
  location_id: number
  location_code: string
  location_type: string
  title: string
  district_id: number
  district_code: string
  total_items: number
  attempted_items: number
  mastered_items: number
  due_review_count: number
  attempt_count: number
  correct_count: number
  foundation: number
  confidence: number
  stability: number
  weakness: number
  progress_percent: number
}

export interface ExerciseAttemptRequest {
  course_code?: string
  learning_item_id?: number
  srs_item_id?: number
  mode: string
  client_attempt_id?: string
  is_correct?: boolean
  score?: number
  quality?: number
  prompt?: Record<string, unknown>
  answer?: Record<string, unknown>
  result?: Record<string, unknown>
  answered_at?: string
}

export interface ExerciseAttemptResult {
  id: number
  user_course_id: number
  learning_item_id?: number
  srs_item_id?: number
  client_attempt_id?: string
  duplicate: boolean
  event_id?: number
  course: CourseMap['course']
  user_course: {
    id: number
    status: string
  }
}

export interface SRSShadowReport {
  course: CourseMap['course']
  user_course: {
    id: number
    status: string
  }
  due: {
    legacy_due_count: number
    linglow_due_count: number
    overlap_count: number
    legacy_only_count: number
    linglow_only_count: number
  }
  review_queue: {
    legacy_due_count: number
    canonical_due_count: number
    overlap_count: number
    legacy_only_count: number
    canonical_only_count: number
    ready_for_canonical_read: boolean
    by_type: Record<string, number>
  }
  mastery: {
    compared_count: number
    average_legacy: number
    average_linglow: number
    average_difference: number
    max_difference: number
  }
  generated_at: string
}

export interface SRSReadinessAggregateReport {
  course: CourseMap['course']
  user_courses_total: number
  ready_count: number
  not_ready_count: number
  ready_for_canonical_read: boolean
  can_enable_srs_read?: boolean
  srs_read_enabled?: boolean
  srs_write_enabled?: boolean
  legacy_due_total: number
  canonical_due_total: number
  legacy_only_total: number
  canonical_only_total: number
  overlap_total: number
  by_type: Record<string, number>
  not_ready_users: Array<{
    user_id: number
    user_course_id: number
    legacy_due_count: number
    canonical_due_count: number
    legacy_only_count: number
    canonical_only_count: number
  }>
  generated_at: string
}

export interface LinglowWordItem {
  learning_item_id: number
  word_card_id: string
  lemma: string
  display_word: string
  translation: string
  cefr_level: string
  state: string
  due_at?: string
  reps: number
  added_at?: string
  last_review_at?: string
  due_count: number
  total_cards: number
  total_reps: number
  mastery_level: string
}

export interface LinglowWordList {
  course: CourseMap['course']
  user_course: { id: number; status: string }
  words: LinglowWordItem[]
  total: number
  limit: number
  offset: number
  generated_at: string
}

export interface LinglowHistory {
  course: CourseMap['course']
  user_course: { id: number; status: string }
  accuracy_percent: number
  total_attempts: number
  correct_attempts: number
  weekly_stats: Array<{ day: string; cards_completed: number; cards_correct: number }>
  words_added_stats: Array<{ day: string; words_added: number }>
  by_mode: Array<{ mode: string; attempt_count: number; correct_count: number }>
  generated_at: string
}

export interface ConversationTask {
  code: string
  title: string
  required: boolean
  completed: boolean
}

export interface ConversationScenarioSummary {
  code: string
  title: string
  npc_name: string
  npc_code: string
  place_type: string
  cefr_level: string
  is_quest: boolean
  prerequisite_code: string
  image_url: string
  npc_image_url: string
  cooldown_until: string | null
  locked: boolean
  tasks: ConversationTask[]
  session_status: string
  quest_passed: boolean
  all_tasks_done: boolean
}

export interface PictureQuestSummary {
  code: string
  title: string
  cefr_level: string
  image_url: string
  tasks: ConversationTask[]
  session_status: string
  quest_passed: boolean
  all_tasks_done: boolean
}

export interface PictureQuestSessionState {
  session_id: number
  quest_code: string
  title: string
  cefr_level: string
  image_url: string
  opening_line: string
  messages: ConversationMessage[]
  tasks: ConversationTask[]
  turn_count: number
  max_turns: number
  status: string
  quest_passed: boolean
}

export interface ConversationCorrection {
  original: string
  corrected: string
  explanation: string
}

export interface ConversationMessage {
  role: string
  content: string
  corrections?: ConversationCorrection[]
}

export interface ConversationSessionState {
  session_id: number
  scenario_code: string
  title: string
  npc_name: string
  place_type: string
  cefr_level: string
  is_quest: boolean
  scene_setup: string
  image_url: string
  opening_line: string
  messages: ConversationMessage[]
  tasks: ConversationTask[]
  turn_count: number
  max_turns: number
  status: string
  quest_passed: boolean
}

export interface ConversationTurnResult {
  reply: string
  corrections?: ConversationCorrection[]
  tasks: ConversationTask[]
  turn_count: number
  max_turns: number
  quest_passed: boolean
  budget_exhausted: boolean
  status: string
}

export const courseClient = {
  getCourses(): Promise<{ courses: CourseSummary[] }> {
    return apiClient.request('/api/courses')
  },
  selectCourse(courseCode: string): Promise<CurrentCourse> {
    return apiClient.request('/api/user/courses/select', {
      method: 'POST',
      body: JSON.stringify({ course_code: courseCode }),
    })
  },
  getCourseMap(courseCode?: string): Promise<CourseMap> {
    return apiClient.request(withCourseCode('/api/linglow/city', courseCode))
  },
  getDailyRoute(limit = 8, courseCode?: string): Promise<DailyRoute> {
    return apiClient.request(withCourseCode(`/api/linglow/daily-route?limit=${encodeURIComponent(String(limit))}`, courseCode))
  },
  getReviewQueue(limit = 20, courseCode?: string): Promise<ReviewQueue> {
    return apiClient.request(withCourseCode(`/api/linglow/review?limit=${encodeURIComponent(String(limit))}`, courseCode))
  },
  getProgress(courseCode?: string): Promise<CourseProgress> {
    return apiClient.request(withCourseCode('/api/linglow/progress', courseCode))
  },
  getSRSShadowReport(courseCode?: string): Promise<SRSShadowReport> {
    return apiClient.request(withCourseCode('/api/linglow/srs-shadow', courseCode))
  },
  getSRSReadinessAggregate(courseCode?: string, limit = 20): Promise<SRSReadinessAggregateReport> {
    return apiClient.request(withCourseCode(`/api/admin/linglow/srs-readiness?limit=${encodeURIComponent(String(limit))}`, courseCode))
  },
  recordExerciseAttempt(payload: ExerciseAttemptRequest): Promise<ExerciseAttemptResult> {
    return apiClient.request('/api/linglow/exercise-attempts', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  getWordList(params: { courseCode?: string; q?: string; status?: string; sort?: string; limit?: number; offset?: number } = {}): Promise<LinglowWordList> {
    const p = new URLSearchParams()
    if (params.courseCode) p.set('course_code', params.courseCode)
    if (params.q) p.set('q', params.q)
    if (params.status) p.set('status', params.status)
    if (params.sort) p.set('sort', params.sort)
    if (params.limit != null) p.set('limit', String(params.limit))
    if (params.offset != null) p.set('offset', String(params.offset))
    const qs = p.toString()
    return apiClient.request(`/api/linglow/words${qs ? '?' + qs : ''}`)
  },
  getWordLevelProgress(courseCode?: string): Promise<{ levels: Record<string, { total: number; mastered: number }> }> {
    const p = new URLSearchParams()
    if (courseCode) p.set('course_code', courseCode)
    const qs = p.toString()
    return apiClient.request(`/api/linglow/word-progress${qs ? '?' + qs : ''}`)
  },
  getDistrictExtras(districtCode: string, courseCode?: string): Promise<DistrictExtras> {
    const p = new URLSearchParams({ district_code: districtCode })
    if (courseCode) p.set('course_code', courseCode)
    return apiClient.request(`/api/linglow/district-extras?${p.toString()}`)
  },
  getHistory(params: { courseCode?: string; days?: number } = {}): Promise<LinglowHistory> {
    const p = new URLSearchParams()
    if (params.courseCode) p.set('course_code', params.courseCode)
    if (params.days != null) p.set('days', String(params.days))
    const qs = p.toString()
    return apiClient.request(`/api/linglow/history${qs ? '?' + qs : ''}`)
  },
  listConversationScenarios(districtCode: string, courseCode?: string): Promise<{ scenarios: ConversationScenarioSummary[] }> {
    const p = new URLSearchParams({ district_code: districtCode })
    if (courseCode) p.set('course_code', courseCode)
    return apiClient.request(`/api/linglow/conversation/scenarios?${p.toString()}`)
  },
  startConversationSession(scenarioCode: string, courseCode?: string): Promise<ConversationSessionState> {
    return apiClient.request('/api/linglow/conversation/sessions', {
      method: 'POST',
      body: JSON.stringify({ scenario_code: scenarioCode, course_code: courseCode }),
    })
  },
  getConversationSession(sessionId: number): Promise<ConversationSessionState> {
    return apiClient.request(`/api/linglow/conversation/sessions/${sessionId}`)
  },
  postConversationMessage(sessionId: number, text: string): Promise<ConversationTurnResult> {
    return apiClient.request(`/api/linglow/conversation/sessions/${sessionId}/message`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
  },
  endConversationSession(sessionId: number, status?: string): Promise<{ status: string }> {
    return apiClient.request(`/api/linglow/conversation/sessions/${sessionId}/end`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    })
  },
  resetConversationSession(sessionId: number): Promise<ConversationSessionState> {
    return apiClient.request(`/api/linglow/conversation/sessions/${sessionId}/reset`, {
      method: 'POST',
    })
  },

  listPictureQuests(districtCode: string, courseCode?: string): Promise<{ quests: PictureQuestSummary[] }> {
    const p = new URLSearchParams({ district_code: districtCode })
    if (courseCode) p.set('course_code', courseCode)
    return apiClient.request(`/api/linglow/picture-quests?${p.toString()}`)
  },
  startPictureQuestSession(questCode: string, courseCode?: string): Promise<PictureQuestSessionState> {
    return apiClient.request('/api/linglow/picture-quests/sessions', {
      method: 'POST',
      body: JSON.stringify({ quest_code: questCode, course_code: courseCode }),
    })
  },
  getPictureQuestSession(sessionId: number): Promise<PictureQuestSessionState> {
    return apiClient.request(`/api/linglow/picture-quests/sessions/${sessionId}`)
  },
  postPictureQuestMessage(sessionId: number, text: string): Promise<ConversationTurnResult> {
    return apiClient.request(`/api/linglow/picture-quests/sessions/${sessionId}/message`, {
      method: 'POST',
      body: JSON.stringify({ text }),
    })
  },
  endPictureQuestSession(sessionId: number, status?: string): Promise<{ status: string }> {
    return apiClient.request(`/api/linglow/picture-quests/sessions/${sessionId}/end`, {
      method: 'POST',
      body: JSON.stringify({ status }),
    })
  },
  resetPictureQuestSession(sessionId: number): Promise<PictureQuestSessionState> {
    return apiClient.request(`/api/linglow/picture-quests/sessions/${sessionId}/reset`, {
      method: 'POST',
    })
  },

  uploadAdminMedia(file: File, type: 'npc' | 'quest' | 'picture'): Promise<{ url: string }> {
    const fd = new FormData()
    fd.append('file', file)
    return apiClient.request(`/api/admin/media/upload?type=${encodeURIComponent(type)}`, {
      method: 'POST',
      body: fd,
    })
  },

  setAdminNpcImage(npcCode: string, imageUrl: string, courseCode: string): Promise<{ success: boolean }> {
    return apiClient.request(
      `/api/admin/conversations/npcs/${encodeURIComponent(npcCode)}/image?course_code=${encodeURIComponent(courseCode)}`,
      { method: 'PUT', body: JSON.stringify({ image_url: imageUrl }) },
    )
  },
}

function withCourseCode(url: string, courseCode?: string): string {
  if (!courseCode) return url
  const sep = url.includes('?') ? '&' : '?'
  return `${url}${sep}course_code=${encodeURIComponent(courseCode)}`
}
