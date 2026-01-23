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
- **Conservative parsing**: Only extracts data that clearly exists in descriptions (no hallucinations)
- **Multiple formats**: Supports various tracklist formats:
  - `00:00 - Artist - Track`
  - `[00:00] Artist - Track`
  - `1. 00:00 - Artist - Track`
  - `00:00 Track Name` (when artist unclear)
- **Batch results**: Displays tracklists from multiple videos at once
- **Direct links**: Provides YouTube URLs for each mix found
- **Automatic export**: Saves results in JSON, CSV, and Markdown formats
- **Verification**: All exports include raw lines from descriptions for accuracy verification

## Export Formats

When tracklists are found, they are automatically exported to `./tracklists/` in three formats:

### JSON Export
- **File**: `{DJ_Name}_{timestamp}.json`
- **Contents**: Complete structured data with all metadata
- **Use case**: Machine-readable format for further processing
- **Verification**: Includes `raw_line` field for every track

### CSV Export
- **File**: `{DJ_Name}_{timestamp}.csv`
- **Columns**: Video Title, Video URL, Timestamp, Artist, Track, Raw Line (Verification)
- **Use case**: Import into spreadsheets, track management tools, databases

### Markdown Export
- **File**: `{DJ_Name}_{timestamp}.md`
- **Contents**: Human-readable format with collapsible verification sections
- **Use case**: Documentation, sharing, archiving

## Anti-Hallucination Guarantees

This tool is designed to extract **only** what exists in video descriptions:

1. **No invented data**: If artist/track split is unclear, full info is stored in track field with empty artist
2. **Raw line preservation**: Every extracted track includes the original line for verification
3. **Conservative patterns**: Only extracts lines with clear timestamps to avoid false positives
4. **Minimum length checks**: Skips very short lines that are likely noise
5. **Deduplication**: Prevents duplicate entries from being added

**You can always verify**: Check the "Raw Line" column in CSV exports or the verification sections in Markdown to confirm accuracy.

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

─────────────────────────────────────────────────────────────

✨ Search complete! Found tracklists in 8/10 videos.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Exporting tracklists...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ JSON exported to: ./tracklists/Amelie_Lens_20260123_143022.json
✅ CSV exported to: ./tracklists/Amelie_Lens_20260123_143022.csv
✅ Markdown exported to: ./tracklists/Amelie_Lens_20260123_143022.md

💾 All exports saved to: ./tracklists
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Downloading a mix:**
```
> https://youtube.com/watch?v=xxxxx
YouTube DJ Set Scraper
Downloading with yt-dlp (best audio quality, with metadata)...
``` 
