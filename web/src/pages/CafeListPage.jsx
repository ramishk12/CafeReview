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

  if (error) return <p className="alert-error">{error}</p>
  if (!cafes) return <p className="py-12 text-center text-slate-500">Loading cafes...</p>
  if (cafes.length === 0) return <p className="py-12 text-center text-slate-500">No cafes yet.</p>

  return (
    <section>
      <div className="mb-6">
        <h1 className="text-3xl font-bold tracking-tight">Cafes</h1>
        <p className="text-slate-500 dark:text-slate-400">Find a spot and see what people think.</p>
      </div>

      <ul className="grid grid-cols-[repeat(auto-fill,minmax(270px,1fr))] gap-4">
        {cafes.map((cafe) => (
          <li key={cafe.id}>
            <Link
              to={`/cafes/${cafe.id}`}
              className="card block h-full p-5 no-underline transition duration-200 hover:-translate-y-0.5 hover:shadow-lg"
            >
              <h2 className="m-0 text-lg font-semibold tracking-tight text-slate-900 dark:text-slate-100">{cafe.name}</h2>
              <p className="mt-1 mb-3 text-sm text-slate-500 dark:text-slate-400">{cafe.address}</p>
              <div className="flex items-center gap-2 text-sm">
                <Stars value={cafe.avg_rating} />
                <span className="text-slate-500 dark:text-slate-400">
                  {cafe.review_count === 0
                    ? 'No reviews yet'
                    : `${cafe.avg_rating.toFixed(1)} (${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'})`}
                </span>
              </div>
            </Link>
          </li>
        ))}
      </ul>
    </section>
  )
}
