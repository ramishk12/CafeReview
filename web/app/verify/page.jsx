import { Suspense } from 'react'
import VerifyClient from './VerifyClient'

export const metadata = { title: 'Confirm email' }

export default function VerifyPage() {
  return (
    <Suspense fallback={<p className="py-16 text-center text-slate-500">Confirming your email…</p>}>
      <VerifyClient />
    </Suspense>
  )
}
