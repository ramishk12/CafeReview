import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ReviewSection, { reviewPhotos } from '@/components/ReviewSection'
import { listMyReviewVotes, listTags } from '@/lib/api'

let authState = { user: null }
let params = new URLSearchParams()
const push = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({ refresh: vi.fn(), push }),
  usePathname: () => '/cafes/1',
  useSearchParams: () => params,
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
vi.mock('@/lib/api', async (importOriginal) => ({
  ...(await importOriginal()),
  listTags: vi.fn(),
  listMyReviewVotes: vi.fn(),
}))

const cafe = { id: 1, name: 'Bean There' }
const reviews = [
  {
    id: 10, user_id: 7, author_name: 'Demo User', rating: 5, body: 'Excellent.', visited_on: '2026-09-28',
    created_at: '2026-10-01T10:00:00Z', updated_at: '2026-10-01T10:00:00Z',
    images: [{ id: 100, url: '/uploads/a.jpg' }], tags: [{ id: 1, name: 'Seasonal menu' }], revision_count: 2,
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
    params = new URLSearchParams()
    push.mockReset()
    listTags.mockResolvedValue([{ id: 3, name: 'Limited edition', scope: 'review' }])
    listMyReviewVotes.mockResolvedValue([])
  })

  // The list is sorted, filtered, and paged on the server, so the component
  // renders the reviews it is given and the controls only update the URL.
  it('renders the reviews it is given, in order', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    const bodies = screen.getAllByText(/Excellent\.|Too cold\./).map((el) => el.textContent)
    expect(bodies).toEqual(['Excellent.', 'Too cold.'])
  })

  it('navigates to the chosen sort', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    fireEvent.change(screen.getByRole('combobox', { name: 'Sort' }), { target: { value: 'helpful' } })
    expect(push).toHaveBeenCalledWith('/cafes/1?sort=helpful')
  })

  it('navigates when the rating filter changes', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    fireEvent.change(screen.getByRole('combobox', { name: 'Rating' }), { target: { value: '4' } })
    expect(push).toHaveBeenCalledWith('/cafes/1?rating=4')
  })

  it('toggles the with-photos filter', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    fireEvent.click(screen.getByRole('button', { name: 'With photos' }))
    expect(push).toHaveBeenCalledWith('/cafes/1?photos=1')
  })

  it('filters to the signed-in user\'s own reviews', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' } }
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    fireEvent.click(screen.getByRole('button', { name: 'My reviews' }))
    expect(push).toHaveBeenCalledWith('/cafes/1?user_id=7')
  })

  it('toggles a review tag filter', async () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={2} />)
    fireEvent.click(await screen.findByRole('button', { name: 'Limited edition' }))
    expect(push).toHaveBeenCalledWith('/cafes/1?tags=3')
  })

  it('shows a filter empty state and a way to clear filters', () => {
    params = new URLSearchParams('rating=4')
    render(<ReviewSection cafe={cafe} reviews={[]} total={0} />)
    expect(screen.getByText('No reviews match your filters.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Clear filters' })).toBeInTheDocument()
  })

  it('shows a pager and navigates between pages', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} total={25} page={1} pageSize={10} />)
    fireEvent.click(screen.getByRole('button', { name: '2' }))
    expect(push).toHaveBeenCalledWith('/cafes/1?page=2')
  })

  it('asks signed-out visitors to log in instead of showing a form', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByRole('link', { name: 'Log in to review' })).toHaveAttribute('href', '/login?from=/cafes/1')
    expect(screen.queryByRole('button', { name: 'Post review' })).not.toBeInTheDocument()
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
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }))
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
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument()
  })

  it('links a review author to their profile', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByRole('link', { name: 'Demo User' })).toHaveAttribute('href', '/users/7')
  })

  it('shows the edit history control to the owner of an edited review', () => {
    authState = { user: { id: 7, display_name: 'Demo User', role: 'user' }, isAdmin: false }
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.getByRole('button', { name: /view history \(2\)/i })).toBeInTheDocument()
  })

  it('hides the edit history from other visitors', () => {
    render(<ReviewSection cafe={cafe} reviews={reviews} />)
    expect(screen.queryByRole('button', { name: /view history/i })).not.toBeInTheDocument()
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
