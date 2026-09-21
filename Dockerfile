FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM debian:bookworm-slim
WORKDIR /app

# Install dependencies natively supported in Debian
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    nodejs \
    python3 \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Download latest yt-dlp binary directly to system PATH
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]