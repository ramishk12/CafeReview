# Agent Guidelines for CafeReview

## Project Overview
Cafe review app with user login, picture uploads, and reviews.
Go backend (Gin + PostgreSQL) and Next.js frontend (App Router, JavaScript, Tailwind CSS) in `web/`.

## Commands
```bash
go test ./...        # backend tests
go vet ./...         # lint
go fmt ./...         # format
go build ./...       # build
cd web && npm run dev    # frontend dev server (http://localhost:3000)
cd web && npm run build  # frontend production build
cd web && npm run start  # serve the production build
docker compose up -d     # start Postgres
```

## Conventions
- Follow the Go conventions from FoodOrderTracking: camelCase functions, PascalCase types, snake_case DB tables.
- Handlers: parse and validate input, query, map `sql.ErrNoRows` to 404, other errors to 500.
- Passwords: bcrypt only. Never store or return password hashes.
- Uploads: validate MIME type and size server-side; store files under `uploads/`, keep only paths in the DB.
- Frontend (Next.js App Router, JavaScript):
  - Routes in `web/app/` (`page.jsx`, `loading.jsx`, `error.jsx`, `not-found.jsx`); route-only helpers can sit beside them.
  - Shared UI in `web/components/`, React context in `web/context/`, helpers in `web/lib/`.
  - Server Components read data with `lib/serverApi.js`. Client Components call the API through `lib/api.js`. Never import `serverApi.js` from a Client Component.
  - Mark components that use state, effects, or event handlers with `'use client'`.
  - In route handlers and pages, `params` is a Promise; `await` it.
  - Styles: Tailwind utilities, with shared component classes in `web/app/globals.css`.
  - The browser talks to `/api` and `/uploads` on the frontend origin; `next.config.mjs` proxies them to the Go API (`BACKEND_URL`, default `http://localhost:8080`).

## Git
- Branches: `feature/{description}`, `fix/{description}`, `test/{description}`.
- Base PRs on `main`; run all tests before pushing.
