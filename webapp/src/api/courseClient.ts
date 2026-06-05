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
  mastery: {
    compared_count: number
    average_legacy: number
    average_linglow: number
    average_difference: number
    max_difference: number
  }
  generated_at: string
}

export const courseClient = {
  getCourseMap(): Promise<CourseMap> {
    return apiClient.request('/api/linglow/city')
  },
  getDailyRoute(limit = 8): Promise<DailyRoute> {
    return apiClient.request(`/api/linglow/daily-route?limit=${encodeURIComponent(String(limit))}`)
  },
  getReviewQueue(limit = 20): Promise<ReviewQueue> {
    return apiClient.request(`/api/linglow/review?limit=${encodeURIComponent(String(limit))}`)
  },
  getProgress(): Promise<CourseProgress> {
    return apiClient.request('/api/linglow/progress')
  },
  getSRSShadowReport(): Promise<SRSShadowReport> {
    return apiClient.request('/api/linglow/srs-shadow')
  },
  recordExerciseAttempt(payload: ExerciseAttemptRequest): Promise<ExerciseAttemptResult> {
    return apiClient.request('/api/linglow/exercise-attempts', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
