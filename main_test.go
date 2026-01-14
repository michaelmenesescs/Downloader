package main

import (
	"errors"
	"testing"
)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteFunc func(name string, args ...string) error
	CallHistory []CommandCall
}

// CommandCall records a command execution call
type CommandCall struct {
	Name string
	Args []string
}

// Execute records the call and executes the mock function
func (m *MockCommandExecutor) Execute(name string, args ...string) error {
	m.CallHistory = append(m.CallHistory, CommandCall{
		Name: name,
		Args: args,
	})
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	return nil
}

// TestDetectService tests URL pattern matching for all supported services
func TestDetectService(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected ServiceType
	}{
		// Tidal URLs
		{
			name:     "Tidal URL with listen subdomain",
			url:      "https://listen.tidal.com/track/123456",
			expected: ServiceTidal,
		},
		{
			name:     "Tidal URL without subdomain",
			url:      "https://tidal.com/browse/album/123456",
			expected: ServiceTidal,
		},
		{
			name:     "Tidal URL with http",
			url:      "http://listen.tidal.com/track/123456",
			expected: ServiceTidal,
		},
		{
			name:     "Tidal URL with http without subdomain",
			url:      "http://tidal.com/browse/album/123456",
			expected: ServiceTidal,
		},

		// SoundCloud URLs
		{
			name:     "SoundCloud URL with www",
			url:      "https://www.soundcloud.com/artist/track",
			expected: ServiceSoundCloud,
		},
		{
			name:     "SoundCloud URL without www",
			url:      "https://soundcloud.com/artist/track",
			expected: ServiceSoundCloud,
		},
		{
			name:     "SoundCloud URL with http",
			url:      "http://soundcloud.com/artist/track",
			expected: ServiceSoundCloud,
		},
		{
			name:     "SoundCloud URL with http and www",
			url:      "http://www.soundcloud.com/artist/track",
			expected: ServiceSoundCloud,
		},

		// YouTube URLs
		{
			name:     "YouTube URL with www",
			url:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: ServiceYouTube,
		},
		{
			name:     "YouTube URL without www",
			url:      "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expected: ServiceYouTube,
		},
		{
			name:     "YouTube short URL with www",
			url:      "https://www.youtu.be/dQw4w9WgXcQ",
			expected: ServiceYouTube,
		},
		{
			name:     "YouTube short URL without www",
			url:      "https://youtu.be/dQw4w9WgXcQ",
			expected: ServiceYouTube,
		},
		{
			name:     "YouTube URL with http",
			url:      "http://youtube.com/watch?v=dQw4w9WgXcQ",
			expected: ServiceYouTube,
		},

		// Unknown/Unsupported URLs
		{
			name:     "Spotify URL (unsupported)",
			url:      "https://open.spotify.com/track/123456",
			expected: ServiceUnknown,
		},
		{
			name:     "Invalid URL",
			url:      "not-a-valid-url",
			expected: ServiceUnknown,
		},
		{
			name:     "Empty string",
			url:      "",
			expected: ServiceUnknown,
		},
		{
			name:     "Random website",
			url:      "https://example.com",
			expected: ServiceUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectService(tt.url)
			if result != tt.expected {
				t.Errorf("DetectService(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

// TestServiceTypeString tests the String method of ServiceType
func TestServiceTypeString(t *testing.T) {
	tests := []struct {
		service  ServiceType
		expected string
	}{
		{ServiceTidal, "Tidal"},
		{ServiceSoundCloud, "SoundCloud"},
		{ServiceYouTube, "YouTube"},
		{ServiceUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.service.String()
			if result != tt.expected {
				t.Errorf("ServiceType.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestDownloadTidal tests the Tidal download function
func TestDownloadTidal(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		username    string
		password    string
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful Tidal download",
			url:         "https://listen.tidal.com/track/123456",
			username:    "testuser",
			password:    "testpass",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Failed Tidal download",
			url:         "https://listen.tidal.com/track/999999",
			username:    "testuser",
			password:    "wrongpass",
			mockError:   errors.New("authentication failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				ExecuteFunc: func(name string, args ...string) error {
					return tt.mockError
				},
			}

			err := downloadTidal(tt.url, tt.username, tt.password, mock)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify the command was called correctly
			if len(mock.CallHistory) != 1 {
				t.Fatalf("Expected 1 command call, got %d", len(mock.CallHistory))
			}

			call := mock.CallHistory[0]
			if call.Name != "tidal-dl" {
				t.Errorf("Expected command 'tidal-dl', got %q", call.Name)
			}

			expectedArgs := []string{"-u", tt.username, "-p", tt.password, tt.url}
			if len(call.Args) != len(expectedArgs) {
				t.Errorf("Expected %d args, got %d", len(expectedArgs), len(call.Args))
			}

			for i, arg := range expectedArgs {
				if i < len(call.Args) && call.Args[i] != arg {
					t.Errorf("Arg %d: expected %q, got %q", i, arg, call.Args[i])
				}
			}
		})
	}
}

// TestDownloadSoundcloud tests the SoundCloud download function
func TestDownloadSoundcloud(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful SoundCloud download",
			url:         "https://soundcloud.com/artist/track",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Failed SoundCloud download",
			url:         "https://soundcloud.com/artist/invalid-track",
			mockError:   errors.New("track not found"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				ExecuteFunc: func(name string, args ...string) error {
					return tt.mockError
				},
			}

			err := downloadSoundcloud(tt.url, mock)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify the command was called correctly
			if len(mock.CallHistory) != 1 {
				t.Fatalf("Expected 1 command call, got %d", len(mock.CallHistory))
			}

			call := mock.CallHistory[0]
			if call.Name != "sh" {
				t.Errorf("Expected command 'sh', got %q", call.Name)
			}

			if len(call.Args) < 2 || call.Args[0] != "-c" {
				t.Error("Expected sh -c command format")
			}
		})
	}
}

// TestDownloadYoutube tests the YouTube download function
func TestDownloadYoutube(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful YouTube download",
			url:         "https://youtube.com/watch?v=dQw4w9WgXcQ",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Successful YouTube short URL download",
			url:         "https://youtu.be/dQw4w9WgXcQ",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Failed YouTube download",
			url:         "https://youtube.com/watch?v=invalid",
			mockError:   errors.New("video not available"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandExecutor{
				ExecuteFunc: func(name string, args ...string) error {
					return tt.mockError
				},
			}

			err := downloadYoutube(tt.url, mock)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify the command was called correctly
			if len(mock.CallHistory) != 1 {
				t.Fatalf("Expected 1 command call, got %d", len(mock.CallHistory))
			}

			call := mock.CallHistory[0]
			if call.Name != "ytmdl" {
				t.Errorf("Expected command 'ytmdl', got %q", call.Name)
			}

			if len(call.Args) != 1 || call.Args[0] != tt.url {
				t.Errorf("Expected args [%q], got %v", tt.url, call.Args)
			}
		})
	}
}

// TestMockCommandExecutor tests the mock executor itself
func TestMockCommandExecutor(t *testing.T) {
	mock := &MockCommandExecutor{
		ExecuteFunc: func(name string, args ...string) error {
			if name == "fail" {
				return errors.New("mock error")
			}
			return nil
		},
	}

	// Test successful execution
	err := mock.Execute("success", "arg1", "arg2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test failed execution
	err = mock.Execute("fail")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Verify call history
	if len(mock.CallHistory) != 2 {
		t.Errorf("Expected 2 calls, got %d", len(mock.CallHistory))
	}

	if mock.CallHistory[0].Name != "success" {
		t.Errorf("First call name: expected 'success', got %q", mock.CallHistory[0].Name)
	}

	if mock.CallHistory[1].Name != "fail" {
		t.Errorf("Second call name: expected 'fail', got %q", mock.CallHistory[1].Name)
	}
}
