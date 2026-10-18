import { LoaderCircle } from 'lucide-react'

// Inline spinner for buttons and loading messages. It is decorative: the
// adjacent text carries the meaning, so it is hidden from assistive tech.
export default function Spinner({ size = 16, className = '' }) {
  return <LoaderCircle size={size} strokeWidth={2.5} aria-hidden="true" className={`animate-spin ${className}`} />
}
