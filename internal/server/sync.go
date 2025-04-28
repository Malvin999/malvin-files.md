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
	StorageDir = "/app/mystorage"
	AuthToken  = "your-really-secret-token-2"
)

type File struct {
	Path         string `json:"path"`
	LastModified int64  `json:"last_modified"`
	IsDir        bool   `json:"is_directory"`
	Content      string `json:"content,omitempty"`
}

type syncRequest struct {
	Timestamps map[string]int64 `json:"timestamps"`
}

type syncResponse struct {
	Files      []File           `json:"files"`       // Files with content that need syncing
	Timestamps map[string]int64 `json:"timestamps"`  // Current server timestamps in Unix format
	ServerTime int64            `json:"server_time"` // Current server time in Unix format
}

// validateAuthToken checks if the request has a valid auth token
func validateAuthToken(r *http.Request) bool {
	token := r.Header.Get("Authorization")

	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}

	return token == AuthToken
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateAuthToken(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func Timestamps(w http.ResponseWriter, r *http.Request) {
	timestamps, err := timestamps(StorageDir)
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

	var request syncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error parsing request JSON: %v", err)
		http.Error(w, "Invalid request JSON", http.StatusBadRequest)
		return
	}

	serverTimestamps, err := timestamps(StorageDir)
	if err != nil {
		log.Printf("Error getting server timestamps: %v", err)
		http.Error(w, fmt.Sprintf("Failed to get timestamps: %v", err), http.StatusInternalServerError)
		return
	}

	missingFiles := make([]File, 0)
	for path, serverTime := range serverTimestamps {
		parts := strings.Split(path, string(os.PathSeparator))
		dir := parts[0]
		requestTime, exists := request.Timestamps[dir]
		if !exists || serverTime > requestTime {
			missingFiles = append(missingFiles, File{path, 0, false, "content"})
		}
	}

	response := syncResponse{
		Files:      missingFiles,
		Timestamps: serverTimestamps,
		ServerTime: time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding sync response: %v", err)
	}
}

// timestamps recursively scans a directory and returns the latest modification time
// for directories only (including root directory) as Unix timestamps
func timestamps(rootPath string) (map[string]int64, error) {
	timestamps := make(map[string]int64)
	realPath, err := filepath.EvalSymlinks(rootPath)
	if err != nil {
		log.Printf("Warning: Could not resolve symlink: %v. Using original path.", err)
		realPath = rootPath
	} else {
		log.Printf("Resolved symlink: %s -> %s", rootPath, realPath)
	}

	err = filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") && path != realPath {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(realPath, path)
		if err != nil {
			return nil
		}

		// Skip non-markdown files for file processing
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		if relPath == "" {
			relPath = "."
		}

		// Record timestamp for all directories and files
		modTime := info.ModTime()
		timestamps[relPath] = modTime.Unix()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return timestamps, nil
}
