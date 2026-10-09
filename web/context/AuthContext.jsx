'use client'

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import * as api from '@/lib/api'
import { AUTH_EXPIRED_EVENT, getToken, setToken } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const toast = useToast()
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  // Restore the session from the stored token.
  useEffect(() => {
    if (!getToken()) {
      setLoading(false)
      return
    }
    api
      .me()
      .then(setUser)
      .catch(() => setToken(null))
      .finally(() => setLoading(false))
  }, [])

  // Any request that gets a 401 dispatches this event; drop the user and say why.
  useEffect(() => {
    const onExpired = () => {
      setUser(null)
      toast.show('Your session expired. Please log in again.', 'error')
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired)
  }, [toast])

  const saveSession = useCallback(({ token, user: nextUser }) => {
    setToken(token)
    setUser(nextUser)
    return nextUser
  }, [])

  const value = useMemo(
    () => ({
      user,
      loading,
      isAdmin: user?.role === 'admin',
      login: async (email, password) => saveSession(await api.login(email, password)),
      register: async (email, password, displayName) =>
        saveSession(await api.register(email, password, displayName)),
      logout: () => {
        setToken(null)
        setUser(null)
      },
    }),
    [user, loading, saveSession],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
