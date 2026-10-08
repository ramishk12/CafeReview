import { useEffect, useState } from 'react'
import { checkHealth } from './services/api.js'

export default function App() {
  const [status, setStatus] = useState('checking...')

  useEffect(() => {
    checkHealth()
      .then((data) => setStatus(data.status))
      .catch(() => setStatus('backend unreachable'))
  }, [])

  return (
    <main className="app">
      <h1>CafeReview</h1>
      <p>Backend status: {status}</p>
    </main>
  )
}
