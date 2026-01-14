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

### Local

```bash
./downloader
```

### Docker

```bash
docker run -it -v $(pwd)/downloads:/app/downloads downloader
```

The application will prompt you for a URL and automatically detect the service to download from. 
