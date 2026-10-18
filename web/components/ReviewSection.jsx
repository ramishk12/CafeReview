'use client'

import { MessageSquare } from 'lucide-react'
import Link from 'next/link'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import { listMyReviewVotes, listTags, unvoteReview, voteReview } from '@/lib/api'
import EmptyState from './EmptyState'
import PhotoLightbox from './PhotoLightbox'
import ReviewCard from './ReviewCard'
import ReviewForm from './ReviewForm'

// SORTS must match reviewSortClauses in internal/handlers/reviews.go.
const SORTS = {
  newest: 'Newest',
  visit: 'Latest visit',
  helpful: 'Most helpful',
  highest: 'Highest rated',
  lowest: 'Lowest rated',
}
const PREVIEW_PHOTOS = 12

const chipClass = (active) =>
  `rounded-full px-3 py-1 text-sm transition ${
    active
      ? 'bg-indigo-600 text-white'
      : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-white/10 dark:text-slate-200 dark:hover:bg-white/20'
  }`

// Maps a review's images to the shape PhotoLightbox expects.
export function reviewPhotos(review) {
  return (review.images || []).map((img) => ({ id: img.id, url: img.url, alt: `Photo by ${review.author_name}` }))
}

// Review list, sorting, filtering, photo gallery, and the review form. The list
// is sorted, filtered, and paged on the server (see `listReviews`); the controls
// write to the URL, so views are shareable and the back button works.
export default function ReviewSection({ cafe, reviews, total = 0, page = 1, pageSize = 10 }) {
  const { user } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const closed = Boolean(cafe.closed)
  const [editingId, setEditingId] = useState(null)
  const [lightbox, setLightbox] = useState(null) // { images, index } or null
  const [reviewTags, setReviewTags] = useState([])
  const [votes, setVotes] = useState({})

  const sortParam = searchParams.get('sort')
  const sort = SORTS[sortParam] ? sortParam : 'newest'
  const minRating = Number(searchParams.get('rating')) || 0
  const onlyPhotos = searchParams.get('photos') === '1'
  const tagParam = searchParams.get('tags') || ''
  const filterTagIds = useMemo(() => tagParam.split(',').filter(Boolean).map(Number), [tagParam])
  const onlyMine = Boolean(user) && searchParams.get('user_id') === String(user.id)

  useEffect(() => {
    listTags('review')
      .then(setReviewTags)
      .catch(() => {})
  }, [])

  // Helpful counts start from the review payload; the signed-in user's own
  // votes are filled in from a separate call.
  useEffect(() => {
    const initial = {}
    for (const review of reviews) initial[review.id] = { voted: false, count: review.helpful_count || 0 }
    setVotes(initial)
  }, [reviews])

  useEffect(() => {
    if (!user) return
    listMyReviewVotes(cafe.id)
      .then((ids) =>
        setVotes((current) => {
          const next = { ...current }
          for (const id of ids) if (next[id]) next[id] = { ...next[id], voted: true }
          return next
        }),
      )
      .catch(() => {})
  }, [user, cafe.id, reviews])

  const toggleVote = async (review) => {
    const current = votes[review.id] || { voted: false, count: review.helpful_count || 0 }
    try {
      const result = current.voted ? await unvoteReview(review.id) : await voteReview(review.id)
      setVotes((v) => ({ ...v, [review.id]: { voted: result.voted, count: result.helpful_count } }))
    } catch (err) {
      toast.show(err.message, 'error')
    }
  }

  // navigate keeps the current view in the URL. Changing sort or a filter resets
  // paging, because it changes the result set.
  const navigate = (updates) => {
    const next = new URLSearchParams(searchParams.toString())
    for (const [key, value] of Object.entries(updates)) {
      if (value === undefined || value === null || value === '' || value === false) next.delete(key)
      else next.set(key, String(value))
    }
    const qs = next.toString()
    router.push(qs ? `${pathname}?${qs}` : pathname)
  }
  const apply = (updates) => navigate({ ...updates, page: undefined })

  const hasFilters = minRating > 0 || onlyPhotos || onlyMine || filterTagIds.length > 0
  const photos = useMemo(() => reviews.flatMap(reviewPhotos), [reviews])
  const openPhoto = (images, index) => setLightbox({ images, index })
  const setLightboxIndex = (index) => setLightbox((l) => (l ? { ...l, index } : l))

  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  let start = Math.max(1, page - 2)
  let end = Math.min(totalPages, start + 4)
  start = Math.max(1, end - 4)
  const pages = []
  for (let p = start; p <= end; p++) pages.push(p)

  return (
    <>
      {closed ? (
        <div className="card mb-10 p-6 text-slate-600 dark:text-slate-300">
          This cafe is closed, so it is no longer accepting new reviews.
        </div>
      ) : user ? (
        editingId === null && (
          <div id="write-a-review" className="mb-10">
            <ReviewForm key="new" cafeId={cafe.id} />
          </div>
        )
      ) : reviews.length > 0 || total > 0 ? (
        <div className="card mb-10 flex flex-wrap items-center justify-between gap-4 p-6">
          <p className="text-slate-600 dark:text-slate-300">Share your experience with {cafe.name}.</p>
          <Link href={`/login?from=/cafes/${cafe.id}`} className="btn">
            Log in to review
          </Link>
        </div>
      ) : null}

      {photos.length > 0 && (
        <section aria-labelledby="photos-heading" className="mb-10">
          <h2 id="photos-heading" className="mb-4 text-xl font-bold tracking-tight">
            Photos <span className="text-base font-medium text-slate-400">({photos.length})</span>
          </h2>
          <ul className="grid grid-cols-3 gap-2 sm:grid-cols-4 md:grid-cols-6">
            {photos.slice(0, PREVIEW_PHOTOS).map((p, i) => (
              <li key={p.id}>
                <button
                  type="button"
                  onClick={() => openPhoto(photos, i)}
                  className="block aspect-square w-full cursor-zoom-in overflow-hidden rounded-xl"
                  aria-label={`Open ${p.alt.toLowerCase()}`}
                >
                  <img src={p.url} alt="" loading="lazy" className="h-full w-full object-cover transition duration-300 hover:scale-105" />
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-2xl font-bold tracking-tight">Reviews</h2>
        {total > 1 && (
          <label className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
            Sort
            <select value={sort} onChange={(e) => apply({ sort: e.target.value })} className="field w-auto py-2 pr-8">
              {Object.entries(SORTS).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
            </select>
          </label>
        )}
      </div>

      {(hasFilters || total > 0) && (
        <div className="mb-6 flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <label className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
              Rating
              <select
                value={minRating || ''}
                onChange={(e) => apply({ rating: e.target.value === '' ? undefined : e.target.value })}
                className="field w-auto py-2 pr-8"
              >
                <option value="">Any</option>
                <option value="4">4+ stars</option>
                <option value="3">3+ stars</option>
                <option value="2">2+ stars</option>
              </select>
            </label>
            <button
              type="button"
              onClick={() => apply({ photos: onlyPhotos ? undefined : '1' })}
              aria-pressed={onlyPhotos}
              className={chipClass(onlyPhotos)}
            >
              With photos
            </button>
            {user && (
              <button
                type="button"
                onClick={() => apply({ user_id: onlyMine ? undefined : user.id })}
                aria-pressed={onlyMine}
                className={chipClass(onlyMine)}
              >
                My reviews
              </button>
            )}
            {hasFilters && (
              <button
                type="button"
                onClick={() => apply({ rating: undefined, photos: undefined, user_id: undefined, tags: undefined })}
                className="text-sm font-semibold text-indigo-600 dark:text-indigo-400"
              >
                Clear filters
              </button>
            )}
          </div>

          {reviewTags.length > 0 && (
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-semibold tracking-wide text-slate-400 uppercase">Tags</span>
              {reviewTags.map((tag) => {
                const active = filterTagIds.includes(tag.id)
                const ids = active ? filterTagIds.filter((t) => t !== tag.id) : [...filterTagIds, tag.id]
                return (
                  <button
                    key={tag.id}
                    type="button"
                    onClick={() => apply({ tags: ids.length ? ids.join(',') : undefined })}
                    aria-pressed={active}
                    className={chipClass(active)}
                  >
                    {tag.name}
                  </button>
                )
              })}
            </div>
          )}
        </div>
      )}

      {reviews.length === 0 ? (
        hasFilters ? (
          <EmptyState
            title="No reviews match your filters"
            description="Try a lower minimum rating or remove a filter."
            action={
              <button
                type="button"
                onClick={() => apply({ rating: undefined, photos: undefined, user_id: undefined, tags: undefined })}
                className="btn"
              >
                Show all reviews
              </button>
            }
          />
        ) : (
          <EmptyState
            icon={MessageSquare}
            title="No reviews yet"
            description={
              closed
                ? `${cafe.name} is closed, so it isn't accepting new reviews.`
                : `Be the first to share your experience with ${cafe.name}.`
            }
            action={
              closed ? null : user ? (
                <a href="#write-a-review" className="btn">
                  Write a review
                </a>
              ) : (
                <Link href={`/login?from=/cafes/${cafe.id}`} className="btn">
                  Log in to review
                </Link>
              )
            }
          />
        )
      ) : (
        <ul className="grid gap-4">
          {reviews.map((review) => (
            <li key={review.id}>
              {editingId === review.id ? (
                <ReviewForm cafeId={cafe.id} existing={review} onDone={() => setEditingId(null)} />
              ) : (
                <ReviewCard
                  review={review}
                  isOwner={user?.id === review.user_id}
                  voted={votes[review.id]?.voted ?? false}
                  helpfulCount={votes[review.id]?.count ?? review.helpful_count ?? 0}
                  onVote={() => toggleVote(review)}
                  onOpenPhoto={(index) => openPhoto(reviewPhotos(review), index)}
                  onEdit={closed ? undefined : () => setEditingId(review.id)}
                />
              )}
            </li>
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <nav className="mt-10 flex items-center justify-center gap-2" aria-label="Pagination">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => navigate({ page: page - 1 === 1 ? undefined : page - 1 })}
            className="btn btn-ghost px-3.5 py-2 text-sm"
          >
            Previous
          </button>
          {pages.map((p) => (
            <button
              key={p}
              type="button"
              onClick={() => navigate({ page: p === 1 ? undefined : p })}
              aria-current={p === page ? 'page' : undefined}
              className={`rounded-full px-3.5 py-1.5 text-sm ${
                p === page ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200'
              }`}
            >
              {p}
            </button>
          ))}
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => navigate({ page: page + 1 })}
            className="btn btn-ghost px-3.5 py-2 text-sm"
          >
            Next
          </button>
        </nav>
      )}

      <PhotoLightbox
        images={lightbox?.images ?? []}
        index={lightbox?.index ?? null}
        onIndexChange={setLightboxIndex}
        onClose={() => setLightbox(null)}
      />
    </>
  )
}
