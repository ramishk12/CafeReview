import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CafeBrowser from '@/components/CafeBrowser'

const push = vi.fn()
let params = new URLSearchParams()

vi.mock('next/link', () => ({
  default: ({ href, children, ...rest }) => <a href={href} {...rest}>{children}</a>,
}))
vi.mock('next/navigation', () => ({
  useRouter: () => ({ push }),
  usePathname: () => '/',
  useSearchParams: () => params,
}))
let authState = { isAdmin: false }
vi.mock('@/context/AuthContext', () => ({ useAuth: () => authState }))

const cafes = [
  { id: 1, name: 'Bean There', address: '12 Market St', avg_rating: 4.5, review_count: 2 },
  { id: 2, name: 'Morning Crumb', address: '48 Elm Ave', avg_rating: 0, review_count: 0 },
]

describe('CafeBrowser', () => {
  beforeEach(() => {
    authState = { isAdmin: false }
    push.mockReset()
    params = new URLSearchParams()
  })

  it('renders the cafes it is given, in order', () => {
    render(<CafeBrowser cafes={cafes} />)
    const names = screen.getAllByRole('heading', { level: 2 }).map((h) => h.textContent)
    expect(names).toEqual(['Bean There', 'Morning Crumb'])
  })

  it('updates the URL when the sort changes', () => {
    render(<CafeBrowser cafes={cafes} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'name' } })
    expect(push).toHaveBeenCalledWith('/?sort=name')
  })

  it('offers every API sort mode, including Newest', () => {
    render(<CafeBrowser cafes={cafes} />)
    const options = screen.getAllByRole('option').map((o) => o.textContent)
    expect(options).toEqual(['Top rated', 'Most reviewed', 'Newest', 'A–Z'])
  })

  it('sorts by Newest when selected', () => {
    render(<CafeBrowser cafes={cafes} />)
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'newest' } })
    expect(push).toHaveBeenCalledWith('/?sort=newest')
  })

  it('updates the URL when a tag is toggled', () => {
    const tags = [{ id: 10, name: 'Quiet', category: 'vibe' }]
    render(<CafeBrowser cafes={cafes} tags={tags} />)
    fireEvent.click(screen.getByRole('button', { name: 'Quiet' }))
    expect(push).toHaveBeenCalledWith('/?tags=10')
  })

  it('shows a pager and navigates between pages', () => {
    render(<CafeBrowser cafes={cafes} total={50} page={1} pageSize={24} />)
    fireEvent.click(screen.getByRole('button', { name: '2' }))
    expect(push).toHaveBeenCalledWith('/?page=2')
  })

  it('shows an empty state when there are no cafes at all', () => {
    render(<CafeBrowser cafes={[]} />)
    expect(screen.getByText('No cafes yet')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Clear search' })).not.toBeInTheDocument()
  })

  it('offers a way to clear the search when nothing matches', () => {
    params = new URLSearchParams('q=tea&tags=3')
    render(<CafeBrowser cafes={[]} />)
    expect(screen.getByText('No cafes match your search')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }))
    expect(push).toHaveBeenCalledWith('/')
  })

  it('offers to add a cafe to an admin looking at an empty list', () => {
    authState = { isAdmin: true }
    render(<CafeBrowser cafes={[]} />)
    expect(screen.getByRole('link', { name: 'Add a cafe' })).toHaveAttribute('href', '/cafes/new')
  })
})
