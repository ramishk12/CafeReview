import { describe, expect, it } from 'vitest'
import { buildCsp } from '@/lib/csp'

describe('buildCsp', () => {
  it('includes the nonce in script-src and style-src', () => {
    const csp = buildCsp('abc123')
    expect(csp).toContain("script-src 'self' 'nonce-abc123' 'strict-dynamic'")
    expect(csp).toContain("style-src 'self' 'nonce-abc123'")
  })

  it('locks down the dangerous directives', () => {
    const csp = buildCsp('n')
    expect(csp).toContain("object-src 'none'")
    expect(csp).toContain("base-uri 'self'")
    expect(csp).toContain("frame-ancestors 'none'")
    expect(csp).toContain("default-src 'self'")
  })

  it('allows the image sources the app uses: self, blob previews, and data', () => {
    expect(buildCsp('n')).toContain("img-src 'self' blob: data:")
  })

  it('only allows unsafe-eval in development', () => {
    expect(buildCsp('n', { isDev: true })).toContain("'unsafe-eval'")
    expect(buildCsp('n', { isDev: false })).not.toContain("'unsafe-eval'")
  })

  it('never allows unsafe-inline for scripts', () => {
    expect(buildCsp('n')).not.toContain("'unsafe-inline'")
  })
})
