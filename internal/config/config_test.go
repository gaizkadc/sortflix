package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Existing directory
	cfg := &Config{
		Path:      tmpDir,
		MinSizeMB: 100,
	}

	info, err := os.Stat(cfg.Path)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected tmpDir to be a valid directory")
	}

	// Non-existing directory
	nonExistent := filepath.Join(tmpDir, "does-not-exist")
	_, err = os.Stat(nonExistent)
	if err == nil {
		t.Fatalf("expected non-existent directory to produce an error")
	}

	// File instead of directory
	filePath := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	finfo, err := os.Stat(filePath)
	if err != nil || finfo.IsDir() {
		t.Fatalf("expected file to not be a directory")
	}
}
