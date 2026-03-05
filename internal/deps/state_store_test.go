package deps

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateStore_ReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	store, err := NewStateStore(path)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}

	state := DependencyState{
		Name:           "yt-dlp",
		CurrentVersion: "2024.1.1",
		LastChecked:    12345,
		Status:         "ok",
	}

	store.Set("yt-dlp", state)

	if err := store.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	store2, err := NewStateStore(path)
	if err != nil {
		t.Fatalf("NewStateStore (reload): %v", err)
	}

	got, ok := store2.Get("yt-dlp")
	if !ok {
		t.Fatal("Get: expected ok")
	}
	if got.Name != state.Name || got.CurrentVersion != state.CurrentVersion {
		t.Errorf("Get: got %+v, want %+v", got, state)
	}
}

func TestStateStore_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := os.WriteFile(path, []byte("not valid json {{{"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store, err := NewStateStore(path)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}

	// Should have empty state, not panic
	_, ok := store.Get("yt-dlp")
	if ok {
		t.Error("Get: expected !ok for empty/invalid state")
	}
}
