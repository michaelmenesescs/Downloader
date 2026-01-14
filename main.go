package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ServiceType represents the type of media service
type ServiceType int

const (
	ServiceUnknown ServiceType = iota
	ServiceTidal
	ServiceSoundCloud
	ServiceYouTube
)

// String returns the string representation of the service type
func (s ServiceType) String() string {
	switch s {
	case ServiceTidal:
		return "Tidal"
	case ServiceSoundCloud:
		return "SoundCloud"
	case ServiceYouTube:
		return "YouTube"
	default:
		return "Unknown"
	}
}

// CommandExecutor is an interface for executing commands (useful for testing)
type CommandExecutor interface {
	Execute(name string, args ...string) error
}

// RealCommandExecutor executes real system commands
type RealCommandExecutor struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Execute runs a real system command
func (r *RealCommandExecutor) Execute(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

// DetectService determines which service a URL belongs to
func DetectService(url string) ServiceType {
	soundcloudRegex := regexp.MustCompile(`^https?://(www\.)?soundcloud\.com/`)
	tidalRegex1 := regexp.MustCompile(`^https?://listen\.tidal\.com/`)
	tidalRegex2 := regexp.MustCompile(`^https?://tidal\.com/`)
	youtubeRegex1 := regexp.MustCompile(`^https?://(www\.)?youtube\.com/`)
	youtubeRegex2 := regexp.MustCompile(`^https?://(www\.)?youtu\.be/`)

	if soundcloudRegex.MatchString(url) {
		return ServiceSoundCloud
	} else if tidalRegex1.MatchString(url) || tidalRegex2.MatchString(url) {
		return ServiceTidal
	} else if youtubeRegex1.MatchString(url) || youtubeRegex2.MatchString(url) {
		return ServiceYouTube
	}
	return ServiceUnknown
}

// downloadTidal downloads media from Tidal using tidal-dl
func downloadTidal(url, username, password string, executor CommandExecutor) error {
	return executor.Execute("tidal-dl", "-u", username, "-p", password, url)
}

// downloadSoundcloud downloads media from SoundCloud using scdl
func downloadSoundcloud(url string, executor CommandExecutor) error {
	return executor.Execute("sh", "-c", fmt.Sprintf("scdl -l %s", url))
}

// downloadYoutube downloads media from YouTube using ytmdl
func downloadYoutube(url string, executor CommandExecutor) error {
	return executor.Execute("ytmdl", url)
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	executor := &RealCommandExecutor{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	for {
		fmt.Println("Download media from Tidal, Soundcloud, or YouTube.")

		// Get URL from user
		fmt.Print("URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		url = strings.TrimSpace(url)

		// Detect which service the URL belongs to
		service := DetectService(url)

		// Match URL pattern and call appropriate downloader
		switch service {
		case ServiceSoundCloud:
			fmt.Println("SoundCloud")
			if err := downloadSoundcloud(url, executor); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from SoundCloud: %v\n", err)
				os.Exit(1)
			}
		case ServiceTidal:
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

			if err := downloadTidal(url, username, password, executor); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from Tidal: %v\n", err)
				os.Exit(1)
			}
		case ServiceYouTube:
			fmt.Println("YouTube")
			if err := downloadYoutube(url, executor); err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading from YouTube: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: URL %s is not supported.\n", url)
			os.Exit(1)
		}
	}
}
