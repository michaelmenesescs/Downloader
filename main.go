package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
// RawLine is included to verify extraction accuracy (no hallucinations)
type TracklistEntry struct {
	Timestamp string `json:"timestamp"`
	Artist    string `json:"artist"`
	Track     string `json:"track"`
	RawLine   string `json:"raw_line"` // Original line from description for verification
}

// VideoTracklist holds a video and its extracted tracklist
type VideoTracklist struct {
	VideoID     string           `json:"video_id"`
	Title       string           `json:"title"`
	URL         string           `json:"url"`
	Duration    string           `json:"duration"`
	Uploader    string           `json:"uploader"`
	Tracklist   []TracklistEntry `json:"tracklist"`
	TrackCount  int              `json:"track_count"`
	HasTracklist bool            `json:"has_tracklist"`
}

// DJSearchResult holds all results for a DJ search
type DJSearchResult struct {
	DJName       string           `json:"dj_name"`
	SearchDate   string           `json:"search_date"`
	TotalVideos  int              `json:"total_videos"`
	VideosWithTracklists int      `json:"videos_with_tracklists"`
	Videos       []VideoTracklist `json:"videos"`
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
// CONSERVATIVE PARSING: Only extracts data that clearly exists in the description
// No hallucinations - if format is ambiguous, we preserve the raw info
func parseTracklist(description string) []TracklistEntry {
	var tracklist []TracklistEntry

	// Conservative timestamp patterns - require clear timestamps to avoid false positives
	// Pattern priority: most specific to least specific
	timestampPatterns := []*regexp.Regexp{
		// 00:00 - Artist - Track (with dash separator)
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]\s*(.+)$`),
		// [00:00] Artist - Track (bracketed timestamp)
		regexp.MustCompile(`(?m)^\[(\d{1,2}:\d{2}(?::\d{2})?)\]\s*(.+)$`),
		// 1. 00:00 - Artist - Track (numbered with timestamp)
		regexp.MustCompile(`(?m)^\d+\.\s*(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]?\s*(.+)$`),
		// 00:00 Artist - Track (timestamp with space, no dash)
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+(.+)$`),
	}

	lines := strings.Split(description, "\n")
	seenLines := make(map[string]bool) // Deduplicate

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines or very short lines (likely not tracks)
		if line == "" || len(line) < 8 {
			continue
		}

		// Skip if already processed (deduplicate)
		if seenLines[line] {
			continue
		}

		for _, pattern := range timestampPatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				timestamp := matches[1]
				trackInfo := strings.TrimSpace(matches[2])

				// Skip if track info is too short (likely noise)
				if len(trackInfo) < 3 {
					continue
				}

				// Conservative artist/track splitting
				// Only split on dash if it exists, otherwise store full info in track field
				var artist, track string

				// Try to split on dash (various dash types: -, –, —)
				dashIndex := -1
				for i, r := range trackInfo {
					if r == '-' || r == '–' || r == '—' {
						dashIndex = i
						break
					}
				}

				if dashIndex > 0 && dashIndex < len(trackInfo)-1 {
					// Found a dash separator - split into artist and track
					artist = strings.TrimSpace(trackInfo[:dashIndex])
					track = strings.TrimSpace(trackInfo[dashIndex+1:])

					// Validate: both parts should have reasonable length
					if len(artist) < 1 || len(track) < 1 {
						// Invalid split - store full info as track
						artist = ""
						track = trackInfo
					}
				} else {
					// No clear dash separator - store full info as track
					// DO NOT hallucinate an "Unknown" artist
					artist = ""
					track = trackInfo
				}

				tracklist = append(tracklist, TracklistEntry{
					Timestamp: timestamp,
					Artist:    artist,
					Track:     track,
					RawLine:   line, // Always preserve original for verification
				})

				seenLines[line] = true
				break
			}
		}
	}

	return tracklist
}

// exportToJSON exports search results to JSON file with full verification data
func exportToJSON(result *DJSearchResult, outputDir string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.json",
		sanitizeFilename(result.DJName),
		time.Now().Format("20060102_150405")))

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("✅ JSON exported to: %s\n", filename)
	return nil
}

