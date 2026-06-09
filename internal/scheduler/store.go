package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ABHINAVGUPTA02/SnapDB/config"
)

type Schedule struct {
	ID        string    `json:"id"`
	Profile   string    `json:"profile,omitempty"`
	Every     string    `json:"every,omitempty"`
	Cron      string    `json:"cron,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	Command   string    `json:"command"`
}

type Store struct {
	Schedules []Schedule `json:"schedules"`
}

func Load() (*Store, error) {
	path, err := config.SchedulesPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Store{Schedules: []Schedule{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var store Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Schedules == nil {
		store.Schedules = []Schedule{}
	}
	return &store, nil
}

func Save(store *Store) error {
	if store == nil {
		return errors.New("schedule store is nil")
	}

	path, err := config.SchedulesPath()
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

func (s *Store) Add(schedule Schedule) {
	s.Schedules = append(s.Schedules, schedule)
	sort.SliceStable(s.Schedules, func(i, j int) bool {
		return s.Schedules[i].CreatedAt.After(s.Schedules[j].CreatedAt)
	})
}

func (s *Store) Delete(id string) (*Schedule, error) {
	for i := range s.Schedules {
		if s.Schedules[i].ID == id {
			deleted := s.Schedules[i]
			s.Schedules = append(s.Schedules[:i], s.Schedules[i+1:]...)
			return &deleted, nil
		}
	}
	return nil, fmt.Errorf("schedule %q not found", id)
}

func NewID(createdAt time.Time) string {
	return "schedule-" + createdAt.UTC().Format("20060102T150405")
}
