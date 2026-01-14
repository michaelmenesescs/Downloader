package main

import (
	"time"
)

// Track represents a music track with metadata
type Track struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Album       string    `json:"album,omitempty"`
	Service     string    `json:"service"` // tidal, soundcloud, youtube
	Duration    int       `json:"duration,omitempty"`
	FilePath    string    `json:"file_path,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Playlist represents a collection of tracks
type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	TrackIDs    []string  `json:"track_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DownloadJob represents a download task
type DownloadJob struct {
	ID          string    `json:"id"`
	TrackID     string    `json:"track_id"`
	PlaylistID  string    `json:"playlist_id,omitempty"`
	Status      string    `json:"status"` // pending, downloading, completed, failed
	Error       string    `json:"error,omitempty"`
	Progress    int       `json:"progress"` // 0-100
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// PlaylistWithTracks represents a playlist with full track details
type PlaylistWithTracks struct {
	Playlist
	Tracks []Track `json:"tracks"`
}

// DownloadRequest represents a request to download a track
type DownloadRequest struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Album      string `json:"album,omitempty"`
	PlaylistID string `json:"playlist_id,omitempty"`
	Username   string `json:"username,omitempty"` // For Tidal
	Password   string `json:"password,omitempty"` // For Tidal
}

// CreatePlaylistRequest represents a request to create a playlist
type CreatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AddTracksToPlaylistRequest represents adding tracks to a playlist
type AddTracksToPlaylistRequest struct {
	TrackIDs []string `json:"track_ids"`
}
