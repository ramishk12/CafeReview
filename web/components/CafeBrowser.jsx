'use client'

import { Coffee, Search, SlidersHorizontal } from 'lucide-react'
import Link from 'next/link'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import CafeCard from './CafeCard'
import EmptyState from './EmptyState'

const SORTS = {
  rating: 'Top rated',
  reviews: 'Most reviewed',
  newest: 'Newest',
  name: 'A–Z',
}

// Presentational list plus controls. Filtering, sorting, and paging all happen
// on the server; the controls write to the URL, so views are shareable.
export default function CafeBrowser({ cafes, tags = [], total = 0, page = 1, pageSize = 24 }) {
  const { isAdmin } = useAuth()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const sort = searchParams.get('sort') || 'rating'
  const currentTags = (searchParams.get('tags') || '').split(',').filter(Boolean)
  const hasFilters = Boolean((searchParams.get('q') || '').trim()) || currentTags.length > 0
  const [query, setQuery] = useState(searchParams.get('q') || '')

  const buildHref = (updates) => {
    const next = new URLSearchParams(searchParams.toString())
    for (const [key, value] of Object.entries(updates)) {
      if (value === undefined || value === null || value === '') next.delete(key)
      else next.set(key, String(value))
    }
    const qs = next.toString()
    return qs ? `${pathname}?${qs}` : pathname
  }

  // Debounced search: typing updates the URL, which resets to page 1.
  useEffect(() => {
    if (query === (searchParams.get('q') || '')) return
    const timer = setTimeout(() => router.push(buildHref({ q: query, page: undefined })), 300)
    return () => clearTimeout(timer)
  }, [query]) // eslint-disable-line react-hooks/exhaustive-deps

  const toggleTag = (id) => {
    const value = String(id)
    const ids = currentTags.includes(value) ? currentTags.filter((t) => t !== value) : [...currentTags, value]
    router.push(buildHref({ tags: ids.join(','), page: undefined }))
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  let start = Math.max(1, page - 2)
  let end = Math.min(totalPages, start + 4)
  start = Math.max(1, end - 4)
  const pages = []
  for (let p = start; p <= end; p++) pages.push(p)

  const tagGroups = {}
  for (const tag of tags) (tagGroups[tag.category] ||= []).push(tag)

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
          <select
            value={sort}
            onChange={(e) => router.push(buildHref({ sort: e.target.value, page: undefined }))}
            className="field w-auto py-2.5 pr-8"
          >
            {Object.entries(SORTS).map(([key, label]) => (
              <option key={key} value={key}>
                {label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {Object.keys(tagGroups).length > 0 && (
        <div className="mb-6 flex flex-col gap-2">
          {Object.entries(tagGroups).map(([category, list]) => (
            <div key={category} className="flex flex-wrap items-center gap-2">
              <span className="text-xs font-semibold tracking-wide text-slate-400 uppercase">{category}</span>
              {list.map((tag) => {
                const active = currentTags.includes(String(tag.id))
                return (
                  <button
                    key={tag.id}
                    type="button"
                    onClick={() => toggleTag(tag.id)}
                    aria-pressed={active}
                    className={`rounded-full px-3 py-1 text-sm transition ${
                      active
                        ? 'bg-indigo-600 text-white'
                        : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-white/10 dark:text-slate-200 dark:hover:bg-white/20'
                    }`}
                  >
                    {tag.name}
                  </button>
                )
              })}
            </div>
          ))}
        </div>
      )}

      {cafes.length === 0 ? (
        hasFilters ? (
          <EmptyState
            icon={Coffee}
            title="No cafes match your search"
            description="Try a different name or address, or remove a tag filter."
            action={
              <button
                type="button"
                onClick={() => router.push(buildHref({ q: undefined, tags: undefined, page: undefined }))}
                className="btn"
              >
                Clear search
              </button>
            }
          />
        ) : total > 0 ? (
          <EmptyState
            icon={Coffee}
            title="No cafes on this page"
            description="This page is past the end of the list."
            action={
              <button type="button" onClick={() => router.push(buildHref({ page: undefined }))} className="btn">
                Back to the first page
              </button>
            }
          />
        ) : (
          <EmptyState
            icon={Coffee}
            title="No cafes yet"
            description="Cafes will show up here once they are added."
            action={
              isAdmin ? (
                <Link href="/cafes/new" className="btn">
                  Add a cafe
                </Link>
              ) : null
            }
          />
        )
      ) : (
        <ul className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
          {cafes.map((cafe) => (
            <li key={cafe.id}>
              <CafeCard cafe={cafe} isAdmin={isAdmin} />
            </li>
          ))}
        </ul>
      )}

      {totalPages > 1 && (
        <nav className="mt-10 flex items-center justify-center gap-2" aria-label="Pagination">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => router.push(buildHref({ page: page - 1 === 1 ? undefined : page - 1 }))}
            className="btn btn-ghost px-3.5 py-2 text-sm"
          >
            Previous
          </button>
          {pages.map((p) => (
            <button
              key={p}
              type="button"
              onClick={() => router.push(buildHref({ page: p === 1 ? undefined : p }))}
              aria-current={p === page ? 'page' : undefined}
              className={`rounded-full px-3.5 py-1.5 text-sm ${
                p === page ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200'
              }`}
            >
              {p}
            </button>
          ))}
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => router.push(buildHref({ page: page + 1 }))}
            className="btn btn-ghost px-3.5 py-2 text-sm"
          >
            Next
          </button>
        </nav>
      )}
    </div>
  )
}
