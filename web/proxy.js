import { NextResponse } from 'next/server'
import { buildCsp } from '@/lib/csp'

// Runs on page requests. Generates a per-request nonce, hands it to the renderer
// through the x-nonce header (Next adds it to its own scripts), and sends the
// Content-Security-Policy to the browser.
//
// The policy is report-only for now: the browser logs violations in the console
// but blocks nothing. A follow-up switches the header to enforcing when the app
// goes live, after the reports are clean. See "Enforce the Content-Security-Policy".
export function proxy(request) {
  const nonce = Buffer.from(crypto.randomUUID()).toString('base64')
  const csp = buildCsp(nonce, { isDev: process.env.NODE_ENV === 'development' })

  const headers = new Headers(request.headers)
  headers.set('x-nonce', nonce)

  const response = NextResponse.next({ request: { headers } })
  response.headers.set('Content-Security-Policy-Report-Only', csp)
  response.headers.set('X-Content-Type-Options', 'nosniff')
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin')
  response.headers.set('X-Frame-Options', 'DENY')
  return response
}

export const config = {
  // Pages only. /api and /uploads are rewritten to the Go API, which sets its own
  // headers, and static assets do not need a per-request nonce.
  matcher: [
    {
      source: '/((?!api|uploads|_next/static|_next/image|favicon.ico).*)',
      missing: [
        { type: 'header', key: 'next-router-prefetch' },
        { type: 'header', key: 'purpose', value: 'prefetch' },
      ],
    },
  ],
}
