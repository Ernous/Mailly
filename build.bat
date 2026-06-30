@echo off
echo Building Frontend...
cd web
call bun install
call bun run build
cd ..

echo Copying frontend to go embed directory...
if exist cmd\server\dist rmdir /s /q cmd\server\dist
xcopy /E /I web\dist cmd\server\dist

echo Building Backend...
go build -o mailly.exe ./cmd/server

echo Build Complete! Run mailly.exe to start.
