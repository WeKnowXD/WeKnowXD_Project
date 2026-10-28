# Sæt samme version som i din go.mod (linjen der starter med "go ")
ARG GO_VERSION=1.27.1

# ---------- Trin 1: byg programmet ----------
# Debian-baseret image, fordi go-sqlite3 kræver cgo og dermed gcc (det følger med her)
FROM golang:${GO_VERSION}-bookworm AS build
WORKDIR /src

# Kopiér kun go.mod/go.sum først, så dependencies caches mellem builds
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download

COPY src/backend/ ./
RUN CGO_ENABLED=1 go build -tags fts5 -o /out/server .

# ---------- Trin 2: lille runtime-image ----------
FROM debian:bookworm-slim

# Uden ca-certificates fejler HTTPS-kaldet til weatherapi.com (x509-fejl)
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Programmet bruger relative stier (templates/, static/, ../whoknows.db),
# så arbejdsmappen skal ligne den, du kører go run fra
WORKDIR /app/backend
COPY --from=build /out/server ./server
COPY src/backend/templates ./templates
COPY src/backend/static ./static

EXPOSE 8081
CMD ["./server"]