// exportToCSV exports all tracklists to CSV file
func exportToCSV(result *DJSearchResult, outputDir string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.csv",
		sanitizeFilename(result.DJName),
		time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Video Title", "Video URL", "Timestamp", "Artist", "Track", "Raw Line (Verification)"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, video := range result.Videos {
		for _, track := range video.Tracklist {
			row := []string{
				video.Title,
				video.URL,
				track.Timestamp,
				track.Artist,
				track.Track,
				track.RawLine,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	fmt.Printf("✅ CSV exported to: %s\n", filename)
	return nil
}

// exportToMarkdown exports results to human-readable markdown file
func exportToMarkdown(result *DJSearchResult, outputDir string) error {
	filename := filepath.Join(outputDir, fmt.Sprintf("%s_%s.md",
		sanitizeFilename(result.DJName),
		time.Now().Format("20060102_150405")))

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# DJ Tracklists: %s\n\n", result.DJName)
	fmt.Fprintf(file, "**Search Date:** %s\n", result.SearchDate)
	fmt.Fprintf(file, "**Total Videos:** %d\n", result.TotalVideos)
	fmt.Fprintf(file, "**Videos with Tracklists:** %d\n\n", result.VideosWithTracklists)
	fmt.Fprintf(file, "---\n\n")

	// Write each video
	for i, video := range result.Videos {
		fmt.Fprintf(file, "## Video %d: %s\n\n", i+1, video.Title)
		fmt.Fprintf(file, "- **URL:** %s\n", video.URL)
		fmt.Fprintf(file, "- **Duration:** %s\n", video.Duration)
		fmt.Fprintf(file, "- **Uploader:** %s\n", video.Uploader)
		fmt.Fprintf(file, "- **Tracks:** %d\n\n", video.TrackCount)

		if video.HasTracklist {
			fmt.Fprintf(file, "### Tracklist\n\n")
			for _, track := range video.Tracklist {
				fmt.Fprintf(file, "- **%s** - %s - %s\n", track.Timestamp, track.Artist, track.Track)
			}
			fmt.Fprintf(file, "\n")

			// Add verification section with raw lines
			fmt.Fprintf(file, "<details>\n<summary>Raw Lines (for verification)</summary>\n\n")
			fmt.Fprintf(file, "```\n")
			for _, track := range video.Tracklist {
				fmt.Fprintf(file, "%s\n", track.RawLine)
			}
			fmt.Fprintf(file, "```\n</details>\n\n")
		} else {
			fmt.Fprintf(file, "*No tracklist found in description*\n\n")
		}

		fmt.Fprintf(file, "---\n\n")
	}

	fmt.Printf("✅ Markdown exported to: %s\n", filename)
	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace spaces and invalid characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	return reg.ReplaceAllString(strings.ReplaceAll(name, " ", "_"), "_")
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

	// Initialize result structure
	result := &DJSearchResult{
		DJName:       djName,
		SearchDate:   time.Now().Format("2006-01-02 15:04:05"),
		TotalVideos:  len(videos),
		VideosWithTracklists: 0,
		Videos:       make([]VideoTracklist, 0),
	}

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

		// Build video tracklist entry
		videoTracklist := VideoTracklist{
			VideoID:     video.ID,
			Title:       video.Title,
			URL:         fmt.Sprintf("https://youtube.com/watch?v=%s", video.ID),
			Duration:    durationStr,
			Uploader:    fullVideo.Uploader,
			Tracklist:   tracklist,
			TrackCount:  len(tracklist),
			HasTracklist: len(tracklist) > 0,
		}

		result.Videos = append(result.Videos, videoTracklist)

		if len(tracklist) == 0 {
			fmt.Printf("❌ No tracklist found in description\n\n")
		} else {
			result.VideosWithTracklists++
			fmt.Printf("✅ Tracklist (%d tracks):\n", len(tracklist))
			for _, track := range tracklist {
				fmt.Printf("  %s - %s - %s\n", track.Timestamp, track.Artist, track.Track)
			}
			fmt.Println()
		}
	}

	fmt.Printf("─────────────────────────────────────────────────────────────\n")
	fmt.Printf("\n✨ Search complete! Found tracklists in %d/%d videos.\n\n",
		result.VideosWithTracklists, result.TotalVideos)

	// Export results
	if result.VideosWithTracklists > 0 {
		exportDir := "./tracklists"
		if err := os.MkdirAll(exportDir, 0755); err != nil {
			return fmt.Errorf("failed to create export directory: %w", err)
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("Exporting tracklists...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Export to all formats
		if err := exportToJSON(result, exportDir); err != nil {
			fmt.Printf("⚠️  JSON export failed: %v\n", err)
		}

		if err := exportToCSV(result, exportDir); err != nil {
			fmt.Printf("⚠️  CSV export failed: %v\n", err)
		}

		if err := exportToMarkdown(result, exportDir); err != nil {
			fmt.Printf("⚠️  Markdown export failed: %v\n", err)
		}

		fmt.Println("\n💾 All exports saved to:", exportDir)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	} else {
		fmt.Println("ℹ️  No tracklists found to export.\n")
	}

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
