'use client'

import { useParams, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import CafeCard from '@/components/CafeCard'
import { deleteList, getList, removeCafeFromList } from '@/lib/api'
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

  const remove = async (cafeId) => {
    try {
      await removeCafeFromList(id, cafeId)
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

  return (
    <section>
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-3xl font-bold tracking-tight">{list.name}</h1>
        <button type="button" onClick={removeList} className="btn btn-danger px-3.5 py-2 text-sm">
          Delete list
        </button>
      </div>

      {list.cafes.length === 0 ? (
        <p className="card p-6 text-slate-500 dark:text-slate-400">No cafes in this list yet. Add some from a cafe page.</p>
      ) : (
        <ul className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5">
          {list.cafes.map((cafe) => (
            <li key={cafe.id} className="relative">
              <CafeCard cafe={cafe} />
              <button
                type="button"
                onClick={() => remove(cafe.id)}
                className="btn btn-ghost absolute right-3 bottom-3 bg-white/90 px-3 py-1.5 text-xs dark:bg-slate-900/90"
              >
                Remove
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
