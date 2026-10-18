import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import AddToListButton from '@/components/AddToListButton'

const listMyLists = vi.fn()
const addCafeToList = vi.fn()
const createList = vi.fn()

vi.mock('@/lib/api', () => ({
  listMyLists: (...args) => listMyLists(...args),
  addCafeToList: (...args) => addCafeToList(...args),
  removeCafeFromList: vi.fn(),
  addUserToList: vi.fn(),
  removeUserFromList: vi.fn(),
  addReviewToList: vi.fn(),
  removeReviewFromList: vi.fn(),
  createList: (...args) => createList(...args),
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ user: { id: 7 } }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('AddToListButton', () => {
  beforeEach(() => {
    listMyLists.mockReset()
    addCafeToList.mockReset()
    createList.mockReset()
  })

  it('adds the cafe to a list', async () => {
    listMyLists.mockResolvedValue([{ id: 1, name: 'Visited', contains_cafe: false, cafe_count: 0 }])
    addCafeToList.mockResolvedValue({})
    render(<AddToListButton type="cafe" id={5} />)

    const trigger = screen.getByRole('button', { name: 'Save to list' })
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    fireEvent.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')

    fireEvent.click(await screen.findByRole('checkbox'))

    await waitFor(() => expect(addCafeToList).toHaveBeenCalledWith(1, 5))
  })

  it('dismisses the panel on Escape and on an outside click', async () => {
    listMyLists.mockResolvedValue([])
    render(<AddToListButton type="cafe" id={5} />)

    const trigger = screen.getByRole('button', { name: 'Save to list' })

    fireEvent.click(trigger)
    await screen.findByText('Your lists')
    fireEvent.keyDown(document, { key: 'Escape' })
    expect(screen.queryByText('Your lists')).not.toBeInTheDocument()
    expect(trigger).toHaveFocus()

    fireEvent.click(trigger)
    await screen.findByText('Your lists')
    fireEvent.pointerDown(document.body)
    expect(screen.queryByText('Your lists')).not.toBeInTheDocument()
  })

  it('creates a list when Enter is pressed in the name field', async () => {
    listMyLists.mockResolvedValue([])
    createList.mockResolvedValue({ id: 9, name: 'Weekends', cafe_count: 0 })
    addCafeToList.mockResolvedValue({})
    render(<AddToListButton type="cafe" id={5} />)

    fireEvent.click(screen.getByRole('button', { name: 'Save to list' }))
    fireEvent.change(await screen.findByPlaceholderText('New list name'), { target: { value: 'Weekends' } })
    fireEvent.submit(screen.getByPlaceholderText('New list name').closest('form'))

    await waitFor(() => expect(createList).toHaveBeenCalledWith('Weekends'))
    await waitFor(() => expect(addCafeToList).toHaveBeenCalledWith(9, 5))
  })
})
