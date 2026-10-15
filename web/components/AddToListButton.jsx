'use client'

import { useEffect, useState } from 'react'
import { addCafeToList, createList, listMyLists, removeCafeFromList } from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

const IDEAS = ['Visited', 'Want to visit', 'Favourites']

// Signed-in users can add a cafe to their lists from its page. Renders nothing
// when signed out.
export default function AddToListButton({ cafeId }) {
  const { user } = useAuth()
  const toast = useToast()
  const [open, setOpen] = useState(false)
  const [lists, setLists] = useState([])
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!open) return
    listMyLists(cafeId)
      .then(setLists)
      .catch(() => {})
  }, [open, cafeId])

  if (!user) return null

  const toggle = async (list) => {
    setBusy(true)
    try {
      if (list.contains_cafe) {
        await removeCafeFromList(list.id, cafeId)
        setLists((current) => current.map((l) => (l.id === list.id ? { ...l, contains_cafe: false, cafe_count: l.cafe_count - 1 } : l)))
      } else {
        await addCafeToList(list.id, cafeId)
        setLists((current) => current.map((l) => (l.id === list.id ? { ...l, contains_cafe: true, cafe_count: l.cafe_count + 1 } : l)))
      }
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setBusy(false)
    }
  }

  const create = async () => {
    if (!name.trim()) return
    setBusy(true)
    try {
      const created = await createList(name.trim())
      await addCafeToList(created.id, cafeId)
      setLists((current) => [{ ...created, contains_cafe: true, cafe_count: 1 }, ...current])
      setName('')
      toast.show('List created')
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="relative">
      <button type="button" onClick={() => setOpen((o) => !o)} className="btn btn-ghost px-3.5 py-2 text-sm">
        Save to list
      </button>
      {open && (
        <div className="absolute right-0 z-20 mt-2 w-64 rounded-2xl border border-slate-200 bg-white p-4 text-left shadow-xl dark:border-white/10 dark:bg-slate-900">
          <p className="mb-2 text-sm font-semibold">Your lists</p>
          {lists.length === 0 && <p className="mb-2 text-sm text-slate-500 dark:text-slate-400">No lists yet.</p>}
          <ul className="mb-3 grid gap-1">
            {lists.map((list) => (
              <li key={list.id}>
                <label className="flex items-center gap-2 text-sm">
                  <input type="checkbox" checked={list.contains_cafe} disabled={busy} onChange={() => toggle(list)} />
                  {list.name}
                </label>
              </li>
            ))}
          </ul>
          <div className="flex items-center gap-2">
            <input className="field py-1.5 text-sm" placeholder="New list name" value={name} onChange={(e) => setName(e.target.value)} />
            <button type="button" onClick={create} disabled={busy || !name.trim()} className="btn px-3 py-1.5 text-sm">
              Add
            </button>
          </div>
          <p className="mt-2 text-xs text-slate-400">Ideas: {IDEAS.join(', ')}</p>
        </div>
      )}
    </div>
  )
}
