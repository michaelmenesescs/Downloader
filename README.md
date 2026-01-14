# Downloader

A Go-based media downloader for Tidal, SoundCloud, and YouTube.

## Description

This application allows you to download media from:
- **Tidal** (requires username and password)
- **SoundCloud**
- **YouTube**

The application detects the URL type and uses the appropriate downloader tool:
- `tidal-dl` for Tidal
- `scdl` for SoundCloud
- `ytmdl` for YouTube

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
