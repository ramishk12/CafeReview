import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'

export default function RegisterPage() {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      await register(email, password, displayName)
      navigate('/', { replace: true })
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  return (
    <section className="card mx-auto mt-8 max-w-md p-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Sign up</h1>
      <form onSubmit={handleSubmit} className="grid gap-4">
        <label className="label">
          Display name
          <input
            className="field"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            maxLength={100}
            required
          />
        </label>
        <label className="label">
          Email
          <input type="email" className="field" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label className="label">
          Password (at least 8 characters)
          <input
            type="password"
            className="field"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <button type="submit" disabled={submitting} className="btn mt-2">
          {submitting ? 'Creating account...' : 'Sign up'}
        </button>
      </form>
      <p className="mt-6 text-sm text-slate-500 dark:text-slate-400">
        Already have an account?{' '}
        <Link to="/login" className="font-semibold">
          Log in
        </Link>
      </p>
    </section>
  )
}
