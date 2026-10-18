package state

import (
	"errors"
	"sort"
	"sync"

	"example.edu/6cc545/cw1/internal/model"
)

// Catalog contains metadata, not mutable file contents. File I/O belongs in CW2.
type Catalog struct {
	mu    sync.RWMutex
	files []model.FileMetadata
}

func NewCatalog() *Catalog { return &Catalog{files: []model.FileMetadata{}} }
func (c *Catalog) Replace(input []model.FileMetadata) error {
	// Prepare an independent candidate collection.
	candidate := make([]model.FileMetadata, 0, len(input))

	// Track filenames to reject duplicates.
	seen := make(map[string]bool)

	// Validate the entire input before changing state.
	for _, file := range input {
		if err := file.Validate(); err != nil {
			return err
		}

		if seen[file.Name] {
			return errors.New("duplicate filename: " + file.Name)
		}

		seen[file.Name] = true
		candidate = append(candidate, file)
	}

	// Sort the independent candidate by filename.
	sort.Slice(candidate, func(i, j int) bool {
		return candidate[i].Name < candidate[j].Name
	})

	// Atomically replace the existing collection.
	c.mu.Lock()
	defer c.mu.Unlock()

	c.files = candidate

	return nil
}

func (c *Catalog) Snapshot() []model.FileMetadata {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy rather than exposing internal storage.
	files := make([]model.FileMetadata, len(c.files))
	copy(files, c.files)

	return files
}

var _ = errors.New
var _ = sort.Slice
