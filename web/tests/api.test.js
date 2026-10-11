import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as api from '@/lib/api'

const { AUTH_EXPIRED_EVENT, ApiError } = api

function mockFetch(status, body) {
  const response = body === undefined
    ? new Response(null, { status })
    : new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
  const fetchMock = vi.fn().mockResolvedValue(response)
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

describe('api client', () => {
  beforeEach(() => {
    window.localStorage.clear()
  })
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('stores and clears the token in localStorage', () => {
    api.setToken('abc')
    expect(api.getToken()).toBe('abc')
    api.setToken(null)
    expect(api.getToken()).toBeNull()
  })

  it('sends a bearer token and a JSON body when signed in', async () => {
    api.setToken('tok-1')
    const fetchMock = mockFetch(201, { id: 9 })

    const result = await api.createReview(4, { rating: 5, body: 'Nice' })

    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/cafes/4/reviews')
    expect(init.method).toBe('POST')
    expect(init.headers.Authorization).toBe('Bearer tok-1')
    expect(init.headers['Content-Type']).toBe('application/json')
    expect(JSON.parse(init.body)).toEqual({ rating: 5, body: 'Nice' })
    expect(result).toEqual({ id: 9 })
  })

  it('omits the Authorization header when signed out', async () => {
    const fetchMock = mockFetch(200, [])
    await api.me().catch(() => {})
    expect(fetchMock.mock.calls[0][1].headers.Authorization).toBeUndefined()
  })

  it('does not set a Content-Type for multipart uploads, so the browser can add the boundary', async () => {
    const fetchMock = mockFetch(201, { id: 1, url: '/uploads/x.jpg' })
    const file = new File(['data'], 'cup.png', { type: 'image/png' })

    await api.uploadReviewImage(10, file)

    const init = fetchMock.mock.calls[0][1]
    expect(init.body).toBeInstanceOf(FormData)
    expect(init.headers['Content-Type']).toBeUndefined()
  })

  it('returns null for 204 responses', async () => {
    mockFetch(204)
    await expect(api.deleteReview(3)).resolves.toBeNull()
  })

  it('throws an ApiError carrying the server message and status', async () => {
    mockFetch(409, { error: 'You have already reviewed this cafe' })
    const err = await api.createReview(1, { rating: 4, body: 'x' }).catch((e) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect(err.message).toBe('You have already reviewed this cafe')
    expect(err.status).toBe(409)
  })

  it('falls back to a generic message when the error body is not JSON', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('<html>oops</html>', { status: 502 })))
    const err = await api.me().catch((e) => e)
    expect(err.message).toBe('Request failed (502)')
  })

  it('clears the token and announces expiry when a signed-in request gets 401', async () => {
    api.setToken('stale')
    mockFetch(401, { error: 'Invalid or expired token' })
    const listener = vi.fn()
    window.addEventListener(AUTH_EXPIRED_EVENT, listener)

    await expect(api.deleteReview(2)).rejects.toThrow('Invalid or expired token')

    expect(api.getToken()).toBeNull()
    expect(listener).toHaveBeenCalledTimes(1)
    window.removeEventListener(AUTH_EXPIRED_EVENT, listener)
  })

  it('does not treat a failed login as an expired session', async () => {
    api.setToken('keep-me')
    mockFetch(401, { error: 'Invalid email or password' })
    const listener = vi.fn()
    window.addEventListener(AUTH_EXPIRED_EVENT, listener)

    await expect(api.login('a@example.com', 'bad')).rejects.toThrow('Invalid email or password')

    expect(listener).not.toHaveBeenCalled()
    window.removeEventListener(AUTH_EXPIRED_EVENT, listener)
  })
})
