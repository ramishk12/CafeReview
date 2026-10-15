import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ReviewImageTagger from '@/components/ReviewImageTagger'

const setReviewImageTags = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => ({ refresh: vi.fn() }) }))
vi.mock('@/lib/api', () => ({
  listTags: vi.fn().mockResolvedValue([{ id: 1, name: 'Exterior', category: 'photo', scope: 'image' }]),
  setReviewImageTags: (...args) => setReviewImageTags(...args),
}))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('ReviewImageTagger', () => {
  beforeEach(() => setReviewImageTags.mockReset())

  it('tags a photo', async () => {
    setReviewImageTags.mockResolvedValue({})
    render(<ReviewImageTagger image={{ id: 5, tags: [] }} />)

    fireEvent.click(screen.getByRole('button', { name: 'Tag photo' }))
    fireEvent.click(await screen.findByRole('button', { name: 'Exterior' }))

    await waitFor(() => expect(setReviewImageTags).toHaveBeenCalledWith(5, [1]))
  })
})
