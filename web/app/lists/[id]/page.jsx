import RequireAuth from '@/components/RequireAuth'
import ListDetailClient from './ListDetailClient'

export const metadata = { title: 'List' }

export default function ListPage() {
  return (
    <RequireAuth>
      <ListDetailClient />
    </RequireAuth>
  )
}
