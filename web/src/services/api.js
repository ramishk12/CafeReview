const API_BASE = '/api'
const TOKEN_KEY = 'cafe_review_token'

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

async function request(path, { method = 'GET', body, form } = {}) {
  const headers = {}
  const token = getToken()
  if (token) headers.Authorization = `Bearer ${token}`

  let payload
  if (form) {
    // Let the browser set the multipart Content-Type with boundary.
    payload = form
  } else if (body !== undefined) {
    headers['Content-Type'] = 'application/json'
    payload = JSON.stringify(body)
  }

  const res = await fetch(`${API_BASE}${path}`, { method, headers, body: payload })
  if (res.status === 204) return null

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    const err = new Error(data.error || `Request failed (${res.status})`)
    err.status = res.status
    throw err
  }
  return data
}

// Auth
export const register = (email, password, displayName) =>
  request('/auth/register', { method: 'POST', body: { email, password, display_name: displayName } })

export const login = (email, password) =>
  request('/auth/login', { method: 'POST', body: { email, password } })

export const me = () => request('/auth/me')

// Cafes
export const listCafes = () => request('/cafes')
export const getCafe = (id) => request(`/cafes/${id}`)
export const createCafe = (cafe) => request('/cafes', { method: 'POST', body: cafe })

// Reviews
export const listReviews = (cafeId) => request(`/cafes/${cafeId}/reviews`)
export const createReview = (cafeId, review) =>
  request(`/cafes/${cafeId}/reviews`, { method: 'POST', body: review })
export const updateReview = (reviewId, review) =>
  request(`/reviews/${reviewId}`, { method: 'PUT', body: review })
export const deleteReview = (reviewId) => request(`/reviews/${reviewId}`, { method: 'DELETE' })

export function uploadReviewImage(reviewId, file) {
  const form = new FormData()
  form.append('image', file)
  return request(`/reviews/${reviewId}/images`, { method: 'POST', form })
}
