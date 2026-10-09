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

  if (error) return <p className="alert-error">{error}</p>
  if (!cafe) return <p className="py-12 text-center text-slate-500">Loading...</p>

  const myReview = user ? reviews.find((r) => r.user_id === user.id) : null

  return (
    <section>
      <Link to="/" className="text-sm font-medium no-underline hover:underline">
        &larr; All cafes
      </Link>

      <header className="mt-3 mb-6">
        <h1 className="text-3xl font-bold tracking-tight">{cafe.name}</h1>
        {cafe.address && <p className="text-slate-500 dark:text-slate-400">{cafe.address}</p>}
        {cafe.description && <p className="mt-2 max-w-2xl">{cafe.description}</p>}
        <div className="mt-3 flex items-center gap-2 text-sm">
          <Stars value={cafe.avg_rating} />
          <span className="text-slate-500 dark:text-slate-400">
            {cafe.review_count === 0
              ? 'No reviews yet'
              : `${cafe.avg_rating.toFixed(1)} from ${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'}`}
          </span>
        </div>
      </header>

      {user ? (
        <ReviewForm key={myReview?.id ?? 'new'} cafeId={cafe.id} existing={myReview} onSaved={reload} />
      ) : (
        <p className="card mb-8 p-5 text-slate-600 dark:text-slate-300">
          <Link to="/login" state={{ from: `/cafes/${id}` }} className="font-semibold">
            Log in
          </Link>{' '}
          to write a review.
        </p>
      )}

      <h2 className="mb-4 text-xl font-semibold tracking-tight">Reviews</h2>
      {reviews.length === 0 ? (
        <p className="text-slate-500 dark:text-slate-400">No reviews yet.</p>
      ) : (
        reviews.map((review) => (
          <ReviewCard key={review.id} review={review} currentUserId={user?.id} onDeleted={reload} />
        ))
      )}
    </section>
  )
}
