'use client'

import { Trash2 } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { deleteReview } from '@/lib/api'
import { formatDate } from '@/lib/cafeTheme'
import { useToast } from '@/context/ToastContext'
import Avatar from './Avatar'
import Stars from './Stars'

export default function ReviewCard({ review, isOwner, onOpenPhoto }) {
  const router = useRouter()
  const toast = useToast()
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState('')

  const handleDelete = async () => {
    if (!window.confirm('Delete this review? This also removes its photos.')) return
    setDeleting(true)
    setError('')
    try {
      await deleteReview(review.id)
      toast.show('Review deleted')
      router.refresh()
    } catch (err) {
      setError(err.message)
      setDeleting(false)
    }
  }

  return (
    <article className="card p-6">
      <header className="flex items-center gap-3">
        <Avatar name={review.author_name} />
        <div className="min-w-0 flex-1">
          <p className="font-semibold leading-tight">{review.author_name}</p>
          <time dateTime={review.created_at} className="text-xs text-slate-500 dark:text-slate-400">
            {formatDate(review.created_at)}
          </time>
        </div>
        <Stars value={review.rating} size={15} />
      </header>

      <p className="mt-4 whitespace-pre-wrap leading-relaxed text-slate-700 dark:text-slate-300">{review.body}</p>

      {review.images?.length > 0 && (
        <ul className="mt-4 flex flex-wrap gap-3">
          {review.images.map((img, i) => (
            <li key={img.id}>
              <button
                type="button"
                onClick={() => onOpenPhoto(i)}
                className="block cursor-zoom-in overflow-hidden rounded-2xl ring-1 ring-slate-900/5"
                aria-label={`Open photo by ${review.author_name}`}
              >
                <img
                  src={img.url}
                  alt={`Photo by ${review.author_name}`}
                  loading="lazy"
                  className="h-32 w-32 object-cover transition duration-300 hover:scale-105 sm:h-36 sm:w-36"
                />
              </button>
            </li>
          ))}
        </ul>
      )}

      {isOwner && (
        <div className="mt-4 flex items-center gap-3 border-t border-slate-200/70 pt-4 dark:border-white/5">
          <button type="button" onClick={handleDelete} disabled={deleting} className="btn btn-danger px-3.5 py-2 text-sm">
            <Trash2 size={15} aria-hidden="true" /> {deleting ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      )}
      {error && <p className="alert-error mt-3">{error}</p>}
    </article>
  )
}
