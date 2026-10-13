import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

afterEach(() => {
  cleanup()
  // Clear any cookies a test set.
  for (const name of ['cafe_csrf', 'cafe_session']) {
    document.cookie = `${name}=; Max-Age=0; path=/`
  }
})
