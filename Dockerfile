# Stage 1: Build the Go binary
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Stage 2: Final runtime container
FROM debian:bookworm-slim
WORKDIR /app

# Install dependencies: ffmpeg, nodejs, python3, ca-certificates, and curl
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    nodejs \
    python3 \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Download latest yt-dlp binary
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

# Copy compiled Go server from builder
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]