import { apiClient } from './client'

export interface StatsDay {
  date: string
  active_seconds: number
  attempt_count: number
  status: 'done' | 'today' | 'empty'
}

export interface SkillStat {
  mode: string
  attempt_count: number
  correct_count: number
  accuracy_percent: number
}

export interface FavoriteDistrict {
  district_code: string
  title: string
  attempt_count: number
  progress_percent: number
}

export interface LinglowStats {
  streak: { current_days: number; best_days: number; today_active: boolean }
  today: { active_seconds: number; attempt_count: number }
  week: StatsDay[]
  month: {
    month: string
    active_minutes: number
    words_learned: number
    texts_read: number
    chat_messages: number
    active_days: number
  }
  skills: SkillStat[]
  skills_period: 'month' | 'all'
  favorite_district?: FavoriteDistrict
  achievements?: AchievementStat[]
  improvements?: ImprovementStat[]
  generated_at: string
}

export interface AchievementStat {
  code: string
  value: number
  unlocked: boolean
}

export interface ImprovementStat {
  kind: string
  mode?: string
  count?: number
  district_code?: string
  title?: string
  accuracy?: number
}

export const statsClient = {
  getStats(params: { courseCode?: string; month?: string } = {}): Promise<LinglowStats> {
    const p = new URLSearchParams()
    if (params.courseCode) p.set('course_code', params.courseCode)
    if (params.month) p.set('month', params.month)
    const qs = p.toString()
    return apiClient.request(`/api/linglow/stats${qs ? '?' + qs : ''}`)
  },
}
