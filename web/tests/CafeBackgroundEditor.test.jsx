import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeBackgroundEditor from '@/components/CafeBackgroundEditor'

const setCafeBackground = vi.fn()
const deleteCafeBackground = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => ({ refresh: vi.fn() }) }))
vi.mock('@/lib/api', () => ({
  setCafeBackground: (...args) => setCafeBackground(...args),
  deleteCafeBackground: (...args) => deleteCafeBackground(...args),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ isAdmin: true }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('CafeBackgroundEditor', () => {
  beforeEach(() => {
    setCafeBackground.mockReset()
    deleteCafeBackground.mockReset()
  })

  it('uploads a chosen image', async () => {
    setCafeBackground.mockResolvedValue({})
    const { container } = render(<CafeBackgroundEditor cafeId={1} backgroundUrl="" />)

    const file = new File(['x'], 'bg.png', { type: 'image/png' })
    fireEvent.change(container.querySelector('input[type="file"]'), { target: { files: [file] } })

    await waitFor(() => expect(setCafeBackground).toHaveBeenCalledWith(1, file))
  })

  it('removes the current image', async () => {
    deleteCafeBackground.mockResolvedValue({})
    render(<CafeBackgroundEditor cafeId={1} backgroundUrl="/uploads/bg.png" />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove' }))

    await waitFor(() => expect(deleteCafeBackground).toHaveBeenCalledWith(1))
  })
})
