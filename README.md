# Audio Scrapper

A Go web service for extracting audio from video URLs using yt-dlp and ffmpeg. The extracted audio files are automatically named with the original video title.

## Project Structure

```
audio_scrapper/
├── main.go                 # Entry point - initializes and starts the server
├── handlers/               # HTTP request handlers
│   └── audio_handler.go    # Handles audio extraction requests
├── routes/                 # Route definitions
│   └── routes.go          # Sets up HTTP routes and middleware
├── services/               # Business logic layer
│   └── audio_service.go    # Audio scraping and processing logic
├── utils/                  # Utility functions
│   └── file_utils.go       # File operations and validation
├── models/                 # Data models
│   └── audio_req_model.go  # Request/response data structures
└── output/                 # Generated audio files
```

## Architecture

### Separation of Concerns

1. **Handlers** (`handlers/`): Handle HTTP requests and responses

   - Validate input
   - Call appropriate services
   - Format responses

2. **Routes** (`routes/`): Define URL patterns and middleware

   - Map URLs to handlers
   - Set up static file serving
   - Configure middleware

3. **Services** (`services/`): Business logic

   - Core application functionality
   - External tool integration (yt-dlp, ffmpeg)
   - Video title extraction
   - Data processing

4. **Utils** (`utils/`): Helper functions

   - File operations
   - Validation utilities
   - Common functionality

5. **Models** (`models/`): Data structures
   - Request/response models
   - Database models (if applicable)

## Usage

### Prerequisites

- Go 1.16+
- yt-dlp
- ffmpeg

### Running the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

### API Endpoints

- **POST** `/extract-audio` - Extract audio from a video URL

  - Request body: `{"url": "https://www.youtube.com/watch?v=VIDEO_ID"}`
  - Response:
    ```json
    {
      "audio_url": "/audio/Video_Title.mp3",
      "video_title": "Video_Title",
      "filename": "Video_Title.mp3"
    }
    ```

- **GET** `/audio/filename.mp3` - Download extracted audio file

### Features

- **Automatic Video Title Extraction**: Audio files are named using the original YouTube video title
- **Safe Filename Generation**: Invalid characters are automatically sanitized for filesystem compatibility
- **Unique Naming**: Prevents filename conflicts by cleaning special characters and spaces

## Benefits of This Structure

1. **Maintainability**: Each component has a single responsibility
2. **Testability**: Easy to unit test individual components
3. **Scalability**: Easy to add new features and endpoints
4. **Reusability**: Services can be reused across different handlers
5. **Clean Code**: Clear separation makes the codebase easier to understand

## Adding New Features

1. **New Handler**: Create a new file in `handlers/`
2. **New Route**: Add to `routes/routes.go`
3. **New Service**: Create a new file in `services/`
4. **New Model**: Add to `models/` directory
5. **New Utility**: Add to `utils/` directory
