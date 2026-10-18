'use client'

import { Pencil, ThumbsUp, Trash2 } from 'lucide-react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { deleteReview, deleteReviewImage, getReviewRevisions } from '@/lib/api'
import { formatDate } from '@/lib/cafeTheme'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import AddToListButton from './AddToListButton'
import Avatar from './Avatar'
import ReviewImageTagger from './ReviewImageTagger'
import Stars from './Stars'

export default function ReviewCard({ review, isOwner, voted = false, helpfulCount = 0, onVote, onOpenPhoto, onEdit }) {
  const router = useRouter()
  const toast = useToast()
  const { user, isAdmin } = useAuth()
  const [deleting, setDeleting] = useState(false)
  const [removingPhotoId, setRemovingPhotoId] = useState(null)
  const [error, setError] = useState('')
  const [history, setHistory] = useState(null)

  const canSeeHistory = review.revision_count > 0 && (isOwner || isAdmin)

  const toggleHistory = async () => {
    if (history) {
      setHistory(null)
      return
    }
    try {
      setHistory(await getReviewRevisions(review.id))
    } catch (err) {
      setError(err.message)
    }
  }

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

  const handleRemovePhoto = async (imageId) => {
    if (!window.confirm('Remove this photo? This cannot be undone.')) return
    setRemovingPhotoId(imageId)
    setError('')
    try {
      await deleteReviewImage(review.id, imageId)
      toast.show('Photo removed')
      router.refresh()
    } catch (err) {
      setError(err.message)
    } finally {
      setRemovingPhotoId(null)
    }
  }

  return (
    <article className="card p-6">
      <header className="flex items-center gap-3">
        <Avatar name={review.author_name} />
        <div className="min-w-0 flex-1">
          <Link href={`/users/${review.user_id}`} className="font-semibold leading-tight no-underline hover:underline">
            {review.author_name}
          </Link>
          <time dateTime={review.visited_on || review.created_at} className="text-xs text-slate-500 dark:text-slate-400">
            {review.visited_on ? `Visited ${formatDate(review.visited_on)}` : formatDate(review.created_at)}
          </time>
        </div>
        <Stars value={review.rating} size={15} />
      </header>

      <p className="mt-4 whitespace-pre-wrap leading-relaxed text-slate-700 dark:text-slate-300">{review.body}</p>

      {review.tags?.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-2">
          {review.tags.map((tag) => (
            <span
              key={tag.id}
              className="rounded-full bg-slate-100 px-2.5 py-0.5 text-xs font-medium text-slate-600 dark:bg-white/10 dark:text-slate-300"
            >
              {tag.name}
            </span>
          ))}
        </div>
      )}

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
              {(isOwner || isAdmin) && <ReviewImageTagger image={img} />}
              {isOwner && (
                <button
                  type="button"
                  onClick={() => handleRemovePhoto(img.id)}
                  disabled={removingPhotoId === img.id}
                  className="mt-1 text-xs font-semibold text-rose-600 hover:text-rose-500 disabled:opacity-50 dark:text-rose-400"
                >
                  {removingPhotoId === img.id ? 'Removing…' : 'Remove photo'}
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      {canSeeHistory && (
        <div className="mt-3">
          <button type="button" onClick={toggleHistory} className="text-xs font-semibold text-indigo-600 dark:text-indigo-400">
            {history ? 'Hide history' : `Edited · view history (${review.revision_count})`}
          </button>
          {history && (
            <ul className="mt-2 grid gap-2">
              {history.map((revision) => (
                <li key={revision.id} className="rounded-xl bg-slate-50 p-3 text-sm dark:bg-white/5">
                  <div className="flex items-center justify-between gap-2">
                    <Stars value={revision.rating} size={13} />
                    <span className="text-xs text-slate-400">{formatDate(revision.visited_on)}</span>
                  </div>
                  <p className="mt-1 whitespace-pre-wrap text-slate-600 dark:text-slate-300">{revision.body}</p>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {user && (
        <div className="mt-4 flex flex-wrap items-center gap-2">
          {!isOwner && (
            <button
              type="button"
              onClick={onVote}
              aria-pressed={voted}
              className={`inline-flex items-center gap-1.5 rounded-full px-3.5 py-2 text-sm transition ${
                voted
                  ? 'bg-indigo-600 text-white'
                  : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-white/10 dark:text-slate-200 dark:hover:bg-white/20'
              }`}
            >
              <ThumbsUp size={14} aria-hidden="true" /> Helpful{helpfulCount > 0 ? ` (${helpfulCount})` : ''}
            </button>
          )}
          <AddToListButton type="review" id={review.id} label="Save review" />
          <AddToListButton type="user" id={review.user_id} label="Save reviewer" />
        </div>
      )}

      {isOwner && (
        <div className="mt-4 flex items-center gap-3 border-t border-slate-200/70 pt-4 dark:border-white/5">
          {onEdit && (
            <button type="button" onClick={onEdit} className="btn btn-ghost px-3.5 py-2 text-sm">
              <Pencil size={15} aria-hidden="true" /> Edit
            </button>
          )}
          <button type="button" onClick={handleDelete} disabled={deleting} className="btn btn-danger px-3.5 py-2 text-sm">
            <Trash2 size={15} aria-hidden="true" /> {deleting ? 'Deleting…' : 'Delete'}
          </button>
        </div>
      )}
      {error && <p className="alert-error mt-3">{error}</p>}
    </article>
  )
}
