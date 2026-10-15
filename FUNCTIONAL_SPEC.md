# CafeReview - Functional Specification

A living description of what the app does today. Update it in the same pull
request whenever a feature is added or changed, so it stays a accurate picture of
current behaviour.

- **Status:** living document
- **Stack:** Go (Gin, PostgreSQL) API, Next.js (App Router, JavaScript) web app

## 1. Accounts and authentication

- **Register:** a user signs up with an email, password (at least 8 characters),
  and a display name. The account starts unverified, a confirmation email is
  sent, and the response is the same whether or not the email is already
  registered (no account enumeration).
- **Email verification:** the confirmation link (`/verify?token=...`) confirms
  the address. Links are single use and expire after 24 hours. A resend endpoint
  issues a new link.
- **Login:** email and password. A correct password on an unverified account is
  refused with a message to confirm the email. On success the server creates a
  session and sets an HttpOnly cookie.
- **Sessions:** stored server-side, sent as an HttpOnly `cafe_session` cookie,
  with a 30-day sliding expiry that is extended at most once per hour. Logout
  deletes the session. Expired sessions are swept periodically.
- **CSRF:** write requests must send the `X-CSRF-Token` header, copied from the
  readable `cafe_csrf` cookie.
- **Roles:** `user` and `admin`. The role is read from the database on each
  request, so a change takes effect immediately. Admins are created with the
  `createadmin` command.
- **Rate limits:** login is 5 requests per minute per IP and an account locks for
  15 minutes after 10 failed logins in an hour; registration and verification
  resends are 5 per hour per IP. Over the limit returns `429` with `Retry-After`.
- **Passwords:** hashed with bcrypt, never returned. There is no password reset
  yet.
- **Public profiles:** each user has a public page at `/users/:id` showing their
  display name, join date, review count, and their reviews (newest first, each
  linking to its cafe). Only public data is shown; the email is never exposed.
  Review author names link to profiles across the app.

## 2. Cafes

- **Browse:** list of cafes with search over name and address, tag filters,
  sorting by name, rating, number of reviews, or newest, and paging. The list is
  paged on the server (24 per page) and the view is driven by the URL
  (`q`, `sort`, `tags`, `page`), so views are shareable and the back button works.
  The total number of matches is returned in the `X-Total-Count` header.
- **Detail:** a cafe page shows its name, address, description, average rating,
  review count and reviewer count, its reviews and photos.
- **Admin only:** create, edit (name, address, description), and delete a cafe.
  Admins have an Edit, Close/Reopen, and Delete toolbar on the cafe page, an Edit
  action on the browse list cards, and a dedicated edit page. Deleting a cafe
  also deletes its reviews and their photo files. This admin surface is the base
  for later actions (such as setting a background image).
- **Closing (admin):** an admin can close or reopen a cafe. A closed cafe shows a
  "Closed" banner and a "Closed" badge on cards, accepts no new reviews, and its
  existing reviews can no longer be edited (they can still be deleted). Closing
  is reversible. Filtering closed cafes out of the list is planned to use tags
  once tagging exists.
- **Summary values:** the average rating is the mean of all reviews; the review
  count is the number of reviews (visits); the reviewer count is the number of
  distinct people who reviewed it. Labels read "N reviews from M people".
- **Tags:** cafes carry tags from a shared, admin-curated set grouped into
  categories (location, vibe, drinks). Tags show as chips on the cafe page and as
  filters on the browse list; filtering uses AND, so a cafe must have every
  selected tag. Admins edit a cafe's tags from the cafe page. The tag framework is
  shared by cafes, reviews, and images through a `scope` column, and a `pending`
  status is reserved for future user-suggested tags awaiting approval.
- **Photo tags:** review photos can be tagged from the image-scoped set (`Drink`,
  `Interior`, `Exterior`). The photo's owner tags their own photos; admins may tag
  any. A photo can have several tags.
- **Background image (admin):** an admin can upload a background image for a
  cafe (JPEG/PNG/WebP, up to 5 MB, sanitized like review photos). It is shown
  behind the cafe header with a dark overlay for readable text; without one the
  header keeps its colour gradient. Admins can replace or remove it.
- **Background from a photo (admin):** an admin can also choose one of the cafe's
  review photos as the background, optionally filtered by photo tag. A chosen
  photo wins over the uploaded image; clearing it falls back to the uploaded
  image (or the gradient).

