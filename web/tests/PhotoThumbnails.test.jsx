import { fireEvent, render, screen } from '@testing-library/react'
import { beforeAll, describe, expect, it } from 'vitest'
import PhotoThumbnails from '@/components/PhotoThumbnails'

// jsdom does not always implement <dialog>; the lightbox only needs showModal
// and close to toggle the open attribute for this test.
beforeAll(() => {
  const proto = window.HTMLDialogElement?.prototype
  if (proto && !proto.showModal) {
    proto.showModal = function showModal() {
      this.setAttribute('open', '')
    }
    proto.close = function close() {
      this.removeAttribute('open')
    }
  }
})

describe('PhotoThumbnails', () => {
  it('opens the lightbox for the clicked photo', () => {
    render(<PhotoThumbnails images={[{ id: 1, url: '/uploads/a.jpg' }]} authorName="Demo User" />)

    fireEvent.click(screen.getByRole('button', { name: 'Open photo by demo user' }))

    expect(screen.getByAltText('Photo by Demo User')).toBeInTheDocument()
  })

  it('renders nothing without photos', () => {
    const { container } = render(<PhotoThumbnails images={[]} authorName="Demo User" />)
    expect(container).toBeEmptyDOMElement()
  })
})
