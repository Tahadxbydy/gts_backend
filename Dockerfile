FROM golang:1.21-alpine

# Set working dir
WORKDIR /app

# Copy static binaries
COPY bin/yt-dlp /usr/local/bin/yt-dlp
COPY bin/ffmpeg /usr/local/bin/ffmpeg
RUN chmod +x /usr/local/bin/yt-dlp /usr/local/bin/ffmpeg

# Copy Go project
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build Go app
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
