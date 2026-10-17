import { ArrowLeft, MapPin } from 'lucide-react'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import AddToListButton from '@/components/AddToListButton'
import CafeAdminActions from '@/components/CafeAdminActions'
import CafeBackgroundEditor from '@/components/CafeBackgroundEditor'
import CafeTagEditor from '@/components/CafeTagEditor'
import ReviewSection from '@/components/ReviewSection'
import Stars from '@/components/Stars'
import { getCafe, listReviews } from '@/lib/serverApi'
import { cafeGradient, ratingLabel } from '@/lib/cafeTheme'

// REVIEW_SORTS must match reviewSortClauses in internal/handlers/reviews.go.
const REVIEW_SORTS = ['newest', 'visit', 'helpful', 'highest', 'lowest']
const REVIEW_RATINGS = ['1', '2', '3', '4', '5']
const REVIEW_PAGE_SIZE = 10

export async function generateMetadata({ params }) {
  const { id } = await params
  const cafe = await getCafe(id).catch(() => null)
  if (!cafe) return { title: 'Cafe not found' }
  return {
    title: cafe.name,
    description: cafe.description || `Reviews and photos for ${cafe.name}.`,
  }
}

export default async function CafePage({ params, searchParams }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()

  const cafe = await getCafe(id)
  if (!cafe) notFound()

  // The review list is server-sorted, filtered, and paged; the controls write
  // these values to the URL so the view stays shareable.
  const sp = await searchParams
  const sort = REVIEW_SORTS.includes(sp.sort) ? sp.sort : 'newest'
  const page = Math.max(1, Number(sp.page) || 1)
  const rating = REVIEW_RATINGS.includes(String(sp.rating)) ? String(sp.rating) : ''
  const photos = sp.photos === '1'
  const tags = typeof sp.tags === 'string' ? sp.tags : ''
  const userId = /^\d+$/.test(String(sp.user_id ?? '')) ? String(sp.user_id) : ''

  const { reviews, total } = await listReviews(id, {
    sort,
    rating,
    photos,
    tags,
    userId,
    page,
    pageSize: REVIEW_PAGE_SIZE,
  })

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
        <div className="flex items-center gap-2">
          <AddToListButton type="cafe" id={cafe.id} />
          <CafeAdminActions cafeId={cafe.id} cafeName={cafe.name} closed={cafe.closed} />
        </div>
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

      <CafeBackgroundEditor
        cafeId={cafe.id}
        backgroundUrl={cafe.background_image_url}
        backgroundSource={cafe.background_source}
        images={reviews.flatMap((r) => r.images ?? [])}
      />

      <CafeTagEditor cafeId={cafe.id} initialTagIds={(cafe.tags ?? []).map((t) => t.id)} />

      <ReviewSection cafe={cafe} reviews={reviews} total={total} page={page} pageSize={REVIEW_PAGE_SIZE} />
    </section>
  )
}
