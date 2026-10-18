import { DetailSkeleton } from '@/components/Skeleton'

// Skeleton for a profile page while the user and their reviews load.
export default function Loading() {
  return (
    <div className="mx-auto max-w-3xl">
      <DetailSkeleton rows={2} />
    </div>
  )
}
