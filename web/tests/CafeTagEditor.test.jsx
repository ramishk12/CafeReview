import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeTagEditor from '@/components/CafeTagEditor'

const setCafeTags = vi.fn()
const listTags = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => ({ refresh: vi.fn() }) }))
vi.mock('@/lib/api', () => ({
  listTags: (...args) => listTags(...args),
  setCafeTags: (...args) => setCafeTags(...args),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ isAdmin: true }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('CafeTagEditor', () => {
  beforeEach(() => {
    setCafeTags.mockReset()
    listTags.mockReset()
    listTags.mockResolvedValue([
      { id: 10, name: 'Quiet', category: 'vibe' },
      { id: 11, name: 'Matcha', category: 'drinks' },
    ])
  })

  it('saves the selected tags', async () => {
    setCafeTags.mockResolvedValue({})
    render(<CafeTagEditor cafeId={1} initialTagIds={[]} />)

    fireEvent.click(await screen.findByRole('button', { name: 'Quiet' }))
    fireEvent.click(screen.getByRole('button', { name: 'Save tags' }))

    await waitFor(() => expect(setCafeTags).toHaveBeenCalledWith(1, [10]))
  })

  it('starts with the cafe\'s current tags selected and can clear them', async () => {
    setCafeTags.mockResolvedValue({})
    render(<CafeTagEditor cafeId={1} initialTagIds={[11]} />)

    const matcha = await screen.findByRole('button', { name: 'Matcha' })
    expect(matcha).toHaveAttribute('aria-pressed', 'true')

    fireEvent.click(matcha) // deselect
    fireEvent.click(screen.getByRole('button', { name: 'Save tags' }))
    await waitFor(() => expect(setCafeTags).toHaveBeenCalledWith(1, []))
  })
})
