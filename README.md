# Murmur

A self-hosted, multi-user RSS/Atom feed reader — a Go backend with a Svelte
frontend, built as a deep dive into **Go concurrency** (goroutines, channels,
`context`, `sync`) and designed to run on a Raspberry Pi.

> A *murmuration* is a starling flock: thousands of individual birds moving as
> one shape — like many feeds aggregated into a single reading stream.

**Status: work in progress.** Auth, feed management, and the concurrent
background fetcher are done; the posts/reading UI is next. See
[Roadmap](#roadmap).

## Highlights

- **Concurrent feed fetcher** (`internal/feedfetch`) — the centerpiece. A
  scheduler ticks on an interval, queries feeds due for a refresh, and fans
  them out over a jobs channel to a **bounded worker pool** (default 5
  workers), with results fanned back in through a single collector goroutine.
  Each fetch runs under its own `context.WithTimeout`, so one hanging feed
  can't stall a worker.
- **Graceful shutdown** — `signal.NotifyContext` (SIGINT/SIGTERM) propagates
  through the HTTP server *and* every in-flight fetch: jobs channel closes,
  workers drain, collector finishes. No leaked goroutines.
- **Zero-dependency HTTP layer** — Go 1.22+ stdlib `net/http.ServeMux` with
  method + pattern routing; custom `logger`, `recoverer`, and `requireAuth`
  middleware in ~40 lines of plain Go. No router framework.
- **SQL-first data access** — hand-written SQL compiled to type-safe Go with
  [sqlc](https://sqlc.dev), `pgx/v5` + `pgxpool` as the driver, and
  [goose](https://github.com/pressly/goose) migrations. No ORM.
- **Stateless auth** — bcrypt password hashing, JWT (HS256) in an httpOnly,
  SameSite=Lax cookie. Same-origin SPA setup means no CORS complexity.

## Stack

| Layer     | Choice                                            |
| --------- | ------------------------------------------------- |
| Backend   | Go (stdlib `net/http`)                            |
| Database  | PostgreSQL — sqlc, pgx/v5, goose migrations       |
| Auth      | bcrypt + JWT in httpOnly cookie                   |
| Frontend  | Svelte 5 (runes) + Vite SPA, dev proxy to the API |
| Deploy    | Docker Compose; target: self-hosted Raspberry Pi 5 |

## API

| Method        | Path                     | Auth |
| ------------- | ------------------------ | ---- |
| POST          | `/api/register`          | –    |
| POST          | `/api/login`             | –    |
| POST          | `/api/logout`            | –    |
| GET           | `/api/me`                | ✅   |
| GET / POST    | `/api/feeds`             | ✅   |
| GET           | `/api/feeds/following`   | ✅   |
| POST / DELETE | `/api/feeds/{id}/follow` | ✅   |

## Getting started

Requires Go 1.26+, Node 20+, and Docker (or a local Postgres).

```sh
# 1. Start Postgres
docker compose up -d

# 2. Run migrations
goose -dir sql/schema postgres "postgres://murmur:murmur@localhost:5432/murmur?sslmode=disable" up

# 3. Configure (~/.murmurconfig.json)
#    { "db_url": "postgres://murmur:murmur@localhost:5432/murmur?sslmode=disable",
#      "jwt_secret": "change-me" }

# 4. Run the API server (port 8080, override with MURMUR_PORT)
go run ./cmd/apiserver

# 5. Run the frontend dev server (proxies /api to :8080)
cd web && npm install && npm run dev
```

## Roadmap

- [x] Auth, feeds/follows REST API
- [x] Concurrent fetcher: scheduler + bounded worker pool + graceful shutdown
- [x] Svelte frontend: auth + feed management
- [ ] Fetch-status endpoint (`RWMutex` status map) + per-host rate limiting
- [ ] Posts API + reading UI (pagination, read tracking)
- [ ] Structured logging (`log/slog`), env-based config
- [ ] Single-binary deploy: `embed.FS` frontend build, Docker image, Pi 5
