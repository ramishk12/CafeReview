# CafeReview

Cafe reviews with user accounts and picture uploads.

## Stack
- Backend: Go, Gin, PostgreSQL
- Frontend: Next.js (App Router), React, Tailwind CSS

## Setup
```bash
# Database (seeds sample cafes, users, and reviews on first start)
docker compose up -d

# Re-seed from scratch (deletes all database data)
docker compose down -v && docker compose up -d

# Backend (port 8080)
go run ./cmd/server

# Frontend (port 3000; /api and /uploads are proxied to the backend)
cd web
npm install
npm run dev
```

To point the frontend at a different API address, set `BACKEND_URL` before starting. It is read when the app is built, so rebuild after changing it:

```bash
BACKEND_URL=http://localhost:8080 npm run build && npm run start
```

For a production build, run `npm run build` and then `npm run start`.

## Tests
```bash
go test ./...
```
