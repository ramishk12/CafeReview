'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { useAuth } from '@/context/AuthContext'
import { Skeleton } from './Skeleton'

// Client-side guard: sends signed-out users to /login and non-admins home for adminOnly pages.
export default function RequireAuth({ children, adminOnly = false }) {
  const { user, loading, isAdmin } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  const allowed = !!user && (!adminOnly || isAdmin)

  useEffect(() => {
    if (loading || allowed) return
    if (!user) router.replace(`/login?from=${encodeURIComponent(pathname)}`)
    else router.replace('/')
  }, [loading, allowed, user, router, pathname])

  // The shared skeleton replaces the plain "Loading..." text this used to show.
  if (loading || !allowed) {
    return (
      <div aria-busy="true" aria-label="Loading" className="mx-auto mt-8 max-w-md space-y-4">
        <Skeleton className="h-8 w-40 rounded-full" />
        <Skeleton className="h-72" />
      </div>
    )
  }
  return children
}
