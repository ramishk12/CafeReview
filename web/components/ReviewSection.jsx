'use client'

import Link from 'next/link'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { listTags } from '@/lib/api'
import PhotoLightbox from './PhotoLightbox'
import ReviewCard from './ReviewCard'
import ReviewForm from './ReviewForm'

const ORDERS = {
  newest: { label: 'Newest', compare: (a, b) => new Date(b.created_at) - new Date(a.created_at) },
  visit: {
    label: 'Latest visit',
    compare: (a, b) => String(b.visited_on || '').localeCompare(String(a.visited_on || '')) || new Date(b.created_at) - new Date(a.created_at),
  },
  highest: { label: 'Highest rated', compare: (a, b) => b.rating - a.rating || new Date(b.created_at) - new Date(a.created_at) },
  lowest: { label: 'Lowest rated', compare: (a, b) => a.rating - b.rating || new Date(b.created_at) - new Date(a.created_at) },
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

// Review list, sorting, filtering, photo gallery, and the review form. Filters
// (minimum rating, has photos, review tags, mine) run in the browser and are kept
// in the URL so the view is shareable.
export default function ReviewSection({ cafe, reviews }) {
  const { user } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const closed = Boolean(cafe.closed)
  const [order, setOrder] = useState('newest')
  const [editingId, setEditingId] = useState(null)
  const [lightbox, setLightbox] = useState(null) // { images, index } or null
  const [reviewTags, setReviewTags] = useState([])

  useEffect(() => {
    listTags('review')
      .then(setReviewTags)
      .catch(() => {})
  }, [])

  const minRating = Number(searchParams.get('rating')) || 0
  const onlyPhotos = searchParams.get('photos') === '1'
  const onlyMine = searchParams.get('mine') === '1'
  const tagParam = searchParams.get('tags') || ''
  const filterTagIds = useMemo(() => tagParam.split(',').filter(Boolean).map(Number), [tagParam])

  const setParam = (updates) => {
    const next = new URLSearchParams(searchParams.toString())
    for (const [key, value] of Object.entries(updates)) {
      if (value === undefined || value === null || value === '' || value === false) next.delete(key)
      else next.set(key, String(value))
    }
    const qs = next.toString()
    router.replace(qs ? `${pathname}?${qs}` : pathname)
  }

  const sorted = useMemo(() => [...reviews].sort(ORDERS[order].compare), [reviews, order])
  const visible = useMemo(
    () =>
      sorted.filter((review) => {
        if (minRating && review.rating < minRating) return false
        if (onlyPhotos && !(review.images?.length > 0)) return false
        if (onlyMine && user && review.user_id !== user.id) return false
        if (filterTagIds.length > 0 && !(review.tags || []).some((t) => filterTagIds.includes(t.id))) return false
        return true
      }),
    [sorted, minRating, onlyPhotos, onlyMine, user, filterTagIds],
  )

  // Every photo for this cafe, in review order, for the gallery and the lightbox.
  const photos = useMemo(() => sorted.flatMap(reviewPhotos), [sorted])

  const openPhoto = (images, index) => setLightbox({ images, index })
  const setLightboxIndex = (index) => setLightbox((l) => (l ? { ...l, index } : l))
  const hasFilters = minRating > 0 || onlyPhotos || onlyMine || filterTagIds.length > 0

  return (
    <>
      {closed ? (
        <div className="card mb-10 p-6 text-slate-600 dark:text-slate-300">
          This cafe is closed, so it is no longer accepting new reviews.
        </div>
      ) : user ? (
        editingId === null && (
          <div className="mb-10">
            <ReviewForm key="new" cafeId={cafe.id} />
          </div>
        )
      ) : (
        <div className="card mb-10 flex flex-wrap items-center justify-between gap-4 p-6">
          <p className="text-slate-600 dark:text-slate-300">Share your experience with {cafe.name}.</p>
          <Link href={`/login?from=/cafes/${cafe.id}`} className="btn">
            Log in to review
          </Link>
        </div>
      )}

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
        {reviews.length > 1 && (
          <label className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
            Sort
            <select value={order} onChange={(e) => setOrder(e.target.value)} className="field w-auto py-2 pr-8">
              {Object.entries(ORDERS).map(([key, o]) => (
                <option key={key} value={key}>
                  {o.label}
                </option>
              ))}
            </select>
          </label>
        )}
      </div>

      {reviews.length > 0 && (
        <div className="mb-6 flex flex-col gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <label className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
              Rating
              <select
                value={minRating || ''}
                onChange={(e) => setParam({ rating: e.target.value === '' ? undefined : e.target.value })}
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
              onClick={() => setParam({ photos: onlyPhotos ? undefined : '1' })}
              aria-pressed={onlyPhotos}
              className={chipClass(onlyPhotos)}
            >
              With photos
            </button>
            {user && (
              <button
                type="button"
                onClick={() => setParam({ mine: onlyMine ? undefined : '1' })}
                aria-pressed={onlyMine}
                className={chipClass(onlyMine)}
              >
                My reviews
              </button>
            )}
            {hasFilters && (
              <button
                type="button"
                onClick={() => setParam({ rating: undefined, photos: undefined, mine: undefined, tags: undefined })}
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
                    onClick={() => setParam({ tags: ids.length ? ids.join(',') : undefined })}
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
        <p className="card py-12 text-center text-slate-500">No reviews yet. Be the first.</p>
      ) : visible.length === 0 ? (
        <p className="card py-12 text-center text-slate-500">No reviews match your filters.</p>
      ) : (
        <ul className="grid gap-4">
          {visible.map((review) => (
            <li key={review.id}>
              {editingId === review.id ? (
                <ReviewForm cafeId={cafe.id} existing={review} onDone={() => setEditingId(null)} />
              ) : (
                <ReviewCard
                  review={review}
                  isOwner={user?.id === review.user_id}
                  onOpenPhoto={(index) => openPhoto(reviewPhotos(review), index)}
                  onEdit={closed ? undefined : () => setEditingId(review.id)}
                />
              )}
            </li>
          ))}
        </ul>
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
