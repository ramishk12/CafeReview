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

export class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
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

  const res = await fetch(`/api${path}`, { method, headers, body: payload })
  if (res.status === 204) return null

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    // A rejected token means the session expired; sign the user out everywhere.
    if (res.status === 401 && token && path !== '/auth/login' && path !== '/auth/register') {
      setToken(null)
      window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT))
    }
    throw new ApiError(data.error || `Request failed (${res.status})`, res.status)
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
