import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterForm from '@/app/register/RegisterForm'

const registerMock = vi.fn()

vi.mock('@/context/AuthContext', () => ({ useAuth: () => ({ register: registerMock }) }))
vi.mock('@/lib/api', () => ({ resendVerification: vi.fn().mockResolvedValue({}) }))
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
})
