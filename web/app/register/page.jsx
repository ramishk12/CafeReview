import { Suspense } from 'react'
import { Skeleton } from '@/components/Skeleton'
import RegisterForm from './RegisterForm'

export const metadata = { title: 'Sign up' }

export default function RegisterPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto mt-8 h-96 max-w-md" />}>
      <RegisterForm />
    </Suspense>
  )
}
