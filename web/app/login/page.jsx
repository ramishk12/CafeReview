import { Suspense } from 'react'
import { Skeleton } from '@/components/Skeleton'
import LoginForm from './LoginForm'

export const metadata = { title: 'Log in' }

export default function LoginPage() {
  return (
    <Suspense fallback={<Skeleton className="mx-auto mt-8 h-80 max-w-md" />}>
      <LoginForm />
    </Suspense>
  )
}
