# Mailly

A lightweight self-hosted email client built with Go and Vue.js.

## Getting Started

### Using Docker (Recommended)

```bash
docker-compose up -d
```

Available at `http://localhost:3000`.

---

### Manual Build

You need: **Go 1.21+**, **Bun**

#### Linux / macOS

```bash
chmod +x build.sh
./build.sh
```

Then run:
```bash
./mailly
```

#### Windows

```bat
build.bat
```

Then run:
```bat
mailly.exe
```

#### Step by step (without scripts)

```bash
# 1. Build frontend
cd web
bun install
bun run build
cd ..

# 2. Copy dist to Go embed directory
rm -rf cmd/server/dist
cp -r web/dist cmd/server/dist

# 3. Build Go binary
go build -o mailly ./cmd/server   # Linux/macOS
go build -o mailly.exe ./cmd/server  # Windows
```

#### Cross-compile for Linux from any OS

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o mailly-linux ./cmd/server
```

#### Cross-compile for Windows from Linux/macOS

```bash
GOOS=windows GOARCH=amd64 go build -o mailly.exe ./cmd/server
```

---

### Configuration

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

Key variables:
- `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` — for Gmail OAuth
- `MICROSOFT_CLIENT_ID` / `MICROSOFT_CLIENT_SECRET` — for Outlook OAuth
- `SESSION_SECRET` — random secret for session encryption
- `PORT` — server port (default: 3000)
