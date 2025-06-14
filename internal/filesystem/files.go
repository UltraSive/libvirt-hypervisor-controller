package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// SaveFile saves data to a file within a specified directory.
// It will overwrite the file if it already exists.
func SaveFile(dir string, filename string, data []byte) error {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, data, 0644) // 0644 is a common file permission
	if err != nil {
		return fmt.Errorf("failed to save file %s: %w", filePath, err)
	}
	return nil
}

// DeleteFile deletes a file at the specified path.
func DeleteFile(dir, filename string) error {
	filePath := filepath.Join(dir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

// UpdateFile updates the content of an existing file.
func UpdateFile(dir, filename string, data []byte) error {
	filePath := filepath.Join(dir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	return os.WriteFile(filePath, data, 0644) // Overwrite the file with new data
}

// downloadFile handles actual downloading from the URL to a specified path
func DownloadFile(url, filePath string, mode os.FileMode) error {
	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Set file permissions
	return os.Chmod(filePath, mode)
}

// DownloadCachedFile manages the cache logic and uses downloadFile if necessary
func DownloadCachedFile(url string, name string, mode os.FileMode) error {
	// Get cache directory from environment
	cacheDir := os.Getenv("CACHE_DIR")
	useCache := cacheDir != "" // Determine if caching should be used

	// Determine cache duration
	var cacheDuration time.Duration
	cacheSecondsStr := os.Getenv("CACHE_SECONDS")
	if cacheSecondsStr != "" {
		seconds, err := strconv.Atoi(cacheSecondsStr)
		if err == nil {
			cacheDuration = time.Duration(seconds) * time.Second
		} else {
			// Fallback to default if conversion fails
			cacheDuration = 604800 * time.Second // 7 days in seconds
		}
	} else {
		cacheDuration = 604800 * time.Second // Default: 7 days (604800 seconds)
	}

	// If no cache directory is set, directly download and copy the file
	if !useCache {
		// Download the file directly to the destination
		return DownloadFile(url, name, mode)
	}

	// Ensure cache directory exists if caching is enabled
	err := os.MkdirAll(cacheDir, os.ModePerm)
	if err != nil {
		return err
	}

	// Perform a cache clean-up before checking for the file
	err = CleanCache(cacheDir, cacheDuration)
	if err != nil {
		// Log the error but proceed with download logic
		fmt.Printf("Error cleaning cache directory %s: %v\n", cacheDir, err)
	}

	// Determine the filename from the URL
	fileName := filepath.Base(url)
	cacheFilePath := filepath.Join(cacheDir, fileName)

	// Check if file is in the cache and not older than the specified duration
	/*if FileExists(cacheFilePath) && !IsFileOlderThan(cacheFilePath, cacheDuration) {
		// Copy the file from cache to the destination
		return CopyFile(cacheFilePath, name, mode)
	}*/

	// Check if file is in the cache (after cleanup)
	if FileExists(cacheFilePath) {
		// Copy the file from cache to the destination
		return CopyFile(cacheFilePath, name, mode)
	}

	// Download the file into the cache
	err = DownloadFile(url, cacheFilePath, mode)
	if err != nil {
		return err
	}

	// Copy the cached file to the destination
	return CopyFile(cacheFilePath, name, mode)
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsFileOlderThan checks if a file is older than the specified duration
func IsFileOlderThan(path string, duration time.Duration) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > duration
}

// CleanCache sweeps through the cache directory and deletes files older than the specified duration.
func CleanCache(cacheDir string, duration time.Duration) error {
	files, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue // Skip subdirectories
		}
		filePath := filepath.Join(cacheDir, file.Name())
		if time.Since(file.ModTime()) > duration {
			err := os.Remove(filePath)
			if err != nil {
				// Log the error but continue to clean other files
				// fmt.Printf("Error deleting file %s: %v\n", filePath, err)
			}
		}
	}
	return nil
}

// CopyFile copies a file from src to dst with the specified mode
func CopyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}
