import { render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RequireAuth from '@/components/RequireAuth'

let auth = { user: null, loading: true, isAdmin: false }
const replace = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => ({ replace }),
  usePathname: () => '/lists',
}))
vi.mock('@/context/AuthContext', () => ({ useAuth: () => auth }))

describe('RequireAuth', () => {
  beforeEach(() => {
    auth = { user: null, loading: true, isAdmin: false }
    replace.mockReset()
  })

  it('shows the shared skeleton while the session is loading', () => {
    render(
      <RequireAuth>
        <p>Secret</p>
      </RequireAuth>,
    )

    expect(screen.getByLabelText('Loading')).toBeInTheDocument()
    expect(screen.queryByText('Secret')).not.toBeInTheDocument()
  })

  it('renders the page for a signed-in user', () => {
    auth = { user: { id: 1, role: 'user' }, loading: false, isAdmin: false }
    render(
      <RequireAuth>
        <p>Secret</p>
      </RequireAuth>,
    )

    expect(screen.getByText('Secret')).toBeInTheDocument()
  })

  it('sends signed-out visitors to the login page', () => {
    auth = { user: null, loading: false, isAdmin: false }
    render(
      <RequireAuth>
        <p>Secret</p>
      </RequireAuth>,
    )

    expect(replace).toHaveBeenCalledWith('/login?from=%2Flists')
  })
})
