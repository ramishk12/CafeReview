import { avatarColor, initials } from '@/lib/cafeTheme'

export default function Avatar({ name, size = 'md' }) {
  const dimensions = size === 'sm' ? 'h-8 w-8 text-xs' : 'h-10 w-10 text-sm'
  return (
    <span
      aria-hidden="true"
      className={`inline-flex shrink-0 items-center justify-center rounded-full font-semibold text-white shadow-inner ${dimensions} ${avatarColor(name)}`}
    >
      {initials(name)}
    </span>
  )
}
