import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ListsClient from '@/app/lists/ListsClient'

const listMyLists = vi.fn()
const createList = vi.fn()

vi.mock('next/link', () => ({ default: ({ href, children }) => <a href={href}>{children}</a> }))
vi.mock('@/lib/api', () => ({
  listMyLists: (...args) => listMyLists(...args),
  createList: (...args) => createList(...args),
  renameList: vi.fn(),
  deleteList: vi.fn(),
}))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: vi.fn() }) }))

describe('ListsClient', () => {
  beforeEach(() => {
    listMyLists.mockReset()
    createList.mockReset()
  })

  it('lists the user\'s lists', async () => {
    listMyLists.mockResolvedValue([{ id: 1, name: 'Visited', cafe_count: 2 }])
    render(<ListsClient />)
    expect(await screen.findByText(/Visited/)).toBeInTheDocument()
  })

  it('creates a list from a suggestion', async () => {
    listMyLists.mockResolvedValue([])
    createList.mockResolvedValue({ id: 2, name: 'Favourites' })
    render(<ListsClient />)

    fireEvent.click(screen.getByRole('button', { name: 'Favourites' }))
    fireEvent.click(screen.getByRole('button', { name: 'Create' }))

    await waitFor(() => expect(createList).toHaveBeenCalledWith('Favourites'))
  })
})
