import { currentUserScope } from '../api/sessionScope'
import { ref, computed } from 'vue'
import { courseClient, CourseSummary } from '../api/courseClient'
import { setGrammarCourse } from '../api/grammarClient'
import { resetLearning } from './useLearningConfig'
import { emitAppDataEvent, setActiveCourseCodeForInvalidation } from '../api/cacheInvalidation'

const courses = ref<CourseSummary[]>([])
const currentCourse = ref<CourseSummary | null>(null)
let loadPromise: Promise<void> | null = null
const cacheKey = () => `linglow.courseCache.v2:${currentUserScope()}`
let loadedScope = ''

interface CourseCache {
  courses: CourseSummary[]
  currentCourseCode: string
}

const isBrowserOffline = () => typeof navigator !== 'undefined' && navigator.onLine === false

const readCache = (): CourseCache | null => {
  if (typeof localStorage === 'undefined') return null
  try {
    const raw = localStorage.getItem(cacheKey())
    if (!raw) return null
    const parsed = JSON.parse(raw) as CourseCache
    if (!Array.isArray(parsed.courses)) return null
    return parsed
  } catch {
    return null
  }
}

const writeCache = () => {
  if (typeof localStorage === 'undefined' || courses.value.length === 0) return
  localStorage.setItem(cacheKey(), JSON.stringify({
    courses: courses.value,
    currentCourseCode: currentCourse.value?.code || '',
  }))
}

const hydrateFromCache = (): boolean => {
  const cached = readCache()
  if (!cached || cached.courses.length === 0) return false
  courses.value = cached.courses
  currentCourse.value =
    cached.courses.find(c => c.code === cached.currentCourseCode) ||
    cached.courses.find(c => c.is_current) ||
    cached.courses[0] ||
    null
  if (currentCourse.value?.code) {
    setGrammarCourse(currentCourse.value.code)
    setActiveCourseCodeForInvalidation(currentCourse.value.code)
  }
  return true
}

const setCurrentCourse = (course: CourseSummary) => {
  currentCourse.value = course
  courses.value = courses.value.map(c => ({ ...c, is_current: c.code === course.code }))
  setGrammarCourse(course.code)
  setActiveCourseCodeForInvalidation(course.code)
  resetLearning()
  writeCache()
}

async function ensureCourseLoaded(): Promise<void> {
  const scope = currentUserScope()
  if (loadedScope !== scope) { resetCourse(); loadedScope = scope }
  if (loadPromise) {
    await loadPromise
    return
  }
  loadPromise = (async () => {
    if (isBrowserOffline() && hydrateFromCache()) return
    try {
      const data = await courseClient.getCourses()
      if (currentUserScope() !== scope) return
      courses.value = data.courses || []
      currentCourse.value = courses.value.find(c => c.is_current) || courses.value[0] || null
      if (currentCourse.value?.code) {
        setGrammarCourse(currentCourse.value.code)
        setActiveCourseCodeForInvalidation(currentCourse.value.code)
      }
      writeCache()
    } catch {
      if (currentUserScope() !== scope) return
      // Offline or flaky network: keep the course selector usable from last known data.
      hydrateFromCache()
    }
  })()
  await loadPromise
}

export function resetCourse(): void {
  courses.value = []
  currentCourse.value = null
  loadPromise = null
  setGrammarCourse('')
  setActiveCourseCodeForInvalidation('')
  resetLearning()
}

export function useCourse() {
  const currentCourseCode = computed(() => currentCourse.value?.code ?? '')

  async function selectCourse(code: string): Promise<void> {
    if (isBrowserOffline()) {
      const course = courses.value.find(c => c.code === code)
      if (course) setCurrentCourse(course)
      return
    }
    const result = await courseClient.selectCourse(code)
    setCurrentCourse(result.course)
    emitAppDataEvent('course-selected', result.course.code)
    // allow next ensureCourseLoaded to refresh the list
    loadPromise = null
  }

  return {
    courses,
    currentCourse,
    currentCourseCode,
    ensureCourseLoaded,
    selectCourse,
    resetCourse,
  }
}
