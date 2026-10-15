'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { deleteCafeBackground, setCafeBackground } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

// Admin editor for a cafe's background image. Renders nothing for non-admins.
export default function CafeBackgroundEditor({ cafeId, backgroundUrl }) {
  const { isAdmin } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const [uploading, setUploading] = useState(false)
  const [removing, setRemoving] = useState(false)

  if (!isAdmin) return null

  const handleFile = async (e) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    setUploading(true)
    try {
      await setCafeBackground(cafeId, file)
      toast.show('Background updated')
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setUploading(false)
    }
  }

  const handleRemove = async () => {
    setRemoving(true)
    try {
      await deleteCafeBackground(cafeId)
      toast.show('Background removed')
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setRemoving(false)
    }
  }

  return (
    <section className="card mb-10 p-6">
      <h2 className="mb-4 text-lg font-semibold tracking-tight">Background image</h2>
      {backgroundUrl ? (
        <img src={backgroundUrl} alt="Cafe background" className="mb-3 h-32 w-full rounded-2xl object-cover" />
      ) : (
        <p className="mb-3 text-sm text-slate-500 dark:text-slate-400">No background image set.</p>
      )}
      <div className="flex flex-wrap items-center gap-3">
        <label className="btn cursor-pointer">
          {uploading ? 'Uploading…' : 'Choose image'}
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            className="sr-only"
            onChange={handleFile}
            disabled={uploading}
          />
        </label>
        {backgroundUrl && (
          <button type="button" onClick={handleRemove} disabled={removing} className="btn btn-danger">
            {removing ? 'Removing…' : 'Remove'}
          </button>
        )}
      </div>
      <p className="mt-2 text-xs text-slate-400">JPEG, PNG, or WebP · up to 5 MB</p>
    </section>
  )
}
