# Agent Guidelines for CafeReview

## Project Overview
Cafe review app with user login, picture uploads, and reviews.
Go backend (Gin + PostgreSQL) and React frontend (Vite) in `web/`.

## Commands
```bash
go test ./...        # backend tests
go vet ./...         # lint
go fmt ./...         # format
go build ./...       # build
cd web && npm run dev    # frontend dev server
cd web && npm run build  # frontend build
docker compose up -d     # start Postgres
```

## Conventions
- Follow the Go conventions from FoodOrderTracking: camelCase functions, PascalCase types, snake_case DB tables.
- Handlers: parse and validate input, query, map `sql.ErrNoRows` to 404, other errors to 500.
- Passwords: bcrypt only. Never store or return password hashes.
- Uploads: validate MIME type and size server-side; store files under `uploads/`, keep only paths in the DB.
- Frontend: pages in `web/src/pages/`, components in `web/src/components/`, HTTP calls in `web/src/services/api.js`, styles in `web/src/index.css`.

## Git
- Branches: `feature/{description}`, `fix/{description}`, `test/{description}`.
- Base PRs on `main`; run all tests before pushing.
