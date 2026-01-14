package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

const (
	tracksIndex       = "tracks"
	playlistsIndex    = "playlists"
	downloadJobsIndex = "download_jobs"
)

// Storage handles all Elasticsearch operations
type Storage struct {
	client *elasticsearch.Client
}

// NewStorage creates a new Storage instance
func NewStorage(addresses []string) (*Storage, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	storage := &Storage{client: client}

	// Initialize indices
	if err := storage.initIndices(); err != nil {
		return nil, fmt.Errorf("error initializing indices: %w", err)
	}

	return storage, nil
}

// initIndices creates the necessary indices if they don't exist
func (s *Storage) initIndices() error {
	ctx := context.Background()

	indices := []string{tracksIndex, playlistsIndex, downloadJobsIndex}

	for _, index := range indices {
		req := esapi.IndicesExistsRequest{Index: []string{index}}
		res, err := req.Do(ctx, s.client)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode == 404 {
			// Create index
			createReq := esapi.IndicesCreateRequest{Index: index}
			createRes, err := createReq.Do(ctx, s.client)
			if err != nil {
				return err
			}
			defer createRes.Body.Close()

			if createRes.IsError() {
				return fmt.Errorf("error creating index %s: %s", index, createRes.String())
			}
		}
	}

	return nil
}

// CreateTrack creates a new track
func (s *Storage) CreateTrack(track *Track) error {
	track.ID = uuid.New().String()
	track.CreatedAt = time.Now()
	track.UpdatedAt = time.Now()

	data, err := json.Marshal(track)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      tracksIndex,
		DocumentID: track.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing track: %s", res.String())
	}

	return nil
}

// GetTrack retrieves a track by ID
func (s *Storage) GetTrack(id string) (*Track, error) {
	req := esapi.GetRequest{
		Index:      tracksIndex,
		DocumentID: id,
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("track not found")
		}
		return nil, fmt.Errorf("error getting track: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	source := result["_source"].(map[string]interface{})
	trackData, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := json.Unmarshal(trackData, &track); err != nil {
		return nil, err
	}

	return &track, nil
}

// UpdateTrack updates an existing track
func (s *Storage) UpdateTrack(track *Track) error {
	track.UpdatedAt = time.Now()

	data, err := json.Marshal(track)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      tracksIndex,
		DocumentID: track.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating track: %s", res.String())
	}

	return nil
}

// DeleteTrack deletes a track
func (s *Storage) DeleteTrack(id string) error {
	req := esapi.DeleteRequest{
		Index:      tracksIndex,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting track: %s", res.String())
	}

	return nil
}

// SearchTracks searches for tracks
func (s *Storage) SearchTracks(query string) ([]Track, error) {
	var buf bytes.Buffer
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title", "artist", "album"},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(searchQuery); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(tracksIndex),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching tracks: %s", res.String())
	}

	return s.parseTrackSearchResults(res.Body)
}

// ListAllTracks retrieves all tracks
func (s *Storage) ListAllTracks() ([]Track, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 1000,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(tracksIndex),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error listing tracks: %s", res.String())
	}

	return s.parseTrackSearchResults(res.Body)
}

