package main

import (
	"bufio"
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

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Regular expressions for URL matching
	soundcloudRegex := regexp.MustCompile(`^https?://(www\.)?soundcloud\.com/`)
	tidalRegex1 := regexp.MustCompile(`^https?://listen\.tidal\.com/`)
	tidalRegex2 := regexp.MustCompile(`^https?://tidal\.com/`)
	youtubeRegex1 := regexp.MustCompile(`^https?://(www\.)?youtube\.com/`)
	youtubeRegex2 := regexp.MustCompile(`^https?://(www\.)?youtu\.be/`)

	for {
		fmt.Println("Download media from Tidal, Soundcloud, or YouTube (optimized for DJ sets).")

		// Get URL from user
		fmt.Print("URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		url = strings.TrimSpace(url)

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
