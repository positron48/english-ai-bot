import { describe, expect, it } from 'vitest'
import { compareVersions } from './version'

describe('compareVersions', () => {
  it('orders numeric segments, not lexicographically', () => {
    expect(compareVersions('0.12.5', '0.12.49')).toBe(-1)
    expect(compareVersions('0.12.49', '0.12.5')).toBe(1)
  })

  it('treats equal versions as 0', () => {
    expect(compareVersions('1.2.3', '1.2.3')).toBe(0)
  })

  it('pads missing trailing segments with 0', () => {
    expect(compareVersions('0.12', '0.12.0')).toBe(0)
    expect(compareVersions('0.12', '0.12.1')).toBe(-1)
  })

  it('strips a leading v and trims whitespace', () => {
    expect(compareVersions('v1.0.0', ' 1.0.0 ')).toBe(0)
  })

  it('handles major version bumps', () => {
    expect(compareVersions('2.0.0', '1.99.99')).toBe(1)
  })
})
