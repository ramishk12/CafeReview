'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { useAuth } from '@/context/AuthContext'

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

  if (loading || !allowed) return <p className="py-16 text-center text-slate-500">Loading...</p>
  return children
}
