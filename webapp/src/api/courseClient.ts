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

export const courseClient = {
  getCourseMap(): Promise<CourseMap> {
    return apiClient.request('/api/learning/course')
  },
}
