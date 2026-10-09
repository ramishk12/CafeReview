'use client'

import { createContext, useCallback, useContext, useMemo, useRef, useState } from 'react'
import { CheckCircle2, CircleAlert, X } from 'lucide-react'

const ToastContext = createContext(null)
const DISMISS_AFTER_MS = 4000

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])
  const nextId = useRef(1)

  const dismiss = useCallback((id) => setToasts((list) => list.filter((t) => t.id !== id)), [])

  const show = useCallback(
    (message, type = 'success') => {
      const id = nextId.current++
      setToasts((list) => [...list, { id, message, type }])
      setTimeout(() => dismiss(id), DISMISS_AFTER_MS)
    },
    [dismiss],
  )

  const value = useMemo(() => ({ show }), [show])

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div aria-live="polite" className="pointer-events-none fixed right-4 bottom-4 z-50 flex w-full max-w-sm flex-col gap-2">
        {toasts.map((t) => {
          const isError = t.type === 'error'
          const Icon = isError ? CircleAlert : CheckCircle2
          return (
            <div
              key={t.id}
              role="status"
              className={`pointer-events-auto flex items-start gap-3 rounded-2xl border p-4 text-sm shadow-xl backdrop-blur-md ${
                isError
                  ? 'border-red-200 bg-red-50/95 text-red-800 dark:border-red-900/50 dark:bg-red-950/90 dark:text-red-200'
                  : 'border-emerald-200 bg-white/95 text-slate-800 dark:border-emerald-900/50 dark:bg-slate-900/95 dark:text-slate-100'
              }`}
            >
              <Icon size={18} className={`mt-0.5 shrink-0 ${isError ? 'text-red-500' : 'text-emerald-500'}`} />
              <p className="flex-1 leading-snug">{t.message}</p>
              <button type="button" onClick={() => dismiss(t.id)} className="btn-ghost -m-1 rounded-full p-1" aria-label="Dismiss">
                <X size={14} />
              </button>
            </div>
          )
        })}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used inside ToastProvider')
  return ctx
}