## 3. Reviews

- **Writing:** a signed-in user writes a review with a star rating (1-5), text
  (up to 2000 characters), an optional visit date, and optional photos.
- **One review per day:** a user may leave at most one review for a given cafe on
  a given visit date. Another visit on another day can be reviewed separately.
  The visit date defaults to today and cannot be in the future.
- **Editing and deleting:** the author can edit or delete their own reviews.
  Editing can change the rating, text, and visit date. Deleting a review also
  deletes its photo files.
- **Listing:** reviews for a cafe are listed with sorting by newest, latest visit,
  highest rated, or lowest rated, and paging.
- **Photos:** each review has its own photos, shown on the review and in a
  per-cafe photo gallery.
- **Closed cafes:** no new reviews can be added, and existing reviews can no
  longer be edited. Existing reviews can still be deleted by their owner.
- **Review tags:** a review can be tagged by its author with admin-curated
  visit tags (for example `Seasonal menu`, `Limited edition`, `Event`). Tags
  appear on the review and can be set in the review form.
- **Filtering:** the review list on a cafe page can be filtered by minimum
  rating, reviews with photos, review tags, and the signed-in user's own
  reviews. Filters run in the browser and are kept in the URL so the view is
  shareable.

## 4. Lists

- Signed-in users create named lists (for example Visited, Want to visit,
  Favourites) and add cafes to them. A cafe can be in several lists.
- The "create list" screen suggests example names; nothing is created
  automatically.
- Lists are private. Public lists are designed for (`lists.is_public`) but not
  built yet.
