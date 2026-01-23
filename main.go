package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// downloadTidal downloads media from Tidal using tidal-dl
func downloadTidal(url, username, password string) error {
	cmd := exec.Command("tidal-dl", "-u", username, "-p", password, url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// downloadSoundcloud downloads media from SoundCloud using scdl
func downloadSoundcloud(url string) error {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("scdl -l %s", url))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// downloadYoutubeDJSet downloads DJ sets from YouTube using yt-dlp with enhanced features
// Features: best audio quality, metadata extraction, playlist support, embedded thumbnails
func downloadYoutubeDJSet(url string) error {
	// Create downloads directory if it doesn't exist
	downloadsDir := "./downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create downloads directory: %w", err)
	}

	// yt-dlp command with optimized settings for DJ sets:
	// -f bestaudio: Download best audio quality
	// --extract-audio: Extract audio from video
	// --audio-format mp3: Convert to MP3 format
	// --audio-quality 0: Best audio quality (VBR)
	// --embed-thumbnail: Embed video thumbnail as album art
	// --add-metadata: Add metadata to the file
	// --metadata-from-title: Extract additional metadata from title
	// -o: Output template with metadata in filename
	// --yes-playlist: Download playlists if URL is a playlist
	args := []string{
		"-f", "bestaudio",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--embed-thumbnail",
		"--add-metadata",
		"--metadata-from-title", "%(artist)s - %(title)s",
		"-o", downloadsDir + "/%(uploader)s - %(title)s.%(ext)s",
		"--yes-playlist",
		"--no-mtime",
		url,
	}

	cmd := exec.Command("yt-dlp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("\nDownloading with yt-dlp (best audio quality, with metadata)...")
	fmt.Printf("Output directory: %s\n\n", downloadsDir)

	return cmd.Run()
}

// VideoMetadata holds information about a YouTube video
type VideoMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	URL         string `json:"webpage_url"`
	Uploader    string `json:"uploader"`
}

// TracklistEntry represents a single track in a DJ mix
type TracklistEntry struct {
	Timestamp string
	Artist    string
	Track     string
	RawLine   string
}

// searchDJMixes searches YouTube for DJ mixes and returns video metadata
func searchDJMixes(djName string, maxResults int) ([]VideoMetadata, error) {
	searchQueries := []string{
		fmt.Sprintf("ytsearch%d:%s mix", maxResults/2, djName),
		fmt.Sprintf("ytsearch%d:%s live set", maxResults/2, djName),
	}

	var allVideos []VideoMetadata

	for _, query := range searchQueries {
		// Use yt-dlp to search and get JSON metadata
		cmd := exec.Command("yt-dlp", "-j", "--flat-playlist", "--no-check-certificate", query)
		output, err := cmd.Output()
		if err != nil {
			continue // Skip if search fails
		}

		// Parse JSON output (one JSON object per line)
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			var video VideoMetadata
			if err := json.Unmarshal([]byte(line), &video); err != nil {
				continue
			}
			allVideos = append(allVideos, video)
		}
	}

	return allVideos, nil
}

// getVideoMetadata fetches full metadata for a video including description
func getVideoMetadata(videoID string) (*VideoMetadata, error) {
	cmd := exec.Command("yt-dlp", "-j", "--no-playlist", "--no-check-certificate", fmt.Sprintf("https://youtube.com/watch?v=%s", videoID))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var video VideoMetadata
	if err := json.Unmarshal(output, &video); err != nil {
		return nil, err
	}

	return &video, nil
}

// parseTracklist extracts tracklist from video description
func parseTracklist(description string) []TracklistEntry {
	var tracklist []TracklistEntry

	// Common timestamp patterns:
	// 00:00 Artist - Track
	// [00:00] Artist - Track
	// 1. Artist - Track
	// 00:00:00 Artist - Track
	timestampPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]\s*(.+)$`),
		regexp.MustCompile(`(?m)^\[(\d{1,2}:\d{2}(?::\d{2})?)\]\s*(.+)$`),
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+(.+)$`),
		regexp.MustCompile(`(?m)^\d+\.\s*(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]\s*(.+)$`),
		regexp.MustCompile(`(?m)^\d+\.\s+(.+?)\s+[-–—]\s+(.+)$`), // Numbered list without timestamp
	}

	lines := strings.Split(description, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for _, pattern := range timestampPatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				timestamp := matches[1]
				trackInfo := matches[2]

				// Try to split Artist - Track
				parts := strings.SplitN(trackInfo, "-", 2)
				var artist, track string
				if len(parts) == 2 {
					artist = strings.TrimSpace(parts[0])
					track = strings.TrimSpace(parts[1])
				} else {
					artist = "Unknown"
					track = strings.TrimSpace(trackInfo)
				}

				tracklist = append(tracklist, TracklistEntry{
					Timestamp: timestamp,
					Artist:    artist,
					Track:     track,
					RawLine:   line,
				})
				break
			}
		}
	}

	return tracklist
}

