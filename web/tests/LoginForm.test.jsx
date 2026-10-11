import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LoginForm from '@/app/login/LoginForm'

const routerMock = { replace: vi.fn(), push: vi.fn() }
let searchParams = new URLSearchParams()
const loginMock = vi.fn()
const showMock = vi.fn()

vi.mock('next/navigation', () => ({
  useRouter: () => routerMock,
  useSearchParams: () => searchParams,
}))
vi.mock('next/link', () => ({
  default: ({ href, children }) => <a href={href}>{children}</a>,
}))
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({ login: loginMock }),
}))
vi.mock('@/context/ToastContext', () => ({
  useToast: () => ({ show: showMock }),
}))

function fillAndSubmit() {
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'demo@example.com' } })
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'password123' } })
  fireEvent.click(screen.getByRole('button', { name: 'Log in' }))
}

describe('LoginForm', () => {
  beforeEach(() => {
    searchParams = new URLSearchParams()
  })

  it('redirects to the requested page after a successful login', async () => {
    searchParams = new URLSearchParams('from=/cafes/3')
    loginMock.mockResolvedValue({ display_name: 'Demo User' })

    render(<LoginForm />)
    fillAndSubmit()

    await waitFor(() => expect(routerMock.replace).toHaveBeenCalledWith('/cafes/3'))
    expect(loginMock).toHaveBeenCalledWith('demo@example.com', 'password123')
    expect(showMock).toHaveBeenCalledWith('Welcome back, Demo User')
  })

  it('ignores an external redirect target', async () => {
    searchParams = new URLSearchParams('from=/%5Cevil.example')
    loginMock.mockResolvedValue({ display_name: 'Demo User' })

    render(<LoginForm />)
    fillAndSubmit()

    await waitFor(() => expect(routerMock.replace).toHaveBeenCalledWith('/'))
  })

  it('shows the server error and re-enables the form on failure', async () => {
    const { ApiError } = await import('@/lib/api')
    loginMock.mockRejectedValue(new ApiError('Invalid email or password', 401))

    render(<LoginForm />)
    fillAndSubmit()

    expect(await screen.findByText('Invalid email or password')).toBeInTheDocument()
    expect(routerMock.replace).not.toHaveBeenCalled()
    expect(screen.getByRole('button', { name: 'Log in' })).toBeEnabled()
  })
})
