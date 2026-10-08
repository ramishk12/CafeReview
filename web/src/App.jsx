import { Route, Routes } from 'react-router-dom'
import Navbar from './components/Navbar.jsx'
import RequireAuth from './components/RequireAuth.jsx'
import CafeListPage from './pages/CafeListPage.jsx'
import CafeDetailPage from './pages/CafeDetailPage.jsx'
import LoginPage from './pages/LoginPage.jsx'
import RegisterPage from './pages/RegisterPage.jsx'
import NewCafePage from './pages/NewCafePage.jsx'

export default function App() {
  return (
    <>
      <Navbar />
      <main className="app">
        <Routes>
          <Route path="/" element={<CafeListPage />} />
          <Route path="/cafes/:id" element={<CafeDetailPage />} />
          <Route
            path="/cafes/new"
            element={
              <RequireAuth adminOnly>
                <NewCafePage />
              </RequireAuth>
            }
          />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="*" element={<p className="status">Page not found.</p>} />
        </Routes>
      </main>
    </>
  )
}
