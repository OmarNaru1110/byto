package deps

import (
	"path/filepath"
	"testing"
	"time"
)

func TestManager_GetPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	store, err := NewStateStore(path)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}

	m := NewManager(store)
	dep := NewYTDLPDependency(dir, time.Hour)
	m.Add(dep)

	got, ok := m.GetPath("yt-dlp")
	if !ok {
		t.Fatal("GetPath: expected ok")
	}
	want := dep.Path()
	if got != want {
		t.Errorf("GetPath: got %q, want %q", got, want)
	}

	_, ok = m.GetPath("nonexistent")
	if ok {
		t.Error("GetPath: expected !ok for nonexistent dep")
	}
}
