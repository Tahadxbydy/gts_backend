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

# Install runtime binaries (ffmpeg, python3, yt-dlp)
RUN apk add --no-cache ffmpeg python3 ca-certificates curl
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/gts-server .
RUN mkdir -p /app/storage

EXPOSE 8080
CMD ["./gts-server"]