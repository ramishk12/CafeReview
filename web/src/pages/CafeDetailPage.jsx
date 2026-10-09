import { ArrowLeft, MapPin } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getCafe, listReviews } from '../services/api.js'
import { useAuth } from '../context/AuthContext.jsx'
import { cafeGradient } from '../lib/cafeTheme.js'
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
  if (!cafe) return <p className="py-16 text-center text-slate-500">Loading...</p>

  const myReview = user ? reviews.find((r) => r.user_id === user.id) : null

  return (
    <section>
      <Link to="/" className="inline-flex items-center gap-1.5 text-sm font-medium text-slate-500 no-underline hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">
        <ArrowLeft size={16} /> All cafes
      </Link>

      <header className={`relative mt-4 mb-10 overflow-hidden rounded-3xl bg-gradient-to-br p-8 text-white shadow-xl ${cafeGradient(cafe.name)}`}>
        <div className="absolute inset-0 bg-black/10" />
        <div className="relative">
          <h1 className="text-4xl font-extrabold tracking-tight drop-shadow-sm">{cafe.name}</h1>
          {cafe.address && (
            <p className="mt-2 flex items-center gap-1.5 text-white/90">
              <MapPin size={16} /> {cafe.address}
            </p>
          )}
          {cafe.description && <p className="mt-3 max-w-2xl text-white/90">{cafe.description}</p>}
          <div className="mt-5 inline-flex items-center gap-3 rounded-2xl bg-white/15 px-4 py-2 backdrop-blur-sm">
            <span className="text-2xl font-bold">{cafe.review_count === 0 ? '–' : cafe.avg_rating.toFixed(1)}</span>
            <Stars value={cafe.avg_rating} size={18} />
            <span className="text-sm text-white/85">
              {cafe.review_count === 0
                ? 'No reviews yet'
                : `${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'}`}
            </span>
          </div>
        </div>
      </header>

      {user ? (
        <ReviewForm key={myReview?.id ?? 'new'} cafeId={cafe.id} existing={myReview} onSaved={reload} />
      ) : (
        <div className="card mb-10 flex flex-wrap items-center justify-between gap-4 p-6">
          <p className="text-slate-600 dark:text-slate-300">Share your experience with this cafe.</p>
          <Link
            to="/login"
            state={{ from: `/cafes/${id}` }}
            className="btn no-underline"
          >
            Log in to review
          </Link>
        </div>
      )}

      <h2 className="mb-5 text-2xl font-bold tracking-tight">Reviews</h2>
      {reviews.length === 0 ? (
        <p className="card py-12 text-center text-slate-500">No reviews yet. Be the first.</p>
      ) : (
        reviews.map((review) => (
          <ReviewCard key={review.id} review={review} currentUserId={user?.id} onDeleted={reload} />
        ))
      )}
    </section>
  )
}
