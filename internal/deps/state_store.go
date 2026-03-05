package deps

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
)

type DependencyState struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	LastChecked    int64  `json:"last_checked"`
	Status         string `json:"status"`
	NeedsUpdate    bool   `json:"needs_update"`
}

type StateStore struct {
	mu   sync.RWMutex
	path string
	data map[string]DependencyState
}

func NewStateStore(path string) (*StateStore, error) {
	s := &StateStore{
		path: path,
		data: make(map[string]DependencyState),
	}

	file, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(file, &s.data); err != nil {
			slog.Warn("invalid state file, using empty state", "path", path, "error", err)
		}
	}

	return s, nil
}

func (s *StateStore) Save() error {
	s.mu.RLock()
	bytes, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, bytes, 0644)
}

func (s *StateStore) Get(name string) (state DependencyState, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok = s.data[name]
	return state, ok
}

func (s *StateStore) Set(name string, state DependencyState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[name] = state
}

func (s *StateStore) All() []DependencyState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	states := make([]DependencyState, 0, len(s.data))
	for _, state := range s.data {
		states = append(states, state)
	}
	return states
}
