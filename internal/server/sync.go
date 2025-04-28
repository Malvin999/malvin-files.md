package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Configuration
const (
	StorageDir = "/app/mystorage"             // Base directory for markdown storage
	AuthToken  = "your-really-secret-token-2" // Replace with your actual token or get from environment
)

// FileInfo represents metadata about a file with Unix timestamp
type FileInfo struct {
	Path         string `json:"path"`
	LastModified int64  `json:"last_modified"` // Unix timestamp
	IsDirectory  bool   `json:"is_directory"`
	Content      string `json:"content,omitempty"` // Only filled for sync requests
}

type SyncRequest struct {
	Timestamps map[string]int64 `json:"timestamps"` // Map of paths to last modified times in Unix format
}

// SyncResponse is the structure returned when syncing with Unix timestamps
type SyncResponse struct {
	Files      []FileInfo       `json:"files"`       // Files with content that need syncing
	Timestamps map[string]int64 `json:"timestamps"`  // Current server timestamps in Unix format
	ServerTime int64            `json:"server_time"` // Current server time in Unix format
}

// getDirectoryTimestamps recursively scans a directory and returns the latest modification time
// for directories only (including root directory) as Unix timestamps
func getDirectoryTimestamps(rootPath string) (map[string]int64, error) {
	timestamps := make(map[string]int64)
	timeObjects := make(map[string]time.Time) // Used for comparing times

	// Resolve the symlink if it exists
	realPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		log.Printf("Warning: Could not resolve symlink: %v. Using original path.", err)
		realPath = rootPath
	} else {
		log.Printf("Resolved symlink: %s -> %s", rootPath, realPath)
	}

	// Walk through the directory
	err = filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip hidden directories and files (starting with .)
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != realPath {
			if info.IsDir() {
				return filepath.SkipDir // Skip the entire directory
			}
			return nil // Skip the file
		}

		// Get the relative path
		relPath, err := filepath.Rel(realPath, path)
		if err != nil {
			return nil
		}

		// Use "." for the root directory
		if relPath == "." || relPath == "" {
			relPath = "."
		}

		// For directories, track their timestamp
		if info.IsDir() {
			modTime := info.ModTime()
			timeObjects[relPath] = modTime
			timestamps[relPath] = modTime.Unix()
			return nil
		}

		// Skip non-markdown files for file processing
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// For files, only update the parent directory's timestamp if needed
		dirPath := filepath.Dir(relPath)
		if dirPath == "." || dirPath == "" {
			dirPath = "."
		}

		modTime := info.ModTime()
		if dirTime, exists := timeObjects[dirPath]; !exists || modTime.After(dirTime) {
			timeObjects[dirPath] = modTime
			timestamps[dirPath] = modTime.Unix()
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return timestamps, nil
}

// validateAuthToken checks if the request has a valid auth token
func validateAuthToken(r *http.Request) bool {
	token := r.Header.Get("Authorization")

	// Check for "Bearer " prefix and remove if present
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	return token == AuthToken
}

// AuthMiddleware wraps a handler and adds token authentication
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateAuthToken(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// Timestamps handles the endpoint for getting current timestamps
func Timestamps(w http.ResponseWriter, r *http.Request) {
	// Auth check is now handled by middleware
	// Get the latest timestamps for all directories and files
	timestamps, err := getDirectoryTimestamps(StorageDir)
	if err != nil {
		log.Printf("Error getting timestamps: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get timestamps: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the timestamps
	response := struct {
		Timestamps map[string]int64 `json:"timestamps"`
		ServerTime int64            `json:"server_time"`
	}{
		Timestamps: timestamps,
		ServerTime: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding timestamp response: %v", err)
	}
}

func Sync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error parsing request JSON: %v", err)
		http.Error(w, "Invalid request JSON", http.StatusBadRequest)
		return
	}

	serverTimestampsUnix, err := getDirectoryTimestamps(StorageDir)
	if err != nil {
		log.Printf("Error getting server timestamps: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get timestamps: %v", err), http.StatusInternalServerError)
		return
	}

	serverTimestamps := make(map[string]time.Time)

	serverFileTimestamps := make(map[string]time.Time)
	for path, unixTime := range serverTimestampsUnix {
		timeObj := time.Unix(unixTime, 0)
		serverTimestamps[path] = timeObj
		serverFileTimestamps[path] = timeObj
	}

	realStorageDir, _ := filepath.EvalSymlinks(StorageDir)
	filepath.Walk(realStorageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != realStorageDir {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// TODO add config
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(realStorageDir, path)
		if err != nil {
			return nil
		}

		serverTimestamps[relPath] = info.ModTime()

		return nil
	})

	filesToSync := make([]FileInfo, 0)

	realStorageDir, err = filepath.EvalSymlinks(StorageDir)
	if err != nil {
		log.Printf("Warning: Could not resolve symlink: %v. Using original path.", err)
		realStorageDir = StorageDir
	}

	err = filepath.Walk(realStorageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and files
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != realStorageDir {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Get the relative path
		relPath, err := filepath.Rel(realStorageDir, path)
		if err != nil {
			return nil
		}

		clientUnixTime, clientHasFile := request.Timestamps[relPath]
		clientTime := time.Time{}
		if clientHasFile {
			clientTime = time.Unix(clientUnixTime, 0)
		}

		if !clientHasFile || info.ModTime().After(clientTime) {
			// Client doesn't have the file or has an older version
			// Read the file content
			content, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Error reading file %s: %v", relPath, err)
				return nil
			}

			// Add the file to the response
			filesToSync = append(filesToSync, FileInfo{
				Path:         relPath,
				LastModified: info.ModTime().Unix(),
				IsDirectory:  false,
				Content:      string(content),
			})
		}

		return nil
	})

	if err != nil {
		log.Printf("Error scanning files: %v", err)
		http.Error(w, fmt.Sprintf("Error scanning files: %v", err), http.StatusInternalServerError)
		return
	}

	response := SyncResponse{
		Files:      filesToSync,
		Timestamps: serverTimestampsUnix,
		ServerTime: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding sync response: %v", err)
	}
}
