// Server-side reads for Server Components. They call the Go API directly.
// Do not import this file from a Client Component.
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

async function get(path) {
  const res = await fetch(`${BACKEND_URL}/api${path}`, { cache: 'no-store' })
  if (res.status === 404) return null
  if (!res.ok) throw new Error(`API ${path} failed with status ${res.status}`)
  return res.json()
}

export const listCafes = () => get('/cafes')
export const getCafe = (id) => get(`/cafes/${id}`)
export const listReviews = (cafeId) => get(`/cafes/${cafeId}/reviews`)
