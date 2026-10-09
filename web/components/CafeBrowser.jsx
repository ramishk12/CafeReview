'use client'

import { Search, SlidersHorizontal } from 'lucide-react'
import { useMemo, useState } from 'react'
import CafeCard from './CafeCard'

const SORTS = {
  rating: { label: 'Top rated', compare: (a, b) => b.avg_rating - a.avg_rating || b.review_count - a.review_count },
  reviews: { label: 'Most reviewed', compare: (a, b) => b.review_count - a.review_count || b.avg_rating - a.avg_rating },
  name: { label: 'A–Z', compare: (a, b) => a.name.localeCompare(b.name) },
}

// Search and sort run in the browser; the list itself is fetched on the server.
export default function CafeBrowser({ cafes }) {
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('rating')

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase()
    const matches = q
      ? cafes.filter((c) => c.name.toLowerCase().includes(q) || (c.address || '').toLowerCase().includes(q))
      : cafes
    return [...matches].sort(SORTS[sort].compare)
  }, [cafes, query, sort])

  return (
    <div>
      <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center">
        <label className="relative flex-1">
          <span className="sr-only">Search cafes</span>
          <Search size={17} aria-hidden="true" className="pointer-events-none absolute top-1/2 left-3.5 -translate-y-1/2 text-slate-400" />
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search by name or address"
            className="field pl-10"
          />
        </label>
        <label className="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
          <SlidersHorizontal size={16} aria-hidden="true" />
          <span className="sr-only sm:not-sr-only">Sort</span>
          <select value={sort} onChange={(e) => setSort(e.target.value)} className="field w-auto py-2.5 pr-8">
            {Object.entries(SORTS).map(([key, s]) => (
              <option key={key} value={key}>
                {s.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {visible.length === 0 ? (
        <p className="card py-16 text-center text-slate-500">
          {cafes.length === 0 ? 'No cafes yet.' : `No cafes match “${query.trim()}”.`}
        </p>
      ) : (
        <ul className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
          {visible.map((cafe) => (
            <li key={cafe.id}>
              <CafeCard cafe={cafe} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
