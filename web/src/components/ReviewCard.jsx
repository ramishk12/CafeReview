import { useState } from 'react'
import { deleteReview } from '../services/api.js'
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
    <article className="card mb-4 p-5">
      <header className="flex flex-wrap items-center gap-3">
        <strong className="font-semibold">{review.author_name}</strong>
        <Stars value={review.rating} />
        <time dateTime={review.created_at} className="ml-auto text-sm text-slate-500 dark:text-slate-400">
          {new Date(review.created_at).toLocaleDateString()}
        </time>
      </header>

      <p className="my-3 whitespace-pre-wrap leading-relaxed">{review.body}</p>

      {review.images?.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-2.5">
          {review.images.map((img) => (
            <a key={img.id} href={img.url} target="_blank" rel="noreferrer" className="block overflow-hidden rounded-xl">
              <img
                src={img.url}
                alt={`Photo by ${review.author_name}`}
                loading="lazy"
                className="h-32 w-32 border border-slate-200 object-cover transition duration-200 hover:scale-105 dark:border-slate-800"
              />
            </a>
          ))}
        </div>
      )}

      {isOwner && (
        <div className="mt-3">
          <button type="button" onClick={handleDelete} className="btn btn-danger px-3 py-1.5 text-sm">
            Delete
          </button>
        </div>
      )}
      {error && <p className="alert-error mt-3">{error}</p>}
    </article>
  )
}
