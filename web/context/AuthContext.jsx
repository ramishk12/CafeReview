'use client'

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { AUTH_EXPIRED_EVENT, login as apiLogin, logout as apiLogout, me, register as apiRegister } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const toast = useToast()
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  // Track the current user so the expiry handler knows whether a session was active.
  const userRef = useRef(null)
  useEffect(() => {
    userRef.current = user
  }, [user])

  // Restore the session from the HttpOnly cookie. A 401 here just means signed
  // out; anything else (network, 5xx) looks identical, so say so.
  useEffect(() => {
    me()
      .then(setUser)
      .catch((err) => {
        setUser(null)
        if (err?.status !== 401) {
          toast.show("We couldn't check your session. Refresh the page to try again.", 'error')
        }
      })
      .finally(() => setLoading(false))
  }, [toast])

  // A 401 on any other request means the session expired. Only notify if a user
  // was signed in, so the initial check does not raise a false alarm.
  useEffect(() => {
    const onExpired = () => {
      if (!userRef.current) return
      setUser(null)
      toast.show('Your session expired. Please log in again.', 'error')
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired)
  }, [toast])

  const value = useMemo(
    () => ({
      user,
      loading,
      isAdmin: user?.role === 'admin',
      login: async (email, password) => {
        const next = await apiLogin(email, password)
        setUser(next)
        return next
      },
      // Registration no longer signs the user in; they must confirm their email first.
      register: (email, password, displayName) => apiRegister(email, password, displayName),
      logout: async () => {
        // Best effort: clear the session on the server, then locally.
        try {
          await apiLogout()
        } catch {
          // Ignore; the local state is cleared regardless.
        }
        setUser(null)
      },
    }),
    [user, loading],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
