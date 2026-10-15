import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import ReviewCard from '@/components/ReviewCard'

const onVote = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => ({ refresh: vi.fn(), push: vi.fn() }) }))
vi.mock('next/link', () => ({ default: ({ href, children }) => <a href={href}>{children}</a> }))
vi.mock('@/lib/api', () => ({
  deleteReview: vi.fn(),
  getReviewRevisions: vi.fn(),
  listMyLists: vi.fn().mockResolvedValue([]),
  addCafeToList: vi.fn(),
  removeCafeFromList: vi.fn(),
  addUserToList: vi.fn(),
  removeUserFromList: vi.fn(),
  addReviewToList: vi.fn(),
  removeReviewFromList: vi.fn(),
  createList: vi.fn(),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ user: { id: 7 }, isAdmin: false }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

const review = {
  id: 10, user_id: 8, author_name: 'Sam Lee', rating: 4, body: 'Nice.', visited_on: '2026-10-01',
  created_at: '2026-10-01T10:00:00Z', updated_at: '2026-10-01T10:00:00Z', images: [], tags: [], revision_count: 0,
}

describe('ReviewCard', () => {
  it('lets a signed-in user mark another review helpful', () => {
    render(<ReviewCard review={review} isOwner={false} voted={false} helpfulCount={2} onVote={onVote} onOpenPhoto={() => {}} />)

    const button = screen.getByRole('button', { name: /helpful \(2\)/i })
    fireEvent.click(button)

    expect(onVote).toHaveBeenCalled()
  })

  it('does not offer a helpful vote on your own review', () => {
    render(<ReviewCard review={review} isOwner voted={false} helpfulCount={0} onVote={onVote} onOpenPhoto={() => {}} />)
    expect(screen.queryByRole('button', { name: /helpful/i })).not.toBeInTheDocument()
  })
})
