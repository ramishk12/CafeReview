'use client'

import { useEffect } from 'react'

// Catches errors thrown while rendering a route (for example, when the API is down).
export default function ErrorPage({ error, reset }) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="card mx-auto mt-12 max-w-md p-8 text-center">
      <h1 className="text-2xl font-bold tracking-tight">Something went wrong</h1>
      <p className="mt-2 text-slate-500 dark:text-slate-400">
        We couldn&apos;t load this page. Check that the API is running, then try again.
      </p>
      <button type="button" onClick={reset} className="btn mt-6">
        Try again
      </button>
    </div>
  )
}
