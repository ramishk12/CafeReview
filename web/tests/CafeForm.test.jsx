import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeForm from '@/components/CafeForm'

const routerMock = { push: vi.fn(), refresh: vi.fn() }
const createCafe = vi.fn()
const updateCafe = vi.fn()

vi.mock('next/navigation', () => ({ useRouter: () => routerMock }))
vi.mock('@/lib/api', () => ({
  createCafe: (...args) => createCafe(...args),
  updateCafe: (...args) => updateCafe(...args),
}))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('CafeForm', () => {
  beforeEach(() => {
    createCafe.mockReset()
    updateCafe.mockReset()
    routerMock.push.mockReset()
    routerMock.refresh.mockReset()
  })

  it('creates a cafe and navigates to it', async () => {
    createCafe.mockResolvedValue({ id: 5, name: 'New Cafe' })

    render(<CafeForm />)
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'New Cafe' } })
    fireEvent.click(screen.getByRole('button', { name: 'Create cafe' }))

    await waitFor(() => expect(createCafe).toHaveBeenCalledWith({ name: 'New Cafe', address: '', description: '' }))
    expect(routerMock.push).toHaveBeenCalledWith('/cafes/5')
  })

  it('edits an existing cafe and prefills the fields', async () => {
    updateCafe.mockResolvedValue({ id: 3, name: 'Bean II' })

    render(<CafeForm cafe={{ id: 3, name: 'Bean', address: '12 Market St', description: 'Nice.' }} />)
    expect(screen.getByRole('heading', { name: 'Edit cafe' })).toBeInTheDocument()
    expect(screen.getByLabelText('Name')).toHaveValue('Bean')

    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'Bean II' } })
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }))

    await waitFor(() =>
      expect(updateCafe).toHaveBeenCalledWith(3, { name: 'Bean II', address: '12 Market St', description: 'Nice.' }),
    )
    expect(routerMock.push).toHaveBeenCalledWith('/cafes/3')
  })
})
