package deps

import (
	"log/slog"
	"sync"
	"time"
)

type Manager struct {
	deps  []Dependency
	store *StateStore
	mu    sync.Mutex
}

func NewManager(store *StateStore) *Manager {
	return &Manager{
		store: store,
	}
}

func (m *Manager) Add(dep Dependency) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.deps {
		if d.GetName() == dep.GetName() {
			return
		}
	}
	m.deps = append(m.deps, dep)
}

func (m *Manager) Bootstrap(progress ProgressCallback) error {
	m.mu.Lock()
	deps := make([]Dependency, len(m.deps))
	copy(deps, m.deps)
	m.mu.Unlock()

	for _, dep := range deps {

		wrappedProgress := func(downloaded, total int64) {
			if progress == nil {
				return
			}
			percentage := 0
			if total > 0 {
				percentage = int(float64(downloaded) / float64(total) * 100)
			}
			progress(ProgressEvent{
				Name:            dep.GetName(),
				DownloadedBytes: downloaded,
				TotalBytes:      total,
				Percentage:      percentage,
			})
		}

		installed, err := dep.CheckInstalled()
		if err != nil {
			return err
		}

		if !installed {
			if err := dep.Install(wrappedProgress); err != nil {
				m.markFailed(dep)
				return err
			}
			m.markChecked(dep)
			continue
		}

		if m.ShouldUpdate(dep) {
			if err := dep.Update(wrappedProgress); err != nil {
				m.markFailed(dep)
				return err
			}
			m.markChecked(dep)
		} else {
			m.markChecked(dep)
		}

	}
	return nil
}

func (m *Manager) ShouldUpdate(dep Dependency) bool {
	ttl := dep.TTL()
	if ttl == 0 {
		return false
	}
	state, ok := m.store.Get(dep.GetName())
	if !ok {
		return true
	}

	last := time.Unix(state.LastChecked, 0)
	return time.Since(last) > ttl
}

func (m *Manager) DependenciesState() []DependencyState {
	m.mu.Lock()
	deps := make([]Dependency, len(m.deps))
	copy(deps, m.deps)
	m.mu.Unlock()

	states := make([]DependencyState, 0, len(deps))
	for _, dep := range deps {
		state, _ := m.store.Get(dep.GetName())
		if state.Name == "" {
			state.Name = dep.GetName()
		}

		installed, err := dep.CheckInstalled()
		if err != nil || !installed {
			state.Status = "failed"
			state.CurrentVersion = ""
			state.NeedsUpdate = true
			states = append(states, state)
			continue
		}

		actualVersion, err := dep.Version()
		if err != nil {
			state.Status = "failed"
			state.NeedsUpdate = true
			states = append(states, state)
			continue
		}

		storedVersion := state.CurrentVersion
		state.CurrentVersion = actualVersion

		if storedVersion != actualVersion {
			state.Status = "failed"
			state.NeedsUpdate = true
		} else {
			state.Status = "ok"
			state.NeedsUpdate = m.ShouldUpdate(dep)
		}
		states = append(states, state)
	}
	return states
}

func (m *Manager) GetPath(name string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, dep := range m.deps {
		if dep.GetName() == name {
			return dep.Path(), true
		}
	}
	return "", false
}

func (m *Manager) mark(dep Dependency, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	version, _ := dep.Version()

	m.store.Set(dep.GetName(), DependencyState{
		CurrentVersion: version,
		LastChecked:    time.Now().Unix(),
		Status:         status,
		Name:           dep.GetName(),
	})

	if err := m.store.Save(); err != nil {
		slog.Error("failed to save dependency state", "dep", dep.GetName(), "error", err)
	}
}

func (m *Manager) markChecked(dep Dependency) {
	m.mark(dep, "ok")
}

func (m *Manager) markFailed(dep Dependency) {
	m.mark(dep, "failed")
}
