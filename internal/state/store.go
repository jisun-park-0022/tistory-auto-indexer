package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type State struct {
	LastRunAt       time.Time `json:"last_run_at"`
	LastSubmittedAt time.Time `json:"last_submitted_at"`
	KnownURLs       []string  `json:"known_urls"`
}

type StateStore interface {
	Load() (*State, error)
	Save(state *State) error
}

type FileStore struct {
	path string
}

func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (s *FileStore) Load() (*State, error) {
	slog.Debug("loading state", "path", s.path)

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		slog.Info("state file not found, starting fresh", "path", s.path)
		return &State{KnownURLs: []string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		slog.Warn("state file corrupted, starting fresh", "path", s.path, "err", err)
		return &State{KnownURLs: []string{}}, nil
	}

	slog.Info("state loaded",
		"path", s.path,
		"last_run_at", st.LastRunAt,
		"last_submitted_at", st.LastSubmittedAt,
		"known_url_count", len(st.KnownURLs),
	)
	return &st, nil
}

func (s *FileStore) Save(st *State) error {
	slog.Debug("saving state", "path", s.path, "known_url_count", len(st.KnownURLs))

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	slog.Info("state saved", "path", s.path)
	return nil
}
