import { Link, NavLink, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'

const navLinkClass = ({ isActive }) =>
  `rounded-full px-3 py-1.5 text-sm font-medium transition hover:bg-indigo-50 hover:text-indigo-600 dark:hover:bg-slate-800 dark:hover:text-indigo-300 ${
    isActive ? 'text-indigo-600 dark:text-indigo-300' : 'text-slate-700 dark:text-slate-200'
  }`

export default function Navbar() {
  const { user, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <header className="sticky top-0 z-10 border-b border-slate-200 bg-white/80 backdrop-blur-md dark:border-slate-800 dark:bg-slate-950/80">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-5 py-3">
        <Link
          to="/"
          className="bg-gradient-to-r from-indigo-600 to-pink-500 bg-clip-text text-lg font-extrabold tracking-tight text-transparent"
        >
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
            <>
              <span className="px-3 text-sm text-slate-500 dark:text-slate-400">{user.display_name}</span>
              <button type="button" onClick={handleLogout} className="btn btn-ghost rounded-full px-3 py-1.5 text-sm">
                Log out
              </button>
            </>
          ) : (
            <>
              <NavLink to="/login" className={navLinkClass}>
                Log in
              </NavLink>
              <NavLink to="/register" className={navLinkClass}>
                Sign up
              </NavLink>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}
