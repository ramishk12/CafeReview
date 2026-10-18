'use client'

import Link from 'next/link'
import { useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import Spinner from '@/components/Spinner'
import { verifyEmail } from '@/lib/api'

// Confirms the email from the link in the verification email.
export default function VerifyClient() {
  const searchParams = useSearchParams()
  const token = searchParams.get('token')
  const [state, setState] = useState('loading')
  const [message, setMessage] = useState('')

  useEffect(() => {
    if (!token) {
      setState('error')
      setMessage('This link is missing its confirmation token.')
      return
    }
    let active = true
    verifyEmail(token)
      .then((res) => {
        if (!active) return
        setState('ok')
        setMessage(res?.message || 'Your email is confirmed.')
      })
      .catch((err) => {
        if (!active) return
        setState('error')
        setMessage(err.message)
      })
    return () => {
      active = false
    }
  }, [token])

  return (
    <section className="card mx-auto mt-12 max-w-md p-8 text-center">
      {state === 'loading' && (
        <p className="flex items-center justify-center gap-2 text-slate-500 dark:text-slate-400">
          <Spinner /> Confirming your email…
        </p>
      )}

      {state === 'ok' && (
        <>
          <h1 className="text-2xl font-bold tracking-tight">Email confirmed</h1>
          <p className="mt-2 text-slate-500 dark:text-slate-400">{message}</p>
          <Link href="/login" className="btn mt-6">
            Log in
          </Link>
        </>
      )}

      {state === 'error' && (
        <>
          <h1 className="text-2xl font-bold tracking-tight">We couldn&apos;t confirm your email</h1>
          <p className="mt-2 text-slate-500 dark:text-slate-400">{message}</p>
          <p className="mt-4 text-sm text-slate-500 dark:text-slate-400">
            You can request a new link from the{' '}
            <Link href="/login" className="font-semibold">
              log in page
            </Link>
            .
          </p>
        </>
      )}
    </section>
  )
}
