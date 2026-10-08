import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listCafes } from '../services/api.js'
import Stars from '../components/Stars.jsx'

export default function CafeListPage() {
  const [cafes, setCafes] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    listCafes().then(setCafes).catch((err) => setError(err.message))
  }, [])

  if (error) return <p className="error">{error}</p>
  if (!cafes) return <p className="status">Loading cafes...</p>
  if (cafes.length === 0) return <p className="status">No cafes yet.</p>

  return (
    <section>
      <h1>Cafes</h1>
      <ul className="cafe-list">
        {cafes.map((cafe) => (
          <li key={cafe.id} className="cafe-item">
            <Link to={`/cafes/${cafe.id}`}>
              <strong>{cafe.name}</strong>
            </Link>
            <div className="muted">{cafe.address}</div>
            <div>
              <Stars value={cafe.avg_rating} />{' '}
              <span className="muted">
                {cafe.review_count === 0
                  ? 'No reviews yet'
                  : `${cafe.avg_rating.toFixed(1)} (${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'})`}
              </span>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
