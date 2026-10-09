'use client'

import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

// Only allow in-app redirects after login (prevents open redirects via ?from=).
function safeRedirect(from) {
  return from && from.startsWith('/') && !from.startsWith('//') ? from : '/'
}

export default function LoginForm() {
  const { login } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const searchParams = useSearchParams()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      const user = await login(email, password)
      toast.show(`Welcome back, ${user.display_name}`)
      router.replace(safeRedirect(searchParams.get('from')))
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  return (
    <section className="card mx-auto mt-8 max-w-md p-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Log in</h1>
      <form onSubmit={handleSubmit} className="grid gap-4">
        <label className="label">
          Email
          <input type="email" autoComplete="email" className="field" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label className="label">
          Password
          <input
            type="password"
            autoComplete="current-password"
            className="field"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <button type="submit" disabled={submitting} className="btn mt-2">
          {submitting ? 'Logging in…' : 'Log in'}
        </button>
      </form>
      <p className="mt-6 text-sm text-slate-500 dark:text-slate-400">
        No account?{' '}
        <Link href="/register" className="font-semibold">
          Sign up
        </Link>
      </p>
    </section>
  )
}
