import { describe, expect, it, vi } from 'vitest'
import { useAudio } from './useAudio'
const { request } = vi.hoisted(() => ({ request: vi.fn() }))
vi.mock('../api/client', () => ({ apiClient: { request } }))
vi.mock('../api/grammarClient', () => ({ getGrammarCourseCode: () => 'en_ru' }))
describe('pronunciation course scope', () => {
 it('uses explicit journey course and keeps its cache separate from the global course', async () => {
  request.mockImplementation(async (url: string) => ({ available: true, url: url.includes('course_code=es_ru') ? '/spanish.mp3' : '/english.mp3' }))
  const audio=useAudio()
  expect(await audio.getWordPronunciationURL('hotel','es_ru')).toBe('/spanish.mp3')
  expect(await audio.getWordPronunciationURL('hotel')).toBe('/english.mp3')
  expect(await audio.getWordPronunciationURL('hotel','es_ru')).toBe('/spanish.mp3')
  expect(request).toHaveBeenCalledTimes(2)
 })
})
