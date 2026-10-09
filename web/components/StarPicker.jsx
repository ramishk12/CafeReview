'use client'

import { useState } from 'react'
import { Star } from 'lucide-react'

// Clickable 1-5 rating. Keyboard users can arrow through it as a radio group.
export default function StarPicker({ value, onChange }) {
  const [hover, setHover] = useState(0)
  const shown = hover || value

  return (
    <div role="radiogroup" aria-label="Rating" className="inline-flex items-center gap-1" onMouseLeave={() => setHover(0)}>
      {[1, 2, 3, 4, 5].map((n) => (
        <button
          key={n}
          type="button"
          role="radio"
          aria-checked={value === n}
          aria-label={`${n} ${n === 1 ? 'star' : 'stars'}`}
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
