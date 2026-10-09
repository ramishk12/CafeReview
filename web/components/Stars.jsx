import { Star } from 'lucide-react'

// Read-only rating out of 5. value may be fractional (e.g. an average).
export default function Stars({ value, size = 16 }) {
  const rounded = Math.round(Number(value) || 0)
  return (
    <span className="inline-flex items-center gap-0.5" role="img" aria-label={`${rounded} out of 5 stars`}>
      {Array.from({ length: 5 }, (_, i) => (
        <Star
          key={i}
          size={size}
          strokeWidth={2}
          aria-hidden="true"
          className={i < rounded ? 'fill-amber-400 text-amber-400' : 'text-slate-300 dark:text-slate-700'}
        />
      ))}
    </span>
  )
}
