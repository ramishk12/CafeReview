'use client'

import Link from 'next/link'
import { useMemo, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import PhotoLightbox from './PhotoLightbox'
import ReviewCard from './ReviewCard'
import ReviewForm from './ReviewForm'

const ORDERS = {
  newest: { label: 'Newest', compare: (a, b) => new Date(b.created_at) - new Date(a.created_at) },
  highest: { label: 'Highest rated', compare: (a, b) => b.rating - a.rating || new Date(b.created_at) - new Date(a.created_at) },
  lowest: { label: 'Lowest rated', compare: (a, b) => a.rating - b.rating || new Date(b.created_at) - new Date(a.created_at) },
}

// Review list, sorting, photo gallery, and the review form for the signed-in user.
export default function ReviewSection({ cafe, reviews }) {
  const { user } = useAuth()
  const [order, setOrder] = useState('newest')
  const [lightbox, setLightbox] = useState(null) // { images, index } or null

  const mine = user ? reviews.find((r) => r.user_id === user.id) : null
  const sorted = useMemo(() => [...reviews].sort(ORDERS[order].compare), [reviews, order])

  // Every photo for this cafe, in review order, for the gallery and the lightbox.
  const photos = useMemo(
    () =>
      sorted.flatMap((r) => (r.images || []).map((img) => ({ id: img.id, url: img.url, alt: `Photo by ${r.author_name}` }))),
    [sorted],
  )

  const toPhotos = (review) =>
    (review.images || []).map((img) => ({ id: img.id, url: img.url, alt: `Photo by ${review.author_name}` }))
  const openPhoto = (images, index) => setLightbox({ images, index })
  const setLightboxIndex = (index) => setLightbox((l) => (l ? { ...l, index } : l))

  return (
    <>
      {user ? (
        <div className="mb-10">
          <ReviewForm key={mine?.id ?? 'new'} cafeId={cafe.id} existing={mine} />
        </div>
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
            {photos.slice(0, 12).map((p, i) => (
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

      {sorted.length === 0 ? (
        <p className="card py-12 text-center text-slate-500">No reviews yet. Be the first.</p>
      ) : (
        <ul className="grid gap-4">
          {sorted.map((review) => (
            <li key={review.id}>
              <ReviewCard
                review={review}
                isOwner={user?.id === review.user_id}
                onOpenPhoto={(index) => openPhoto(toPhotos(review), index)}
              />
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
