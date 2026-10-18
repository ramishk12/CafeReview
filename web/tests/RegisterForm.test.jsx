import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterForm from '@/app/register/RegisterForm'

const registerMock = vi.fn()
const showToast = vi.fn()
const resendVerification = vi.fn()

vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ register: registerMock }) }))
vi.mock('@/context/ToastContext', () => ({ useToast: () => ({ show: showToast }) }))
vi.mock('@/lib/api', () => ({ resendVerification: (...args) => resendVerification(...args) }))
vi.mock('next/link', () => ({ default: ({ href, children }) => <a href={href}>{children}</a> }))

function fillForm() {
  fireEvent.change(screen.getByLabelText('Display name'), { target: { value: 'Ann' } })
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'ann@example.com' } })
  fireEvent.change(screen.getByLabelText('Password (at least 8 characters)'), { target: { value: 'password123' } })
  fireEvent.click(screen.getByRole('button', { name: 'Sign up' }))
}

describe('RegisterForm', () => {
  beforeEach(() => {
    registerMock.mockReset()
    showToast.mockReset()
    resendVerification.mockReset()
    resendVerification.mockResolvedValue({})
  })

  it('shows the check-your-inbox screen after signup and does not sign the user in', async () => {
    registerMock.mockResolvedValue({ message: 'ok' })

    render(<RegisterForm />)
    fillForm()

    expect(await screen.findByText('Check your inbox')).toBeInTheDocument()
    expect(screen.getByText('ann@example.com')).toBeInTheDocument()
    expect(registerMock).toHaveBeenCalledWith('ann@example.com', 'password123', 'Ann')
    expect(screen.queryByRole('button', { name: 'Sign up' })).not.toBeInTheDocument()
  })

  it('shows the error and keeps the form on failure', async () => {
    registerMock.mockRejectedValue(new Error('Email is required'))

    render(<RegisterForm />)
    fireEvent.change(screen.getByLabelText('Display name'), { target: { value: 'Ann' } })
    fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'ann@example.com' } })
    fireEvent.change(screen.getByLabelText('Password (at least 8 characters)'), { target: { value: 'password123' } })
    fillForm()

    expect(await screen.findByText('Email is required')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Sign up' })).toBeEnabled()
  })

  it('confirms a verification resend', async () => {
    registerMock.mockResolvedValue({ message: 'ok' })

    render(<RegisterForm />)
    fillForm()
    await screen.findByText('Check your inbox')

    fireEvent.click(screen.getByRole('button', { name: 'resend the email' }))

    await waitFor(() => expect(showToast).toHaveBeenCalledWith('Verification email sent'))
  })

  it('reports a failed verification resend', async () => {
    registerMock.mockResolvedValue({ message: 'ok' })
    resendVerification.mockRejectedValue(new Error('Too many attempts'))

    render(<RegisterForm />)
    fillForm()
    await screen.findByText('Check your inbox')

    fireEvent.click(screen.getByRole('button', { name: 'resend the email' }))

    await waitFor(() => expect(showToast).toHaveBeenCalledWith('Too many attempts', 'error'))
  })
})
