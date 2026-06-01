# Multi-stage build for the skip-migrator backend.
# Stage 1: build the Go binaries (API + migrate-cli).
# Stage 2: runtime image with Node.js 22 + pnpm so the worker can run `pnpm install/build`
#          on user-uploaded projects.

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Copy the migrator module first (it's a local replace target)
COPY migrator/go.mod migrator/go.sum* ./migrator/
COPY backend/go.mod backend/go.sum* ./backend/
WORKDIR /src/backend
RUN go mod download

# Copy sources
COPY migrator /src/migrator
COPY backend /src/backend

# Build both binaries (CGO off → static, alpine-compatible)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' \
      -o /out/server ./cmd/server
RUN cd /src/migrator && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' \
      -o /out/migrate-cli ./cmd/migrate-cli

# ---------------------------------------------------------------------------
FROM node:22-alpine AS runtime
RUN apk add --no-cache git ca-certificates tzdata \
 && corepack enable \
 && corepack prepare pnpm@10.12.3 --activate

WORKDIR /app
COPY --from=builder /out/server      /usr/local/bin/server
COPY --from=builder /out/migrate-cli /usr/local/bin/migrate-cli

ENV PORT=8000
EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/server"]
