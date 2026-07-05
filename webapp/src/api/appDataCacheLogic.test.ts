import { describe, expect, it } from 'vitest'
import {
  buildScreenStorageKey,
  computeStaleAt,
  entryNeedsRefresh,
  hashTokenScope,
  screensForTags,
  tagsAffectScreen,
  type CachedScreenPayload,
} from './appDataCacheLogic'

describe('appDataCacheLogic', () => {
  it('builds stable screen storage keys', () => {
    expect(buildScreenStorageKey('user:1', 'en_ru', 'dashboard', 'ru'))
      .toBe('user:1|en_ru|ru|dashboard')
  })

  it('computes staleAt from screen TTL', () => {
    const fetchedAt = '2026-07-05T12:00:00.000Z'
    const staleAt = computeStaleAt(fetchedAt, 'dashboard', Date.parse(fetchedAt))
    expect(staleAt).toBe('2026-07-05T12:05:00.000Z')
  })

  it('marks entry stale after staleAt', () => {
    const entry: CachedScreenPayload = {
      key: 'city',
      userScope: 'user:1',
      courseCode: 'en_ru',
      locale: 'ru',
      appVersion: 'x',
      dataVersion: 1,
      payload: {},
      fetchedAt: '2026-07-05T12:00:00.000Z',
      staleAt: '2026-07-05T12:15:00.000Z',
      dirtyTags: [],
      pendingLocalMutations: 0,
    }
    expect(entryNeedsRefresh(entry, Date.parse('2026-07-05T12:14:59.000Z'))).toBe(false)
    expect(entryNeedsRefresh(entry, Date.parse('2026-07-05T12:15:01.000Z'))).toBe(true)
  })

  it('marks dirty entries as needing refresh before staleAt', () => {
    const entry: CachedScreenPayload = {
      key: 'dashboard',
      userScope: 'user:1',
      courseCode: 'en_ru',
      locale: 'ru',
      appVersion: 'x',
      dataVersion: 1,
      payload: {},
      fetchedAt: '2026-07-05T12:00:00.000Z',
      staleAt: '2026-07-05T12:05:00.000Z',
      dirtyTags: ['words'],
      pendingLocalMutations: 1,
    }
    expect(entryNeedsRefresh(entry, Date.parse('2026-07-05T12:01:00.000Z'))).toBe(true)
  })

  it('maps event tags to affected screens', () => {
    expect(screensForTags(['words', 'srs'])).toContain('dashboard')
    expect(screensForTags(['words', 'srs'])).toContain('daily-route')
    expect(tagsAffectScreen('review', ['srs'])).toBe(true)
    expect(tagsAffectScreen('learning', ['srs'])).toBe(false)
    expect(screensForTags(['course']).length).toBe(6)
  })

  it('hashes token scopes deterministically', () => {
    expect(hashTokenScope('abc')).toBe(hashTokenScope('abc'))
    expect(hashTokenScope('abc')).not.toBe(hashTokenScope('abcd'))
  })
})
