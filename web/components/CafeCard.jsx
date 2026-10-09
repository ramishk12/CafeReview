import { MapPin } from 'lucide-react'
import Link from 'next/link'
import { cafeGradient } from '@/lib/cafeTheme'
import Stars from './Stars'

export default function CafeCard({ cafe }) {
  const hasReviews = cafe.review_count > 0
  return (
    <Link
      href={`/cafes/${cafe.id}`}
      className="card group flex h-full flex-col overflow-hidden no-underline transition duration-300 hover:-translate-y-1 hover:shadow-xl focus-visible:-translate-y-1"
    >
      <div className={`relative h-28 ${cafeGradient(cafe.name)}`}>
        <span className="absolute top-3 right-3 flex items-center gap-1 rounded-full bg-white/90 px-2.5 py-1 text-sm font-semibold text-slate-900 shadow-sm">
          {hasReviews ? cafe.avg_rating.toFixed(1) : 'New'}
          {hasReviews && <span aria-hidden="true" className="text-amber-500">★</span>}
        </span>
      </div>
      <div className="flex flex-1 flex-col p-5">
        <h2 className="text-lg font-semibold tracking-tight text-slate-900 group-hover:text-indigo-600 dark:text-white dark:group-hover:text-indigo-300">
          {cafe.name}
        </h2>
        {cafe.address && (
          <p className="mt-1 flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400">
            <MapPin size={14} className="shrink-0" aria-hidden="true" />
            {cafe.address}
          </p>
        )}
        <div className="mt-auto flex items-center gap-2 pt-4 text-sm">
          <Stars value={cafe.avg_rating} size={14} />
          <span className="text-slate-500 dark:text-slate-400">
            {hasReviews ? `${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'}` : 'No reviews yet'}
          </span>
        </div>
      </div>
    </Link>
  )
}
