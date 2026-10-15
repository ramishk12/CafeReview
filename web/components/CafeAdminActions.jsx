'use client'

import { Pencil, Trash2 } from 'lucide-react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { closeCafe, deleteCafe, reopenCafe } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

// Admin-only controls on a cafe page: edit and delete. This is the place to add
// later admin actions such as close/reopen or a background image.
export default function CafeAdminActions({ cafeId, cafeName, closed = false }) {
  const { isAdmin } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const [deleting, setDeleting] = useState(false)
  const [toggling, setToggling] = useState(false)

  if (!isAdmin) return null

  const handleToggleClosed = async () => {
    setToggling(true)
    try {
      if (closed) {
        await reopenCafe(cafeId)
        toast.show(`${cafeName} reopened`)
      } else {
        await closeCafe(cafeId)
        toast.show(`${cafeName} closed`)
      }
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setToggling(false)
    }
  }

  const handleDelete = async () => {
    if (!window.confirm(`Delete ${cafeName}? Its reviews and photos are removed too.`)) return
    setDeleting(true)
    try {
      await deleteCafe(cafeId)
      toast.show(`${cafeName} deleted`)
      router.push('/')
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
      setDeleting(false)
    }
  }

  return (
    <div className="flex items-center gap-2">
      <Link href={`/cafes/${cafeId}/edit`} className="btn btn-ghost px-3.5 py-2 text-sm no-underline">
        <Pencil size={15} aria-hidden="true" /> Edit
      </Link>
      <button type="button" onClick={handleToggleClosed} disabled={toggling} className="btn btn-ghost px-3.5 py-2 text-sm">
        {toggling ? '…' : closed ? 'Reopen' : 'Close'}
      </button>
      <button type="button" onClick={handleDelete} disabled={deleting} className="btn btn-danger px-3.5 py-2 text-sm">
        <Trash2 size={15} aria-hidden="true" /> {deleting ? 'Deleting…' : 'Delete'}
      </button>
    </div>
  )
}
