# API Examples

This document provides practical examples for using the Downloader API.

## Setup

All examples assume the server is running at `http://localhost:8080`.

Start the server:
```bash
docker-compose up -d
```

## Complete DJ Set Workflow

### 1. Create a Playlist

```bash
curl -X POST http://localhost:8080/api/playlists \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekend House Mix",
    "description": "Deep house tracks for Saturday night"
  }'
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Weekend House Mix",
  "description": "Deep house tracks for Saturday night",
  "track_ids": [],
  "created_at": "2026-01-14T12:00:00Z",
  "updated_at": "2026-01-14T12:00:00Z"
}
```

### 2. Add Tracks

#### Add a YouTube Track

```bash
curl -X POST http://localhost:8080/api/tracks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
    "title": "Never Gonna Give You Up",
    "artist": "Rick Astley",
    "album": "Whenever You Need Somebody"
  }'
```

#### Add a SoundCloud Track

```bash
curl -X POST http://localhost:8080/api/tracks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://soundcloud.com/artist/deep-house-track",
    "title": "Deep House Vibes",
    "artist": "House Producer"
  }'
```

### 3. Add Tracks to Playlist

```bash
curl -X POST http://localhost:8080/api/playlists/550e8400-e29b-41d4-a716-446655440000/tracks \
  -H "Content-Type: application/json" \
  -d '{
    "track_ids": [
      "track-id-1",
      "track-id-2"
    ]
  }'
```

### 4. Download Entire Playlist

```bash
curl -X POST http://localhost:8080/api/download/playlist/550e8400-e29b-41d4-a716-446655440000
```

Response:
```json
{
  "message": "Queued 2 tracks for download",
  "jobs": [
    {
      "id": "job-id-1",
      "track_id": "track-id-1",
      "playlist_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "pending",
      "progress": 0,
      "created_at": "2026-01-14T12:05:00Z",
      "updated_at": "2026-01-14T12:05:00Z"
    },
    {
      "id": "job-id-2",
      "track_id": "track-id-2",
      "playlist_id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "pending",
      "progress": 0,
      "created_at": "2026-01-14T12:05:00Z",
      "updated_at": "2026-01-14T12:05:00Z"
    }
  ]
}
```

### 5. Monitor Download Progress

```bash
# Check all downloading jobs
curl http://localhost:8080/api/download/jobs?status=downloading

# Check specific job
curl http://localhost:8080/api/download/jobs/job-id-1
```

### 6. View Playlist with All Tracks

```bash
curl "http://localhost:8080/api/playlists/550e8400-e29b-41d4-a716-446655440000?include_tracks=true"
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Weekend House Mix",
  "description": "Deep house tracks for Saturday night",
  "track_ids": ["track-id-1", "track-id-2"],
  "created_at": "2026-01-14T12:00:00Z",
  "updated_at": "2026-01-14T12:05:00Z",
  "tracks": [
    {
      "id": "track-id-1",
      "url": "https://youtube.com/watch?v=dQw4w9WgXcQ",
      "title": "Never Gonna Give You Up",
      "artist": "Rick Astley",
      "album": "Whenever You Need Somebody",
      "service": "youtube",
      "file_path": "/app/downloads/Never_Gonna_Give_You_Up.mp3",
      "created_at": "2026-01-14T12:01:00Z",
      "updated_at": "2026-01-14T12:06:00Z"
    },
    {
      "id": "track-id-2",
      "url": "https://soundcloud.com/artist/deep-house-track",
      "title": "Deep House Vibes",
      "artist": "House Producer",
      "service": "soundcloud",
      "file_path": "/app/downloads/Deep_House_Vibes.mp3",
      "created_at": "2026-01-14T12:02:00Z",
      "updated_at": "2026-01-14T12:07:00Z"
    }
  ]
}
```

## Search and Discovery

### Search for Tracks

```bash
# Search by artist
curl "http://localhost:8080/api/tracks/search?q=Rick%20Astley"

# Search by title
curl "http://localhost:8080/api/tracks/search?q=deep%20house"
```

### Filter Tracks by Service

```bash
# Get all YouTube tracks
curl "http://localhost:8080/api/tracks?service=youtube"

# Get all SoundCloud tracks
curl "http://localhost:8080/api/tracks?service=soundcloud"

# Get all Tidal tracks
curl "http://localhost:8080/api/tracks?service=tidal"
```

## Playlist Management

### Update Playlist Info