// parseTrackSearchResults parses Elasticsearch search results into Track slice
func (s *Storage) parseTrackSearchResults(body io.Reader) ([]Track, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	tracks := make([]Track, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		trackData, err := json.Marshal(source)
		if err != nil {
			continue
		}

		var track Track
		if err := json.Unmarshal(trackData, &track); err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// CreatePlaylist creates a new playlist
func (s *Storage) CreatePlaylist(playlist *Playlist) error {
	playlist.ID = uuid.New().String()
	playlist.CreatedAt = time.Now()
	playlist.UpdatedAt = time.Now()
	if playlist.TrackIDs == nil {
		playlist.TrackIDs = []string{}
	}

	data, err := json.Marshal(playlist)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      playlistsIndex,
		DocumentID: playlist.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing playlist: %s", res.String())
	}

	return nil
}

// GetPlaylist retrieves a playlist by ID
func (s *Storage) GetPlaylist(id string) (*Playlist, error) {
	req := esapi.GetRequest{
		Index:      playlistsIndex,
		DocumentID: id,
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("playlist not found")
		}
		return nil, fmt.Errorf("error getting playlist: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	source := result["_source"].(map[string]interface{})
	playlistData, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var playlist Playlist
	if err := json.Unmarshal(playlistData, &playlist); err != nil {
		return nil, err
	}

	return &playlist, nil
}

// GetPlaylistWithTracks retrieves a playlist with all track details
func (s *Storage) GetPlaylistWithTracks(id string) (*PlaylistWithTracks, error) {
	playlist, err := s.GetPlaylist(id)
	if err != nil {
		return nil, err
	}

	tracks := make([]Track, 0, len(playlist.TrackIDs))
	for _, trackID := range playlist.TrackIDs {
		track, err := s.GetTrack(trackID)
		if err != nil {
			continue // Skip tracks that can't be found
		}
		tracks = append(tracks, *track)
	}

	return &PlaylistWithTracks{
		Playlist: *playlist,
		Tracks:   tracks,
	}, nil
}

// UpdatePlaylist updates an existing playlist
func (s *Storage) UpdatePlaylist(playlist *Playlist) error {
	playlist.UpdatedAt = time.Now()

	data, err := json.Marshal(playlist)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      playlistsIndex,
		DocumentID: playlist.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating playlist: %s", res.String())
	}

	return nil
}

// DeletePlaylist deletes a playlist
func (s *Storage) DeletePlaylist(id string) error {
	req := esapi.DeleteRequest{
		Index:      playlistsIndex,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error deleting playlist: %s", res.String())
	}

	return nil
}

// ListAllPlaylists retrieves all playlists
func (s *Storage) ListAllPlaylists() ([]Playlist, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 1000,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(playlistsIndex),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error listing playlists: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	playlists := make([]Playlist, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		playlistData, err := json.Marshal(source)
		if err != nil {
			continue
		}

		var playlist Playlist
		if err := json.Unmarshal(playlistData, &playlist); err != nil {
			continue
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

// CreateDownloadJob creates a new download job
func (s *Storage) CreateDownloadJob(job *DownloadJob) error {
	job.ID = uuid.New().String()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      downloadJobsIndex,
		DocumentID: job.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing download job: %s", res.String())
	}

	return nil
}

// GetDownloadJob retrieves a download job by ID
func (s *Storage) GetDownloadJob(id string) (*DownloadJob, error) {
	req := esapi.GetRequest{
		Index:      downloadJobsIndex,
		DocumentID: id,
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("download job not found")
		}
		return nil, fmt.Errorf("error getting download job: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	source := result["_source"].(map[string]interface{})
	jobData, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var job DownloadJob
	if err := json.Unmarshal(jobData, &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// UpdateDownloadJob updates an existing download job
func (s *Storage) UpdateDownloadJob(job *DownloadJob) error {
	job.UpdatedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      downloadJobsIndex,
		DocumentID: job.ID,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error updating download job: %s", res.String())
	}

	return nil
}

// ListDownloadJobs retrieves download jobs by status
func (s *Storage) ListDownloadJobs(status string) ([]DownloadJob, error) {
	var buf bytes.Buffer
	var query map[string]interface{}

	if status == "" {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"size": 1000,
			"sort": []map[string]interface{}{
				{"created_at": map[string]string{"order": "desc"}},
			},
		}
	} else {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"term": map[string]interface{}{
					"status": status,
				},
			},
			"size": 1000,
			"sort": []map[string]interface{}{
				{"created_at": map[string]string{"order": "desc"}},
			},
		}
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(downloadJobsIndex),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error listing download jobs: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	jobs := make([]DownloadJob, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		jobData, err := json.Marshal(source)
		if err != nil {
			continue
		}

		var job DownloadJob
		if err := json.Unmarshal(jobData, &job); err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Ping checks if Elasticsearch is reachable
func (s *Storage) Ping() error {
	res, err := s.client.Info()
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("elasticsearch error: %s", string(body))
	}

	var info map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return err
	}

	version := info["version"].(map[string]interface{})["number"].(string)
	fmt.Printf("Connected to Elasticsearch version %s\n", version)

	return nil
}

// AddTracksToPlaylist adds tracks to a playlist
func (s *Storage) AddTracksToPlaylist(playlistID string, trackIDs []string) error {
	playlist, err := s.GetPlaylist(playlistID)
	if err != nil {
		return err
	}

	// Create a map to avoid duplicates
	trackMap := make(map[string]bool)
	for _, id := range playlist.TrackIDs {
		trackMap[id] = true
	}

	// Add new tracks
	for _, id := range trackIDs {
		if !trackMap[id] {
			playlist.TrackIDs = append(playlist.TrackIDs, id)
			trackMap[id] = true
		}
	}

	return s.UpdatePlaylist(playlist)
}

// RemoveTracksFromPlaylist removes tracks from a playlist
func (s *Storage) RemoveTracksFromPlaylist(playlistID string, trackIDs []string) error {
	playlist, err := s.GetPlaylist(playlistID)
	if err != nil {
		return err
	}

	// Create a set of tracks to remove
	removeSet := make(map[string]bool)
	for _, id := range trackIDs {
		removeSet[id] = true
	}

	// Filter out removed tracks
	newTrackIDs := make([]string, 0)
	for _, id := range playlist.TrackIDs {
		if !removeSet[id] {
			newTrackIDs = append(newTrackIDs, id)
		}
	}

	playlist.TrackIDs = newTrackIDs
	return s.UpdatePlaylist(playlist)
}

// GetTracksByService retrieves all tracks for a specific service
func (s *Storage) GetTracksByService(service string) ([]Track, error) {
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				"service": strings.ToLower(service),
			},
		},
		"size": 1000,
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex(tracksIndex),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error searching tracks by service: %s", res.String())
	}

	return s.parseTrackSearchResults(res.Body)
}
