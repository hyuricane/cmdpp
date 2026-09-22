package store

import (
	"path/filepath"
	"testing"
)

func TestStoreAddUpdateGetRemove(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "commands.json")

	s, err := LoadFrom(filePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// 1. Add command
	cmd, updated, err := s.AddOrUpdate("tunnel-one", "ssh -N -L user@ssh.example.org", "My SSH tunnel")
	if err != nil {
		t.Fatalf("AddOrUpdate failed: %v", err)
	}
	if updated {
		t.Errorf("expected updated=false, got true")
	}
	if cmd.Name != "tunnel-one" {
		t.Errorf("expected name 'tunnel-one', got %s", cmd.Name)
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Reload store from disk
	s2, err := LoadFrom(filePath)
	if err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	found, ok := s2.Get("tunnel-one")
	if !ok {
		t.Fatalf("expected to find 'tunnel-one'")
	}
	if found.Cmd != "ssh -N -L user@ssh.example.org" {
		t.Errorf("unexpected cmd: %s", found.Cmd)
	}
	if found.Description != "My SSH tunnel" {
		t.Errorf("unexpected desc: %s", found.Description)
	}

	// 3. Update command
	_, updated2, err := s2.AddOrUpdate("tunnel-one", "ssh -v -N -L user@ssh.example.org", "Updated tunnel")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !updated2 {
		t.Errorf("expected updated=true, got false")
	}
	_ = s2.Save()

	// 4. Record Run
	if err := s2.RecordRun("tunnel-one"); err != nil {
		t.Fatalf("RecordRun failed: %v", err)
	}
	c, _ := s2.Get("tunnel-one")
	if c.RunCount != 1 || c.LastRunAt == nil {
		t.Errorf("expected RunCount=1 and LastRunAt set, got %d, %v", c.RunCount, c.LastRunAt)
	}

	// 5. Remove command
	removed, err := s2.Remove("tunnel-one")
	if err != nil || !removed {
		t.Fatalf("Remove failed: %v, %v", removed, err)
	}
	if _, ok := s2.Get("tunnel-one"); ok {
		t.Errorf("expected 'tunnel-one' to be removed")
	}
}
