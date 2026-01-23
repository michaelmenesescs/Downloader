# Downloader

A Go-based media downloader for Tidal, SoundCloud, and YouTube, optimized for DJ sets and mixes.

## Description

This application allows you to download media from:
- **Tidal** (requires username and password)
- **SoundCloud**
- **YouTube** (optimized for DJ sets with enhanced features)

The application detects the URL type and uses the appropriate downloader tool:
- `tidal-dl` for Tidal
- `scdl` for SoundCloud
- `yt-dlp` for YouTube (DJ set scraper)

## YouTube DJ Set Scraper Features

The YouTube downloader is specifically optimized for DJ sets and mixes with the following features:
- **Best audio quality**: Downloads highest quality audio available
- **MP3 format**: Automatic conversion to MP3 with VBR encoding
- **Metadata extraction**: Includes title, uploader, and other metadata
- **Embedded thumbnails**: Video thumbnail embedded as album art
- **Playlist support**: Automatically downloads entire playlists
- **Organized output**: Files saved to `./downloads` directory with descriptive names

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

The application will prompt you for a URL and automatically detect the service to download from. 
