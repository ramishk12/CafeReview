import { Trash2 } from 'lucide-react'
import { useState } from 'react'
import { deleteReview } from '../services/api.js'
import Avatar from './Avatar.jsx'
import Stars from './Stars.jsx'

export default function ReviewCard({ review, currentUserId, onDeleted }) {
  const [error, setError] = useState('')
  const isOwner = currentUserId === review.user_id

  const handleDelete = async () => {
    if (!window.confirm('Delete this review?')) return
    setError('')
    try {
      await deleteReview(review.id)
      onDeleted()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <article className="card mb-4 p-6">
      <header className="flex items-center gap-3">
        <Avatar name={review.author_name} />
        <div className="min-w-0 flex-1">
          <p className="font-semibold leading-tight">{review.author_name}</p>
          <time dateTime={review.created_at} className="text-xs text-slate-500 dark:text-slate-400">
            {new Date(review.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
          </time>
        </div>
        <Stars value={review.rating} size={15} />
      </header>

      <p className="mt-4 whitespace-pre-wrap leading-relaxed text-slate-700 dark:text-slate-300">{review.body}</p>

      {review.images?.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-3">
          {review.images.map((img) => (
            <a key={img.id} href={img.url} target="_blank" rel="noreferrer" className="block overflow-hidden rounded-2xl ring-1 ring-slate-900/5">
              <img
                src={img.url}
                alt={`Photo by ${review.author_name}`}
                loading="lazy"
                className="h-36 w-36 object-cover transition duration-300 hover:scale-105"
              />
            </a>
          ))}
        </div>
      )}

      {isOwner && (
        <div className="mt-4 border-t border-slate-200/70 pt-4 dark:border-white/5">
          <button type="button" onClick={handleDelete} className="btn btn-danger px-3.5 py-2 text-sm">
            <Trash2 size={15} /> Delete
          </button>
        </div>
      )}
      {error && <p className="alert-error mt-3">{error}</p>}
    </article>
  )
}
