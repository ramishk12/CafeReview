'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import RequireAuth from '@/components/RequireAuth'
import { createCafe } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

export default function NewCafePage() {
  return (
    <RequireAuth adminOnly>
      <NewCafeForm />
    </RequireAuth>
  )
}

function NewCafeForm() {
  const router = useRouter()
  const toast = useToast()
  const [name, setName] = useState('')
  const [address, setAddress] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const cafe = await createCafe({ name, address, description })
      toast.show(`${cafe.name} added`)
      router.push(`/cafes/${cafe.id}`)
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  return (
    <section className="card mx-auto mt-8 max-w-md p-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Add a cafe</h1>
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
          {submitting ? 'Saving…' : 'Create cafe'}
        </button>
      </form>
    </section>
  )
}
