import { useState } from 'react'
import { createReview, updateReview, uploadReviewImage } from '../services/api.js'

const ACCEPTED_TYPES = 'image/jpeg,image/png,image/webp'
const MAX_IMAGE_BYTES = 5 * 1024 * 1024

// ReviewForm creates a review, or edits one when `existing` is passed.
// Selected photos are uploaded one at a time after the review is saved.
export default function ReviewForm({ cafeId, existing, onSaved }) {
  const [rating, setRating] = useState(existing?.rating ?? 5)
  const [body, setBody] = useState(existing?.body ?? '')
  const [files, setFiles] = useState([])
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleFiles = (e) => {
    const selected = Array.from(e.target.files || [])
    const tooBig = selected.find((f) => f.size > MAX_IMAGE_BYTES)
    setError(tooBig ? `${tooBig.name} is larger than 5 MB` : '')
    setFiles(tooBig ? [] : selected)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)

    let review
    try {
      const payload = { rating: Number(rating), body }
      review = existing
        ? await updateReview(existing.id, payload)
        : await createReview(cafeId, payload)
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
      return
    }

    const failed = []
    for (const file of files) {
      try {
        await uploadReviewImage(review.id, file)
      } catch (err) {
        failed.push(`${file.name}: ${err.message}`)
      }
    }

    setSubmitting(false)
    setFiles([])
    e.target.reset()
    if (failed.length > 0) setError(`Review saved, but some photos failed. ${failed.join('; ')}`)
    onSaved(review)
  }

  return (
    <form className="review-form" onSubmit={handleSubmit}>
      <h3>{existing ? 'Edit your review' : 'Write a review'}</h3>

      <label>
        Rating
        <select value={rating} onChange={(e) => setRating(e.target.value)}>
          {[5, 4, 3, 2, 1].map((n) => (
            <option key={n} value={n}>
              {n} {n === 1 ? 'star' : 'stars'}
            </option>
          ))}
        </select>
      </label>

      <label>
        Review
        <textarea
          value={body}
          onChange={(e) => setBody(e.target.value)}
          maxLength={2000}
          rows={4}
          required
        />
      </label>

      <label>
        Photos (JPEG, PNG, or WebP, up to 5 MB each)
        <input type="file" accept={ACCEPTED_TYPES} multiple onChange={handleFiles} />
      </label>

      {error && <p className="error">{error}</p>}

      <button type="submit" disabled={submitting}>
        {submitting ? 'Saving...' : existing ? 'Save changes' : 'Post review'}
      </button>
    </form>
  )
}
