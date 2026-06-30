# Mailly

A lightweight email client built with Go and Vue.js.

## Getting Started

### Using Docker (Recommended)
1. Provide your OAuth credentials in the `docker-compose.yml` file.
2. Run the application:
```bash
docker-compose up -d
```
The application will be available at `http://localhost:3000`.

### Manual Build
To build the application manually into a single executable, use the provided build script. You'll need `go` and `bun` installed.

**Windows:**
```bat
build.bat
```

This will:
1. Build the frontend (Vue) and copy it to the Go server directory.
2. Embed the frontend inside the Go binary.
3. Output a single `mailly.exe` binary.

**Configuration:**
Create a `.env` file based on `.env.example` and fill in your OAuth credentials (Google/Microsoft) and other settings.

Run the binary:
```bat
mailly.exe
```