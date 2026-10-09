import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createCafe } from '../services/api.js'

export default function NewCafePage() {
  const navigate = useNavigate()
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
      navigate(`/cafes/${cafe.id}`)
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
          <input className="field" value={address} onChange={(e) => setAddress(e.target.value)} />
        </label>
        <label className="label">
          Description
          <textarea
            className="field min-h-24 resize-y"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
          />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <button type="submit" disabled={submitting} className="btn mt-2">
          {submitting ? 'Saving...' : 'Create cafe'}
        </button>
      </form>
    </section>
  )
}
