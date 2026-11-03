# Murmur

RSS/Atom feed reader — Go backend, Svelte frontend

## Highlights

- **Concurrent feed fetcher** (`internal/feedfetch`) — scheduler → bounded
  worker pool → collector, each fetch under its own `context.WithTimeout`.
- **Graceful shutdown** — `signal.NotifyContext` drains the HTTP server and
  every in-flight fetch on SIGINT/SIGTERM.
- **Zero-dependency HTTP layer** — stdlib `net/http.ServeMux`, no router.
- **SQL-first** — sqlc + pgx/v5 + goose migrations, no ORM.
- **Stateless auth** — bcrypt + JWT in an httpOnly cookie, plus Google OAuth.
- **Single binary** — the Svelte build is embedded via `embed.FS`, and the
  server applies its own migrations at startup. No goose binary on the host.

## Run it with Docker

```sh
cp .env.example .env   # set JWT_SECRET and POSTGRES_PASSWORD
docker compose up -d --build
```

Serves the UI and the API on <http://localhost:8080>. The server migrates the
database itself on boot, so there is no separate migration step.

## Local development

```sh
cp .env.example .env            # set JWT_SECRET; DB_URL defaults to localhost
cd web && bun install && bun run build && cd ..
go run ./cmd/apiserver          # :8080, API + embedded UI
cd web && bun run dev           # :5173, HMR, proxies /api to :8080
```

The one-off `bun run build` matters: `web/embed.go` embeds `web/dist`, so Go
cannot compile without that directory. Only a `.gitkeep` is committed, which
is enough to build but serves an empty UI until you build the frontend.

Config is entirely env-driven (loaded from `.env` via `godotenv`, or the
process environment in production) — see `internal/config/config.go` for
every variable and its default.

```sh
go test ./... -race              # unit + httptest concurrency tests
goose -dir sql/schema postgres "$DB_URL" status   # optional, for inspection
sqlc generate                    # regenerate internal/database after SQL edits
```

## Deploying

The image is a single container (distroless, non-root) plus Postgres, so
`compose.yaml` is the whole deployment. On the target host:

```sh
docker compose up -d --build
```

Build on the host when its architecture differs from your dev machine
(a Raspberry Pi is arm64), or cross-build with
`docker buildx build --platform linux/arm64`.

To expose it publicly without opening a port or renting a static IP, run a
Cloudflare Tunnel pointed at `localhost:8080`.

## Roadmap

- [x] Auth, feeds/follows, posts REST API
- [x] Concurrent fetcher: scheduler + bounded worker pool + graceful shutdown
- [x] Svelte frontend: auth, feeds, posts
- [x] Read tracking + pagination
- [x] Env-based config
- [x] Single-binary deploy: embedded frontend, self-applying migrations, Docker
- [ ] Fetch-status endpoint + per-host rate limiting
- [ ] Structured logging (`log/slog`)
- [ ] OPML import
