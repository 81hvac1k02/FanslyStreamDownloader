# Dockerfile.build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -ldflags="-s -w" -trimpath -o FanslyStreamDownloader

### Step 2: Create a Dockerfile for the final image with FFmpeg and the Go binary

# Dockerfile
FROM linuxserver/ffmpeg:version-7.1-cli

COPY --from=builder /app/FanslyStreamDownloader /usr/local/bin/FanslyStreamDownloader
WORKDIR /app

ENTRYPOINT  ["FanslyStreamDownloader"]