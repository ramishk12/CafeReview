import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import EmptyState from '@/components/EmptyState'

describe('EmptyState', () => {
  it('shows a title, description, and action', () => {
    render(
      <EmptyState
        title="Nothing here"
        description="Add something to get started."
        action={<button type="button">Do it</button>}
      />,
    )

    expect(screen.getByRole('heading', { name: 'Nothing here' })).toBeInTheDocument()
    expect(screen.getByText('Add something to get started.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Do it' })).toBeInTheDocument()
  })

  it('can render its title as another heading level', () => {
    render(<EmptyState titleAs="h1" title="Something broke" />)

    expect(screen.getByRole('heading', { level: 1, name: 'Something broke' })).toBeInTheDocument()
  })
})
