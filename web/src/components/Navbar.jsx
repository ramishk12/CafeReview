import { Coffee, LogOut } from 'lucide-react'
import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'
import Avatar from './Avatar.jsx'

const navLinkClass = ({ isActive }) =>
  `rounded-full px-3.5 py-2 text-sm font-medium transition hover:bg-slate-900/5 dark:hover:bg-white/10 ${
    isActive ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-500/15 dark:text-indigo-300' : 'text-slate-600 hover:text-slate-900 dark:text-slate-300 dark:hover:text-white'
  }`

export default function Navbar() {
  const { user, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <header className="sticky top-0 z-20 border-b border-slate-200/70 bg-white/70 backdrop-blur-xl dark:border-white/5 dark:bg-slate-950/70">
      <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-5 py-3">
        <Link to="/" className="flex items-center gap-2 text-lg font-bold tracking-tight text-slate-900 no-underline dark:text-white">
          <span className="flex h-8 w-8 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-fuchsia-500 text-white shadow-md shadow-indigo-500/30">
            <Coffee size={17} strokeWidth={2.25} />
          </span>
          CafeReview
        </Link>

        <nav className="flex items-center gap-1">
          <NavLink to="/" end className={navLinkClass}>
            Cafes
          </NavLink>
          {isAdmin && (
            <NavLink to="/cafes/new" className={navLinkClass}>
              Add cafe
            </NavLink>
          )}
          {user ? (
            <div className="ml-2 flex items-center gap-2 border-l border-slate-200 pl-3 dark:border-white/10">
              <Avatar name={user.display_name} size="sm" />
              <span className="hidden text-sm font-medium text-slate-700 sm:inline dark:text-slate-200">{user.display_name}</span>
              <button
                type="button"
                onClick={handleLogout}
                title="Log out"
                className="btn btn-ghost rounded-full p-2"
              >
                <LogOut size={16} />
                <span className="sr-only">Log out</span>
              </button>
            </div>
          ) : (
            <>
              <NavLink to="/login" className={navLinkClass}>
                Log in
              </NavLink>
              <Link
                to="/register"
                className="ml-1 rounded-full bg-slate-900 px-4 py-2 text-sm font-semibold text-white no-underline transition hover:bg-slate-700 dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200"
              >
                Sign up
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}
