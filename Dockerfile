# Use official Go image
FROM golang:1.21

# Install ffmpeg, python3, pip, and yt-dlp
RUN apt-get update && \
    apt-get install -y ffmpeg python3 python3-pip && \
    pip3 install yt-dlp

# Set the working directory
WORKDIR /app

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the app
COPY . .

# Build your Go app
RUN go build -o main .

# Expose port (matches what your app uses)
EXPOSE 8080

# Start your app
CMD ["./main"]
