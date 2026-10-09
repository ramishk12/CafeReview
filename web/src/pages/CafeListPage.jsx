import { MapPin } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listCafes } from '../services/api.js'
import Stars from '../components/Stars.jsx'
import { cafeGradient } from '../lib/cafeTheme.js'

export default function CafeListPage() {
  const [cafes, setCafes] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    listCafes().then(setCafes).catch((err) => setError(err.message))
  }, [])

  if (error) return <p className="alert-error">{error}</p>
  if (!cafes) return <p className="py-16 text-center text-slate-500">Loading cafes...</p>

  return (
    <section>
      <div className="mb-10 pt-4">
        <p className="eyebrow">Discover</p>
        <h1 className="mt-2 text-4xl font-extrabold tracking-tight sm:text-5xl">
          Find your next favorite <span className="bg-gradient-to-r from-indigo-600 to-fuchsia-500 bg-clip-text text-transparent">cafe</span>
        </h1>
        <p className="mt-3 max-w-xl text-lg text-slate-500 dark:text-slate-400">
          Honest reviews and photos from people who actually drink the coffee.
        </p>
      </div>

      {cafes.length === 0 ? (
        <p className="card py-16 text-center text-slate-500">No cafes yet.</p>
      ) : (
        <ul className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
          {cafes.map((cafe) => (
            <li key={cafe.id}>
              <Link
                to={`/cafes/${cafe.id}`}
                className="card group flex h-full flex-col overflow-hidden no-underline transition duration-300 hover:-translate-y-1 hover:shadow-xl"
              >
                <div className={`relative h-28 bg-gradient-to-br ${cafeGradient(cafe.name)}`}>
                  <span className="absolute top-3 right-3 flex items-center gap-1 rounded-full bg-white/90 px-2.5 py-1 text-sm font-semibold text-slate-900 shadow-sm">
                    {cafe.review_count === 0 ? 'New' : cafe.avg_rating.toFixed(1)}
                    {cafe.review_count > 0 && <span className="text-amber-500">★</span>}
                  </span>
                </div>
                <div className="flex flex-1 flex-col p-5">
                  <h2 className="text-lg font-semibold tracking-tight text-slate-900 dark:text-white">{cafe.name}</h2>
                  {cafe.address && (
                    <p className="mt-1 flex items-center gap-1.5 text-sm text-slate-500 dark:text-slate-400">
                      <MapPin size={14} className="shrink-0" />
                      {cafe.address}
                    </p>
                  )}
                  <div className="mt-auto flex items-center gap-2 pt-4 text-sm">
                    <Stars value={cafe.avg_rating} size={14} />
                    <span className="text-slate-500 dark:text-slate-400">
                      {cafe.review_count === 0
                        ? 'No reviews yet'
                        : `${cafe.review_count} ${cafe.review_count === 1 ? 'review' : 'reviews'}`}
                    </span>
                  </div>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
