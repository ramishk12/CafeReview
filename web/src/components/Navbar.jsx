import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext.jsx'

export default function Navbar() {
  const { user, isAdmin, logout } = useAuth()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/')
  }

  return (
    <header className="navbar">
      <Link to="/" className="brand">CafeReview</Link>
      <nav>
        <Link to="/">Cafes</Link>
        {isAdmin && <Link to="/cafes/new">Add cafe</Link>}
        {user ? (
          <>
            <span className="muted">{user.display_name}</span>
            <button type="button" className="link-button" onClick={handleLogout}>
              Log out
            </button>
          </>
        ) : (
          <>
            <Link to="/login">Log in</Link>
            <Link to="/register">Sign up</Link>
          </>
        )}
      </nav>
    </header>
  )
}