- Limits: 50 lists per user and 500 cafes per list.
- The schema also has `list_reviews` and `list_users` for future lists of reviews
  and reviewers (issue #78).

## 5. Images

- **Accepted formats:** JPEG, PNG, and WebP, up to 5 MB per file. The type is
  confirmed by content sniffing, not the client header.
- **Sanitisation:** on upload, privacy metadata is stripped at the container
  level (EXIF, GPS, device details, comments, and similar) without re-encoding
  the pixels. Colour profiles and EXIF orientation are preserved, and any data
  appended after the end of the image is removed.
- **Storage:** files are stored under `uploads/` with random names; only paths
  are kept in the database. They are served at `/uploads/...`.

## 6. Security

- bcrypt password hashing; password hashes never leave the server.
- Server-side sessions (see section 1) and CSRF protection on writes.
- Per-IP and per-account rate limits on the authentication endpoints.
- Hardening headers on every API response: `X-Content-Type-Options: nosniff`,
  `Referrer-Policy`, `X-Frame-Options`, `Cross-Origin-Resource-Policy`, and a
  locked-down `Content-Security-Policy` on API and upload responses.
- The web app sends a report-only `Content-Security-Policy` with a per-request
  nonce (to be enforced at launch).
- Internal errors are logged server-side and returned to clients as a generic
  message with a request ID. Database error text is never exposed.
- Secrets and credentials come from the environment; the app refuses to start
  without a database password.

## 7. Email

- Transactional email is sent through a single `Sender` interface. Configuring
  SMTP uses that provider; leaving `SMTP_HOST` empty writes emails to the server
  log instead.
- Local development uses Mailpit (SMTP on 1025, web UI on 8025).
- Currently only account verification emails are sent.

## 8. Command-line tools

- `createadmin`: creates an admin user from `ADMIN_EMAIL`, `ADMIN_PASSWORD`, and
  `ADMIN_NAME`, or promotes an existing user. The account is created verified.

## 9. API

All endpoints are under `/api`. Errors return `{"error": "..."}`. Writes require
the session cookie and the `X-CSRF-Token` header; admin routes require an admin
account.

| Method | Path | Notes |
| --- | --- | --- |
| POST | `/auth/register` | Creates an unverified account and emails a link |
| POST | `/auth/login` | Blocked until the email is verified |
| POST | `/auth/logout` | Deletes the session |
| GET | `/auth/me` | Current user |
| GET | `/auth/verify` | Confirms an email with `?token=` |
| POST | `/auth/verify/resend` | Sends a new link |
| GET | `/cafes` | Search `q`, tag filter `tags` (comma-separated IDs, AND), `sort`, `limit`, `offset` |
| GET | `/cafes/:id` | |
| GET | `/tags` | Query: `scope` (`cafe`, `review`, `image`). Approved tags only |
| GET | `/users/:id` | Public profile: display name, join date, review count |
| GET | `/users/:id/reviews` | A user's reviews, newest first, with `cafe_name`. Paged |
| POST | `/cafes` | Admin |
| PUT | `/cafes/:id` | Admin |
| DELETE | `/cafes/:id` | Admin; also removes reviews and photos |
| POST | `/cafes/:id/close` | Admin |
| POST | `/cafes/:id/reopen` | Admin |
| PUT | `/cafes/:id/tags` | Admin. Body: `{"tag_ids": [...]}`, replaces the set |
| PUT | `/cafes/:id/background` | Admin. Multipart field `image` |
| DELETE | `/cafes/:id/background` | Admin |
| PUT | `/cafes/:id/background/review-image` | Admin. Body: `{"image_id": id}` |
| DELETE | `/cafes/:id/background/review-image` | Admin |
| PUT | `/review-images/:id/tags` | Owner or admin. Body: `{"tag_ids": [...]}` |
| GET | `/me/lists` | Signed-in. Query: `cafe=<id>` adds `contains_cafe` |
| POST | `/me/lists` | Signed-in. Body: `{"name": "..."}` |
| GET | `/lists/:id` | Owner. The list with its cafes |
| PUT | `/lists/:id` | Owner. Rename |
| DELETE | `/lists/:id` | Owner |
| POST | `/lists/:id/cafes` | Owner. Body: `{"cafe_id": id}` |
| DELETE | `/lists/:id/cafes/:cafeId` | Owner |
| GET | `/cafes/:id/reviews` | `sort`, `limit`, `offset` |
| POST | `/cafes/:id/reviews` | One per user per cafe per day. Body includes `visited_on` and optional `tag_ids` |
| PUT | `/reviews/:id` | Owner only. Body includes rating, body, `visited_on`, and `tag_ids` |
| DELETE | `/reviews/:id` | Owner only; also removes photos |
| POST | `/reviews/:id/images` | Owner only |
| DELETE | `/reviews/:id/images/:imageId` | Owner only |

## 10. Web pages

- `/` browse cafes (search, tag filters, sort, and pages; state is in the URL).
- `/cafes/:id` cafe detail, reviews (with sorting and filtering), photos, and the review form.
- `/cafes/new` admin: add a cafe.
- `/cafes/:id/edit` admin: edit an existing cafe.
- `/users/:id` public profile.
- `/login`, `/register`, `/verify` account pages.

## 11. Data model

- `users`: id, email (unique), display_name, password_hash, role, created_at,
  email_verified_at.
- `sessions`: id, token_hash (unique), user_id, csrf_token, created_at,
  expires_at.
- `email_tokens`: id, user_id, token_hash (unique), kind, expires_at, used_at,
  created_at.
- `cafes`: id, name, address, description, created_at, closed_at,
  background_image_path.
- `reviews`: id, cafe_id, user_id, rating (1-5), body, visited_on, created_at,
  updated_at. Unique on (user_id, cafe_id, visited_on).
- `review_images`: id, review_id, filename, created_at.
- `tags`: id, name, category, scope (`cafe`/`review`/`image`), status
  (`approved`/`pending`), created_at. Unique on (scope, name).
- `cafe_tags`: cafe_id, tag_id.
- `review_image_tags`: image_id, tag_id.
- `review_tags`: review_id, tag_id.
- `lists`: id, user_id, name, is_public, created_at.
- `list_cafes`, `list_reviews`, `list_users`: list_id + the item id, added_at.

## 12. Not built yet

Tracked as issues: city search and maps (#16), social sign-in (#33), account settings (#45), admin moderation (#35), reporting reviews (#40), helpful votes (#41), review edit history (#42), review drafts (#43), lists of reviewers and reviews (#78), following cafes and email (#46), similar-cafe warning (#36), loading and empty states (#39), and the launch work: image storage (#9), Docker (#48), enforcing the CSP (#62), structured logging (#49), and accessibility (#50).
closing and archiving cafes, user lists, following cafes, social sign-in,
account settings, public profiles, review reports and moderation, helpful votes,
review edit history, review drafts, maps integration, Dockerfiles, structured
logging, and the accessibility audit.
