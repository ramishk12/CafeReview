import RequireAuth from '@/components/RequireAuth'
import ListsClient from './ListsClient'

export const metadata = { title: 'My lists' }

export default function ListsPage() {
  return (
    <RequireAuth>
      <ListsClient />
    </RequireAuth>
  )
}
