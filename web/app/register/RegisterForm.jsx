'use client'

import Link from 'next/link'
import { useState } from 'react'
import Spinner from '@/components/Spinner'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'
import { resendVerification } from '@/lib/api'

export default function RegisterForm() {
  const { register } = useAuth()
  const toast = useToast()
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [resending, setResending] = useState(false)
  const [sent, setSent] = useState(false)

  const handleResend = async () => {
    setResending(true)
    try {
      await resendVerification(email)
      toast.show('Verification email sent')
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setResending(false)
    }
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)
    try {
      await register(email, password, displayName)
      setSent(true)
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
    }
  }

  if (sent) {
    return (
      <section className="card mx-auto mt-8 max-w-md p-8 text-center">
        <h1 className="mb-2 text-2xl font-bold tracking-tight">Check your inbox</h1>
        <p className="text-slate-500 dark:text-slate-400">
          We&apos;ve sent a confirmation link to <span className="font-medium">{email}</span>. Open it to activate your
          account.
        </p>
        <p className="mt-4 text-sm text-slate-500 dark:text-slate-400">
          Didn&apos;t get it? Check your spam folder, or{' '}
          <button type="button" className="font-semibold" onClick={handleResend} disabled={resending}>
            {resending ? 'sending…' : 'resend the email'}
          </button>
          .
        </p>
        <p className="mt-6 text-sm text-slate-500 dark:text-slate-400">
          Already confirmed?{' '}
          <Link href="/login" className="font-semibold">
            Log in
          </Link>
        </p>
      </section>
    )
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
          {submitting ? (
            <>
              <Spinner /> Creating account…
            </>
          ) : (
            'Sign up'
          )}
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
