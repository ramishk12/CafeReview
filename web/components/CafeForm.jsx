'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import Spinner from '@/components/Spinner'
import { createCafe, updateCafe } from '@/lib/api'
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
        <label className="label">
          Description
          <textarea className="field min-h-24 resize-y" value={description} onChange={(e) => setDescription(e.target.value)} rows={4} />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <button type="submit" disabled={submitting} className="btn mt-2">
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
      </form>
    </section>
  )
}
