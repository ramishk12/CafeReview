'use client'

import { AuthProvider } from '@/context/AuthContext'
import { ToastProvider } from '@/context/ToastContext'

// Client-side providers. ToastProvider is outermost because AuthProvider uses it.
export default function Providers({ children }) {
  return (
    <ToastProvider>
      <AuthProvider>{children}</AuthProvider>
    </ToastProvider>
  )
}
