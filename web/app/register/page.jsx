import { Suspense } from 'react'
import RegisterForm from './RegisterForm'

export const metadata = { title: 'Sign up' }

export default function RegisterPage() {
  return (
    <Suspense fallback={<p className="py-16 text-center text-slate-500">Loading...</p>}>
      <RegisterForm />
    </Suspense>
  )
}
