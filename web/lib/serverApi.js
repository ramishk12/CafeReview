// Server-side reads for Server Components. They call the Go API directly.
// Do not import this file from a Client Component.
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

async function get(path) {
  const res = await fetch(`${BACKEND_URL}/api${path}`, { cache: 'no-store' })
  if (res.status === 404) return null
  if (!res.ok) throw new Error(`API ${path} failed with status ${res.status}`)
  return res.json()
}

// listCafes fetches one page and the total count from X-Total-Count.
export async function listCafes({ q, sort, tags, page = 1, pageSize = 24 } = {}) {
  const params = new URLSearchParams()
  if (q) params.set('q', q)
  if (sort) params.set('sort', sort)
  if (tags) params.set('tags', tags)
  params.set('limit', pageSize)
  params.set('offset', (page - 1) * pageSize)

  const res = await fetch(`${BACKEND_URL}/api/cafes?${params.toString()}`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`API /cafes failed with status ${res.status}`)
  const total = Number(res.headers.get('X-Total-Count') || 0)
  const cafes = await res.json()
  return { cafes, total }
}
export const listTags = () => get('/tags?scope=cafe')
export const getUser = (id) => get(`/users/${id}`)

export async function listUserReviews(id, { page = 1, pageSize = 20 } = {}) {
  const params = new URLSearchParams()
  params.set('limit', pageSize)
  params.set('offset', (page - 1) * pageSize)

  const res = await fetch(`${BACKEND_URL}/api/users/${id}/reviews?${params.toString()}`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`API /users/${id}/reviews failed with status ${res.status}`)
  const total = Number(res.headers.get('X-Total-Count') || 0)
  const reviews = await res.json()
  return { reviews, total }
}
export const getCafe = (id) => get(`/cafes/${id}`)

// listReviews fetches one page of a cafe's reviews and the total count from
// X-Total-Count. sort and the filters map to the query parameters of
// GET /cafes/:id/reviews. Sorting and filtering happen on the server so the
// page does not have to load every review.
export async function listReviews(cafeId, { sort, rating, photos, tags, userId, page = 1, pageSize = 10 } = {}) {
  const params = new URLSearchParams()
  if (sort) params.set('sort', sort)
  if (rating) params.set('rating', rating)
  if (photos) params.set('photos', '1')
  if (tags) params.set('tags', tags)
  if (userId) params.set('user_id', userId)
  params.set('limit', pageSize)
  params.set('offset', (page - 1) * pageSize)

  const res = await fetch(`${BACKEND_URL}/api/cafes/${cafeId}/reviews?${params.toString()}`, { cache: 'no-store' })
  if (!res.ok) throw new Error(`API /cafes/${cafeId}/reviews failed with status ${res.status}`)
  const total = Number(res.headers.get('X-Total-Count') || 0)
  const reviews = await res.json()
  return { reviews, total }
}
