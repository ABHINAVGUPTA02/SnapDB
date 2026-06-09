package metadata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
)

type Backup struct {
	ID             string    `json:"id"`
	DatabaseType   string    `json:"databaseType"`
	DatabaseName   string    `json:"databaseName"`
	Profile        string    `json:"profile,omitempty"`
	Path           string    `json:"path"`
	OriginalPath   string    `json:"originalPath,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	SizeBytes      int64     `json:"sizeBytes"`
	ChecksumSHA256 string    `json:"checksumSha256"`
	Compressed     bool      `json:"compressed"`
	Compression    string    `json:"compression,omitempty"`
	Encrypted      bool      `json:"encrypted"`
	UploadedTo     []string  `json:"uploadedTo,omitempty"`
	Status         string    `json:"status"`
	Kind           string    `json:"kind,omitempty"`
}

type Store struct {
	Backups []Backup `json:"backups"`
}

func Load() (*Store, error) {
	path, err := config.MetadataPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Store{Backups: []Backup{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Backups == nil {
		store.Backups = []Backup{}
	}
	return &store, nil
}

func Save(store *Store) error {
	if store == nil {
		return errors.New("metadata store is nil")
	}

	path, err := config.MetadataPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (s *Store) Add(backup Backup) {
	s.Backups = append(s.Backups, backup)
	sort.SliceStable(s.Backups, func(i, j int) bool {
		return s.Backups[i].CreatedAt.After(s.Backups[j].CreatedAt)
	})
}

func (s *Store) Find(idOrPath string) (*Backup, int) {
	for i := range s.Backups {
		if s.Backups[i].ID == idOrPath || s.Backups[i].Path == idOrPath {
			return &s.Backups[i], i
		}
	}
	return nil, -1
}

func (s *Store) Delete(id string) (*Backup, error) {
	backup, idx := s.Find(id)
	if backup == nil {
		return nil, fmt.Errorf("backup %q not found", id)
	}

	deleted := *backup
	s.Backups = append(s.Backups[:idx], s.Backups[idx+1:]...)
	return &deleted, nil
}

func NewID(databaseType string, createdAt time.Time) string {
	return fmt.Sprintf("%s-%s", databaseType, createdAt.UTC().Format("20060102T150405"))
}

func ChecksumFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
