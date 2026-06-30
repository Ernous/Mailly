# Build frontend
FROM oven/bun:1 AS frontend
WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN bun install --frozen-lockfile
COPY web/ ./
RUN bun run build

# Build backend
FROM golang:alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy frontend dist to cmd/server so go:embed can find it
COPY --from=frontend /app/web/dist ./cmd/server/dist
RUN CGO_ENABLED=0 go build -o mailly ./cmd/server

# Final image
FROM alpine:latest
WORKDIR /app
# Install ca-certificates for TLS (IMAP/OAuth)
RUN apk --no-cache add ca-certificates
COPY --from=backend /app/mailly .

EXPOSE 3000
CMD ["./mailly"]
