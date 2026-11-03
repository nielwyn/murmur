FROM oven/bun:1 AS frontend-builder
WORKDIR /src/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY sql/ ./sql/
COPY web/embed.go ./web/embed.go
COPY --from=frontend-builder /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o apiserver ./cmd/apiserver

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=builder /src/apiserver /apiserver
USER nonroot
EXPOSE 8080
ENTRYPOINT ["/apiserver"]
