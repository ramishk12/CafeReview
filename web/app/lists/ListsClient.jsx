'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { createList, deleteList, listMyLists, renameList } from '@/lib/api'
import { useToast } from '@/context/ToastContext'

const IDEAS = ['Visited', 'Want to visit', 'Favourites']

export default function ListsClient() {
  const toast = useToast()
  const [lists, setLists] = useState([])
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)

  const load = () =>
    listMyLists()
      .then(setLists)
      .catch(() => {})

  useEffect(() => {
    load()
  }, [])

  const create = async () => {
    if (!name.trim()) return
    setBusy(true)
    try {
      await createList(name.trim())
      setName('')
      load()
    } catch (err) {
      toast.show(err.message, 'error')
    } finally {
      setBusy(false)
    }
  }

  const rename = async (list) => {
    const next = window.prompt('Rename list', list.name)
    if (!next || next === list.name) return
    try {
      await renameList(list.id, next)
      load()
    } catch (err) {
      toast.show(err.message, 'error')
    }
  }

  const remove = async (list) => {
    if (!window.confirm(`Delete "${list.name}"?`)) return
    try {
      await deleteList(list.id)
      toast.show('List deleted')
      load()
    } catch (err) {
      toast.show(err.message, 'error')
    }
  }

  return (
    <section className="mx-auto mt-8 max-w-2xl">
      <h1 className="mb-6 text-3xl font-bold tracking-tight">My lists</h1>

      <div className="card mb-8 p-6">
        <h2 className="mb-3 text-lg font-semibold tracking-tight">Create a list</h2>
        <div className="flex flex-wrap items-center gap-2">
          <input className="field flex-1" placeholder="List name" value={name} onChange={(e) => setName(e.target.value)} />
          <button type="button" onClick={create} disabled={busy || !name.trim()} className="btn">
            Create
          </button>
        </div>
        <div className="mt-3 flex flex-wrap gap-2">
          {IDEAS.map((idea) => (
            <button
              key={idea}
              type="button"
              onClick={() => setName(idea)}
              className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-700 dark:bg-white/10 dark:text-slate-200"
            >
              {idea}
            </button>
          ))}
        </div>
      </div>

      {lists.length === 0 ? (
        <p className="card p-6 text-slate-500 dark:text-slate-400">No lists yet. Create one above.</p>
      ) : (
        <ul className="grid gap-3">
          {lists.map((list) => (
            <li key={list.id} className="card flex items-center justify-between gap-3 p-5">
              <Link href={`/lists/${list.id}`} className="font-semibold no-underline">
                {list.name} <span className="text-sm font-normal text-slate-400">({list.cafe_count})</span>
              </Link>
              <div className="flex items-center gap-2">
                <button type="button" onClick={() => rename(list)} className="btn btn-ghost px-3 py-1.5 text-sm">
                  Rename
                </button>
                <button type="button" onClick={() => remove(list)} className="btn btn-danger px-3 py-1.5 text-sm">
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
