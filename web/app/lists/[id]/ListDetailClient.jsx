'use client'

import Link from 'next/link'
import { useParams, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import CafeCard from '@/components/CafeCard'
import Stars from '@/components/Stars'
import { deleteList, getList, removeCafeFromList, removeReviewFromList, removeUserFromList } from '@/lib/api'
import { formatDate } from '@/lib/cafeTheme'
import { useToast } from '@/context/ToastContext'

export default function ListDetailClient() {
  const { id } = useParams()
  const router = useRouter()
  const toast = useToast()
  const [list, setList] = useState(null)
  const [error, setError] = useState('')

  const load = () =>
    getList(id)
      .then(setList)
      .catch((err) => setError(err.message))

  useEffect(() => {
    load()
  }, [id])

  const run = async (fn) => {
    try {
      await fn()
      load()
    } catch (err) {
      toast.show(err.message, 'error')
    }
  }

  const removeList = async () => {
    if (!window.confirm('Delete this list?')) return
    try {
      await deleteList(id)
      router.push('/lists')
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    }
  }

  if (error) return <p className="card mx-auto mt-8 max-w-2xl p-6 text-slate-500">{error}</p>
  if (!list) return <p className="py-16 text-center text-slate-500">Loading…</p>

  const isEmpty = list.cafes.length === 0 && list.users.length === 0 && list.reviews.length === 0

  return (
    <section>
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-3xl font-bold tracking-tight">{list.name}</h1>
        <button type="button" onClick={removeList} className="btn btn-danger px-3.5 py-2 text-sm">
          Delete list
        </button>
      </div>

      {isEmpty && <p className="card p-6 text-slate-500 dark:text-slate-400">Nothing in this list yet. Add cafes, reviewers, or reviews from their pages.</p>}

      {list.cafes.length > 0 && (
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-bold tracking-tight">Cafes</h2>
          <ul className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
            {list.cafes.map((cafe) => (
              <li key={cafe.id} className="relative">
                <CafeCard cafe={cafe} />
                <button
                  type="button"
                  onClick={() => run(() => removeCafeFromList(id, cafe.id))}
                  className="btn btn-ghost absolute right-3 bottom-3 bg-white/90 px-3 py-1.5 text-xs dark:bg-slate-900/90"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      {list.users.length > 0 && (
        <section className="mb-10">
          <h2 className="mb-4 text-xl font-bold tracking-tight">Reviewers</h2>
          <ul className="grid gap-3">
            {list.users.map((reviewer) => (
              <li key={reviewer.id} className="card flex items-center justify-between gap-3 p-5">
                <Link href={`/users/${reviewer.id}`} className="font-semibold no-underline">
                  {reviewer.display_name}
                </Link>
                <button
                  type="button"
                  onClick={() => run(() => removeUserFromList(id, reviewer.id))}
                  className="btn btn-ghost px-3 py-1.5 text-xs"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      {list.reviews.length > 0 && (
        <section>
          <h2 className="mb-4 text-xl font-bold tracking-tight">Reviews</h2>
          <ul className="grid gap-4">
            {list.reviews.map((review) => (
              <li key={review.id} className="card p-6">
                <div className="flex items-center justify-between gap-3">
                  <Link href={`/cafes/${review.cafe_id}`} className="font-semibold no-underline">
                    {review.cafe_name}
                  </Link>
                  <Stars value={review.rating} size={15} />
                </div>
                <p className="text-xs text-slate-400">
                  by{' '}
                  <Link href={`/users/${review.user_id}`} className="no-underline">
                    {review.author_name}
                  </Link>{' '}
                  · {review.visited_on ? `visited ${formatDate(review.visited_on)}` : formatDate(review.created_at)}
                </p>
                <p className="mt-3 whitespace-pre-wrap leading-relaxed text-slate-700 dark:text-slate-300">{review.body}</p>
                <button
                  type="button"
                  onClick={() => run(() => removeReviewFromList(id, review.id))}
                  className="btn btn-ghost mt-3 px-3 py-1.5 text-xs"
                >
                  Remove
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}
    </section>
  )
}