// findDJTracklists searches for a DJ and displays tracklists from their mixes
func findDJTracklists(djName string) error {
	fmt.Printf("\n🔍 Searching for %s mixes on YouTube...\n\n", djName)

	// Search for videos
	videos, err := searchDJMixes(djName, 10)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(videos) == 0 {
		fmt.Println("No videos found.")
		return nil
	}

	fmt.Printf("Found %d videos. Extracting tracklists...\n\n", len(videos))

	// Process each video
	for i, video := range videos {
		fmt.Printf("─────────────────────────────────────────────────────────────\n")
		fmt.Printf("Video %d/%d\n", i+1, len(videos))
		fmt.Printf("Title: %s\n", video.Title)
		fmt.Printf("URL: https://youtube.com/watch?v=%s\n", video.ID)

		// Fetch full metadata with description
		fullVideo, err := getVideoMetadata(video.ID)
		if err != nil {
			fmt.Printf("⚠️  Could not fetch video details\n\n")
			continue
		}

		// Format duration
		duration := fullVideo.Duration
		hours := duration / 3600
		minutes := (duration % 3600) / 60
		seconds := duration % 60
		var durationStr string
		if hours > 0 {
			durationStr = fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
		} else {
			durationStr = fmt.Sprintf("%d:%02d", minutes, seconds)
		}
		fmt.Printf("Duration: %s\n", durationStr)
		fmt.Printf("Uploader: %s\n\n", fullVideo.Uploader)

		// Parse tracklist
		tracklist := parseTracklist(fullVideo.Description)

		if len(tracklist) == 0 {
			fmt.Printf("❌ No tracklist found in description\n\n")
		} else {
			fmt.Printf("✅ Tracklist (%d tracks):\n", len(tracklist))
			for _, track := range tracklist {
				fmt.Printf("  %s - %s - %s\n", track.Timestamp, track.Artist, track.Track)
			}
			fmt.Println()
		}
	}

	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	fmt.Printf("\n✨ Search complete! Found tracklists in %d videos.\n\n", len(videos))

	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Regular expressions for URL matching
	soundcloudRegex := regexp.MustCompile(`^https?://(www\.)?soundcloud\.com/`)
	tidalRegex1 := regexp.MustCompile(`^https?://listen\.tidal\.com/`)
	tidalRegex2 := regexp.MustCompile(`^https?://tidal\.com/`)
	youtubeRegex1 := regexp.MustCompile(`^https?://(www\.)?youtube\.com/`)
	youtubeRegex2 := regexp.MustCompile(`^https?://(www\.)?youtu\.be/`)

	for {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Download media or find DJ tracklists")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Enter a URL to download, or a DJ name to find tracklists")
		fmt.Print("\n> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		input = strings.TrimSpace(input)

		// Check if input is a URL or DJ name
		isURL := soundcloudRegex.MatchString(input) ||
			tidalRegex1.MatchString(input) ||
			tidalRegex2.MatchString(input) ||
			youtubeRegex1.MatchString(input) ||
			youtubeRegex2.MatchString(input)

		if !isURL {
			// Treat as DJ name search
			if err := findDJTracklists(input); err != nil {
				fmt.Fprintf(os.Stderr, "Error searching for DJ tracklists: %v\n", err)
			}
			continue
		}

		// URL-based download mode
		url := input

		// Match URL pattern and call appropriate downloader
		if soundcloudRegex.MatchString(url) {
			fmt.Println("SoundCloud")
			if err := downloadSoundcloud(url); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from SoundCloud: %v\n", err)
				os.Exit(1)
			}
		} else if tidalRegex1.MatchString(url) || tidalRegex2.MatchString(url) {
			fmt.Println("Tidal")

			// Get Tidal credentials
			fmt.Print("Tidal username: ")
			username, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading username: %v\n", err)
				os.Exit(1)
			}
			username = strings.TrimSpace(username)

			fmt.Print("Tidal password: ")
			password, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
			password = strings.TrimSpace(password)

			if err := downloadTidal(url, username, password); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from Tidal: %v\n", err)
				os.Exit(1)
			}
		} else if youtubeRegex1.MatchString(url) || youtubeRegex2.MatchString(url) {
			fmt.Println("YouTube DJ Set Scraper")
			if err := downloadYoutubeDJSet(url); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from YouTube: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: URL %s is not supported.\n", url)
			os.Exit(1)
		}
	}
}
