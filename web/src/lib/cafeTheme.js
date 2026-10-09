// Tailwind needs literal class names, so the gradients are listed in full.
const GRADIENTS = [
  'from-indigo-500 via-violet-500 to-fuchsia-500',
  'from-amber-400 via-orange-500 to-rose-500',
  'from-emerald-400 via-teal-500 to-cyan-600',
  'from-sky-400 via-blue-500 to-indigo-600',
  'from-rose-400 via-pink-500 to-purple-600',
  'from-lime-400 via-emerald-500 to-teal-600',
]

const AVATAR_COLORS = [
  'bg-indigo-500',
  'bg-rose-500',
  'bg-emerald-500',
  'bg-amber-500',
  'bg-sky-500',
  'bg-fuchsia-500',
]

// Picks a stable value from a string, so a cafe or user always gets the same color.
function pick(list, text) {
  let hash = 0
  for (const ch of String(text)) hash = (hash * 31 + ch.charCodeAt(0)) >>> 0
  return list[hash % list.length]
}

export const cafeGradient = (name) => pick(GRADIENTS, name)
export const avatarColor = (name) => pick(AVATAR_COLORS, name)

export function initials(name = '') {
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('') || '?'
}
