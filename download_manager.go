package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DownloadManager manages the download queue and workers
type DownloadManager struct {
	storage       *Storage
	executor      CommandExecutor
	downloadPath  string
	workers       int
	queue         chan *DownloadJob
	wg            sync.WaitGroup
	stopChan      chan struct{}
	mu            sync.Mutex
}

// NewDownloadManager creates a new DownloadManager
func NewDownloadManager(storage *Storage, executor CommandExecutor, downloadPath string, workers int) *DownloadManager {
	return &DownloadManager{
		storage:      storage,
		executor:     executor,
		downloadPath: downloadPath,
		workers:      workers,
		queue:        make(chan *DownloadJob, 100),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the download workers
func (dm *DownloadManager) Start() {
	for i := 0; i < dm.workers; i++ {
		dm.wg.Add(1)
		go dm.worker(i)
	}

	// Start the queue processor
	go dm.processQueue()

	log.Printf("Download manager started with %d workers", dm.workers)
}

// Stop stops the download manager gracefully
func (dm *DownloadManager) Stop() {
	close(dm.stopChan)
	close(dm.queue)
	dm.wg.Wait()
	log.Println("Download manager stopped")
}

// worker processes download jobs from the queue
func (dm *DownloadManager) worker(id int) {
	defer dm.wg.Done()

	for job := range dm.queue {
		log.Printf("Worker %d: Processing job %s", id, job.ID)
		dm.processJob(job)
	}
}

// processQueue continuously checks for pending downloads
func (dm *DownloadManager) processQueue() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Process pending jobs on startup
	dm.enqueuePendingJobs()

	for {
		select {
		case <-ticker.C:
			dm.enqueuePendingJobs()
		case <-dm.stopChan:
			return
		}
	}
}

// enqueuePendingJobs adds pending jobs to the queue
func (dm *DownloadManager) enqueuePendingJobs() {
	jobs, err := dm.storage.ListDownloadJobs("pending")
	if err != nil {
		log.Printf("Error listing pending jobs: %v", err)
		return
	}

	for i := range jobs {
		select {
		case dm.queue <- &jobs[i]:
		default:
			// Queue is full, will try again next tick
			return
		}
	}
}

// processJob handles the actual download
func (dm *DownloadManager) processJob(job *DownloadJob) {
	// Update status to downloading
	job.Status = "downloading"
	job.Progress = 0
	if err := dm.storage.UpdateDownloadJob(job); err != nil {
		log.Printf("Error updating job status: %v", err)
		return
	}

	// Get track details
	track, err := dm.storage.GetTrack(job.TrackID)
	if err != nil {
		dm.failJob(job, fmt.Sprintf("Error getting track: %v", err))
		return
	}

	// Detect service
	service := DetectService(track.URL)
	if service == ServiceUnknown {
		dm.failJob(job, "Unsupported service")
		return
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(dm.downloadPath, 0755); err != nil {
		dm.failJob(job, fmt.Sprintf("Error creating download directory: %v", err))
		return
	}

	// Change to download directory
	oldDir, err := os.Getwd()
	if err != nil {
		dm.failJob(job, fmt.Sprintf("Error getting current directory: %v", err))
		return
	}

	if err := os.Chdir(dm.downloadPath); err != nil {
		dm.failJob(job, fmt.Sprintf("Error changing to download directory: %v", err))
		return
	}
	defer os.Chdir(oldDir)

	// Update progress
	job.Progress = 50
	dm.storage.UpdateDownloadJob(job)

	// Perform download based on service
	var downloadErr error
	switch service {
	case ServiceTidal:
		// For Tidal, we need credentials - these should be stored in the track metadata
		// For now, we'll fail if credentials aren't provided
		downloadErr = fmt.Errorf("Tidal downloads require credentials - use the download endpoint with username/password")
	case ServiceSoundCloud:
		downloadErr = downloadSoundcloud(track.URL, dm.executor)
	case ServiceYouTube:
		downloadErr = downloadYoutube(track.URL, dm.executor)
	}

	if downloadErr != nil {
		dm.failJob(job, fmt.Sprintf("Download failed: %v", downloadErr))
		return
	}

	// Update track with file path (simplified - would need better file discovery)
	track.FilePath = filepath.Join(dm.downloadPath, sanitizeFilename(track.Title))
	if err := dm.storage.UpdateTrack(track); err != nil {
		log.Printf("Warning: Failed to update track file path: %v", err)
	}

	// Mark job as completed
	now := time.Now()
	job.Status = "completed"
	job.Progress = 100
	job.CompletedAt = &now
	if err := dm.storage.UpdateDownloadJob(job); err != nil {
		log.Printf("Error updating job to completed: %v", err)
		return
	}

	log.Printf("Job %s completed successfully", job.ID)
}

// failJob marks a job as failed
func (dm *DownloadManager) failJob(job *DownloadJob, errorMsg string) {
	log.Printf("Job %s failed: %s", job.ID, errorMsg)
	job.Status = "failed"
	job.Error = errorMsg
	if err := dm.storage.UpdateDownloadJob(job); err != nil {
		log.Printf("Error updating failed job: %v", err)
	}
}

// QueueDownload creates and queues a new download job
func (dm *DownloadManager) QueueDownload(trackID, playlistID string) (*DownloadJob, error) {
	job := &DownloadJob{
		TrackID:    trackID,
		PlaylistID: playlistID,
		Status:     "pending",
		Progress:   0,
	}

	if err := dm.storage.CreateDownloadJob(job); err != nil {
		return nil, err
	}

	return job, nil
}

// QueuePlaylistDownload queues all tracks in a playlist for download
func (dm *DownloadManager) QueuePlaylistDownload(playlistID string) ([]DownloadJob, error) {
	playlist, err := dm.storage.GetPlaylist(playlistID)
	if err != nil {
		return nil, err
	}

	jobs := make([]DownloadJob, 0, len(playlist.TrackIDs))

	for _, trackID := range playlist.TrackIDs {
		job, err := dm.QueueDownload(trackID, playlistID)
		if err != nil {
			log.Printf("Error queuing track %s: %v", trackID, err)
			continue
		}
		jobs = append(jobs, *job)
	}

	return jobs, nil
}

// sanitizeFilename removes invalid characters from filenames
func sanitizeFilename(filename string) string {
	// Basic sanitization - replace invalid characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	return result
}

// DownloadTrackNow immediately downloads a track (for API requests with credentials)
func (dm *DownloadManager) DownloadTrackNow(track *Track, username, password string) error {
	service := DetectService(track.URL)
	if service == ServiceUnknown {
		return fmt.Errorf("unsupported service")
	}

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(dm.downloadPath, 0755); err != nil {
		return fmt.Errorf("error creating download directory: %w", err)
	}

	// Change to download directory
	oldDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	if err := os.Chdir(dm.downloadPath); err != nil {
		return fmt.Errorf("error changing to download directory: %w", err)
	}
	defer os.Chdir(oldDir)

	// Perform download based on service
	var downloadErr error
	switch service {
	case ServiceTidal:
		downloadErr = downloadTidal(track.URL, username, password, dm.executor)
	case ServiceSoundCloud:
		downloadErr = downloadSoundcloud(track.URL, dm.executor)
	case ServiceYouTube:
		downloadErr = downloadYoutube(track.URL, dm.executor)
	}

	if downloadErr != nil {
		return fmt.Errorf("download failed: %w", downloadErr)
	}

	// Update track with file path
	track.FilePath = filepath.Join(dm.downloadPath, sanitizeFilename(track.Title))
	if err := dm.storage.UpdateTrack(track); err != nil {
		log.Printf("Warning: Failed to update track file path: %v", err)
	}

	return nil
}
