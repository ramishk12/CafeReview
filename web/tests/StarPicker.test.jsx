import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import StarPicker from '@/components/StarPicker'

describe('StarPicker', () => {
  it('selects a star on click', () => {
    const onChange = vi.fn()
    render(<StarPicker value={3} onChange={onChange} />)

    fireEvent.click(screen.getByRole('radio', { name: '5 stars' }))

    expect(onChange).toHaveBeenCalledWith(5)
  })

  it('moves between stars with the arrow keys', () => {
    const onChange = vi.fn()
    render(<StarPicker value={3} onChange={onChange} />)

    const group = screen.getByRole('radiogroup')
    fireEvent.keyDown(group, { key: 'ArrowRight' })
    expect(onChange).toHaveBeenCalledWith(4)

    fireEvent.keyDown(group, { key: 'ArrowLeft' })
    expect(onChange).toHaveBeenCalledWith(2)
  })

  it('keeps only the selected star in the tab order', () => {
    render(<StarPicker value={3} onChange={vi.fn()} />)

    expect(screen.getByRole('radio', { name: '3 stars' })).toHaveAttribute('tabindex', '0')
    expect(screen.getByRole('radio', { name: '1 star' })).toHaveAttribute('tabindex', '-1')
  })
})
