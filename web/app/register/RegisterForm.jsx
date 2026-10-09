'use client'

import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

function safeRedirect(from) {
  return from && from.startsWith('/') && !from.startsWith('//') ? from : '/'
}

export default function RegisterForm() {
  const { register } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const searchParams = useSearchParams()
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
      toast.show('Account created. Welcome!')
      router.replace(safeRedirect(searchParams.get('from')))
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  return (
    <section className="card mx-auto mt-8 max-w-md p-8">
      <h1 className="mb-6 text-2xl font-bold tracking-tight">Create an account</h1>
      <form onSubmit={handleSubmit} className="grid gap-4">
        <label className="label">
          Display name
          <input className="field" value={displayName} onChange={(e) => setDisplayName(e.target.value)} maxLength={100} required />
        </label>
        <label className="label">
          Email
          <input type="email" autoComplete="email" className="field" value={email} onChange={(e) => setEmail(e.target.value)} required />
        </label>
        <label className="label">
          Password (at least 8 characters)
          <input
            type="password"
            autoComplete="new-password"
            className="field"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            minLength={8}
            required
          />
        </label>
        {error && <p className="alert-error">{error}</p>}
        <button type="submit" disabled={submitting} className="btn mt-2">
          {submitting ? 'Creating account…' : 'Sign up'}
        </button>
      </form>
      <p className="mt-6 text-sm text-slate-500 dark:text-slate-400">
        Already have an account?{' '}
        <Link href="/login" className="font-semibold">
          Log in
        </Link>
      </p>
    </section>
  )
}
