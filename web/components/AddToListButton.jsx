'use client'

import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import {
  addCafeToList,
  addReviewToList,
  addUserToList,
  createList,
  listMyLists,
  removeCafeFromList,
  removeReviewFromList,
  removeUserFromList,
} from '@/lib/api'
import { useAuth } from '@/context/AuthContext'
import { useToast } from '@/context/ToastContext'

const IDEAS = ['Visited', 'Want to visit', 'Favourites', 'Trusted reviewers', 'Helpful reviews']

const CONFIG = {
  cafe: { add: addCafeToList, remove: removeCafeFromList, flag: 'contains_cafe' },
  user: { add: addUserToList, remove: removeUserFromList, flag: 'contains_user' },
  review: { add: addReviewToList, remove: removeReviewFromList, flag: 'contains_review' },
}

// Signed-in users save a cafe, reviewer, or review to their lists. Renders
// nothing when signed out. type is 'cafe', 'user', or 'review'.
export default function AddToListButton({ type = 'cafe', id, label = 'Save to list' }) {
  const { user } = useAuth()
  const toast = useToast()
  const [open, setOpen] = useState(false)
  const [lists, setLists] = useState([])
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const [coords, setCoords] = useState(null)
  const triggerRef = useRef(null)
  const panelRef = useRef(null)

  const config = CONFIG[type]
  const panelId = `add-to-list-${type}-${id}`

  // The panel is portaled to the body with fixed positioning. It used to be an
  // absolutely positioned child, but cards carry backdrop-blur, which creates a
  // stacking context, so following cards painted over it.
  useEffect(() => {
    if (!open) return
    listMyLists({ [type]: id })
      .then(setLists)
      .catch(() => {})

    const place = () => {
      const rect = triggerRef.current?.getBoundingClientRect()
      if (!rect) return
      if (rect.bottom < 0 || rect.top > window.innerHeight) {
        setOpen(false) // the trigger scrolled out of view
        return
      }
      setCoords({ top: rect.bottom + 8, right: Math.max(8, window.innerWidth - rect.right) })
    }
    place()

    // Keep the panel anchored while the page scrolls or the window resizes.
    let frame = 0
    const onViewportChange = () => {
      if (frame) return
      frame = requestAnimationFrame(() => {
        frame = 0
        place()
      })
    }
    const onPointerDown = (event) => {
      const target = event.target
      if (triggerRef.current?.contains(target) || panelRef.current?.contains(target)) return
      setOpen(false)
    }
    const onKeyDown = (event) => {
      if (event.key !== 'Escape') return
      setOpen(false)
      triggerRef.current?.focus()
    }

    document.addEventListener('pointerdown', onPointerDown)
    document.addEventListener('keydown', onKeyDown)
    window.addEventListener('scroll', onViewportChange, true)
    window.addEventListener('resize', onViewportChange)
    return () => {
      if (frame) cancelAnimationFrame(frame)
      document.removeEventListener('pointerdown', onPointerDown)
      document.removeEventListener('keydown', onKeyDown)
      window.removeEventListener('scroll', onViewportChange, true)
      window.removeEventListener('resize', onViewportChange)
    }
  }, [open, type, id])

  if (!user) return null

  const toggle = async (list) => {
    setBusy(true)
    try {
      if (list[config.flag]) {
        await config.remove(list.id, id)
        setLists((current) => current.map((l) => (l.id === list.id ? { ...l, [config.flag]: false } : l)))
      } else {
        await config.add(list.id, id)
        setLists((current) => current.map((l) => (l.id === list.id ? { ...l, [config.flag]: true } : l)))
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
      await config.add(created.id, id)
      setLists((current) => [{ ...created, [config.flag]: true }, ...current])
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
      <button
        ref={triggerRef}
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-expanded={open}
        aria-haspopup="true"
        aria-controls={panelId}
        className="btn btn-ghost px-3.5 py-2 text-sm"
      >
        {label}
      </button>
      {open &&
        coords &&
        createPortal(
          <div
            ref={panelRef}
            id={panelId}
            role="dialog"
            aria-label="Your lists"
            style={{ top: coords.top, right: coords.right }}
            className="fixed z-50 w-64 rounded-2xl border border-slate-200 bg-white p-4 text-left shadow-xl dark:border-white/10 dark:bg-slate-900"
          >
            <p className="mb-2 text-sm font-semibold">Your lists</p>
            {lists.length === 0 && <p className="mb-2 text-sm text-slate-500 dark:text-slate-400">No lists yet.</p>}
            <ul className="mb-3 grid gap-1">
              {lists.map((list) => (
                <li key={list.id}>
                  <label className="flex items-center gap-2 text-sm">
                    <input type="checkbox" checked={Boolean(list[config.flag])} disabled={busy} onChange={() => toggle(list)} />
                    {list.name}
                  </label>
                </li>
              ))}
            </ul>
            <form
              className="flex items-center gap-2"
              onSubmit={(event) => {
                event.preventDefault()
                create()
              }}
            >
              <input className="field py-1.5 text-sm" placeholder="New list name" value={name} onChange={(e) => setName(e.target.value)} />
              <button type="submit" disabled={busy || !name.trim()} className="btn px-3 py-1.5 text-sm">
                Add
              </button>
            </form>
            <p className="mt-2 text-xs text-slate-400">Ideas: {IDEAS.join(', ')}</p>
          </div>,
          document.body,
        )}
    </div>
  )
}
