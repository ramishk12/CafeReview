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
    <article className="review-card">
      <header className="review-header">
        <strong>{review.author_name}</strong>
        <Stars value={review.rating} />
        <time dateTime={review.created_at}>{new Date(review.created_at).toLocaleDateString()}</time>
      </header>
      <p className="review-body">{review.body}</p>

      {review.images?.length > 0 && (
        <div className="review-images">
          {review.images.map((img) => (
            <a key={img.id} href={img.url} target="_blank" rel="noreferrer">
              <img src={img.url} alt={`Photo by ${review.author_name}`} loading="lazy" />
            </a>
          ))}
        </div>
      )}

      {isOwner && (
        <div className="review-actions">
          <button type="button" className="danger" onClick={handleDelete}>
            Delete
          </button>
        </div>
      )}
      {error && <p className="error">{error}</p>}
    </article>
  )
}
