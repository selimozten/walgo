package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// HashFile computes the SHA-256 hash of a file
func HashFile(filePath string) (string, error) {
	// #nosec G304 - filePath is controlled by the application
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HashDirectory computes hashes for all files in a directory
func HashDirectory(dir string) (map[string]string, error) {
	hashes := make(map[string]string)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path first to check for cache directory
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip the cache directory
		if relPath == CacheDir || filepath.HasPrefix(relPath, CacheDir+string(filepath.Separator)) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Compute hash
		hash, err := HashFile(path)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", relPath, err)
		}

		hashes[relPath] = hash
		return nil
	})

	if err != nil {
		return nil, err
	}

	return hashes, nil
}

// CompareHashes compares two hash maps and returns the changes
func CompareHashes(old, new map[string]string) *ChangeSet {
	changes := &ChangeSet{
		Added:    make([]string, 0),
		Modified: make([]string, 0),
		Deleted:  make([]string, 0),
		Unchanged: make([]string, 0),
	}

	// Find added and modified files
	for path, newHash := range new {
		if oldHash, exists := old[path]; exists {
			if oldHash == newHash {
				changes.Unchanged = append(changes.Unchanged, path)
			} else {
				changes.Modified = append(changes.Modified, path)
			}
		} else {
			changes.Added = append(changes.Added, path)
		}
	}

	// Find deleted files
	for path := range old {
		if _, exists := new[path]; !exists {
			changes.Deleted = append(changes.Deleted, path)
		}
	}

	return changes
}
