import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

const enCourse = {
  id: 1,
  code: 'en_ru',
  title: 'English',
  city_name: 'London',
  target_language: 'en',
  native_language: 'ru',
  ui_locale: 'ru',
  status: 'published',
  is_current: true,
}

const esCourse = {
  ...enCourse,
  id: 2,
  code: 'es_ru',
  title: 'Spanish',
  city_name: 'Madrid',
  target_language: 'es',
  is_current: false,
}

const getCourses = vi.fn()
const selectCourse = vi.fn()
const setGrammarCourse = vi.fn()
const resetLearning = vi.fn()

vi.mock('../api/courseClient', () => ({
  courseClient: { getCourses, selectCourse },
}))

vi.mock('../api/grammarClient', () => ({
  setGrammarCourse,
}))

vi.mock('./useLearningConfig', () => ({
  resetLearning,
}))

const setOnline = (online: boolean) => {
  Object.defineProperty(navigator, 'onLine', {
    value: online,
    configurable: true,
  })
}

const freshUseCourse = async () => {
  vi.resetModules()
  const mod = await import('./useCourse')
  return mod.useCourse()
}

describe('useCourse', () => {
  beforeEach(() => {
    localStorage.clear()
    getCourses.mockReset()
    selectCourse.mockReset()
    setGrammarCourse.mockReset()
    resetLearning.mockReset()
    setOnline(true)
  })

  afterEach(() => {
    localStorage.clear()
  })

  it('loads courses online and caches them', async () => {
    getCourses.mockResolvedValue({ courses: [enCourse, esCourse] })
    const c = await freshUseCourse()
    await c.ensureCourseLoaded()

    expect(c.currentCourseCode.value).toBe('en_ru')
    expect(setGrammarCourse).toHaveBeenCalledWith('en_ru')
    expect(localStorage.getItem('linglow.courseCache.v1')).toContain('en_ru')
  })

  it('hydrates courses from cache while offline', async () => {
    localStorage.setItem('linglow.courseCache.v1', JSON.stringify({
      courses: [enCourse, { ...esCourse, is_current: true }],
      currentCourseCode: 'es_ru',
    }))
    setOnline(false)

    const c = await freshUseCourse()
    await c.ensureCourseLoaded()

    expect(getCourses).not.toHaveBeenCalled()
    expect(c.currentCourseCode.value).toBe('es_ru')
    expect(c.courses.value).toHaveLength(2)
  })

  it('switches course locally while offline', async () => {
    localStorage.setItem('linglow.courseCache.v1', JSON.stringify({
      courses: [enCourse, esCourse],
      currentCourseCode: 'en_ru',
    }))
    setOnline(false)

    const c = await freshUseCourse()
    await c.ensureCourseLoaded()
    await c.selectCourse('es_ru')

    expect(selectCourse).not.toHaveBeenCalled()
    expect(c.currentCourseCode.value).toBe('es_ru')
    expect(resetLearning).toHaveBeenCalled()
    expect(JSON.parse(localStorage.getItem('linglow.courseCache.v1') || '{}').currentCourseCode).toBe('es_ru')
  })

  it('keeps online select backed by API', async () => {
    getCourses.mockResolvedValue({ courses: [enCourse, esCourse] })
    selectCourse.mockResolvedValue({ course: esCourse, user_course: { id: 10, status: 'active' } })

    const c = await freshUseCourse()
    await c.ensureCourseLoaded()
    await c.selectCourse('es_ru')

    expect(selectCourse).toHaveBeenCalledWith('es_ru')
    expect(c.currentCourseCode.value).toBe('es_ru')
    expect(localStorage.getItem('linglow.courseCache.v1')).toContain('es_ru')
  })
})
