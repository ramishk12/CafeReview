'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { listTags, setReviewImageTags } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

// Lets the owner tag a photo (drink, interior, exterior) and saves immediately.
export default function ReviewImageTagger({ image }) {
  const toast = useToast()
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [tags, setTags] = useState([])
  const [selected, setSelected] = useState(() => new Set((image.tags ?? []).map((t) => t.id)))
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open && tags.length === 0) {
      listTags('image')
        .then(setTags)
        .catch(() => {})
    }
  }, [open, tags.length])

  const toggle = async (id) => {
    const next = new Set(selected)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    const previous = selected
    setSelected(next)
    setSaving(true)
    try {
      await setReviewImageTags(image.id, [...next])
      router.refresh()
    } catch (err) {
      setSelected(previous)
      toast.show(err.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const names = new Map(tags.map((t) => [t.id, t.name]))

  return (
    <div className="mt-1.5 w-32 sm:w-36">
      <button type="button" onClick={() => setOpen((o) => !o)} className="text-xs font-semibold text-indigo-600 dark:text-indigo-400">
        {open ? 'Done' : 'Tag photo'}
      </button>
      {open && (
        <div className="mt-1 flex flex-wrap gap-1">
          {tags.map((tag) => (
            <button
              key={tag.id}
              type="button"
              disabled={saving}
              onClick={() => toggle(tag.id)}
              aria-pressed={selected.has(tag.id)}
              className={`rounded-full px-2 py-0.5 text-xs ${
                selected.has(tag.id) ? 'bg-indigo-600 text-white' : 'bg-slate-100 text-slate-700 dark:bg-white/10 dark:text-slate-200'
              }`}
            >
              {names.get(tag.id)}
            </button>
          ))}
          {tags.length === 0 && <span className="text-xs text-slate-400">Loading…</span>}
        </div>
      )}
    </div>
  )
}
