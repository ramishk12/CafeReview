import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AddToListButton from '@/components/AddToListButton'

const listMyLists = vi.fn()
const addCafeToList = vi.fn()

vi.mock('@/lib/api', () => ({
  listMyLists: (...args) => listMyLists(...args),
  addCafeToList: (...args) => addCafeToList(...args),
  removeCafeFromList: vi.fn(),
  createList: vi.fn(),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ user: { id: 7 } }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('AddToListButton', () => {
  beforeEach(() => {
    listMyLists.mockReset()
    addCafeToList.mockReset()
  })

  it('adds the cafe to a list', async () => {
    listMyLists.mockResolvedValue([{ id: 1, name: 'Visited', contains_cafe: false, cafe_count: 0 }])
    addCafeToList.mockResolvedValue({})
    render(<AddToListButton cafeId={5} />)

    fireEvent.click(screen.getByRole('button', { name: 'Save to list' }))
    fireEvent.click(await screen.findByRole('checkbox'))

    await waitFor(() => expect(addCafeToList).toHaveBeenCalledWith(1, 5))
  })
})
