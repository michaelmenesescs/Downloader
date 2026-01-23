# Downloader

A Go-based media downloader and DJ tracklist finder for Tidal, SoundCloud, and YouTube, optimized for DJ sets and mixes.

## Description

This application provides two main features:

### 1. Media Downloads
Download media from:
- **Tidal** (requires username and password)
- **SoundCloud**
- **YouTube** (optimized for DJ sets with enhanced features)

The application detects the URL type and uses the appropriate downloader tool:
- `tidal-dl` for Tidal
- `scdl` for SoundCloud
- `yt-dlp` for YouTube (DJ set scraper)

### 2. DJ Tracklist Finder
Search for DJ mixes on YouTube and automatically extract tracklists from video descriptions. Simply enter a DJ name (e.g., "Andrey Pushkarev") and the tool will:
- Search YouTube for their mixes
- Extract tracklists from video descriptions
- Parse timestamps and track information
- Display organized results with video links

## Features

### YouTube DJ Set Downloader
The YouTube downloader is specifically optimized for DJ sets and mixes with the following features:
- **Best audio quality**: Downloads highest quality audio available
- **MP3 format**: Automatic conversion to MP3 with VBR encoding
- **Metadata extraction**: Includes title, uploader, and other metadata
- **Embedded thumbnails**: Video thumbnail embedded as album art
- **Playlist support**: Automatically downloads entire playlists
- **Organized output**: Files saved to `./downloads` directory with descriptive names

### DJ Tracklist Finder
Automatically discover and extract tracklists from DJ mixes:
- **YouTube search**: Searches for mixes by DJ name
- **Smart parsing**: Extracts timestamps and track information from descriptions
- **Multiple formats**: Supports various tracklist formats:
  - `00:00 - Artist - Track`
  - `[00:00] Artist - Track`
  - `1. Artist - Track`
- **Batch results**: Displays tracklists from multiple videos at once
- **Direct links**: Provides YouTube URLs for each mix found

## Building

### Local Build

```bash
go build -o downloader main.go
```

### Docker Build

```bash
docker build -t downloader .
```

## Usage

### Local

```bash
./downloader
```

### Docker

```bash
docker run -it -v $(pwd)/downloads:/app/downloads downloader
```

### Interactive Mode

The application runs in interactive mode and accepts two types of input:

**1. URL for downloading:**
```
> https://youtube.com/watch?v=xxxxx
```
The application will detect the service and download the media.

**2. DJ name for tracklist search:**
```
> Andrey Pushkarev
```
The application will search YouTube for the DJ's mixes and extract tracklists.

### Examples

**Finding DJ tracklists:**
```
$ ./downloader
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Download media or find DJ tracklists
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Enter a URL to download, or a DJ name to find tracklists

> Amelie Lens

🔍 Searching for Amelie Lens mixes on YouTube...

Found 10 videos. Extracting tracklists...

─────────────────────────────────────────────────────────────
Video 1/10
Title: Amelie Lens - Boiler Room Set
URL: https://youtube.com/watch?v=xxxxx
Duration: 1:30:00
Uploader: Boiler Room

✅ Tracklist (15 tracks):
  00:00 - Amelie Lens - In My Mind
  08:30 - VTSS - Berlin
  ...
```

**Downloading a mix:**
```
> https://youtube.com/watch?v=xxxxx
YouTube DJ Set Scraper
Downloading with yt-dlp (best audio quality, with metadata)...
``` 
