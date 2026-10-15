'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { clearCafeBackgroundImage, deleteCafeBackground, listTags, setCafeBackground, setCafeBackgroundImage } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

// Admin editor for a cafe's background. There are two sources: an uploaded image
// and a chosen review photo. A chosen review photo wins; removing it falls back
// to the uploaded image. Renders nothing for non-admins.
export default function CafeBackgroundEditor({ cafeId, backgroundUrl, backgroundSource = '', images = [] }) {
  const { isAdmin } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const [uploading, setUploading] = useState(false)
  const [removingUpload, setRemovingUpload] = useState(false)
  const [choosing, setChoosing] = useState(0)
  const [clearingChosen, setClearingChosen] = useState(false)
  const [imageTags, setImageTags] = useState([])
  const [tagFilter, setTagFilter] = useState(null)

  useEffect(() => {
    if (!isAdmin) return
    listTags('image')
      .then(setImageTags)
      .catch(() => {})
  }, [isAdmin])

  if (!isAdmin) return null

  const run = async (fn, message, setBusy) => {
    setBusy(true)
    try {
      await fn()
      toast.show(message)
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setBusy(false)
    }
  }

  const handleFile = (e) => {
    const file = e.target.files?.[0]
    e.target.value = ''
    if (!file) return
    run(() => setCafeBackground(cafeId, file), 'Background uploaded', setUploading)
  }

  const filteredImages = tagFilter === null ? images : images.filter((img) => (img.tags || []).some((t) => t.id === tagFilter))
  const sourceLabel = backgroundSource === 'review_image' ? 'a chosen review photo' : backgroundSource === 'uploaded' ? 'an uploaded image' : ''

  return (
    <section className="card mb-10 p-6">
      <h2 className="mb-4 text-lg font-semibold tracking-tight">Background image</h2>

      {backgroundUrl ? (
        <figure className="mb-3">
          <img src={backgroundUrl} alt="Cafe background" className="h-32 w-full rounded-2xl object-cover" />
          {sourceLabel && <figcaption className="mt-1 text-xs text-slate-400">Using {sourceLabel}</figcaption>}
        </figure>
      ) : (
        <p className="mb-3 text-sm text-slate-500 dark:text-slate-400">No background image set.</p>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <label className="btn cursor-pointer">
          {uploading ? 'Uploading…' : 'Upload image'}
          <input type="file" accept="image/jpeg,image/png,image/webp" className="sr-only" onChange={handleFile} disabled={uploading} />
        </label>
        {backgroundSource === 'uploaded' && (
          <button
            type="button"
            onClick={() => run(() => deleteCafeBackground(cafeId), 'Background removed', setRemovingUpload)}
            disabled={removingUpload}
            className="btn btn-danger"
          >
            {removingUpload ? 'Removing…' : 'Remove uploaded'}
          </button>
        )}
        {backgroundSource === 'review_image' && (
          <button
            type="button"
            onClick={() => run(() => clearCafeBackgroundImage(cafeId), 'Cleared chosen photo', setClearingChosen)}
            disabled={clearingChosen}
            className="btn btn-ghost"
          >
            {clearingChosen ? 'Clearing…' : 'Clear chosen photo'}
          </button>
        )}
      </div>
      <p className="mt-2 text-xs text-slate-400">JPEG, PNG, or WebP · up to 5 MB</p>

      {images.length > 0 && (
        <div className="mt-6 border-t border-slate-200/70 pt-5 dark:border-white/5">
          <h3 className="mb-3 text-sm font-semibold">Choose from this cafe's photos</h3>
          {imageTags.length > 0 && (
            <div className="mb-3 flex flex-wrap gap-2">
              {imageTags.map((tag) => (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => setTagFilter((current) => (current === tag.id ? null : tag.id))}
                  aria-pressed={tagFilter === tag.id}
                  className={`rounded-full px-3 py-1 text-xs ${
                    tagFilter === tag.id ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200'
                  }`}
                >
                  {tag.name}
                </button>
              ))}
            </div>
          )}
          <ul className="grid grid-cols-3 gap-3 sm:grid-cols-5">
            {filteredImages.map((img) => (
              <li key={img.id}>
                <button
                  type="button"
                  onClick={() => run(() => setCafeBackgroundImage(cafeId, img.id), 'Background set', setChoosing)}
                  className="block w-full cursor-pointer overflow-hidden rounded-xl ring-1 ring-slate-900/10"
                  aria-label="Use this photo as the background"
                >
                  <img src={img.url} alt="" loading="lazy" className="h-24 w-full object-cover" />
                </button>
              </li>
            ))}
          </ul>
          {filteredImages.length === 0 && <p className="text-sm text-slate-500 dark:text-slate-400">No photos with that tag.</p>}
        </div>
      )}
    </section>
  )
}
