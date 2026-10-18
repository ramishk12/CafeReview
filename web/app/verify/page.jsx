import { Suspense } from 'react'
import Spinner from '@/components/Spinner'
import VerifyClient from './VerifyClient'

export const metadata = { title: 'Confirm email' }

export default function VerifyPage() {
  return (
    <Suspense
      fallback={
        <p className="flex items-center justify-center gap-2 py-16 text-slate-500">
          <Spinner /> Confirming your email…
        </p>
      }
    >
      <VerifyClient />
    </Suspense>
  )
}
