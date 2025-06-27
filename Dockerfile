FROM golang:1.21-alpine

# Install system deps
RUN apk add --no-cache ffmpeg python3 py3-pip && \
    pip3 install yt-dlp

# Set working dir
WORKDIR /app

# Download Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build Go app
RUN go build -o main .

# Expose app port
EXPOSE 8080

# Start server
CMD ["./main"]
