# Multi-stage build for Go + yt-dlp + ffmpeg runtime
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency definitions
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o gts-server main.go

# Production runtime stage
FROM alpine:3.19

# Install system dependencies: ffmpeg, python3 (required by yt-dlp), ca-certificates, and curl
RUN apk add --no-cache ffmpeg python3 ca-certificates curl

# Install yt-dlp CLI binary directly to system PATH
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/gts-server .

# Create persistent storage folder for downloads
RUN mkdir -p /app/storage

# Expose default HTTP server port
EXPOSE 8080

CMD ["./gts-server"]