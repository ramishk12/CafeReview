import { PageSkeleton } from '@/components/Skeleton'

// Skeleton for the cafe list while its data loads. Scoped to "/" so cafe detail
// pages return a real 404 when missing.
export default function Loading() {
  return <PageSkeleton />
}
