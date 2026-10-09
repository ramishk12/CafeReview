'use client'

import { ChevronLeft, ChevronRight, X } from 'lucide-react'
import { useEffect, useRef } from 'react'

// Full-screen photo viewer built on <dialog>. Escape closes it natively; arrow keys page through.
export default function PhotoLightbox({ images, index, onIndexChange, onClose }) {
  const ref = useRef(null)
  const open = index !== null

  useEffect(() => {
    const dialog = ref.current
    if (!dialog) return
    if (open && !dialog.open) dialog.showModal()
    if (!open && dialog.open) dialog.close()
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKey = (e) => {
      if (e.key === 'ArrowRight') onIndexChange((index + 1) % images.length)
      if (e.key === 'ArrowLeft') onIndexChange((index - 1 + images.length) % images.length)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, index, images.length, onIndexChange])

  const current = open ? images[index] : null
  const multiple = images.length > 1

  return (
    <dialog
      ref={ref}
      onClose={onClose}
      onClick={(e) => {
        if (e.target === ref.current) onClose()
      }}
      className="m-auto max-h-screen max-w-screen bg-transparent p-4 backdrop:bg-black/80 backdrop:backdrop-blur-sm"
    >
      {current && (
        <div className="relative flex flex-col items-center gap-3">
          <img src={current.url} alt={current.alt} className="max-h-[80vh] max-w-[90vw] rounded-2xl object-contain shadow-2xl" />
          <div className="flex items-center gap-3">
            {multiple && (
              <button
                type="button"
                onClick={() => onIndexChange((index - 1 + images.length) % images.length)}
                className="btn btn-ghost rounded-full bg-white/10 p-2 text-white hover:bg-white/20"
                aria-label="Previous photo"
              >
                <ChevronLeft size={20} />
              </button>
            )}
            <span className="text-sm text-white/80">
              {index + 1} / {images.length}
            </span>
            {multiple && (
              <button
                type="button"
                onClick={() => onIndexChange((index + 1) % images.length)}
                className="btn btn-ghost rounded-full bg-white/10 p-2 text-white hover:bg-white/20"
                aria-label="Next photo"
              >
                <ChevronRight size={20} />
              </button>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="btn btn-ghost absolute -top-2 -right-2 rounded-full bg-white/15 p-2 text-white hover:bg-white/25"
            aria-label="Close"
          >
            <X size={18} />
          </button>
        </div>
      )}
    </dialog>
  )
}
