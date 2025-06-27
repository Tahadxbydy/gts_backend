FROM golang:1.21-bullseye

# Install ffmpeg, python3, pip, and yt-dlp
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ffmpeg python3 python3-pip && \
    pip3 install yt-dlp && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set workdir
WORKDIR /app

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy app code
COPY . .

# Build the app
RUN go build -o main .

# Expose app port
EXPOSE 8080

# Run the app
CMD ["./main"]
