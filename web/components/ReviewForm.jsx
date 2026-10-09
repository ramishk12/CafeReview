'use client'

import { ImagePlus, X } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useEffect, useRef, useState } from 'react'
import { createReview, updateReview, uploadReviewImage } from '@/lib/api'
import { useToast } from '@/context/ToastContext'
import StarPicker from './StarPicker'

const MAX_BODY = 2000
const MAX_IMAGE_BYTES = 5 * 1024 * 1024
const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/webp']

// Creates a review, or edits the current user's review when `existing` is set.
// New photos are uploaded one at a time after the review itself is saved.
export default function ReviewForm({ cafeId, existing }) {
  const router = useRouter()
  const toast = useToast()
  const [rating, setRating] = useState(existing?.rating ?? 5)
  const [body, setBody] = useState(existing?.body ?? '')
  const [files, setFiles] = useState([]) // { file, url }
  const [error, setError] = useState('')
  const [progress, setProgress] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Preview URLs are revoked when removed, after submit, or on unmount.
  const filesRef = useRef(files)
  useEffect(() => {
    filesRef.current = files
  }, [files])
  useEffect(() => () => filesRef.current.forEach((f) => URL.revokeObjectURL(f.url)), [])

  const addFiles = (list) => {
    for (const file of list) {
      if (!ACCEPTED_TYPES.includes(file.type)) {
        setError(`${file.name} is not a JPEG, PNG, or WebP image`)
        return
      }
      if (file.size > MAX_IMAGE_BYTES) {
        setError(`${file.name} is larger than 5 MB`)
        return
      }
    }
    setError('')
    const picked = list.map((file) => ({ file, url: URL.createObjectURL(file) }))
    setFiles((current) => [...current, ...picked])
  }

  const removeFile = (index) => {
    URL.revokeObjectURL(files[index].url)
    setFiles((current) => current.filter((_, i) => i !== index))
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setSubmitting(true)

    let review
    try {
      const payload = { rating, body: body.trim() }
      review = existing ? await updateReview(existing.id, payload) : await createReview(cafeId, payload)
    } catch (err) {
      setError(err.message)
      setSubmitting(false)
      return
    }

    const failed = []
    for (const [i, item] of files.entries()) {
      setProgress(`Uploading photo ${i + 1} of ${files.length}…`)
      try {
        await uploadReviewImage(review.id, item.file)
      } catch (err) {
        failed.push(`${item.file.name}: ${err.message}`)
      }
    }

    files.forEach((f) => URL.revokeObjectURL(f.url))
    setProgress('')
    setSubmitting(false)
    if (failed.length > 0) {
      setError(`Review saved, but some photos failed. ${failed.join('; ')}`)
      toast.show('Review saved with some photo errors', 'error')
    } else {
      toast.show(existing ? 'Review updated' : 'Review posted')
    }
    setFiles([])
    router.refresh()
  }

  return (
    <form onSubmit={handleSubmit} className="card grid gap-5 p-6 sm:p-7">
      <div className="flex items-center justify-between gap-3">
        <h3 className="text-lg font-semibold tracking-tight">{existing ? 'Edit your review' : 'Write a review'}</h3>
      </div>

      <div className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-600 dark:text-slate-400">Rating</span>
        <StarPicker value={rating} onChange={setRating} />
      </div>

      <label className="label">
        Your review
        <textarea
          className="field min-h-28 resize-y"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          maxLength={MAX_BODY}
          rows={4}
          required
          placeholder="What did you order? How was the atmosphere?"
        />
        <span className="text-right text-xs font-normal text-slate-400">
          {body.length}/{MAX_BODY}
        </span>
      </label>

      <div className="grid gap-2">
        <span className="text-sm font-medium text-slate-600 dark:text-slate-400">Photos (optional)</span>
        <label className="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed border-slate-300 bg-slate-50/60 px-4 py-6 text-center text-sm text-slate-500 transition hover:border-indigo-400 hover:bg-indigo-50/40 dark:border-white/10 dark:bg-white/[0.02] dark:text-slate-400 dark:hover:border-indigo-500/50">
          <ImagePlus size={22} aria-hidden="true" />
          <span>
            <span className="font-semibold text-indigo-600 dark:text-indigo-400">Choose photos</span> or drop them here
          </span>
          <span className="text-xs">JPEG, PNG, or WebP · up to 5 MB each</span>
          <input
            type="file"
            accept={ACCEPTED_TYPES.join(',')}
            multiple
            className="sr-only"
            onChange={(e) => {
              addFiles(Array.from(e.target.files || []))
              e.target.value = ''
            }}
          />
        </label>

        {files.length > 0 && (
          <ul className="flex flex-wrap gap-3">
            {files.map((item, i) => (
              <li key={item.url} className="relative">
                <img src={item.url} alt={item.file.name} className="h-24 w-24 rounded-xl object-cover ring-1 ring-slate-900/10" />
                <button
                  type="button"
                  onClick={() => removeFile(i)}
                  className="btn btn-ghost absolute -top-2 -right-2 rounded-full bg-white p-1 text-slate-700 shadow-md dark:bg-slate-800 dark:text-slate-200"
                  aria-label={`Remove ${item.file.name}`}
                >
                  <X size={14} />
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>

      {error && <p className="alert-error">{error}</p>}

      <div className="flex flex-wrap items-center gap-4">
        <button type="submit" disabled={submitting || body.trim() === ''} className="btn">
          {submitting ? progress || 'Saving…' : existing ? 'Save changes' : 'Post review'}
        </button>
      </div>
    </form>
  )
}
