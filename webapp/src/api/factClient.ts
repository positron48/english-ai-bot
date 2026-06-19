import { apiClient } from './client'

export interface LumiFactDTO {
  id: number
  body: string
  context: string
}

export const factClient = {
  // null when there are no facts for the scope (HTTP 204) or on any error
  async getDailyFact(context: string, courseCode?: string, locale?: string): Promise<LumiFactDTO | null> {
    const p = new URLSearchParams({ context })
    if (courseCode) p.set('course_code', courseCode)
    if (locale) p.set('locale', locale)
    try {
      const res = await apiClient.request<LumiFactDTO | null>(`/api/linglow/lumi-fact?${p.toString()}`)
      return res && (res as LumiFactDTO).body ? (res as LumiFactDTO) : null
    } catch {
      return null
    }
  },
}
