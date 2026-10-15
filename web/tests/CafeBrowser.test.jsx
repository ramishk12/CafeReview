import { fireEvent, render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeBrowser from '@/components/CafeBrowser'

let authState = { isAdmin: false }

vi.mock('next/link', () => ({
  default: ({ href, children, ...rest }) => <a href={href} {...rest}>{children}</a>,
}))
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => authState,
}))

const cafes = [
  { id: 1, name: 'Bean There', address: '12 Market St', avg_rating: 4.5, review_count: 2 },
  { id: 2, name: 'Morning Crumb', address: '48 Elm Ave', avg_rating: 0, review_count: 0 },
  { id: 3, name: 'Matcha House', address: 'Market Lane', avg_rating: 4.8, review_count: 5 },
]

function cardNames() {
  return screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent)
}

describe('CafeBrowser', () => {
  beforeEach(() => {
    authState = { isAdmin: false }
  })

  it('sorts by top rated by default', () => {
    render(<CafeBrowser cafes={cafes} />)
    expect(cardNames()).toEqual(['Matcha House', 'Bean There', 'Morning Crumb'])
  })

  it('filters by name or address, case-insensitively', () => {
    render(<CafeBrowser cafes={cafes} />)
    fireEvent.change(screen.getByPlaceholderText(/search by name or address/i), { target: { value: 'MARKET' } })
    expect(cardNames()).toEqual(['Matcha House', 'Bean There'])
  })

  it('shows a message when nothing matches', () => {
    render(<CafeBrowser cafes={cafes} />)
    fireEvent.change(screen.getByPlaceholderText(/search by name or address/i), { target: { value: 'zzz' } })
    expect(screen.getByText(/no cafes match/i)).toBeInTheDocument()
  })

  it('changes the sort order from the dropdown', () => {
    render(<CafeBrowser cafes={cafes} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'name' } })
    expect(cardNames()).toEqual(['Bean There', 'Matcha House', 'Morning Crumb'])
  })

  it('shows the empty state when there are no cafes at all', () => {
    render(<CafeBrowser cafes={[]} />)
    expect(screen.getByText('No cafes yet.')).toBeInTheDocument()
  })

  it('links each card to its cafe page', () => {
    render(<CafeBrowser cafes={cafes} />)
    const link = within(screen.getByRole('list')).getAllByRole('link')[0]
    expect(link).toHaveAttribute('href', '/cafes/3')
  })

  it('filters cafes by tags with AND', () => {
    const tagged = [
      { id: 1, name: 'A', address: '', avg_rating: 0, review_count: 0, tags: [{ id: 10, name: 'Quiet' }, { id: 11, name: 'Matcha' }] },
      { id: 2, name: 'B', address: '', avg_rating: 0, review_count: 0, tags: [{ id: 10, name: 'Quiet' }] },
    ]
    const tags = [
      { id: 10, name: 'Quiet', category: 'vibe' },
      { id: 11, name: 'Matcha', category: 'drinks' },
    ]
    render(<CafeBrowser cafes={tagged} tags={tags} />)

    fireEvent.click(screen.getByRole('button', { name: 'Quiet' }))
    expect(cardNames()).toEqual(['A', 'B'])
    fireEvent.click(screen.getByRole('button', { name: 'Matcha' }))
    expect(cardNames()).toEqual(['A'])
  })

  it('shows a Closed badge on closed cafes', () => {
    render(<CafeBrowser cafes={[{ id: 9, name: 'Shut Cafe', address: '', avg_rating: 0, review_count: 0, closed: true }]} />)
    expect(screen.getByText('Closed')).toBeInTheDocument()
  })

  it('shows an admin edit quick-action only to admins', () => {
    render(<CafeBrowser cafes={cafes} />)
    expect(screen.queryByRole('link', { name: 'Edit' })).not.toBeInTheDocument()

    authState = { isAdmin: true }
    render(<CafeBrowser cafes={cafes} />)
    const editLinks = screen.getAllByRole('link', { name: 'Edit' })
    expect(editLinks[0]).toHaveAttribute('href', '/cafes/3/edit')
  })
})
