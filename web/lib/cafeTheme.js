// Tailwind needs literal class names, so the gradients are listed in full.
const GRADIENTS = [
  'bg-linear-to-br from-indigo-500 via-violet-500 to-fuchsia-500',
  'bg-linear-to-br from-amber-400 via-orange-500 to-rose-500',
  'bg-linear-to-br from-emerald-400 via-teal-500 to-cyan-600',
  'bg-linear-to-br from-sky-400 via-blue-500 to-indigo-600',
  'bg-linear-to-br from-rose-400 via-pink-500 to-purple-600',
  'bg-linear-to-br from-lime-400 via-emerald-500 to-teal-600',
]

const AVATAR_COLORS = ['bg-indigo-500', 'bg-rose-500', 'bg-emerald-500', 'bg-amber-500', 'bg-sky-500', 'bg-fuchsia-500']

// Picks a stable value from a string, so the same cafe or user always gets the same color.
function pick(list, text) {
  let hash = 0
  for (const ch of String(text)) hash = (hash * 31 + ch.charCodeAt(0)) >>> 0
  return list[hash % list.length]
}

export const cafeGradient = (name) => pick(GRADIENTS, name)
export const avatarColor = (name) => pick(AVATAR_COLORS, name)

export function initials(name = '') {
  return (
    name
      .trim()
      .split(/\s+/)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? '')
      .join('') || '?'
  )
}

// Fixed locale and timezone so server and browser render the same text (no hydration mismatch).
export function formatDate(value) {
  return new Date(value).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC' })
}

export function ratingLabel(count, average) {
  if (!count) return 'No reviews yet'
  return `${count} ${count === 1 ? 'review' : 'reviews'}${average ? ` · ${average.toFixed(1)} avg` : ''}`
}
