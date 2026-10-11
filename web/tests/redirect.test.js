import { describe, expect, it } from 'vitest'
import { safeRedirect } from '@/lib/redirect'

describe('safeRedirect', () => {
  it.each([
    ['/cafes/3', '/cafes/3'],
    ['/', '/'],
    ['/cafes/new?x=1', '/cafes/new?x=1'],
  ])('keeps the same-site path %s', (input, expected) => {
    expect(safeRedirect(input)).toBe(expected)
  })

  it.each([
    ['https://evil.example', 'absolute URL'],
    ['//evil.example', 'protocol-relative URL'],
    ['/\\evil.example', 'backslash variant browsers treat as //'],
    ['javascript:alert(1)', 'javascript scheme'],
    ['cafes/3', 'relative path without a leading slash'],
    ['', 'empty string'],
    [null, 'missing value'],
    [undefined, 'undefined'],
  ])('falls back to / for %s (%s)', (input) => {
    expect(safeRedirect(input)).toBe('/')
  })
})
