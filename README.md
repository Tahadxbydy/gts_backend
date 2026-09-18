# GTS — Get That Song 🎵

A clean, lightweight Go backend that extracts audio from YouTube URLs via **yt-dlp** + **ffmpeg** and serves it to a Flutter mobile app over HTTP.

---

## Table of Contents

- [Overview](#overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [API Reference](#api-reference)
- [Architecture](#architecture)
- [Background Services](#background-services)
- [Development Notes](#development-notes)

---

## Overview

GTS accepts a YouTube URL from a Flutter client, runs `yt-dlp` to extract the audio track as an MP3, stores the file temporarily in a local `./storage` directory, and provides a download endpoint the app can stream from. Files are automatically cleaned up after **1 hour** by a background ticker.

---

## Tech Stack

| Layer            | Technology                                    |
| ---------------- | --------------------------------------------- |
| Language         | Go 1.22+                                      |
| HTTP Framework   | [Echo v4](https://echo.labstack.com/)         |
| Audio Extraction | [yt-dlp](https://github.com/yt-dlp/yt-dlp)    |
| Audio Processing | [ffmpeg](https://ffmpeg.org/)                 |
| Unique IDs       | [google/uuid](https://github.com/google/uuid) |

---

## Project Structure

```
gts_backend/
├── main.go               # Entry point — server init, graceful shutdown
├── go.mod / go.sum       # Go module files
├── models/
│   └── models.go         # Core data structs (AudioRequest, ExtractionResponse, AudioMetadata)
├── services/
│   ├── extractor.go      # yt-dlp extraction service (os/exec wrapper)
│   ├── storage.go        # Local file-system helpers
│   └── cleanup.go        # Background time.Ticker — deletes files > 1 hour old
├── handlers/
│   ├── extract.go        # POST /api/v1/extract
│   ├── download.go       # GET  /api/v1/download/:id
│   └── health.go         # GET  /health
├── routes/
│   └── router.go         # Central route registration
└── storage/              # Auto-created — extracted audio files live here (gitignored)
```

---

## Prerequisites

Install the following tools before running the server:

```bash
# macOS (Homebrew)
brew install yt-dlp ffmpeg

# Verify
yt-dlp --version
ffmpeg -version
```

Go 1.22 or later is required:

```bash
go version
```

---

## Getting Started

```bash
# 1. Clone the repo
git clone https://github.com/tahakhan/gts_backend.git
cd gts_backend

# 2. Download Go dependencies
go mod tidy

# 3. Build
go build -o gts .

# 4. Run
./gts
# or simply:
go run .
```

The server starts on **port 8080** by default.

---

## Configuration

| Environment Variable | Default | Description                         |
| -------------------- | ------- | ----------------------------------- |
| `PORT`               | `8080`  | TCP port the HTTP server listens on |

Set variables before running:

```bash
PORT=9000 ./gts
```

---

## API Reference

### Health Check

```
GET /health
```

**Response 200**

```json
{
  "status": "ok",
  "timestamp": "2026-09-19T04:22:00Z",
  "service": "gts-backend",
  "version": "1.0.0"
}
```

---

### Extract Audio

```
POST /api/v1/extract
Content-Type: application/json
```

**Request Body**

| Field    | Type   | Required | Description                                   |
| -------- | ------ | -------- | --------------------------------------------- |
| `url`    | string | ✅       | Full YouTube video URL                        |
| `format` | string | ❌       | Output format: `mp3` (default), `m4a`, `opus` |

**Example Request**

```bash
curl -X POST http://localhost:8080/api/v1/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "format": "mp3"}'
```

**Response 201**

```json
{
  "id": "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
  "title": "Rick Astley - Never Gonna Give You Up",
  "format": "mp3",
  "download_url": "/api/v1/download/3f2504e0-4f89-11d3-9a0c-0305e82c3301",
  "created_at": "2026-09-19T04:22:00Z"
}
```

**Error Responses**

| Status | Error Key           | Cause                    |
| ------ | ------------------- | ------------------------ |
| 400    | `invalid_request`   | Malformed JSON body      |
| 422    | `missing_url`       | `url` field not provided |
| 500    | `extraction_failed` | yt-dlp or ffmpeg error   |

---

### Download Audio

```
GET /api/v1/download/:id
```

Streams the extracted audio file. Supports **HTTP Range requests** (resumable downloads).

**Example**

```bash
curl -O http://localhost:8080/api/v1/download/3f2504e0-4f89-11d3-9a0c-0305e82c3301
```

**Response Headers**

```
Content-Disposition: attachment; filename="Rick Astley - Never Gonna Give You Up.mp3"
Content-Type: audio/mpeg
Accept-Ranges: bytes
```

**Error Responses**

| Status | Error Key    | Cause                            |
| ------ | ------------ | -------------------------------- |
| 400    | `missing_id` | No `:id` path parameter          |
| 404    | `not_found`  | ID not found or file has expired |

---

## Architecture

```
Flutter App
    │
    │  POST /api/v1/extract  { url, format }
    ▼
┌─────────────────────────────────────────────┐
│               Echo HTTP Server               │
│                                             │
│  handlers/extract.go                        │
│      └─► services/extractor.go             │
│              └─► yt-dlp (os/exec)          │
│                  └─► ffmpeg                 │
│                      └─► ./storage/<id>.mp3 │
│                                             │
│  handlers/download.go                       │
│      └─► http.ServeFile (range-aware)      │
└─────────────────────────────────────────────┘
    │
    │  GET /api/v1/download/:id
    ▼
Flutter App streams/saves the MP3
```

---

## Background Services

### File Cleanup (`services/cleanup.go`)

A `time.Ticker` goroutine fires every **10 minutes** and removes any audio file whose `created_at` timestamp is more than **1 hour** in the past.

```
Interval : 10 minutes
Max age  : 1 hour
```

The worker is started in `main.go` and shut down cleanly via a `done` channel when the server receives SIGINT/SIGTERM.

---

## Development Notes

- **Concurrency limit**: `ExtractionService` uses a buffered channel semaphore (`maxConcurrency = 5`) to cap simultaneous yt-dlp processes.
- **Blocking extraction**: The current `/extract` endpoint is synchronous — yt-dlp must finish before the HTTP response is sent. For long videos consider adding an async job queue with a status-poll endpoint.
- **Storage backend**: Files are currently stored on local disk. Swap `StorageService` for an S3-compatible implementation for horizontal scaling.
- **ffmpeg path**: The default `--ffmpeg-location` is set to `/usr/local/bin` (Homebrew default). Adjust in `services/extractor.go` if your ffmpeg lives elsewhere (e.g., `/opt/homebrew/bin` on Apple Silicon).

---

## License

MIT
