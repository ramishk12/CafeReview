'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import Spinner from '@/components/Spinner'
import { createCafe, listSimilarCafes, updateCafe } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

// Creates a cafe, or edits one when `cafe` is provided. Admin only (guarded by
// the page that renders it).
export default function CafeForm({ cafe }) {
  const router = useRouter()
  const toast = useToast()
  const editing = Boolean(cafe)
  const [name, setName] = useState(cafe?.name ?? '')
  const [address, setAddress] = useState(cafe?.address ?? '')
  const [description, setDescription] = useState(cafe?.description ?? '')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [similar, setSimilar] = useState([])

  // Advisory duplicate check while adding a cafe: look for existing cafes with
  // a similar name or address and show them, without blocking the save.
  useEffect(() => {
    const trimmed = name.trim()
    if (editing || trimmed.length < 3) {
      setSimilar([])
      return
    }
    let active = true
    const timer = setTimeout(() => {
      listSimilarCafes(trimmed, address.trim())
        .then((matches) => {
          if (active) setSimilar(matches)
        })
        .catch(() => {})
    }, 350)
    return () => {
      active = false
      clearTimeout(timer)
    }
  }, [editing, name, address])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const saved = editing
        ? await updateCafe(cafe.id, { name, address, description })
        : await createCafe({ name, address, description })
      toast.show(editing ? `${saved.name} updated` : `${saved.name} added`)
      router.push(`/cafes/${saved.id}`)
      router.refresh()
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  return (
    <section className="card mx-auto mt-8 max-w-md p-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">{editing ? 'Edit cafe' : 'Add a cafe'}</h1>
      <form onSubmit={handleSubmit} className="grid gap-4">
        <label className="label">
          Name
          <input className="field" value={name} onChange={(e) => setName(e.target.value)} maxLength={255} required />
        </label>
        <label className="label">
          Address
          <input className="field" value={address} onChange={(e) => setAddress(e.target.value)} maxLength={255} />
        </label>

        {similar.length > 0 && (
          <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm dark:border-amber-900/50 dark:bg-amber-950/40">
            <p className="font-semibold text-amber-800 dark:text-amber-200">Similar cafes already exist</p>
            <p className="mt-1 text-amber-700 dark:text-amber-300">
              Check these before creating a duplicate. You can still add this cafe.
            </p>
            <ul className="mt-3 grid gap-2">
              {similar.map((match) => (
                <li key={match.id}>
                  <Link
                    href={`/cafes/${match.id}`}
                    className="font-semibold text-amber-900 no-underline hover:underline dark:text-amber-100"
                  >
                    {match.name}
                  </Link>
                  {match.address && <span className="text-amber-700 dark:text-amber-300"> · {match.address}</span>}
                </li>
              ))}
            </ul>
          </div>
        )}
        <label className="label">
          Description
          <textarea className="field min-h-24 resize-y" value={description} onChange={(e) => setDescription(e.target.value)} rows={4} />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <div className="mt-2 flex flex-wrap items-center gap-3">
          <button type="submit" disabled={submitting} className="btn">
            {submitting ? (
              <>
                <Spinner /> Saving…
              </>
            ) : editing ? (
              'Save changes'
            ) : (
              'Create cafe'
            )}
          </button>
          <Link href={editing ? `/cafes/${cafe.id}` : '/'} className="btn btn-ghost">
            Cancel
          </Link>
        </div>
      </form>
    </section>
  )
}
