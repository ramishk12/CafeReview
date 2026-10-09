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
## API

All endpoints are under `/api`. Errors return `{"error": "..."}`. Writes need `Authorization: Bearer <token>`, and admin routes need an admin token.

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/cafes` | Query: `q` (search name or address, max 100 chars), `sort` (`name`, `rating`, `reviews`, `newest`; default `name`), `limit` (1-100), `offset` |
| GET | `/cafes/:id` | |
| POST | `/cafes` | Admin |
| PUT | `/cafes/:id` | Admin. Replaces name, address, and description |
| DELETE | `/cafes/:id` | Admin. Also deletes the cafe's reviews and photo files |
| GET | `/cafes/:id/reviews` | Query: `sort` (`newest`, `highest`, `lowest`; default `newest`), `limit` (1-100), `offset` |
| POST | `/cafes/:id/reviews` | One review per user per cafe |
| PUT | `/reviews/:id` | Owner only |
| DELETE | `/reviews/:id` | Owner only. Also deletes the review's photo files |
| POST | `/reviews/:id/images` | Owner only. Multipart field `image`: JPEG, PNG, or WebP, up to 5 MB |
| DELETE | `/reviews/:id/images/:imageId` | Owner only. Removes one photo |

**Paging:** without `limit`, the full list is returned. With `limit`, the response is one page. Either way, the `X-Total-Count` response header gives the number of matching rows.

## Tests
```bash
go test ./...
```
