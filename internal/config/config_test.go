package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &Config{
		Path:      tmpDir,
		MinSizeMB: 100,
	}

	info, err := os.Stat(cfg.Path)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected tmpDir to be a valid directory")
	}

	nonExistent := filepath.Join(tmpDir, "does-not-exist")
	_, err = os.Stat(nonExistent)
	if err == nil {
		t.Fatalf("expected non-existent directory to produce an error")
	}
}

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test_config.json")

	content := `{
		"path": "/some/test/path",
		"min_size_mb": 150,
		"video_extensions": [".mp4", ".mkv", ".iso"],
		"delete_junk": false,
		"clean_empty_dirs": true,
		"ignore_articles": true,
		"ignored_articles": ["The", "El"],
		"ignored_files": [".driveignore"],
		"collision_strategy": "skip",
		"dry_run": true,
		"verbose": true
	}`

	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	if err := LoadConfigFile(cfg, jsonFile); err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	if cfg.Path != "/some/test/path" {
		t.Errorf("expected path /some/test/path, got %s", cfg.Path)
	}
	if cfg.MinSizeMB != 150 {
		t.Errorf("expected min_size_mb 150, got %d", cfg.MinSizeMB)
	}
	if len(cfg.VideoExtensions) != 3 {
		t.Errorf("expected 3 video extensions, got %d", len(cfg.VideoExtensions))
	}
	if cfg.DeleteJunk != false {
		t.Errorf("expected delete_junk false, got %v", cfg.DeleteJunk)
	}
	if cfg.IgnoreArticles != true {
		t.Errorf("expected ignore_articles true, got %v", cfg.IgnoreArticles)
	}
	if cfg.CollisionStrategy != "skip" {
		t.Errorf("expected collision_strategy skip, got %s", cfg.CollisionStrategy)
	}
	if cfg.DryRun != true {
		t.Errorf("expected dry_run true, got %v", cfg.DryRun)
	}
	if cfg.Verbose != true {
		t.Errorf("expected verbose true, got %v", cfg.Verbose)
	}
	if cfg.ConfigFile != jsonFile {
		t.Errorf("expected config file %s, got %s", jsonFile, cfg.ConfigFile)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("skipping home dir test: no home dir found")
	}

	expanded := ExpandPath("~/Movies")
	expected := filepath.Join(home, "Movies")
	if expanded != expected {
		t.Errorf("ExpandPath('~/Movies') = %s, expected %s", expanded, expected)
	}

	regular := ExpandPath("/tmp/movies")
	if regular != "/tmp/movies" {
		t.Errorf("ExpandPath('/tmp/movies') = %s, expected /tmp/movies", regular)
	}
}
