import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ListDetailClient from '@/app/lists/[id]/ListDetailClient'

const getList = vi.fn()

vi.mock('next/navigation', () => ({
  useParams: () => ({ id: '1' }),
  useRouter: () => ({ push: vi.fn(), refresh: vi.fn() }),
}))
vi.mock('next/link', () => ({ default: ({ href, children }) => <a href={href}>{children}</a> }))
vi.mock('@/lib/api', () => ({
  getList: (...args) => getList(...args),
  deleteList: vi.fn(),
  removeCafeFromList: vi.fn(),
  removeUserFromList: vi.fn(),
  removeReviewFromList: vi.fn(),
}))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('ListDetailClient', () => {
  beforeEach(() => getList.mockReset())

  it('shows cafes, reviewers, and reviews', async () => {
    getList.mockResolvedValue({
      id: 1,
      name: 'Favourites',
      cafes: [{ id: 1, name: 'Bean There', address: '', avg_rating: 4.5, review_count: 1 }],
      users: [{ id: 7, display_name: 'Demo User' }],
      reviews: [
        {
          id: 10, cafe_id: 1, cafe_name: 'Bean There', user_id: 7, author_name: 'Demo User',
          rating: 5, body: 'Great.', visited_on: '2026-10-01', created_at: '2026-10-01T10:00:00Z', images: [], tags: [],
        },
      ],
    })

    render(<ListDetailClient />)

    expect(await screen.findByRole('heading', { name: 'Cafes' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Reviewers' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'Reviews' })).toBeInTheDocument()
    expect(screen.getAllByRole('link', { name: 'Demo User' })[0]).toHaveAttribute('href', '/users/7')
    expect(screen.getByText('Great.')).toBeInTheDocument()
  })

  it('shows an empty state for a list with nothing in it', async () => {
    getList.mockResolvedValue({ id: 1, name: 'Favourites', cafes: [], users: [], reviews: [] })

    render(<ListDetailClient />)

    expect(await screen.findByText('This list is empty')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Browse cafes' })).toHaveAttribute('href', '/')
  })
})
