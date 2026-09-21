# Update base image to Go 1.25 or latest alpine builder
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gts-server main.go

# Production runtime stage
FROM alpine:3.19

# Install runtime dependencies: ffmpeg, nodejs (JS runtime solver), python3, ca-certificates, and curl
RUN apk add --no-cache \
    ffmpeg \
    nodejs \
    python3 \
    ca-certificates \
    curl

# Download the latest yt-dlp binary into /usr/local/bin
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/gts-server .
RUN mkdir -p /app/storage

EXPOSE 8080
CMD ["./gts-server"]