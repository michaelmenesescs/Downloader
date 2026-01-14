package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// API handles HTTP requests
type API struct {
	storage         *Storage
	downloadManager *DownloadManager
	router          *mux.Router
}

// NewAPI creates a new API instance
func NewAPI(storage *Storage, downloadManager *DownloadManager) *API {
	api := &API{
		storage:         storage,
		downloadManager: downloadManager,
		router:          mux.NewRouter(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes configures all API routes
func (api *API) setupRoutes() {
	// Health check
	api.router.HandleFunc("/health", api.healthCheck).Methods("GET")

	// Track endpoints
	api.router.HandleFunc("/api/tracks", api.createTrack).Methods("POST")
	api.router.HandleFunc("/api/tracks", api.listTracks).Methods("GET")
	api.router.HandleFunc("/api/tracks/{id}", api.getTrack).Methods("GET")
	api.router.HandleFunc("/api/tracks/{id}", api.updateTrack).Methods("PUT")
	api.router.HandleFunc("/api/tracks/{id}", api.deleteTrack).Methods("DELETE")
	api.router.HandleFunc("/api/tracks/search", api.searchTracks).Methods("GET")

	// Playlist endpoints
	api.router.HandleFunc("/api/playlists", api.createPlaylist).Methods("POST")
	api.router.HandleFunc("/api/playlists", api.listPlaylists).Methods("GET")
	api.router.HandleFunc("/api/playlists/{id}", api.getPlaylist).Methods("GET")
	api.router.HandleFunc("/api/playlists/{id}", api.updatePlaylist).Methods("PUT")
	api.router.HandleFunc("/api/playlists/{id}", api.deletePlaylist).Methods("DELETE")
	api.router.HandleFunc("/api/playlists/{id}/tracks", api.addTracksToPlaylist).Methods("POST")
	api.router.HandleFunc("/api/playlists/{id}/tracks", api.removeTracksFromPlaylist).Methods("DELETE")

	// Download endpoints
	api.router.HandleFunc("/api/download", api.downloadTrack).Methods("POST")
	api.router.HandleFunc("/api/download/playlist/{id}", api.downloadPlaylist).Methods("POST")
	api.router.HandleFunc("/api/download/jobs", api.listDownloadJobs).Methods("GET")
	api.router.HandleFunc("/api/download/jobs/{id}", api.getDownloadJob).Methods("GET")
}

// Start starts the API server
func (api *API) Start(addr string) error {
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, api.router)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Handlers

func (api *API) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := api.storage.Ping(); err != nil {
		respondError(w, http.StatusServiceUnavailable, "Elasticsearch unavailable")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Track handlers

func (api *API) createTrack(w http.ResponseWriter, r *http.Request) {
	var track Track
	if err := json.NewDecoder(r.Body).Decode(&track); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Detect service from URL
	service := DetectService(track.URL)
	if service == ServiceUnknown {
		respondError(w, http.StatusBadRequest, "Unsupported URL")
		return
	}
	track.Service = strings.ToLower(service.String())

	if err := api.storage.CreateTrack(&track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, track)
}

func (api *API) listTracks(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")

	var tracks []Track
	var err error

	if service != "" {
		tracks, err = api.storage.GetTracksByService(service)
	} else {
		tracks, err = api.storage.ListAllTracks()
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tracks)
}

func (api *API) getTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	track, err := api.storage.GetTrack(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, track)
}

func (api *API) updateTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var track Track
	if err := json.NewDecoder(r.Body).Decode(&track); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	track.ID = id
	if err := api.storage.UpdateTrack(&track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, track)
}

func (api *API) deleteTrack(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := api.storage.DeleteTrack(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Track deleted"})
}

func (api *API) searchTracks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
		return
	}

	tracks, err := api.storage.SearchTracks(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, tracks)
}

// Playlist handlers

func (api *API) createPlaylist(w http.ResponseWriter, r *http.Request) {
	var req CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	playlist := &Playlist{
		Name:        req.Name,
		Description: req.Description,
		TrackIDs:    []string{},
	}

	if err := api.storage.CreatePlaylist(playlist); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, playlist)
}

func (api *API) listPlaylists(w http.ResponseWriter, r *http.Request) {
	playlists, err := api.storage.ListAllPlaylists()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, playlists)
}

func (api *API) getPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if we should include tracks
	includeTracks := r.URL.Query().Get("include_tracks") == "true"

	if includeTracks {
		playlist, err := api.storage.GetPlaylistWithTracks(id)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, playlist)
	} else {
		playlist, err := api.storage.GetPlaylist(id)
		if err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondJSON(w, http.StatusOK, playlist)
	}
}

func (api *API) updatePlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var playlist Playlist
	if err := json.NewDecoder(r.Body).Decode(&playlist); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	playlist.ID = id
	if err := api.storage.UpdatePlaylist(&playlist); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, playlist)
}

func (api *API) deletePlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := api.storage.DeletePlaylist(id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Playlist deleted"})
}

func (api *API) addTracksToPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req AddTracksToPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.storage.AddTracksToPlaylist(id, req.TrackIDs); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	playlist, err := api.storage.GetPlaylist(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, playlist)
}

func (api *API) removeTracksFromPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req AddTracksToPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := api.storage.RemoveTracksFromPlaylist(id, req.TrackIDs); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	playlist, err := api.storage.GetPlaylist(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, playlist)
}

// Download handlers

func (api *API) downloadTrack(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create track
	track := &Track{
		URL:    req.URL,
		Title:  req.Title,
		Artist: req.Artist,
		Album:  req.Album,
	}

	service := DetectService(track.URL)
	if service == ServiceUnknown {
		respondError(w, http.StatusBadRequest, "Unsupported URL")
		return
	}
	track.Service = strings.ToLower(service.String())

	if err := api.storage.CreateTrack(track); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// If Tidal and credentials provided, download immediately
	if service == ServiceTidal && req.Username != "" && req.Password != "" {
		if err := api.downloadManager.DownloadTrackNow(track, req.Username, req.Password); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Download failed: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Download completed",
			"track":   track,
		})
		return
	}

	// Queue download
	job, err := api.downloadManager.QueueDownload(track.ID, req.PlaylistID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": "Download queued",
		"track":   track,
		"job":     job,
	})
}

func (api *API) downloadPlaylist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	jobs, err := api.downloadManager.QueuePlaylistDownload(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message": fmt.Sprintf("Queued %d tracks for download", len(jobs)),
		"jobs":    jobs,
	})
}

func (api *API) listDownloadJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	jobs, err := api.storage.ListDownloadJobs(status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, jobs)
}

func (api *API) getDownloadJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	job, err := api.storage.GetDownloadJob(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, job)
}
