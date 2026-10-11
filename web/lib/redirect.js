// Returns a same-site path for post-login redirects, or '/' for anything else.
// Rejects protocol-relative ("//host") and backslash variants ("/\host"), which browsers treat as external URLs.
export function safeRedirect(from) {
  if (typeof from !== 'string' || !from.startsWith('/')) return '/'
  if (from.startsWith('//') || from.includes('\\')) return '/'
  return from
}
