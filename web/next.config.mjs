// The Go API runs separately. Requests to /api and /uploads are proxied to it so the
// browser stays on one origin (no CORS, and the stored token works everywhere).
// Rewrites are read at build time, so rebuild after changing BACKEND_URL.
const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080'

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      { source: '/api/:path*', destination: `${BACKEND_URL}/api/:path*` },
      { source: '/uploads/:path*', destination: `${BACKEND_URL}/uploads/:path*` },
    ]
  },
}

export default nextConfig
