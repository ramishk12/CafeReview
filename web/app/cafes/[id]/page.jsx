import { ArrowLeft, MapPin } from 'lucide-react'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import ReviewSection from '@/components/ReviewSection'
import Stars from '@/components/Stars'
import { getCafe, listReviews } from '@/lib/serverApi'
import { cafeGradient, ratingLabel } from '@/lib/cafeTheme'

export async function generateMetadata({ params }) {
  const { id } = await params
  const cafe = await getCafe(id).catch(() => null)
  if (!cafe) return { title: 'Cafe not found' }
  return {
    title: cafe.name,
    description: cafe.description || `Reviews and photos for ${cafe.name}.`,
  }
}

export default async function CafePage({ params }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()

  const [cafe, reviews] = await Promise.all([getCafe(id), listReviews(id)])
  if (!cafe) notFound()

  const hasReviews = cafe.review_count > 0

  return (
    <section>
      <Link
        href="/"
        className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 no-underline hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
      >
        <ArrowLeft size={16} aria-hidden="true" /> All cafes
      </Link>

      <header className={`relative mt-4 mb-10 overflow-hidden rounded-3xl p-8 text-white shadow-xl ${cafeGradient(cafe.name)}`}>
        <div className="absolute inset-0 bg-black/10" aria-hidden="true" />
        <div className="relative">
          <h1 className="text-4xl font-extrabold tracking-tight drop-shadow-sm">{cafe.name}</h1>
          {cafe.address && (
            <p className="mt-2 flex items-center gap-1.5 text-white/90">
              <MapPin size={16} aria-hidden="true" /> {cafe.address}
            </p>
          )}
          {cafe.description && <p className="mt-3 max-w-2xl text-white/90">{cafe.description}</p>}
          <div className="mt-5 inline-flex items-center gap-3 rounded-2xl bg-white/15 px-4 py-2 backdrop-blur-sm">
            <span className="text-2xl font-bold">{hasReviews ? cafe.avg_rating.toFixed(1) : '–'}</span>
            <Stars value={cafe.avg_rating} size={18} />
            <span className="text-sm text-white/85">{ratingLabel(cafe.review_count)}</span>
          </div>
        </div>
      </header>

      <ReviewSection cafe={cafe} reviews={reviews} />
    </section>
  )
}
