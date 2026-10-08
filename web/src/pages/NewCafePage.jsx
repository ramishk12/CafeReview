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
    <section className="narrow">
      <h1>Add a cafe</h1>
      <form onSubmit={handleSubmit} className="stacked-form">
        <label>
          Name
          <input value={name} onChange={(e) => setName(e.target.value)} maxLength={255} required />
        </label>
        <label>
          Address
          <input value={address} onChange={(e) => setAddress(e.target.value)} />
        </label>
        <label>
          Description
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={4} />
        </label>
        {error && <p className="error">{error}</p>}
        <button type="submit" disabled={submitting}>
          {submitting ? 'Saving...' : 'Create cafe'}
        </button>
      </form>
    </section>
  )
}
