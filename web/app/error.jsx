'use client'

import { TriangleAlert } from 'lucide-react'
import { useEffect } from 'react'
import EmptyState from '@/components/EmptyState'

// Catches errors thrown while rendering a route (for example, when the API is
// down). It uses the shared empty-state design and offers a retry.
export default function ErrorPage({ error, reset }) {
  useEffect(() => {
    console.error(error)
  }, [error])

  return (
    <div className="mx-auto mt-12 max-w-md">
      <EmptyState
        icon={TriangleAlert}
        titleAs="h1"
        title="We're having trouble loading this page"
        description="Please try again in a moment. If the problem continues, let us know."
        action={
          <button type="button" onClick={reset} className="btn">
            Try again
          </button>
        }
      />
    </div>
  )
}
