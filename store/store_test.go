package store

import (
	"os"
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

func TestStorePrivatePermissions(t *testing.T) {
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "subconfig", "cmdpp")
	filePath := filepath.Join(storeDir, "commands.json")

	s, err := LoadFrom(filePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	_, _, err = s.AddOrUpdate("token-cmd", "export TOKEN=secret123", "sensitive token")
	if err != nil {
		t.Fatalf("AddOrUpdate failed: %v", err)
	}

	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	dirInfo, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("failed to stat storeDir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != DirPerm {
		t.Errorf("expected directory permissions %04o, got %04o", DirPerm, perm)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat filePath: %v", err)
	}
	if perm := fileInfo.Mode().Perm(); perm != FilePerm {
		t.Errorf("expected file permissions %04o, got %04o", FilePerm, perm)
	}
}

func TestStorePermissionsMigrationOnLoad(t *testing.T) {
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "legacy_config", "cmdpp")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.Chmod(storeDir, 0755); err != nil {
		t.Fatalf("Chmod storeDir failed: %v", err)
	}

	filePath := filepath.Join(storeDir, "commands.json")
	if err := os.WriteFile(filePath, []byte("[]"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := os.Chmod(filePath, 0644); err != nil {
		t.Fatalf("Chmod filePath failed: %v", err)
	}

	// Verify pre-migration state is open
	dInfo, _ := os.Stat(storeDir)
	if dInfo.Mode().Perm() != 0755 {
		t.Fatalf("expected initial dir perm 0755, got %04o", dInfo.Mode().Perm())
	}
	fInfo, _ := os.Stat(filePath)
	if fInfo.Mode().Perm() != 0644 {
		t.Fatalf("expected initial file perm 0644, got %04o", fInfo.Mode().Perm())
	}

	// Loading store should automatically migrate permissions
	s, err := LoadFrom(filePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}

	dInfoAfter, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("stat storeDir failed: %v", err)
	}
	if perm := dInfoAfter.Mode().Perm(); perm != DirPerm {
		t.Errorf("expected migrated dir perm %04o, got %04o", DirPerm, perm)
	}

	fInfoAfter, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat filePath failed: %v", err)
	}
	if perm := fInfoAfter.Mode().Perm(); perm != FilePerm {
		t.Errorf("expected migrated file perm %04o, got %04o", FilePerm, perm)
	}
}

func TestStorePermissionsMigrationOnSave(t *testing.T) {
	tempDir := t.TempDir()
	storeDir := filepath.Join(tempDir, "existing_dir")
	if err := os.MkdirAll(storeDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.Chmod(storeDir, 0755); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	filePath := filepath.Join(storeDir, "commands.json")
	s, err := LoadFrom(filePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// Reset storeDir to 0755 to simulate an existing directory created by older version
	_ = os.Chmod(storeDir, 0755)

	_, _, _ = s.AddOrUpdate("my-cmd", "echo hello", "desc")
	if err := s.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	dInfo, err := os.Stat(storeDir)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := dInfo.Mode().Perm(); perm != DirPerm {
		t.Errorf("expected dir perm %04o, got %04o", DirPerm, perm)
	}

	fInfo, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if perm := fInfo.Mode().Perm(); perm != FilePerm {
		t.Errorf("expected file perm %04o, got %04o", FilePerm, perm)
	}
}

func TestIsSystemOrSharedDir(t *testing.T) {
	if !isSystemOrSharedDir("/") {
		t.Error("expected / to be system dir")
	}
	if !isSystemOrSharedDir(".") {
		t.Error("expected . to be system dir")
	}
	if !isSystemOrSharedDir(os.TempDir()) {
		t.Error("expected os.TempDir() to be system dir")
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		if !isSystemOrSharedDir(home) {
			t.Error("expected home dir to be shared dir")
		}
		if !isSystemOrSharedDir(filepath.Join(home, ".config")) {
			t.Error("expected ~/.config to be shared dir")
		}
		if isSystemOrSharedDir(filepath.Join(home, ".config", "cmdpp")) {
			t.Error("expected ~/.config/cmdpp NOT to be shared dir")
		}
	}
}


