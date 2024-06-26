package datastorage

import (
	"os"
)

// LocalStorage represents a file on the filesytem.
type LocalStorage struct {
	path string
}

// NewLocalStorage returns a local storage object given a file path.
func NewLocalStorage(path string) *LocalStorage {
	return &LocalStorage{
		path: path,
	}
}

// Load returns a file contents from the filesystem.
func (s *LocalStorage) Load() ([]byte, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Save writes a file to the filesystem and returns an error if one occurs.
func (s *LocalStorage) Save(b []byte) error {
	return os.WriteFile(s.path, b, 0644)
}
