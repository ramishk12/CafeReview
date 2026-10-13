// Browser-side API client. Requests go to /api on the same origin; next.config.mjs proxies them.
//
// Authentication uses a server-side session sent as an HttpOnly cookie, so the
// session token is never readable from JavaScript. A second, readable cookie
// (cafe_csrf) carries a CSRF token that is echoed in the X-CSRF-Token header on
// state-changing requests. The server compares the two.
export const AUTH_EXPIRED_EVENT = 'auth:expired'

const CSRF_COOKIE = 'cafe_csrf'
const SAFE_METHODS = new Set(['GET', 'HEAD', 'OPTIONS'])

// getCookie reads a non-HttpOnly cookie by name.
export function getCookie(name) {
  if (typeof document === 'undefined') return ''
  const match = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'))
  return match ? decodeURIComponent(match[1]) : ''
}

// ApiError.message is always safe to show to a user. requestId is set for server errors
// and matches the ID in the server logs, so support can find the failure.
export class ApiError extends Error {
  constructor(message, status, requestId = '') {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.requestId = requestId
  }
}

export const NETWORK_ERROR_MESSAGE = "We couldn't reach the server. Check your connection and try again."

// Plain-language text for an HTTP status when the server gave no usable message.
export function fallbackMessage(status) {
  if (status === 502 || status === 503 || status === 504) {
    return 'The service is temporarily unavailable. Please try again in a moment.'
  }
  if (status >= 500) return 'Something went wrong on our end. Please try again.'
  if (status === 404) return "We couldn't find what you were looking for."
  if (status === 403) return "You don't have permission to do that."
  return 'We could not complete that request. Please check the details and try again.'
}

async function request(path, { method = 'GET', body, form } = {}) {
  const headers = {}
  if (!SAFE_METHODS.has(method)) {
    const csrf = getCookie(CSRF_COOKIE)
    if (csrf) headers['X-CSRF-Token'] = csrf
  }

  let payload
  if (form) {
    payload = form // the browser sets the multipart boundary itself
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }

  let res
  try {
    res = await fetch(`/api${path}`, { method, headers, body: payload, credentials: 'same-origin' })
  } catch {
    throw new ApiError(NETWORK_ERROR_MESSAGE, 0)
  }
  if (res.status === 204) return null

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    // A rejected session on a non-auth request means it expired; let the app sign out.
    if (res.status === 401 && path !== '/auth/login' && path !== '/auth/register') {
      window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT))
    }

    if (res.status >= 500) {
      // Server messages for 5xx are not shown; the reference lets support find the logs.
      const requestId = data.request_id || res.headers.get('X-Request-ID') || ''
      const message = requestId ? `${fallbackMessage(res.status)} (Reference: ${requestId})` : fallbackMessage(res.status)
      throw new ApiError(message, res.status, requestId)
    }
    // 4xx messages come from validation and are written to be read by users.
    throw new ApiError(data.error || fallbackMessage(res.status), res.status)
  }
  return data
}

// Auth
export const register = (email, password, displayName) =>
  request('/auth/register', { method: 'POST', body: { email, password, display_name: displayName } })
export const login = (email, password) => request('/auth/login', { method: 'POST', body: { email, password } })
export const logout = () => request('/auth/logout', { method: 'POST' })
export const me = () => request('/auth/me')
export const verifyEmail = (token) => request(`/auth/verify?token=${encodeURIComponent(token)}`)
export const resendVerification = (email) => request('/auth/verify/resend', { method: 'POST', body: { email } })

// Cafes (reads are also fetched on the server; see lib/serverApi.js)
export const createCafe = (cafe) => request('/cafes', { method: 'POST', body: cafe })

// Reviews
export const createReview = (cafeId, review) => request(`/cafes/${cafeId}/reviews`, { method: 'POST', body: review })
export const updateReview = (reviewId, review) => request(`/reviews/${reviewId}`, { method: 'PUT', body: review })
export const deleteReview = (reviewId) => request(`/reviews/${reviewId}`, { method: 'DELETE' })

export function uploadReviewImage(reviewId, file) {
  const form = new FormData()
  form.append('image', file)
  return request(`/reviews/${reviewId}/images`, { method: 'POST', form })
}
