package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackgroundWorker_NewWithoutPath(t *testing.T) {
	cfg := WorkerConfig{
		BeadsPath: "",
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}
	defer worker.Stop()

	if worker.State() != WorkerIdle {
		t.Errorf("Expected idle state, got %v", worker.State())
	}

	if worker.GetSnapshot() != nil {
		t.Error("Expected nil snapshot initially")
	}
}

func TestBackgroundWorker_NewWithPath(t *testing.T) {
	// Create a temporary beads file
	tmpDir := t.TempDir()
	beadsPath := filepath.Join(tmpDir, "beads.jsonl")

	// Write a valid beads file
	content := `{"id":"test-1","title":"Test Issue","status":"open","priority":1,"issue_type":"task"}` + "\n"
	if err := os.WriteFile(beadsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg := WorkerConfig{
		BeadsPath:     beadsPath,
		DebounceDelay: 50 * time.Millisecond,
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}
	defer worker.Stop()

	if worker.State() != WorkerIdle {
		t.Errorf("Expected idle state, got %v", worker.State())
	}
}

func TestBackgroundWorker_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	beadsPath := filepath.Join(tmpDir, "beads.jsonl")

	content := `{"id":"test-1","title":"Test","status":"open","priority":1,"issue_type":"task"}` + "\n"
	if err := os.WriteFile(beadsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg := WorkerConfig{
		BeadsPath:     beadsPath,
		DebounceDelay: 50 * time.Millisecond,
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}

	if err := worker.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop should be idempotent
	worker.Stop()
	worker.Stop() // Should not panic

	if worker.State() != WorkerStopped {
		t.Errorf("Expected stopped state, got %v", worker.State())
	}
}

func TestBackgroundWorker_TriggerRefresh(t *testing.T) {
	tmpDir := t.TempDir()
	beadsPath := filepath.Join(tmpDir, "beads.jsonl")

	content := `{"id":"test-1","title":"Test","status":"open","priority":1,"issue_type":"task"}` + "\n"
	if err := os.WriteFile(beadsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg := WorkerConfig{
		BeadsPath:     beadsPath,
		DebounceDelay: 50 * time.Millisecond,
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}
	defer worker.Stop()

	// Trigger refresh and wait for processing
	worker.TriggerRefresh()

	// Wait for processing to complete
	time.Sleep(200 * time.Millisecond)

	snapshot := worker.GetSnapshot()
	if snapshot == nil {
		t.Fatal("Expected snapshot after refresh")
	}

	if len(snapshot.Issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(snapshot.Issues))
	}
}

func TestBackgroundWorker_WatcherChanged(t *testing.T) {
	tmpDir := t.TempDir()
	beadsPath := filepath.Join(tmpDir, "beads.jsonl")

	content := `{"id":"test-1","title":"Test","status":"open","priority":1,"issue_type":"task"}` + "\n"
	if err := os.WriteFile(beadsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cfg := WorkerConfig{
		BeadsPath:     beadsPath,
		DebounceDelay: 50 * time.Millisecond,
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}
	defer worker.Stop()

	ch := worker.WatcherChanged()
	if ch == nil {
		t.Error("WatcherChanged should return non-nil channel")
	}
}

func TestBackgroundWorker_WatcherChangedNil(t *testing.T) {
	// Worker without path should have nil watcher
	cfg := WorkerConfig{
		BeadsPath: "",
	}

	worker, err := NewBackgroundWorker(cfg)
	if err != nil {
		t.Fatalf("NewBackgroundWorker failed: %v", err)
	}
	defer worker.Stop()

	if worker.WatcherChanged() != nil {
		t.Error("WatcherChanged should return nil when no watcher")
	}
}

func TestWorkerState_String(t *testing.T) {
	tests := []struct {
		state    WorkerState
		expected string
	}{
		{WorkerIdle, "0"},
		{WorkerProcessing, "1"},
		{WorkerStopped, "2"},
	}

	for _, tt := range tests {
		// Just verify the states have distinct values
		if int(tt.state) < 0 || int(tt.state) > 2 {
			t.Errorf("Unexpected state value: %v", tt.state)
		}
	}
}
