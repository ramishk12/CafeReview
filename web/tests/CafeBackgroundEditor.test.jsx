import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeBackgroundEditor from '@/components/CafeBackgroundEditor'

const setCafeBackground = vi.fn()
const deleteCafeBackground = vi.fn()
const setCafeBackgroundImage = vi.fn()
const clearCafeBackgroundImage = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => ({ refresh: vi.fn() }) }))
vi.mock('@/lib/api', () => ({
  listTags: vi.fn().mockResolvedValue([]),
  setCafeBackground: (...args) => setCafeBackground(...args),
  deleteCafeBackground: (...args) => deleteCafeBackground(...args),
  setCafeBackgroundImage: (...args) => setCafeBackgroundImage(...args),
  clearCafeBackgroundImage: (...args) => clearCafeBackgroundImage(...args),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ isAdmin: true }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('CafeBackgroundEditor', () => {
  beforeEach(() => {
    setCafeBackground.mockReset()
    deleteCafeBackground.mockReset()
    setCafeBackgroundImage.mockReset()
    clearCafeBackgroundImage.mockReset()
  })

  it('uploads a chosen image', async () => {
    setCafeBackground.mockResolvedValue({})
    const { container } = render(<CafeBackgroundEditor cafeId={1} backgroundUrl="" />)

    const file = new File(['x'], 'bg.png', { type: 'image/png' })
    fireEvent.change(container.querySelector('input[type="file"]'), { target: { files: [file] } })

    await waitFor(() => expect(setCafeBackground).toHaveBeenCalledWith(1, file))
  })

  it('removes the uploaded image', async () => {
    deleteCafeBackground.mockResolvedValue({})
    render(<CafeBackgroundEditor cafeId={1} backgroundUrl="/uploads/bg.png" backgroundSource="uploaded" />)

    fireEvent.click(screen.getByRole('button', { name: 'Remove uploaded' }))

    await waitFor(() => expect(deleteCafeBackground).toHaveBeenCalledWith(1))
  })

  it('sets the background from a review photo', async () => {
    setCafeBackgroundImage.mockResolvedValue({})
    const images = [{ id: 5, url: '/uploads/a.jpg', tags: [] }]
    render(<CafeBackgroundEditor cafeId={1} backgroundUrl="" images={images} />)

    fireEvent.click(screen.getByRole('button', { name: 'Use this photo as the background' }))

    await waitFor(() => expect(setCafeBackgroundImage).toHaveBeenCalledWith(1, 5))
  })
})
