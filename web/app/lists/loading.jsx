import { Skeleton } from '@/components/Skeleton'

// Skeleton for the lists page while the signed-in user's lists load.
export default function Loading() {
  return (
    <div aria-busy="true" aria-label="Loading" className="mx-auto mt-8 max-w-2xl space-y-4">
      <Skeleton className="h-10 w-40 rounded-full" />
      <Skeleton className="h-36" />
      <Skeleton className="h-20" />
      <Skeleton className="h-20" />
    </div>
  )
}