```bash
curl -X PUT http://localhost:8080/api/playlists/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Weekend House Mix",
    "description": "Updated description with more details",
    "track_ids": ["track-id-1", "track-id-2"]
  }'
```

### Remove Tracks from Playlist

```bash
curl -X DELETE http://localhost:8080/api/playlists/550e8400-e29b-41d4-a716-446655440000/tracks \
  -H "Content-Type: application/json" \
  -d '{
    "track_ids": ["track-id-1"]
  }'
```

### Delete Playlist

```bash
curl -X DELETE http://localhost:8080/api/playlists/550e8400-e29b-41d4-a716-446655440000
```

## Tidal Downloads (with Authentication)

### Download Tidal Track with Credentials

```bash
curl -X POST http://localhost:8080/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://listen.tidal.com/track/123456",
    "title": "Premium Track",
    "artist": "Tidal Artist",
    "username": "your-tidal-username",
    "password": "your-tidal-password"
  }'
```

**Note:** When username and password are provided, the download happens immediately (synchronous). Otherwise, it's queued for background processing.

## Health Check

```bash
curl http://localhost:8080/health
```

Response when healthy:
```json
{
  "status": "healthy"
}
```

## Using with Scripts

### Bash Script Example

```bash
#!/bin/bash

API_URL="http://localhost:8080"

# Create playlist
PLAYLIST=$(curl -s -X POST "$API_URL/api/playlists" \
  -H "Content-Type: application/json" \
  -d '{"name":"My DJ Set"}')

PLAYLIST_ID=$(echo $PLAYLIST | jq -r '.id')
echo "Created playlist: $PLAYLIST_ID"

# Add multiple tracks
TRACKS=(
  "https://youtube.com/watch?v=track1"
  "https://youtube.com/watch?v=track2"
  "https://youtube.com/watch?v=track3"
)

TRACK_IDS=()

for URL in "${TRACKS[@]}"; do
  TRACK=$(curl -s -X POST "$API_URL/api/tracks" \
    -H "Content-Type: application/json" \
    -d "{\"url\":\"$URL\",\"title\":\"Track\",\"artist\":\"Artist\"}")

  TRACK_ID=$(echo $TRACK | jq -r '.id')
  TRACK_IDS+=("\"$TRACK_ID\"")
  echo "Added track: $TRACK_ID"
done

# Add all tracks to playlist
TRACK_IDS_JSON=$(IFS=,; echo "[${TRACK_IDS[*]}]")
curl -s -X POST "$API_URL/api/playlists/$PLAYLIST_ID/tracks" \
  -H "Content-Type: application/json" \
  -d "{\"track_ids\":$TRACK_IDS_JSON}"

# Download playlist
curl -s -X POST "$API_URL/api/download/playlist/$PLAYLIST_ID"

echo "Playlist queued for download!"
```

### Python Script Example

```python
import requests
import json

API_URL = "http://localhost:8080"

# Create playlist
playlist_data = {
    "name": "My Python Playlist",
    "description": "Created from Python script"
}
response = requests.post(f"{API_URL}/api/playlists", json=playlist_data)
playlist = response.json()
playlist_id = playlist['id']

print(f"Created playlist: {playlist_id}")

# Add tracks
tracks = [
    {"url": "https://youtube.com/watch?v=track1", "title": "Track 1", "artist": "Artist 1"},
    {"url": "https://youtube.com/watch?v=track2", "title": "Track 2", "artist": "Artist 2"},
]

track_ids = []
for track_data in tracks:
    response = requests.post(f"{API_URL}/api/tracks", json=track_data)
    track = response.json()
    track_ids.append(track['id'])
    print(f"Added track: {track['id']}")

# Add tracks to playlist
requests.post(
    f"{API_URL}/api/playlists/{playlist_id}/tracks",
    json={"track_ids": track_ids}
)

# Download playlist
response = requests.post(f"{API_URL}/api/download/playlist/{playlist_id}")
result = response.json()
print(f"Queued {len(result['jobs'])} tracks for download")
```

## Troubleshooting

### Check if Elasticsearch is Connected

```bash
curl http://localhost:8080/health
```

If you get an error, check Elasticsearch status:
```bash
docker-compose logs elasticsearch
```

### View Download Logs

```bash
docker-compose logs -f downloader-api
```

### Check Failed Downloads

```bash
curl "http://localhost:8080/api/download/jobs?status=failed"
```

### Restart Download Workers

```bash
docker-compose restart downloader-api
```
