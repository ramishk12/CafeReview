// Browser-side API client. Requests go to /api on the same origin; next.config.mjs proxies them.
const TOKEN_KEY = 'cafe_review_token'
export const AUTH_EXPIRED_EVENT = 'auth:expired'

export function getToken() {
  if (typeof window === 'undefined') return null
  return window.localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  if (token) window.localStorage.setItem(TOKEN_KEY, token)
  else window.localStorage.removeItem(TOKEN_KEY)
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
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  let payload
  if (form) {
    payload = form // the browser sets the multipart boundary itself
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }

  let res
  try {
    res = await fetch(`/api${path}`, { method, headers, body: payload })
  } catch {
    throw new ApiError(NETWORK_ERROR_MESSAGE, 0)
  }
  if (res.status === 204) return null

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    // A rejected token means the session expired; sign the user out everywhere.
    if (res.status === 401 && token && path !== '/auth/login' && path !== '/auth/register') {
      setToken(null)
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
export const me = () => request('/auth/me')

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
