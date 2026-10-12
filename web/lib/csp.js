// Builds the Content-Security-Policy for pages.
//
// The policy is sent report-only for now (see web/proxy.js). Keeping it here, as
// one pure function, makes it easy to test and to compare with the follow-up that
// will enforce it at launch.
//
// 'strict-dynamic' lets scripts loaded by a trusted (nonce-carrying) script run,
// which is how Next.js loads its own chunks. 'unsafe-eval' is only needed by React
// in development; production does not use eval.
export function buildCsp(nonce, { isDev = false } = {}) {
  const scriptExtra = isDev ? " 'unsafe-eval'" : ''
  return [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'${scriptExtra}`,
    `style-src 'self' 'nonce-${nonce}'`,
    "img-src 'self' blob: data:",
    "font-src 'self'",
    "connect-src 'self'",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join('; ')
}
