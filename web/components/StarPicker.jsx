'use client'

import { useRef, useState } from 'react'
import { Star } from 'lucide-react'

// Clickable 1-5 rating. Arrow keys move between the stars like a radio group,
// and only the checked star is a tab stop.
export default function StarPicker({ value, onChange }) {
  const [hover, setHover] = useState(0)
  const shown = hover || value
  const refs = useRef([])

  const move = (next) => {
    const clamped = Math.min(5, Math.max(1, next))
    if (clamped === value) return
    onChange(clamped)
    refs.current[clamped - 1]?.focus()
  }

  const onKeyDown = (event) => {
    switch (event.key) {
      case 'ArrowRight':
      case 'ArrowUp':
        event.preventDefault()
        move(value + 1)
        break
      case 'ArrowLeft':
      case 'ArrowDown':
        event.preventDefault()
        move(value - 1)
        break
      default:
        break
    }
  }

  return (
    <div role="radiogroup" aria-label="Rating" className="inline-flex items-center gap-1" onMouseLeave={() => setHover(0)} onKeyDown={onKeyDown}>
      {[1, 2, 3, 4, 5].map((n) => (
        <button
          key={n}
          ref={(element) => {
            refs.current[n - 1] = element
          }}
          type="button"
          role="radio"
          aria-checked={value === n}
          aria-label={`${n} ${n === 1 ? 'star' : 'stars'}`}
          tabIndex={value === n ? 0 : -1}
          onClick={() => onChange(n)}
          onMouseEnter={() => setHover(n)}
          className="cursor-pointer rounded-md p-0.5 transition hover:scale-110"
        >
          <Star size={26} strokeWidth={2} className={n <= shown ? 'fill-amber-400 text-amber-400' : 'text-slate-300 dark:text-slate-600'} />
        </button>
      ))}
      <span className="ml-2 text-sm text-slate-500 dark:text-slate-400">{value} / 5</span>
    </div>
  )
}
