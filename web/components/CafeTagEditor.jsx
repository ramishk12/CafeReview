'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { listTags, setCafeTags } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

// Admin editor for a cafe's tags. Loads the approved cafe tags and saves the
// selected set. Renders nothing for non-admins.
export default function CafeTagEditor({ cafeId, initialTagIds = [] }) {
  const { isAdmin } = useAuth()
  const toast = useToast()
  const router = useRouter()
  const [tags, setTags] = useState([])
  const [selected, setSelected] = useState(() => new Set(initialTagIds))
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!isAdmin) return
    listTags('cafe')
      .then(setTags)
      .catch(() => {})
  }, [isAdmin])

  if (!isAdmin) return null

  const toggle = (id) =>
    setSelected((current) => {
      const next = new Set(current)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  const save = async () => {
    setSaving(true)
    try {
      await setCafeTags(cafeId, [...selected])
      toast.show('Tags saved')
      router.refresh()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setSaving(false)
    }
  }

  const groups = {}
  for (const tag of tags) (groups[tag.category] ||= []).push(tag)

  return (
    <section className="card mb-10 p-6">
      <h2 className="mb-4 text-lg font-semibold tracking-tight">Tags</h2>
      {tags.length === 0 && <p className="text-sm text-slate-500 dark:text-slate-400">No tags available yet.</p>}
      {Object.entries(groups).map(([category, list]) => (
        <div key={category} className="mb-3">
          <p className="mb-1 text-xs font-semibold tracking-wide text-slate-400 uppercase">{category}</p>
          <div className="flex flex-wrap gap-2">
            {list.map((tag) => {
              const active = selected.has(tag.id)
              return (
                <button
                  key={tag.id}
                  type="button"
                  onClick={() => toggle(tag.id)}
                  aria-pressed={active}
                  className={`rounded-full px-3 py-1 text-sm transition ${
                    active
                      ? 'bg-indigo-600 text-white'
                      : 'bg-slate-100 text-slate-700 hover:bg-slate-200 dark:bg-white/10 dark:text-slate-200 dark:hover:bg-white/20'
                  }`}
                >
                  {tag.name}
                </button>
              )
            })}
          </div>
        </div>
      ))}
      <button type="button" onClick={save} disabled={saving} className="btn mt-3">
        {saving ? 'Saving…' : 'Save tags'}
      </button>
    </section>
  )
}
