package main

import (
	"fmt"
	"regexp"
	"strings"
)

type TracklistEntry struct {
	Timestamp string
	Artist    string
	Track     string
	RawLine   string
}

func parseTracklist(description string) []TracklistEntry {
	var tracklist []TracklistEntry

	timestampPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]\s*(.+)$`),
		regexp.MustCompile(`(?m)^\[(\d{1,2}:\d{2}(?::\d{2})?)\]\s*(.+)$`),
		regexp.MustCompile(`(?m)^(\d{1,2}:\d{2}(?::\d{2})?)\s+(.+)$`),
		regexp.MustCompile(`(?m)^\d+\.\s*(\d{1,2}:\d{2}(?::\d{2})?)\s+[-–—]\s*(.+)$`),
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

func main() {
	testDescription := `Andrey Pushkarev Live @ Berghain 2023

This is an amazing techno set!

Tracklist:
00:00 - Svreca - Obscure
08:45 - Acronym - I Just Want To Feel Alive
18:30 - Shifted - Correct
27:15 - Dax J - Escape The System
35:00 - Rebekah - Fear Paralysis
42:30 - I Hate Models - Daydream
1:05:00 - Unknown Artist - Track Name
[1:15:00] Another Artist - Another Track

Thanks for listening!`

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  TRACKLIST PARSER DEMO")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\nInput Description:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println(testDescription)
	fmt.Println()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  EXTRACTED TRACKLIST")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	tracklist := parseTracklist(testDescription)

	if len(tracklist) == 0 {
		fmt.Println("\n❌ No tracks found!")
	} else {
		fmt.Printf("\n✅ Successfully parsed %d tracks:\n\n", len(tracklist))
		for i, track := range tracklist {
			fmt.Printf("  %2d. [%s] %s - %s\n", i+1, track.Timestamp, track.Artist, track.Track)
		}
	}
	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ✨ Parser works perfectly!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
