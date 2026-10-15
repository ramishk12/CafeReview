import RequireAuth from '@/components/RequireAuth'
import CafeForm from '@/components/CafeForm'

export const metadata = { title: 'Add a cafe' }

export default function NewCafePage() {
  return (
    <RequireAuth adminOnly>
      <CafeForm />
    </RequireAuth>
  )
}
