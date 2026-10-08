import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getCafe, listReviews } from '../services/api.js'
import { useAuth } from '../context/AuthContext.jsx'
import Stars from '../components/Stars.jsx'
import ReviewCard from '../components/ReviewCard.jsx'
import ReviewForm from '../components/ReviewForm.jsx'

export default function CafeDetailPage() {
  const { id } = useParams()
  const { user } = useAuth()
  const [cafe, setCafe] = useState(null)
  const [reviews, setReviews] = useState([])
  const [error, setError] = useState('')

  // Reload the cafe (for the updated average) and its reviews.
  const reload = useCallback(() => {
    Promise.all([getCafe(id), listReviews(id)])
      .then(([cafeData, reviewData]) => {
        setCafe(cafeData)
        setReviews(reviewData)
      })
      .catch((err) => setError(err.message))
  }, [id])

  useEffect(() => {
    reload()
  }, [reload])

  if (error) return <p className="error">{error}</p>
  if (!cafe) return <p className="status">Loading...</p>

  const myReview = user ? reviews.find((r) => r.user_id === user.id) : null

  return (
    <section>
      <Link to="/">&larr; All cafes</Link>
      <h1>{cafe.name}</h1>
      {cafe.address && <p className="muted">{cafe.address}</p>}
      {cafe.description && <p>{cafe.description}</p>}
      <p>
        <Stars value={cafe.avg_rating} />{' '}
        <span className="muted">
          {cafe.review_count === 0
            ? 'No reviews yet'
            : `${cafe.avg_rating.toFixed(1)} from ${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'}`}
        </span>
      </p>

      {user ? (
        <ReviewForm
          key={myReview?.id ?? 'new'}
          cafeId={cafe.id}
          existing={myReview}
          onSaved={reload}
        />
      ) : (
        <p>
          <Link to="/login" state={{ from: `/cafes/${id}` }}>Log in</Link> to write a review.
        </p>
      )}

      <h2>Reviews</h2>
      {reviews.length === 0 ? (
        <p className="muted">No reviews yet.</p>
      ) : (
        reviews.map((review) => (
          <ReviewCard
            key={review.id}
            review={review}
            currentUserId={user?.id}
            onDeleted={reload}
          />
        ))
      )}
    </section>
  )
}
