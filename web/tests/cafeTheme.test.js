import { describe, expect, it } from 'vitest'
import { avatarColor, cafeGradient, formatDate, initials, ratingLabel } from '@/lib/cafeTheme'

describe('cafeTheme helpers', () => {
  it('returns the same gradient and color for the same name', () => {
    expect(cafeGradient('Bean There')).toBe(cafeGradient('Bean There'))
    expect(avatarColor('Sam Lee')).toBe(avatarColor('Sam Lee'))
  })

  it('builds initials from up to two words', () => {
    expect(initials('Demo User')).toBe('DU')
    expect(initials('  ann  ')).toBe('A')
    expect(initials('Mary Jane Watson')).toBe('MJ')
    expect(initials('')).toBe('?')
  })

  it('formats dates in UTC so server and browser agree', () => {
    expect(formatDate('2026-10-08T23:30:00Z')).toBe('Oct 8, 2026')
  })

  it('labels review and reviewer counts', () => {
    expect(ratingLabel(0)).toBe('No reviews yet')
    expect(ratingLabel(1)).toBe('1 review')
    expect(ratingLabel(1, 1)).toBe('1 review from 1 person')
    expect(ratingLabel(3)).toBe('3 reviews')
    expect(ratingLabel(3, 2)).toBe('3 reviews from 2 people')
  })
})
