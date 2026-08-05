# Murmur

RSS/Atom feed reader — Go backend, Svelte frontend

## Highlights

- **Concurrent feed fetcher** (`internal/feedfetch`) — scheduler → bounded
  worker pool → collector, each fetch under its own `context.WithTimeout`.
- **Graceful shutdown** — `signal.NotifyContext` drains the HTTP server and
  every in-flight fetch on SIGINT/SIGTERM.
- **Zero-dependency HTTP layer** — stdlib `net/http.ServeMux`, no router.
- **SQL-first** — sqlc + pgx/v5 + goose migrations, no ORM.
- **Stateless auth** — bcrypt + JWT in an httpOnly cookie.

## Getting started

Requires Go 1.26+, Bun, and Docker (or a local Postgres).

```sh
docker compose up -d
goose -dir sql/schema postgres "postgres://user:pass@localhost:5432/murmur?sslmode=disable" up
cp .env.example .env  # fill in DB_URL, JWT_SECRET, etc.
go run ./cmd/apiserver
cd web && bun install && bun run dev
```

Config is entirely env-driven (loaded from `.env` via `godotenv`, or the
process environment in production) — see `internal/config/config.go` for
every variable and its default.

## Roadmap

- [x] Auth, feeds/follows, posts REST API
- [x] Concurrent fetcher: scheduler + bounded worker pool + graceful shutdown
- [x] Svelte frontend: auth, feeds, posts
- [ ] Read tracking + pagination
- [ ] Fetch-status endpoint + per-host rate limiting
- [x] Env-based config
- [ ] Structured logging (`log/slog`)
- [ ] Single-binary deploy: embed frontend, Docker image
