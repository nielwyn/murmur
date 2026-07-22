# Murmur

A self-hosted, multi-user RSS/Atom feed reader — Go backend, Svelte frontend,
built to learn Go concurrency (goroutines, channels, context, sync).

> A *murmuration* is a starling flock moving as one shape — like many feeds
> aggregated into one reading stream.

**Status: work in progress.** Auth, feeds, and posts are up; read tracking
and pagination are next.

## Highlights

- **Concurrent feed fetcher** (`internal/feedfetch`) — scheduler → bounded
  worker pool → collector, each fetch under its own `context.WithTimeout`.
- **Graceful shutdown** — `signal.NotifyContext` drains the HTTP server and
  every in-flight fetch on SIGINT/SIGTERM.
- **Zero-dependency HTTP layer** — stdlib `net/http.ServeMux`, no router.
- **SQL-first** — sqlc + pgx/v5 + goose migrations, no ORM.
- **Stateless auth** — bcrypt + JWT in an httpOnly cookie.

## Stack

| Layer    | Choice                                 |
| -------- | --------------------------------------- |
| Backend  | Go, stdlib `net/http`                   |
| Database | PostgreSQL — sqlc, pgx/v5, goose         |
| Auth     | bcrypt + JWT in httpOnly cookie         |
| Frontend | Svelte 5 (runes) + Vite, bun            |
| Deploy   | Docker Compose, self-hosted             |

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
| GET           | `/api/posts`             | ✅   |

## Getting started

Requires Go 1.26+, Bun, and Docker (or a local Postgres).

```sh
docker compose up -d
goose -dir sql/schema postgres "postgres://user:pass@localhost:5432/murmur?sslmode=disable" up
go run ./cmd/apiserver
cd web && bun install && bun run dev
```

## Roadmap

- [x] Auth, feeds/follows, posts REST API
- [x] Concurrent fetcher: scheduler + bounded worker pool + graceful shutdown
- [x] Svelte frontend: auth, feeds, posts
- [ ] Read tracking + pagination
- [ ] Fetch-status endpoint + per-host rate limiting
- [ ] Structured logging, env-based config
- [ ] Single-binary deploy: embed frontend, Docker image
