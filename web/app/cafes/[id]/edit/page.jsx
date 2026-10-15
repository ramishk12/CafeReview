import { notFound } from 'next/navigation'
import RequireAuth from '@/components/RequireAuth'
import CafeForm from '@/components/CafeForm'
import { getCafe } from '@/lib/serverApi'

export const metadata = { title: 'Edit cafe' }

export default async function EditCafePage({ params }) {
  const { id } = await params
  if (!/^\d+$/.test(id)) notFound()

  const cafe = await getCafe(id)
  if (!cafe) notFound()

  return (
    <RequireAuth adminOnly>
      <CafeForm cafe={cafe} />
    </RequireAuth>
  )
}
