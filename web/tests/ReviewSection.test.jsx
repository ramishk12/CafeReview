import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ReviewSection, { reviewPhotos } from '@/components/ReviewSection'

let authState = { user: null }

vi.mock('next/navigation', () => ({
  useRouter: () => ({ refresh: vi.fn(), replace: vi.fn() }),
}))
vi.mock('next/link', () => ({
  default: ({ href, children }) => <a href={href}>{children}</a>,
}))
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => authState,
}))
vi.mock('@/context/ToastContext', () => ({
  useToast: () => ({ show: vi.fn() }),
}))

const cafe = { id: 1, name: 'Bean There' }
const reviews = [
  {
    id: 10, user_id: 7, author_name: 'Demo User', rating: 5, body: 'Excellent.', visited_on: '2026-09-28',
    created_at: '2026-10-01T10:00:00Z', updated_at: '2026-10-01T10:00:00Z',
    images: [{ id: 100, url: '/uploads/a.jpg' }], tags: [{ id: 1, name: 'Seasonal menu' }],
  },
  {
    id: 11, user_id: 8, author_name: 'Sam Lee', rating: 2, body: 'Too cold.', visited_on: '2026-10-02',
    created_at: '2026-10-02T10:00:00Z', updated_at: '2026-10-02T10:00:00Z', images: [],
  },
]

describe('reviewPhotos', () => {
  it('maps images to lightbox items labelled with the author', () => {
    expect(reviewPhotos(reviews[0])).toEqual([{ id: 100, url: '/uploads/a.jpg', alt: 'Photo by Demo User' }])
  })

  it('returns an empty list when a review has no images field', () => {
    expect(reviewPhotos({ author_name: 'X' })).toEqual([])
  })
})

describe('ReviewSection', () => {
  beforeEach(() => {
    authState = { user: null }
  })

  it('asks signed-out visitors to log in instead of showing a form', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByRole('link', { name: 'Log in to review' })).toHaveAttribute('href', '/login?from=/cafes/1')
    expect(screen.queryByRole('button', { name: 'Post review' })).not.toBeInTheDocument()
  })

  it('lists reviews newest first by default', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    const bodies = screen.getAllByText(/Excellent\.|Too cold\./).map((el) => el.textContent)
    expect(bodies).toEqual(['Too cold.', 'Excellent.'])
  })

  it('re-sorts by highest rated', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'highest' } })
    const bodies = screen.getAllByText(/Excellent\.|Too cold\./).map((el) => el.textContent)
    expect(bodies).toEqual(['Excellent.', 'Too cold.'])
  })

  it('shows a form to add a visit for a signed-in user', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' } }
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByRole('heading', { name: 'Write a review' })).toBeInTheDocument()
    expect(screen.getByLabelText('Visit date')).toBeInTheDocument()
  })

  it('opens the edit form when a user edits their own review', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' } }
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    fireEvent.click(screen.getByRole('button', { name: /edit/i }))
    expect(screen.getByRole('heading', { name: 'Edit your review' })).toBeInTheDocument()
  })

  it('shows the delete button only on the signed-in user\'s own reviews', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' } }
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getAllByRole('button', { name: /^delete$/i })).toHaveLength(1)
  })

  it('shows a closed notice and hides the add and edit actions for a closed cafe', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' } }
    render(<ReviewSection cafe={{ id: 1, name: 'Bean There', closed: true }} reviews={reviews} />)
    expect(screen.getByText(/no longer accepting new reviews/i)).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Write a review' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument()
  })

  it('shows a review\'s tags', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByText('Seasonal menu')).toBeInTheDocument()
  })

  it('shows an empty state with no reviews', () => {
    render(<ReviewSection cafe={cafe} reviews={[]} />)
    expect(screen.getByText('No reviews yet. Be the first.')).toBeInTheDocument()
  })
})
