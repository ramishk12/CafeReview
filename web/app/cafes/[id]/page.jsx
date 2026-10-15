import { ArrowLeft, MapPin } from 'lucide-react'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import CafeAdminActions from '@/components/CafeAdminActions'
import CafeBackgroundEditor from '@/components/CafeBackgroundEditor'
import CafeTagEditor from '@/components/CafeTagEditor'
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
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 no-underline hover:text-slate-900 dark:text-slate-400 dark:hover:text-white"
        >
          <ArrowLeft size={16} aria-hidden="true" /> All cafes
        </Link>
        <CafeAdminActions cafeId={cafe.id} cafeName={cafe.name} closed={cafe.closed} />
      </div>

      <header
        className={`relative mt-4 mb-10 overflow-hidden rounded-3xl p-8 text-white shadow-xl ${cafe.background_image_url ? 'bg-slate-900' : cafeGradient(cafe.name)}`}
        style={
          cafe.background_image_url
            ? { backgroundImage: `url(${cafe.background_image_url})`, backgroundSize: 'cover', backgroundPosition: 'center' }
            : undefined
        }
      >
        <div className={`absolute inset-0 ${cafe.background_image_url ? 'bg-black/50' : 'bg-black/10'}`} aria-hidden="true" />
        <div className="relative">
          <div className="flex flex-wrap items-center gap-3">
            <h1 className="text-4xl font-extrabold tracking-tight drop-shadow-sm">{cafe.name}</h1>
            {cafe.closed && (
              <span className="rounded-full bg-white/20 px-3 py-1 text-xs font-bold uppercase tracking-wide backdrop-blur-sm">
                Closed
              </span>
            )}
          </div>
          {cafe.address && (
            <p className="mt-2 flex items-center gap-1.5 text-white/90">
              <MapPin size={16} aria-hidden="true" /> {cafe.address}
            </p>
          )}
          {cafe.description && <p className="mt-3 max-w-2xl text-white/90">{cafe.description}</p>}
          {cafe.tags?.length > 0 && (
            <div className="mt-4 flex flex-wrap gap-2">
              {cafe.tags.map((tag) => (
                <span key={tag.id} className="rounded-full bg-white/20 px-3 py-1 text-xs font-medium backdrop-blur-sm">
                  {tag.name}
                </span>
              ))}
            </div>
          )}
          <div className="mt-5 inline-flex items-center gap-3 rounded-2xl bg-white/15 px-4 py-2 backdrop-blur-sm">
            <span className="text-2xl font-bold">{hasReviews ? cafe.avg_rating.toFixed(1) : '–'}</span>
            <Stars value={cafe.avg_rating} size={18} />
            <span className="text-sm text-white/85">{ratingLabel(cafe.review_count, cafe.reviewer_count)}</span>
          </div>
        </div>
      </header>

      {cafe.closed && (
        <div className="mb-8 rounded-2xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800 dark:border-rose-900/50 dark:bg-rose-950/40 dark:text-rose-200">
          This cafe is closed. New reviews are not accepted, and existing reviews can no longer be edited. Reviews can
          still be deleted.
        </div>
      )}

      <CafeBackgroundEditor cafeId={cafe.id} backgroundUrl={cafe.background_image_url} />

      <CafeTagEditor cafeId={cafe.id} initialTagIds={(cafe.tags ?? []).map((t) => t.id)} />

      <ReviewSection cafe={cafe} reviews={reviews} />
    </section>
  )
}
