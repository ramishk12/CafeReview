# CafeReview

Cafe reviews with user accounts and picture uploads.

## Stack
- Backend: Go, Gin, PostgreSQL
- Frontend: Next.js (App Router), React, Tailwind CSS

## Setup
```bash
# Database (seeds sample cafes on first start; create an admin with cmd/createadmin)
docker compose up -d

# Re-seed from scratch (deletes all database data)
docker compose down -v && docker compose up -d

# Admin account (set ADMIN_EMAIL and ADMIN_PASSWORD in .env first)
go run ./cmd/createadmin

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

All endpoints are under `/api`. Errors return `{"error": "..."}`. Signing in sets an HttpOnly session cookie. Write requests also need the `X-CSRF-Token` header, copied from the `cafe_csrf` cookie. Admin routes require an admin account. `POST /auth/logout` deletes the session.

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

**Rate limits:** login allows 5 requests per minute per IP, and an account is locked for 15 minutes after 10 failed logins within an hour. Registration allows 5 per hour per IP. Over the limit returns `429` with a `Retry-After` header. The client IP is taken from the direct connection unless `TRUSTED_PROXIES` lists the proxies to trust.

## Production checklist
Before running anywhere other than your machine:

- Set `COOKIE_SECURE=true` (the default) so session cookies are only sent over HTTPS. Local http development sets it false.
- Set `DB_PASSWORD`. The server and `docker compose` both refuse to start without it, and the repo ships no default.
- Change any credentials that came from an example. Never keep the defaults.
- Set `TRUSTED_PROXIES` to the address of the proxy or frontend server that forwards requests, so the client IP is correct. Leave it empty only when the API is reached directly.
- Keep the database off the public network. `docker-compose.yml` binds it to `127.0.0.1`; keep that or place it on a private network.
- Do not load demo or seed data in production. `db/init` seeds sample cafes only; the sample accounts have been removed.
- Serve over HTTPS, and see the follow-up to enforce the Content-Security-Policy (#62).

## Tests
```bash
# Backend (unit tests; database calls are mocked with sqlmock)
go test ./...
go vet ./...

# Frontend (Vitest + Testing Library, jsdom)
cd web
npm test
```
