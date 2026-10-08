import { createContext, useContext, useEffect, useState } from 'react'
import * as api from '../services/api.js'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(Boolean(api.getToken()))

  // Restore the session from a stored token on first load.
  useEffect(() => {
    if (!api.getToken()) return
    api
      .me()
      .then(setUser)
      .catch(() => api.setToken(null))
      .finally(() => setLoading(false))
  }, [])

  const saveSession = ({ token, user }) => {
    api.setToken(token)
    setUser(user)
    return user
  }

  const value = {
    user,
    loading,
    isAdmin: user?.role === 'admin',
    login: async (email, password) => saveSession(await api.login(email, password)),
    register: async (email, password, displayName) =>
      saveSession(await api.register(email, password, displayName)),
    logout: () => {
      api.setToken(null)
      setUser(null)
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
