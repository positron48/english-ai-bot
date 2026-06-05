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

export interface CourseMapDistrict {
  id: number
  code: string
  level_code: string
  title: string
  order: number
  status: string
  locations: CourseMapLocation[]
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
  generated_at: string
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
}

function withCourseCode(url: string, courseCode?: string): string {
  if (!courseCode) return url
  const sep = url.includes('?') ? '&' : '?'
  return `${url}${sep}course_code=${encodeURIComponent(courseCode)}`
}
