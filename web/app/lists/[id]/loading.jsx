import { DetailSkeleton } from '@/components/Skeleton'

// Skeleton for a list page while the list and its items load.
export default function Loading() {
  return <DetailSkeleton rows={2} />
}
