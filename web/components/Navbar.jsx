'use client'

import { Coffee, LogOut, Menu, X } from 'lucide-react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useAuth } from '@/context/AuthContext'
import Avatar from './Avatar'

export default function Navbar() {
  const { user, isAdmin, logout } = useAuth()
  const pathname = usePathname()
  const router = useRouter()
  const [menuOpen, setMenuOpen] = useState(false)

  // Close the mobile menu whenever the route changes.
  useEffect(() => setMenuOpen(false), [pathname])

  const handleLogout = () => {
    logout()
    router.push('/')
  }

  const links = [
    { href: '/', label: 'Cafes' },
    ...(isAdmin ? [{ href: '/cafes/new', label: 'Add cafe' }] : []),
  ]

  const isActive = (href) => (href === '/' ? pathname === '/' : pathname.startsWith(href))

  const linkClass = (href) =>
    `rounded-full px-3.5 py-2 text-sm font-medium no-underline transition ${
      isActive(href)
        ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300'
        : 'text-slate-600 hover:bg-slate-900/5 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white'
    }`

  return (
    <header className="sticky top-0 z-30 border-b border-slate-200/70 bg-white/70 backdrop-blur-xl dark:border-white/5 dark:bg-slate-950/70">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-5 py-3">
        <Link href="/" className="flex items-center gap-2 text-lg font-bold tracking-tight text-slate-900 no-underline dark:text-white">
          <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-linear-to-br from-indigo-500 to-fuchsia-500 text-white shadow-md shadow-indigo-500/30">
            <Coffee size={17} strokeWidth={2.25} />
          </span>
          CafeReview
        </Link>

        {/* Desktop navigation */}
        <nav aria-label="Main" className="hidden items-center gap-1 md:flex">
          {links.map((l) => (
            <Link key={l.href} href={l.href} className={linkClass(l.href)} aria-current={isActive(l.href) ? 'page' : undefined}>
              {l.label}
            </Link>
          ))}
          {user ? (
            <div className="ml-2 flex items-center gap-2 border-l border-slate-200 pl-3 dark:border-white/10">
              <Avatar name={user.display_name} size="sm" />
              <span className="text-sm font-medium text-slate-700 dark:text-slate-200">{user.display_name}</span>
              <button type="button" onClick={handleLogout} className="btn btn-ghost rounded-full p-2" title="Log out">
                <LogOut size={16} />
                <span className="sr-only">Log out</span>
              </button>
            </div>
          ) : (
            <>
              <Link href="/login" className={linkClass('/login')}>
                Log in
              </Link>
              <Link
                href="/register"
                className="ml-1 rounded-full bg-slate-900 px-4 py-2 text-sm font-semibold text-white no-underline transition hover:bg-slate-700 dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
              >
                Sign up
              </Link>
            </>
          )}
        </nav>

        {/* Mobile menu toggle */}
        <button
          type="button"
          onClick={() => setMenuOpen((o) => !o)}
          aria-expanded={menuOpen}
          aria-controls="mobile-menu"
          className="btn btn-ghost rounded-full p-2 md:hidden"
        >
          {menuOpen ? <X size={20} /> : <Menu size={20} />}
          <span className="sr-only">{menuOpen ? 'Close menu' : 'Open menu'}</span>
        </button>
      </div>

      {menuOpen && (
        <nav id="mobile-menu" aria-label="Mobile" className="border-t border-slate-200/70 px-5 py-3 md:hidden dark:border-white/5">
          <div className="flex flex-col gap-1">
            {links.map((l) => (
              <Link key={l.href} href={l.href} className={linkClass(l.href)}>
                {l.label}
              </Link>
            ))}
            {user ? (
              <div className="mt-2 flex items-center justify-between border-t border-slate-200/70 pt-3 dark:border-white/5">
                <span className="flex items-center gap-2 text-sm font-medium">
                  <Avatar name={user.display_name} size="sm" />
                  {user.display_name}
                </span>
                <button type="button" onClick={handleLogout} className="btn btn-ghost rounded-full px-3 py-1.5 text-sm">
                  <LogOut size={14} /> Log out
                </button>
              </div>
            ) : (
              <>
                <Link href="/login" className={linkClass('/login')}>
                  Log in
                </Link>
                <Link href="/register" className={linkClass('/register')}>
                  Sign up
                </Link>
              </>
            )}
          </div>
        </nav>
      )}
    </header>
  )
}
