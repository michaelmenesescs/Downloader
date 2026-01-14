# Downloader

A Go-based media downloader and playlist manager for DJs, featuring Elasticsearch backend for advanced playlist management.

## Features

### 🎵 Multi-Platform Support
Download media from:
- **Tidal** (requires username and password)
- **SoundCloud**
- **YouTube**

### 🎛️ Dual Operation Modes

**1. Interactive Mode** - Simple CLI for quick downloads
**2. Server Mode** - Full-featured API server with:
- Playlist management
- Track metadata storage
- Download queue system
- Background workers
- Elasticsearch-powered search

### 🗄️ Playlist Management
- Create and manage playlists
- Add/remove tracks
- Search tracks by artist, title, or album
- Track download status
- Export playlists to USB for DJing

### ⚡ Download Queue
- Background download workers
- Track download progress
- Automatic retry on failure
- Batch playlist downloads

## Architecture

```
┌─────────────────┐
│   Client/API    │
└────────┬────────┘
         │
         v
┌─────────────────┐     ┌──────────────────┐
│   REST API      │────>│  Elasticsearch   │
│   (Go/Mux)      │     │  (Storage)       │
└────────┬────────┘     └──────────────────┘
         │
         v
┌─────────────────┐     ┌──────────────────┐
│ Download Manager│────>│  Download Queue  │
│  (Workers)      │     │  (Background)    │
└────────┬────────┘     └──────────────────┘
         │
         v
┌─────────────────────────────────┐
│  External Downloaders           │
│  - tidal-dl (Tidal)             │
│  - scdl (SoundCloud)            │
│  - ytmdl (YouTube)              │
└─────────────────────────────────┘
```

## Prerequisites

- **Go 1.21+** (for building)
- **Docker & Docker Compose** (recommended for deployment)
- **Elasticsearch 8.x** (for server mode)

External downloaders (included in Docker image):
- `tidal-dl` - Tidal Media Downloader
- `scdl` - SoundCloud Downloader
- `ytmdl` - YouTube Music Downloader
- `ffmpeg` - Audio processing

## Building

### Local Build

```bash
go build -o downloader main.go
```

### Docker Build

```bash
docker build -t downloader .
```

## Testing

Run all tests:

```bash
go test -v
```

Run tests with coverage:

```bash
go test -cover
```

Generate detailed coverage report:

```bash
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
```

View coverage in browser:

```bash
go tool cover -html=coverage.out
```

### Test Coverage

The test suite covers:
- ✅ URL pattern matching for all services (100% coverage)
- ✅ Download function logic with mocked command execution (100% coverage)
- ✅ ServiceType enum and string representation (100% coverage)
- ✅ Error handling for all download operations

## Usage

### Interactive Mode (Simple Downloads)

Run without flags for interactive CLI:

```bash
./downloader
```

Or with Docker:

```bash
docker run -it -v $(pwd)/downloads:/app/downloads downloader
```

The application will prompt you for a URL and automatically detect the service.

### Server Mode (Full Playlist Manager)

#### Using Docker Compose (Recommended)

```bash
# Start Elasticsearch and API server
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f downloader-api

# Stop services
docker-compose down
```

#### Local Server Mode

```bash
# Start Elasticsearch first (required)
# Then run:
./downloader -mode=server -es-url=http://localhost:9200 -api-addr=:8080
```

### Command-Line Flags

```
-mode string
    Run mode: 'interactive' or 'server' (default "interactive")

-es-url string
    Elasticsearch URL (default "http://localhost:9200")

-api-addr string
    API server address (default ":8080")

-download-path string
    Download directory path (default "./downloads")

-workers int
    Number of concurrent download workers (default 3)
```

## API Reference

Base URL: `http://localhost:8080`

### Health Check

```bash
GET /health
```

### Tracks

#### Create Track
```bash
POST /api/tracks
Content-Type: application/json

{
  "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
  "title": "Track Title",
  "artist": "Artist Name",
  "album": "Album Name"
}
```

#### List All Tracks
```bash
GET /api/tracks

# Filter by service
GET /api/tracks?service=youtube
```

#### Get Track by ID
```bash
GET /api/tracks/{id}
```

#### Update Track
```bash
PUT /api/tracks/{id}
Content-Type: application/json

{
  "title": "Updated Title",
  "artist": "Updated Artist"
}
```

#### Delete Track
```bash
DELETE /api/tracks/{id}
```

#### Search Tracks
```bash
GET /api/tracks/search?q=artist%20name
```

### Playlists

#### Create Playlist
```bash
POST /api/playlists
Content-Type: application/json

{
  "name": "My DJ Set",
  "description": "House music for Friday night"
}
```

#### List All Playlists
```bash
GET /api/playlists
```

#### Get Playlist
```bash
# Basic info
GET /api/playlists/{id}

# Include full track details
GET /api/playlists/{id}?include_tracks=true
```

#### Update Playlist
```bash
PUT /api/playlists/{id}
Content-Type: application/json

{
  "name": "Updated Playlist Name",
  "description": "Updated description"
}
```

#### Add Tracks to Playlist
```bash
POST /api/playlists/{id}/tracks
Content-Type: application/json

{
  "track_ids": ["track-id-1", "track-id-2"]
}
```

#### Remove Tracks from Playlist
```bash
DELETE /api/playlists/{id}/tracks
Content-Type: application/json

{
  "track_ids": ["track-id-1"]
}
```

#### Delete Playlist
```bash
DELETE /api/playlists/{id}
```

### Downloads

#### Download Single Track
```bash
POST /api/download
Content-Type: application/json

{
  "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
  "title": "Track Title",
  "artist": "Artist Name",
  "username": "tidal_user",    # For Tidal only
  "password": "tidal_pass",    # For Tidal only
  "playlist_id": "optional-playlist-id"
}
```

#### Download Entire Playlist
```bash
POST /api/download/playlist/{playlist_id}
```

#### List Download Jobs
```bash
# All jobs
GET /api/download/jobs

# Filter by status (pending, downloading, completed, failed)
GET /api/download/jobs?status=completed
```

#### Get Download Job Status
```bash
GET /api/download/jobs/{id}
```

## DJ Workflow Example

```bash
# 1. Create a playlist for your set
curl -X POST http://localhost:8080/api/playlists \
  -H "Content-Type: application/json" \
  -d '{"name":"Friday Night House Set","description":"Deep house vibes"}'

# Response: {"id":"playlist-id",...}

# 2. Add tracks to your playlist
curl -X POST http://localhost:8080/api/tracks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://youtube.com/watch?v=...","title":"Track 1","artist":"Artist"}'

# Response: {"id":"track-id-1",...}

curl -X POST http://localhost:8080/api/playlists/playlist-id/tracks \
  -H "Content-Type: application/json" \
  -d '{"track_ids":["track-id-1"]}'

# 3. Download the entire playlist
curl -X POST http://localhost:8080/api/download/playlist/playlist-id

# 4. Check download progress
curl http://localhost:8080/api/download/jobs?status=downloading

# 5. Get playlist with tracks once downloads complete
curl "http://localhost:8080/api/playlists/playlist-id?include_tracks=true"

# 6. Copy downloads to USB
cp -r downloads/* /media/usb/DJ_Sets/Friday_Night/
``` 
