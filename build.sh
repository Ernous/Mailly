#!/usr/bin/env bash
set -e

echo "Building Frontend..."
cd web
bun install
bun run build
cd ..

echo "Copying frontend to go embed directory..."
rm -rf cmd/server/dist
cp -r web/dist cmd/server/dist

echo "Building Backend..."
# Detect OS and set output binary name
if [[ "$OSTYPE" == "darwin"* ]]; then
    OUTPUT="mailly"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OUTPUT="mailly"
else
    OUTPUT="mailly"
fi

go build -o "$OUTPUT" ./cmd/server

echo "Build complete! Run ./$OUTPUT to start."
