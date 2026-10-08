# CafeReview

Cafe reviews with user accounts and picture uploads.

## Stack
- Backend: Go, Gin, PostgreSQL
- Frontend: React + Vite

## Setup
```bash
# Database
docker compose up -d

# Backend (port 8080)
go run ./cmd/server

# Frontend (port 5173, proxies /api to the backend)
cd web
npm install
npm run dev
```

## Tests
```bash
go test ./...
```
